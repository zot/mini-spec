# Carve: minispecsdom moves home — the mini-spec readers leave simple-dom

`minispecsdom` is mini-spec's document model: the readers and writers for carves, the
three trajectory files, test designs, gaps, requirements and traceability comments, built
over `github.com/zot/simple-dom`'s generic DOM. It was written in `~/work/mini-spec-tool`
because that is where the DOM was being built, and it stayed there because on 2026-09-04
the rule was *depend on the module, never copy its packages in*. That rule was about the
DOM. The readers are not the DOM; they know what a part line and a `**Pulled:**` field are,
and simple-dom should not. This carve moves them here, with their design, and leaves
simple-dom a library that knows nothing about mini-spec.

**Provenance.** Bill, 2026-09-14, in conversation. Supersedes the "never copy its packages
in" clause of the 2026-09-04 restart decision for `minispecsdom` only; `sdom` and
`sdom/schema` stay a module dependency behind the existing `replace`.

## Status

- [x] ~~**Item 1 — the package.**~~ **LANDED (`9c6c796`, 2026-09-14 — `#80`.)**
- [x] ~~**Item 2 — the design comes with it, renumbered.**~~ **LANDED (`9c6c796`, 2026-09-14 — `#80`.)**
- [x] ~~**Item 3 — gaps, requests, and the far side.**~~ **LANDED (`28c01dd`, 2026-09-14 — `#82`.)**
- [x] ~~**Item 4 — the alarms are re-pulled where their sites moved.**~~ **LANDED (`1c8395f`, 2026-09-14 — `#81`.)**

## Decisions

**DECIDED (Bill, 2026-09-14): the package lands at `tool/internal/minispecsdom`.** Internal,
not public: nothing outside this module should read a carve except through the CLI.

**DECIDED (Bill, 2026-09-14): the design artifacts are brought in as ordinary artifacts of
this project, with requirements renumbered into this project's sequence.** Not a second
design root, not a separate numbering. The renumbering is the one edit the skill calls
permanent and silent, so it is done by a mapping table, on a copy outside both trees, before
anything lands beside this project's own `R210`–`R336` — see Item 2.

**DECIDED (Bill, 2026-09-14): plain copy, no history import.** The moving commit names
mini-spec-tool's head at the time of the copy; the decisions are in its `carves/done/` and
the code is days old. A subtree import would carry a history that names files by paths that
no longer exist.

**What is measured (2026-09-14, on mini-spec-tool at `b9f4c70`):**

| the slice | count |
|---|---|
| Go files in `minispecsdom/` (with tests and `testdata/`) | 24 + 6 fixtures |
| files here importing `github.com/zot/simple-dom/minispecsdom` | 16 |
| specs (`traceability-comment`, `markdown`, `part-line`, and the seven `*-schema`) | 10 |
| CRC cards / sequences / test designs mapped to `minispecsdom/` | 10 / 9 / 9 |
| requirements in those features, of which retired | 111 (4) |
| distinct `Rn` cited from the Go sources | 132 |
| fire alarms (`**Inject:**` lines) in the nine test designs | 56 |
| open gaps in mini-spec-tool about this package | 3 (`O28`–`O30`) |
| request files this project sent that carry no `RESP-` | 6 |

The cards' `**Requirements:**` lines cite nothing below `R210` — the slice is closed under
its own references — and no sdom-side card cites a card in the slice. Both were checked,
because the renumbering is only safe if they hold. *Two corrections found while landing
(2026-09-14):* the slice is a **set, not a range** — 22 numbers inside 210–355 belong to
sdom features, appended there later — and the `markdown base` feature (`markdown.md`, 13
requirements) is simple-dom's own: implemented in `sdom/schema/markdown.go`, cited by
`crc-MarkdownParser.md`, and cited by nothing in `minispecsdom`. It stayed. So: nine specs,
111 requirements, four retired.

**Item 1** (the package). Copy `minispecsdom/` to `tool/internal/minispecsdom/`, rewrite the
16 import paths, and build. `tool/go.mod` keeps its `replace` for `github.com/zot/simple-dom`,
since `sdom` and `sdom/schema` are still used from 21 and 9 files respectively inside the
moved package alone. Nothing in the `Makefile` changes. The far side — deleting the package
from mini-spec-tool — is Item 3, not this one, so that the two trees agree for as long as
both build.

**Item 2** (the design). Nine specs, ten cards, nine sequences, nine test designs, and nine
`## Feature:` blocks of `requirements.md`. Every `Rn` in the slice is
mapped by a table: the 111 ids sorted, onto `R337`–`R447` densely, so this project's
sequence has no hole (the uniform offset first planned would have left 22). The table is
`rmap.tsv` in the moving session's scratch; the mapping ran single-pass over every file, so
no number was mapped twice. The four retired requirements keep their `~~Rn:~~ (Retired Tn —
see Rm)` shape with both numbers mapped and their `Tn` entries re-minted as `T4`–`T7`.

*The order is the hazard.* This project already owns `R210`–`R336`. A substitution run
over files that sit in `tool/design/` would rewrite this project's references, and the
result validates green. So: copy the slice to a scratch directory, renumber there with
word-bounded replacement over the table, grep the copy for any `R` in `210`–`355` that
survived, and only then move the files in. The Go sources' 132 refs are mapped in the same
pass, in the same scratch copy — Item 1's package copy is taken from that copy, not from
mini-spec-tool directly, or the two items are done in the other order and Item 1's code
lands citing numbers that mean something else here.

*Two names collide.* This project already has `crc-Carve.md`, `test-Carve.md`,
`crc-Pending.md` and `test-Pending.md`: the first pair is the file-system adapter over the
reader, the second is the orchestrator of the three verbs. The incoming pair are the
readers themselves. **DECIDED (Bill, 2026-09-14): the incoming pair are renamed
`crc-CarveSdom.md` / `test-CarveSdom.md` and `crc-PendingSdom.md` / `test-PendingSdom.md`**,
and the card titles follow (`# CarveSdom`, `# PendingSdom`). This project's adapters keep
their names. The rename touches only the moved files and the 28 + 40 `// CRC:` lines in the
moved Go that name the two cards; `seq-carve.md` and `seq-pending.md` do not collide and
keep theirs.

Afterwards: Artifacts lines added here (`O29`'s `unread.go` gets its mapping in the same
pass rather than being ported as a gap), `specs/index.md` reconciled and
`query unindexed-specs` empty, `file-formats.md` updated since the tool's own readers now
define the formats it lists, and the three specs here that say "read through
`github.com/zot/simple-dom`'s `minispecsdom`" (`backup.md`, `queries.md`, `queue-items.md`)
rewritten at the source.

**Item 3** (the far side). Here: `O28` and `O30` ported as gaps with their measurements;
every gap and spec here that says *raised with mini-spec-tool* (`O13`–`O17`, `O24`, `O21`)
rewritten, because the reader is now ours and the phrase would send a future agent to file
a request against a package that no longer holds it. The six unanswered `requests/` files
are read, and each is either already landed (the three `*-landed.md` and
`three-readers-ready.md` are notices, not defects) or becomes a gap here. In mini-spec-tool:
delete `minispecsdom/`, its Artifacts lines, cards, sequences, test designs, the ten specs
and the ten feature blocks — leaving `R210`–`R355` as a numbered hole that its `validate`
will report as `numbering gaps`, which is the honest record there and is closed by a note in
its `requirements.md` naming this carve. Its `requests/` directory retires. Then update the
2026-09-04 memory and the `alarm-puller` notes here, which still say the sibling holds the
readers.

**Item 4** (the alarms). All 56 `**Inject:**` lines name `minispecsdom/<file>:<symbol>`
and every one of those paths changes. The symbols do not, so a census after the landing
commit should read them as fresh; if it reads them `unchecked` or stale because the path
is new to git, the re-pull is fanned out per *Delegating the re-pull*, with the
`mini-spec-tool` sibling linked beside each worktree since `sdom` is still resolved through
the `replace`. Measured before deciding whether any pull is needed, not assumed.
