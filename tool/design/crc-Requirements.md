# Requirements
**Requirements:** R437, R438, R439, R440, R441, R442, R443, R444, R445

The requirements schema: embeds the markdown base, owns the document, and reads every heading
as a section with its own content, its source, and its numbered entries in live and retired
form. The third design-document reader, on the same base as TestDoc and Gaps.

## Knows
- its `Doc`, the markdown parser and its context, through `markdownDoc`
- its sections, in order, each with title, level, source, parent, its own content span and
  the end of its last non-blank line
- its entries, in order, each with number, folded text, retirement, section, deviations and
  its line at parse time

## Does
- `ParseRequirements`: renders the document to lines and walks them — a heading outside a
  code group opens a section, a keyed bullet an entry, `Source:` names the source, other lines
  fold into the open entry
- `Sections`, `Section(title)`, `Requirements`, `Requirement(id)`, `Render`, `Doc`; `Unread`:
  unkeyed column-0 bullets, later `Source:` lines, deviant entries, and every group the context
  reports open or closing nothing, ordered by line
- `Add(title, id, text)`: one line at the end of the section's own content
- `Retire(id, tn, clause)`: the head line struck and claused, the body untouched
- after every write, re-reads the render and reads the write back on the addressed entry, or
  panics with `ReadBackError`

## Constraints
- **A section's content is its own.** A sub-heading ends it, so a new entry lands before the
  first sub-heading
- **A code group's lines are body**
- **Refusal is decided before any marking**: `ErrNoSection`, `ErrManySections`, `ErrBadReqID`,
  `ErrReqExists`, `ErrNoRequirement`, `ErrRetired`, `ErrBadClause`, and a `DeviationError`
- **No minting, no coverage, no `Tn` gap.** Those are the verb's and the gaps reader's

## Collaborators
- markdownDoc: the base parse, `inCode`, `replaceSpan`
- Deviation, DeviationError, Unread, mustReadBack: shared with the other readers

## Sequences
- seq-requirements.md
