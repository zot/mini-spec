# MarkerSpan
**Requirements:** R352, R356, R357, R359, R397

`**VERB (attribution)**` as a stencil over the bold run the base emitted: the verb
bound, the attribution derived, and one canonical write.

## Knows
- its children: the `**` opener and closer, reused; the verb `Text`; glue; whatever the
  attribution spans — texts and code-span groups

## Does
- `Verb()`: the bound `Text`
- `Attribution()`: renders the nodes between the parentheses
- `QueueID()`: the `#N` inside the attribution, and whether there was one
- `Set(verb, attribution)`: rewrites the interior as one canonical text, verb in
  capitals, after re-parsing the render as a marker and requiring the same verb and
  attribution back; refuses otherwise with the literal unchanged

## Constraints
- **The attribution is derived, not bound**, because it crosses code spans — a value that
  spans nodes is read, never stored
- **A write is canonical and leaves one text** where a fresh parse would give code
  spans. Bytes identical, structure not: the corner a canonical write accepts, stated
- **Recognition is by content**: capitals, words joined by spaces or hyphens, an
  optional parenthesised attribution, an optional trailing period

## Collaborators
- PartLine: holds it
- StencilBuilder: cuts the interior
- BracketContext: the run's closer

## Sequences
- seq-partline.md
