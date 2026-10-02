# Carve: traceability comments over sdom — the last reader, and what it unlocks

> **DRAFT (Daneel, 2026-09-18) — for Bill to edit before its first part is worked.**

The code-file reader is the last file-kind not on sdom, and it is [sdom.md](sdom.md)'s
Item 5. `minispecsdom.TraceabilityComment` (`comment.go`) is written and tested but
**wired to nothing**: `validate.go` and `query.go` still harvest inline `Rn` refs
through the regex `parser.ParseTraceability`. This carve wires the sdom reader in,
retires the extractor, and builds what a position-preserving, round-tripping code
reader unlocks — a positioned reverse lookup, and eventually ref rewriting.

**Provenance.** Split from [sdom.md](sdom.md) Item 5. The reverse-lookup request came
from ark (`minispec-query-implementation`, 2026-09-18): "where is `Rn` implemented?" —
the positive of the `missing impl coverage` check, honoring the harvest's shape rules
rather than a raw `grep Rn`, with a pattern form for the "know the concern, not the
number" case. Developed from planning notes worked up with Bill on 2026-09-18.

**What is already true (verified 2026-09-18), so we do not rebuild it:**

- `minispecsdom.TraceabilityComment` implements the spec (`tool/specs/traceability-comment.md`)
  faithfully: `Test:` field, reqlist as a field in any order, the bare `// R563-R570: …`
  form, and `--`/`—`/`:` descriptions — with typed views `CRC/Seq/Test/Refs/Description`
  and a writable `Description`.
- **Ranges are already handled.** `sdom.RequirementList` keeps a range as one faithful
  item (`R563-R570` stays literal) and `Items()` expands it to the full member list. So
  `565 ∈ Refs().Items()` already answers "is R565 implemented?" correctly.
- The sdom nodes carry positions, so `file:line` is free — no line-capture work is owed
  (that was only true of the old positionless extractor).

## Status

- [x] ~~**Item 1 — wire `minispecsdom.Comments` into the harvest; retire `parser.ParseTraceability`.**~~ **LANDED (2026-09-29 — `#93`.)** Re-point `validate.go`'s impl-coverage harvest and `query.go`'s code-ref reader onto the sdom reader. Build one positioned harvest both consume.
- [x] ~~**Item 2 — `minispec query implementation` (ark's request).**~~ **LANDED (2026-09-29 — `#92`.)** Reverse lookup `Rn` (or a pattern) → the code that implements it, over the sdom harvest. Rides on Item 1's positioned harvest.
- **Item 4 — C++ raw strings.** **MOVED (Bill, 2026-09-25 — [sdom.md](sdom.md) Item 5.5.)** Language support is tracked in the sdom carve.
- [x] ~~**Item 3 — a canonical writer on `RequirementList`.**~~ **DISCHARGED (mini-spec-tool `355c36d`, 2026-09-30.)** The writer already existed — `SetItems([]int)` since their `#15`, with `RequirementText([]int)` beside it; `Ranges()` and `Contains(n)` were added on our request (their R364, R365). Unreleased; reached through the workspace. Item 5.2's precondition is met — see Decisions.
- **Item 5 — requirement lists in minimal range form.** No checkbox: the subparts carry the state.
  - [ ] **5.1 — a CRC card's `**Requirements:**` line is read through `sdom.RequirementList`.** **OPEN (not queued.)**
  - [ ] **5.2 — `add-ref` and `remove-ref` rewrite the line in sorted minimal form.** **OPEN (not queued.)**
  - [ ] **5.3 — the grep lookup is superseded by the range-aware queries.** **OPEN (not queued.)**

## Decisions

**DECIDED (Bill, 2026-09-18): `query implementation` arg and output shape.**

- **Arg form.** Multiple `R#`s allowed, with `Rn-Rm` ranges and optional commas. The
  CLI first classifies the args: if they parse cleanly as a list of R-refs it is
  **number mode**; otherwise **text mode** (a pattern matched against requirement text).
  The anchored ID/range grammar in `query.ExpandGapRefs` is the working model, one
  namespace over.
- **Retired requirements.** By **number**, included implicitly (the cleanup query: "what
  code still points at a retired `Rn`?"). By **text**, excluded unless `--retired`.
- **Output.** Number mode prints just the code locations (the caller knows the `Rn`).
  Text mode prints the matched requirement line(s), then their locations, plus an
  explicit "no impl refs" — the positive twin of `missing impl coverage`.
- **Flag.** `--retired`, so "no retired" is the default and it only affects text mode.

**DECIDED (Bill, 2026-09-18): build `query implementation` on `minispecsdom.Comments`,
not the old extractor.** Positions and the full grammar are already right there;
building on `ParseTraceability` would invest in the reader Item 1 retires. This is why
Item 2 rides Item 1.

**Ranges vs. `RangeSet` — the record.** `RequirementList` expands to a flat `[]int` via
`Items()`, which is correct for the query (`Contains` is `slices.Contains(Refs().Items(),
n)`, and requirement spans are tiny). A `RangeSet` of `(lo,hi)` intervals is worth doing
for the **write** side — editing a ref while keeping the range literal, cheap inclusion
without materializing — but it does not block the read query, so Item 3 is independent
of Item 2. *Its write-side consumer arrived 2026-09-29 as Item 5.2, and the whole-line decision below replaced the interval view with a canonical writer — see the next decision.*

**DECIDED (Bill, 2026-09-25): Item 1 lands before Item 2, in the same batch.** Measured the
same day over this repository's 92 Artifacts code files: of about 890 distinct (file, Rn)
pairs, the regex harvest and the sdom reader disagree on 36. `query implementation` promises a
ref counts exactly where validate counts it, which cannot hold while the two read differently.
The 36 fall into three classes, and each is decided:

- **A ref followed by a full stop and prose counts** (`// R271. A gap is a source…`, 20
  pairs). The regex counted it and the sdom grammar rejected the whole comment, so the grammar
  widens: a `.` directly after the refs segment starts the description, as `:` does.
- **A Seq-only comment counts, refs included** (`// Seq: seq-backup.md#2.2 | R232`, 15 pairs).
  The sdom reader already counts it; the regex required `CRC:`. This supersedes the skill's
  "a `Seq:`-only line does not trigger".
- **A `//` quoted inside a comment's prose does not count** (1 pair). The regex matched a
  comment leader inside backticks; the sdom reader sees one comment and is right.

**DECIDED (Bill, 2026-09-25): minimal language tables for what sdom does not ship.** The
harvest picks a reader by file extension. Go, JavaScript/TypeScript, Lua, Shell and Python come
from sdom; HTML, Markdown, CSS and C/C++ are constructed in `minispecsdom` (*superseded the same
day for C/C++: C, C++ and Java are three tables — see below*). HTML's embedded
JavaScript and CSS are read through bracket groups, the way a template literal restricts to
`${`: `<!--` is raw and live everywhere, `<script` is code mode, `<style` is restricted to
comments and strings, and every JavaScript group lists `<script` and the JavaScript brackets
as its allowed parents, since `AllowedParent` checks the immediate parent only. A file whose
extension has no table is reported as not read, never skipped. Where a language has several
comment forms, the one to write is its table's `Comment` style, which every table sets: bracket
order belongs to matching (Lua's `--[[` must precede `--`), so it cannot also mean "preferred".

**DECIDED (Bill, 2026-09-25): the configurable comment patterns retire.** Once the harvest
reads through the language tables, nothing reads `comment_patterns` or `comment_closers`, so
the keys, their defaults and `parser.ParseTraceability` retire, superseded at their source. A
config that still sets them is told the keys are retired, not silently obeyed or silently
ignored. **`query comment-patterns` stays in role**: it teaches an agent how to write a
comment in each extension, closers included, so it reports each table's comment style instead
of a regex. ~~No config key maps new extensions to a table yet; an unmapped file is reported until the tool learns its language.~~ *Superseded the same day: projects define languages in their configuration — see below.*

**DECIDED (Bill, 2026-09-25): C, C++ and Java are three built-in tables.** Their comment
shapes agree from C99 on; their strings do not. C++'s table carries raw strings,
`R"delim( … )delim"` ([sdom.md](sdom.md) Item 5.5), and Java's carries text blocks, `"""…"""`, as a group ahead
of `"` so a text block is not read as an empty string followed by another. Extensions: `.c`,
`.h` for C; `.cpp`, `.hpp`, `.cc` for C++; `.java` for Java.

**DECIDED (Bill, 2026-09-29): the grammar stays strict about the separator; ark fixes its
comments.** The regex reader's bare-annotation rule counted a ref followed straight by prose,
`// R5 handles the retry`; the grammar reads that as prose, since its interior is not wholly
fields. This repository has none. Ark had 29 (measured by grep over its 258 tracked code
files, an upper bound). Widening the grammar once more would have kept them, at the cost of
another exception to recognition-is-consumption and of making `// R5 is wrong here` a claim
that R5 is implemented. Bill chose the separator instead: it is what makes the intent
unambiguous. Ark is asked to add a colon after the refs before the new binary replaces its
minispec (`requests/ark-ref-separators.md`), and SKILL.md teaches the separator form.

**DECIDED (Bill, 2026-09-27): Emacs Lisp is built in, and Pascal is mapped.** A census of
YAML configurations under `~/work` found two projects on languages no table covered: NitroPascal
(`.pas`, `.dpr`, `{ … }` comments) and nuterm (`.el`, `;` comments). Pascal is sdom's own
`LangPascal`, so it joins the extension map. Emacs Lisp gets a table here: `;` line comments,
`"` strings with `\` escapes, `( )` and `[ ]` brackets. Its character literals (`?(`, `?\)`)
need a group that opens only at a token start, because `?` also ends predicate names
(`f-exists?`): measured over 10,730 installed `.el` files, 17% hold a bracket char literal, and
16,571 `?` follow a symbol character. sdom cannot yet test what precedes an opener, so
`BeforeOpen` — the dual of `BeforeClose`, mechanism rather than language — is requested from
simple-dom ([sdom.md](sdom.md) Item 5.7); the table itself stays here.

**DECIDED (Bill, 2026-09-25): projects define languages in their configuration, mirroring
sdom's own structs.** A `languages` entry in `.minispec/config.toml` is a `BracketLang` (or
`IndentLang`) written field for field in snake case — `brackets` with `open`, `close`,
`escape`, `separators`, `allowed_inner`, `allowed_parent`, `kind` and the rest, plus the
`comment` style — rather than ark's chunker categories, which special-case what the structs
already say. Bracket order is matching order, exactly as in Go, and `comment` is required,
since it is the form written. A definition names its extensions and overrides the built-in
table for them; a same-named definition in a design root replaces the repository's whole.
The code/raw distinction survives the file: TOML decodes an absent `allowed_inner` as nil
and `allowed_inner = []` as empty (measured, BurntSushi/toml v1.5.0), and has no null to
blur them. A definition sdom rejects is an error naming the file and the language.
**An example configuration ships in the skill directory** defining C, C++, Java, Go and
Python, as a template to copy from rather than a second source of the built-ins.

**DECIDED (Bill, 2026-09-25): configuration moves to TOML first.** The language definitions
are written once, in TOML, so the format change goes ahead of Item 1 as its own item (`#94`,
from gap `O30`); Item 1 (`#93`) is paused behind it.

**SENT (Daneel, 2026-09-25): two requests to simple-dom** (`mini-spec-tool`,
`requests/sdom-check-and-matching-close-groups.md`). An exported `Check() error`, because
`NewBracketParser` panics on a malformed table as a library invariant, and a table read from
a user's configuration is caller input. And a pattern closer whose named groups must equal
the opener's, generalizing `CloseIsOpen` — Bill's design — which C++ raw strings need and
which also fixes Lua's long brackets (`[==[ … ]==]`).

*Answered and landed the same day* (mini-spec-tool `#40`–`#44`, response
`RESP-sdom-check-and-matching-close-groups.md`): `BracketLang.Check() error`, and
`BracketGroup.CloseRegex` — with **no** `CloseGroupsMatchOpen` flag, since naming the same
capture groups in `OpenRegex` and `CloseRegex` is the declaration, and `Check()` refuses
patterns whose group sets differ. A configured language is therefore checked with `Check()`,
never by recovering a panic, and the configuration mirrors `close_regex` with no flag. A
behaviour change arrives with it: in code mode a closer now closes an enclosing group from
inside a child, ending the groups between (`( { )` pairs `(` and lists `{` unclosed), so
unclosed and stray counts on malformed code shift, and the harvest's unread report reads
those. **It reaches this tool only through a simple-dom release:** `go.mod` requires the
published v1.0.0, and `~/work/go.work`, which uses the local checkout, is off in release
builds and in worktree alarm pulls.

**DECIDED (Bill, 2026-09-29): `add-ref` writes requirement lists in minimal range form —
pairs as lists, and the whole line rewritten.** Bill's proposal, for the write side the
`RangeSet` record above was waiting on. A run of three or more consecutive numbers is written
as a range (`R502-R505`); two neighbours stay a list (`R5, R6`), since a range saves nothing
on a pair and reads worse. Every `add-ref` and `remove-ref` rewrites the whole line sorted
and minimal, rather than merging into an adjacent run — rigid on output — at the cost of one
reordering diff per card, once. Measured the same day over the 51 `**Requirements:**` lines
in the CRC, sequence and test designs: all plain `Rn` lists, 12 out of numeric order, 46 that
would shrink, 4611 characters to 2531. This spans Items 3 and 5, so it sits here.

**Item 5** carries it, in the order that keeps every step safe. **5.1** first: today
`ParseCRCCard` splits the line on commas and keeps each token as written, so `validate` and
`Coverage()` — both reached through `GlobCRCCards`, CRC cards only — would read `R502-R505`
as one unknown ref and R503, R504 as uncovered. Loud rather than silent, but a range must not
be written until the reader expands it, through the grammar code comments already use.
**5.2** then: `add-ref` sorts, dedupes and compacts; `remove-ref` is where a range splits
(`R502-R505` less R504 is `R502-R503, R505`), which is what Item 3's `RangeSet` is for.
*Spelling and the empty-line rule superseded 2026-09-30 — see the answer below.* The
`**Requirements:**` lines of sequence and test designs are read by nothing and written by no
verb, so they are out of scope. **5.3** last: `/minimap` teaches *Requirement → everywhere it
lands: grep -rn "R5" design/ src/*, and a grep for R503 misses a card that writes
`R502-R505`. Code comments already allow ranges, which is why `query implementation` exists;
`query coverage` maps each Rn to its design files and becomes the lookup to teach. The grep
line in minimap and any echo of it in SKILL.md is rewritten at the source, or it is a trap.

**DECIDED (Bill, 2026-09-29): Item 3 asks simple-dom for a canonical writer, not an interval
view.** The interval view was for byte-preserving edits inside a range; with the whole line
rewritten nothing edits inside one — the verb reads the members through `Items()`, changes the
set, and writes it back. So the request is a `SetItems([]int)` on `RequirementList` that renders
the minimal form (runs of three or more as `Rn-Rm`, pairs and singles as a list, `, ` between),
so the reader and writer share one grammar by construction. `Ranges()` and `Contains(n)` are
named in the request as optional. *Supersedes the `Ranges() [][2]int` / `Contains(n)` shape
Item 3 was filed with on 2026-09-18.*

**ANSWERED (mini-spec-tool, 2026-09-30 — `RESP-requirement-list-writer.md`): the writer was
already there.** `SetItems` predates the request, which was filed without reading sdom's
`list.go` — the *ask the tool what exists* lesson, paid for with one round trip. Three things
came back with it:

- **DECIDED (Bill, 2026-09-30, with mini-spec-tool): a range is written `R5-8`, not `R5-R8`.**
  A grep finds only a range's two ends in either spelling, so the second `R` buys little and
  the short form is the smaller line. The reader accepts both, so existing lines still read.
  *This supersedes "`Rn-Rm`" and "`R502-R505`" in the 2026-09-29 decisions above*, which are
  left as written because they were the decision of that day. Re-measured with the short
  form: 51 lines, 46 shrink, 4611 characters to 2420.
- **An empty set does not read back** (their R366): `SetItems` of nothing writes an empty
  literal, and a list has at least one item. **Bill's rule: removing the last requirement
  removes the `**Requirements:**` line.** `validate` already reports a card with no
  requirements as an orphan whether the line is missing or empty, so the rule opens no
  second corner there.
- **But it exposes one in `AddRef`, live today.** It looks for the `**Requirements:**` line
  and, finding none, writes the file back unchanged and returns success — so `add-ref` on a
  card whose line was removed does nothing and says nothing. Item 5.2 inserts the line when
  it is absent (beneath the card's heading), and that path gets a test and an alarm.

## The CLAUDE.md obligation

Adding a `query` subcommand trips the summary-spec rule: `tool/specs/cli-commands.md`
gains the `implementation` entry in the same pass. And the feature runs the full
mini-spec phases, including a `test-*.md` + fire alarm for the classifier and the
positioned harvest (cheap, deterministic — no infra).
