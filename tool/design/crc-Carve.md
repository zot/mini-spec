# Carve
**Requirements:** R207, R208, R209, R210, R211, R212, R214, R216, R217, R218, R219, R220

The carve documents at the repository root, read and written **through
`github.com/zot/simple-dom`'s `minispecsdom.Carve`**. This card is an adapter: it owns the
file system side — which files are carves, reading bytes, writing them back atomically — and
the census counts, and it owns nothing about what a part line is. The shapes are
`trajectory-format.md`'s; the rules that read them are the dependency's, stated in its
`part-line.md` and `carve-schema.md`.

*Why an adapter and not a port.* Every consumer on the branch this reclaims from — the backup
slot, the `pending` verbs, `query carves`, `validate trajectory` — called path-taking functions.
The dependency's readers are source in, `Render()` out, and deliberately do not write a file or
print a report. So the path-level signatures stay, and the seam where "listed by read paths,
refused by write paths" meets the file system is one place (R219, R220).

## Knows
- CarveScan: which carve directories existed and what each document contributed. Directories
  are reported present-or-absent because no directory at all has no answer, and an empty
  census is a confident wrong one (R214)
- Carve: one document — its repository-relative path, whether a status block was found, its
  parts, and its stateless lines. A document with no status block keeps its entry: the tool
  cannot tell a carve missing its block from a file that was never a carve (R211)
- Part: the dependency's `*minispecsdom.Part` — key, checkbox, markers, deviations, depth,
  line — with the census's derived view over it: `State` (open, landed), `QueueID` from the
  first `OPEN` marker's attribution, `Title`, and a JSON shape that names its fields for a
  reader without the Go source open
- Stateless: a status-block line with no checkbox — its 1-based line, derived from the node's
  offset in the source since the dependency numbers only parts, its head text so a reader can
  recognise it, and the deviations the part-line reader found on it (R216)

## Does
- ScanCarves(repoRoot): read every `*.md` directly in `carves/` and `.carves/`; `carves/done/`
  is not entered; both absent is `ErrNoCarveDirs` (R208, R214)
- ReadCarve(path, rel): read the bytes, `minispecsdom.ParseCarve`, and take `Parts()` and
  `Stateless()` as the reader hands them — the status region, the fence rule and the subpart
  depth are all the reader's (R209, R210)
- Open(), Landed(), Stateless(), NonConforming(), NoStatus(): the counts, per carve and across
  the scan, every one stated so a zero is evidence the check ran (R218)
- SetMarker(path, key, verb, attribution), SetPartLanded(path, key, attribution): read, apply
  `Carve.SetMarker` / `Carve.Land`, write by temp-file-and-rename. The reader's refusals —
  `DeviationError`, `ErrReopen`, `ErrLanded`, `ErrNoPart` — pass through unchanged, and no
  byte reaches the file on any of them (R220)
- PartIsLanded(path, key): the checkbox of the keyed part; false with no error when no part
  keys so, which the backup slot's release rule reads as "nothing to protect"

## Collaborators
- minispecsdom.Carve: the reader and writer over one document's bytes
- Project: to resolve the repository root
- CLI: renders the census (`printCarves`) and refuses nothing itself

## Sequences
- seq-carve-status.md
