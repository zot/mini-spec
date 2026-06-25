# File Format Inventory

**Status:** Summary spec. Indexes every file the tool reads or writes, with a link to the per-feature spec that owns each format. No new behavior is defined here — when this file and an owning spec disagree, the owning spec wins.

**Maintenance:** When you add, change, or retire a file format the tool handles, update this file by hand. Mini-spec's normal anchoring does not catch drift here.

## Design files (`design/`)

| File | Purpose | Owning spec(s) |
|---|---|---|
| `design.md` — Artifacts section | Inventory of design files paired with the code files that implement them. Inline-checkbox form: `- [x] crc-Foo.md → src/foo.go`. | [config.md](config.md) (canonical format), [validate.md](validate.md) (structural rules), [queries.md](queries.md) (`query artifacts`) |
| `design.md` — Gaps section | Open issues raised during design. Each line: `- [ ] S1: description` or `- A1: description` (no checkbox for `A`/`T` types). | [updates.md](updates.md) (`add-gap`, `resolve-gap`, `approve-gap`), [validate.md](validate.md) (structural rules), [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) (no-checkbox `A`/`T` form, `Tn` retirement entries) |
| `requirements.md` | Requirements grouped by feature with source-spec backref. Each requirement: `- **R5:** text` or `- **R5:** (inferred) text`. Retired form: `- **~~R5:~~** (Retired T1 — see R10) text`. | [validate.md](validate.md) (parse and format rules), [phase.md](phase.md) (`phase requirements`), [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) (retired form) |
| `crc-*.md` | CRC card. Heading sets class name; `**Requirements:** R1, R3, R7` line links to requirements; optional `## Sequences` section lists related `seq-*.md` files. | [validate.md](validate.md) (Requirements field rules, Sequences validation), [updates.md](updates.md) (`add-ref`, `remove-ref`) |
| `seq-*.md` | Sequence diagram. Optional dotted-number anchors (`1.`, `1.1`, `1.1.1.`) within tree/UML/lane drawings. Numbering is opt-in per file; once present it must be contiguous from `1.` per diagram. | [validate.md](validate.md) (numbering rules, anchor resolution) |
| `ui-*.md`, `test-*.md`, `manifest-*.md` | Other design artifacts. Tool tracks presence/listing only — content is not parsed. | [validate.md](validate.md) (Artifacts manifest completeness) |

## Spec files (`specs/`)

| File | Purpose | Owning spec(s) |
|---|---|---|
| `specs/*.md` | Per-feature specs. Referenced from `requirements.md` via `**Source:** specs/foo.md`. Tool validates the path resolves; content is free-form. | [validate.md](validate.md) (Spec Source Validation), [phase.md](phase.md) (`phase spec`) |
| `specs/migrations/*.md` | In-flight migration specs (non-recursive). Listed by `query migrations`. Content is free-form. | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) (`query migrations`) |
| `specs/migrations/complete/<NNN>-*.md` | Completed migrations, renumbered in landing order. `update migration-complete` performs the move. | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) (`update migration-complete`, Source migration-completion fallback) |

## Code files

| Format | Purpose | Owning spec(s) |
|---|---|---|
| Traceability comment | One or more lines like `// CRC: crc-Foo.md \| Seq: seq-bar.md#1.2 \| R5, R12` in code files listed under Artifacts. Pipe-separated sections, third section optional. Block-comment languages require a configured closer. A **bare annotation** — a comment that leads with a requirement ref (`// R5: desc`, `// R5, R6`, trailing `foo() // R7`) — also supplies inline Rn refs; a ref not leading the comment (prose `// see R5`) does not count. Both the `\| Rn` tail and the bare annotation expand `Rn-Rm` ranges (`// R5-R8` covers R5 through R8). | [validate.md](validate.md) (Traceability Comments, Sequence Anchor Validation, Implementation Coverage), [config.md](config.md) (`comment_patterns`, `comment_closers`), [queries.md](queries.md) (`query traceability`, `query comment-patterns`) |

## Configuration & metadata

| File | Purpose | Owning spec(s) |
|---|---|---|
| `.minispec.yaml` | Optional project config: `design_dir`, `src_dir`, `comment_patterns`, `comment_closers`. | [config.md](config.md) |
| `.claude/skills/mini-spec/README.md` | Skill version source. `check-version` reads the `Version:` line (project-level first, then `~/`). | [config.md](config.md) (Version section) |
