# Roadmap

NAS Audit will be developed incrementally, with each phase producing a usable and reviewable slice of the system.

## Phase 0 - Project Foundation

Goal: make the project understandable, safe, and easy to evolve.

Deliverables:

- README with goals and non-goals.
- Architecture Decision Records in `docs/adr/`.
- Initial roadmap.
- GitHub milestones and issue backlog.
- Basic repository structure.

## Phase 1 - Inventory MVP

Goal: scan configured filesystem roots and store inventory metadata safely.

Deliverables:

- Initial configuration format for monitored roots.
- SQLite schema and migrations.
- Scan lifecycle states: pending, running, completed, failed, cancelled.
- Recursive filesystem traversal.
- Exclusions.
- Error recording.
- Basic totals: files, folders, bytes.

Acceptance test:

A complete scan finishes without excessive RAM usage and produces totals comparable to operating-system file listing tools.

## Phase 2 - Reliable Incremental Scans

Goal: detect changes without false deletion reports.

Deliverables:

- Added detection.
- Modified detection.
- Missing detection only after successful scans.
- Failed-scan protection.
- Atomic summary reports.
- SQLite backup procedure.

Acceptance test:

Disconnecting or unmounting a monitored source during a scan does not produce a mass-deletion report.

## Phase 3 - Web UI and Git Reports

Goal: make results reviewable by humans.

Deliverables:

- Dashboard.
- Browse/search page.
- Changes page.
- Scan history page.
- Compact Markdown and CSV reports.
- Optional automatic Git commits for successful reports.

## Phase 4 - Duplicate Candidates

Goal: identify likely duplicate files and folders without modifying data.

Deliverables:

- Size-based candidate grouping.
- Quick hashes.
- Full hashes for candidates.
- Duplicate file reports.
- Duplicate folder fingerprints.
- Reclaimable-space estimates.

## Phase 5 - Activity and Audit Correlation

Goal: correlate inventory changes with optional source-specific activity events.

Deliverables:

- First audit/activity importer.
- Activity page.
- Correlation between inventory changes and activity events.
- Confidence levels for uncertain attribution.

## Suggested Initial GitHub Milestones

- Phase 0 - Project Foundation
- Phase 1 - Inventory MVP
- Phase 2 - Reliable Incremental Scans
- Phase 3 - Web UI and Git Reports
- Phase 4 - Duplicate Candidates
- Phase 5 - Activity and Audit Correlation

## Suggested Initial Issues

- Write initial README.
- Add ADR-0001 for observation-only architecture.
- Add ADR-0002 for Go, SQLite, and server-rendered HTML.
- Define initial configuration format.
- Create SQLite schema and migrations.
- Implement scan lifecycle.
- Implement filesystem traversal.
- Add exclusions.
- Add failed-scan protection.
- Generate atomic summary report.
- Build basic dashboard.
