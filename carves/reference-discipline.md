# Carve: reference discipline — links that resolve for a cloner

Publishing a document publishes its pointers. A public document citing a private
working note hands a cloner a dangling link: it resolves to nothing, and nothing
warned that it was ever going to. This carve is the rule for which documents may
cite which, and the check that enforces it.

**Provenance.** Split out of [trajectory-tool.md](done/trajectory-tool.md) on
2026-08-04 (Bill's call). It arrived there because trajectory documents are where
the problem was noticed, but the rule is a validator over markdown links in any
project document and is useful on `design/` and `specs/` today. Nothing about it
is trajectory-specific.

## Status

- **Item 1 — the checker.** **SPLIT (Bill, 2026-08-16.)** No checkbox: the sub-items carry
  the state. It blocked [trajectory-tool.md](done/trajectory-tool.md) Item 3 and 8.2, whose
  markdown reading is shared rather than reimplemented — a fenced example is not data,
  whether it holds a link or a status entry; both re-landed over the shared reading on
  2026-09-04 and 2026-09-05, so nothing waits on this carve any more.
  - [x] ~~**1.1 — the simple DOM: a position-preserving markdown parse.**~~ **LANDED (`2a050a2`, 2026-09-04 — `#67`.)**
    Landed twice: first as `#11` (`70abbfe`, 2026-08-16, `internal/mdom`), abandoned with
    `old-sdom` at the 2026-09-04 restart; then as `github.com/zot/simple-dom`'s `sdom` with
    `schema.LangMarkdown`, a module dependency, and the `minispecsdom` readers over it, moved
    into `tool/internal/minispecsdom` on 2026-09-14 (`#80`). See *What landed, and what did
    not* below — two things this part promised are still owed to 1.2.
  - [x] ~~**1.2 — extraction, resolution, git status, on top of it.**~~ **LANDED (`6d0dd9c`, 2026-09-15 — `#83`.)**
    Needs a link reader first: the DOM does not model links.
- [ ] **Item 2 — the document-class model.** **OPEN (#90.)**
- [x] ~~**Item 3 — wire into `validate` and report.**~~ **LANDED (2026-09-15 — `#89`.)**
- [x] ~~**Item 4 — repair links broken by a carve's move, both directions.**~~ **LANDED (`22974e6`, 2026-09-15 — `#84`.)**
  Added 2026-09-15 after `query links`' first run found 32 of them (gap `O27`).
- [x] ~~**Item 5 — a move verb that rewrites links as it moves.**~~ **LANDED (2026-09-15 — `#85`.)** Prevention
  for the class Item 4 repairs; needs Item 4's rewrite.

## Decisions

**DECIDED (Bill, 2026-08-04): a reference is a real markdown link**, `[text](path)`
— not a prose mention. That is what makes it checkable at all, and it is the
precondition for everything else here.

**What each document class may point at:**

| document                | may reference                                           |
|-------------------------|---------------------------------------------------------|
| private carve directory | anything, including planning scratch                    |
| public `carves/`        | **only VCS-managed files**                              |
| `specs/migrations/`     | **only VCS-managed files**                              |
| the pending file        | in between — it points at both scratch notes and carves |

**DECIDED (Bill, 2026-08-04): git only, through the `git` command line.** No fossil
support and no linked-in git library. Supporting a second VCS means the tool has to
know how to *operate* it, a large surface for a check this small; shelling out gets
tracked/ignored status for free and stays correct as git changes.

*A real narrowing, stated rather than hidden:* a project whose public documents are
managed by something else gets no reference checking. In a git+fossil project the
rule still works wherever the public documents are the git-managed ones — the
arrangement that motivated this — but a fossil-only document's references go
unchecked, and the tool should say so plainly rather than pass silently.

**DECIDED (Bill, 2026-08-04): error on ignored, warn on untracked-but-not-ignored.**
The first is certainly wrong — a link into an ignored path can never resolve for a
cloner. The second is usually just early: writing a carve and its referenced doc in
one session and filing the item before committing is the ordinary workflow, not a
mistake.

## The split

**Item 1** — the checker. Extract links, resolve them relative to the containing
file, classify each as tracked / untracked-but-not-ignored / ignored / missing.

*Cheap, but not trivial — it must parse markdown, not grep it.* A throwaway version
run over the trajectory carve on 2026-08-04 flagged a missing file that was not a
link at all: the literal `[text](path)` inside the code span in the decision above.
Links in code spans and fenced blocks are examples, not references, and a regex
cannot tell the difference. That failure is the safe direction — a false alarm
rather than a silent pass — but it is the same lesson as the instrument table in
the trajectory carve, found on the document that argues it.

**DECIDED (Bill, 2026-08-16): the markdown reading is a *simple DOM*, and it is split out
as 1.1 because three parts now stand on it.** Parse into only the nodes we operate on —
headings, list items and their checkboxes, fenced blocks, code spans, links — and keep
**every other byte exactly where it was**. Not an AST: the shape, with everything else
carried as opaque spans.

*Two halves, and neither works alone.* Selective structure without total preservation is
the lossy round-trip — a document read into structs and written back, with the comments
and the unmodelled bits gone. Total preservation without selective structure is a full
markdown parser we have no use for. Together they round-trip, which is the property the
tests below can actually check.

*Fences and code spans are the mechanism, not a special case.* Borrowed from microfts2's
bracket chunker, where a group can be **scan-restricted** — inside it, only its close and
escape are recognized and everything else is literal text. A fenced block and a backticked
span are exactly that, so fence-awareness falls out of the lexicon rather than being
bolted on. Two known bugs go with it: the `[text](path)` above, and a `## Status` example
inside a fence counted as real open work — measured 2026-08-16, a grep widened past
`carves/` reports 32 open items where 12 exist.

*This is also how a tool may write into a document a human owns.* Mini-spec's files are an
heirloom: markdown in the places people already read, with the tool a replaceable consumer.
An engine that rewrites what it does not fully model will eventually delete something it
never saw. With a position-preserving DOM an edit is a byte-range splice, so untouched text
is untouched byte-for-byte and that class of bug cannot be written.

**DECIDED (Bill, 2026-08-16): 1.1 is markdown and only markdown**, kept separate from the
YAML reading the tool already does.

*The justification is the sharing, not a rule about mode flags.* An earlier draft here
said a parser that grows a flag for a second lexicon has stopped being simple. That is a
bright line standing where judgment belongs, and it would forbid something obviously
right: mail and HTTP are a control line, then headers of the same shape, then a blank
line, then a MIME body, and one parser with a pluggable piece captures all of that.
Splitting is **earned specialization** — you divide when carrying both has become
burdensome, not when a difference first appears. Markdown against YAML earns it easily,
sharing essentially no lexicon; that is why these are two and not one with a switch.

*Within* markdown the split is not earned **here**, and the reason is worth stating rather
than inheriting as a law. All three consumers model the **same region** of markdown:
headings, list items and their checkboxes, fenced blocks, code spans, links. 1.2 wants
links, Item 3 wants status entries, 8.2 wants checkbox lines across carves — three
*schemas* over one set of nodes. So one DOM, with schema in readers on top: the DOM knows
headings, list items, fences and code spans; it does not know what a carve is.

*The general form of that has no rule in it, which is why the reason is spelled out.* One
lexicon does not imply one DOM — two uses of markdown modelling disjoint regions (headings
and checkboxes versus paragraphs and emphasis spans) are legitimately two, sharing a
tokenizer and nothing else worth sharing. The unit is **region modelled**, not format, and
the answer here is "one" because the regions coincide, not because they share a file
extension. See the [earned-specialization] pattern, where three attempts to reduce this to
a test each died to a counterexample.

*The rule is not hypothetical — this project already has two, and never noticed.* The
comment-eating `--repair` bug was cited here as motivation for **this** part, which is the
right lesson and the wrong parser: `--repair` reads **YAML**. Its fix, landed in `5ea35f1`,
is `setTrack` in [tool/internal/project/init.go](../tool/internal/project/init.go) parsing
into a `yaml.Node` and rewriting **only the value node** — a YAML simple DOM, borrowed
rather than written, whose own comment gives the pattern's argument exactly: unmarshalling
into `Config` "discards three things at once: the file's comments, the order of its keys,
and any setting written by a newer tool version." So markdown gets 1.1 and YAML already has
`yaml.Node`, two lexicons, two parsers, arrived at independently and correctly never fused.

*One caveat on the borrowed one, which is the honest reason to prefer writing your own.*
`setTrack` states its limit in the source: **"byte-fidelity is not claimed"** — a blank line
between a comment and what it annotates is lost, though the comment and its attachment
survive. So the YAML DOM would **fail** the verbatim-reproduction test below. That is
tolerable for a config file the tool owns the schema of, and it would not be tolerable for
a carve, which is prose a human writes. Worth knowing before anyone reaches for an
off-the-shelf markdown library on the strength of this decision: 1.1 needs the byte
fidelity that `yaml.Node` explicitly does not offer.

**DECIDED (Bill, 2026-08-16): the test discipline is part of 1.1, not a follow-up.** The
pattern makes a claim a test can check exactly, and reaching for it without the tests buys
nothing.

1. **Verbatim reproduction.** `emit(parse(x))` equals `x` byte for byte — not
   semantically, not modulo whitespace. **Run it over the real corpus, not fixtures:**
   every document in `carves/`, `tool/design/`, `tool/specs/`, and ark's tree. That is the
   direct lesson of the `--repair` bug — a fixture contains only what its author thought
   to include, and what a lossy parse eats is exactly what nobody thought of. The corpus is
   a test suite nobody has to write, and it grows on its own as documents are added.
2. **Edit equivalence, as a commuting diagram.** For each edit operation: parse, change the
   DOM, emit — and separately edit the text directly. Require the two results identical.
   Round-trip identity says the document can be put back; this says the change made
   *through* the structure is the change meant for the file.

**The reference edit must be independently written, and that is the trap.** If the textual
path runs through the DOM, or shares its span arithmetic, or calls the same helper, the
test proves a function equals itself. Write it naive and obviously correct — replace this
line, splice at this offset — accept that it is slow and handles only simple cases, and
keep it in the test file where nobody is tempted to reuse it.

**Order matters:** verbatim reproduction is the precondition. If `emit` is unfaithful,
edit equivalence can pass while both paths are equally wrong, which is a green test over a
corrupted file.

*The failure mode is unusually good, which is worth knowing before writing the alarm.*
Drop one span kind in the parser and the round-trip goes red on the first real document
containing one, naming the file — where a coverage test stays green while the same span
silently disappears. Pull it deliberately anyway; it is just louder by construction than
most guards.

### What landed, and what did not (2026-09-15)

The DOM in this tree is simple-dom's `sdom` under `schema.LangMarkdown`, and it holds the
08-16 decisions point by point: headings, list items, checkboxes, fences and code spans are
the modelled nodes and everything else is opaque `Text`; `Doc.Render` rebuilds from the nodes
rather than handing back its source, which was the R216 lesson of the first landing; and the
fence and the code span are one scan-restricted group — a run of backticks closing only on a
run of the same length, the parity fix of 2026-09-05 — so fence-awareness is the lexicon, as
this carve asked. `TestByteRoundTripPerLanguageOverTheCorpus` runs every shipped language over
mini-spec-tool's own sources, specs, design and carves, byte for byte. Edit equivalence lives
in the readers' write tests rather than as the commuting diagram prescribed above.

**Two things 1.1 promised are not there, and they are 1.2's first steps.**

1. **Links are not a node.** `schema/markdown.go` says so by design: `[` opens no group,
   because a line-head marker may share no first byte with an opener (R231). So 1.2 cannot
   read links off the DOM. The shape that fits what exists is a link reader in
   `minispecsdom`, like the others — scan the `Text` runs outside code groups and bind their
   locations — which is cheap exactly because the code-group test already exists, and that
   was the whole reason for wanting a DOM under a link checker.
2. **The round-trip corpus is theirs, not ours.** The test's globs name mini-spec-tool's
   tree. Nothing round-trips this repository's `carves/`, `tool/specs/`, `tool/design/` or
   ark's tree, which is the population the test discipline above names. One test on this
   side, before 1.2 writes anything on top.

**1.2** is what the original Item 1 described — link extraction, resolution relative to the
containing file, and tracked / untracked-but-not-ignored / ignored / missing classification
— now written against the DOM instead of against lines, beginning with the two residues above.

**Item 2** — the document-class model: which classes exist in this project, which
are public, and what each may cite. It was coupled to Item 1 of
[trajectory-tool.md](done/trajectory-tool.md), which decided how a project declares its
siting; that carve is done, so the class of a document is now a fact about where it lives
under a layout that is settled, and open question 1 below is this carve's alone to answer.

**DECIDED (Bill, 2026-09-15): the class is git's to say.** A document is public exactly when
git tracks it, and a public document may cite only what git tracks — which is what the
classifier already checks. Nothing is declared: no configured directory list, no filename
convention, no marker. The population of the check is every tracked markdown file under the
repository root; the private files — the trajectory ledgers, ignored scratch — are outside
it by the same rule. **A staged file counts as tracked**, new and uncommitted included: the
question is asked of the index, not of history, so a carve written and staged this session
cites cleanly before its first commit. Open question 1 is answered by this; Item 2 becomes
the population change in `validate trajectory` and in `finished-carve`'s incoming rewrite,
closing gap `O28`.

**Ark made the case that this cannot be inferred.** Its queue files are fossil-only
and untracked in git — private by a filename-case convention (top-level uppercase =
private) that git cannot see — while its `carves/` are public and pushed. Untracked
and not ignored is exactly the state of a file written five minutes ago, so the one
mechanism that could have answered returns the ambiguous answer for the real case.

**DECIDED (Bill, 2026-08-04): trajectory files are gitignored in every project**, so
the queue's class stops being inferred and starts being declared where the checker
already looks. That settles one class universally and makes "a public carve may not
link the queue" a constant rather than a per-project answer — but it settles only
that class. Everything else here still needs the model.

**Item 3** — wire into `validate` and report. Errors and warnings distinguished per
the decision above; the report names the citing file, the link, and why it failed.
`validate trajectory` already parses every carve and both queue files through the readers,
so this is one more pass over documents the tool has in hand, not a new reading.

**Item 4** — repair links broken by a carve's move. When `trajectory-tool.md` moved to
`carves/done/` on 2026-09-14, every relative link in it kept pointing where it used to live,
and the first run of `query links` over the done carves read 32 `missing` in that one file.
The skill says to rewrite links as part of a move; the move is a hand `git mv`, so the rule
has no forcing function, and this part is the repair for what has already broken.

**DECIDED (Bill, 2026-09-15): a repair, and it lives with the tool's writes, not on
`validate`.** `validate` is read-only by its spec and is run constantly; the one `--repair`
in the tool is on `init`, which is itself a write over a file the tool owns the schema of.
The home is `update repair-links [file...]` (Bill, same day, choosing it over a
`query links --repair` flag), built on the classifier `query links` already runs.

**The predicate is what makes writing into a human's document safe.** A link qualifies when
it is `missing` now and resolves once re-based at the carve's former directory — for
`carves/done/x.md`, at `carves/` — which is the classifier run twice. The rewrite is then
mechanical: `../tool/…` becomes `../../tool/…`, `done/y.md` becomes `y.md`, spliced by byte
range through the DOM so nothing else in the file moves. A link that resolves both ways, or
neither, is reported and left alone; it is not the tool's to guess.

**DECIDED (Bill, 2026-09-15): both directions.** When a carve moves, every document that
linked it breaks too — `reference-discipline.md`'s own links to `trajectory-tool.md` were
fixed by hand at the time — so the predicate is *missing, but resolves under a sibling
relocation*, not *the moved file's own links*. Outgoing: re-base the citing file at the
carve's former directory. Incoming: re-base the target under `done/`, or out of it. The
population is every document `query links` reads, not the moved file alone.

**Item 5** — a move verb that rewrites links as it moves. **DECIDED (Bill, 2026-09-15): the
prevention is built too**, not left as a rule the skill states in prose and a hand `git mv`
ignores. The verb moves a carve to `carves/done/` and rewrites both directions in the same act, so
a moved carve never enters the state Item 4 repairs.

**DECIDED (Bill, 2026-09-15, at the start of `#85`): the verb is `update finished-carve
<carve>`, one direction, and it is a plain rename — nothing staged.** *Finished* names the
event, so the verb also refuses a carve whose status block still has an open part, or has
none; the reverse move is rare enough to stay a hand move followed by `repair-links`. The
tool never stages, and git finds the rename at commit time.

*Superseded the same day, at the decision above it:* the rewrite is not Item 4's predicate
called after the move — that would let a link resolve to a different file that happens to
sit at the relocated path. The verb knows the destination, so every rewrite is computed from
where each link resolves **now**: outgoing links reach the same target from `done/`, incoming
links from every document the tool reads reach the carve at its new path. A link that does
not resolve today is reported and left, not the move's to fix. What is refused before any
byte moves: a target already present, an open part, and a rewrite that fails to read back.

## The hole the tool cannot close

The rule checks markdown links, so a *prose* mention of a private file is invisible
to it. Ark's carve README already governs that case: a working note named in prose
without a link must have its contribution *stated in the sentence that names it*, so
nothing is lost by not having the file. The two conventions compose, but only the
linked half is machine-checkable.

**The prose half stays a human obligation**, and it is load-bearing rather than
decorative: ark's own convention exists because its carves must reference private
queue items, which can never be links. So the unlinkable reference is not an
edge case to be stamped out — it is the normal way a public document points at a
private one, and the check must not push people toward deleting the mention
instead of writing the sentence.

## Validated against a real project, 2026-08-04

The rule was prototyped by hand over ark's six public carves before being written
down. Every markdown link resolved to a tracked file; nothing was flagged. Ark's
planning scratch is gitignored, so a link into it from a public carve *would* error
— the case the rule exists for is reachable, and currently unviolated.

The check also found a live violation of the prose half, written that same
afternoon: a carve line reading "verified live by <two private rig files>", which
tells a reader with only the repository nothing at all. Rewritten to state what the
rigs showed.

## Open questions

1. **How does a document declare its class?** By location (a configured list of
   public directories), by filename convention (ark's uppercase rule), or by a
   marker in the file? Location is the most mechanical; ark shows convention is
   real in the wild.
2. **What about links out of the repository** — a URL, or a path above the project
   root? Ignore them, or check only that they are well-formed?
3. **Does this run on every `validate`, or on demand?** It shells out to `git` once
   per referenced path unless batched, and `validate` is run constantly.
