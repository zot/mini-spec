# Query
**Requirements:** R10-17, R79, R102, R185-187, R189, R191-193, R198-204, R317-322, R326, R502-507, R518, R519, R531, R538, R539, R549

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
- SelectCoverage(cov, ids): the entries `query coverage` prints — every requirement, or only the
  selected IDs (classified by `ClassifyImplArgs`, number mode only) — ascending, each with its
  design files by base name; a selected ID no requirement carries is marked unknown (R538, R539)
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
- Traceability(path): check one code file through the harvest (`HarvestFile`); a file it
  cannot read reports the reason rather than a missing comment (R519)
- TraceabilityAll(): the harvest over every code file in Artifacts, its unread list included
  (R519)
- CommentForms(): for every extension a table reads — built in or configured — the comment
  style to write, the other comment forms the table accepts, and whether the written form has
  a closer, for `query comment-patterns` to warn on (R518)
- NextID(class): the next free identifier for `item`, `gap` or `req`. The class picks
  the root as well as the count — `item` is repository-scoped and delegates to
  Trajectory, `gap` and `req` are design-scoped and read what is already parsed (R189,
  R191). `gap` answers for **every** gap type, since numbering runs a separate sequence
  per type (R192), and `req` counts retired requirements, whose numbers are permanently
  taken (R193); `T` also counts the `(Retired Tn …)` markers in requirements.md, the rule
  `retire` mints by (R549)
- Implementation(args, retired): the reverse lookup — where a requirement is implemented.
  Classify the args: a clean list of requirement refs (the `ExpandGapRefs` grammar, R-only)
  is number mode, otherwise the sole arg is a regexp over requirement text. Number mode
  prints code locations only; text mode prints each matched requirement (its Rn and one-line
  text) then its locations, with an explicit "no impl refs" for a match with none; retired is
  included by number and excluded by text unless `--retired` (R502, R503, R504, R505, R507)
- Implementation reads the one harvest `validate` reads (`HarvestArtifacts`), indexing each
  Rn to its `file:line` and comment, so a ref counts in both exactly alike (R506)
- ClassifyImplArgs(args), SelectImplementation(reqs, harvest, sel): the two pure halves of
  Implementation, so neither needs a project on disk. The classifier returns number mode with
  the IDs sorted and deduplicated, or text mode with the compiled pattern; more than one arg
  that is not a clean ref list, or a pattern that does not compile, is an error naming it. The
  selection keeps one entry per selected requirement, ascending, each with its locations in
  manifest then line order — an entry with none is kept, empty, never dropped (R503, R504, R505).
  In number mode an ID no requirement carries is marked unknown, its locations still listed
  (R531)

## Collaborators
- Project: to locate files
- Parser: to parse design files
- Trajectory: to read item IDs, which live outside the design root
- Alarm: assesses the alarms and renders a brief; Query supplies the manifest facts it
  needs and never the wording
- RepoRoot: to express the design root as a path a delegated agent can resolve in its own
  checkout, since an absolute path is wrong in a worktree
- Harvest: the one reader of code files' traceability, shared with Validate
- Languages: the tables each extension reads with, for CommentForms

## Sequences
- seq-query.md
