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

## Test: ParseTraceability_Found
**Purpose:** Find traceability comments in code
**Input:**
```go
// CRC: crc-Store.md | Seq: seq-crud.md
func Add() {}
```
**Expected:** Traceability{CRCRefs: ["crc-Store.md"], SeqRefs: ["seq-crud.md"]}
**Refs:** crc-Parser.md

## Test: ParseTraceability_Missing
**Purpose:** Handle code file without traceability comments
**Input:** Go file with no // CRC: comments
**Expected:** Traceability with empty slices
**Refs:** crc-Parser.md

## Test: ParseTraceability_CustomPattern
**Purpose:** Use configurable comment pattern per file extension
**Input:**
```python
# CRC: crc-Store.md | Seq: seq-crud.md
def add(): pass
```
**Expected:** With pattern `#\s*`, finds Traceability{CRCRefs: ["crc-Store.md"], SeqRefs: ["seq-crud.md"]}
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

