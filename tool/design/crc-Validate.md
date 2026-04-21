# Validate
**Requirements:** R24, R25, R26, R27, R28, R29, R30, R31, R3, R40, R41, R42, R43, R63, R64, R65, R66, R68, R69, R70, R72

Runs structural validations and reports findings.

## Knows
- project: loaded Project instance
- findings: accumulated validation results
- issues: list of problems found

## Does
- Run(): execute all validations, return ValidationResult
- ValidateRequirements(): check format, unique numbering (no duplicates/gaps, order-independent)
- ValidateCRCCards(): check Requirements fields, valid Rn refs
- ValidateArtifacts(): check structure, file existence
- ValidateGaps(): check ID format (S/R/D/C/I/O/A), no duplicates, exclude A from open count
- approvedGapReqs(): extract Rn references (individual and ranges) from approved gap descriptions
- ValidateTraceability(): check code files have CRC comments, CRC/Seq refs exist, inline Rn refs exist in requirements.md
- ValidateImplementationCoverage(): check every requirement appears as inline Rn ref in at least one code file; exclude approved-gap requirements; report uncovered as I-type gaps
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
