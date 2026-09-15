# Test Design: FinishedCarve
**Source:** crc-FinishedCarve.md

## Test: a finished carve moves and every link follows, both directions
**Purpose:** R469, R471, R472, R474
**Input:** a root with `carves/x.md` (status block, every part landed) linking `../tool/a.md#s`, `done/old.md`, `other.md` and `nowhere.md`; `carves/other.md` linking `x.md#4`; `carves/done/old.md` linking `../x.md`; `PENDING.md` linking `carves/x.md`; `carves/done/x-twin.md` that must not be touched, and `carves/done/nowhere.md`, a file the carve never pointed at
**Expected:** `carves/done/x.md` exists with `../../tool/a.md#s`, `old.md`, `../other.md`, `nowhere.md` (left and reported, not retargeted at the twin); `carves/x.md` is gone; `other.md` reads `done/x.md#4`, `old.md` reads `x.md`, `PENDING.md` reads `carves/done/x.md`; every other byte of every file unchanged; the twin untouched; counts rewritten 6, left 1
**Refs:** crc-FinishedCarve.md, seq-links.md#4.3
**Code:** internal/update/finish_test.go
**Alarm:** 1
**Fire alarm:** compute the outgoing rewrite by the repair's search after the move rather than from the current target, with `carves/done/nowhere.md` present. Red: `nowhere.md` is retargeted to a file the carve never pointed at.
**Inject:** internal/update/finish.go:outgoingPlan
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the left link re-searched from the old location: `nowhere.md` was rewritten at the twin `carves/done/nowhere.md` and the report read rewritten 7, left 0; restore checksummed clean

## Test: the refusals leave every file as it was
**Purpose:** R469, R470, R473
**Input:** a carve with an open part; one with no status block; a path under `done/`; a path outside any carve directory; and a destination already present
**Expected:** each call returns an error naming the reason — the open part by its key — and a byte-for-byte comparison of the whole tree before and after shows nothing changed, including no file at the destination
**Refs:** crc-FinishedCarve.md, seq-links.md#4.2
**Code:** internal/update/finish_test.go
**Alarm:** 2
**Fire alarm:** check the destination after the rename instead of before. Red: `os.Rename` overwrites the file already at `done/`, the twin's bytes are gone, and the tree comparison names it.
**Inject:** internal/update/finish.go:FinishCarve
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the destination check moved after the write: `a refusal created or removed a file: 11 files before, 10 after` — the twin at `done/` overwritten and the old file removed; restore checksummed clean. *First attempt did not distinguish:* the colliding fixture carve had no status block, so it was refused on that before either ordering reached the destination, and red came from the message rather than the tree; the fixture gained a landed status block before this record

## Test: the CLI verb, the report and the design-root independence
**Purpose:** R474, R475
**Input:** the root above with a `.git` marker, run from a subdirectory with no design root
**Expected:** exit 0, the report names the move, each rewrite as `file:line old → new`, the left link, and closes with the counts; a second run exits 1 because the carve is no longer in `carves/`
**Refs:** crc-FinishedCarve.md, seq-links.md#4.6
**Code:** internal/cli/cli_finish_test.go
**Alarm:** 3
**Fire alarm:** resolve the carve through `getProject()` like the other update verbs. Red: the run refuses with no design root.
**Inject:** internal/cli/cli.go:runUpdate
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the dispatch moved into the project-bound switch: `first run exited 1; want 0 with no design root`; restore checksummed clean. The same day a probe past the list — dropping the `MkdirAll` of `done/` — rang on this test too, whose fixture is the only one without a `done/` directory
