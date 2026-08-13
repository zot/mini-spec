# Validate Command

Structural validation of design files. Checks format and references, not intent.

## minispec validate

Run all validations and report issues.

## Checks Performed

### Requirements Format
- requirements.md exists
- All requirements have Rn format with unique numbering (no duplicates, no gaps; file order doesn't matter)
- Each requirement has a Source spec reference
- Inferred requirements are marked

### CRC Card Format
- Every crc-*.md has a **Requirements:** field
- Requirements field references valid Rn identifiers
- No duplicate Rn references within a card

### Artifacts Structure
- design.md has Artifacts section
- All listed design files exist
- All code file paths are valid
- Checkbox syntax is correct

### Gaps Structure
- design.md has Gaps section
- Gap IDs follow S/R/D/C/I/O/A + number format
- No duplicate gap IDs

### Approved Gap Coverage
- Approved (A-type) gaps may reference requirements via Rn or Rn-Rm ranges in their description
- Requirements referenced by approved gaps are treated as covered for validation purposes
- These requirements do not appear in the "uncovered" list or trigger "uncovered requirements" issues

### Traceability Comments
- Code files in Artifacts have `// CRC:` comments
- Referenced CRC and Seq files in code comments exist in design/
- Optional third pipe-delimited section contains inline requirement refs: `// CRC: crc-X.md | Seq: seq-Y.md | R5, R12`
- Parser extracts Rn refs from the third section (comma-separated)
- Inline Rn refs may also come from a **bare annotation**: a comment whose first token after the comment leader is a requirement ref (e.g. `// R5: description`, `// R5, R6`, or a trailing `foo() // R7`). The leading comma-separated refs are collected. This counts the deliberate field/line annotations that sit alongside a type's or function's `// CRC:` header.
- Both the `| Rn` tail and the bare annotation expand `Rn-Rm` **ranges** into every member, so a range-form annotation (`// R5-R8`) counts R5, R6, R7, and R8. Ranges and comma lists may be mixed (`// R5-R7, R10`).
- A ref that does not immediately follow the comment leader — a prose mention like `// see R5` or `// computed lazily (R5)` — is **not** counted; only a comment that leads with the ref signals intent.
- Inline Rn refs are validated: each must exist in requirements.md

### Implementation Coverage
- Every requirement in requirements.md should appear as an inline Rn ref in at least one code file's traceability comment
- Requirements covered only at the design level (CRC card) but not in any code file are reported as implementation gaps (I-type)
- Requirements covered by approved gaps (A-type) are excluded from this check
- Implementation coverage is reported in validate output alongside design coverage

### Artifacts Manifest Completeness
- All `crc-*.md`, `seq-*.md`, `ui-*.md`, `test-*.md`, `manifest-*.md` files in `design/` are listed in Artifacts section
- Detects orphaned design files not tracked in design.md

### Spec Source Validation
- `**Source:**` fields in requirements.md reference files that exist in `specs/`
- Validates the requirements→specs traceability link
- A Source line may carry a comma-separated list of paths (`**Source:** a.md, b.md`); each is validated independently
- Source values are checked for clean shape: relative path, ends in `.md`, no whitespace/parens/backticks/leading-slash/annotations. Malformed values are reported separately from missing files.
- Lines that look like Source markers but don't match the canonical pattern (e.g. `**Source**:`, `*Source:*`, `Source:`) are reported as suspicious so the author can fix them.
- When malformed values or suspicious lines are present, validate appends a `fix instructions:` block at the bottom of the output describing the canonical Source format.
- When a Source path is **missing**, validate appends a `fix instructions:` block naming the three legitimate repairs. It does so because the obvious repair is the one that must never be taken: deleting the orphaned requirements opens a numbering gap, and closing that gap by renumbering silently repoints every design and code anchor at a different requirement — a failure no check can detect, since all the numbers still exist. The block says:
  - **spec renamed** → rewrite the `**Source:**` to the new path
  - **spec merged into another** → repoint the `**Source:**` at the absorbing spec; the requirements live on and nothing retires
  - **spec deleted outright** → retire each of its requirements with `update retire`, regroup them under a `**Source:** specs/deleted.md` block, and record the dead spec's name, a one-line description, and its requirement numbers in `specs/deleted.md`
  - and, in every case: requirement numbers are never renumbered and never reused
- Migration-completion fallback: a Source value like `specs/migrations/X.md` resolves to `specs/migrations/complete/<NNN>-X.md` (NNN digits) when the literal path is missing but the renumbered completed file exists. This means `update migration-complete` does not require rewriting `**Source:**` lines in requirements.md.

### CRC Sequences Validation
- Files listed in CRC card `## Sequences` sections exist in `design/`
- Validates CRC→sequence traceability

### Sequence Anchor Validation
- A `Seq:` reference may carry a `#K.x.y` fragment pointing at a numbered step.
- When a fragment is present, the referenced sequence file is parsed for numbered items and the fragment must resolve to one of them.
- Missing fragments are reported per code file alongside other missing design refs.

### Sequence Numbering Validation
- A sequence file is "numbered" if it contains any line whose first non-whitespace, non-lane content is a dotted-number token (`1.`, `1.1`, `1.1.1.`, etc.). Lane characters allowed before the number: `│ ├ └ ─ |` and ASCII tree connectors.
- For each numbered sequence file, the tool:
  - Groups all dotted IDs by their first segment K (the diagram index).
  - Verifies the set of K values is contiguous starting at 1 (no gap between diagrams).
  - For each K, verifies the tree of children is contiguous at every level (K.1, K.2, ..., K.N with no holes; under K.x, children must be K.x.1, K.x.2, ...).
  - Verifies every dotted ID appears at most once in the file.
- Unnumbered sequence files are silently skipped — numbering is opt-in per file.
- Numbering gaps and duplicates are reported per sequence file.

### Fire Alarm Freshness

A `test-*.md` may record a fault injection that proved one of its tests — `**Fire
alarm:**` describing the injection, `**Inject:**` naming the `file:symbol` sites it
edits, and `**Pulled:**` giving the date it was run. A proof obtained against code
that has since been rewritten is void, and nothing about a green suite says so.

- For every alarm carrying **both** `**Inject:**` and `**Pulled:**`, the tool asks git
  whether any listed symbol has changed since that date, and reports the alarm as
  **stale** when one has.
- The question is asked of the **function**, not the file. A file-level check reports
  every alarm in a busy file as stale and so reports nothing at all; git can answer
  "has this function changed" directly, and that is the granularity the claim is about.
- **A change must be dated strictly after the pull.** A date is a coarser clock than
  git's history, and the first draft of this rule counted same-day changes as stale on
  the argument that over-reporting is the safe direction. Running it proved otherwise:
  the normal workflow is to fix the code, pull the alarm and commit both together, so
  the code's last change and the pull share a date for *every freshly recorded alarm*.
  The rule marked two of ark's three verified alarms stale the day they were written.
  A check that fires on arrival is ignored, which is a worse failure than the blind
  spot it was avoiding — a change made later the same day is missed, and caught by the
  next change on any later day.
- An `**Inject:**` naming a symbol git cannot find is reported as **unresolvable**,
  not silently skipped. That is the anchor rotting, which is the failure the field
  exists to prevent.

**Only stale alarms are reported here, and that is deliberate.** An alarm with
`**Inject:**` but no `**Pulled:**` is a *prescription* — an injection someone wrote
down and may never have run — and an alarm with neither is unanchored. Both are worth
knowing and neither belongs in `validate`: a project adopting the convention has many
of each, the counts fall slowly, and a line that reports a non-zero number every run
for months is the recurring nag this project distinguishes from a closable gripe.
Those two live in `minispec query alarms`. Stale is the closable one — it should
normally read zero, and a non-zero reading is a specific worklist.

**Silent without git**, like every other check that needs it. A tree with no
repository cannot answer the question, and a check that could not look must not
return a clean result.

## Output

Show what was found so the AI can verify assumptions and correct mismatches:

```
requirements.md:
  found: R1, R2, R3, R4, R5, R6, R7, R8
  sources: specs/auth.md (R1-R3), specs/storage.md (R4-R8)
  inferred: R3, R7

design files:
  crc-Store.md: R4, R5, R6
  crc-View.md: R1, R2
  crc-Auth.md: (no Requirements field)
  seq-login.md: (sequences don't have requirements)

coverage:
  covered: R1, R2, R4, R5, R6
  uncovered: R3, R7, R8

artifacts:
  crc-Store.md:
    [x] src/store.ts
    [ ] src/store_test.ts
  crc-View.md:
    [x] src/view.ts (file missing)

gaps:
  [ ] S1: ...
  [x] D1: ...

issues:
  - crc-Auth.md: no Requirements field
  - src/view.ts: listed in artifacts but file missing
  - R3, R7, R8: no design coverage
  - seq-logout.md: not listed in Artifacts
  - specs/old-feature.md: referenced as Source but file missing
  - src/store.ts: references crc-Missing.md which does not exist
  - crc-Store.md Sequences: seq-missing.md does not exist
```

The "found" lists let the AI verify parsing matched expectations. If formatting is unusual but parseable, the AI sees what was extracted and can decide if corrections are needed.

Exit code: 0 if no issues, 1 if any issues found.
