# FinishedCarve
**Requirements:** R469, R470, R471, R472, R473, R474

Moves a carve to its directory's `done/` and rewrites every link the move would break, both
directions, from where each link resolves now. Refuses, before any byte moves, to finish a
carve that is not finished.

## Knows
- the repository root, the carve's current and destination paths
- the incoming population: live carves, done carves, the three trajectory files
- per file: the rewrites planned (index, old, new) and the links left

## Does
- `FinishCarve(root, carve)`: checks the path sits directly in a carve directory and nothing
  is at the destination; reads the carve's status block and refuses on none or on any open
  part, naming them
- plans the carve's outgoing rewrites: each link that resolves now, re-expressed from `done/`;
  a link that does not resolve is recorded as left
- plans incoming rewrites: each link in each population document that resolves to the carve,
  re-expressed to the destination, fragment kept
- applies every plan in memory through `SetDest`, so a read-back failure refuses the whole
  move with nothing written; then writes the incoming documents in place, writes the carve's
  new content at the destination, and removes the old file
- `Report`: the move, each rewrite, each link left, the counts

## Constraints
- **Computed, never searched.** The destination is known, so a rewrite comes from the link's
  current target; the repair's exactly-one rule is never needed and cannot misfire here
- **Refusal precedes the rename**, and every refusal leaves every file as it was
- **Plain rename, nothing staged**: the tool never stages; git finds the rename at commit time
- **No design root, no git**

## Collaborators
- minispecsdom.Markdown: `Links`, `SetDest`, `Render`
- query: `ClassifyLink` (where a link resolves now), `LiveCarves`
- parser: the carve reader for the status block and its open parts; `EditFile` for the
  in-place writes
- CLI: `update finished-carve <carve>`

## Sequences
- seq-links.md
