# Sequence: the queue file
**Requirements:** R371, R372, R373, R375, R399, R400, R401, R407, R408, R412

## 1. Reading

1. `ParsePending(src)`
   1.1. Parse with the markdown base
   1.2. For each level-2 `Heading`: if the text after it opens `N.`, an entry begins;
        otherwise it is listed as unread, with its line
   1.3. The region runs to the next level-2-or-higher heading, or to a text holding a
        `---` that is a whole document line outside a code group, where the run is cut
   1.4. Derive the values from the run's rendered bytes: the heading line, the
        `Source:` line, the `Next:` line
        1.4.1. The `Source:` line reads as `part` or `gap` by its word; a part key drops
               its `#`, a gap key must be one gap ID; neither is `SourceNone`
        1.4.2. A `Source:` line that read as neither is listed as unread, with its line

## 2. Placing

2. `Place(e, pos)`
   2.1. Refuse a position outside `1 … len+1`
        2.1.1. Refuse a gap source whose key is not one gap ID
   2.2. Render the canonical entry text, ending in a blank line, in the part or gap form
        by `Kind`
   2.3. Inside a window, `Insert` it as one synthetic text before the entry at `pos`, or
        where the entries end; re-read the document and re-derive the entries
        2.3.1. When a rule closes the region, split the last entry's tail at the rule and
               insert before it
        2.3.2. Else at end of file: open one blank line before it and drop the text's own
               trailing blank line, so the file ends in one newline
        2.3.3. With no entries, after the header's rule, one blank line between

## 3. Removing

3. `Remove(id)`
   3.1. Find the entry; unknown is an error
   3.2. Inside a window, split the shared tail text at the region's end if the run
        does not end on a node boundary, then `Remove` every node of the run
        3.2.1. If the file now ends in a blank line, trim it to one newline — the
               separator placement opened
   3.3. Re-read the document and re-derive the entries
