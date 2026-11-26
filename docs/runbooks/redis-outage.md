# Redis admission outage drill

Against a disposable Compose stack, run `bash tests/redis-outage-smoke.sh`. The test stops Redis, asserts that event ingestion fails closed with HTTP 503 while liveness remains available, then restarts Redis and checks ingestion recovery. An EXIT trap attempts to restart Redis after failure. Existing incident state lives in PostgreSQL, not Redis. Restart resets the local request budget; no Redis persistence or HA is configured. The drill is not a benchmark or a production availability guarantee.
