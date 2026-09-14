# CarveSdom
**Requirements:** R361, R362, R363, R364, R365, R366, R368, R369, R391, R392, R394, R395, R404, R409, R411, R415, R417, R445

The carve file schema: embeds the markdown base, owns the document, and adds the
status block. The first of the four trajectory file schemas.

## Knows
- its `Doc`, the markdown parser and its contexts
- the status heading, or that there is none
- its parts, in order, each a `PartLine` with a depth, a parent and its line at parse time
- its stateless status lines

## Does
- `ParseCarve`: parses with the base, finds the status region, turns the list items
  inside it — and only those — into part lines, splitting them into parts and
  stateless lines, deriving depth and parent
- `Part(key)`, `Parts`, `Stateless`, `HasStatus`, `Render`; `Unread`: every group the context
  reports open at end of input, at its opener's line
- `SetMarker(key, …)`: the tool's rule on the keyed line — replace the first transient
  (`OPEN` or `REVERTED`), remove other transients, and when none insert before the trailing
  prose
- `Land(key, …)`: box, strike, and a `LANDED` record, in one act — refused with
  `ErrLanded` when the box is already checked
- after either write, parses the render afresh and reads the marker back on the keyed
  part — for `Land` the checked, struck box too — or panics with `ReadBackError`

## Constraints
- **Bounded by the region.** Other checkbox lists in a carve track other things
- **No checkbox means something**, and is recorded as stateless rather than dropped or
  given a state
- **Depth from whitespace, parent by derivation** — the flat array has no children
- **Writes touch the node's own children**; a write to an unknown key is an error
- **Refusal is decided before any marking**, so a refused write is byte-identical; a line
  carrying deviations refuses with a `DeviationError` naming each rule and target

## Collaborators
- schema.MarkdownParser, schema.Heading, schema.ListItem: the base
- PartLine, MarkerSpan: the lines and their markers
- BracketContext: for the part lines' pairing
- Doc: source offsets for depth, the splice

## Sequences
- seq-carve.md
