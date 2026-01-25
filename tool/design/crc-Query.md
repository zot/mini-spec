# Query
**Requirements:** R10, R11, R12, R13, R14, R15, R16, R17

Read-only operations that query parsed design data.

## Knows
- project: loaded Project instance
- requirements: parsed requirements
- crcCards: parsed CRC cards
- artifacts: parsed artifacts
- gaps: parsed gaps

## Does
- Requirements(): list all requirements with text and source
- Coverage(): map each Rn to design files that reference it
- Uncovered(): list Rn with no design references
- OrphanDesigns(): list CRC cards with no/empty Requirements field
- Artifacts(): list artifacts with checkbox states
- Gaps(): list gap items
- Traceability(path): check single file for CRC/Seq comments
- TraceabilityAll(): check all code files in Artifacts

## Collaborators
- Project: to locate files
- Parser: to parse design files

## Sequences
- seq-query.md
