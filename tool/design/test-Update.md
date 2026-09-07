# Test Design: Update
**Source:** crc-Update.md

## Test: Check_GapItem
**Purpose:** Check a gap item checkbox
**Input:** design.md with "- [ ] D1: description"
**Expected:** Line changed to "- [x] D1: description"
**Refs:** crc-Update.md, seq-update.md

## Test: Check_ArtifactFile
**Purpose:** Check a code file checkbox in Artifacts
**Input:** design.md with "  - [ ] src/store.go"
**Expected:** Line changed to "  - [x] src/store.go"
**Refs:** crc-Update.md

## Test: Uncheck_Preserves
**Purpose:** Uncheck preserves surrounding content
**Input:** design.md with other content around checkbox
**Expected:** Only checkbox changed, rest of file unchanged
**Refs:** crc-Update.md

## Test: AddRef_NewRequirement
**Purpose:** Add requirement to existing list
**Input:** crc-Store.md with "**Requirements:** R1, R3"
**Expected:** Changed to "**Requirements:** R1, R3, R5"
**Refs:** crc-Update.md, seq-update.md

## Test: AddRef_FirstRequirement
**Purpose:** Add requirement to empty list
**Input:** crc-Store.md with "**Requirements:**" (empty)
**Expected:** Changed to "**Requirements:** R5"
**Refs:** crc-Update.md

## Test: AddRef_Duplicate
**Purpose:** Don't add duplicate requirement
**Input:** crc-Store.md already has R5
**Expected:** No change, no error
**Refs:** crc-Update.md

## Test: RemoveRef_Middle
**Purpose:** Remove requirement from middle of list
**Input:** "**Requirements:** R1, R3, R5"
**Expected:** "**Requirements:** R1, R5"
**Refs:** crc-Update.md

## Test: AddGap_AutoNumber
**Purpose:** Auto-number new gap
**Input:** Gaps section has S1, R1, R2, D1
**Expected:** New R gap gets ID R3
**Refs:** crc-Update.md, seq-update.md

## Test: ResolveGap_Alias
**Purpose:** resolve-gap is alias for check
**Input:** minispec update resolve-gap D1
**Expected:** Same as minispec update check design.md D1
**Refs:** crc-Update.md

## Test: numbering is append-only and idempotent
**Purpose:** validates R311 and R313 — a freed number is never reused, existing numbers are untouched, the migration adds its own lines and nothing else, and a second run writes nothing
**Input:** a test design holding alarms 1 and 3 — the shape deleting 2 leaves — and one unnumbered alarm
**Expected:** the unnumbered alarm becomes 4, inserted above its `**Fire alarm:**`; stripping that line reproduces the input byte for byte; a second run assigns nothing and changes no byte
**Refs:** crc-Update.md — R311, R313
**Code:** internal/update/alarmfields_test.go
**Fire alarm:** skip the reader's write — `assigned, werr = nil, nil` in place of `td.NumberAlarms()` — and confirm the run assigns nothing (dropping the call outright leaves `werr` unused and does not compile, which is not a pull)
**Inject:** internal/update/alarmfields.go:NumberAlarms
**Pulled:** 2026-09-07 — rang: `assigned [... []], want [4] — the freed 2 must not be reused`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature

## Test: a re-pull moves the leading date and a first pull is inserted after the site
**Purpose:** validates R314 — the census reads the leading date, so a re-pull moves it and folds the old record after; the body's backticks survive; a first pull lands after the `**Inject:**` it vouches for; a number the document does not hold, or a key with none, is refused
**Input:** an alarm carrying `2026-08-05 — rang, the original record`, re-pulled with a backticked body; an unpulled alarm pulled once; two bad keys
**Expected:** `**Pulled:** 2026-09-07 — … *Earlier —* 2026-08-05 — …`; the first pull directly under its `**Inject:**`; both bad keys refused
**Refs:** crc-Update.md — R314
**Code:** internal/update/alarmfields_test.go
**Fire alarm:** format the date as `02-01-2006` and confirm the write is refused by the reader's read-back — a `**Pulled:**` must lead with `YYYY-MM-DD`, so a wrongly formatted stamp cannot reach the file
**Inject:** internal/update/alarmfields.go:SetPulled
**Pulled:** 2026-09-07 — rang, through the reader's guard rather than the assertion: `TestDoc.SetPulled on "1" did not read back … leads with a YYYY-MM-DD date … nothing was written`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature

## Test: re-siting voids the record only when the code moves
**Purpose:** validates R315 — a rewrite to the same text changes nothing; a disambiguation resolving to the same lines keeps the record; a move to other lines demotes it to history naming the old sites; an empty site list is refused
**Input:** a fake ranger placing `a.go:F` in HEAD and `a.go:T.F` on disk at the same lines, and `c.go:Moved` elsewhere
**Expected:** no-op: unchanged and byte-identical; `T.F`: rewritten, `**Pulled:**` standing; `Moved`: `**Pulled:**` gone, a `*Pulled at \`a.go:T.F\`` sentence in its place
**Refs:** crc-Update.md, crc-Git.md — R315
**Code:** internal/update/alarmfields_test.go
**Fire alarm:** void on any text change — drop the `!sameCode(...)` conjunct — and confirm the disambiguation case goes red with the record cleared
**Inject:** internal/update/alarmfields.go:SetInject
**Pulled:** 2026-09-07 — rang: `disambiguating to the same lines: cleared=true err=<nil>; want the record kept`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature
