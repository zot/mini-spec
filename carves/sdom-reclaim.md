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

- [ ] **Item 1 — the module dependency and the carve reader.** **OPEN (#67.)**
- [ ] **Item 2 — the backup slot.** **OPEN (not queued.)** Needs Item 1.
- [ ] **Item 3 — the `pending` verbs: `add-item`, `start`, `finish`, `revert`, `replay`.** **OPEN (not queued.)** Needs Items 1 and 2.
- [ ] **Item 4 — `validate trajectory`.** **OPEN (not queued.)** Needs Items 1 and 3.
- [ ] **Item 5 — the alarm-field verbs: `pulled`, `inject`, `number-alarms`.** **OPEN (not queued.)** Needs Item 1.
- [ ] **Item 6 — the alarm site as a parsed extent, not a git pattern.** **OPEN (not queued.)** Needs Item 1.
- [ ] **Item 7 — `update add-req` and the `query gaps` selectors.** **OPEN (not queued.)** Needs Item 1.
- [ ] **Item 8 — the skill re-derived as each verb returns.** **OPEN (not queued.)**

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

## The requests exchange

`requests/trajectory-reader-requirements.md` (2026-09-04, completed): seven items their readers
owed, all landed the same day — write-path refusal (`DeviationError`, `ErrReopen`, `ErrLanded`),
`Line()` on every entry and part, lenient `OPEN` attribution reads, and a gap ID as a valid
`Source:` (their `Entry.Kind`, `SourceKey`, `ErrBadGapSource`). Item 7, several parts per entry,
is a door kept open. `requests/RESP-alarm-method-anchors.md` (in-progress) is this side's
answer to their one request; Item 6 above closes it.
