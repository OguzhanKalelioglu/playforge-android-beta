-- 0003 — Task jobs audit table
-- Asynq task'larının audit trail'i (idempotency + debug için)

CREATE TABLE IF NOT EXISTS task_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id VARCHAR(255) UNIQUE NOT NULL,        -- Asynq job ID
    test_id UUID REFERENCES tests(id) ON DELETE CASCADE,
    assignment_id UUID REFERENCES test_assignments(id) ON DELETE SET NULL,
    task_type VARCHAR(50) NOT NULL,
    day INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
        -- pending | running | completed | failed | retrying | dead
    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,
    payload_encrypted BYTEA,                   -- şifrelenmiş task payload
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_task_jobs_test
    ON task_jobs(test_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_task_jobs_status
    ON task_jobs(status, updated_at DESC) WHERE status IN ('pending', 'running', 'retrying');
CREATE INDEX IF NOT EXISTS idx_task_jobs_type_day
    ON task_jobs(task_type, day);

-- updated_at trigger
DROP TRIGGER IF EXISTS update_task_jobs_updated_at ON task_jobs;
CREATE TRIGGER update_task_jobs_updated_at
    BEFORE UPDATE ON task_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
