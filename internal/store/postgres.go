package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/its-ernest/rundown-workers/pkg/engine"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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
		created_at  TIMESTAMP DEFAULT NOW(),
		updated_at  TIMESTAMP DEFAULT NOW(),
		next_run_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_jobs_queue_status_next ON jobs(queue, status, next_run_at);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) GetJob(id string) (*engine.Job, error) {
	var job engine.Job
	err := s.db.QueryRow(
		`SELECT id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at
         FROM jobs WHERE id = $1`, id,
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

func (s *PostgresStore) GetJobByTag(tag string) (*engine.Job, error) {
	var job engine.Job
	err := s.db.QueryRow(
		`SELECT id, queue, tag, payload, status, retries, max_retries, timeout, created_at, updated_at, next_run_at
         FROM jobs WHERE tag = $1 ORDER BY created_at DESC LIMIT 1`, tag,
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

func (s *PostgresStore) Enqueue(queue, tag, payload string, timeout, maxRetries int) (*engine.Job, error) {
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
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		job.ID, job.Queue, job.Tag, job.Payload, job.Status,
		job.Retries, job.MaxRetries, job.Timeout,
		job.CreatedAt, job.UpdatedAt, job.NextRunAt,
	)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *PostgresStore) Poll(queue string) (*engine.Job, error) {
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
		 WHERE queue = $1 AND status = $2 AND next_run_at <= $3
		 ORDER BY created_at ASC
		 LIMIT 1
		 FOR UPDATE SKIP LOCKED`,
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
		`UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`,
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

func (s *PostgresStore) Complete(id string) error {
	_, err := s.db.Exec(
		`UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`,
		engine.StatusDone, time.Now().UTC(), id,
	)
	return err
}

func (s *PostgresStore) Fail(id string) error {
	ctx := context.Background()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var retries, maxRetries int
	err = tx.QueryRowContext(ctx,
		`SELECT retries, max_retries FROM jobs WHERE id = $1`, id,
	).Scan(&retries, &maxRetries)
	if err != nil {
		return err
	}

	now := time.Now().UTC()

	if retries < maxRetries {
		backoff := 5 * (retries + 1) * (retries + 1)
		nextRun := now.Add(time.Duration(backoff) * time.Second)
		_, err = tx.ExecContext(ctx,
			`UPDATE jobs SET status = $1, retries = retries + 1, updated_at = $2, next_run_at = $3 WHERE id = $4`,
			engine.StatusPending, now, nextRun, id,
		)
	} else {
		_, err = tx.ExecContext(ctx,
			`UPDATE jobs SET status = $1, updated_at = $2 WHERE id = $3`,
			engine.StatusFailed, now, id,
		)
	}

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *PostgresStore) CleanupStale() (int64, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(
		`UPDATE jobs
		 SET status = $1, updated_at = $2, retries = retries + 1
		 WHERE status = $3
		 AND EXTRACT(EPOCH FROM ($2::timestamp - updated_at)) > timeout`,
		engine.StatusPending, now, engine.StatusRunning,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
