# ADR-0002: Go, SQLite, and Server-Rendered HTML

## Status

Accepted

## Context

NAS Audit needs to scan large filesystem trees, persist inventory history, run reliably on a modest Linux machine, and provide a lightweight web interface.

The first version should minimize operational complexity and avoid unnecessary services.

## Decision

The first production version will use:

- Go for the scanner, backend, and web service;
- SQLite for inventory and scan history;
- server-rendered HTML for the web interface;
- systemd services and timers for Linux deployment and scheduling;
- Git for source code, documentation, configuration templates, and compact reports.

The first version will avoid Docker, Redis, PostgreSQL, Kubernetes, and a separate JavaScript frontend.

## Consequences

Positive:

- Small deployment footprint.
- Simple backup and restore story.
- Good performance for filesystem walking and concurrent hashing work.
- Fewer moving parts for a service that should be reliable and boring.

Negative:

- SQLite requires careful transaction and backup handling.
- A future multi-node or high-concurrency deployment may require revisiting the database choice.
- Server-rendered HTML limits highly interactive UI patterns, which is acceptable for the first version.

## Implementation Notes

SQLite must live on local storage, not inside a monitored root. Backups must use SQLite's backup mechanism rather than copying a live database file.
