# Harvest
**Requirements:** R513, R514, R515, R502

The one reader of code files' traceability. `validate`, `query traceability` and `query
implementation` all consume it, so a ref counts in one exactly where it counts in the
others. It lives in `parser`, the path-taking layer over `minispecsdom`, which both
`validate` and `query` already import.

## Knows
- Harvest: {Files []FileHarvest, Unread []UnreadFile}
- FileHarvest: {Path, Comments []Comment}
- HarvestComment: {Line, Text, CRC []string, Seq []string, Refs []string} — Refs expanded, `R5-R8` as
  four entries, in the `Rn` spelling every other check uses; Text is the comment as written,
  opener through closer, its whitespace runs collapsed to one space so it prints on one line,
  for `query implementation` to show where a ref was written (R502)
- UnreadFile: {Path, Line, Reason} — Line is 0 when the whole file went unread

## Does
- HarvestFile(root, path, configured): read the file, pick its table by extension —
  configured first, then built in — parse, run the
  traceability-comment pass, and keep every comment with its line. A file with no table
  returns unread, reason `no language for <ext>`. A parse that leaves a string or comment open at
  end of input — an opener in the context's `Unclosed()` whose group is restricted — keeps
  what it read and is also listed unread, at that opener's line, since everything after it
  was never searched. Unpaired closers and unclosed code brackets hide no comment and are
  not reported
- HarvestArtifacts(root, artifacts, configured): HarvestFile over every code file the manifest lists,
  each once, in manifest order. A listed file that does not exist is not the harvest's
  concern: `validate` reports it as a missing artifact

## Collaborators
- Languages: the table for each extension, given the project's configured languages
- Project: supplies the configured languages, already layered and checked
- TraceabilityComment: the pass and the comment's typed fields
- Artifact, CodeFile: the manifest, from `ParseArtifacts`

## Sequences
- seq-harvest.md
