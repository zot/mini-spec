# Links
**Requirements:** R455, R456, R457, R459, R460

Resolves and classifies every link the Markdown reader found, against the file system
and git. The checker `carves/reference-discipline.md` was opened for: a public document
citing a private file hands a cloner a dangling link.

## Knows
- the repository root, and the citing files (given, or the live carves)
- per link: the citing file, line, link as written, resolved path, class

## Does
- `Check(root, files)`: for each file, parse with `Markdown`, resolve each link relative to
  the file's directory with the fragment removed, and classify
- classifies `external` (a scheme) and `local` (fragment only) before touching the disk;
  `outside` when the resolved path leaves the root or was absolute; `missing` when nothing
  is there; then `ignored` by `check-ignore`, `tracked` by the index, else `untracked`
- `Report`: the lines carrying a decision (or every link with `all`), and the count per class,
  zeros included; `Errors()` says whether any link is an error

## Constraints
- **Exactly one class per link**, decided in the order above so a missing file is never asked
  of git and a URL never of the disk
- **Severity is the 2026-08-04 decision**: ignored, missing, outside are errors; untracked is
  a warning; the rest carry none
- **Git through the command line, batched per citing file**; no working tree is a refusal,
  never a clean report
- **A directory resolves when it exists** and is tracked when git holds anything under it

## Collaborators
- minispecsdom.Markdown: the links
- Git: `IsRepo`, `Ignored`, `Tracked`
- parser.ScanCarves: the default population
- CLI: `query links [--all] [file...]`, exit status, `--json`

## Sequences
- seq-links.md
