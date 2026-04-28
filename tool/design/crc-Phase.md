# Phase
**Requirements:** R44, R45, R46, R47, R48, R49, R50, R63, R76, R87

Phase-specific validation for post-phase checks in the mini-spec workflow.

## Knows
- phaseName: which phase to validate (spec, requirements, design, implementation, gaps)

## Does
- RunSpec(): validate spec files exist and are non-empty
- RunRequirements(): validate requirements.md format and spec sources
- RunDesign(): validate design files, CRC cards, requirement coverage
- RunImplementation(): validate code files and traceability comments
- RunGaps(): validate gaps section structure; flag A/T entries that carry a checkbox; report open/resolved/approved/retired separately
- FormatResult(): format phase-specific output using ranges and dedup; on success a single OK summary line plus only the sparse findings the AI cannot get from a query subcommand

## Collaborators
- Project: to locate files
- Validate: reuses validation logic
- Query: for coverage queries
- Parser: for parsing design files

## Sequences
- seq-phase.md
