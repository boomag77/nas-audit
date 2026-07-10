# NAS Audit

NAS Audit is a read-only filesystem inventory, change tracking, duplicate detection, and audit reporting service.

It monitors configured filesystem roots. A root can be a local directory, a mounted SMB or NFS share, an external disk, a Synology share, or any other path visible to the host operating system.

The first production version is strictly observation-only: it must never modify, move, rename, delete, quarantine, deduplicate, or rewrite monitored files.

## Goals

- Inventory files and directories across configured roots.
- Detect added, missing, modified, moved, and renamed files.
- Preserve scan history in SQLite.
- Avoid false deletion reports after failed or interrupted scans.
- Calculate folder sizes and file counts.
- Identify duplicate file and duplicate folder candidates.
- Import optional activity/audit events from source-specific adapters.
- Provide a lightweight web interface.
- Store compact reports and operational history in Git.

## Non-Goals

NAS Audit does not modify monitored data.

The observation-only release will not include:

- automatic deletion;
- quarantine;
- automatic deduplication;
- hard-link creation;
- file rewriting;
- permission changes;
- automatic restore;
- Git tracking of source documents or full inventory databases.

## Planned Stack

- Go backend and scanner.
- SQLite database stored on local disk.
- Server-rendered HTML web UI.
- systemd services and timers for Linux deployments.
- Git for source code, configuration history, compact reports, and documentation.

## Safety Model

Configured roots should be exposed read-only whenever possible. For network shares, prefer both a read-only mount and an upstream account that cannot write to the source.

A file may be marked missing only after a complete successful scan. Failed, interrupted, or cancelled scans must never produce mass-missing or deletion reports.

## Project Status

Early planning and project setup.

See [docs/roadmap.md](docs/roadmap.md) and [NAS_Audit_Git_Project_Plan.md](NAS_Audit_Git_Project_Plan.md) for the current direction.
