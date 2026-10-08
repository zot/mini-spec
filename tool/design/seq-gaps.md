# Sequence: reading the gaps section and writing an entry
**Requirements:** R429-432, R434-436, R550

## 1. Reading

1. `ParseGaps(src)`
   1.1. Parse with the markdown base
   1.2. Find the level-2 heading whose text is `Gaps`; none means no section. The region
        ends at the next heading of level 2 or higher, or the end
   1.3. Render the region's nodes to lines and walk them
        1.3.1. A blank line closes the open entry's body
        1.3.2. A line inside a code group is skipped
        1.3.3. A keyed bullet opens a gap: depth from its whitespace, parent the nearest
               preceding shallower gap, deviations for a checkbox on a permanent letter, none
               on a tracked one, or a repeated ID
        1.3.4. An indented un-keyed bullet under an open gap is a sub-item; any other bullet
               is unread
        1.3.5. Any other line folds into the open sub-item or the open gap's text
        1.3.6. A `###` line closes the open entry and opens a subsection with its title; every
               non-blank line moves the current region's end — the head's before the first
               `###`, a subsection's after it   // R551
   1.4. Append the context's unclosed openers and unpaired closers; order by line

## 2. Writing

2. `Add(id, text)`, `Resolve(id)`, `Approve(id, newID)`
   2.1. Find the entry by ID — `ErrNoGap`; refuse with a `DeviationError` over deviations
   2.2. `Add`: `ErrBadGapID`, `ErrNoSection`, `ErrGapExists`; the line by letter; the insertion
        point after the last gap's span, or after the heading's line, when the section has no
        subsections
        2.2.1. With subsections: after the last gap of the same letter; else the end of the
               subsection whose heading names the type; else the end of the head   // R550
   2.3. `Resolve`: `ErrPermanent`, `ErrResolved`; the head line with `[x]`
   2.4. `Approve`: `ErrBadGapID` unless an `A` number, `ErrGapExists`, `ErrPermanent`; the head
        line rewritten `- A<n>: <head text>` at its depth
   2.5. Inside `Mutate`: replace the head line's span, or insert at the point — after a newline
        of its own when the point ends no line (`insertLine`)
   2.6. Re-read the render; read the entry back or panic with `ReadBackError`
