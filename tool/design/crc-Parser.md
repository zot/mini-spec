# Parser
**Requirements:** R5, R6, R7, R8, R9, R51, R52, R53, R59, R61, R66, R67, R71, R73, R74, R75, R77, R90, R91, R94, R95, R96, R104

Parses mini-spec design file formats into structured data.

## Knows
- Requirement: {ID, Text, Sources []string, Inferred bool, Retired bool, Line int}
- SourceLineIssue: {LineNum int, Line string}
- CRCCard: {Name, Requirements []string, Sequences []string, Path string}
- Artifact: {DesignFile, CodeFiles []CodeFile}
- CodeFile: {Path, Checked bool, Line int}
- Gap: {ID, Type, Description, Resolved bool, HasCheckbox bool, Line int}
- Traceability: {CRCRefs []string, SeqRefs []SeqRef, ReqRefs []string}
- SeqRef: {File string, Fragment string} — Fragment is "" for file-only refs
- SeqDoc: {Path string, Items map[string]int (id -> line), Ks []int, Trees map[int]*SeqNode}
- SeqNode: {ID string, Children []*SeqNode}

## Does
- ParseRequirements(path): parse requirements.md -> []Requirement
  - Accepts strikethrough retired form `- **~~Rn:~~** (Retired Tk — see Rxxx) <text>`; sets Retired flag
  - Splits comma-separated paths on `**Source:**` line into Sources []string
- ScanSourceLineIssues(path): re-scan requirements.md for lines that look like Source markers but don't match the canonical `**Source:** ...` pattern -> []SourceLineIssue
- ParseCRCCard(path): parse crc-*.md -> CRCCard
- ParseArtifacts(path): parse design.md Artifacts section -> []Artifact
  - Supports inline format: `- [x] design.md → code.ts, code2.ts`
  - Skips subsection headers (`### CRC Cards`, etc.)
  - Parses comma-separated code files after `→`
  - Strips backticks from code file paths
- ParseGaps(path): parse design.md Gaps section -> []Gap (types: S/R/D/C/I/O/A/T)
  - Tn entries always have no checkbox; HasCheckbox=false
  - An entries: prefer no checkbox; legacy `- [ ] An` form still accepted with HasCheckbox=true
- ParseTraceability(path, commentPattern, commentCloser): scan code file for CRC: comments using the provided pattern; strips commentCloser from refs; stops each section at next `|` delimiter; extracts Rn refs from optional third section; also extracts Rn refs from a bare annotation — a comment whose first token after the comment leader is a requirement ref (prose mentions not leading with a ref are ignored); splits each Seq ref at `#` into SeqRef{File, Fragment} -> Traceability
- ParseSeqDoc(path): scan a sequence-diagram file for dotted-number items (allowing lane characters `│ ├ └ ─ |` before the number); group IDs by first segment K; build per-K trees -> SeqDoc

## Collaborators
- os: file reading
- regexp: pattern matching
- bufio: line-by-line scanning
- Project: provides comment patterns and closers for file extensions

## Sequences
- seq-parse.md
