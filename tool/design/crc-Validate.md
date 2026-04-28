# Validate
**Requirements:** R24, R25, R26, R27, R28, R29, R30, R31, R3, R40, R41, R42, R43, R63, R64, R65, R66, R68, R69, R70, R72, R76, R78, R84, R85, R86, R88

Runs structural validations and reports findings.

## Knows
- project: loaded Project instance
- findings: accumulated validation results
- issues: deduplicated list of problems, bucketed by category

## Does
- Run(): execute all validations, return ValidationResult
- ValidateRequirements(): check format, unique numbering (no duplicates/gaps, order-independent)
- ValidateCRCCards(): check Requirements fields, valid Rn refs
- ValidateArtifacts(): check structure, file existence
- ValidateGaps(): check ID format (S/R/D/C/I/O/A/T), no duplicates, flag A/T entries that carry a checkbox
- approvedGapReqs(): extract Rn references (individual and ranges) from approved gap descriptions
- retiredReqs(): set of Rn IDs marked retired in requirements.md
- ValidateTraceability(): check code files have CRC comments, CRC/Seq refs exist, inline Rn refs exist in requirements.md
- ValidateImplementationCoverage(): check every non-retired, non-approved requirement appears as inline Rn ref in at least one code file
- ValidateArtifactsCompleteness(): check all design files are listed in Artifacts
- ValidateSpecSources(): check Source fields reference existing spec files
- ValidateCRCSequences(): check files in CRC Sequences sections exist
- FormatText(): emit issues-only output with Rn ranges, deduplicated; on success a single `phase: validate OK` line

## Collaborators
- Project: to locate files
- Parser: to parse and extract data
- Query: to compute coverage for reporting
- os: to check file existence

## Sequences
- seq-validate.md

## Notes
Output is issues-only and machine-greppable. AIs that need to inspect parsed state use the `query` subcommands.
