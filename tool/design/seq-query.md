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

## `query implementation` — the reverse lookup

The args classify first: parsed cleanly as requirement refs (the `ExpandGapRefs` grammar,
R-only) they are the selected set and it is number mode; otherwise the sole arg compiles to a
regexp and the selected set is the requirements whose text it matches (text mode). The harvest
then runs once — for each code file the Artifacts manifest lists, parse it with the sdom
language its extension maps to, run the traceability-comment reader, and index each Rn (ranges
expanded) to its `file:line` and comment. Each selected requirement is looked up in that index;
number mode prints only its locations, text mode prints the requirement then its locations, and
a selected requirement absent from the index prints "no impl refs". Retired requirements are
kept in the selected set by number and dropped in text mode unless `--retired` (R502–R507).
