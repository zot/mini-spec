# Sequence: the ledger
**Requirements:** R396, R380, R381, R383

## 1. Reading

1. `ParseDone(src)`
   1.1. Parse with the markdown base
   1.2. For each `ListItem` at column 0: an entry if the text after it opens with bold,
        otherwise counted as entry-like
   1.3. The region runs to the next entry-like bullet at column 0 or a heading of
        level 2 or higher
   1.4. Derive: the header's date, the identifier slot between em dash and colon and
        its `#N`s, the title, the first backquoted commit, and the backquoted part
        pointer from the header or the body

## 2. Prepending

2. `Prepend(header, body)`
   2.1. Render `header`, a newline, the body, and a blank line as one synthetic text
   2.2. Inside a window, `Insert` it before the first entry's `ListItem`, or at the end
   2.3. Re-read the document
