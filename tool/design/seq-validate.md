# Sequence: Validate

```
User -> CLI: minispec validate
CLI -> Project: Detect()
Project -> Project: walk up to find design/
Project -> Project: LoadConfig()
Project --> CLI: project

CLI -> Validate: Run(project)

Validate -> Parser: ParseRequirements(requirements.md)
Parser --> Validate: []Requirement
Validate -> Validate: collect all Rn numbers, sort, check for duplicates and gaps
Validate -> Validate: record findings: "found: R1, R2, ..."

Validate -> Project: glob(design/crc-*.md)
loop each CRC file
    Validate -> Parser: ParseCRCCard(path)
    Parser --> Validate: CRCCard
    Validate -> Validate: check Requirements field exists
    Validate -> Validate: check Rn refs are valid
    Validate -> Validate: record findings: "crc-X.md: R1, R3"
end

Validate -> Parser: ParseGaps(design.md)
Parser --> Validate: []Gap
Validate -> Validate: check ID format, no duplicates
Validate -> Validate: extract Rn refs from approved (A-type) gap descriptions

Validate -> Query: Coverage()
Query --> Validate: map[Rn][]files
Validate -> Validate: merge approved-gap Rn refs into covered set
Validate -> Validate: record coverage, find uncovered

Validate -> Parser: ParseArtifacts(design.md)
Parser --> Validate: []Artifact
loop each code file
    Validate -> os: Stat(path)
    Validate -> Validate: record: "[x] path" or "(missing)"
end

Validate -> Validate: record gap findings

loop each code file in artifacts
    Validate -> Project: CommentPattern(ext)
    Project --> Validate: pattern
    Validate -> Parser: ParseTraceability(path, pattern)
    Parser --> Validate: Traceability
    Validate -> Validate: check CRC/Seq refs exist
    Validate -> Validate: check inline Rn refs exist in requirements
    Validate -> Validate: collect all inline Rn refs across code files
end

Validate -> Validate: compute implementation coverage
note: each requirement should appear in at least one code file's inline Rn refs
note: exclude requirements covered by approved gaps
Validate -> Validate: report uncovered as I-type implementation gaps

Validate -> Validate: compile issues list
Validate --> CLI: ValidationResult

CLI -> CLI: Output(result)
CLI --> User: formatted output + exit code
```
