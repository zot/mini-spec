# Test Design: Parser
**Source:** crc-Parser.md

## Test: ParseRequirements_ValidFile
**Purpose:** Parse well-formed requirements.md
**Input:**
```markdown
## Feature: Auth
**Source:** specs/auth.md
- **R1:** User can log in
- **R2:** (inferred) Session expires after 30 min
```
**Expected:** 2 requirements, R2 marked inferred, both have source specs/auth.md
**Refs:** crc-Parser.md

## Test: ParseRequirements_MissingSource
**Purpose:** Handle requirements without Source line
**Input:** Requirements section without **Source:** line
**Expected:** Requirements parsed, Source field empty
**Refs:** crc-Parser.md

## Test: ParseCRCCard_ValidCard
**Purpose:** Parse CRC card with Requirements field
**Input:**
```markdown
# Store
**Requirements:** R1, R3, R7
## Knows
...
```
**Expected:** CRCCard{Name: "Store", Requirements: ["R1", "R3", "R7"]}
**Refs:** crc-Parser.md

## Test: ParseCRCCard_NoRequirements
**Purpose:** Handle CRC card missing Requirements field
**Input:** CRC card without **Requirements:** line
**Expected:** CRCCard with empty Requirements slice
**Refs:** crc-Parser.md

## Test: ParseArtifacts_NestedCheckboxes
**Purpose:** Parse artifacts with nested code file checkboxes
**Input:**
```markdown
## Artifacts
- crc-Store.md
  - [x] src/store.go
  - [ ] src/store_test.go
- crc-View.md
  - [x] src/view.go
```
**Expected:** 2 artifacts, first has 2 code files (one checked, one not)
**Refs:** crc-Parser.md

## Test: ParseGaps_AllTypes
**Purpose:** Parse gaps with different type prefixes
**Input:**
```markdown
## Gaps
- [ ] S1: spec gap
- [x] R1: resolved requirement gap
- [ ] D1: design gap
- [ ] C1: code gap
- [ ] O1: oversight
```
**Expected:** 5 gaps, R1 marked resolved, correct types
**Refs:** crc-Parser.md

## Test: an alarm never adopts the next test's fields
**Purpose:** validates R178 — attributing a `**Pulled:**` to an alarm that never had one
is the strongest false claim this parser could make: it would report an unverified
guard as verified
**Input:** an unrecorded alarm followed by a recorded one
**Expected:** the first has no sites and no pull date; the second keeps both
**Alarm:** 1
**Fire alarm:** *structural since 2026-09-07:* the entry is a region of the dependency's
reader, so a field cannot cross into the next entry by construction and no edit in this
tool reaches the property. The test still guards it; the injection would be in
`minispecsdom.ParseTestDoc`, which is theirs. Until 2026-09-07 the site was
`readAlarmFields`, pulled 2026-08-13: the first alarm adopted `2026-08-05`
**Refs:** crc-Parser.md — R178, R316

## Test: a half-written injection site is dropped, not guessed at
**Purpose:** validates R178 — an entry with no colon has no symbol, and half an anchor
points somewhere, which is worse than nowhere because every future check follows it
**Input:** an `**Inject:**` mixing one good site with three malformed ones, and an
unparseable `**Pulled:**`
**Expected:** one site kept; no pull date recorded
**Alarm:** 2
**Fire alarm:** keep a site with an empty symbol — drop the `symbol == ""` half of the
filter in `alarmOf` — and confirm the malformed sites appear. *Re-sited 2026-09-07 from
`parseSites`, which went with the line reader*
**Inject:** internal/parser/testdoc.go:alarmOf
**Pulled:** 2026-09-07 — rang: `Sites = [good.go:Fine garbage-no-colon: nosymbol:], want only good.go:Fine`; restore byte-clean by copy, and again after the simplification pass touched the file, same signature. Previously 2026-08-13 against `parseSites`
**Refs:** crc-Parser.md — R178, R316

## Test: a CRC card's Requirements field reads ranges as their members
**Purpose:** validates R532 — a range in either spelling contributes every member; a plain list is unchanged; text the grammar does not consume is kept as tokens, so validate can still report it
**Input:** cards whose field is `R5-8, R10`; `R5-R7`; `R1, R3, R7`
**Expected:** `R5 R6 R7 R8 R10`; `R5 R6 R7`; `R1 R3 R7`
**Refs:** crc-Parser.md — R532
**Code:** internal/parser/parser_test.go
**Alarm:** 3
**Fire alarm:** go back to the comma split — have `RequirementsField` return the field's comma tokens as written, the reader this replaced. (Skipping only the head parse no longer reaches the property: the per-token reading expands `R5-8` on its own.) `R5-8` comes back as one token, so R6 and R7 vanish from coverage while an unknown-ref finding is the only trace
**Inject:** internal/parser/crc.go:RequirementsField
**Pulled:** 2026-10-03 — rang (re-pulled with the injection rewritten for the per-token reader): with the field returned as its comma tokens, `"R5-8, R10": got ["R5-8" "R10"]` and three more red; restore byte-clean by copy

## Test: leftover text in the field is kept, not dropped
**Purpose:** validates R532 — a field the grammar stops short on still yields every token after the stop, so a junk token is reported and never silently lost
**Input:** a card whose field is `R5, TBD, R9`; one whose field is `R5, TBD, R10-12`
**Expected:** `R5`, `R9`, then `TBD` — refs first, the non-ref after; and `R5`, `R10`, `R11`, `R12`, `TBD`: a range after a stray word still names its members
**Refs:** crc-Parser.md — R532
**Code:** internal/parser/parser_test.go
**Alarm:** 4
**Fire alarm:** discard what `ParseRequirementList` did not consume in `RequirementsField` — the card reads as `R5` alone, `TBD` is never reported, and R9 falls out of coverage with nothing saying why
**Inject:** internal/parser/crc.go:RequirementsField
**Pulled:** 2026-10-03 — rang (re-pulled, test changed): with the unconsumed text discarded, `got ["R5"], want ["R5" "R9" "TBD"]`; restore byte-clean by copy

## Test: a field that opens with a non-ref keeps every token
**Purpose:** validates R532 — when the grammar cannot read even the head, the whole field is kept as tokens, so nothing in it goes unreported
**Input:** a card whose field is `TBD, R5`
**Expected:** `R5`, then `TBD`
**Refs:** crc-Parser.md — R532
**Code:** internal/parser/parser_test.go
**Alarm:** 5
**Fire alarm:** return nothing from `RequirementsField` when `ParseRequirementList` does not match — the card reads as having no requirements, and `TBD` and R5 both vanish. Found past the list: the suite stayed green under exactly this injection until this test existed
**Inject:** internal/parser/crc.go:RequirementsField
**Pulled:** 2026-10-03 — rang (re-pulled, test changed): with a non-matching field returning nothing, `got [], want ["R5" "TBD"]`; restore byte-clean by copy

