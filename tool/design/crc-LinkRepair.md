# LinkRepair
**Requirements:** R463, R464, R465, R466, R467, R489

Repairs the links a carve's move broke, both directions: for every `missing` link in the
population, tries the four sibling relocations a move between `carves/` and `carves/done/`
can produce, and rewrites the destination when exactly one resolves.

## Knows
- the repository root and the population: every tracked markdown file, or the files given
- per link considered: the citing file, line, old destination, and the outcome — rewritten
  to a new destination, unresolvable, or ambiguous

## Does
- `Repair(root, files)`: for each file, `ParseMarkdown`, classify each link with the
  `query links` classifier, and for each `missing` one compute the candidates
- `candidates`: re-base the citing directory `carves/` ⇄ `carves/done/`; insert or remove
  `done/` before the target's name; keep those that exist inside the root, deduplicated
- rewrites through `Markdown.SetDest` — the relocated target relative to the citing file's
  directory, fragment kept, `<…>` kept if present — then writes the file atomically
- `Report`: every link considered with its outcome, and the counts, zeros included;
  `Unrepaired()` for the exit status

## Constraints
- **Exactly one candidate.** None is not the move's doing; two is not the tool's to choose.
  Both are reported and left, and every class but `missing` is untouched
- **Bytes move only inside the parentheses**, and only in files with at least one rewrite
- **Idempotent**: a second run considers the same links, finds none missing, writes nothing
- **No git.** A relocation exists on disk or it does not; the class afterwards is the query's

## Collaborators
- minispecsdom.Markdown: `Links`, `SetDest`, `Render`
- query: `classify`, `PublicDocuments`
- parser.editFile: the atomic write with the read-back refusal
- CLI: `update repair-links [file...]`, exit status

## Sequences
- seq-links.md
