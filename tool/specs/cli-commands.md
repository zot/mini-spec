# CLI Command Inventory

**Status:** Summary spec. Indexes the full `minispec` CLI surface, with a link from each command to the per-feature spec that owns it. No new behavior is defined here — when this file and an owning spec disagree, the owning spec wins.

**Maintenance:** When you add, rename, or retire a subcommand, update this file by hand. Mini-spec's normal anchoring (specs → requirements → design → code) does not catch drift here.

## Top-level commands

| Command | Owning spec | Summary |
|---|---|---|
| `minispec init --track-<style>` | [initialization.md](initialization.md) | Create `<repo root>/.minispec/config.yaml` recording `track`, and write the ignore lines that value implies. Sole creator of the repository configuration; the flag is mandatory. Refuses when `.minispec/` already exists. |
| `minispec init --track-<style> --repair` | [initialization.md](initialization.md) | Set `track` on an existing configuration *and* bring `.gitignore` into agreement with it. Requires `.minispec/` to exist. The agent confirms the value with the user before running it. |
| `minispec init carve <name>` | [initialization.md](initialization.md) | Scaffolds `carves/<name>.md` with its status block holding one open part; refuses an existing file (R334). |
| `minispec check-version` | [config.md](config.md), [repository-root.md](repository-root.md) | Verify tool binary version matches the skill's `README.md`, looked up under the repository root first, then `~/`. Exits 0 on match. Runs without a configuration. |
| `minispec query <sub>` | [queries.md](queries.md) | Read design files and print results. No modifications. |
| `minispec update <sub>` | [updates.md](updates.md) | Atomic modifications to structured parts of design files. |
| `minispec validate` | [validate.md](validate.md) | Run all structural validations and report issues. |
| `minispec validate trajectory` | [validate.md](validate.md) | Check the trajectory layer — the queue files at the repository root against the carves that point at them — in both directions, plus the ledger, numbering, marking-agreement and current-file shape checks. Repository-scoped, never a design root; `make validate` runs both. |
| `minispec phase <sub>` | [phase.md](phase.md) | Phase-specific validation after each workflow phase. |
| `minispec pending <sub>` | [queue-items.md](queue-items.md), [backup.md](backup.md) | Trajectory-item operations, and the backup slot beneath them. `add-item` / `start` / `finish` write the trajectory files and both sides of the item↔part link; `revert` / `replay` give one level of undo and one of redo. Resolves the **repository root**, never a design root. |

## `query` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `query project` | [queries.md](queries.md), [repository-root.md](repository-root.md) | Show resolved paths: repository root, design root, design, src, specs. States when the two roots are the same directory. |
| `query config` | [queries.md](queries.md), [config.md](config.md) | Show every effective setting with its value and the file that supplied it. |
| `query requirements` | [queries.md](queries.md) | List all requirements with text and source spec. |
| `query coverage` | [queries.md](queries.md) | For each requirement, list referencing design files. |
| `query uncovered` | [queries.md](queries.md) | List requirements with no design file references. |
| `query orphan-designs` | [queries.md](queries.md) | List CRC cards missing or with empty Requirements field. |
| `query artifacts` | [queries.md](queries.md) | List all artifacts from design.md with checkbox states. |
| `query gaps [RANGE...] [--open] [--closed]` | [queries.md](queries.md) | Gap items from `design.md`'s Gaps section, read through `minispecsdom`'s gaps reader; RANGE selects by ID in the inline-ref grammar (`O22`, `O22-O28`, `O22,O25`), one type per range, nothing matched refused, partly unassigned fine; `--open`/`--closed` select by checkbox and never claim a permanent gap; both flags mean every checkbox; a valid empty selection says so. |
| `query traceability [file]` | [queries.md](queries.md) | Check traceability comments in one code file. |
| `query traceability --all` | [queries.md](queries.md) | Scan all artifact code files and report traceability status. |
| `query migrations` | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | List in-flight migration specs under `specs/migrations/` (non-recursive). |
| `query comment-patterns` | [config.md](config.md) | Show recognized comment patterns and closers per file extension. |
| `query unindexed-specs` | [queries.md](queries.md) | List per-feature specs not referenced in the root index `specs/index.md` (exact `.md`-token match; lists all when index absent). |
| `query alarms [--unverified] [--brief]` | [queries.md](queries.md) | Census of fault injections recorded in `design/test-*.md`, one line per alarm, with its state: `verified`, `stale`, `unrecorded`, `unanchored`, `unresolvable` — or `unchecked` where git cannot answer. `--unverified` lists only the states that carry a decision while the closing count still reports the whole population. `--brief` replaces each line with the spawn prompt for a delegated re-pull: sites, the test files the Artifacts manifest maps the document to, the `**Fire alarm:**` prose verbatim, and the evidence-never-a-verdict contract. The two compose. |
| `query carves [--open]` | [queries.md](queries.md) | Cross-document status view over `carves/` and `.carves/` at the repository root: open, landed, stateless and unread counts per carve from its status block only, subparts included, fenced samples excluded; `--open` lists the open parts, clean stateless lines and unread lines; non-conforming lines are listed always. Needs no design root. |
| `query links [--all] [file...]` | [queries.md](queries.md), [links-schema.md](links-schema.md) | Every markdown link in the named documents (default: the live carves), resolved relative to the citing file and classified against git: `tracked`, `untracked` (warning), `ignored` (error), `missing` (error), `outside` (error), `external`, `local`. Links inside code groups are never read. Lines carrying a decision are listed, `--all` lists every link, the closing count states every class; exit 1 on any error; refuses outside a git tree. Needs no design root. |
| `query next-id <item\|gap\|req>` | [queries.md](queries.md) | Next free identifier for a class. `item` counts the pending **and** done files at the repository root; `gap` reports every gap type; `req` includes retired requirements. Missing files are reported, never defaulted to `1`. |

## `pending` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `pending add-item --from <doc>#<part>\|<gap-id> "<title>" --status <text> [--skill <s>] [--next-action <text>] [--next\|--nth N\|--after N\|--last]` | [queue-items.md](queue-items.md) | Mint the next item ID and write **both** sides of the link in one act: the whole queue entry — heading, status sentence, part pointer or gap ID, optional `Next:` line — at the stated position, and the part line's `**OPEN (#N.)**` marker (nothing on a gap's side). Every prose slot has a `--<slot>-file` twin read byte for byte, refused alongside the inline form. The `written` line names each file's git kind — tracked and uncommitted, ignored, untracked — so the carve flip reads as the write that still needs a commit. `--create` writes the missing trajectory files with their preambles first; without it a missing file is a refusal that names it (R332, R333). |
| `pending start <N> [--context <text>\|--context-file <path>]` | [queue-items.md](queue-items.md) | **Open an item**: write the current file's `## Active` section — the identity line `` `#N` — <title> `` from the queue entry, the caller's context beneath it byte for byte. Refuses a section that already holds an item. The `written` line names each file's git kind — tracked and uncommitted, ignored, untracked — so the carve flip reads as the write that still needs a commit. |
| `pending finish <N> --commit <hash> [--body <text>\|--body-file <path>] [--discharged <text>\|--discharged-file <path>] (--resolve\|--no-resolve)` | [queue-items.md](queue-items.md) | Check off each discharged part in its carve — box, strikethrough, appended `LANDED` record — then reset `## Active`, then move the entry to the done file with its header (`#N / <discharged>`) and the body when given. A gap-sourced item must say `--resolve` or `--no-resolve`. The `written` line names each file's git kind — tracked and uncommitted, ignored, untracked — so the carve flip reads as the write that still needs a commit. |
| `pending revert` | [backup.md](backup.md) | Undo the most recent trajectory change, and mark the part `**REVERTED (#N.)**`. |
| `pending replay` | [backup.md](backup.md) | Redo what revert undid, returning the part to `**OPEN (#N.)**`. |
| `pending changes` | [backup.md](backup.md) | What has moved in the working tree since the anchor the last transition wrote — `M` changed, `A` new, `D` deleted and recoverable — from one tree-vs-tree diff; names its question and the two repair commands; alters nothing; refuses with no anchor. |

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
| `update retire R<old> <R<new>\|-> "<reason>"` | [updates.md](updates.md), [migrations/complete/001](migrations/complete/001-migration-and-retirement.md) | Retire a requirement: strikethrough the Rn line, append a Tn gap, and print a supersede-at-source reconcile reminder (naming the retired requirement's **Source:** spec) to stderr. Reports the minted value as a sentence on stdout, hidden by `--quiet`, carried by `--json` (R331). |
| `update add-req --section <heading> (--req <text>... \| --req-file <path>...)` | [updates.md](updates.md) | Mints the next free `Rn` for each text and appends it to the named section of `requirements.md` in one act, before the section's first sub-heading; unknown or ambiguous heading refused; a body writing its own `**Rn:**` refused. Reports the minted value as a sentence on stdout, hidden by `--quiet`, carried by `--json` (R331). |
| `update number-alarms [file...]` | [updates.md](updates.md) | Assigns `**Alarm:**` numbers to every alarm lacking one, all of `design/test-*.md` by default. Append-only, never renumbering, idempotent; writes nothing but the added lines and reports per file. |
| `update pulled <doc>#<n> --body-file <f>` | [updates.md](updates.md) | Records a fire alarm as pulled: today's date from the system clock, then the body from a file byte for byte; the previous line folds after it as history. |
| `update repair-links [file...]` | [updates.md](updates.md), [links-schema.md](links-schema.md) | Rewrites links a carve's move broke, both directions: a `missing` link that exactly one sibling relocation between `carves/` and `carves/done/` resolves gets its destination bytes rewritten through the `Markdown` reader; none or two resolving is reported and left. Default population: live carves plus `carves/done/`. Reports every link considered with counts; exit 1 when any was left. Needs no design root. |
| `update inject <doc>#<n> <file:symbol>...` | [updates.md](updates.md) | Re-sites a fire alarm and voids its `**Pulled:**` when the sites resolve to different code — old sites in HEAD, new on disk — demoting the record to history rather than deleting it. |
| `update migration-complete <name>` | [migrations/complete/001-migration-and-retirement.md](migrations/complete/001-migration-and-retirement.md) | Move `specs/migrations/<name>.md` to `specs/migrations/complete/<NNN>-<name>.md`. Reports the minted value as a sentence on stdout, hidden by `--quiet`, carried by `--json` (R331). |

## `phase` subcommands

| Subcommand | Owning spec | Summary |
|---|---|---|
| `phase spec` | [phase.md](phase.md) | Validate Spec Phase: at least one non-empty spec file exists. |
| `phase requirements` | [phase.md](phase.md) | Validate Requirements Phase: requirements.md parseable, Rn unique, sources resolve. |
| `phase design` | [phase.md](phase.md) | Validate Design Phase: design.md Artifacts complete, CRC Rn references valid. |
| `phase implementation` | [phase.md](phase.md) | Validate Implementation Phase: artifact code files exist with traceability comments. |
| `phase gaps` | [phase.md](phase.md) | Validate Gaps Phase: gap IDs follow S/R/D/C/I/O/A/T format, no duplicates. |

## `init` flags

| Flag | Behavior |
|---|---|
| `--track-none` | Record `track: none` — the project is not under version control. |
| `--track-private-trajectory` | Record `track: private-trajectory` — git, trajectory files ignored. |
| `--track-all` | Record `track: all` — git, trajectory files tracked. |
| `--repair` | Operate on an existing configuration instead of creating one; sets `track` and reconciles `.gitignore`. |

Exactly one `--track-*` flag is required. `--repair` inverts the `.minispec/` precondition rather than adding a mode.

## Global flags

| Flag | Behavior |
|---|---|
| `--design-dir <path>` | Override design directory location. |
| `--src-dir <path>` | Override source directory location. |
| `--quiet` | Minimal output. |
| `--json` | Emit JSON output where supported (most `query` subcommands). |
| `--version`, `-v` | Print `minispec <version>` and exit. |
| `-h`, `--help` | Show top-level help. |
