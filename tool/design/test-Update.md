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
