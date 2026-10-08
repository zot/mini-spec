# Gaps
**Requirements:** R429-432, R434-436, R445, R540, R550, R551

The gaps schema: embeds the markdown base, owns the document, and reads the `## Gaps`
region of a design document as typed, numbered entries. The second design-document reader,
on the same base as TestDoc.

## Knows
- its `Doc`, the markdown parser and its context, through `markdownDoc`
- the `Gaps` heading, or that there is none, and where the region ends
- its regions: the head before the first `###`, then each subsection with its title, and where
  each one's last non-blank line ends
- its entries, in order, each with type, number, checkbox state, folded text, sub-items,
  depth, parent, deviations and its line at parse time

## Does
- `ParseGaps`: parses with the base, finds the region, reads each bullet outside a code group
  as a gap, a sub-item, or an unread line, folding continuation lines
- `Items`, `Gap(id)`, `HasGaps`, `Render`, `Doc`; `Unread`: unkeyed column-0 bullets, deviant
  entries, and every group the context reports open or closing nothing, ordered by line
- `Add(id, text)`: one line after the last gap's span, or after the heading; checkbox by letter;
  on a line of its own, ending an unterminated previous line first. With `###` subsections,
  after the last gap of its letter, else at the end of the subsection whose heading names the
  type, else at the end of the head
- `Resolve(id)`: `[ ]` to `[x]` on the head line
- `Approve(id, newID)`: the head line rewritten permanent, everything beneath untouched
- after every write, re-reads the render and reads the write back on the addressed entry, or
  panics with `ReadBackError`

## Constraints
- **Bounded by the region.** Bullets elsewhere in the document are not gaps
- **A code group's lines are body**
- **Refusal is decided before any marking**: `ErrNoGap`, `ErrNoSection`, `ErrBadGapID`,
  `ErrGapExists`, `ErrPermanent`, `ErrResolved`, and a `DeviationError` naming each deviation
- **No minting, no filtering.** Numbers come from the caller; open/closed and ranges are the query's

## Collaborators
- markdownDoc: the base parse, `regionEnd`, `inCode`, `replaceSpan`, `insertLine`
- schema.Heading: the section heading
- Deviation, DeviationError, Unread, mustReadBack: shared with the other readers

## Sequences
- seq-gaps.md
