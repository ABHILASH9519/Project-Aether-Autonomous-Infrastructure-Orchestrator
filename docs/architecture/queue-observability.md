# Durable queue observability

The workflow metrics endpoint includes a bounded, one-second SQL snapshot of pending outbox records, oldest unpublished age, pending remediation jobs and historical failed remediation jobs. The collection success gauge distinguishes an unavailable database from an empty queue. Failed collection omits queue metrics rather than publishing zeros. Metrics contain no event or incident identifiers.

Prometheus evaluates alerts for outbox stagnation, failed queue collection and workflow scrape failure. The Grafana HTTP dashboard also contains queue depth and oldest-age panels. The local alert rules do not send notifications because no Alertmanager is configured. Counts are database-wide; when scraping multiple workflow replicas use max, not sum. Full table counts may become expensive at larger scale; retention and preaggregated metrics remain future work. CI verifies metric exposition and loaded rule names.
