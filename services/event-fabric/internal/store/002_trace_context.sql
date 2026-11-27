ALTER TABLE incident_outbox ADD COLUMN IF NOT EXISTS traceparent text NOT NULL DEFAULT '';
ALTER TABLE remediation_jobs ADD COLUMN IF NOT EXISTS traceparent text NOT NULL DEFAULT '';
