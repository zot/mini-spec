# Current
**Requirements:** R385, R386, R387, R388, R389, R390, R405, R406, R411, R416, R445

The current file schema: embeds the markdown base, owns the document, and adds the
`## Active` region — the only region a tool may write.

## Knows
- its `Doc`, the markdown parser and its contexts
- the Active heading and the node range of its region
- the standing level-2 headings

## Does
- `ParseCurrent`: parses with the base; finds exactly one level-2 `Active` heading or
  refuses with `ErrNoActive` / `ErrManyActive`; bounds its region at the next heading of
  level 2 or higher
- `Active`, `Occupied`, `Standing`, `Render`; `Unread`: every group the context reports open
  at end of input, at its opener's line
- `SetActive(body)`: refuses when occupied; otherwise replaces the region's body
- `Reset`: replaces the region's body with the placeholder
- after either write, reads the body or the placeholder back, or panics with `ReadBackError`
- the replacement: split the heading's text after the title line, remove the body's
  nodes, insert one synthetic text, re-read

## Constraints
- **Exactly one Active**, refused otherwise — never picked
- **The region ends at the next `##` or higher**, so standing context is never inside it
- **A write addresses no node outside the region**; the guarantee is structural
- **Refuse over a held item**; the pending file is the stack

## Collaborators
- schema.MarkdownParser, schema.Heading: the base
- Doc: `Split`, `Remove`, `Insert`

## Sequences
- seq-current.md
