# Test Design: CarveSdom
**Source:** crc-CarveSdom.md

## Test: the fixture's status block reads as parts
**Purpose:** R361, R362, R363, R364, R365, R369
**Input:** `sdom/schema/testdata/trajectory-sample.md`
**Expected:** `HasStatus`; four parts keyed `1`, `2.1`, `2.2`, `4` with depths `0, 2, 2, 0` and parents nil, `1`, `1`, nil — the subparts follow `Item 1` because the split parent between them has no checkbox; one stateless line keyed `Item 2`; the body's `- [ ] inside a fence` is not a part; the render is byte-exact
**Refs:** crc-CarveSdom.md, seq-carve.md#1
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 1
**Fire alarm:** drop the region bound so every list item in the file is tried (consume `from` and `to` so it builds). Red: nothing here, since the fixture's other bullets are fenced — so the test also parses a source with a bullet under a later `## Notes`, which becomes a part under the injection; a `### Sub` inside the block does not end it.
**Inject:** internal/minispecsdom/carve.go:ParseCarve
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after the reader gained the unclosed report; rang: `a bullet outside the region became a part: "Item 1,Item 2,Item 9"`, only that test. Previously 2026-09-04 — re-pulled at `ddcf09f` by delegation: rang, `a bullet outside the region became a part: "Item 1,Item 2,Item 9"`, only that test; restore clean. Previously 2026-09-03 — rang: `a bullet outside the region became a part: "Item 1,Item 2,Item 9"`, only that test. The delegate's first attempt left `from`/`to` unused and did not build; re-pulled by hand with them consumed.


## Test: no status block, and a fenced one
**Purpose:** R362, R363
**Input:** a document with no `## Status`; a document whose only `## Status` is inside a fence
**Expected:** `HasStatus` false and no parts for both
**Refs:** crc-CarveSdom.md, seq-carve.md#1.2
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 2
**Fire alarm:** when no `Heading` node keys the block, fall back to searching the source text for `## Status` and treat a hit as the heading. Red: the fenced document reports a status block. (A first prescription — match each node's render by prefix — could not reach the property: a heading node renders only its `## `, the title being the sibling text, so that injection matched nothing anywhere and every status test failed for the wrong reason.)
**Inject:** internal/minispecsdom/carve.go:Carve.statusRegion
**Pulled:** 2026-09-03 — rang, on the rewritten injection: `no status block found` inverted — the fenced document reported one. The first prescription was unreachable; see the alarm.

## Test: SetMarker follows the tool's rule
**Purpose:** R366, R409
**Input:** `Item 4` (`OPEN (not queued.)`) → `SetMarker("OPEN", "#9.")`; a line carrying `**NOT VERIFIED.** **OPEN (#8.)**` → `SetMarker("LANDED", "x")`; a line with no marker → `SetMarker("DEFERRED", "Bill, 2026-09-03")`; a line carrying `**OPEN (#4.)** **OPEN (#5.)**` → `SetMarker("OPEN", "#6.")`; a line carrying `**REVERTED (#70.)** Needs Item 1.` → `SetMarker("OPEN", "#70.")`; a line with trailing prose and no marker → `SetMarker("OPEN", "#3.")`; an unknown key
**Expected:** `OPEN (#9.)` replaces in place; `NOT VERIFIED` stands and the `OPEN` becomes `LANDED (x)`; the bare line gains ` **DEFERRED (Bill, 2026-09-03)**`; the two-transient line becomes `**OPEN (#6.)**` alone; `REVERTED` is replaced, not joined — `**OPEN (#70.)** Needs Item 1.`; the prose line reads `**Item 5 — e.** **OPEN (#3.)** Needs Item 1.`, marker before prose; the unknown key errors and nothing changes
**Refs:** crc-CarveSdom.md, seq-carve.md#2.3
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 3
**Fire alarm:** treat only `OPEN` as transient again (`isTransient` = `isOpen`). Red: `**REVERTED (#70.)** … **OPEN (#70.)**`, both states on one line — mini-spec's 2026-09-05 smoke test. A second: append at the end of the line instead of before the prose. Red: `Needs Item 1. **OPEN (#3.)**`. The older injections: replace the first marker regardless of verb. Red: `NOT VERIFIED` is overwritten. A second injection: stop dropping the other transients (`drop` never appended). Red: `**OPEN (#6.)** **OPEN (#5.)**` — this case was added 2026-09-03 after a past-the-list probe; that probe's green was a mis-applied injection (a `sed` that matched nothing), so whether the earlier suite guarded the rule was never established — no earlier test carried two transients, which is the inference the case rests on. Re-run with the injection verified applied, it rang here.
**Inject:** internal/minispecsdom/partline.go:PartLine.SetMarker
**Pulled:** 2026-09-05 — the two new injections pulled again after the simplifier restructured `isTransient` and `insertMarker`: both rang, only this test. Earlier the same day: Item 7: the two new injections rang (`REVERTED` kept beside `OPEN`; the marker after the prose), and the two older ones re-pulled by hand against the rewritten site rang; each only in this test. Previously 2026-09-04 — re-pulled at `ddcf09f`: the second injection by delegation rang with `**OPEN (#6.)** **OPEN (#5.)**`; the first by hand rang at the same assertion; only that test each time; restore clean. Previously 2026-09-03 — rang: `NOT VERIFIED` overwritten by `LANDED (x)`, only that test.

## Test: Land is three markings in one act
**Purpose:** R368
**Input:** `2.2` → `Land("` + "`abc`" + `, 2026-09-03 — ` + "`#8`" + `.")`
**Expected:** the line renders `  - [x] ~~**2.2 — the config move.**~~ **LANDED (`abc`, 2026-09-03 — `#8`.)**`
**Refs:** crc-CarveSdom.md, seq-carve.md#2
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 4
**Fire alarm:** skip the strike in `Land`. Red: the render keeps the head unstruck while the box and marker changed — the three no longer agree.
**Inject:** internal/minispecsdom/carve.go:Carve.Land
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 10 added the read-back to the site; rang: the read-back itself caught it — `Carve.Land on "2.2" did not read back: want … checked and struck, got … struck false` — only that test. Previously 2026-09-04 — re-pulled at `ddcf09f` by delegation: rang, `2.2` landed with `[x]` and the record but the head unstruck, only that test; restore clean. Previously 2026-09-03 — rang: the line landed with `[x]` and the record but the head unstruck — the three no longer agree.

## Test: writes refuse over deviations
**Purpose:** R391, R392
**Input:** a status line `- [ ] **Item 1 — a.** **OPEN (#3)**` (the bare-`#N` scheme, a deviation) → `SetMarker("Item 1", "LANDED", "x")` and `Land("Item 1", "x")`
**Expected:** each returns a `DeviationError` with key `1` and one deviation, rule `OPEN attribution`; its message names the rule and its target; the render equals the source byte for byte
**Refs:** crc-CarveSdom.md, crc-PartLine.md, seq-carve.md#2.1.1
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 5
**Fire alarm:** make `refuse` return nil unconditionally. Red: both writes succeed and the line gains a canonical marker over the deviating one. A second injection, the past-the-list probe of 2026-09-04: move `SetChecked(true)` in `Land` above the refusals. Red: `Land changed a refused line` — the byte-identical assertion is what guards R391's ordering, and it also broke the older landing test with `already landed`.
**Inject:** internal/minispecsdom/partline.go:PartLine.refuse
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 7 rewrote the site; rang: `SetMarker over a deviating line: <nil>` and `Land over a deviating line: <nil>`, only that test (first injection only; the second is the older probe). Previously 2026-09-04 — rang: `SetMarker over a deviating line: <nil>` and `Land over a deviating line: <nil>`; only that test. The changed-line assertion did not print because the test `continue`s after a wrong error type.

## Test: OPEN never reopens a landed part
**Purpose:** R391, R393
**Input:** the fixture's `Item 1` (`[x]`, struck, `LANDED`) → `SetMarker("1", "open", "not queued.")`; then `SetMarker("1", "NOT VERIFIED", "Bill, 2026-09-04")`
**Expected:** the first is `ErrReopen` and the render equals the source; the second succeeds — the guard is on `OPEN`, not on every write over a landed part
**Refs:** crc-PartLine.md, seq-carve.md#2.3.1
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 6
**Fire alarm:** drop the `ErrReopen` guard from `SetMarker`. Red: `OPEN (not queued.)` is written beside the `LANDED` record on a checked, struck line.
**Inject:** internal/minispecsdom/partline.go:PartLine.SetMarker
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 7 rewrote the site; rang: `OPEN over a landed part: <nil>` and `a refused reopen changed the line`, with `OPEN (not queued.)` written beside the `LANDED` record, only that test. Previously 2026-09-04 — rang: `OPEN over a landed part: <nil>`, and the render showed `**LANDED (…)** **OPEN (not queued.)**` on the checked, struck line; only that test.

## Test: Land refuses over a landed part
**Purpose:** R391, R394
**Input:** the fixture's `Item 1` → `Land("1", "` + "`fff`" + `, 2026-09-04 — ` + "`#9`" + `.")`
**Expected:** `ErrLanded`, and the render equals the source — no second `LANDED` beside the first
**Refs:** crc-CarveSdom.md, seq-carve.md#2.1.2
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 7
**Fire alarm:** drop the `Checked()` guard from `Land`. Red: the write succeeds and the line carries two `LANDED` records.
**Inject:** internal/minispecsdom/carve.go:Carve.Land
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 10 added the read-back to the site; rang: `Land over a landed part: <nil>` and the line carrying two `LANDED` records, only that test. Previously 2026-09-04 — rang: `Land over a landed part: <nil>`, and the render showed two `LANDED` records side by side; only that test.

## Test: every part knows its line
**Purpose:** R395
**Input:** the fixture
**Expected:** `Item 1` line 7, `2.1` line 9, `2.2` line 10, `Item 4` line 11 — 1-based, as parsed
**Refs:** crc-CarveSdom.md
**Code:** internal/minispecsdom/carve_test.go
**Alarm:** 8
**Fire alarm:** build each `Part` without its line (`line` left zero). Red: every part reports 0.
**Inject:** internal/minispecsdom/carve.go:ParseCarve
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after `ParseCarve` gained the unclosed report; rang: `Item 1: line 0, want 7` and the three other parts, only that test. Previously 2026-09-04 — rang: `Item 1: line 0, want 7` and the other three parts; only that test.

## Test: a group open at end of input is unread
**Purpose:** R404
**Input:** a carve whose status block is followed by a code span that never closes
**Expected:** `Unread` holds one item at the span's line naming the marker; the parts still read
**Refs:** crc-CarveSdom.md
**Code:** internal/minispecsdom/carve_test.go

## Test: a closer that closes nothing is unread
**Purpose:** R411
**Input:** a carve whose prose holds a two-run span with a three-run inside
**Expected:** three unread items on that line — the span the three-run ended, never closed; the three-run, which closes nothing; and the span the trailing two-run opened, demoted at end of input and read as text (R445). Previously the trailing span was never closed and took the file's tail with ithing
**Refs:** crc-CarveSdom.md
**Code:** internal/minispecsdom/carve_test.go
**Fire alarm:** drop the unpaired closers from `unbalanced`. Red: two items, no `closes nothing`.
**Inject:** internal/minispecsdom/unread.go:unbalanced
**Pulled:** 2026-09-12 — re-pulled by delegate after the checkpoint commit; rang: this test alone, two items and no `closes nothing`. Previously 2026-09-07 — re-pulled after the demoted line joined the helper; rang: this test alone. Previously 2026-09-05 — rang: two items and no `closes nothing`, only that test.

## Test: SetMarker and Land read back
**Purpose:** R415 — through the guard, by the existing write tests
**Input:** the marker and landing tests' writes
**Expected:** silent
**Refs:** crc-CarveSdom.md
**Code:** internal/minispecsdom/carve_test.go
**Fire alarm:** in `Land`, skip `Strike(true)`. Red: `Carve.Land … want LANDED (…), checked and struck`. A second: in `PartLine.SetMarker`, write the attribution with a trailing `?`. Red: `Carve.SetMarker … did not read back`.
**Inject:** internal/minispecsdom/carve.go:Carve.Land, internal/minispecsdom/partline.go:PartLine.SetMarker
**Pulled:** 2026-09-05 — pulled again after the simplifier restructured the guards, rang again; first: both rang: the unstruck landing and the corrupted attribution each panicked through the guard.
