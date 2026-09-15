# Sequence: reading a document's links, classifying them, repairing a move, and finishing a carve
**Requirements:** R448, R449, R450, R451, R453, R455, R456, R457, R459, R460, R462, R463, R464, R465, R466, R467, R469, R470, R471, R472, R473, R474

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

## 4. Finishing a carve

4. `FinishCarve(root, carve)`
   4.1. Refuse unless the path is directly in `carves/` or `.carves/`; refuse when the
        destination `<dir>/done/<name>` exists
   4.2. Read the carve's status block; refuse on none, or on any open part, naming each
   4.3. Outgoing plan: for each link in the carve, `ClassifyLink`; one that resolves inside
        the tree is rewritten to reach the same target from `done/`, fragment kept; one that
        does not is recorded as left
   4.4. Incoming plan: for each population document (live carves, done carves, the three
        trajectory files) and each of its links, the one whose resolved path is the carve is
        rewritten to reach the destination from that document's directory
   4.5. Apply every plan in memory through `SetDest` — a read-back panic refuses the whole
        move before any file is written
   4.6. Write the incoming documents in place, write the carve's content at the destination,
        remove the old file; report the move, the rewrites, the links left, the counts
