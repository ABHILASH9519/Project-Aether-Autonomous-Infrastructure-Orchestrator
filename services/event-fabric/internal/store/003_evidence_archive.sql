ALTER TABLE remediation_jobs ADD COLUMN IF NOT EXISTS archive_attempts bigint NOT NULL DEFAULT 0 CHECK (archive_attempts >= 0);
ALTER TABLE remediation_jobs ADD COLUMN IF NOT EXISTS archive_available_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE remediation_jobs ADD COLUMN IF NOT EXISTS archived_at timestamptz;
ALTER TABLE remediation_jobs ADD COLUMN IF NOT EXISTS archive_sha256 text NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS remediation_jobs_archive_pending ON remediation_jobs(archive_available_at,correlation_id) WHERE status IN ('succeeded','failed') AND archived_at IS NULL;
