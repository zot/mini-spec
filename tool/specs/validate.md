# Validate Command

Structural validation of design files. Checks format and references, not intent.

## minispec validate

Run all validations and report issues.

## Checks Performed

### Requirements Format
- requirements.md exists
- All requirements have Rn format with unique numbering (no duplicates, no gaps; file order doesn't matter)
- Each requirement has a Source spec reference
- Inferred requirements are marked

### CRC Card Format
- Every crc-*.md has a **Requirements:** field
- Requirements field references valid Rn identifiers
- No duplicate Rn references within a card

### Artifacts Structure
- design.md has Artifacts section
- All listed design files exist
- All code file paths are valid
- Checkbox syntax is correct

### Gaps Structure
- design.md has Gaps section
- Gap IDs follow S/R/D/C/I/O/A + number format
- No duplicate gap IDs

### Approved Gap Coverage
- Approved (A-type) gaps may reference requirements via Rn or Rn-Rm ranges in their description
- Requirements referenced by approved gaps are treated as covered for validation purposes
- These requirements do not appear in the "uncovered" list or trigger "uncovered requirements" issues

### Traceability Comments
- Code files in Artifacts have `// CRC:` comments
- Referenced CRC and Seq files in code comments exist in design/
- Optional third pipe-delimited section contains inline requirement refs: `// CRC: crc-X.md | Seq: seq-Y.md | R5, R12`
- Parser extracts Rn refs from the third section (comma-separated)
- Inline Rn refs are validated: each must exist in requirements.md

### Implementation Coverage
- Every requirement in requirements.md should appear as an inline Rn ref in at least one code file's traceability comment
- Requirements covered only at the design level (CRC card) but not in any code file are reported as implementation gaps (I-type)
- Requirements covered by approved gaps (A-type) are excluded from this check
- Implementation coverage is reported in validate output alongside design coverage

### Artifacts Manifest Completeness
- All `crc-*.md`, `seq-*.md`, `ui-*.md`, `test-*.md`, `manifest-*.md` files in `design/` are listed in Artifacts section
- Detects orphaned design files not tracked in design.md

### Spec Source Validation
- `**Source:**` fields in requirements.md reference files that exist in `specs/`
- Validates the requirements→specs traceability link

### CRC Sequences Validation
- Files listed in CRC card `## Sequences` sections exist in `design/`
- Validates CRC→sequence traceability

## Output

Show what was found so the AI can verify assumptions and correct mismatches:

```
requirements.md:
  found: R1, R2, R3, R4, R5, R6, R7, R8
  sources: specs/auth.md (R1-R3), specs/storage.md (R4-R8)
  inferred: R3, R7

design files:
  crc-Store.md: R4, R5, R6
  crc-View.md: R1, R2
  crc-Auth.md: (no Requirements field)
  seq-login.md: (sequences don't have requirements)

coverage:
  covered: R1, R2, R4, R5, R6
  uncovered: R3, R7, R8

artifacts:
  crc-Store.md:
    [x] src/store.ts
    [ ] src/store_test.ts
  crc-View.md:
    [x] src/view.ts (file missing)

gaps:
  [ ] S1: ...
  [x] D1: ...

issues:
  - crc-Auth.md: no Requirements field
  - src/view.ts: listed in artifacts but file missing
  - R3, R7, R8: no design coverage
  - seq-logout.md: not listed in Artifacts
  - specs/old-feature.md: referenced as Source but file missing
  - src/store.ts: references crc-Missing.md which does not exist
  - crc-Store.md Sequences: seq-missing.md does not exist
```

The "found" lists let the AI verify parsing matched expectations. If formatting is unusual but parseable, the AI sees what was extracted and can decide if corrections are needed.

Exit code: 0 if no issues, 1 if any issues found.
