# Sequence: reading and writing traceability comments
**Requirements:** R342, R343, R344, R345, R346, R347

Three diagrams: the pass, one comment's parse, and a guarded list write.

## 1. The second pass

1. A consumer calls `Comments(d, ctx)` on a bracket-parsed document
   1.1. For each node in document order that is an `*Opener`, the pass resolves the
        group by the opener's bytes through `ctx.Language()` and compares its `Kind`
        to `Language().Comment.Kind` — two configured strings, no word spelled here
   1.2. A candidate gets a fresh `TraceabilityComment` and `Parse(opener, ctx)`
        1.2.1. false: the node is dropped and the walk advances
        1.2.2. true: inside one `Mutate` window, `Replace(opener, node)`, then `Remove`
               the interior text and the closer — they now live inside the node
   1.3. The pass returns every node it spliced, in document order

## 2. One comment parses itself

2. `Parse(cmt, ctx)`
   2.1. `ctx.Closer(cmt)` and the nodes between: exactly one `*Text` or nothing to do
   2.2. Split the interior on `|`; peel the description off the first `--`, `—` or `:`
        that does not immediately follow `CRC`, `Seq` or `Test`
   2.3. For each segment, a `StencilBuilder` over the alternation regex — `CRC:`,
        `Seq:`, `Test:` or a leading `R` — so the participating group names the field
        2.3.1. No match, or a field kind seen twice: return false
        2.3.2. A list group's text goes to `ParseList` or `ParseRequirementList`; the
               remainder must be empty, or return false; the node is `Put` in its slot
   2.4. The description, if any, becomes a `Text` the node keeps a pointer to
   2.5. Children: the original opener, every segment's children and the `|` glue
        between them flat, the description and its separator, the original closer
   2.6. If the walk did not consume the whole interior, return false

## 3. A guarded list write

3. `SetItems(items)` on a `List`
   3.1. Render the canonical literal
   3.2. Re-parse it with the same parser; require an empty remainder and equal items
        3.2.1. Refused: return the error, the literal untouched
   3.3. `SetText` on the one child — the field is altered, the comment is altered
        because a child is, and nothing else moved
