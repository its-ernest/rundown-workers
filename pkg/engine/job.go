package engine

import "time"

// JobStatus represents the current lifecycle state of a workflow job.
type JobStatus string

const (
	// StatusPending means the job is in the queue and waiting for a worker.
	StatusPending JobStatus = "pending"

	// StatusRunning means a worker has picked up the job and is currently executing it.
	StatusRunning JobStatus = "running"

	// StatusDone means the job completed successfully.
	StatusDone JobStatus = "done"

	// StatusFailed means the job failed all retry attempts or was marked as fatal.
	StatusFailed JobStatus = "failed"
)

// Job is the primary unit of work in Rundown-Workers.
//
// It contains the payload to be processed, the current status, and
// metadata required for retry logic and timeout enforcement.
type Job struct {
	ID         string    `json:"id"`
	Queue      string    `json:"queue"`
	Tag        string    `json:"tag,omitempty"`
	Payload    string    `json:"payload"`
	Status     JobStatus `json:"status"`
	Retries    int       `json:"retries"`     // How many times this job has been attempted.
	MaxRetries int       `json:"max_retries"` // Maximum allowed retries before moving to StatusFailed.
	Timeout    int       `json:"timeout"`     // Total seconds allowed for execution before being marked stale.
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	NextRunAt  time.Time `json:"next_run_at"` // The time when the job is eligible to be picked up again (for retries).
}
