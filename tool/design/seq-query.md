# Sequence: Query Coverage

```
User -> CLI: minispec query coverage
CLI -> Project: Detect()
Project --> CLI: project

CLI -> Query: Coverage(project)

Query -> Parser: ParseRequirements(requirements.md)
Parser --> Query: []Requirement {R1, R2, R3, ...}

Query -> Project: glob(design/crc-*.md)
loop each CRC file
    Query -> Parser: ParseCRCCard(path)
    Parser --> Query: CRCCard {Name, Requirements: [R1, R3]}
end

Query -> Query: build map[Rn] -> []files
Query --> CLI: CoverageResult

CLI -> CLI: Output(result)
CLI --> User: "R1: crc-Store.md, crc-View.md\nR2: crc-Store.md\nR3: (none)"
```

# Sequence: Query Uncovered

```
User -> CLI: minispec query uncovered
CLI -> Query: Uncovered(project)
Query -> Query: Coverage()  // reuse coverage logic
Query -> Query: filter where files == empty
Query --> CLI: []string{R3, R7}
CLI --> User: "R3\nR7"
```

## `query gaps` with a selection

The RANGE arguments are expanded in the inline-ref grammar first (a range across two types is
refused there), the flags parsed wherever they sit, the gaps read through the dependency's
reader, and the selection applied — IDs, then checkbox state — before the output form is
chosen, so `--json` renders exactly the selected set. Nothing matched by ID is an error naming
the ask; a valid selection matching nothing prints what was asked (R317–R323).
