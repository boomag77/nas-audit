# ADR-0001: Observation-Only Architecture

## Status

Accepted

## Context

NAS Audit is intended to monitor important filesystem trees and produce inventory, change, duplicate, and audit reports. These roots may be local directories, mounted network shares, external disks, or NAS shares.

Because the monitored data may be large, sensitive, or operationally important, the first production version must not be able to modify it.

## Decision

NAS Audit will be observation-only.

The service must not create, edit, move, rename, delete, quarantine, deduplicate, chmod, chown, rewrite, or otherwise modify monitored files or directories.

The web interface and API must not expose write operations for monitored roots.

When possible, monitored roots should be protected outside the application as well:

- read-only mounts;
- read-only network share accounts;
- operating-system permissions that deny writes to the service account.

## Consequences

Positive:

- The system can be trusted as an audit and reporting tool.
- Bugs in the application should not be able to damage monitored data.
- Duplicate detection can be introduced safely as candidate reporting only.

Negative:

- Cleanup, restore, permission repair, and deduplication workflows require separate tooling or a future project.
- Users may need external procedures to act on reports.

## Safety Invariants

- Failed, interrupted, or cancelled scans must not mark files as missing.
- A file may be marked missing only after a complete successful scan.
- Reports from incomplete scans must not claim files were deleted.
- Duplicate reports are candidates only and must never trigger automatic changes.
