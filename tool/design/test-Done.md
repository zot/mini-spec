# Test Design: Done
**Source:** crc-Done.md

## Test: the fixture's entries read back
**Purpose:** R378, R379, R380, R381, R382, R384
**Input:** `internal/minispecsdom/testdata/done-sample.md`: a preamble and rule; an entry with `#8 / R189–R198` in the slot, a commit and a part pointer in the header, and an indented bullet in its body; an entry with an `O201` slot and a `#4` quoted in its body; an entry with no slot; a body holding a fenced quotation of an entry; an entry-like `- not an entry` bullet
**Expected:** three entries, the first's region keeping its indented bullet; IDs `[8]`, none, none; `HasSlot` true, true, false; the body's `#4` contributes nothing; part pointers as written, the second found in the body; `MaxID` 8; `Unread` 1; the render is byte-exact
**Refs:** crc-Done.md, seq-done.md#1
**Code:** internal/minispecsdom/done_test.go
**Alarm:** 1
**Fire alarm:** read every `#N` in the entry rather than only the slot. Red: the second entry gains ID 4 and `MaxID` still reads 8 — so the test asserts the IDs, not only the max. A second injection, in `regionEnd`: let any list item end a region, not only one at column 0. Red: entry 0 loses its indented bullet — added 2026-09-03 after a past-the-list probe found the column-0 condition unguarded.
**Inject:** internal/minispecsdom/done.go:DoneEntry.derive, internal/minispecsdom/done.go:Done.regionEnd
**Pulled:** 2026-09-03 — rang: `ids [8 1] [3 4] []` — the body's `#4` and even `#1` from a part pointer counted — and the prepend test's max moved too. The `regionEnd` injection was pulled by hand the same day: `the body's indented bullet ended entry 0's region`.

## Test: Prepend lands after the rule
**Purpose:** R383
**Input:** `Prepend` on the fixture; `Prepend` on a ledger with a rule and no entries
**Expected:** the new entry is the first, separated from the rule and from the old first entry by blank lines; on the empty ledger it follows the rule; `Entries` reflects it
**Refs:** crc-Done.md, seq-done.md#2
**Code:** internal/minispecsdom/done_test.go
**Alarm:** 2
**Fire alarm:** insert at the end instead of before the first entry. Red: the new entry is last.
**Inject:** internal/minispecsdom/done.go:Done.Prepend
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after Item 10 added the read-back to the site; rang: the read-back caught it — `Done.Prepend … did not read back`, the old first entry still first — only that test. Previously 2026-09-03 — rang: the new entry rendered last, after the entry-like bullet, only that test.

## Test: entries and entry-like bullets carry their line
**Purpose:** R395, R396
**Input:** the fixture (folded into "the fixture's entries read back")
**Expected:** entries at lines 7, 11, 14; `Unread` is one `{21, "- not an entry, but entry-like"}`
**Refs:** crc-Done.md, seq-done.md#1
**Code:** internal/minispecsdom/done_test.go
**Alarm:** 4
**Fire alarm:** record every line as 0 in `scan`. Red: the line assertions and the unread comparison both fail.
**Inject:** internal/minispecsdom/done.go:Done.scan
**Pulled:** 2026-09-05 — re-pulled by a delegated puller after the reader gained the unclosed report; rang: `Unread [{Line:0 Text:- not an entry, but entry-like}]` and `lines 0 0 0`, only the fixture read-back. Previously 2026-09-04 — rang: `Unread [{Line:0 Text:- not an entry, but entry-like}]` and `lines 0 0 0`; only that test. Same non-building first attempt as the pending alarm. Re-pulled the same day after the simplification pass hoisted `off`: the unread line alone was injected and rang on the unread assertion.

## Test: a group open at end of input is unread
**Purpose:** R402, R445 — the failure `Unread` exists to prevent, arriving one layer below it; since the base demotes, the tail reads too
**Input:** a done file whose entry is followed by a fence that never closes and then a second entry
**Expected:** both entries read; `Unread` holds one item at the fence's line whose text names the marker and says it is never closed, read as text
**Refs:** crc-Done.md
**Code:** internal/minispecsdom/done_test.go
**Fire alarm:** make `unbalanced` return nil. Red: this test and its three siblings in the pending, carve and current designs — one helper, four readers. A second: drop the demoted loop alone. Red: the same five, since every one now reads a demoted opener.
**Inject:** internal/minispecsdom/unread.go:unbalanced
**Alarm:** 5
**Pulled:** 2026-09-12 — re-pulled by delegate after the checkpoint commit; both injections rang the same five tests, `unread []`. Previously 2026-09-07 — both injections rang in the same five tests: this, its pending, carve and current siblings, and the carve's closes-nothing test. Previously 2026-09-05 — rang in exactly the four sibling tests; re-pulled the same day after the simplifier dropped the helper's `doc` parameter, rang in the same four. **Past the list, same day:** disabling the line sort that first followed this helper rang nothing, and on inspection could not — nothing structured follows a group still open at the end, so the unclosed lines are last in file order by construction. The sort was removed; the requirements say *after*, not *sorted*.

## Test: Prepend reads back
**Purpose:** R414 — through the guard, by the existing prepend test
**Input:** the prepend test's writes
**Expected:** silent
**Refs:** crc-Done.md
**Code:** internal/minispecsdom/done_test.go
**Fire alarm:** write `header + "x"` as the header line. Red: `Done.Prepend … did not read back`.
**Inject:** internal/minispecsdom/done.go:Done.Prepend
**Pulled:** 2026-09-05 — pulled again after the simplifier restructured the guards, rang again; first: rang.
