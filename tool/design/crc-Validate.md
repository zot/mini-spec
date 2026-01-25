# Validate
**Requirements:** R24, R25, R26, R27, R28, R29, R30, R31, R3, R40, R41, R42, R43

Runs structural validations and reports findings.

## Knows
- project: loaded Project instance
- findings: accumulated validation results
- issues: list of problems found

## Does
- Run(): execute all validations, return ValidationResult
- ValidateRequirements(): check format, sequential numbering
- ValidateCRCCards(): check Requirements fields, valid Rn refs
- ValidateArtifacts(): check structure, file existence
- ValidateGaps(): check ID format, no duplicates
- ValidateTraceability(): check code files have CRC comments, refs exist
- ValidateArtifactsCompleteness(): check all design files are listed in Artifacts
- ValidateSpecSources(): check Source fields reference existing spec files
- ValidateCRCSequences(): check files in CRC Sequences sections exist

## Collaborators
- Project: to locate files
- Parser: to parse and extract data
- Query: to compute coverage for reporting
- os: to check file existence

## Sequences
- seq-validate.md

## Notes
Output shows what was found at each step (requirements found, coverage map, etc.) so AI can verify parsing assumptions. Issues are listed separately at the end.
