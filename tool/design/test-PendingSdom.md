# Test Design: PendingSdom
**Source:** crc-PendingSdom.md

## Test: the fixture's entries read back
**Purpose:** R370, R371, R372, R376
**Input:** `internal/minispecsdom/testdata/pending-sample.md`: a preamble and rule, five entries — one with a skill, one with a `Next:` line and a nested sub-item, one whose body holds a fenced `## 9.`, a gap-sourced one, one whose gap source is a range — and a `## Notes` heading between
**Expected:** ids `8, 12, 3, 14, 15` in order; titles, skill, status, Source document and key, Next as written; `MaxID` 15; `Unread` lists `Notes` and the range-sourced `Source:` line, each with its line; the fenced `## 9.` is no entry; the render is byte-exact
**Refs:** crc-PendingSdom.md, seq-pending.md#1
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 1
**Fire alarm:** end a region at any heading regardless of level. Red: the entry with the nested `###` sub-item loses its Next line to the sub-item's region.
**Inject:** internal/minispecsdom/pending.go:Pending.regionEnd
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `the sub-item is not inside entry 12's region`, and `Remove` left the sub-item behind; restore clean. *Earlier —* 2026-09-03 — rang: `the sub-item is not inside entry 12's region`, and `Remove(12)` left the sub-item behind — two tests.
## Test: Place by position, refused not clamped
**Purpose:** R407, R374, R377
**Input:** `Place` at 1, at `len+1`, and after id 12; then at 0 and at `len+2`
**Expected:** the renders show the new entry before the first, at the end — the file still ending in one newline — and after 12's region, each separated by a blank line; `Entries` reflects each write; the two bad positions error and the render is unchanged
**Refs:** crc-PendingSdom.md, seq-pending.md#2
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 2
**Fire alarm:** clamp an out-of-range position to the nearest valid one. Red: `Place` at 0 succeeds and the entry lands first.
**Inject:** internal/minispecsdom/pending.go:Pending.Place
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `position 0 accepted`, then the read-back guard panicked on the entry landing at 10; restore clean. *Earlier —* 2026-09-05 — re-pulled after the simplifier restructured `Place`; rang in the same test. Earlier the same day: re-pulled by hand after Item 6 rewrote the site; rang: `position 0 accepted`, `position len+2 accepted`, `a refused Place changed the document`, only that test. Previously 2026-09-05 — re-pulled by a delegated puller (stale since Item 5 moved `Place` on 2026-09-04); rang: `position 0 accepted`, `position len+2 accepted`, `a refused Place changed the document`, only that test. Previously 2026-09-03 — rang: `position 0 accepted`, `position len+2 accepted`, `a refused Place changed the document`, only that test.
## Test: Remove drops exactly the run
**Purpose:** R408
**Input:** `Remove(12)` on the fixture, then `Remove(99)`
**Expected:** the render equals the fixture with entry 12's lines gone and its neighbours' bytes untouched; entries read `8, 3, 14, 15`; the unknown id errors
**Refs:** crc-PendingSdom.md, seq-pending.md#3
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 3
**Fire alarm:** skip the tail split so the next entry's leading bytes go with the removed run. Red: the render loses the following entry's heading.
**Inject:** internal/minispecsdom/pending.go:Pending.Remove
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `after Remove: "# Pending\n\n---\n"` and the round trip left a `Source:` tail; restore clean. *Earlier —* 2026-09-05 — re-pulled by hand after Item 6 rewrote the site; rang: the rule-ends-region test (`after Remove:`) and the new identity test's rule-and-commentary case, which lost its rule to the run. Previously 2026-09-03 — rang: the rule-ends-region test, `after Remove: "# Pending\n\n---\n\n"` — the rule and the prose went with the run.
## Test: a rule ends a region
**Purpose:** R371, R375 — found by injecting past the alarm list, not by design
**Input:** one entry followed by `---` and trailing prose
**Expected:** the entry's region stops before the rule; `Remove` leaves the rule and the prose exactly as they were
**Refs:** crc-PendingSdom.md, seq-pending.md#1.3
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 4
**Fire alarm:** drop the `---` branch in `regionEnd`. Red: the region runs to the end of the file and `Remove` takes the rule and the prose with it. Every other test stayed green under this injection on 2026-09-03, since the fixture's only rule precedes its entries.
**Inject:** internal/minispecsdom/pending.go:Pending.regionEnd
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `the region ran past the rule`, and the last-position Place landed after the commentary with it; restore clean. *Earlier —* 2026-09-03 — rang: `the region ran past the rule`, only this test; injection verified applied, restored by reverse edit.
## Test: entries and unread headings carry their line
**Purpose:** R395, R396
**Input:** the fixture (folded into "the fixture's entries read back")
**Expected:** entries 8, 12, 3 at lines 8, 12, 19; `Unread` is `{25, "Notes"}` then the range-sourced line at 33
**Refs:** crc-PendingSdom.md, seq-pending.md#1
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 5
**Fire alarm:** record every line as 0 in `scan` — entries and unread alike. Red: the line assertions and the unread comparison both fail.
**Inject:** internal/minispecsdom/pending.go:Pending.scan
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `lines 0 0 0` and the unread items at line 0 and 1; restore clean. *Earlier —* 2026-09-05 — re-pulled by a delegated puller after the reader gained the unclosed report; rang: `unread [{Line:0 Text:Notes} …]` and `lines 0 0 0`, only the fixture read-back. Previously 2026-09-04 — rang: `unread [{Line:0 Text:Notes}]` and `lines 0 0 0`; only that test. A first attempt did not build: the probe comment swallowed the closing parenthesis on the same line. Re-pulled the same day after the simplification pass hoisted the line into a local: rang again.
## Test: a source is a part or a gap
**Purpose:** R399, R400, R401
**Input:** the fixture's entries 8 (`part `#7``), 14 (`gap `O136``), 15 (`gap `O1-O3``, a range) and 3 (no `Source:` line); `EntryText{Kind: SourceGap, SourceKey: "O7"}.Text()`; `Place` of a gap entry keyed `O1, O2`; `Place` of the `O7` entry at the end
**Expected:** 8 is `SourcePart` key `7`; 14 is `SourceGap` key `O136`; 15 is `SourceNone` with an empty key and its document still read; 3 is `SourceNone` and not unread; the text renders `gap `O7``; the list is refused with `ErrBadGapSource`; the placed entry reads back as a gap source
**Refs:** crc-PendingSdom.md, seq-pending.md#1.4.1, seq-pending.md#2.1.1
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 6
**Fire alarm:** drop the `gapIDRe` check from `derive` so any backquoted gap key reads as `SourceGap`. Red: entry 15 reads as a gap keyed `O1-O3` and leaves `Unread`. A second injection: drop the refusal from `Place`. Red: the list is placed, and reads back as `SourceNone`.
**Inject:** internal/minispecsdom/pending.go:Entry.derive, internal/minispecsdom/pending.go:Pending.Place
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang, both injections: entry 15 read as a gap keyed `O1-O3` with the Notes heading left unread; the placed list panicked the read-back with `Kind:0`; restore clean. *Earlier —* 2026-09-05 — the `Place` injection re-pulled by hand after Item 6 rewrote `Place`; rang: `TestASourceIsAPartOrAGap`, only that test. Previously 2026-09-04 — both rang: without the ID check entry 15 read as `Kind:2` keyed `O1-O3` and `Unread` shrank to `Notes` alone (the fixture test objected too); without the refusal `a list as a gap source: <nil>`; only those tests; restore clean. The `derive` injection re-pulled the same day after the simplification pass named the groups: rang again.
## Test: a group open at end of input is unread
**Purpose:** R403, R445
**Input:** a pending file whose entry is followed by a two-backtick span holding a lone backtick, never closed, and then a second entry
**Expected:** both entries read; `Unread` holds two items at the span's line, the two-run and the lone backtick it had enclosed, each never closed and read as text
**Refs:** crc-PendingSdom.md
**Code:** internal/minispecsdom/pending_test.go

## Test: Place at the last position lands before the rule
**Purpose:** R407 — where the entries end, not where the file does
**Input:** the fixture followed by `---` and commentary, `Place` at `len+1`; the same with every entry removed first, then `Place` at 1; a file whose header rule is its last line
**Expected:** the new entry sits after the last entry and before the rule with one blank line each side; with no entries it sits after the header rule; with nothing after the rule it ends the file in one newline
**Refs:** crc-PendingSdom.md, seq-pending.md#2.3.1, seq-pending.md#2.3.3
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 7
**Fire alarm:** make the last-position case fall through to `appendEntry`. Red: the entry lands after the commentary.
**Inject:** internal/minispecsdom/pending.go:Pending.Place
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `entry not between the last entry and the rule`, landing after the commentary; restore clean. *Earlier —* 2026-09-05 — re-pulled after the simplifier restructured `Place` into guard clauses; rang: `TestPlaceAtTheLastPositionLandsBeforeTheRule`, only that test. A first attempt left unreachable code, which `go vet` refuses — no evidence, not silence. Earlier the same day: rang: `entry not between the last entry and the rule`, only that test; the first pull of the sibling alarm below was a false silent — the injection left a variable unused and the build failed, which the pull script read as no failures. A build failure must be read as no evidence, never as green.
## Test: Place then Remove is the identity
**Purpose:** R407, R408 — old R480 on `old-sdom`: `add-item` and `finish` are inverses or they are not
**Input:** the fixture as it is, the fixture with a trailing rule and commentary, and the fixture with no final newline; `Place` at `len+1` then `Remove` of the new id
**Expected:** the render equals the source byte for byte in the first two shapes; the third gains its missing final newline and is otherwise identical
**Refs:** crc-PendingSdom.md, seq-pending.md#2.3.2, seq-pending.md#3.2.1
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 8
**Fire alarm:** drop `closeTail` from `Remove`. Red: the plain fixture ends in a blank line after the round trip — the one-line growth mini-spec measured.
**Inject:** internal/minispecsdom/pending.go:Pending.Remove
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `not the identity`, a `Source:` tail and an extra blank line in both the plain and the no-final-newline cases; restore clean. *Earlier —* 2026-09-05 — rang: `plain: not the identity; tail "…\n\n"` and the no-final-newline case, only `TestPlaceThenRemoveIsTheIdentity`.
## Test: The canonical entry is flush left and an indented one still reads
**Purpose:** R501
**Input:** `Place` of a part entry (the gap form's render is asserted by *A source is a part or a gap*); then `ParsePending` over a file whose `Source:` and `Next:` lines carry the old three-space indent
**Expected:** the rendered `Source:` and `Next:` lines begin at column zero; the indented file's entry reads its document, key and next action exactly as an unindented one does
**Refs:** crc-PendingSdom.md, seq-pending.md#2.2
**Code:** internal/minispecsdom/pending_test.go
**Fire alarm:** restore the three-space prefix in `Place`'s render. Red: the rendered `Source:` line begins with spaces.
**Inject:** internal/minispecsdom/pending.go:Pending.Place
**Pulled:** 2026-09-16 — rang: `not flush left`, the rendered `Source:` line opening with three spaces

## Test: a title with emphasis inside reads whole
**Purpose:** R412 — the title from the node, not a regex
**Input:** `## 5. **A **b** c** (skill). status here.` with a `Source:` line
**Expected:** title `A **b** c`, skill `skill`, status `status here`
**Refs:** crc-PendingSdom.md, seq-pending.md#1.2
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 9
**Fire alarm:** read the title with the old lazy regex. Red: title `A `.
**Inject:** internal/minispecsdom/pending.go:Entry.deriveHead
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `title "A "`, the rest spilling into the status; restore clean. The entry was unreadable to the tool until this pass: the previous alarm's 2026-09-05 record carried an unclosed backtick span that swallowed this entry, repaired by hand the same day. *Earlier —* 2026-09-05 — pulled again after the simplifier restructured the site; rang. Earlier the same day: rang: `title "A "`, only that test.
## Test: a write that does not read back panics
**Purpose:** R413 — the guard, and the shape of what it says
**Input:** `mustReadBack` with a false verdict; every existing write test, which passes through the guard green
**Expected:** a `*ReadBackError` naming reader, write, key, want and got; a true verdict is silent
**Refs:** crc-PendingSdom.md
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 10
**Fire alarm:** corrupt `Place`'s bytes — prefix the inserted text with `#`. Red: every `Place` test panics with `Pending.Place … did not read back`. A second: have `Remove` leave the head node in place. Red: `Remove … the entry still reads`.
**Inject:** internal/minispecsdom/pending.go:Pending.Place, internal/minispecsdom/pending.go:Pending.Remove
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang, both injections: `Pending.Place on "20" did not read back … got <nil>`; the head left in place gave `## ## 3.` and `ids [8 14 15], want [8 3 14 15]`, with three sibling tests failing beside it; restore clean. *Earlier —* 2026-09-05 — pulled again after the simplifier restructured the guards, rang again; first: both rang: the corrupted placement and the unremoved head each panicked through the guard in every write test.
## Test: a rule is a line of its own outside code
**Purpose:** R447 — mini-spec's `pending finish 55` left an entry's tail behind a three-dash code span (2026-09-12); a `---` line in a fence did the same here, measured the same day
**Input:** two entries; the first's body holds a three-dash code span mid-sentence, or a fence whose one line is `---`; `Remove` of the first
**Expected:** both read as two entries; after `Remove` the render is the header and the second entry, no tail
**Refs:** crc-PendingSdom.md, seq-pending.md#1.3
**Code:** internal/minispecsdom/pending_test.go
**Alarm:** 11
**Fire alarm:** in `ruleAt`, drop the `inCode` guard and the two boundary tests, so any node line reading `---` counts. Red: both cases leave a tail after `Remove`, the span case from the closing backtick on.
**Inject:** internal/minispecsdom/pending.go:Pending.ruleAt

**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: both cases left a tail after `Remove`, the span case from the closing backtick on; restore clean. *Earlier —* 2026-09-12 — rang: this test alone, both cases (`the region ended at the dashes, leaving a tail`); injection by hand in the live tree, restored from a copy, diff checked.
