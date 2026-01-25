# Phase Subcommands

Phase-specific validation commands for post-phase checks in the mini-spec workflow.

## Purpose

After completing each workflow phase, run the corresponding phase command to validate that phase's outputs before proceeding. This catches issues early rather than at final validation.

## Commands

### minispec phase spec

Run after Spec Phase. Validates:
- At least one spec file exists in `specs/`
- Spec files have content (non-empty)

### minispec phase requirements

Run after Requirements Phase. Validates:
- requirements.md exists and is parseable
- All requirements have Rn format with sequential numbering
- Each Source field references an existing spec file (R41)
- Reports requirements found and their sources

### minispec phase design

Run after Design Phase. Validates:
- design.md exists with Artifacts section
- All design files (crc-*, seq-*, etc.) are listed in Artifacts
- CRC cards have Requirements field with valid Rn references
- All referenced Rn IDs exist in requirements.md
- Reports coverage (covered vs uncovered requirements)

### minispec phase implementation

Run after Implementation Phase. Validates:
- Code files listed in Artifacts exist
- Code files have traceability comments (// CRC:)
- Traceability refs point to existing design files
- Reports artifact checkbox states

### minispec phase gaps

Run after Gaps Phase. Validates:
- Gaps section exists in design.md
- Gap IDs follow S/R/D/C/O + number format
- No duplicate gap IDs
- Reports open vs resolved gaps

## Output

Each phase command shows findings relevant to that phase, with a focused subset of issues. Exit code 0 if no issues for that phase, 1 otherwise.

Example for `minispec phase requirements`:
```
requirements.md:
  found: R1, R2, R3, R4, R5
  sources: specs/overview.md (R1-R3), specs/auth.md (R4-R5)
  inferred: R2

phase: requirements OK
```

Example with issues:
```
requirements.md:
  found: R1, R2, R4
  sources: specs/overview.md (R1-R2, R4)

issues:
  - non-sequential: expected R3, found R4
  - specs/missing.md: referenced as Source but file missing

phase: requirements FAILED
```
