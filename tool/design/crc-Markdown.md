# Markdown
**Requirements:** R448, R449, R450, R451, R452, R453, R454

The plain-markdown reader: the base as a document of its own, for the one question no
schema reader asks — where does this document point? It owns the DOM of any markdown file
and reads its inline links; it writes nothing.

## Knows
- its `Doc`, the markdown parser and its context (the embedded base)
- its links, in document order, each with text, destination, path, fragment, image flag,
  line and offset at parse time

## Does
- `ParseMarkdown`: parses with the base, then scans the source for the inline form —
  `[text](dest)`, `![alt](dest)` — skipping every `[` the base places inside a code group,
  and binds each match to its offset
- `Links`, `Doc`, `Render`; `Unread`: every group open at end of input or closing nothing
- splits `Dest` into `Path` and `Fragment` at the first `#`; unwraps `<…>`; strips a
  trailing quoted title

## Constraints
- **A code group's bytes hold no link.** The test is the base's `inCode`, structural rather
  than a scan for fence markers — the reason the checker waited for a DOM
- **The inline form only.** Reference links, autolinks and bare URLs are not read; an
  unclosed `[` or `(` is text (measured 2026-08-16: reference links scored zero over 425 files)
- **Byte-exact render, proved over the real corpus** — this repository's documents and a
  sibling `ark` — with the count reported, never a fixture alone
- **No judgment.** Nothing is resolved; the file system and git are the classifier's

## Collaborators
- schema.MarkdownParser: the base
- BracketContext: `Enclosing` for the code-group test, and the unbalanced report
- markdownDoc: `parseBase`, `inCode`, `Render`, shared with TestDoc

## Sequences
- seq-links.md
