# Test Design: Carve
**Source:** crc-Carve.md

Fixtures carry real shapes wherever one exists. The reader's own rules are tested in the
dependency; what is tested here is what this tool adds — the scan, the counts, the write
path's atomicity and refusals, and the rendering.

## Test: a fenced status example is not data
**Purpose:** validates R209 — a document about the format quotes a status block in a fence,
and a line-oriented scan counted it (ark's `carves/README.md`, 14 open reported where 13 existed)
**Input:** a document whose only `## Status` and only checkbox lines sit inside a fence
**Expected:** `HasStatus` false, no parts
**Fire alarm:** none here — the property is the dependency's, proven by its `carve_test.go`;
this test pins that the adapter does not re-scan by line
**Refs:** crc-Carve.md — R209
**Code:** internal/parser/carve_test.go

## Test: subparts are counted, and the status block bounds the count
**Purpose:** validates R209, R210 — the grep reported 9 open where 12 existed and missed every
subpart; a whole-document count reads open questions as parts
**Input:** a status block with a top-level part, a `SPLIT` parent, two subparts, then a
`## Open questions` section with two more checkboxes
**Expected:** 2 open, 1 landed, 1 stateless; the open questions contribute nothing
**Fire alarm:** none here — the status region and the subpart depth are the dependency's,
proven by its `carve_test.go`; this test pins that the adapter takes `Parts()` and
`Stateless()` as handed and re-scans nothing
**Refs:** crc-Carve.md — R209, R210, R216
**Code:** internal/parser/carve_test.go

## Test: a document with no status block is reported, never dropped
**Purpose:** validates R211
**Input:** a carve directory holding one document with a status block and one without
**Expected:** two carves in the scan, one with `HasStatus` false; the census names one document
with no status block
**Fire alarm:** skip a document with no status block in `ScanCarves` and confirm this goes red
with one carve
**Inject:** internal/parser/carve.go:ScanCarves
**Pulled:** 2026-09-04 — rang: `carves=1 noStatus=0 withStatus=1; want 2, 1, 1`; restore byte-clean by copy
**Refs:** crc-Carve.md — R211
**Code:** internal/parser/carve_test.go

## Test: no carve directory at all has no answer
**Purpose:** validates R214 — an empty census over no directory is a confident wrong one
**Input:** a repository root with neither `carves/` nor `.carves/`
**Expected:** `ErrNoCarveDirs`
**Fire alarm:** return an empty scan and nil when both are absent and confirm this goes red
**Inject:** internal/parser/carve.go:ScanCarves
**Pulled:** 2026-09-04 — rang: `got <nil>; want ErrNoCarveDirs`; restore byte-clean by copy
**Refs:** crc-Carve.md — R214
**Code:** internal/parser/carve_test.go

## Test: a stateless line carries its line number and reason
**Purpose:** validates R216 — the dependency numbers parts, not stateless lines, so the adapter
derives the number from the node's offset; a wrong derivation points a reader at the wrong line
**Input:** a status block whose third line is a `SPLIT` parent
**Expected:** one stateless entry, its `Line` the parent's 1-based line, its reason naming the
missing checkbox
**Fire alarm:** report the offset instead of the line and confirm this goes red
**Inject:** internal/parser/carve.go:lineOf
**Pulled:** 2026-09-04 — rang: `stateless = L85 … want L6`, and the CLI listing test went red beside it on
the `L8` column; restore byte-clean by copy
**Refs:** crc-Carve.md — R216
**Code:** internal/parser/carve_test.go

## Test: a marker write is atomic and a refusal leaves the file byte-identical
**Purpose:** validates R220 — the file is rewritten by rename, and none of the reader's four
refusals reaches the file
**Input:** a carve on disk with a conforming open part and a landed part; `SetMarker` on the
open part; then `SetMarker("OPEN", …)` on the landed part, `SetPartLanded` on the landed part,
and `SetMarker` on a key no part carries
**Expected:** the first write changes exactly the marker and the file re-reads as one part
landed and one open; the three refusals return `ErrReopen`, `ErrLanded`, `ErrNoPart` and the
file's bytes are unchanged after each
**Fire alarm:** on a refusal, create the temp file and leave it — confirm the leftover-file
assertion goes red. *The first injection written here could not ring:* writing the rendered
bytes before checking the reader's error left the file byte-identical anyway, because the
reader does not mutate the document when it refuses — so "byte-identical on refusal" has two
independent guards, and only the cleanup half is this adapter's to prove
**Inject:** internal/parser/carve.go:editCarve
**Pulled:** 2026-09-04 — rang: `temp files left behind: 4 entries in the directory`; restore byte-clean by copy.
The write-before-check injection was pulled first the same day and stayed green, recorded above
**Refs:** crc-Carve.md — R220
**Code:** internal/parser/carve_test.go

## Test: the command answers with no design root
**Purpose:** validates R213 — this repository's queue sits above two design roots
**Input:** a repository root with `carves/` and no `design/` anywhere
**Expected:** exit 0 and a census
**Fire alarm:** route `carves` through `getProject()` like the other subcommands and confirm
this goes red with `no design/ directory found`
**Inject:** internal/cli/cli.go:runQuery
**Pulled:** 2026-09-04 — rang: `runQuery exited 1 in a tree with no design root; want 0`; restore byte-clean by copy
**Refs:** crc-CLI.md — R213
**Code:** internal/cli/cli_carves_test.go

## Test: the rendered line and the listing rule
**Purpose:** validates R207, R212, R216, R217, R218 — a non-conforming part is listed without
`--open`, a clean open part and a clean stateless line only with it, the census states zeros
and says `stateless`
**Input:** one carve with an open conforming part, a landed part, a part keyed on a superseded
scheme (`#4`), and a `SPLIT` parent
**Expected:** without `--open` the only part row is the non-conforming one and no stateless
row; with `--open` the open part and the stateless row appear; the census reads
`1 carve: 1 open, 1 landed, 1 stateless, 1 non-conforming; 0 documents with no status block`
**Fire alarm:** list a part only when `--open` is given and confirm the no-flag case goes red
with no rows
**Inject:** internal/cli/cli.go:listed
**Pulled:** 2026-09-04 — rang: `without --open, want only the non-conforming row` — no rows at all;
restore byte-clean by copy
**Refs:** crc-CLI.md — R207, R212, R216, R217, R218
**Code:** internal/cli/cli_carves_test.go

## Test: this repository's carves stay conformant
**Purpose:** validates R219 against the live corpus — the reader's rules over this
repository's own carves, zero deviations
**Input:** `carves/` of this repository
**Expected:** every part conforms and every stateless line carries no deviation
**Fire alarm:** none — a corpus check, not a guard; it goes red when a carve is edited badly,
which is its purpose
**Refs:** crc-Carve.md — R219
**Code:** internal/parser/carve_test.go
