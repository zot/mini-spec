# TraceabilityComment
**Requirements:** R337, R338, R339, R340, R341, R342, R343, R344, R345, R346, R347, R348

Mini-spec's anchor comment as **one node over the whole comment**, opener through
closer, one kind for every language. The first code that knows what a CRC card is,
in package `minispecsdom`, a sibling of `sdom`.

## Knows
- its children, through the embedded `Compound`: the original opener, the interior's
  glue and fields, the original closer
- typed views of the fields it found — `CRC`, `Seq`, `Test` (`*List`), `Refs`
  (`*RequirementList`), `Description` (`*Text`) — nil when absent

## Does
- `Parse(cmt, ctx)`: reads the comment through the context — closer, inner text —
  and fills itself **off to the side**; false when the interior is not one text node
  or the walk does not consume it
- the **segment walk**: split the interior on `|`, peel a description off the first
  separator that is not a field-key colon, run one `StencilBuilder` per segment with
  an alternation regex so each segment classifies itself by which group participated,
  hand list texts to `ParseList` / `ParseRequirementList`, splice the results flat
- `Comments(d, ctx)`: the pass — every opener whose group kind equals the language's
  `Comment.Kind` is tried, and each success replaces its run from opener to closer
  inside its own mutation window
- `New(lang, Fields)`: assembles the canonical interior and runs the same walk at a
  synthetic location inside synthetic markers from `lang.Comment`

## Constraints
- **It tiles the whole group and reuses the markers.** The context keys its links by
  node identity, so moving the original `*Opener` / `*Closer` inside keeps every link
  valid without a rebuild
- **Recognition is consumption.** Leading with a keyword is not enough: a comment
  reading *Test: a repaint frame round-trips (R3136).* leaves prose uncovered and is
  not one of these
- **One regex cannot express arbitrary order** — a repeated capture reports only its
  last iteration — hence the segment walk. It is a parsing device: the tree is flat
- **`New` has no origin.** A node from a separate `sdom.Parse` carries that parse's
  origin and can enter no other document; a synthetic node enters any. The interior
  walk is the shared path; only the prefix/suffix wrapping is not, and it is trivial
- **Nothing below this package knows what a CRC card is.** `sdom` supplies lists,
  stencils and the comment style; the keywords live here
- Association with a declaration is the consumer's job, not this node's

## Collaborators
- StencilBuilder: one per segment
- List, RequirementList: the fields
- BracketContext: closer, inner text, language, identity-keyed links
- BracketLang.Comment: the wrapping for `New` and the kind `Comments` matches
- Doc: the splice, inside a mutation window

## Sequences
- seq-anchor.md
