# Parser
**Requirements:** R5, R6, R7, R8, R9, R51, R52, R53

Parses mini-spec design file formats into structured data.

## Knows
- Requirement: {ID, Text, Source, Inferred bool, Line int}
- CRCCard: {Name, Requirements []string, Sequences []string, Path string}
- Artifact: {DesignFile, CodeFiles []CodeFile}
- CodeFile: {Path, Checked bool, Line int}
- Gap: {ID, Type, Description, Resolved bool, Line int}
- Traceability: {CRCRefs []string, SeqRefs []string}

## Does
- ParseRequirements(path): parse requirements.md -> []Requirement
- ParseCRCCard(path): parse crc-*.md -> CRCCard
- ParseArtifacts(path): parse design.md Artifacts section -> []Artifact
  - Supports inline format: `- [x] design.md → code.ts, code2.ts`
  - Skips subsection headers (`### CRC Cards`, etc.)
  - Parses comma-separated code files after `→`
  - Strips backticks from code file paths
- ParseGaps(path): parse design.md Gaps section -> []Gap
- ParseTraceability(path, commentPattern): scan code file for CRC: comments using the provided pattern -> Traceability

## Collaborators
- os: file reading
- regexp: pattern matching
- bufio: line-by-line scanning
- Project: provides comment patterns for file extensions

## Sequences
- seq-parse.md
