# Test Design: Trajectory
**Source:** crc-Trajectory.md

The three alarms below target properties that **pass by default**. A parser that read
only the pending file, that scanned whole done entries, or that scanned a whole entry
header rather than its identifier slot, would satisfy any test written without these
cases — which is exactly why they are written with them.

*Every fixture here is in the done-entry shape adopted 2026-08-16. The shapes drawn from
ark's ledger are marked; they are there because a fixture contains only what its author
thought to include, and the author of the shape this replaced had read no corpus at all.*

## Test: max spans both files
**Purpose:** the maximum is taken across the pending *and* done files, not either alone (R190)
**Input:** a pending file whose highest item is `## 3.` and a done file whose highest entry leads `— #9:`; then the mirror case, pending `## 9.` and done `— #3:`
**Expected:** both directions answer 10
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
**Alarm:** 1
**Fire alarm:** make `MaxItemID` return the pending file's maximum alone. The first case goes red with `next=4, want 10`; the mirror case still passes, which is the point — one direction alone cannot detect it
**Inject:** internal/parser/trajectory.go:MaxItemID
**Pulled:** 2026-08-16 — rang, `next = 4, want 10`, and only the done-file-highest subtest failed. Restored byte-clean. *Re-pulled the same day after `parseDoneIDs` was rewritten for the adopted shape — same signature*

## Test: a citation in an entry body does not raise the maximum
**Purpose:** only a done entry's **header** is scanned, so prose quoting another item cannot inflate the answer (R190)
**Input:** a done file with one entry `` - **2026-08-14 — #4: a thing.** `` whose body quotes an older pending entry, `- 2026-07-06 — **PENDING #99 — the older shape: seeds and scope.**`, on a later line
**Expected:** 5, not 100
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
**Alarm:** 2
**Fire alarm:** drop the header guard in `parseDoneIDs` so every line is offered to the slot regex. Goes red with `next=100, want 5`. The body line is **ark's real shape** — an older pending entry quoted verbatim inside a later done entry — and measured 2026-08-16, five such lines sit in its ledger
**Inject:** internal/parser/trajectory.go:parseDoneIDs
**Pulled:** 2026-08-16 — rang, `next = 100, want 5`. Restored byte-clean. *This is the second pull: the first, earlier the same day, used a body-prose fixture that the rewritten parser would no longer have caught, so both the fixture and the injection are new*

## Test: only the identifier slot is read
**Purpose:** an entry's leading slot holds whatever it discharged — a queue ID, a gap ID, a requirement range, several separated by `/`, or nothing — so only the `#N`s in it count, and nothing after the title's colon does (R190)
**Input:** six done entries — two verbatim from ark (`— O201 / R3399:` and `— #117 / R3398:`), one discharging two items (`— #84 / #83:`), one whose `Part` pointer ends `` `carves/x.md#13` ``, one with no identifiers at all, and one whose title carries a colon followed by `#500`
**Expected:** exactly `[117 84 83 4]`, maximum 117
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
**Alarm:** 3
**Fire alarm:** scan the whole header line rather than the slot. Goes red with `DONE.md contributed [117 84 83 4 13 500], want [117 84 83 4]` and `max = 500, want 117` — 13 being a **part key** and 500 prose. The part pointer is the sharp one: the adopted format appends `` Part `<doc>#<key>` `` to every entry, spelled exactly like a queue ID, so this slot restriction is load-bearing against the format's own addition
**Inject:** internal/parser/trajectory.go:parseDoneIDs
**Pulled:** 2026-08-16 — rang, both assertions, message as recorded above. Restored byte-clean

## Test: no files at all has no answer
**Purpose:** a project with no trajectory layer is reported as unanswerable rather than told the next ID is 1 (R194)
**Input:** a repository root containing neither file
**Expected:** an error naming both missing files; no number printed
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go

## Test: one file missing still answers, and says so
**Purpose:** a partial read is usable but must be qualified (R195, R197)
**Input:** a done file with one entry leading `— #6:` and no pending file
**Expected:** 7, with the report naming `PENDING.md` as unread
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go

## Test: `next-id item` needs no design root
**Purpose:** the wiring property a parser test cannot reach — the queue is repository-scoped, so the subcommand must answer in a tree with no design root at all (R191, R198)
**Input:** a temp directory with `.git/`, a pending and a done file, and deliberately no `design/` anywhere
**Expected:** exit 0 and an answer
**Refs:** crc-CLI.md, crc-Trajectory.md
**Code:** internal/cli/cli_next_id_test.go
**Alarm:** 4
**Fire alarm:** delete the early-dispatch branch at the top of `runQuery`, so the subcommand falls through to `getProject()`. Goes red with `no design/ directory found` and exit 1. This is a **regression test for a real defect**, found by running the command rather than by reading it: the version first written failed in this very repository, whose design roots are `tool/` and `example/` while the queue sits above both
**Inject:** internal/cli/cli.go:runQuery
**Pulled:** 2026-09-16 — re-pulled by delegation at `ccea59e` after `runQuery` gained the `refs` dispatch for `#91`, same injection and signature: `runQuery exited 1 in a tree with no design root; want 0`; restore clean *Earlier —* 2026-09-15 — rang again after `runQuery` gained the `links` dispatch for `#83`, same injection and signature: `runQuery exited 1 in a tree with no design root; want 0`; restore checksummed clean *Earlier —* 2026-09-07 — rang again after the gaps case was extracted from `runQuery` for `#75`, same injection and signature; restore byte-clean by copy. Previously 2026-09-04 — rang again after the done-entry reader change staled it (delegated `alarm-puller`, evidence read by hand): `TestNextIDItemNeedsNoDesignRoot` exited 1 with `no design/ directory found`, restore byte-clean. First pulled 2026-08-16, same signature
## Test: per-file counts accompany the answer
**Purpose:** the number carries its evidence, so a silent parse failure is visible (R197)
**Input:** a pending file with two items and a done file with three entries
**Expected:** the report states 2 for the pending file and 3 for the done file
**Refs:** crc-Trajectory.md
**Code:** internal/parser/trajectory_test.go
