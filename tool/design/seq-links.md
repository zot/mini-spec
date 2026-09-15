# Sequence: reading a document's links, classifying them, and repairing a move
**Requirements:** R448, R449, R450, R451, R453, R455, R456, R457, R459, R460, R462, R463, R464, R465, R466, R467

## 1. Reading

1. `ParseMarkdown(src)`
   1.1. Parse with the markdown base
   1.2. Scan the source for `[` (or `![`) — the source rather than each `Text` node, since
        link text may hold emphasis, whose markers are nodes of their own
        1.2.1. A match whose opening bracket is inside a code group (`inCode`) is skipped
        1.2.2. The text runs to the matching `]`, one level of nested brackets allowed; a
               `(` must follow directly, and the destination runs to its `)`; otherwise text
        1.2.3. Unwrap `<…>`; strip a trailing quoted title; split path and fragment at `#`
   1.3. Bind the link to its offset in the document and derive its line
   1.4. Append the context's unclosed openers and unpaired closers; order by line

## 2. Classifying

2. `Check(root, files)`
   2.1. With no files, the population is `ScanCarves(root)`'s live carves
   2.2. Refuse when `root` is not a git working tree
   2.3. For each file: read, `ParseMarkdown`, and for each link
        2.3.1. A scheme → `external`; an empty path → `local`
        2.3.2. Resolve the path against the file's directory; absolute, or escaping `root`
               → `outside`
        2.3.3. Nothing on disk → `missing`
   2.4. Ask git once per file: `Ignored` over the surviving paths, then `Tracked` on each
        not ignored → `ignored`, `tracked`, else `untracked`
   2.5. Count every class; the report lists decisions, or everything under `all`

## 3. Repairing

3. `Repair(root, files)`
   3.1. With no files, the population is the live carves and every `*.md` directly under
        `carves/done/` (and the `.carves/` pair)
   3.2. For each file: `ParseMarkdown`, then `classify` every link; only `missing` ones go on
   3.3. Candidates for a missing link: the citing directory re-based `carves/` ⇄
        `carves/done/`; `done/` inserted before, or removed from before, the target's name;
        kept when something is on disk there inside the root
        3.3.1. Zero candidates → `unresolvable`, reported, untouched
        3.3.2. Two or more → `ambiguous`, reported, untouched
   3.4. Exactly one: the new destination is the candidate relative to the citing directory,
        slash-separated, fragment appended as written, `<…>` kept if the old one had it
   3.5. `SetDest(i, new)` inside `Mutate`: `replaceSpan` over the bytes between `(` and `)`,
        then re-read and read link `i` back with the new destination or panic `ReadBackError`
   3.6. A file with at least one rewrite is written atomically; the report lists every link
        considered and closes with the counts
