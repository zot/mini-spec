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

### Reader Agreement
- **The two readers of every design document the dependency owns must agree, and this
  check is the second opinion.** For the Gaps section, `requirements.md` and every test
  design, an independent line scan — a regex over lines, built on nothing the reader is
  built on, bounded for the gaps section by its heading and the next level-2 heading — is
  compared with what the reader returned: IDs for gaps and requirements, entry counts for
  test designs (a title is display, and the reader's may differ from the heading's). Every
  difference is a finding, listed first, because every check below reads through the
  document reader alone. A reader that lost a file's tail to one unclosed span agrees with
  itself forever; only a scan that shares none of its blind spots can say how much of the
  file it actually saw.
- *Measured the day it was written, 2026-09-07, on three repositories.* Here: agreement,
  after a stray backslash-backtick had hidden four of six alarms that morning. mini-spec-tool:
  two test designs short by three and two entries. ark: **`requirements.md` read 1419 of some
  2800 requirements** — R1323 onward gone behind a ``` quoted in prose at line 2157 — and five
  test designs short by up to thirteen entries each, behind lone `*` and `**` in prose. Every
  one was reported by the reader as a group never closed; none was visible from the counts.
- **What the readers could not read is printed as a coverage note**, naming each file and its
  count, whether or not anything else fired — never an issue in itself, because a reader takes
  silence about coverage as a claim of completeness.

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
  exists to prevent. **A prescription is checked too**: an alarm with an `**Inject:**`
  and no `**Pulled:**` is the shape *most* likely to have rotted, because nobody has run
  it, and until 2026-09-06 it was the only shape never looked at. Its sites are resolved
  at the cheap half of the cost — the extent, no history walk — before it reads
  `unrecorded`; a rotted one reads `unresolvable`. Measured on old-sdom's delegation rig:
  the one hard failure among nine pullers was a brief naming a symbol that did not exist,
  which the puller wrote for itself rather than refusing.
- **The reader computes the site's line range; git is asked only when those lines
  changed.** `git log -L :pattern:file` asks git to *find* the symbol as well as bound
  it, and git bounds a declaration at the line before the next one — so its range
  carried the successor's doc comment and the trailing blank line, the last declaration
  in a file read stale on every append, and a bare name's first match could be a use
  or a comment (measured 2026-09-04, the stopgap's gaps `O10` and `O11`). The range now
  comes from the dependency's Go declarations over the file **as committed at HEAD**,
  since `-L <start>,<end>` resolves against HEAD: **the declaring line through the line
  on which every bracket group opened inside it has closed**, a grouped `const` member
  at its own line, and **no comment in the range** — not the successor's, and not the
  declaration's own, which old-sdom measured wrong when three verified alarms went
  stale over traceability lines rewritten inside their doc blocks.
- **A method is named by its receiver — `Type.Method` — and that form resolves to
  exactly that method's declaration.** Three `Parse` methods on three types are one
  symbol to a bare name; the receiver is the only spelling that can tell them apart, and
  the receiver group between keyword and name is what it is matched on. **A symbol
  declared more than once in its file is reported unresolvable with the count and the
  repair**, never resolved to the first: the anchor has been watching an arbitrary one
  since it was written. Only Go sources have an extent today (gap `O21`).

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

## minispec validate trajectory

The consistency of the trajectory layer — the queue files at the repository root and the
carves that point at them. Read-only, and **a separate subcommand rather than part of bare
`validate`.**

**Why separate, and it is not a preference.** `validate` is *design-scoped*: it resolves a
design root and checks `design/` against `specs/` and the code. The trajectory layer is
*repository-scoped* — one repository holds one queue and may hold several design roots, as
this project does with `tool/` and `example/`. Folding a repository-scoped check into a
design-scoped command would run it once per design root and report the same drift twice,
and it is the project's own two-roots theme violated at the seam that theme was written
about. So this resolves the repository root and never a project, like `query next-id item`
and `query carves` before it. `make validate` runs both; the gate is the Makefile's job,
not a scoping compromise.

**A project running no trajectory layer passes.** No queue files and no carve directory is
not a failure — there is nothing that could be inconsistent. It says so and exits 0, which
is a different report from *could not check* and must not be confused with it.

### Referential integrity

Both directions, because each catches what the other cannot.

- **Carve → queue.** Every `#N` a carve's status block cites resolves to an entry in the
  pending file or the done file. A citation to an item that never existed, or to one whose
  ID was reused, is a pointer into nothing.
- **Queue → carve.** Every queue entry whose `Source:` names a carve appears in that
  carve's status block under the key it claims. An item can otherwise complete against a
  part the carve never recorded, which is how a carve comes to under-report its own work.

**Ingest by position, never by pattern.** What the tool reads as a citation comes from a
fixed position — a part line inside the status block, and the marker on it, as the
dependency's carve reader hands it over — and never from a pattern swept over prose. This is stated because the alternative was tried and
failed: an extractor sweeping `**VERB (…)**` across a whole status block read a prose
sentence quoting another project's shape, inside backticks, as a live citation and reported
a dangling `#121` in a repository that never had one. That was the fifth ad-hoc instrument
in this layer's history to return a confident wrong answer, and the first written by
someone who had read the record of the other four.

### Ledger and status checks

- **An item ID held by *both* the pending and the done file.** That is a live item and a
  completed one sharing a number, so every pointer to it is ambiguous.

  **Repetition *within* the done file is not a collision, and checking for it was wrong.**
  An item that lands in stages is legitimately recorded across several entries — measured
  2026-08-16 in ark, `#41` appears in five (`Pass 1`, `2a`, `2b`, `2c`, `CLOSED`) and `#65`
  and `#98` in two each — and from the number alone a staged record is indistinguishable
  from a reuse. The first draft of this check reported all three as reused IDs, which is a
  check firing on correct work: the shape that gets muted and then reports nothing at all.
  Only the cross-file case is decidable, and ark has none.
- **Orphans** — a carve part marked landed against a queue ID with no done entry, and a
  done entry naming a part that no carve records.
- **A done entry whose header carries no identifier slot** is reported as *unmigrated*
  rather than skipped. The slot is read between the date's em dash and the title's colon;
  a header with no colon yields nothing, and nothing is also the correct answer for an
  entry that legitimately discharged no ID, so the two are indistinguishable and the skip
  is silent today. Measured 2026-08-16: 2 of ark's 54 entries, hiding three queue IDs.
- **Item numbers that appear in no readable entry.** Every number from 1 to the maximum
  assigned should be accounted for, since an ID is assigned at creation. A deliberately
  abandoned ID is the expected exception and the format says so — but *expected* is a claim
  worth testing rather than assuming. Measured 2026-08-16: this project has none; ark has
  **17**, and **16 of them are mentioned in its queue files** — `#46` nineteen times, `#35`
  ten — so they are losses rather than abandonments.

- **Entry-like lines the reader could not recognize**, reported as coverage rather than as
  a defect. This is the honest half, and the gap list above is its symptom. A shape-based
  check is blind by construction to a line outside the shape: "does this entry carry an
  identifier slot?" cannot be answered for a line never recognized as an entry, so the
  check reports clean over everything it never saw. Measured 2026-08-16 in ark:
  **139 top-level entry lines in the done file, 54 recognized** — the other 85 put the date
  outside the bold, an older shape — so `query next-id item` reports what it read from 39%
  of the ledger with nothing saying so.

  When both fire, the gap report names the unread count, because it very likely explains
  them. A gap in a fully-readable ledger is a different and more interesting finding.

- **Checkbox agreement.** A status line states its state three ways — checkbox,
  strikethrough, marker — for three readers, and the checkbox is authoritative where they
  disagree. Measured 2026-08-16 the corpus is clean: no line among nine live carves
  disagrees with itself. That makes this a **sentry** rather than a repair — the corpus was
  normalised by hand, and nothing keeps it so.

- **A status-block line the reader could not read as a part and lists as deviating** is an
  issue, named with its file, line and the reader's reason. A checkbox-less line with no
  deviation — a `SPLIT` or `MOVED` parent — is the format's own shape and is not one. A carve that loses a part line loses it from
  `query carves` and from every check above at once, all of them then agreeing on a smaller
  number, so it cannot be a note. *On this branch the finding is the dependency's own
  `Stateless()` list*; the August tree also asked a **second, independent** flat-scan question
  of every status block, and that cross-check is not carried yet — see gap `O18`, whose home is
  queue item `#55`.

- **The two readers of the queue files must agree.** `next-id item`'s line scan and the
  dependency's document readers read the same two files; every item ID one saw and the other
  did not is a finding, listed first, because every finding below it reads through the
  document reader alone. *Found the day the check was ported, on this repository:* the line
  scan read 58 IDs from the done file, the document reader returned 17 entries and reported
  nothing unread — one unclosed backtick in an entry body absorbed the remaining 41 entries
  into a single text node — and four landed parts were reported as orphans. A second reader
  is the only thing that can see what the first one swallowed.

- **`CURRENT.md` carries exactly one `## Active`**, the region `pending finish` clears. The
  write path refuses both the missing heading and a duplicated one, but only when someone runs
  the verb; measured 2026-08-19 on the August tree, a corrupted file sat green until the next
  completion met it. The trajectory files are gitignored, so git cannot diff them, and the
  backup slot holds one level of undo which the next operation spends — a check running
  between operations is the only thing standing where a diff would normally stand. Reported
  before the reference-level findings, because a file whose own shape is wrong makes every
  statement about the references inside it a claim about a document nobody can trust. A
  missing current file is not this finding; absence is already answered.

- **Output.** Markdown to stdout, the global `--json` flag honoured with one key convention
  (snake_case), exit 0 when consistent and 1 when issues were found. Coverage notes — unread
  lines, unreachable citations — print whether or not anything else fired. The unread note
  counts per file across **every document a reader touched**: both queue files, the current
  file, and each carve, since a bracket group never closed in any of them takes that file's
  tail out of every check above.

### Conformance is named, not complained about

Where a document predates a decided format, the report says which target it should migrate
to rather than that it is wrong. A carve keyed on the superseded bare `#N` or `Part X` is
not ambiguous, it is **unmigrated**, and an agent given a target can act where one given a
complaint cannot.

This has a measured consequence worth stating: ark's carves yield only 3 marker citations,
not because they are clean but because they put queue IDs in *key* position. Most of that
project's references are unreachable to the integrity check until it migrates — which the
report says, rather than reporting a clean result over references it could not see.


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
