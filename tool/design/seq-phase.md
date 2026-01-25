# Sequence: Phase Commands

## minispec phase requirements (example)

```
User -> CLI: minispec phase requirements
CLI -> Project: Detect()
Project --> CLI: project

CLI -> Phase: RunRequirements(project)

Phase -> Parser: ParseRequirements(requirements.md)
Parser --> Phase: []Requirement

Phase -> Phase: check sequential numbering
Phase -> Phase: check spec sources exist

Phase -> Phase: build findings
Phase --> CLI: PhaseResult{findings, issues, passed}

CLI -> CLI: format output
CLI --> User: phase-specific output + exit code
```

## minispec phase design (example)

```
User -> CLI: minispec phase design
CLI -> Project: Detect()
Project --> CLI: project

CLI -> Phase: RunDesign(project)

Phase -> Parser: ParseRequirements(requirements.md)
Parser --> Phase: []Requirement

Phase -> Parser: ParseArtifacts(design.md)
Parser --> Phase: []Artifact

Phase -> Project: GlobCRCCards()
loop each CRC card
    Phase -> Parser: ParseCRCCard(path)
    Phase -> Phase: check Requirements field
    Phase -> Phase: check valid Rn refs
end

Phase -> Phase: compute coverage
Phase -> Phase: check all design files listed

Phase --> CLI: PhaseResult{findings, issues, passed}

CLI -> CLI: format output
CLI --> User: phase-specific output + exit code
```
