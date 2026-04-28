# Migration: Migration Workflow and Requirement Retirement

**Language:** Go
**Environment:** CLI, single binary

## Problem

`SKILL.md` v2.6.0 introduces three additions the tool does not yet
support:

1. A migration workflow under `specs/migrations/` (in-flight) and
   `specs/migrations/complete/` (history, numbered in landing
   order). The skill makes a "scan for in-flight migrations" check
   mandatory before every phase, but the AI has no tool command to
   run for it — the present design relies on the model remembering
   to scan a directory.
2. A new gap kind, **Tn (retired)**, that has no checkbox (it is
   permanent, not track-to-completion) and points the original `Rn`
   to its replacement (or to "no replacement"). Retired requirements
   stay in `requirements.md` with a strikethrough marker so old
   design and code references still resolve. The current parser
   regex (`^- \[([ x])\] ([SRDCIOA])(\d+):`) rejects both the
   checkbox-less line and the `T` type, and the requirement regex
   rejects the `**~~Rn:~~**` form.
3. Approved gaps (An) are described in the skill as "never checked
   off" — they should not have a checkbox marker either. Existing
   projects have `- [ ] A1: ...` lines that need to migrate to
   `- A1: ...`. Tn entries follow the same shape from the start.

In addition, `minispec validate` output is verbose enough to defeat
the tool's primary purpose (saving tokens). A run on a real
project (`/tmp/validate`) emits long enumerations of every
`R{found,covered,uncovered}` ID, repeats identical issue strings up
to a dozen times, and shows successful sections that contain no
information for the AI.

## Goal

Bring the tool fully in line with `SKILL.md` v2.6.0:

- Parse and validate the new file shapes (Tn, checkbox-less An/Tn,
  strikethrough retired Rn).
- Add commands for the workflows the skill describes, so the AI
  can run a single command instead of doing multi-step file edits:
  `query migrations`, `update retire`, `update migration-complete`.
- Reformat `validate` output so it shows only what needs fixing,
  uses requirement-ID ranges, and deduplicates repeated messages.
- Crank-handle existing `[ ] An` lines toward the no-checkbox form
  by surfacing them as a validate issue.

## Public CLI Surface (additions)

```
minispec query migrations
    List in-flight migration specs (specs/migrations/*.md, not recursive
    into complete/). One file path per line. Empty output when no
    migrations are in flight. Exit 0 either way.

minispec update retire R<old> <R<new>|-> "<reason>"
    Atomically:
      - rewrite the R<old> line in requirements.md to
        `- **~~R<old>:~~** (Retired T<n> — see R<new>) <original text>`
        (the literal "no replacement" replaces the `see R<new>` clause
        when the second arg is `-`)
      - append `- T<n>: R<old> retired by R<new> (<reason>)` to
        the Gaps section of design.md, where T<n> is the next free
        T-number.
    Returns the assigned T<n> on stdout.

minispec update migration-complete <name>
    Move specs/migrations/<name>.md to specs/migrations/complete/<NNN>-<name>.md
    where NNN is the next zero-padded three-digit number after the
    largest existing prefix in complete/. Returns the new path on
    stdout.

minispec update add-gap T "<reason>"
    Add a new T-typed gap line. Tn entries are written without a
    checkbox.

minispec update add-gap A "<desc>"
    A-typed gap lines are written without a checkbox. (Behavior
    change: previously `[ ] A<n>: ...`.)

minispec update approve-gap <id>
    The resulting An line is written without a checkbox.
```

## File-format Changes

### requirements.md

Retired lines look like:

```
- **~~R5:~~** (Retired T1 — see R10) <original text>
- **~~R7:~~** (Retired T2 — no replacement) <original text>
```

The parser must recognize both the unmarked and the strikethrough
form, capturing a `Retired bool` on the parsed `Requirement`. The
text after the marker is the original requirement text, preserved
verbatim so existing references still mean the same thing.

### design.md Gaps

Tn entries:
```
- T1: R1598 retired by R1833 (2026-04-23 ec-rekey)
  - reason: EC keys moved from (fileID, chunkIdx) to chunkID
- T2: R1099 retired by R1281 (2026-04-09 tag-embeddings)
- T3: R42 retired (no replacement)
```

An entries:
```
- A1: R148-R154 covered by app design
```

Both have no leading `[ ]` or `[x]`. Indented sub-bullets after a
Tn (such as `  - reason: ...`) are part of the Tn block and are
preserved during writes; the tool does not need to parse them.

For backwards tolerance the parser still accepts the legacy
`- [ ] A<n>: ...` form, but `validate` reports a "checkboxed
permanent gap" issue so the AI cleans them up.

## Validate Output

Issues-only, deduplicated, ranged. Successful runs print one line:

```
phase: validate OK
```

Failing runs print only the categories that have problems:

```
issues:
  uncovered requirements: R3, R7-9
  missing impl coverage: R44-72, R83-85
  duplicate requirements: R177-181
  unknown CRC refs: crc-Auth.md → R99
  missing artifacts: src/view.ts (listed but not found)
  missing traceability: src/foo.go, src/bar.go
  missing design refs: internal/mcp/tools.go → crc-MCPTool.md (R169-176); internal/mcp/server.go → seq-mcp-lifecycle.md (Scenario 3)
  unlisted design files: seq-foo.md, ui-bar.md
  missing spec sources: specs/old-feature.md
  CRC sequences not found: crc-Foo.md → seq-bar.md
  permanent gaps with checkbox: A1, T2

phase: validate FAILED
```

Range rule: numerically sort and dedup; collapse runs of two or
more consecutive numbers into `R<lo>-<hi>` (no `R` on the upper
bound). Singletons stay as `R<n>`.

Phase commands follow the same shape. Wherever Rn lists appear
(findings *and* issues — e.g. `phase requirements` reporting
"found:" or per-source breakdowns), they use ranges and dedup,
never enumerate every Rn. The default phase output is one summary
line plus any sparse findings the AI cannot get cheaper from a
`query` subcommand.

## Behaviour Changes

- **Coverage check:** retired requirements are excluded from the
  `Coverage.Uncovered` list and from the implementation-coverage
  check. Their Rn IDs remain valid for cross-reference resolution.
- **Spec phase:** `phase spec` already ignores subdirectories under
  `specs/`, so it is unaffected by `migrations/` and `complete/`.
- **Mandatory pre-phase scan:** `query migrations` is the canonical
  way the AI checks for in-flight work; SKILL.md is updated to
  call it instead of describing a directory scan.

## Versioning

This release ships as **mini-spec 2.6.0**. The file format changes
are additive (older parsers reading new files would fail; newer
parsers reading older files still work), so there is no migration
of existing checked-in design data — the only legacy artefact is
the `[ ] An` checkbox shape that validate flags during normal
work.
