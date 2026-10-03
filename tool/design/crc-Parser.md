# Parser
**Requirements:** R310, R312, R316, R5, R6, R7, R8, R51, R52, R53, R61, R66, R71, R73, R74, R75, R77, R90, R91, R94, R95, R96, R178, R326, R532

Parses mini-spec design file formats into structured data.

## Knows
- Requirement: {ID, Text, Sources []string, Inferred bool, Retired bool, Line int}
- SourceLineIssue: {LineNum int, Line string}
- CRCCard: {Name, Requirements []string, Sequences []string, Path string}
- Artifact: {DesignFile, CodeFiles []CodeFile}
- CodeFile: {Path, Checked bool, Line int}
- Gap: {ID, Type, Description, Resolved bool, HasCheckbox bool, Line int}
- SeqRef: {File string, Fragment string} — Fragment is "" for file-only refs
- SeqDoc: {Path string, Items map[string]int (id -> line), Ks []int, Trees map[int]*SeqNode}
- SeqNode: {ID string, Children []*SeqNode}

## Does
- ParseTestDoc(path), ParseTestDocReport(path): the fire alarms a test design records, read
  through the dependency's `TestDoc` reader, with its unread list beside them; `Alarm` carries
  the entry's number, code files and line, `Key()` its `<doc>#<n>` name (R178, R310, R316)
- ParseRequirements(path): parse requirements.md -> []Requirement
  - Accepts strikethrough retired form `- **~~Rn:~~** (Retired Tk — see Rxxx) <text>`; sets Retired flag
  - Splits comma-separated paths on `**Source:**` line into Sources []string
- ScanSourceLineIssues(path): re-scan requirements.md for lines that look like Source markers but don't match the canonical `**Source:** ...` pattern -> []SourceLineIssue
- ParseCRCCard(path): parse crc-*.md -> CRCCard. The `**Requirements:**` field is read
  through `sdom.ParseRequirementList` from its head: each ref and every member of a range
  (`R5-8`, `R5-R8`) becomes an `Rn` in Requirements, in the field's order. Text the list does
  not consume is split on commas and kept as tokens as written, so validate reports it as an
  unknown reference rather than losing it (R532)
- ParseArtifacts(path): parse design.md Artifacts section -> []Artifact
  - Supports inline format: `- [x] design.md → code.ts, code2.ts`
  - Skips subsection headers (`### CRC Cards`, etc.)
  - Parses comma-separated code files after `→`
  - Strips backticks from code file paths
- ParseGaps(path): parse design.md Gaps section -> []Gap (types: S/R/D/C/I/O/A/T)
  - Tn entries always have no checkbox; HasCheckbox=false
  - An entries: prefer no checkbox; legacy `- [ ] An` form still accepted with HasCheckbox=true
- ParseSeqDoc(path): scan a sequence-diagram file for dotted-number items (allowing lane characters `│ ├ └ ─ |` before the number); group IDs by first segment K; build per-K trees -> SeqDoc

## Collaborators
- os: file reading
- regexp: pattern matching
- bufio: line-by-line scanning

## Sequences
- seq-parse.md
