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

- [ ] **Item 1 — wire `minispecsdom.Comments` into the harvest; retire `parser.ParseTraceability`.** **OPEN (not queued.)** Re-point `validate.go`'s impl-coverage harvest and `query.go`'s code-ref reader onto the sdom reader. Build one positioned harvest both consume.
- [ ] **Item 2 — `minispec query implementation` (ark's request).** **OPEN (#92.)** Reverse lookup `Rn` (or a pattern) → the code that implements it, over the sdom harvest. Rides on Item 1's positioned harvest.
- [ ] **Item 3 — `RangeSet` on `RequirementList`.** **OPEN (not queued.)** An interval view (`Ranges() [][2]int` / `Contains(n)`) beside `Items()`. Enhancement, not a correctness fix — see Decisions.

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
of Item 2.

## The CLAUDE.md obligation

Adding a `query` subcommand trips the summary-spec rule: `tool/specs/cli-commands.md`
gains the `implementation` entry in the same pass. And the feature runs the full
mini-spec phases, including a `test-*.md` + fire alarm for the classifier and the
positioned harvest (cheap, deterministic — no infra).
