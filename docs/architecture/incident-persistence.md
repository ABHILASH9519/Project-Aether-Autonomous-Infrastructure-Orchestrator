# Incident persistence — Phase 3

Status: initial storage implementation; integration verification pending.

```mermaid
flowchart LR
 Kafka --> Consumer
 Consumer --> TX[PostgreSQL transaction]
 TX --> Inbox[Unique event ID inbox]
 TX --> Incidents[Incident projection]
 TX --> Outbox[Pending outbox]
 TX -->|commit succeeds| Offset[Kafka offset commit]
 Outbox -. future relay .-> Kafka
```

The inbox, incident update and outbox insertion are atomic. A replay after database commit but before Kafka offset commit is ignored by the unique inbox key. The current projection counts events by correlation ID; it does not yet implement incident state transitions. Outbox publishing is not implemented. Event IDs must be globally unique; conflicting reuse of an ID is currently treated as a duplicate. PostgreSQL is intended for local development; authentication and operational hardening remain pending.
