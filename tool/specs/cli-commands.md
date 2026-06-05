# CLI Command Inventory

**Status:** Summary spec. Indexes the full `minispec` CLI surface, with a link from each command to the per-feature spec that owns it. No new behavior is defined here — when this file and an owning spec disagree, the owning spec wins.

**Maintenance:** When you add, rename, or retire a subcommand, update this file by hand. Mini-spec's normal anchoring (specs → requirements → design → code) does not catch drift here.

## Top-level commands

| Command | Owning spec | Summary |
|---|---|---|
| `minispec check-version` | [config.md](config.md) | Verify tool binary version matches the skill's `README.md`. Exits 0 on match. |
| `minispec query <sub>` | [queries.md](queries.md) | Read design files and print results. No modifications. |
| `minispec update <sub>` | [updates.md](updates.md) | Atomic modifications to structured parts of design files. |
| `minispec validate` | [validate.md](validate.md) | Run all structural validations and report issues. |
| `minispec phase <sub>` | [phase.md](phase.md) | Phase-specific validation after each workflow phase. |

## `query` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `query project` | [queries.md](queries.md) | Show resolved project paths (root, design, src, specs). |
| `query requirements` | [queries.md](queries.md) | List all requirements with text and source spec. |
| `query coverage` | [queries.md](queries.md) | For each requirement, list referencing design files. |
| `query uncovered` | [queries.md](queries.md) | List requirements with no design file references. |
| `query orphan-designs` | [queries.md](queries.md) | List CRC cards missing or with empty Requirements field. |
| `query artifacts` | [queries.md](queries.md) | List all artifacts from design.md with checkbox states. |
| `query gaps` | [queries.md](queries.md) | List all items from the design.md Gaps section. |
| `query traceability [file]` | [queries.md](queries.md) | Check traceability comments in one code file. |
| `query traceability --all` | [queries.md](queries.md) | Scan all artifact code files and report traceability status. |
| `query migrations` | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | List in-flight migration specs under `specs/migrations/` (non-recursive). |
| `query comment-patterns` | [config.md](config.md) | Show recognized comment patterns and closers per file extension. |
| `query unindexed-specs` | [queries.md](queries.md) | List per-feature specs not referenced in the root index `specs/index.md` (exact `.md`-token match; lists all when index absent). |

## `update` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `update check [file] [item]` | [updates.md](updates.md) | Check a checkbox in the specified file. |
| `update uncheck [file] [item]` | [updates.md](updates.md) | Uncheck a checkbox. |
| `update add-ref [crc-file] [Rn]` | [updates.md](updates.md) | Add a requirement reference to a CRC card. |
| `update remove-ref [crc-file] [Rn]` | [updates.md](updates.md) | Remove a requirement reference from a CRC card. |
| `update add-gap [type] [description]` | [updates.md](updates.md), [migrations/complete/001](migrations/complete/001-migration-and-retirement.md) | Add a new gap with auto-numbered ID. Migration spec amends: `A` and `T` gaps are written without checkboxes. |
| `update resolve-gap [id]` | [updates.md](updates.md) | Mark a gap as resolved. Alias for `update check design.md [id]`. |
| `update approve-gap [id]` | [updates.md](updates.md), [migrations/complete/001](migrations/complete/001-migration-and-retirement.md) | Convert an existing gap to approved (A) type. Migration spec amends: A entries have no checkbox. |
| `update retire R<old> <R<new>\|-> "<reason>"` | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | Retire a requirement: strikethrough the Rn line, append a Tn gap. |
| `update migration-complete <name>` | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | Move `specs/migrations/<name>.md` to `specs/migrations/complete/<NNN>-<name>.md`. |

## `phase` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `phase spec` | [phase.md](phase.md) | Validate Spec Phase: at least one non-empty spec file exists. |
| `phase requirements` | [phase.md](phase.md) | Validate Requirements Phase: requirements.md parseable, Rn unique, sources resolve. |
| `phase design` | [phase.md](phase.md) | Validate Design Phase: design.md Artifacts complete, CRC Rn references valid. |
| `phase implementation` | [phase.md](phase.md) | Validate Implementation Phase: artifact code files exist with traceability comments. |
| `phase gaps` | [phase.md](phase.md) | Validate Gaps Phase: gap IDs follow S/R/D/C/I/O/A/T format, no duplicates. |

## Global flags

| Flag | Behavior |
|---|---|
| `--design-dir <path>` | Override design directory location. |
| `--src-dir <path>` | Override source directory location. |
| `--quiet` | Minimal output. |
| `--json` | Emit JSON output where supported (most `query` subcommands). |
| `--version`, `-v` | Print `minispec <version>` and exit. |
| `-h`, `--help` | Show top-level help. |
