# Carve: the sdom conversion — one model per file, tracked across sub-carves

> **DRAFT (Daneel, 2026-09-18) — for Bill to edit before its first part is worked.**

Mini-spec is mid-conversion from ad-hoc, lossy readers to **sdom** — the
position-preserving DOM in `github.com/zot/simple-dom`, with mini-spec's document
knowledge in `internal/minispecsdom`. The payoff is whole-file: a parse that
round-trips, so a file can be *annotated* and *edited through its structure* instead
of rewritten. This master carve tracks the conversion one file-kind at a time. It is
long-running by design; each kind's own work gets a **sub-carve** as we reach it, and
this carve points at it.

**Provenance.** The DOM was built in `~/work/mini-spec-tool` as `simple-dom` plus the
`minispecsdom` readers; the package moved home into `tool/internal/minispecsdom` at
`#80` ([minispecsdom-move.md](done/minispecsdom-move.md)); the tool reclaimed over it
on branch `new-sdom` at `#67`–`#75` ([sdom-reclaim.md](done/sdom-reclaim.md)). The
per-kind landings below are **subsumed in that package/reclaim work rather than
discrete commits** — the LANDED refs point at the done carve that holds each, not at a
per-kind commit, because there isn't one.

**What "converted" means here.** The old `internal/parser` package has become a thin
file-path *wrapper* over `minispecsdom` for the converted kinds (`parser/trajectory.go`
→ `minispecsdom.ParsePending/Current/Done`, and so on). "Converted" = the parsing goes
through sdom behind that wrapper. The wrapper is not deleted — see Item 7.

## Status

- [x] ~~**Item 1 — trajectory files (PENDING/CURRENT/DONE).**~~ **LANDED (2026-09-14 — `#80`.)** `parser/trajectory.go` → `minispecsdom.Parse{Pending,Current,Done}`.
- [x] ~~**Item 2 — requirements.md.**~~ **LANDED (2026-09-14 — `#80`.)** `parser/requirements.go` → `minispecsdom.ParseRequirements`.
- [x] ~~**Item 3 — carves.**~~ **LANDED (2026-09-04 — `#67`.)** `parser/carve.go` → `minispecsdom.ParseCarve`.
- **Item 4 — design.md (one model, typed views).** **SPLIT (Bill, 2026-09-18.)** No checkbox: the sub-items carry the state. One sdom model per file; Gaps, Artifacts and Intent are views over it, never separate parses — see Decisions.
  - [x] ~~**4.1 — Gaps view.**~~ **LANDED (2026-09-14 — `#80`.)** `minispecsdom.ParseGaps`. Landed as a standalone parse; joining it to Item 4's one model is part of 4.5 — see body.
  - [x] ~~**4.2 — test designs (`design/test-*.md`).**~~ **LANDED (2026-09-14 — `#80`.)** `minispecsdom.ParseTestDoc`.
  - [ ] **4.3 — CRC cards (`design/crc-*.md`).** **OPEN (not queued.)** Still `parser.ParseCRCCard`, no sdom reader.
  - [ ] **4.4 — sequence diagrams (`design/seq-*.md`).** **OPEN (not queued.)** Still `parser.ParseSeqDoc`, no sdom reader.
  - [ ] **4.5 — the Artifacts manifest.** **OPEN (not queued.)** Still `parser.ParseArtifacts`. Must land as a view over Item 4's one model, joining 4.1.
  - [ ] **4.6 — UI layouts (`design/ui-*.md`).** **OPEN (not queued.)** No structural reader today.
- [ ] **Item 5 — code files (traceability comments).** **OPEN (not queued.)** The first open item. Its work lives in [traceability.md](traceability.md).
- [ ] **Item 6 — `specs/*.md`.** **OPEN (not queued.)** Scope to settle — see body.
- [ ] **Item 7 — retire the old `parser` wrappers.** **OPEN (not queued.)** End-state to settle — see body.

## Decisions

**DECIDED (Bill, 2026-09-18): one sdom model per file; sections are typed views, not
separate parses.** The user thinks in files, and annotation and round-trip are
whole-file operations — an annotated design.md cannot be two independent parses
stitched back together, and the user never thought of design.md as two files. So
design.md is one model with Gaps / Artifacts / Intent views, the same shape
`TraceabilityComment` already uses (`CRC`/`Seq`/`Test`/`Refs`/`Description` as views
over one node). This is why Item 4 is one item, not two.

**DECIDED (Bill, 2026-09-18): a view is an order-independent re-granulation pass over
one base parse, applied lazily.** The file is parsed once into a base DOM; each view
(Gaps, Artifacts, Intent) is a pass that splits nodes and rehomes them into inserted
view-nodes — the re-granulation the readers already do (`FindTraceComments`,
`FindDecls`, `Comments`), boundaries moving while bytes and provenance do not. The
required property is **confluence**: starting from the base parse and applying any
subset of views in any order yields the same DOM. That is what makes annotation over
any combination of views well-defined, and what lets a consumer **apply only the views
it needs** — impl-query applies the ref view alone, a gaps report applies Gaps alone,
neither pays for the other.

What must hold is confluence. **How** we get it today is the cheap way — views over
non-overlapping regions, so no two passes contend for the same nodes — but that is a
current simplifying choice, **not a property of the model**. Overlapping views are
possible (a syntax-highlighting view interleaving with declaration re-homing;
`sdom.DeclarationName`'s `Text` node retargeting to a text-behavior interface that
several views share), and nothing here forecloses them; we have simply not earned the
need, so we do not build for it yet. Confluence fails *silently* — order-dependence is
intermittent and reads as working — so Item 4 owes a fire alarm that applies the views
in several orders and asserts an identical DOM. This is uniform across file kinds:
`Comments` is already such a pass over the code base parse.

**DECIDED (Bill, 2026-09-18): sub-carves are made lazily, as the conversion reaches
each item.** Only this master carve and the currently-active sub-carve are live at
once; a finished sub-carve moves to `done/` and this carve points there. That keeps the
live-carve count small, which is the whole point of a carve.

## Item 4 — design.md

The target is **one `minispecsdom` model of the design.md file**, with `Gaps()`,
`Artifacts()` and `Intent()` as typed views over the shared node tree — not the
present split, where `minispecsdom.ParseGaps` and `parser.ParseArtifacts` parse the
same file independently. The satellite design files (crc-, seq-, test-, ui-) are
separate files with their own readers; the Artifacts manifest is design.md's index of
them.

**4.1 — Gaps view.** Gaps parsing is on sdom (`ParseGaps`), but as a *standalone* parse
of the gaps region, not a view over a whole-design.md model. Reaching the one-model
target means the Artifacts reader (4.5) and the Gaps reader share one parse; whether
that re-homes `ParseGaps` or wraps it is a call for whoever lands 4.5.

**4.5 — the Artifacts manifest.** The remaining design.md section still on the old
parser. It carries the item↔code checkboxes, so its conversion is where the one-model
design.md actually lands — 4.1 folds in here.

## Item 5 — code files (traceability)

The code-comment reader `minispecsdom.Comments` / `TraceabilityComment` is **written and
tested but wired to nothing**; `validate` and `query` still harvest inline `Rn` refs
through the regex `parser.ParseTraceability`. Converting code files = wire `Comments`
into the harvest and retire `ParseTraceability`. This is also the reader that unlocks
positioned harvest (`file:line`), ref rewriting, and the reverse lookup ark asked for —
so the traceability sub-problem is a carve of its own: [traceability.md](traceability.md).

## Item 6 — `specs/*.md`

`@undecided:` (resolution: settle when reached) — do specs need structural sdom reading
at all? Today specs are indexed and their `**Source:**` links checked; there may be
little document structure to model. Decide the scope before opening the item.

## Item 7 — retire the old `parser` wrappers

`@undecided:` (resolution: decide when Items 4–6 are near done) — with every kind on
sdom, the old `parser` package is only a file-path wrapper over `minispecsdom`. The
end-state could retire the wrappers (consumers call `minispecsdom` and own their file
IO), or keep them as a deliberate file-IO adapter. Not a decision to force now; named so
it is not lost.
