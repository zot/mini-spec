# Done
**Requirements:** R378, R379, R380, R381, R382, R383, R384, R395, R396, R402, R411, R414, R445

The done file schema: embeds the markdown base, owns the ledger, and adds the
completion entry as a view over a list-item region.

## Knows
- its `Doc`, the markdown parser and its contexts
- its entries, most recent first, each with its run and derived values
- the column-0 bullets that were entry-like but not entries, each with its line and text
- each entry's line at parse time

## Does
- `ParseDone`: parses with the base, walks column-0 `ListItem`s, and for each whose
  text opens with bold collects the run to the region's end and derives the values —
  the identifier slot's `#N`s, the part pointer, date, title, commit
- `Entries`, `MaxID`, `Unread`, `Render`; `Unread` also carries every group the context
  reports open at end of input, at its opener's line, last
- `Prepend(header, body)`: one synthetic text just after the rule, before the first
  entry or at the end; then re-reads the document and reads the entry back as the first,
  or panics with `ReadBackError`

## Constraints
- **A view, not a node**: values read from bytes at the format's positions
- **The slot is the only source of queue IDs.** Bodies quote items freely
- **A fence is the entry's**, by construction of the base
- **Say what could not be read**

## Collaborators
- schema.MarkdownParser, schema.ListItem, schema.Heading: the base
- Doc: `Insert`

## Sequences
- seq-done.md
