# Test Design: Validate
**Source:** crc-Validate.md

## Test: Validate_AllPass
**Purpose:** Full validation with no issues
**Input:** Well-formed project with all files, sequential Rn, valid refs
**Expected:** Exit 0, output shows all findings, no issues
**Refs:** crc-Validate.md, seq-validate.md

## Test: Validate_NonSequentialRequirements
**Purpose:** Detect gap in requirement numbering
**Input:** requirements.md with R1, R2, R4 (missing R3)
**Expected:** Issue reported: "non-sequential: R3 missing"
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

## Test: Validate_OutputShowsFindings
**Purpose:** Output includes what was found for AI verification
**Input:** Any valid project
**Expected:** Output includes "found: R1, R2, R3", "crc-Store.md: R1, R2", coverage map
**Refs:** crc-Validate.md, R30
