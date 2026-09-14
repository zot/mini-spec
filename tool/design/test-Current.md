# Test Design: Current
**Source:** crc-Current.md

## Test: the fixture's regions read back
**Purpose:** R385, R386, R387
**Input:** `internal/minispecsdom/testdata/current-sample.md`: a preamble and rule, `## Active` holding an item with a `###` sub-heading, two standing `##` sections
**Expected:** `Active` is the item text through its sub-heading; `Occupied`; `Standing` lists the two sections; the render is byte-exact
**Refs:** crc-Current.md, seq-current.md#1
**Code:** internal/minispecsdom/current_test.go
**Alarm:** 1
**Fire alarm:** end the region at any heading (drop the level test and the variable). Red: `Active` stops before the `###` sub-heading.
**Inject:** internal/minispecsdom/current.go:Current.regionEnd
**Pulled:** 2026-09-03 — rang: `Active` cut short of the `###` sub-heading. The delegate's injection left the heading variable unused and did not build; re-pulled by hand with it dropped.

## Test: a write reaches the region and nothing else
**Purpose:** R388, R389, R390 — the 2026-08-18 incident, as a property
**Input:** `SetActive` on the occupied fixture; then `Reset`; then `SetActive("new item")`
**Expected:** the first is refused and the render unchanged; after `Reset` the region holds the placeholder and every byte outside the region — the preamble, the rule, both standing sections — is identical to the fixture; after `SetActive` the region holds the new item and the outside is still identical; `Occupied` follows each step
**Refs:** crc-Current.md, seq-current.md#2
**Code:** internal/minispecsdom/current_test.go
**Alarm:** 2
**Fire alarm:** bound the region at the end of the file instead of the next `##`. Red: `Reset` deletes the standing sections — the incident, reproduced.
**Inject:** internal/minispecsdom/current.go:Current.regionEnd
**Pulled:** 2026-09-03 — rang, and it is the incident: `after Reset` shows both standing sections gone; `Active` also read to the end of the file.

## Test: exactly one Active
**Purpose:** R385, R406
**Input:** a document with no `## Active`; one with two; one whose only `## Active` is inside a fence
**Expected:** all three refused by `ParseCurrent`; the first and third with `ErrNoActive`, the second with `ErrManyActive`, by `errors.Is`
**Refs:** crc-Current.md, seq-current.md#1.2
**Code:** internal/minispecsdom/current_test.go
**Alarm:** 3
**Fire alarm:** take the first `Active` heading and ignore a second. Red: the two-heading document parses. For R406: return a fresh `errors.New` with the same message in place of `ErrManyActive`. Red: `errors.Is` fails on the two-heading document while the message is unchanged — which is the whole point of a sentinel.
**Inject:** internal/minispecsdom/current.go:Current.parse
**Pulled:** 2026-09-05 — the R406 injection rang: `want minispecsdom: more than one …` on the two-heading document, only that case. Same day, earlier: re-pulled by a delegated puller after `parse` gained the unclosed report; rang: the two-heading document `parsed`, only `TestExactlyOneActive`. Previously 2026-09-03 — rang: the two-heading document parsed, only that test. Site is `parse`, the helper.

## Test: a group open at end of input is unread
**Purpose:** R405
**Input:** a current file whose standing section ends in a code span that never closes
**Expected:** `Unread` holds one item at the span's line naming the marker; the Active region still reads
**Refs:** crc-Current.md
**Code:** internal/minispecsdom/current_test.go

## Test: SetActive and Reset read back
**Purpose:** R416 — through the guard, by the existing write tests
**Input:** the region-write test's writes
**Expected:** silent
**Refs:** crc-Current.md
**Code:** internal/minispecsdom/current_test.go
**Fire alarm:** have `write` append a stray line to the body. Red: `Current.SetActive … did not read back`.
**Inject:** internal/minispecsdom/current.go:Current.write
**Pulled:** 2026-09-05 — pulled again after the simplifier restructured the guards, rang again; first: rang.
