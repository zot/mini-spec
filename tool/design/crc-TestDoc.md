# TestDoc
**Requirements:** R418, R419, R420, R421, R422, R423, R424, R425, R426, R427, R428, R445, R446

The test-design file schema: embeds the markdown base, owns the document, and adds the
`## Test:` entry with its five alarm fields. The first reader over a design document rather
than a trajectory file.

## Knows
- its `Doc`, the markdown parser and its context
- its entries, in order, each with title, the five fields, its deviations and its line at parse time
- which level-2 headings did not read as tests

## Does
- `ParseTestDoc`: parses with the base, finds each `## Test:` heading and its region, reads the
  field lines inside it — outside any code group — folding bodies to the next field line
- `Tests`, `Alarm(n)`, `Render`, `Doc`; `Unread`: non-test headings, deviant entries, and every
  group the context reports open or closing nothing, ordered by line
- `SetPulled(n, date, body)`: replace with the old content folded as history, or insert after
  `Inject`, else after `Fire alarm`
- `SetInject(n, sites, void)`: rewrite the sites; when `void`, demote the `Pulled` line to the
  history shape naming the old sites, in the same write
- `NumberAlarms`: `**Alarm:** <n>` above every unnumbered alarm's `Fire alarm`, from max+1
- after every write, re-reads the render and reads the write back on the numbered entry, or
  panics with `ReadBackError`

## Constraints
- **Bounded by the entry.** A write edits inside one entry's region and no byte outside it
- **A code group's lines are body**, read side and write side
- **Absence is data.** No `Pulled` is a prescription; no `Alarm` is unmigrated; neither is filled in
- **Refusal is decided before any marking**: `ErrNoAlarm`, `ErrNoInject`, `ErrEmptyInject`, and a
  `DeviationError` naming each doubled or malformed field
- **No judgment.** Sites are recorded, not resolved; freshness is the census's

## Collaborators
- schema.MarkdownParser, schema.Heading: the base
- BracketContext: `Enclosing` for the code-group test, and the unbalanced report
- Doc: `Split`, `Insert`, `Replace`, `Remove` inside `Mutate`; `Line`
- Deviation, DeviationError, Unread, mustReadBack: shared with the trajectory readers

## Sequences
- seq-testdoc.md
