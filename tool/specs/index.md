# Spec Index

The root index for the `minispec` tool. Every per-feature spec appears here under a
system; the summary specs and cross-cutting themes are registered below. Entries are
pointers — the named spec stays canonical, this file mirrors it.

## Systems

### Orientation — how the tool locates what it operates on

| spec | covers |
|---|---|
| [config.md](config.md) | design-root detection, the two config scopes and their inheritance, comment patterns and closers, global flags, version reporting |
| [repository-root.md](repository-root.md) | repository-root detection, the two-root distinction, marker ranking, the home-directory exclusion |
| [initialization.md](initialization.md) | `init` and `--repair`, the `track` setting and its startup check, the no-configuration refusal, git preferences |

### Commands — what the tool does once oriented

| spec | covers |
|---|---|
| [overview.md](overview.md) | what the tool is for and how the pieces fit |
| [queries.md](queries.md) | the read-only `query` subcommands |
| [updates.md](updates.md) | the `update` subcommands — atomic edits to structured parts of design files |
| [validate.md](validate.md) | full validation and its check classes |
| [phase.md](phase.md) | per-phase validation gates |
| [backup.md](backup.md) | the backup slot beneath the queue verbs — one level of undo and redo over the trajectory files, the stamp, drift refusal, release, and the worktree anchor |
| [queue-items.md](queue-items.md) | the `pending` verbs — `add-item`, `start`, `finish` — that write the trajectory files and both sides of the item↔part link, over `minispecsdom`'s readers |

### Document model — the readers and writers over the markdown the tool owns

Ported from `mini-spec-tool` on 2026-09-14 ([carve](../../carves/done/minispecsdom-move.md)); built
over `github.com/zot/simple-dom`, which stays a module dependency.

| spec | covers |
|---|---|
| [traceability-comment.md](traceability-comment.md) | the `// CRC: … | Seq: … | Rn` comment as a declaration: its segments, keywords, and canonical write |
| [part-line.md](part-line.md) | a carve's part line: head, markers, trailing text, the strike, and the marker rule |
| [carve-schema.md](carve-schema.md) | the carve document: status block, parts, decisions, `Land`, `Open`, `SetMarker` |
| [pending-schema.md](pending-schema.md) | the pending file: entries, `Place`, `Remove`, the source slot |
| [done-schema.md](done-schema.md) | the done file: entries and their identifier slot |
| [current-schema.md](current-schema.md) | the current file: the `## Active` section and standing context |
| [testdoc-schema.md](testdoc-schema.md) | a `test-*.md`: test entries and the fire-alarm fields |
| [gaps-schema.md](gaps-schema.md) | `design.md`'s gaps list: add, resolve, approve |
| [requirements-schema.md](requirements-schema.md) | `requirements.md`: feature sections, entries, retire |

### Migrations

In-flight migration specs live in [migrations/](migrations/); completed ones are
numbered in landing order under `migrations/complete/`.

| spec | covers |
|---|---|
| [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | the migration workflow, requirement retirement, and the `A`/`T` no-checkbox gap shapes |

## Summary specs (cross-cutting inventories)

These index behavior owned by the per-feature specs above along one axis. They are
**mirrors, not sources** — where a summary and a per-feature spec disagree, the
per-feature spec wins. Mini-spec's normal anchoring does not catch drift in them, so
they are maintained by hand; `CLAUDE.md` carries the standing instruction.

| spec | axis |
|---|---|
| [cli-commands.md](cli-commands.md) | every subcommand and global flag |
| [file-formats.md](file-formats.md) | every file the tool reads or writes |

## Themes

**Two roots, one word.** "Project root" names both the design root and the
repository root, and the ambiguity has cost real design time. Touched by
[config.md](config.md) and [repository-root.md](repository-root.md), and by
`check-version` in particular. The invariant: repository-scoped artifacts
(`.claude/`, `carves/`, the trajectory files) anchor at the repository root;
project-scoped ones (`design/`, `specs/`, `src/`) anchor at the design root. Neither
search substitutes for the other, and prose that says "project root" without saying
which one is a defect.

**Report absence as error, never as silence.** Several checks can fail in a way that
looks like a pass — a version check that consulted the wrong file still prints a
verdict, a coverage scan that missed a range still reports a number. Touched by
[validate.md](validate.md), [phase.md](phase.md), and
[repository-root.md](repository-root.md). The invariant: a check that could not
examine what it claims to have examined says so, rather than returning a clean
result.
