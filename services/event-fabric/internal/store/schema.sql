CREATE TABLE IF NOT EXISTS incident_inbox (
 event_id text PRIMARY KEY,
 received_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS incidents (
 correlation_id text PRIMARY KEY,
 state text NOT NULL DEFAULT 'open' CHECK (state IN ('open','acknowledged','resolved')),
 event_count bigint NOT NULL DEFAULT 1,
 last_event_type text NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS incident_outbox (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 event_id text NOT NULL UNIQUE REFERENCES incident_inbox(event_id),
 correlation_id text NOT NULL REFERENCES incidents(correlation_id),
 payload jsonb NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(),
 published_at timestamptz
);
CREATE INDEX IF NOT EXISTS incident_outbox_pending ON incident_outbox(id) WHERE published_at IS NULL;
