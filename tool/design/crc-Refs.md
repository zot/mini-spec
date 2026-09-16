# Refs
**Requirements:** R495, R496, R497

The inventory of references: every link and every pointer in the owned documents, where
each stands and where it resolves. The classifier is [Links](crc-Links.md); this passes no
judgment, and it is the query the move is specified against.

## Knows
- the repository root and the owned population: every tracked markdown file plus every
  document the trajectory layer sites
- per reference: the citing file, line, text as written, kind (`link`, `pointer`), the
  resolved repository-relative path or none, and the reader's index for a rewrite

## Does
- `OwnedDocuments`: the tracked documents, then the trajectory files and every `*.md` in
  `carves/`, `.carves/` and their `done/` that exist, deduplicated
- `RefsIn(root, file, doc)`: the links through the classifier's resolution, the pointers
  relative to the citing directory — or from the root in a trajectory file — ordered by line
- `Refs(root, files, to)`: over the files or the owned population, narrowed to those that
  resolve to `to` when given
- `FormatRefs`: one line each and a count

## Constraints
- **No judgment**: no severity, no git class; a reference resolves or it does not
- **A pointer in a trajectory file resolves from the repository root**, which is how every
  `Part` pointer is written
- **The move rewrites exactly what `--to` lists** — the same `RefsIn` feeds both

## Collaborators
- minispecsdom.Markdown: `Links`, `Pointers`
- Links: `classify`, `PublicDocuments`
- parser: the carve directory names
- CLI: `query refs [--to <path>] [file...]`

## Sequences
- seq-links.md
