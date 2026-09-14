# Test Design: TestDoc
**Source:** crc-TestDoc.md

## Test: the fixture reads as entries with fields
**Purpose:** R418, R419, R420, R421, R422, R424
**Input:** `internal/minispecsdom/testdata/testdoc-sample.md`
**Expected:** four entries; the first titled and at line 4 with a folded `Fire alarm`, one site, a `Pulled` dated `2026-09-06`, one `Code` file and `Alarm` 3; `Alarm(3)` is that entry and `Alarm(0)`, `Alarm(99)` nil; the second is not an alarm; the third has two sites and no `Pulled`; the fourth is `Alarm` 1 and its fenced `**Alarm:** 99` names nothing; `Unread` is the `Notes` heading alone; the render is byte-exact
**Refs:** crc-TestDoc.md, seq-testdoc.md#1
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 1
**Fire alarm:** drop the code-group test so every `**Name:**` at a line head is a field. Red: alarm 99 appears, and the fourth entry's `Alarm` reads 99 or is doubled.
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.readFields
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `the fenced fields were read as fields: alarm=99`, three doubled-field deviations, and NumberAlarms assigning 100; restore clean. *Earlier —* 2026-09-07 — re-pulled by hand after the code-group test moved into `markdownDoc`; rang with the same signature. *Earlier —* 2026-09-06 — rang, by hand: `the fenced fields were read as fields: alarm=99`, `Alarm(n) does not resolve as expected`, the fenced `**Fire alarm:**` listed as a doubled field, and `NumberAlarms` assigned `[100]`; restore byte-clean
## Test: doubled and malformed fields are deviations
**Purpose:** R422, R423, R424
**Input:** an entry with two `**Alarm:**` lines and a `**Pulled:**` with no date; an entry whose `**Alarm:**` is `x`
**Expected:** the first `Alarm` is read and two deviations are listed, `Pulled` is nil; the second entry has `Alarm` 0 and one deviation; three unread; `SetPulled` on the first refuses with a `DeviationError`, on an absent number with `ErrNoAlarm`
**Refs:** crc-TestDoc.md, seq-testdoc.md#1.4.2
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 2
**Fire alarm:** let a repeated field overwrite the first rather than deviate. Red: the entry reads `Alarm` 2 and the write over it succeeds.
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.derive
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `first read 2` with only the Pulled-date deviation left; restore clean. *Earlier —* 2026-09-06 — rang, by hand: `first read 2 with [Pulled leads with a date]` — the second `**Alarm:**` overwrote the first and the doubling went unlisted; restore byte-clean
## Test: SetPulled replaces with history, or inserts after Inject or Fire alarm
**Purpose:** R425, R426
**Input:** the fixture's alarm 3 → `SetPulled(3, "2026-09-07", …)`; after `NumberAlarms`, the third entry's first pull; a bare entry with `Fire alarm` and `Alarm` only
**Expected:** one `**Pulled:**` line, dated anew with the old content after ` *Earlier —* `, the file longer by exactly the new prefix and every other entry intact; the first pull sits between `**Inject:**` and `**Refs:**`; with no `Inject`, directly after the folded `Fire alarm` and before `**Alarm:**`
**Refs:** crc-TestDoc.md, seq-testdoc.md#2.2.1
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 3
**Fire alarm:** apply the later span before the earlier one when `void` demotes `Pulled`, so the earlier span's end boundary lands where parsed nodes were already removed. Red: `SetInject` returns *no parsed node at offset* from the boundary guard, where before the guard it silently ran the removal to the end of the file and took the entries after alarm 3 with it. (Sited on the sort, not on `replaceSpan`: the two boundaries inside one span may be resolved in either order and stay green.)
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.SetInject
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `no parsed node at offset 751; a span was already replaced there` from the boundary guard; restore clean. *Earlier —* 2026-09-06 — rang, by hand, on the re-sited injection: `no parsed node at offset 751; a span was already replaced there`; restore byte-clean
## Test: SetInject rewrites, and demotes Pulled on void
**Purpose:** R427
**Input:** alarm 3 → two sites without `void`; the same with `void`; an empty site list; an entry with no `Inject`
**Expected:** the sites joined by `, ` with `Pulled` standing; with `void` the `Pulled` line becomes the history sentence naming `internal/alarm/alarm.go:assessOne` and the entry's `Pulled` is nil; `ErrEmptyInject`; `ErrNoInject`
**Refs:** crc-TestDoc.md, seq-testdoc.md#2.2.2
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 4
**Fire alarm:** fold the list fields like the prose ones. Red: the demoted line reads as `Inject` sites and the read-back panics.
**Inject:** internal/minispecsdom/testdoc.go:folds
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang through the guard: `TestDoc.SetInject on "3" did not read back`, the demoted line folded into the Inject sites; restore clean. *Earlier —* 2026-09-06 — rang, by hand: `SetInject on "3" did not read back: want a.go:A, b.go:B` — the demoted line folded into the sites; restore byte-clean
## Test: NumberAlarms numbers above Fire alarm from max+1, once
**Purpose:** R428
**Input:** the fixture, twice
**Expected:** `[4]` assigned above the third entry's `**Fire alarm:**`, the file longer by that line alone, 1 and 3 unchanged; the second run assigns nothing and changes no byte; the entry with no `Fire alarm` is never numbered
**Refs:** crc-TestDoc.md, seq-testdoc.md#2.2.3
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 5
**Fire alarm:** count from the number of entries rather than the highest number present. Red: the assigned number collides with 3 or is not 4.
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.NumberAlarms
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `assigned [5], want [4]`; restore clean. *Earlier —* 2026-09-06 — rang, by hand: `assigned [5], want [4]`; restore byte-clean
## Test: a title with a code span or bold reads whole
**Purpose:** R446 — mini-spec's *the pre-`track` refusal asks the intent and stops* read as *the pre-*; 25 of its headings carry a code span
**Input:** the fixture, whose headings now hold a code span and a bold run
**Expected:** both titles read whole, markers included, as source
**Refs:** crc-TestDoc.md, seq-testdoc.md#1.2
**Code:** internal/minispecsdom/testdoc_test.go
**Alarm:** 6
**Fire alarm:** cut the title back to the first text node after the marker (`headingText(nodes[i+1])`). Red: no entry titled with the span or the bold; the titles stop at the pre- and at through.
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.scan

**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: no entry titled with the span or the bold, both titles stopped at their marker; restore clean. *Earlier —* 2026-09-12 — rang: this test alone, both titles missing (`the title stopped at a marker`); injection by hand in the live tree, restored by reverse edit, diff checked.
