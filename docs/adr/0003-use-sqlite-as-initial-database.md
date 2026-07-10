# ADR-0003: Use SQLite as the Initial Database

## Status

Accepted

## Context

NAS Audit will be developed locally and should be easy to run on different development machines without requiring a local database server.

SQLite does not require a background service. The application can create and open a local database file directly, which keeps early development simple and reduces operational overhead.

The project may later need PostgreSQL or another server database for shared development, hosted deployments, or higher-concurrency production environments. That decision should remain open.

## Decision

NAS Audit will use SQLite as the initial database backend.

The database location must be configurable. The first configuration shape will use:

```yaml
database:
  driver: sqlite
  path: ./var/nas-audit.db
```

The SQLite database file must not be stored inside a monitored root.

## Consequences

Positive:

- No local database service is required for development.
- The first runnable version can stay small and easy to test.
- Backups and local reset workflows are straightforward.
- The project can focus first on scan correctness and safety.

Negative:

- A single SQLite database file is not a good shared remote development database.
- Future multi-node or high-concurrency deployments may require PostgreSQL.
- Database access code should avoid hardcoding assumptions that make a future backend change unnecessarily difficult.

## Implementation Notes

- Ignore SQLite database files and WAL/SHM files in Git.
- Use transactions and batch writes for scan results.
- Use SQLite's backup mechanism rather than copying a live database file.
- Keep database configuration explicit so a future PostgreSQL backend can be introduced deliberately.
