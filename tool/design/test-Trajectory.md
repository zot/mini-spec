# Test Design: Trajectory
**Source:** crc-Trajectory.md

Both alarms below target properties that **pass by default**. A parser that read only
the pending file, or that scanned whole done entries, would satisfy any test written
without these cases — which is exactly why they are written with them.

## Test: max spans both files
**Purpose:** the maximum is taken across the pending *and* done files, not either alone (R190)
**Input:** a pending file whose highest item is `## 3.` and a done file whose highest is `` (`#9`) ``; then the mirror case, pending `## 9.` and done `` (`#3`) ``
**Expected:** both directions answer 10
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
**Fire alarm:** make `MaxItemID` return the pending file's maximum alone. The first case goes red with `next=4, want 10`; the mirror case still passes, which is the point — one direction alone cannot detect it
**Inject:** internal/parser/trajectory.go:MaxItemID
**Pulled:** 2026-08-14 — rang, `next = 4, want 10`, and only the done-file-highest subtest failed. Restored byte-clean

## Test: a citation in an entry body does not raise the maximum
**Purpose:** only a done entry's first line is scanned, so prose citing another item cannot inflate the answer (R190)
**Input:** a done file with one entry `` - **2026-08-14 — thing (`#4`).** `` whose body prose cites `` `#99` `` on a later line
**Expected:** 5, not 100
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
**Fire alarm:** scan every line of the entry rather than the first. Goes red with `next=100, want 5`. This is not hypothetical — this project's own `#8` ledger entry cites `` `#7` `` in its body
**Inject:** internal/parser/trajectory.go:parseDoneIDs
**Pulled:** 2026-08-14 — rang, `next = 100, want 5`. Restored byte-clean

## Test: no files at all has no answer
**Purpose:** a project with no trajectory layer is reported as unanswerable rather than told the next ID is 1 (R194)
**Input:** a repository root containing neither file
**Expected:** an error naming both missing files; no number printed
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go

## Test: one file missing still answers, and says so
**Purpose:** a partial read is usable but must be qualified (R195, R197)
**Input:** a done file with `` (`#6`) `` and no pending file
**Expected:** 7, with the report naming `PENDING.md` as unread
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go

## Test: `next-id item` needs no design root
**Purpose:** the wiring property a parser test cannot reach — the queue is repository-scoped, so the subcommand must answer in a tree with no design root at all (R191, R198)
**Input:** a temp directory with `.git/`, a pending and a done file, and deliberately no `design/` anywhere
**Expected:** exit 0 and an answer
**Refs:** crc-CLI.md, crc-Trajectory.md
**Code:** internal/cli/cli_next_id_test.go
**Fire alarm:** delete the early-dispatch branch at the top of `runQuery`, so the subcommand falls through to `getProject()`. Goes red with `no design/ directory found` and exit 1. This is a **regression test for a real defect**, found by running the command rather than by reading it: the version first written failed in this very repository, whose design roots are `tool/` and `example/` while the queue sits above both
**Inject:** internal/cli/cli.go:runQuery
**Pulled:** 2026-08-14 — rang, `no design/ directory found`, exit 1. Restored byte-clean

## Test: per-file counts accompany the answer
**Purpose:** the number carries its evidence, so a silent parse failure is visible (R197)
**Input:** a pending file with two items and a done file with three entries
**Expected:** the report states 2 for the pending file and 3 for the done file
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
