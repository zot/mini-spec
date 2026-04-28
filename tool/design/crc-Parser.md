# Parser
**Requirements:** R5, R6, R7, R8, R9, R51, R52, R53, R59, R61, R66, R67, R71, R73, R74, R75, R77

Parses mini-spec design file formats into structured data.

## Knows
- Requirement: {ID, Text, Source, Inferred bool, Retired bool, Line int}
- CRCCard: {Name, Requirements []string, Sequences []string, Path string}
- Artifact: {DesignFile, CodeFiles []CodeFile}
- CodeFile: {Path, Checked bool, Line int}
- Gap: {ID, Type, Description, Resolved bool, HasCheckbox bool, Line int}
- Traceability: {CRCRefs []string, SeqRefs []string, ReqRefs []string}

## Does
- ParseRequirements(path): parse requirements.md -> []Requirement
  - Accepts strikethrough retired form `- **~~Rn:~~** (Retired Tk — see Rxxx) <text>`; sets Retired flag
- ParseCRCCard(path): parse crc-*.md -> CRCCard
- ParseArtifacts(path): parse design.md Artifacts section -> []Artifact
  - Supports inline format: `- [x] design.md → code.ts, code2.ts`
  - Skips subsection headers (`### CRC Cards`, etc.)
  - Parses comma-separated code files after `→`
  - Strips backticks from code file paths
- ParseGaps(path): parse design.md Gaps section -> []Gap (types: S/R/D/C/I/O/A/T)
  - Tn entries always have no checkbox; HasCheckbox=false
  - An entries: prefer no checkbox; legacy `- [ ] An` form still accepted with HasCheckbox=true
- ParseTraceability(path, commentPattern, commentCloser): scan code file for CRC: comments using the provided pattern; strips commentCloser from refs; stops each section at next `|` delimiter; extracts Rn refs from optional third section -> Traceability

## Collaborators
- os: file reading
- regexp: pattern matching
- bufio: line-by-line scanning
- Project: provides comment patterns and closers for file extensions

## Sequences
- seq-parse.md
