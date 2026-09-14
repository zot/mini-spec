# Sequence: reading and writing a part line
**Requirements:** R349, R351, R352, R354, R357, R358

## 1. Parsing one line

1. `Parse(item, ctx)` on a `ListItem` the base emitted
   1.1. Collect the run: from the item to the last node before the text that begins
        the next line — the run's nodes are reused, never copied
   1.2. If the next node is a `Checkbox`, keep it as the line's checkbox
   1.3. Walk the run for bold openers, pairing each through `ctx`
        1.3.1. The first bold run is the **head**: cut its first text with a stencil
               into key, separator glue and title; note whether a `~~` encloses it
        1.3.2. A later bold run whose first text reads `VERB (` — or `VERB` alone — is
               a **marker**: a `MarkerSpan` over opener through closer
        1.3.3. Anything else stays as it was
   1.4. Record deviations: unkeyed, wrong separator, checkbox interior, verb case, an
        `OPEN` attribution off its two forms
   1.5. Tile: the compound is the run with the head's text and the markers replaced

## 2. Striking

2. `Strike(true)` on a line whose head is not struck
   2.1. Insert a `~~` opener before the head's bold opener and a `~~` closer after its
        bold closer, among the node's own children — synthetic nodes, no origin; the
        flat array is untouched, so no window
   2.2. `IsStruck` now reads true, because the head is enclosed

## 3. Setting a marker

3. `MarkerSpan.Set(verb, attribution)`
   3.1. Render the canonical interior `VERB (attribution)`
   3.2. Re-parse it as a marker; require the same verb and attribution — else refuse
   3.3. Replace the interior children with the one text; bytes match a fresh render
