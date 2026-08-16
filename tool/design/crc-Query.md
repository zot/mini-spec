# Query
**Requirements:** R10, R11, R12, R13, R14, R15, R16, R17, R79, R102, R185, R186, R187, R189, R191, R192, R193, R198

Read-only operations that query parsed design data.

## Knows
- project: loaded Project instance
- requirements: parsed requirements
- crcCards: parsed CRC cards
- artifacts: parsed artifacts
- gaps: parsed gaps

## Does
- Requirements(): list all requirements with text and source
- Alarms(): list every recorded fire alarm with its freshness state and close with the
  census. Asked rather than emitted: it carries the two states that stay non-zero for
  months, which validate deliberately omits (R185, R186, R187)
- Coverage(): map each Rn to design files that reference it
- Uncovered(): list Rn with no design references
- OrphanDesigns(): list CRC cards with no/empty Requirements field
- Artifacts(): list artifacts with checkbox states
- Gaps(): list gap items
- Migrations(): list specs/migrations/*.md (non-recursive, excludes complete/)
- UnindexedSpecs(): list specs/*.md not referenced in specs/index.md (exact .md-token match; all specs when index absent)
- Traceability(path): check single file for CRC/Seq comments (passes pattern+closer from Project)
- TraceabilityAll(): check all code files in Artifacts
- CommentPatterns(): return configured comment patterns map
- CommentClosers(): return configured comment closers map
- NextID(class): the next free identifier for `item`, `gap` or `req`. The class picks
  the root as well as the count — `item` is repository-scoped and delegates to
  Trajectory, `gap` and `req` are design-scoped and read what is already parsed (R189,
  R191). `gap` answers for **every** gap type, since numbering runs a separate sequence
  per type (R192), and `req` counts retired requirements, whose numbers are permanently
  taken (R193)

## Collaborators
- Project: to locate files
- Parser: to parse design files
- Trajectory: to read item IDs, which live outside the design root

## Sequences
- seq-query.md
