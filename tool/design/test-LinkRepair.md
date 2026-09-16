# Test Design: LinkRepair
**Source:** crc-LinkRepair.md

## Test: the four relocations, one each, and the two refusals
**Purpose:** R464, R465, R466
**Input:** a temporary root with `carves/live.md`, `carves/done/moved.md`, `tool/x.md`, `carves/done/old.md`; `moved.md` links `../tool/x.md` (outgoing, written from `carves/`) and `done/old.md` (outgoing, `done/` now redundant); `live.md` links `moved.md` (incoming, target moved into `done/`) and `done/back.md` where `carves/back.md` exists (incoming, target moved out); plus a link to `nowhere.md` and a link whose two relocations both exist
**Expected:** five rewritten — `../../tool/x.md#sec`, `old.md`, `<old.md#w>` (the wrap and the fragment kept), `done/moved.md`, `back.md`; `nowhere.md` left `unresolvable`; the double one left `ambiguous`; every rewritten link's surrounding bytes unchanged; fragments kept
**Refs:** crc-LinkRepair.md, seq-links.md#3.3
**Code:** internal/update/links_test.go
**Alarm:** 1
**Fire alarm:** take the first candidate that resolves instead of requiring exactly one. Red: the ambiguous link is rewritten.
**Inject:** internal/update/links.go:candidates
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the candidate list truncated to its first entry: `live.md`'s ambiguous link was rewritten to `done/sub/both.md`; restore checksummed clean. The same day a probe past the list — dropping the angle-bracket wrap in `newDest` — found nothing asserting it, so the fixture gained `[w](<done/old.md#w>)` and its expectation `<old.md#w>` before this record

## Test: a pointer is repaired by the same rule
**Purpose:** R500
**Input:** a ledger `DONE.md` whose Part pointer names carves/moved.md#1 after `carves/done/moved.md` exists, and a pointer to nowhere
**Expected:** the pointer reads `carves/done/moved.md#1`, the key kept; the other is left `unresolvable`
**Refs:** crc-LinkRepair.md, seq-links.md#3.3
**Code:** internal/update/links_test.go
**Alarm:** 4
**Fire alarm:** skip pointers in the repair. Red: the ledger's pointer is left where it was, `rewritten 0`.
**Inject:** internal/update/links.go:RepairLinks
**Pulled:** 2026-09-16 — rang, by hand after the simplification pass, with pointers skipped: the ledger still named carves/moved.md in its Part pointer, rewritten 0; restore checksummed clean

## Test: only missing links are considered, and a second run is a no-op
**Purpose:** R464, R468
**Input:** the same root after `Repair`; and a document whose only links are tracked, external and local
**Expected:** the second run considers only the two links still missing — the rewritten ones resolve now — with none rewritten and no file's mtime or bytes changed; the clean document is reported with zero considered and is not written
**Refs:** crc-LinkRepair.md, seq-links.md#3.2
**Code:** internal/update/links_test.go
**Alarm:** 2
**Fire alarm:** consider every link rather than only the missing ones, so every link enters the candidate search. Red: the second run considers every link in the tree, and the clean document's tracked, external and local links are all listed as considered.
**Inject:** internal/update/links.go:RepairLinks
**Pulled:** 2026-09-16 — re-pulled by delegation at `ccea59e` after `#91` added the pointer rule to `RepairLinks`; rang on two tests with every link considered: ten considered where seven are expected, the clean document's tracked link `unresolvable` and its local link retargeted; restore clean *Earlier —* 2026-09-15 — rang, by hand, at `RepairLinks` — the site the 2026-09-15 pull was made at; the design had named it `Repair` and the census could not resolve that: with the missing-only filter removed, `TestOnlyMissingLinksAndASecondRunIsANoOp` red on every link in the tree being considered; restore checksummed clean *Pulled at `internal/update/links.go:Repair` — 2026-09-15 — rang, by hand after the simplification pass, with the missing-only filter removed: the second run listed every link in the tree as considered, and the clean document's tracked link came back `unresolvable`; restore checksummed clean — and the site has since moved, so this is history rather than a record.*
## Test: the default population, the report and the exit status
**Purpose:** R489, R467, R468
**Input:** the root above as a repository with everything staged, run from a subdirectory with no arguments; then again
**Expected:** every tracked markdown file is the population (R489), `carves/live.md` and `carves/done/moved.md` among them; the report lists each link with `old → new` or its reason and closes with the counts; exit 1 on the first run because one was left; the second run considers only that one, rewrites nothing, writes nothing, and still exits 1
**Refs:** crc-LinkRepair.md, seq-links.md#3.1
**Code:** internal/cli/cli_repair_test.go
**Alarm:** 3
**Fire alarm:** exit 0 whenever something was rewritten. Red: the first run exits 0 with two links left.
**Inject:** internal/cli/cli.go:CLI.runRepairLinks
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, exiting 0 whenever a file was written: `first run exited 0; want 1 with a link left`; restore checksummed clean
