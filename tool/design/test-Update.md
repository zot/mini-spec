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
**Code:** internal/update/update_test.go

## Test: AddRef_FirstRequirement
**Purpose:** Add requirement to empty list
**Input:** crc-Store.md with "**Requirements:**" (empty)
**Expected:** Changed to "**Requirements:** R5"
**Refs:** crc-Update.md
**Code:** internal/update/update_test.go

## Test: AddRef_Duplicate
**Purpose:** Don't add duplicate requirement
**Input:** crc-Store.md already has R5
**Expected:** No change, no error
**Refs:** crc-Update.md
**Code:** internal/update/update_test.go

## Test: RemoveRef_Middle
**Purpose:** Remove requirement from middle of list
**Input:** "**Requirements:** R1, R3, R5"
**Expected:** "**Requirements:** R1, R5"
**Refs:** crc-Update.md
**Code:** internal/update/update_test.go

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
**Alarm:** 1
**Fire alarm:** skip the reader's write — `assigned, werr = nil, nil` in place of `td.NumberAlarms()` — and confirm the run assigns nothing (dropping the call outright leaves `werr` unused and does not compile, which is not a pull)
**Inject:** internal/update/alarmfields.go:NumberAlarms
**Pulled:** 2026-09-07 — rang: `assigned [... []], want [4] — the freed 2 must not be reused`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature

## Test: a re-pull moves the leading date and a first pull is inserted after the site
**Purpose:** validates R314 — the census reads the leading date, so a re-pull moves it and folds the old record after; the body's backticks survive; a first pull lands after the `**Inject:**` it vouches for; a number the document does not hold, or a key with none, is refused
**Input:** an alarm carrying `2026-08-05 — rang, the original record`, re-pulled with a backticked body; an unpulled alarm pulled once; two bad keys
**Expected:** `**Pulled:** 2026-09-07 — … *Earlier —* 2026-08-05 — …`; the first pull directly under its `**Inject:**`; both bad keys refused
**Refs:** crc-Update.md — R314
**Code:** internal/update/alarmfields_test.go
**Alarm:** 2
**Fire alarm:** format the date as `02-01-2006` and confirm the write is refused by the reader's read-back — a `**Pulled:**` must lead with `YYYY-MM-DD`, so a wrongly formatted stamp cannot reach the file
**Inject:** internal/update/alarmfields.go:SetPulled
**Pulled:** 2026-09-07 — rang, through the reader's guard rather than the assertion: `TestDoc.SetPulled on "1" did not read back … leads with a YYYY-MM-DD date … nothing was written`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature

## Test: re-siting voids the record only when the code moves
**Purpose:** validates R315 — a rewrite to the same text changes nothing; a disambiguation resolving to the same lines keeps the record; a move to other lines demotes it to history naming the old sites; an empty site list is refused
**Input:** a fake ranger placing `a.go:F` in HEAD and `a.go:T.F` on disk at the same lines, and `c.go:Moved` elsewhere
**Expected:** no-op: unchanged and byte-identical; `T.F`: rewritten, `**Pulled:**` standing; `Moved`: `**Pulled:**` gone, a `*Pulled at ...` sentence naming `a.go:T.F` in its place
**Refs:** crc-Update.md, crc-Git.md — R315
**Code:** internal/update/alarmfields_test.go
**Alarm:** 3
**Fire alarm:** void on any text change — drop the `!sameCode(...)` conjunct — and confirm the disambiguation case goes red with the record cleared
**Inject:** internal/update/alarmfields.go:SetInject
**Pulled:** 2026-09-07 — rang: `disambiguating to the same lines: cleared=true err=<nil>; want the record kept`; restore byte-clean by copy, and again the same day after the simplification pass restructured the file, same signature

## Test: add-req mints, appends before the sub-heading, and refuses with nothing written
**Purpose:** validates R324 and R325 — numbers count retired and note-level entries, a batch lands in order at the end of the section's own content, the `Feature: ` prefix is optional, an unknown heading and a self-labelled body are refused by name with nothing written
**Input:** a requirements file with a retired R2 and a `### Notes` sub-section holding R4; two texts for `Alpha`, one for `Feature: Beta`, one for an unknown `Gamma`, one opening with `**R9:**`
**Expected:** R5 and R6 before `### Notes`, R7 under Beta, both refusals naming what they refuse, the file byte-identical after them
**Refs:** crc-Update.md — R324, R325
**Code:** internal/update/addreq_test.go
**Alarm:** 4
**Fire alarm:** strip the marker instead of refusing — replace the refusal with `texts[i] = ownMarkerRe.ReplaceAllString(...)` — and confirm the self-labelled body is accepted
**Inject:** internal/update/update.go:AddReq
**Pulled:** 2026-09-07 — rang: `a body writing its own label was not refused naming it: <nil>`; restore byte-clean by copy, and again the same day after the simplification pass restructured `update.go`, same signature

## Test: the gap verbs write through the reader
**Purpose:** validates R326 — add mints counting resolved entries and appends after the last gap, resolve checks and refuses a resolved or permanent gap, approve converts and is idempotent on an approved one
**Input:** a Gaps section with O1 open, O2 resolved, A1, T1
**Expected:** O3 appended before the next heading; O1 checked; O2 and A1 refused; O3 becomes A2; approving A2 returns A2 and writes nothing
**Refs:** crc-Update.md — R326
**Code:** internal/update/addreq_test.go
**Alarm:** 5
**Fire alarm:** return the minted ID from `ApproveGap` without calling the reader's `Approve` — confirm the head line stays `O3`
**Inject:** internal/update/update.go:ApproveGap
**Pulled:** 2026-09-07 — rang: `O3 was not rewritten as A2`; restore byte-clean by copy, and again the same day after the simplification pass restructured `update.go`, same signature

## Test: retire writes both documents through the readers
**Purpose:** validates R80 and R326 — the head line is rewritten in requirements.md, the Tn gap added in design.md, the Source returned, and a second retirement refused
**Input:** R3 retired by R1
**Expected:** `- **~~R3:~~** (Retired T2 — see R1) third`; `- T2: R3 retired by R1 (folded)`; `specs/beta.md`; the second call refused
**Refs:** crc-Update.md — R80, R326
**Code:** internal/update/addreq_test.go
**Alarm:** 6
**Fire alarm:** skip the gaps write — return after the requirements render — and confirm the Tn line is missing
**Inject:** internal/update/update.go:Retire
**Pulled:** 2026-10-08 — rang, by hand at `#102`'s commit after `Retire` moved to `EditFiles`: the design-side render returning its source unchanged fails `the Tn gap was not added`, and the refusal and T-number tests beside it; restore byte-clean *Earlier —* 2026-09-07 — rang: `the Tn gap was not added`; restore byte-clean by copy, and again the same day after the simplification pass restructured `update.go`, same signature
## Test: add-ref writes the canonical form
**Purpose:** validates R533 — the whole field comes back sorted, each ref once, a run of three or more as `R5-8`, a pair listed
**Input:** a card with `**Requirements:** R8, R1, R5, R7`; add R6
**Expected:** `**Requirements:** R1, R5-8`; then add R3 to `R1, R2` gives `R1, R2, R3`'s canonical `R1-3`
**Refs:** crc-Update.md, seq-update.md — R533
**Code:** internal/update/update_test.go
**Alarm:** 7
**Fire alarm:** render each ref as a separate list item in `rewriteRequirements` instead of `sdom.RequirementText` — sorted, but no range ever forms, and a card already written `R5-8` is expanded on its next add. Keep `sdom` referenced, or the build fails on the unused import and nothing is learned
**Inject:** internal/update/update.go:rewriteRequirements
**Pulled:** 2026-10-03 — rang: with each ref listed and no range rendered, `got "…R1, R5, R6, R7, R8…", want "…R1, R5-8…"`; the first attempt replaced the only `sdom` call, broke the build on an unused import and was not counted; restore byte-clean by copy

## Test: add-ref on a card with no Requirements line writes one
**Purpose:** validates R534 and closes O36 — the line appears beneath the `#` heading; a ref inside a range is already present and nothing is written
**Input:** a card `# Store` with no Requirements line; add R5. Then a card with `R5-8`; add R6
**Expected:** `# Store` followed by `**Requirements:** R5`; the second card byte-identical
**Refs:** crc-Update.md — R534
**Code:** internal/update/update_test.go
**Alarm:** 8
**Fire alarm:** return without inserting when `rewriteRequirements` finds no Requirements line — the O36 defect: the add reports success and the card is unchanged
**Inject:** internal/update/update.go:rewriteRequirements
**Pulled:** 2026-10-03 — rang: with no insert when the line is absent, `an add with nowhere to write must make the line (O36)`; restore byte-clean by copy

## Test: remove-ref splits a range, and the last ref takes the line with it
**Purpose:** validates R535 — removing a member from inside a range splits it; removing the only ref removes the line rather than leaving `**Requirements:**` empty
**Input:** `**Requirements:** R5-8`, remove R6; a card with only `R5`, remove R5
**Expected:** `**Requirements:** R5, R7, R8`; the second card has no Requirements line and nothing else changed
**Refs:** crc-Update.md — R535
**Code:** internal/update/update_test.go
**Alarm:** 9
**Fire alarm:** write the empty field instead of removing the line in `rewriteRequirements` — `**Requirements:**` with nothing after it, which the grammar does not read back
**Inject:** internal/update/update.go:rewriteRequirements
**Pulled:** 2026-10-03 — rang: with the empty field written instead of the line removed, `got "# Store\n**Requirements:** \n\nA store.\n"`; restore byte-clean by copy

## Test: tokens that are not refs survive the rewrite
**Purpose:** validates R536 — a non-ref token is kept as written after the refs, so validate still reports it
**Input:** `**Requirements:** R5, TBD, R9`; add R6
**Expected:** `**Requirements:** R5, R6, R9, TBD`
**Refs:** crc-Update.md — R536
**Code:** internal/update/update_test.go
**Alarm:** 10
**Fire alarm:** drop the leftover tokens in `rewriteRequirements` — `TBD` vanishes, and the one thing validate would have reported is gone with no trace
**Inject:** internal/update/update.go:rewriteRequirements
**Pulled:** 2026-10-03 — rang: with the tokens dropped, `got "**Requirements:** R5, R6, R9", want "…, TBD"`; restore byte-clean by copy

## Test: removing a ref the card lacks is an error
**Purpose:** validates R537 — a remove that finds nothing is refused naming the ref and the card, and the file is untouched
**Input:** `**Requirements:** R1, R5`; remove R3
**Expected:** an error naming R3 and the card; the file byte-identical
**Refs:** crc-Update.md — R537
**Code:** internal/update/update_test.go
**Alarm:** 11
**Fire alarm:** let `RemoveRef` succeed when the ref is absent — the old behavior, a silent success that tells the caller something false
**Inject:** internal/update/update.go:RemoveRef
**Pulled:** 2026-10-03 — rang: with the absence check disabled, `want an error naming R3 and the card, got <nil>`; restore byte-clean by copy

## Test: an argument that is not a requirement ref is refused
**Purpose:** validates R533 — `add-ref` takes an `Rn` with n at least 1; anything else is refused and the card is untouched
**Input:** a card with `R1`; add `5`, `R0`, `Rx`, `O5`
**Expected:** four refusals; the card byte-identical
**Refs:** crc-Update.md — R533
**Code:** internal/update/update_test.go
**Alarm:** 12
**Fire alarm:** let `reqNumber` accept any number it can read — `5` writes R5 and `R0` writes a ref to a requirement that cannot exist. Found past the list: the suite stayed green under exactly this injection until this test existed
**Inject:** internal/update/update.go:reqNumber
**Pulled:** 2026-10-03 — rang: with only the number checked, `"5" changed the card: … R1, R5` and `"R0" changed the card: … R0, R1`; restore byte-clean by copy

## Test: retire writes neither document when the gap write is refused
**Purpose:** validates R547 — a refusal from either reader leaves both files as they were
**Input:** a requirements.md and a design.md whose Gaps section holds a deviant entry the gaps reader refuses to write past, or a `T` already carrying the minted number; retire R3
**Expected:** an error; requirements.md and design.md byte-identical to before
**Refs:** crc-Update.md, seq-update.md — R547
**Code:** internal/update/addreq_test.go
**Alarm:** 13
**Fire alarm:** restore the field defect — write requirements.md through its own `EditFile` before the gap edit renders. Red: requirements.md carries `(Retired T…` after the refusal (ui-engine, 2026-10-07)
**Inject:** internal/update/update.go:Retire
**Pulled:** 2026-10-08 — rang, by hand after simplification: writing requirements.md through its own `EditFile` before the paired edits fails `TestRetireWritesNeitherDocumentWhenTheGapWriteIsRefused` (`requirements.md was written although the gap was refused`), and the happy path with it (`the requirement is already retired`); update, parser, query run unfiltered with `-count=1`; restore byte-clean

## Test: the next Tn counts the retired markers
**Purpose:** validates R549 — a `Tn` named only by a requirement's marker is still taken, by `retire` and by `query next-id gap`
**Input:** a requirements.md with `(Retired T5 — see R1)` and a design.md whose highest `T` gap is `T2`; retire another requirement, and ask next-id
**Expected:** the retirement is `T6`; `query next-id gap` reports `T6` before it
**Refs:** crc-Update.md, crc-Query.md — R549
**Code:** internal/update/addreq_test.go
**Alarm:** 14
**Fire alarm:** mint from the gaps alone (`nextGapID(gaps, "T")`). Red: the retirement is `T3`, a number the marker already holds
**Inject:** internal/update/update.go:Retire
**Pulled:** 2026-10-08 — rang, by hand after simplification: minting from the gaps alone gives `retired as "T2", want T6` and `the gap does not carry T6`; restore byte-clean

## Test: EditFiles writes in the order given
**Purpose:** validates R548 — the file passed first is on disk when a later write fails, which is what lets `retire` put the gap first
**Input:** two files, the second in a read-only directory (readable, its temp file uncreatable); both rendered to new content
**Expected:** an error; the first file carries the new content
**Refs:** crc-Update.md — R548
**Code:** internal/update/addreq_test.go
**Alarm:** 15
**Fire alarm:** write the edits in reverse order in `EditFiles`. Red: `the first file was not written before the second failed`
**Inject:** internal/parser/carve.go:EditFiles
**Pulled:** 2026-10-08 — rang, by hand after simplification: writing the edits in reverse order fails `TestEditFilesWritesInTheOrderGiven` (`the first file was not written before the second failed`); restore byte-clean
