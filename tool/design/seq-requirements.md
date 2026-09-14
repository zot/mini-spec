# Sequence: reading requirements.md and writing an entry
**Requirements:** R437, R438, R439, R440, R441, R442, R443, R444

## 1. Reading

1. `ParseRequirements(src)`
   1.1. Parse with the markdown base
   1.2. Render to lines and walk them
        1.2.1. A blank line closes the open entry
        1.2.2. A line inside a code group is content of the section, never a heading or entry
        1.2.3. A heading closes the open section's own content and opens a section, parent
               the nearest preceding shallower one
        1.2.4. A `Source:` line names the section's source; a second is unread
        1.2.5. A keyed bullet opens an entry (1.3); any other bullet is unread
        1.2.6. Any other line folds into the open entry and extends the section's last content
   1.3. The entry: struck or live; a struck head's clause read into `RetiredBy` and
        `Replacement` or a deviation when absent; a repeated ID a deviation
   1.4. Append the context's unclosed openers and unpaired closers; order by line

## 2. Writing

2. `Add(title, id, text)`, `Retire(id, tn, clause)`
   2.1. Refuse first: the ID's shape and presence, the title's count, the clause's shape, the
        entry's deviations and retirement
   2.2. `Add`: the line at the section's last content end
   2.3. `Retire`: the head line rewritten with the strike and the clause
   2.4. Inside `Mutate`: replace the head line's span, or insert at the point
   2.5. Re-read the render; read the entry back or panic with `ReadBackError`
