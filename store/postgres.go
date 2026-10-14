package store // PostgreSQL implementation of the Store interface.

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goqueue/job"
)

// PostgresStore keeps jobs in PostgreSQL so they survive application restarts.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore opens a connection pool and verifies the database is reachable.
func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &PostgresStore{pool: pool}, nil
}

// Create inserts a job and lets PostgreSQL generate its ID.
func (s *PostgresStore) Create(ctx context.Context, newJob *job.Job) error {
	const query = `
		INSERT INTO jobs (
			type,
			status,
			payload,
			retry_count,
			max_retries,
			error,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	return s.pool.QueryRow(
		ctx,
		query,
		newJob.Type,
		newJob.Status,
		newJob.Payload,
		newJob.RetryCount,
		newJob.MaxRetries,
		newJob.Error,
		newJob.CreatedAt,
	).Scan(&newJob.ID)
}

// Get loads one job by ID.
func (s *PostgresStore) Get(ctx context.Context, id int) (job.Job, bool, error) {
	const query = `
		SELECT
			id,
			type,
			status,
			payload,
			retry_count,
			max_retries,
			error,
			created_at
		FROM jobs
		WHERE id = $1
	`

	var storedJob job.Job
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&storedJob.ID,
		&storedJob.Type,
		&storedJob.Status,
		&storedJob.Payload,
		&storedJob.RetryCount,
		&storedJob.MaxRetries,
		&storedJob.Error,
		&storedJob.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return job.Job{}, false, nil
	}
	if err != nil {
		return job.Job{}, false, err
	}

	return storedJob, true, nil
}

// Update persists the latest state of an existing job.
func (s *PostgresStore) Update(ctx context.Context, updatedJob job.Job) error {
	const query = `
		UPDATE jobs
		SET
			status = $2,
			retry_count = $3,
			max_retries = $4,
			error = $5
		WHERE id = $1
	`

	_, err := s.pool.Exec(
		ctx,
		query,
		updatedJob.ID,
		updatedJob.Status,
		updatedJob.RetryCount,
		updatedJob.MaxRetries,
		updatedJob.Error,
	)
	return err
}

// Close releases all PostgreSQL connections in the pool.
func (s *PostgresStore) Close() {
	s.pool.Close()
}
