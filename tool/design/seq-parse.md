# Sequence: Parse Requirements

```
Caller -> Parser: ParseRequirements(path)

Parser -> os: ReadFile(path)
Parser -> bufio: NewScanner(content)

Parser -> Parser: state = seekingFeature
loop each line
    alt line matches "## Feature:"
        Parser -> Parser: currentSource = nil
        Parser -> Parser: state = seekingSource
    else line matches "**Source:**"
        Parser -> Parser: currentSource = extracted path
    else line matches "- **Rn:**"
        Parser -> Parser: extract ID, text, inferred flag
        Parser -> Parser: append Requirement{ID, text, currentSource, inferred, lineNum}
    end
end

Parser --> Caller: []Requirement
```

# Sequence: Parse CRC Card

```
Caller -> Parser: ParseCRCCard(path)

Parser -> os: ReadFile(path)
Parser -> Parser: extract name from "# Name" line
Parser -> Parser: find "**Requirements:**" line
Parser -> Parser: split on comma, trim whitespace
Parser --> Caller: CRCCard{Name, Requirements, Path}
```

# Sequence: Parse Sequence Document

```
Caller -> Parser: ParseSeqDoc(path)

Parser -> os: ReadFile(path)
Parser -> bufio: NewScanner(content)

loop each line
    Parser -> Parser: strip leading whitespace and lane chars (│ ├ └ ─ |)
    alt remaining text begins with dotted-number token
        Parser -> Parser: extract id (1, 1.1, 1.1.1, etc.), strip trailing dot
        Parser -> Parser: append to Items[id] = lineNum
    end
end

Parser -> Parser: group ids by first segment K -> Trees[K]
Parser -> Parser: build child trees per K from sorted ids

Parser --> Caller: SeqDoc{Path, Items, Ks, Trees}
```

# Sequence: Parse Artifacts

```
Caller -> Parser: ParseArtifacts(designMdPath)

Parser -> os: ReadFile(path)
Parser -> Parser: find "## Artifacts" section
Parser -> Parser: state = seekingDesignFile

loop each line until next ##
    alt line matches "^- " (design file)
        Parser -> Parser: currentArtifact = {DesignFile: name}
    else line matches "^  - [x]" or "^  - [ ]"
        Parser -> Parser: extract path, checked state
        Parser -> Parser: append to currentArtifact.CodeFiles
    end
end

Parser --> Caller: []Artifact
```
