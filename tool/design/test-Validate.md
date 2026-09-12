# Test Design: Validate
**Source:** crc-Validate.md

## Test: Validate_AllPass
**Purpose:** Full validation with no issues
**Input:** Well-formed project with all files, unique Rn (may be in any file order), valid refs
**Expected:** Exit 0, output shows all findings, no issues
**Refs:** crc-Validate.md, seq-validate.md

## Test: Validate_GapInRequirements
**Purpose:** Detect gap in requirement numbering regardless of file order
**Input:** requirements.md with R1, R4, R2 (missing R3)
**Expected:** Issue reported: "gap in numbering: R3 missing (have R1, R2, R4)"
**Refs:** crc-Validate.md

## Test: Validate_DuplicateRequirements
**Purpose:** Detect duplicate requirement IDs
**Input:** requirements.md with R1, R2, R2, R3
**Expected:** Issue reported: "duplicate requirement: R2"
**Refs:** crc-Validate.md

## Test: Validate_OutOfOrderRequirements
**Purpose:** Requirements in non-sequential file order pass validation
**Input:** requirements.md with R3, R1, R2 (out of file order but complete)
**Expected:** Exit 0, no numbering issues
**Refs:** crc-Validate.md

## Test: Validate_InvalidRnRef
**Purpose:** Detect CRC card referencing non-existent requirement
**Input:** crc-Store.md references R99, but R99 not in requirements.md
**Expected:** Issue: "crc-Store.md references unknown R99"
**Refs:** crc-Validate.md

## Test: Validate_MissingCodeFile
**Purpose:** Detect artifact referencing missing file
**Input:** Artifacts lists src/missing.go but file doesn't exist
**Expected:** Issue: "src/missing.go listed but not found"
**Refs:** crc-Validate.md

## Test: Validate_MissingTraceability
**Purpose:** Detect code file without CRC comment
**Input:** src/store.go exists but has no // CRC: comment
**Expected:** Issue: "src/store.go missing traceability comment"
**Refs:** crc-Validate.md

## Test: Validate_ApprovedGapSuppressesUncovered
**Purpose:** Approved gaps with Rn references suppress uncovered-requirements issues
**Input:** R5 not in any CRC card, but gaps has `- [ ] A1: R5 (deliberate omission)`
**Expected:** R5 not listed as uncovered, no "uncovered requirements" issue for R5
**Refs:** crc-Validate.md, R65

## Test: Validate_ApprovedGapRangeSuppress
**Purpose:** Approved gaps with Rn-Rm ranges suppress all requirements in range
**Input:** R10-R13 not in CRC cards, gaps has `- [ ] A2: R10-R13 (config concern)`
**Expected:** R10, R11, R12, R13 not listed as uncovered
**Refs:** crc-Validate.md, R65

## Test: each Source diagnostic earns its own half of the fix block
**Purpose:** validates R92 and R188 — the two halves answer different questions (how a
Source line is shaped, what to do when its spec is gone), so firing the wrong one is as
bad as firing none. Also pins that both sit under a single `fix instructions:` header
**Input:** a zero-value `ValidationResult`; then one with only malformed values, only
suspicious lines, only a missing path, and both
**Expected:** empty block for the first; format half only; format half only; missing
half only; both halves under one header
**Refs:** crc-Validate.md — R92, R188
**Code:** internal/validate/validate_test.go

## Test: the missing-Source block names every repair and forecloses the renumber
**Purpose:** validates R188's content rather than its trigger. The block is only worth
emitting if it names all three repairs and rules out the fourth — deleting the orphaned
requirements — which is the failure it exists to prevent and the one no check can catch
**Input:** a `ValidationResult` with one missing Source path
**Expected:** the block mentions renamed, merged, deleted, `minispec update retire`,
`specs/deleted.md`, and that numbers are never renumbered or reused
**Alarm:** 1
**Fire alarm:** restore the pre-R188 gate — in `sourceFixInstructions`, stop appending
`missingSourceFix` (change its guard to `if false`), which is exactly how the function
read before this change — and confirm the missing-only and both cases go red along with
all six content assertions. The `both` case is the one that matters: it stays red while
the format half still emits, which a header-count-only check would sleep through
**Inject:** internal/validate/validate.go:sourceFixInstructions
**Pulled:** 2026-08-13 — rang: 8 assertions red across both tests, restore byte-clean.
Pulled again after the simplification pass restructured the function — still rang
**Refs:** crc-Validate.md — R188
**Code:** internal/validate/validate_test.go

## Test: Validate_OutputShowsFindings
**Purpose:** Output includes what was found for AI verification
**Input:** Any valid project
**Expected:** Output includes "found: R1, R2, R3", "crc-Store.md: R1, R2", coverage map
**Refs:** crc-Validate.md, R30

## Test: the readers disagree over a swallowed tail
**Purpose:** validates R327 and R328 — in each of the three documents, a group never closed loses the tail to the reader and the line scan finds it; the finding leads the report; the unread lines are counted; a gap-shaped bullet outside the Gaps section is not scanned
**Input:** a design.md whose O2 opens a span it never closes with O3 after it and an O9 in another section; a requirements.md whose retired R2 opens a span with R3 after it; a test design whose first entry opens a span with a second entry after it
**Expected:** three findings naming O3, R3 and `2 test entries by line, 1 by the document reader`; no mention of O9; three unread counts; `issues:` leading with `the two readers disagree:`; a clean result still prints the note before `phase: validate OK`
**Refs:** crc-Validate.md — R327, R328
**Code:** internal/validate/agreement_test.go
**Fire alarm:** make `onlyIn` return nothing — the second opinion agrees with everything — and confirm all three findings vanish while the note stays
**Inject:** internal/validate/validate.go:onlyIn
**Pulled:** 2026-09-07 — rang: `missing finding "design.md: the line scan read gap O3 …"` and the other two; restore byte-clean by copy, and again after the simplification pass restructured `readerAgreement`, same signature

## Test: the readers agree over healthy documents
**Purpose:** validates R327 — agreement is silence: a nested sub-bullet under a gap, a retired requirement, a sub-section, two plain test entries produce no finding and no unread line
**Input:** healthy versions of the three documents
**Expected:** no findings, no unread
**Refs:** crc-Validate.md — R327
**Code:** internal/validate/agreement_test.go
**Fire alarm:** widen the gap line scan to any bullet — drop the `X<n>:` key from `gapLineRe` — and confirm the sub-bullet reads as a gap the reader did not return
**Inject:** internal/validate/validate.go:readerAgreement
**Pulled:** 2026-09-07 — rang: `healthy documents: findings [design.md: the line scan read gap sub …]`; restore byte-clean by copy, and again after the simplification pass restructured `readerAgreement`, same signature
