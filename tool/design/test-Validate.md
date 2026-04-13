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

## Test: Validate_OutputShowsFindings
**Purpose:** Output includes what was found for AI verification
**Input:** Any valid project
**Expected:** Output includes "found: R1, R2, R3", "crc-Store.md: R1, R2", coverage map
**Refs:** crc-Validate.md, R30
