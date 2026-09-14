# PartLine
**Requirements:** R349, R350, R351, R353, R354, R355, R358, R359, R360, R392, R393, R397, R398, R409, R410, R417

A carve status line as one node over a markdown-parsed document: bound checkbox and
key, a list of marker spans, and the strike behind an accessor. The first
`minispecsdom` schema piece over the markdown base.

## Knows
- its children, through the embedded `Compound`: the base's `ListItem`, its `Checkbox`
  if any, the head's markers and re-cut texts, the marker spans, interspersed text
- typed views: `Checkbox()` (the base's node, nil when absent), `Key()` — the fragment,
  `4` or `2.2`, with `Item ` left to the display — `Title()`,
  `Markers()`, `Deviations()`
- the head's bold opener, so `IsStruck`/`Strike` can find and wrap it

## Does
- `Parse(item, ctx)`: walks the line's nodes from the `ListItem` to the newline,
  classifies each bold run by content — head first, then markers — re-cuts the head's
  first text into key, separator and title, builds `MarkerSpan`s over marker runs, and
  keeps everything else as it is; records deviations
- `Splice(d)`: replaces the run with itself inside a mutation window
- `IsStruck()`: whether the head sits inside a `~~` group
- `Strike(on)`: inserts or removes the `~~` pair around the head's bold run
- `SetMarker(verb, attribution)`: the tool's rule among its own children, after two
  refusals — a `DeviationError` when the line carries deviations, `ErrReopen` when the
  verb is `OPEN` and the checkbox is checked
- `PartLines(d, ctx)`: every list item, parsed and spliced

## Constraints
- **Every list item parses.** An unkeyed line is a node with deviations, not a
  failure; the checkbox still counts. Which lines belong to a status block is the
  carve schema's call
- **Content decides the runs**, the way the tool reads them today: the first bold run
  is the head, a later `VERB (…)` run is a marker
- **Flexible on input, rigid on output.** An `OPEN` attribution reads in any case, spacing
  and with an optional stop; a superseded scheme — the comma form — is a deviation, since it
  is not a marker and would otherwise go unreported
- **Reuse the base's nodes.** `ListItem`, `Checkbox`, every `**`, `~~` and code-span
  marker are the very objects the parse emitted; only interior texts are re-cut
- **The strike is a structural edit of the node's own children.** The flat array never
  sees inside a compound, so no mutation window is involved; the accessor is what a
  consumer sees

## Collaborators
- schema.ListItem, schema.Checkbox: the base's line-head markers
- MarkerSpan: the markers it holds
- BracketContext: pairing for `**` and `~~`, the document
- Doc: the splice and the strike

## Sequences
- seq-partline.md
