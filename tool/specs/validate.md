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
