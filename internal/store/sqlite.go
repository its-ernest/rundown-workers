package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/its-ernest/rundown-workers/pkg/engine"
	_ "modernc.org/sqlite"
)

// Inside internal/store/store.go

type Store interface {
	Enqueue(queue, tag, payload string, timeout, maxRetries int) (*engine.Job, error)
	Poll(queue string) (*engine.Job, error)
	Complete(id string) error
	Fail(id string) error
	CleanupStale() (int64, error)
	GetJob(id string) (*engine.Job, error)
	GetJobByTag(tag string) (*engine.Job, error)
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes

	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA synchronous=NORMAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")

	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id          TEXT PRIMARY KEY,
		queue       TEXT NOT NULL,
		tag         TEXT DEFAULT '',
		payload     TEXT NOT NULL,
		status      TEXT NOT NULL,
		retries     INTEGER DEFAULT 0,
		max_retries INTEGER DEFAULT 0,
		timeout     INTEGER DEFAULT 300,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		next_run_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_jobs_queue_status_next ON jobs(queue, status, next_run_at);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	// Migration: Add tag column if it doesn't exist (from older versions)
	// We ignore the error because SQLite doesn't have "IF NOT EXISTS" for ADD COLUMN
	_, _ = db.Exec("ALTER TABLE jobs ADD COLUMN tag TEXT DEFAULT '';")

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) GetJob(id string) (*engine.Job, error) {
	var job engine.Job
	err := s.db.QueryRow(
		`SELECT id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at 
         FROM jobs WHERE id = ?`, id,
	).Scan(
		&job.ID, &job.Queue, &job.Tag, &job.Payload, &job.Status,
		&job.Retries, &job.MaxRetries, &job.Timeout,
		&job.CreatedAt, &job.UpdatedAt, &job.NextRunAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("job not found")
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *SQLiteStore) GetJobByTag(tag string) (*engine.Job, error) {
	var job engine.Job
	err := s.db.QueryRow(
		`SELECT id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at 
         FROM jobs WHERE tag = ? ORDER BY created_at DESC LIMIT 1`, tag,
	).Scan(
		&job.ID, &job.Queue, &job.Tag, &job.Payload, &job.Status,
		&job.Retries, &job.MaxRetries, &job.Timeout,
		&job.CreatedAt, &job.UpdatedAt, &job.NextRunAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *SQLiteStore) Enqueue(queue, tag, payload string, timeout, maxRetries int) (*engine.Job, error) {
	if timeout <= 0 {
		timeout = 300
	}
	if maxRetries < 0 {
		maxRetries = 0
	}

	now := time.Now().UTC()
	job := &engine.Job{
		ID:         uuid.New().String(),
		Queue:      queue,
		Tag:        tag,
		Payload:    payload,
		Status:     engine.StatusPending,
		Retries:    0,
		MaxRetries: maxRetries,
		Timeout:    timeout,
		CreatedAt:  now,
		UpdatedAt:  now,
		NextRunAt:  now,
	}

	_, err := s.db.Exec(
		`INSERT INTO jobs (id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Queue, job.Tag, job.Payload, job.Status,
		job.Retries, job.MaxRetries, job.Timeout,
		job.CreatedAt, job.UpdatedAt, job.NextRunAt,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *SQLiteStore) Poll(queue string) (*engine.Job, error) {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	var job engine.Job

	err = tx.QueryRowContext(ctx,
		`SELECT id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at
		 FROM jobs
		 WHERE queue = ? AND status = ? AND next_run_at <= ?
		 ORDER BY created_at ASC
		 LIMIT 1`,
		queue, engine.StatusPending, now,
	).Scan(
		&job.ID, &job.Queue, &job.Tag, &job.Payload, &job.Status,
		&job.Retries, &job.MaxRetries, &job.Timeout,
		&job.CreatedAt, &job.UpdatedAt, &job.NextRunAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`,
		engine.StatusRunning, now, job.ID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	job.Status = engine.StatusRunning
	job.UpdatedAt = now
	return &job, nil
}

func (s *SQLiteStore) Complete(id string) error {
	_, err := s.db.Exec(
		`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`,
		engine.StatusDone, time.Now().UTC(), id,
	)
	return err
}

func (s *SQLiteStore) Fail(id string) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var retries, maxRetries int
	err = tx.QueryRowContext(ctx,
		`SELECT retries, max_retries FROM jobs WHERE id = ?`, id,
	).Scan(&retries, &maxRetries)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	if retries < maxRetries {
		backoff := 5 * (retries + 1) * (retries + 1)
		nextRun := now.Add(time.Duration(backoff) * time.Second)
		_, err = tx.ExecContext(ctx,
			`UPDATE jobs SET status = ?, retries = retries + 1, updated_at = ?, next_run_at = ? WHERE id = ?`,
			engine.StatusPending, now, nextRun, id,
		)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE jobs SET status = ?, updated_at = ? WHERE id = ?`,
			engine.StatusFailed, now, id,
		)
	}

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *SQLiteStore) CleanupStale() (int64, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`UPDATE jobs
		 SET status = ?, updated_at = ?, retries = retries + 1
		 WHERE status = ?
		 AND (strftime('%s', ?) - strftime('%s', updated_at)) > timeout`,
		engine.StatusPending, now, engine.StatusRunning, now,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
