# Carve: reclaiming the tool over simple-dom

> **DRAFT (Daneel, 2026-09-04) — for Bill to edit before its first part is worked.**

On 2026-09-04 this repository restarted from `before-sdom-2` (`abd78fd`, the last tree with no
DOM of its own and the `.minispec/` behaviour already in it) as branch `new-sdom`; the August
DOM work is preserved as `old-sdom`. What `old-sdom` had and this tree does not is listed in
`.scratch/REVISE.md`, measured from the CLI summary spec at both refs. This carve is the public
home for getting it back — **over `github.com/zot/simple-dom`'s `sdom` and `minispecsdom` as
the only readers**, as a module dependency, with thin path-taking adapters in this tool's
`parser` — and it is the consumer half of mini-spec-tool's `carves/sdomification.md` Item 2.

## Status

- [x] ~~**Item 1 — the module dependency and the carve reader.**~~ **LANDED (`2a050a2`, 2026-09-04 — `#67`.)**
- [x] ~~**Item 2 — the backup slot.**~~ **LANDED (`7dd50a0`, 2026-09-04 — `#68`.)** Needs Item 1.
- [x] ~~**Item 3 — the `pending` verbs: `add-item`, `start`, `finish`, `revert`, `replay`.**~~ **LANDED (`eaf9604`, 2026-09-05 — `#69`.)** Needs Items 1 and 2.
- [x] ~~**Item 4 — `validate trajectory`.**~~ **LANDED (`7e6e293`, 2026-09-05 — `#70`.)** Needs Items 1 and 3.
- [ ] **Item 5 — the alarm-field verbs: `pulled`, `inject`, `number-alarms`.** **OPEN (#74.)** Needs Item 1.
- [x] ~~**Item 6 — the alarm site as a parsed extent, not a git pattern.**~~ **LANDED (`3931bcc`, 2026-09-06 — `#72`.)** Needs Item 1.
- [ ] **Item 7 — `update add-req` and the `query gaps` selectors.** **OPEN (not queued.)** Needs Item 1.
- [x] ~~**Item 8 — the skill re-derived as each verb returns.**~~ **LANDED (`0700965`, 2026-09-06 — `#73`.)**
- [x] ~~**Item 9 — the readers' backtick and key-fragment landings, absorbed.**~~ **LANDED (`9716bca`, 2026-09-06 — `#71`.)** Needs Items 3 and 4.

## Decisions

**DECIDED (Bill, 2026-09-04): consume simple-dom as a Go module dependency** — `require
github.com/zot/simple-dom` in `tool/go.mod` with a `replace` to the sibling checkout while both
move — and **never copy its packages in**: nothing owned the schema layer before, and a copy is
a second owner. Measured the same day in a scratch copy: `go 1.26`, `require … v0.0.0`, the
`replace`, one importing file; `go mod tidy` keeps the requirement, build and tests green.

**DECIDED (Bill, 2026-09-04): the path-taking calls live as thin adapters in this tool's
`parser`.** Their readers are source-string in, `Render()` out, and deliberately do three things
*not*: refuse a write over a line with deviations (they list them), write a file (they re-read
their own bytes), and print the per-file unread report. Every old caller — backup, the `pending`
verbs, `query carves`, `validate trajectory` — was written against path-taking functions, so the
adapters keep those signatures and are the one place the atomic temp-and-rename write and the
"listed by read paths, refused by write paths" rule live. *Update, same evening:* their readers
now refuse over deviations themselves (`DeviationError`), refuse `OPEN` over a checked part and
`Land` over a landed one — see the requests exchange below — so the adapters' refusal is the
file-level half only.

**DECIDED (Bill, 2026-09-04): reclaimed requirements take the next free number here**, with
the `old-sdom` number recorded in `.scratch/REVISE.md`'s map, because keeping the old numbers
would leave `validate` reporting a 164-wide numbering gap for good.

**DECIDED (Bill, 2026-09-04): one squashed commit per item, its steps visible in the body, and
no commit hash in a `Pulled` line** — `SKILL.md` carries the rule; gap `O12` banks the tool's half.

**DECIDED (Bill, 2026-09-05): the dependency's markdown table gets bracket groups for runs of
five backticks down to one**, longest first. CommonMark allows any run length and a native
run-of-the-same-character opener would express it in one rule, but five suffices: measured across
every markdown file ark indexes (3,171 files), the longest run is four, twenty-one times. Sent as
`requests/backtick-run-groups.md`, correcting `double-backtick-span-absorbs`: the mechanism is
parity inversion — a two-backtick opener read as an empty one-backtick span — not a span running
to end of file.
@ark-request-sent: requests/backtick-run-groups.md

**Inherited from mini-spec-tool (Bill, 2026-09-03): early comparisons, then wean, then one
cut.** Their readers' acceptance is their own tests over committed fixtures; this side's
acceptance is this tool's tests over its adapters. No test here shells out to their repository
and none of theirs reads this one.

## Item 1

The dependency cannot land alone — `go mod tidy` drops a `require` nothing imports — so it
lands with its first importer, and the first thing on the list that every later part needs
is the carve reader: `ParseCarve`, `PartIsLanded`, `SetMarker`, `SetPartLanded` as
path-taking adapters over `minispecsdom.Carve`, and `query carves` restored on top of them
(the `old-sdom` census, `R223`–`R234`, `R358`–`R362`, `R372`–`R374`, `R501`, re-derived).
The `stateless` column and the hyphenated-verb ruling are already theirs, so this part reclaims
them from spec text, not Go. Their readers gained `Line()` and `[]Unread{Line, Text}` today at
this tool's request; the adapters consume both.

**What the census prints is this tool's presentation** — their `sdomification.md` Item 1 is
the per-file "what could not be read" report, assessed discharged on the reader side; the
printing is here.

## Item 6

`old-sdom`'s `Extent` computed an alarm site's line range from its own parse and handed git
`-L <start>,<end>` (its `R404`–`R406`, `R465`). `fec7441`'s stopgap hands git a
receiver-aware pattern instead and banks two defects as gaps `O10` and `O11`: the first bounded
match may be a use or a doc comment, and git's range includes the trailing blank line so the
last declaration in a file reads stale on any append. Their `DeclarationName` carries the bare
method name and jumps the receiver, so the qualified `Type.Method` is computable from their DOM
but not exposed — noted to them in `requests/RESP-alarm-method-anchors.md`.

*Landed 2026-09-06 as `#72`.* `siteExtent` in `project/extent.go` over `sdom.LangGo` and
`schema.Go`: the declaring line through the line its groups close on, no comment, the receiver
read from the group between keyword and name (so their `DeclarationName` needed no change);
two declarations answering to one bare name is reported, not guessed. `LastChanged` reads
`git show HEAD:./file` once per file — `./` because a design root beneath the repository root
(`tool/`) otherwise reads every site as having no history, measured the same day — and hands
git `-L start,end`. R205 and R206 retired at the source (`T2`, `T3` → R304, R303); `O10`,
`O11` resolved; `O21` banks the markdown extent (old R402) until a non-Go site exists.
Measured on mini-spec-tool's tree: 123 alarms, 123 verified, 0 unresolvable.

## Item 9

Their 2026-09-05 landings — one pattern code group with `Unclosed` on the context and `Unread` on
every reader (`c42cd24`, `f3c947e`), and `PartLine.Key()` returning the fragment (`8819819`) —
reach this tool through the `replace`, not a version. Measured the same evening against their
head: build, tests and `validate` green, so the two requests asking for a bump are mostly
discharged already; Item 4 adopted the fragment key and `trajectory-format.md` states it. What
is left, one squashed commit:

- The `--from` shim in `pending.go` **stays**: it converts a typed `#Item 1` to the fragment by
  decision (R243, pattern 23 — flexible on input, rigid on output), and their `Key()` no longer
  returning that form makes it harmless rather than dead. The queue entry's `Next:` line said
  to refuse the form; it was wrong, and this is the record of why.
- `parser.Carve` and `QueueScan` carry the readers' `Unread()`; `query carves` prints an
  `unread` count per carve and in the census and lists the opener's line under `--open`
  (R301); `validate trajectory`'s coverage note counts per file across both queue files, the
  current file and every carve (R302). Both alarms pulled the same evening.
- Gap `O20` resolved: the done file reads whole (58 entries, reader disagreement gone).
  Both requests had already been answered `completed` the same afternoon
  (`requests/RESP-both-backtick-halves-landed.md`, `RESP-sdomification-key-fragment.md`);
  the evening's inbox sweep missed them, and the duplicates it prompted were removed.

## Item 5

*Landed 2026-09-07 as `#74`.* Blocked until their `TestDoc` reader landed overnight
(`478875e`, asked for in `requests/three-readers.md`); with it the port is thin. The parser's
line reader became an adapter over `minispecsdom.ParseTestDoc` (`Alarm` gained `ID`, `Code`,
`Line`, `Key()`), the census names alarms `<doc>#<n>` and lists an unnumbered one with its
repair, and three verbs sit on the reader's writes: `number-alarms`, `pulled --body-file`,
`inject`. What is decided here and not there: the name, the system-clock date, and `inject`'s
`void` — the old sites resolved in HEAD, the new on disk, which is the asymmetry old-sdom's
`O120` asked for. R310–R316. Dogfooded the same morning: 22 alarms numbered across eight test
designs, nothing else touched; the first run refused a doubled `**Code:**` I had written the
day before, which is the reader's deviation rule doing its job. Their reader cuts a test title
at a code span (`requests/testdoc-title-stops-at-code-span.md`); display only, the name is the
number now. `--alarm` selection and the missing-`Code:`-file report stay unreclaimed.

## Item 8

*Landed 2026-09-06 as `#73`.* Measured first: against `old-sdom` the three files were 503 lines
short, but the diff is not one-directional — this tree carries later text of its own (one commit
per item, `revert`/`replay`, the worktree anchor, no hash in a `Pulled:` line, the fragment key,
`REVERTED`), and much of what `old-sdom` had is process that depends on no verb. So the rule
applied was: **restore what depends on nothing or on a reader that is here; rewrite what
changed; leave out what waits.**

Restored as written: the no-task-tool relocation; injection design in Implementation and the
pull *after* Simplification (the 2026-08-21 measurement, 32 cycles for sixteen proofs); which
properties most need an alarm; a build-breaking injection is not a ring; standing context in
the current file and the `## Active` region (their `Current` reader has `Standing()`, R296
checks the heading); the sited-decisions rule and `@undecided`; the tool/agent contract table
(key row updated to the fragment); the part-line rules the reader enforces — em dash, the
`Item` word, checkbox interior, verbs in capitals (`part-line.md`, checked); read the whole
carve; the leading `Pulled:` date; the puller's step 0, build-failure and verbosity rules;
the delegation hazards — harness worktree at `origin/main`, `GOWORK`, the test cache, the
worktree sweep — plus one measured today, the relative `replace` needing a sibling.

Rewritten: the backtick rule now says where the tool reports it (`validate trajectory`, R302)
and where it does not (`design/`, `specs/` — gap `O22`); the prose-through-a-file rule names
the flags this tree has and banks the two verbs without them (`O23`); "ask the tool what
exists" no longer names `query sdom`.

Left out, each waiting on its verb: batched briefs and `--batch`, the brief's commit line
(R369), `query alarms --alarm`, `query sdom`, `update pulled` / `inject` / `number-alarms`
and `add-req` in the verb list. They return with `requests/three-readers.md` and the parts
behind it. The Cursor generator got its two task-tool edit pairs back with the paragraphs
they anchor on, which is what `make validate` checks.

## The requests exchange

`requests/trajectory-reader-requirements.md` (2026-09-04, completed): seven items their readers
owed, all landed the same day — write-path refusal (`DeviationError`, `ErrReopen`, `ErrLanded`),
`Line()` on every entry and part, lenient `OPEN` attribution reads, and a gap ID as a valid
`Source:` (their `Entry.Kind`, `SourceKey`, `ErrBadGapSource`). Item 7, several parts per entry,
is a door kept open. `requests/RESP-alarm-method-anchors.md` (in-progress) is this side's
answer to their one request; Item 6 above closes it.
`requests/pending-reader-defects.md` (2026-09-05, open) carries seven things Item 3's port found
the readers still owe — five defects, a sentinel request, and the parse-check at the end of `Mutate`.
@ark-request-sent: requests/pending-reader-defects.md
`requests/double-backtick-span-absorbs.md` (2026-09-05, open) is Item 4's finding: a two-backtick span
absorbs the rest of a document, silently — 41 of this repository's 58 done entries (`O20`).
@ark-request-sent: requests/double-backtick-span-absorbs.md
