# Query
**Requirements:** R10, R11, R12, R13, R14, R15, R16, R17, R79, R102, R185, R186, R187, R189, R191, R192, R193, R198, R199, R200, R201, R202, R203, R204, R317, R318, R319, R320, R321, R322, R326, R502, R503, R504, R505, R506, R507

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
- SelectAlarms(assessments, unverified): narrow the list to the states that carry a
  decision — everything that is not `verified`. **The list narrows and the census does
  not**: the closing count is computed over the whole population either way, so the
  filtered form is the full census minus the repetitions of *nothing to do here* and
  minus nothing else. A filtered count would answer *how many are wrong* and drop *out of
  how many*, which is the shape of the `| tail -2` workaround this flag replaces (R199)
- AlarmBriefs(assessments): for each selected alarm, resolve the design root relative to
  the repository root and look its document up in the Artifacts manifest for the test
  files, then ask Alarm to render the brief. Query holds the manifest; Alarm holds the
  wording (R200, R204)
- Coverage(): map each Rn to design files that reference it
- Uncovered(): list Rn with no design references
- OrphanDesigns(): list CRC cards with no/empty Requirements field
- Artifacts(): list artifacts with checkbox states
- Gaps(): list gap items, through the dependency's gaps reader (R326)
- ExpandGapRefs(args), SelectGaps(gaps, selection): the RANGE grammar — one type per range, a
  reversed range its low end — and the selection over it: nothing matched is an error naming
  the ask, partly unassigned is not, `--open`/`--closed` never claim a permanent gap, both
  flags mean every checkbox, document order always (R317, R318, R319, R320, R321, R322)
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
- Implementation(args, retired): the reverse lookup — where a requirement is implemented.
  Classify the args: a clean list of requirement refs (the `ExpandGapRefs` grammar, R-only)
  is number mode, otherwise the sole arg is a regexp over requirement text. Number mode
  prints code locations only; text mode prints each matched requirement (its Rn and one-line
  text) then its locations, with an explicit "no impl refs" for a match with none; retired is
  included by number and excluded by text unless `--retired` (R502, R503, R504, R505, R507)
- implRefs(): the positioned harvest — parse each Artifacts code file with sdom for its
  language (a per-extension map picks the `BracketLang`), run the traceability-comment reader,
  and collect each Rn with its `file:line` and comment, ranges expanded by the reader. Factored
  so Item 1 of the carve can promote it for validate (R506)

## Collaborators
- Project: to locate files
- Parser: to parse design files
- Trajectory: to read item IDs, which live outside the design root
- Alarm: assesses the alarms and renders a brief; Query supplies the manifest facts it
  needs and never the wording
- RepoRoot: to express the design root as a path a delegated agent can resolve in its own
  checkout, since an absolute path is wrong in a worktree
- TraceabilityComment: the minispecsdom reader that parses a code file's traceability comments
  into positioned, range-expanded refs; implRefs runs it per code file

## Sequences
- seq-query.md
