CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payload TEXT NOT NULL DEFAULT '',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

-- The output row is both the effect and its durable idempotency record.
CREATE TABLE IF NOT EXISTS job_outputs (
    job_id INTEGER PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
    payload TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
