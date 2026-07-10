# NAS Audit Project Instructions

## Core Goal

Build a reliable, read-only service for monitoring a large Synology NAS.

The first production version is strictly observation-only.

## Absolute Safety Rules

- Never add functionality that modifies NAS data.
- Do not create, edit, move, rename, delete, quarantine, deduplicate, chmod, chown, or rewrite NAS files.
- Do not add UI buttons or API endpoints that can modify NAS content.
- NAS access must remain read-only at both levels:
  - the Linux SMB mount is read-only;
  - the Synology service account has read-only permissions.
- Even an application administrator must not be able to modify NAS data in the observation-only release.

## Scan Reliability Rules

- A file may be marked missing only after a complete successful scan.
- Failed, interrupted, cancelled, or incomplete scans must never produce deletion or mass-missing reports.
- If the NAS is unavailable or the network drops during a scan, the scan must fail safely and preserve the previous successful inventory as authoritative.
- Missing records should be marked as missing, not immediately deleted from history.

## Stack Choices

- Backend and scanner: Go.
- Database: SQLite on the local Linux SSD, not on the NAS share.
- Web UI: server-rendered HTML.
- Scheduling and service management: systemd services and timers.
- Version control: Git.
- NAS protocol: SMB 3.x.

Avoid Docker, Redis, PostgreSQL, Kubernetes, and a separate JavaScript frontend in the first version.

## Data and Git Rules

Git may store:

- source code;
- configuration templates;
- schema migrations;
- systemd units;
- operational documentation;
- compact reports.

Git must not store:

- TIFF, PDF, or other NAS source documents;
- SQLite database files;
- SQLite WAL/SHM files;
- full large inventory manifests;
- raw high-volume logs;
- credentials or secrets.

## Hashing and Duplicate Rules

- Do not hash the full NAS on every scan.
- Use staged hashing: metadata, quick hash, full hash only for candidates.
- Duplicate files and folders are candidates only.
- No hash result may trigger automatic deletion or consolidation.

## Reporting and Backups

- Reports must be written atomically.
- Reports from failed or incomplete scans must be clearly marked and must not claim files were deleted.
- SQLite backups must use SQLite's backup mechanism, not a plain copy of the live database.
- Git is not the SQLite database backup; it is only history for code, configuration, and compact reports.

## Development Priorities

Initial implementation order:

1. SQLite schema.
2. Scan lifecycle.
3. Filesystem traversal.
4. Exclusions and error handling.
5. Basic dashboard.
6. Atomic summary report.
7. SQLite backup.
8. Git commit after successful report generation.

## Working Rule for Codex

Do not edit, create, delete, rename, reformat, or otherwise change any project file unless the user explicitly asks for that specific change.
