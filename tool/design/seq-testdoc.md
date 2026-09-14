# Sequence: reading a test design and writing an alarm field
**Requirements:** R418, R419, R420, R421, R422, R423, R425, R426, R427, R428

## 1. Reading

1. `ParseTestDoc(src)`
   1.1. Parse with the markdown base
   1.2. For each level-2 `Heading`: the title is the rest of the heading's source line after
        the marker, whatever markers it carries; one not beginning `Test:` is unread
   1.3. The region ends at the next heading of level 2 or higher, or the end
   1.4. Render the region's nodes to lines, noting which lines fall inside a code group
        (`Enclosing` on the node, group kind `code`)
        1.4.1. A line at a field head outside code opens a field; its body folds to the
               next field head or the region's end
        1.4.2. A named field seen twice is a deviation; the first is kept
   1.5. Derive the values: sites split on commas, the leading date on `Pulled`, the
        integer on `Alarm`; a malformed value is a deviation
   1.6. Append the context's unclosed openers and unpaired closers; order by line

## 2. Writing

2. `SetPulled(n, date, body)`, `SetInject(n, sites, void)`, `NumberAlarms()`
   2.1. Find the entry by number — `ErrNoAlarm`; refuse with a `DeviationError` over
        deviations; `SetInject` refuses `ErrNoInject` and `ErrEmptyInject`
   2.2. Compute the new line text from the old field's lines
        2.2.1. `SetPulled`: fold the old content as ` *Earlier —* …`, or choose the insertion
               point after `Inject`, else after `Fire alarm`
        2.2.2. `SetInject` with `void`: rewrite the `Pulled` line to the history shape naming
               the old sites
        2.2.3. `NumberAlarms`: next number is max+1; insertion is above `Fire alarm`
   2.3. Inside `Mutate`: split the region's text at the field's line bounds and replace the
        middle, or insert a synthetic text at the insertion point
   2.4. Re-read the render; read the field back on the numbered entry or panic with
        `ReadBackError`
