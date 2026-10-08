# Test Design: Gaps
**Source:** crc-Gaps.md

## Test: the fixture reads as entries
**Purpose:** R429, R430, R431, R432
**Input:** `internal/minispecsdom/testdata/gaps-sample.md`; a document with no section; a document whose only `## Gaps` is fenced
**Expected:** seven gaps `A1,T1,I1,O1,O2,O3,O4`; `A1` permanent, unboxed, at line 13; `T1` with one folded sub-item; `I1` checked; `O1`'s text folded across three lines; `O2` with two sub-items; `O3`'s fenced `O99` is not a gap and its text stops at the colon; the `O5` after `## Notes` is not a gap; nothing unread; byte-exact render; no section and a fenced section both read empty
**Refs:** crc-Gaps.md, seq-gaps.md#1
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 1
**Fire alarm:** drop the code-group skip so a fenced bullet is read. Red: `O99` resolves as a gap and the ID list grows.
**Inject:** internal/minispecsdom/gaps.go:Gaps.readItems
**Pulled:** 2026-10-08 — rang, by hand at `#103`'s commit after `readItems` gained regions and the `###` break: dropping the code-group skip reads the fenced `O99` — `ids = A1,T1,I1,O1,O2,O3,O99,O4` in `TestGapsReadsTheFixture`; minispecsdom run unfiltered with `-count=1`; restore byte-clean *Earlier —* 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `O99` resolved as a gap and the ID list grew to `A1,T1,I1,O1,O2,O3,O99,O4`; restore clean. *Earlier —* 2026-09-07 — rang, by hand: `ids = A1,T1,I1,O1,O2,O3,O99,O4` and `O99` resolved as a gap; restore byte-clean
## Test: deviations, nesting, and a bare bullet
**Purpose:** R430, R432
**Input:** a section with a boxed `A1`, an unboxed `O2`, an `O3` with a nested `O4`, a second `O3`, and a bare `- reason:` bullet at column 0
**Expected:** one deviation each on `A1` and `O2`; `O4` at depth 2 with parent `O3` and no deviation; `Gap("O3")` is the first and the second carries a deviation; four unread ending with the bare bullet; `Resolve("A1")` refuses with a `DeviationError`; an absent ID is `ErrNoGap`
**Refs:** crc-Gaps.md, seq-gaps.md#1.3.3
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 2
**Fire alarm:** key a gap only at column 0, so an indented keyed bullet is a sub-item. Red: `O4` is not a gap and the ID list is short.
**Inject:** internal/minispecsdom/gaps.go:gapHeadRe
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `the nested O4 was not read as a gap`; restore clean. *Earlier —* 2026-09-07 — rang, by hand: `the nested O4 was not read as a gap`; restore byte-clean
## Test: Add appends after the last gap, or after the heading
**Purpose:** R436, R550
**Input:** `Add("O5", …)` then `Add("A2", …)` on the fixture; a taken ID; a bad ID; no section; a section with no entries
**Expected:** `O5` sits between `O4` and the blank before `## Notes`, the file longer by that line alone; `A2` follows it with no checkbox; `ErrGapExists`, `ErrBadGapID`, `ErrNoSection`; the empty section gains its first entry directly under the heading
**Refs:** crc-Gaps.md, seq-gaps.md#2.2
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 3
**Fire alarm:** insert at the region's end rather than after the last gap's span. Red: `O5` lands after the blank line, directly before `## Notes`, and the placement assertion fails.
**Inject:** internal/minispecsdom/gaps.go:Gaps.Add
**Pulled:** 2026-10-08 — rang, by delegation at the batch commit after `#101` routed `Add` through `insertLine`: inserting at the section end fails `TestGapsAdd` on placement, the permanent add and the empty section; minispecsdom run unfiltered with `-count=1`; restore clean *Earlier —* 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `O5` landed after the blank line, directly before `## Notes`, and the empty-section case failed with it; restore clean. *Earlier —* 2026-09-07 — rang, by hand: the placement and permanent-add assertions both failed, `O5` and `A2` sitting after the blank line; restore byte-clean
## Test: Add after an unterminated last line ends it first
**Purpose:** R540
**Input:** a Gaps section whose last entry ends the file with no newline; `Add("O2", …)`
**Expected:** the render is the source plus `\n- [ ] O2: …\n`, and `O2` reads back; no `ReadBackError`
**Refs:** crc-Gaps.md, seq-gaps.md#2.5
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 5
**Fire alarm:** restore the field defect — `Gaps.Add` inserting through `replaceSpan` rather than `insertLine`. Red: `Add` panics with `ReadBackError` (`O2` glued onto `O1`'s line, ui-engine 2026-10-07)
**Inject:** internal/minispecsdom/gaps.go:Gaps.Add
**Pulled:** 2026-10-08 — rang, by hand after simplification: `TestGapsAddAfterUnterminatedLine` panicked with `ReadBackError` (`O2` glued onto `O1`); package run unfiltered with `-count=1`; restore byte-clean. Past the list: inverting `insertLine`'s condition also fails `TestGapsAdd`

## Test: Resolve flips the head; Approve rewrites it permanent
**Purpose:** R434, R435, R436
**Input:** `Resolve("O1")` twice, `Resolve("A1")`; `Approve("O3", "A2")`; `Approve("O1", "A3")` on a fresh fixture; `Approve` with a non-A ID, a taken ID, a permanent target
**Expected:** `[x] O1` with the file the same length; `ErrResolved`; `ErrPermanent`; `- A2:` with `O3`'s head text and the fence beneath untouched, `O3` gone; `- A3:` with `O1`'s two continuation lines still beneath it and the file shorter by exactly `[ ] O1` less `A3`; `ErrBadGapID`, `ErrGapExists`, `ErrPermanent`
**Refs:** crc-Gaps.md, seq-gaps.md#2.3
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 4
**Fire alarm:** replace the whole body span on approve rather than the head line. Red: `O1`'s continuation lines are gone from the render. (`O3` cannot carry this alarm: a blank line follows its head, so its body is the head alone and the injection cannot reach it — found on the first pull, which stayed green.)
**Inject:** internal/minispecsdom/gaps.go:Gaps.Approve
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: the approved entry's continuation lines were gone from the render; restore clean. *Earlier —* 2026-09-07 — rang, by hand, on the re-targeted test: `approve of a wrapped entry` — `O1`'s continuation lines gone; the first pull against `O3` stayed green and is recorded in the alarm; restore byte-clean

## Test: Add places an entry by subsection
**Purpose:** R550, R551 — same letter first, then the heading that names the type, then the head
**Input:** a section with a head holding `T1`, then `### Incomplete Implementation` (prose only), `### Design → Code Gaps` holding `D1`, and `### Oversights (On)` holding `O1` with an `A1` filed beneath it; add `T2`, `O2`, `A2`, `I1`, `S1`
**Expected:** `T2` after `T1` at the head; `O2` after `O1`; `A2` after `A1` under Oversights; `I1` at the end of Incomplete Implementation; `S1` at the head; every subsection heading still on its own line
**Refs:** crc-Gaps.md, seq-gaps.md#2.2.1
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 6
**Fire alarm:** restore the field behaviour — always insert after the last gap in the section. Red: `T2` lands under Oversights (ark's ui-engine report, 2026-10-07)
**Inject:** internal/minispecsdom/gaps.go:Gaps.insertionPoint
**Pulled:** 2026-10-08 — rang, by hand after simplification: always placing after the last gap in the section fails `TestGapsAddPlacesBySubsection` on all three placements (`T2`, `I1` and the Oversights run); minispecsdom and update run unfiltered with `-count=1`; restore byte-clean

## Test: a subsection heading names a type in either form
**Purpose:** R550 — `(Xn)` or the standard name, arrows `→` or `->`, spaced or not, case ignored, a trailing word allowed
**Input:** headings `Oversights (On)`, `design -> code gaps`, `Spec→Requirements`, `Spec → Design Gaps`, `Notes`
**Expected:** O, D and S named; `Spec → Design Gaps` and `Notes` name nothing
**Refs:** crc-Gaps.md
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 7
**Fire alarm:** compare the title exactly rather than normalised. Red: `design -> code gaps` names nothing
**Inject:** internal/minispecsdom/gaps.go:namesType
**Pulled:** 2026-10-08 — rang, by hand after simplification: comparing titles unnormalised fails `TestASubsectionHeadingNamesAType` and the head and Incomplete Implementation placements in `TestGapsAddPlacesBySubsection`; restore byte-clean

## Test: a heading ends the entry above it
**Purpose:** R551 — a `###` line directly under a gap, with no blank line between, is not folded into its text
**Input:** `- [ ] O1: first` followed at once by `### Later` and `- [ ] O2: second`
**Expected:** `O1`'s text is `first`
**Refs:** crc-Gaps.md, seq-gaps.md#1.3.6
**Code:** internal/minispecsdom/gaps_test.go
**Alarm:** 8
**Fire alarm:** let a `###` line fall through to the fold. Red: `O1`'s text is `first ### Later`
**Inject:** internal/minispecsdom/gaps.go:Gaps.readItems
**Pulled:** 2026-10-08 — rang, by hand after simplification: with the `###` branch unreachable, `TestAHeadingEndsTheEntryAboveIt` reads O1 as `first ### Later`, and every placement in `TestGapsAddPlacesBySubsection` fails; restore byte-clean
