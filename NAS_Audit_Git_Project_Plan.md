# NAS Audit Project Plan

## 1. Project goal

Build a reliable, read-only service for monitoring configured filesystem roots.

A monitored root can be:

- a local directory;
- a mounted SMB share;
- a mounted NFS share;
- an external disk;
- a Synology share;
- any other path visible to the host operating system.

The service must:

- inventory files and folders;
- detect added, missing, modified, moved, and renamed files;
- calculate folder sizes and file counts;
- identify duplicate file and duplicate folder candidates;
- import activity/audit events when available;
- show results in a lightweight web interface;
- store compact change reports and configuration history in Git;
- never modify, move, rename, quarantine, deduplicate, or delete monitored files.

The first production version is strictly **observation only**.

---

## 2. Core safety principle

The monitoring system must have no ability to modify monitored content.

Preferred protection exists at two levels:

1. The monitored source is mounted or exposed read-only to Linux when possible.
2. The upstream account, share, or filesystem permissions deny write operations when possible.

For network shares, even if the Linux mount is accidentally changed to read-write, the upstream account should still be unable to create, modify, rename, or delete files.

---

## 3. High-level architecture

```text
Configured filesystem roots
    |
    | local paths, mounted SMB/NFS shares, external disks
    | optional audit/activity event sources
    v
Linux machine
    |
    +-- single Go service
    +-- SQLite inventory database
    +-- lightweight web interface
    +-- scheduled scans
    +-- report generator
    +-- Git repository
    +-- local logs and database backups
```

The monitored source files remain in their original locations.

Git stores source code, configuration, schema migrations, compact reports, and operational documentation. Git does not store the source documents or the full daily inventory.

---

## 4. Recommended platform

### Linux machine

Minimum practical specification:

- 2 CPU cores;
- 8 GB RAM;
- 100 GB or larger local SSD;
- 1 GbE network;
- reliable wired connection;
- UPS strongly recommended.

Preferred operating system:

```text
Debian Stable
```

Ubuntu Server LTS is also acceptable.

### Application stack

```text
Backend and scanner: Go
Database: SQLite
Web UI: server-rendered HTML
Service management: systemd
Scheduling: systemd timers
Version control: Git
Source access: local filesystem paths and mounted network shares
```

Avoid Docker, Redis, PostgreSQL, Kubernetes, and a separate JavaScript frontend in the first version. The goal is reliability and low resource usage.

---

## 5. Source preparation

Configure one or more monitored roots.

Example configuration:

```yaml
roots:
  - name: archive
    path: /mnt/archive
    type: filesystem
    read_only_required: true
    exclusions:
      - "@eaDir"
      - "#recycle"
      - ".snapshot"

  - name: local-documents
    path: /srv/documents
    type: filesystem
    read_only_required: false
```

For a network share, create a dedicated read-only account when possible.

Example account name:

```text
nas-audit
```

Required permissions:

```text
Read: Allow
Write: Deny
Create: Deny
Delete: Deny
Rename: Deny
Administration: No
Interactive SSH login: No
```

Do not use:

- a Synology administrator account;
- a domain administrator;
- a shared user account;
- credentials used by employees.

Recommended exclusions for Synology-like shares:

```text
@eaDir
@tmp
@database
@docker
@appstore
#recycle
.snapshot
```

---

## 6. Read-only source mounts

For local directories, prefer operating-system permissions that prevent the service account from writing to the monitored root.

For SMB/NFS/network sources, prefer read-only mounts and read-only upstream accounts.

Example SMB mount:

Create the mount point:

```bash
sudo mkdir -p /mnt/archive
```

Create a protected credentials file:

```bash
sudo install -m 600 /dev/null /etc/nas-audit.credentials
sudo nano /etc/nas-audit.credentials
```

Example:

```ini
username=nas-audit
password=REPLACE_WITH_PASSWORD
domain=WORKGROUP
```

Example `/etc/fstab` entry:

```fstab
//SERVER/archive /mnt/archive cifs credentials=/etc/nas-audit.credentials,ro,vers=3.1.1,noserverino,nofail,x-systemd.automount,_netdev 0 0
```

Test:

```bash
sudo mount -a
findmnt /mnt/archive
touch /mnt/archive/write-test
```

The `touch` command must fail with a permission error.

---

## 7. Local directory layout

```text
/opt/nas-audit/
├── nas-audit
├── config.yaml
└── VERSION

/var/lib/nas-audit/
├── inventory.db
├── backups/
└── state/

/var/log/nas-audit/
├── service.log
├── scan.log
└── audit-import.log

/srv/nas-audit-git/
├── reports/
├── config/
├── docs/
├── migrations/
└── scripts/
```

The SQLite database must be stored on the Linux machine's local SSD, not inside a monitored source root.

---

## 8. Inventory database

Recommended SQLite settings:

```sql
PRAGMA journal_mode=WAL;
PRAGMA synchronous=FULL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

Core tables:

```text
files
directories
scans
file_changes
folder_changes
hash_jobs
duplicate_groups
audit_events
service_events
```

Minimum file metadata:

```text
path
parent_path
name
extension
size
mtime
ctime
inode
device
first_seen_scan
last_seen_scan
missing_since_scan
quick_hash
content_hash
scan_error
```

The database must preserve history. Missing files are marked as missing; records are not immediately deleted.

---

## 9. Scan reliability rules

Every scan has a state:

```text
pending
running
completed
failed
cancelled
```

Critical rule:

A file may be marked missing only after a complete successful scan.

Correct workflow:

```text
1. Create scan record with status=running.
2. Walk the entire configured scope.
3. Update last_seen_scan for discovered entries.
4. Finish database transaction batches.
5. Validate scan completeness.
6. Set scan status=completed.
7. Only now mark unseen entries as missing.
```

If the scan is interrupted:

```text
status=failed
no files are marked missing
no deletion report is generated
```

This prevents network outages and mount failures from appearing as mass deletions.

---

## 10. Change detection

Initial comparison uses metadata:

```text
path
size
mtime
```

Change types:

```text
added
missing
modified
moved
renamed
unchanged
scan_error
```

A possible move or rename is detected when:

```text
old path disappears
new path appears
size matches
content hash matches
```

The service must clearly label unverified matches as:

```text
possible move
possible rename
```

Only hash-confirmed matches become verified.

---

## 11. Hashing strategy

Do not hash all 37 TB on every scan.

Use staged hashing.

### Stage 1 — metadata

Group candidates by:

```text
size
```

Files with unique sizes cannot be exact duplicates.

### Stage 2 — quick hash

For candidates, hash samples such as:

```text
first block
middle block
last block
file size
```

Recommended algorithm:

```text
BLAKE3
```

### Stage 3 — full hash

Read the complete file only when both match:

```text
size
quick hash
```

### Stage 4 — optional byte verification

For critical duplicate confirmation, perform a full byte-for-byte comparison.

No hash result should trigger an automatic delete action.

---

## 12. Duplicate folder detection

A folder fingerprint is built from sorted child entries.

File entry:

```text
F|filename|size|content_hash
```

Directory entry:

```text
D|directory_name|directory_hash
```

The folder hash is calculated bottom-up.

Two folders are exact duplicate candidates when these values match:

```text
folder hash
file count
directory count
total bytes
```

The web interface and reports must display candidates only. The service does not remove or consolidate folders.

---

## 13. Web interface

The web UI should remain lightweight and read-only.

Required pages:

### Dashboard

Show:

```text
last successful scan
current scan status
total files
total folders
total logical size
scan duration
scan errors
possible duplicate space
```

### Browse

Allow browsing and searching by:

```text
path
filename
extension
size
date
status
```

### Changes

Show:

```text
added
missing
modified
possible moved
verified moved
scan errors
```

### Duplicate files

Show:

```text
group
hash state
file size
copy count
potentially reclaimable bytes
all paths
```

### Duplicate folders

Show:

```text
folder hash
copy count
file count
total size
all paths
```

### Activity

Show imported audit events:

```text
time
username
source IP
protocol
operation
path
```

### Scans

Show:

```text
scan ID
start time
end time
status
files visited
folders visited
errors
bytes hashed
```

No delete, rename, move, quarantine, or write buttons are permitted in the observation-only release.

---

## 14. User activity auditing

Audit/activity events are optional and source-specific.

Possible event sources:

```text
Synology SMB Transfer Log
Samba logs
Linux auditd
remote syslog
custom CSV/JSON imports
future filesystem event adapters
```

For high-volume systems, avoid importing normal read events initially because millions of reads can generate excessive logs.

Prefer separate user accounts or identity-aware access systems. Shared accounts prevent reliable user attribution.

Recommended log flow:

```text
audit/activity source
    |
    | syslog, file import, API, or adapter
    v
Linux machine
    |
    v
audit event importer
    |
    v
SQLite audit_events table
```

Inventory changes answer:

```text
What changed?
```

Audit/activity events answer:

```text
Who performed the action?
When?
From which IP or host?
Through which protocol or source?
```

The service should correlate inventory changes with activity events but must display confidence levels when attribution is uncertain.

---

## 15. Git integration

Git is used as an audit and configuration history layer.

Git stores:

```text
application source code
configuration templates
database migrations
systemd units
operational documentation
daily summary reports
compact change reports
duplicate candidate reports
audit import summaries
```

Git does not store:

```text
TIFF or PDF source files
inventory.db
SQLite WAL files
full 21-million-file manifests
temporary scan data
raw high-volume logs
credentials
```

Recommended `.gitignore`:

```gitignore
*.db
*.sqlite
*.sqlite-wal
*.sqlite-shm
*.log
tmp/
state/
backups/
secrets/
.env
```

---

## 16. Git report structure

```text
reports/
└── 2026/
    └── 07/
        └── 2026-07-10/
            ├── summary.md
            ├── added.csv
            ├── missing.csv
            ├── modified.csv
            ├── moved.csv
            ├── duplicate-files-summary.csv
            ├── duplicate-folders-summary.csv
            ├── scan-errors.csv
            └── audit-summary.csv
```

Large reports should be compressed or summarized before committing.

The full data remains queryable in SQLite.

Example commit:

```bash
git add reports config docs migrations scripts
git commit -m "Filesystem audit report 2026-07-10"
```

If no tracked report changed, the automation should exit successfully without creating an empty commit.

---

## 17. Report generation rules

Reports must be written atomically.

Correct pattern:

```text
summary.md.tmp
    |
    | successful write and fsync
    v
rename to summary.md
```

A report must include:

```text
scan ID
scan start and end time
scan status
scope
total files
total folders
total bytes
added count
missing count
modified count
move candidates
duplicate candidates
errors
database integrity status
```

Reports from failed or incomplete scans must be clearly marked and must not claim files were deleted.

---

## 18. Database backup

Create a consistent SQLite backup using SQLite's backup mechanism:

```bash
sqlite3 /var/lib/nas-audit/inventory.db   ".backup '/var/lib/nas-audit/backups/inventory-$(date +%F).db'"
```

Recommended retention:

```text
daily: 14 copies
weekly: 8 copies
monthly: 12 copies
```

At least one backup copy must be stored outside the old Linux machine.

Possible destinations:

```text
separate protected Synology backup share
external USB disk
another Linux machine
offline backup
```

Do not back up the live SQLite database with a simple `cp` while it is being written.

---

## 19. Git repository backup

The Git repository should have at least one remote.

Possible targets:

```text
internal Git server
bare repository on a protected NAS share
another Linux machine
private Git hosting, if company policy permits
```

Git is not the backup of the SQLite database. It is the history of code, configuration, and compact reports.

---

## 20. systemd services

Recommended units:

```text
nas-audit.service
nas-audit-scan.service
nas-audit-scan.timer
nas-audit-backup.service
nas-audit-backup.timer
nas-audit-report.service
nas-audit-report.timer
nas-audit-git.service
```

Suggested schedule:

```text
incremental metadata scan: nightly
duplicate candidate hashing: weekends
database backup: daily
Git report commit: after successful report generation
integrity check: weekly
```

Do not start a new scan if a previous scan is still running.

Use a lock file or systemd unit dependency to prevent overlap.

---

## 21. Security requirements

### Monitored source access

- dedicated read-only accounts for network sources when possible;
- credentials readable only by root;
- read-only mounts where supported;
- no write permissions for monitored roots where practical;
- no administrator/root credentials for routine scanning.

### Web interface

Initial deployment:

```text
listen only on the local network
HTTPS through Caddy or Nginx
authentication required
no public Internet exposure
```

Roles for future use:

```text
Viewer
Auditor
Administrator
```

In the observation-only release, even an administrator cannot modify monitored source data from the application.

### Linux host

- automatic security updates;
- SSH keys only;
- firewall enabled;
- service runs as an unprivileged user;
- local SSD health monitored;
- UPS and graceful shutdown configured.

---

## 22. Logging

Record:

```text
service startup and shutdown
scan start and completion
scan failure reason
mount availability
database errors
database integrity checks
files skipped
permission errors
hash errors
audit import status
report creation
Git commit result
```

Logs must not contain:

```text
source credentials
session secrets
private keys
full credential files
```

Use log rotation to prevent disk exhaustion.

---

## 23. Failure scenarios

### Monitored source unavailable

Expected behavior:

```text
scan fails
no files are marked missing
failure is recorded
web UI shows degraded state
no normal change report is committed
```

### Network interruption

Expected behavior:

```text
current scan fails
partial results remain associated with failed scan
previous successful inventory remains authoritative
```

### Linux machine loses power

Expected behavior:

```text
SQLite WAL recovery occurs
running scan becomes failed or abandoned
no mass missing state is committed
```

### Local disk full

Expected behavior:

```text
scanner stops safely
database remains valid
report generation stops
alert appears in UI and logs
```

### Corrupt database

Expected behavior:

```text
service enters read-only degraded mode
latest backup is preserved
automatic destructive repair is forbidden
```

---

## 24. Development phases

## Phase 0 — Infrastructure and safety

Deliverables:

- Linux installed;
- local SSD checked;
- UPS configured where appropriate;
- initial monitored root selected;
- read-only access tested where supported;
- project Git repository initialized;
- backup destination defined.

Acceptance test:

```text
The service account can read the configured root and cannot modify it when read-only mode is required.
```

## Phase 1 — Inventory MVP

Deliverables:

- recursive filesystem traversal;
- SQLite schema;
- scan tracking;
- total files, folders, and bytes;
- exclusions;
- error recording;
- command-line progress;
- basic web dashboard.

Acceptance test:

```text
A complete scan finishes without exhausting RAM and produces totals comparable to operating-system file listing tools.
```

## Phase 2 — Reliable incremental scans

Deliverables:

- added detection;
- modified detection;
- missing detection only after successful scans;
- interrupted-scan handling;
- atomic reports;
- daily database backups.

Acceptance test:

```text
Disconnecting or unmounting a monitored source during a scan does not produce a mass-deletion report.
```

## Phase 3 — Git reporting

Deliverables:

- summary Markdown reports;
- compact CSV change reports;
- automatic Git commits;
- Git remote backup;
- credentials excluded from Git.

Acceptance test:

```text
Each successful scan produces a reviewable Git commit without storing the source documents or full inventory database.
```

## Phase 4 — Duplicate candidate detection

Deliverables:

- grouping by file size;
- quick hashes;
- full hashes for candidates;
- duplicate file reports;
- duplicate folder fingerprints;
- potential reclaimable-space estimates.

Acceptance test:

```text
Known duplicate test folders are found, and no files are changed.
```

## Phase 5 — Audit correlation

Deliverables:

- at least one audit/activity event adapter implemented;
- remote syslog or file import configured for the first adapter;
- audit importer;
- user, host/IP, action, path, source, and time displayed;
- correlation between inventory changes and audit events.

Acceptance test:

```text
A test rename and deletion from a supported event source appears in the activity page and is correlated with the next inventory scan.
```

## Phase 6 — Operational hardening

Deliverables:

- weekly SQLite integrity checks;
- log rotation;
- health page;
- disk-space alerts;
- backup restore procedure;
- disaster recovery documentation;
- service update procedure.

Acceptance test:

```text
The database can be restored on another Linux machine and the web interface can display the last known inventory.
```

---

## 25. Explicitly out of scope

The observation-only project must not include:

- automatic deletion;
- quarantine;
- automatic deduplication;
- hard-link creation;
- file rewriting;
- TIFF or PDF recompression;
- automatic permission changes;
- automatic restore;
- monitored filesystem modifications;
- Git tracking of source documents;
- Git LFS or git-annex for the entire archive.

These capabilities require a separate future project and separate security review.

---

## 26. Initial implementation order

Recommended first sprint:

```text
1. Prepare Linux development/target machine.
2. Select an initial monitored root.
3. Configure and verify read-only access where required.
4. Initialize Git repository.
5. Create SQLite schema.
6. Implement scan lifecycle.
7. Implement filesystem traversal.
8. Add exclusions and error handling.
9. Build basic dashboard.
10. Add atomic summary report.
11. Add SQLite backup.
12. Add automatic Git commit after a successful scan.
```

Second sprint:

```text
1. Incremental change detection.
2. Failed-scan protection.
3. Search and browse UI.
4. Change reports.
5. Audit log import.
```

Third sprint:

```text
1. Duplicate file candidates.
2. Duplicate folder candidates.
3. Hash scheduling.
4. Reclaimable-space reports.
```

---

## 27. Definition of done

The observation-only release is complete when:

- configured read-only roots are verified as read-only where required;
- the service runs on the target Linux machine;
- a full scan can complete without excessive RAM use;
- incomplete scans cannot produce false deletion reports;
- additions, modifications, missing files, and moves can be reported;
- duplicate files and folders are reported as candidates only;
- the web interface contains no write operations for monitored sources;
- at least one audit/activity adapter can identify user or process, host/IP where available, action, and path;
- SQLite is backed up and restore-tested;
- compact reports are committed to Git;
- the Git repository has an independent remote copy;
- operational and recovery procedures are documented.
