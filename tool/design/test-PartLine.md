# Test Design: PartLine and MarkerSpan
**Source:** crc-PartLine.md

## Test: the status block's lines read back
**Purpose:** R349, R351, R352, R353, R356, R360 — the five things the tool reads, from the committed fixture
**Input:** `internal/minispecsdom/testdata/trajectory-sample.md` through `PartLines`
**Expected:** five part lines; keys `1`, `2`, `2.1`, `2.2`, `4`; checkboxes true, none, true, false, false; `Item 1`'s marker verb `LANDED`, queue ID 3, attribution containing the commit; `Item 2` struck false with verb `SPLIT`; `2.2` queue ID 8; the document renders back byte-exact
**Refs:** crc-PartLine.md, seq-partline.md#1
**Code:** internal/minispecsdom/partline_test.go
**Alarm:** 1
**Fire alarm:** classify every bold run as a marker, never a head. Red: every key is empty and the titles are nil.
**Inject:** internal/minispecsdom/partline.go:PartLine.Parse
**Pulled:** 2026-09-04 — re-pulled at `ddcf09f` by delegation: rang, `keys ",,,,"` as prescribed, and the strike test crashed on its nil head as before, aborting the package run; restore clean. Previously 2026-09-03 — rang: `keys ",,,,"` — the assertion; the strike test then crashed on its own nil map entry, which is that test's shape under this injection and not a production path (a headless line is guarded, and now asserted).

## Test: strike is derived and the edit is hidden
**Purpose:** R354
**Input:** `Item 1` (struck) and `2.2` (not) from the fixture; `Strike(true)` on `2.2`, `Strike(false)` on `Item 1`
**Expected:** `IsStruck` reads true then false before the edits and false then true after; the renders gain and lose exactly `~~` … `~~` around the head; no other byte moves
**Refs:** crc-PartLine.md, seq-partline.md#2
**Code:** internal/minispecsdom/partline_test.go
**Alarm:** 2
**Fire alarm:** make `Strike(true)` wrap the whole line rather than the head's bold run. Red: the render puts `~~` before the marker span's closing bytes rather than after the head.
**Inject:** internal/minispecsdom/partline.go:PartLine.Strike
**Pulled:** 2026-09-03 — rang: `IsStruck did not follow Strike` — the whole-line wrap leaves the head's opener with `- ` before it, so the derivation reads false; caught one assertion earlier than the render check predicted.

## Test: deviations name the target
**Purpose:** R350, R355, R410 — an unkeyed line parses and reports
**Input:** `- [ ] **Part A — old scheme.** **OPEN (#8.)**`, `- [X] **Item 3 - hyphen.**`, `- [ ] **Item 5 — ok.** **open (soon.)**`, `- a plain bullet`, and `- [ ] **Item 6 — r.** **REVERTED (Bill)**`
**Expected:** five part lines, all parsed; deviations: key form; checkbox interior and separator; verb case and `OPEN` attribution; key form; `REVERTED` attribution — each carrying its target text; on the plain bullet `IsStruck` is false and `Strike(true)` changes nothing
**Refs:** crc-PartLine.md
**Code:** internal/minispecsdom/partline_test.go
**Alarm:** 3
**Fire alarm:** return false from `Parse` when the head does not key. Red: the first line is missing from the result and its checkbox is not counted.
**Inject:** internal/minispecsdom/partline.go:PartLine.Parse
**Pulled:** 2026-09-04 — re-pulled at `ddcf09f` by delegation: rang, `1 lines, want 4`, only that test; restore clean. Previously 2026-09-03 — rang: `1 lines, want 3` — only the keyed line survived.

## Test: a marker write is canonical and guarded
**Purpose:** R357
**Input:** `2.2`'s marker: `Set("landed", "` + "`abc`" + `, 2026-09-03 — ` + "`#8`" + `.")`; then `Set("BAD)", "x")`
**Expected:** the first renders `**LANDED (`abc`, 2026-09-03 — `#8`.)**` and `QueueID` reads 8; the second is refused and the render unchanged
**Refs:** crc-MarkerSpan.md, seq-partline.md#3
**Code:** internal/minispecsdom/partline_test.go
**Alarm:** 4
**Fire alarm:** drop the re-parse in `Set`. Red: the second write is accepted and the marker's verb reads `BAD)`.
**Inject:** internal/minispecsdom/partline.go:MarkerSpan.Set
**Pulled:** 2026-09-03 — rang: `a verb with a parenthesis was accepted` and `a refused write changed the document`, only that test.

## Test: flexible on input, rigid on output
**Purpose:** R397, R398
**Input:** `OPEN (…)` with `#3.`, `#3`, `not queued.`, `not queued`, `Not  queued.`, `NOT QUEUED`; then `**OPEN, not queued.**` and `**OPEN (soon)**`; then `Set("OPEN", "#4.")` over `**open (Not queued)**`
**Expected:** the six loose spellings report no deviation; the comma form reports `marker scheme` and the off-shape attribution `OPEN attribution`, each with a target; the write renders `**OPEN (#4.)**`
**Refs:** crc-PartLine.md, crc-MarkerSpan.md
**Code:** internal/minispecsdom/partline_test.go
**Alarm:** 5
**Fire alarm:** restore the exact match — `a != "not queued."` in place of `notQueuedRe`. Red: `not queued`, `Not  queued.` and `NOT QUEUED` report `OPEN attribution`. A second injection: drop the `commaSchemeRe` check from `Parse`. Red: the comma form reports nothing — the silent case.
**Inject:** internal/minispecsdom/partline.go:MarkerSpan.deviations, internal/minispecsdom/partline.go:PartLine.Parse
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 7 rewrote the site; rang: both injections — the exact match reported `OPEN attribution` on all three spellings, and without the comma-scheme check the comma form reported nothing — only that test each. Previously 2026-09-04 — both rang: the exact match reported `OPEN attribution` on `not queued`, `Not  queued.` and `NOT QUEUED` (the two stop-ful forms stayed clean, as the old rule allowed); the dropped scheme check reported `[]` on the comma form; only that test each time. Both re-pulled the same day after the simplification pass reordered the operands and inlined `firstText`: rang again.

## Test: the key is the fragment
**Purpose:** R417 — one key form
**Input:** the fixture's keyed lines
**Expected:** keys `1`, `2.1`, `2.2`, `4`, and the stateless line's `2`; `Part("4")` resolves and `Part("Item 4")` does not
**Refs:** crc-PartLine.md, crc-CarveSdom.md
**Code:** internal/minispecsdom/carve_test.go
**Fire alarm:** return the bound text whole. Red: every key expectation in the carve and part-line tests reads `Item N`.
**Inject:** internal/minispecsdom/partline.go:PartLine.Key
**Pulled:** 2026-09-05 — rang in 9 tests across the carve and part-line files.
