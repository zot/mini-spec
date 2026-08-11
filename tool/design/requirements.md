# Requirements

## Feature: Overview
**Source:** specs/overview.md

- **R1:** Tool is a Go CLI producing a single binary
- **R2:** Tool queries structured parts of design files without AI
- **R3:** Tool performs structural gap detection (missing references, not intent)
- **R4:** Tool performs atomic updates to design files
- **R5:** Tool parses requirements.md format (Rn items with Source specs)
- **R6:** Tool parses CRC card format (Requirements field)
- **R7:** Tool parses design.md Artifacts section (nested checkboxes)
- **R8:** Tool parses design.md Gaps section (typed IDs: S/R/D/C/O)
- **R9:** Tool parses code traceability comments (// CRC: pattern)

## Feature: Queries
**Source:** specs/queries.md

- **R10:** `query requirements` lists all Rn with text and source
- **R11:** `query coverage` shows which design files reference each Rn
- **R12:** `query uncovered` lists Rn with no design file references
- **R13:** `query orphan-designs` lists CRC cards missing Requirements field
- **R14:** `query artifacts` lists artifacts with checkbox states
- **R15:** `query gaps` lists gap items from design.md
- **R16:** `query traceability [file]` checks a code file for CRC/Seq comments
- **R17:** `query traceability --all` scans all code files in Artifacts
- **R89:** `query project` reports resolved project paths (design root, design, src, specs) so the AI can verify which project the tool detected when results look surprising

## Feature: Updates
**Source:** specs/updates.md

- **R18:** `update check [file] [item]` checks a checkbox
- **R19:** `update uncheck [file] [item]` unchecks a checkbox
- **R20:** `update add-ref [crc-file] [Rn]` adds requirement to CRC card
- **R21:** `update remove-ref [crc-file] [Rn]` removes requirement from CRC card
- **R22:** `update add-gap [type] [desc]` adds new gap with auto-numbered ID
- **R23:** `update resolve-gap [id]` marks gap as resolved (checks checkbox)
- **R103:** after `update retire` rewrites the Rold line and appends the Tn (R80), it prints a supersede-at-source reconcile reminder to **stderr** — naming Rold's feature `**Source:**` spec(s) from requirements.md (or saying none is recorded), prompting a grep of design/ for stale directives, and stating the completion test — while **stdout** carries only the Tn; the reminder is advisory and is suppressed by `--quiet`

## Feature: Validate
**Source:** specs/validate.md

- **R24:** `validate` runs all structural validations
- **R25:** Validates requirements.md format: unique Rn numbering with no duplicates or gaps (file order doesn't matter)
- **R26:** Validates CRC cards have Requirements field with valid Rn refs
- **R27:** Validates Artifacts section structure and file existence
- **R28:** Validates Gaps section structure and ID format
- **R29:** Validates code files have traceability comments
- **R30:** Validation output shows what was found (not just pass/fail)
- **R31:** Exit code 0 if no issues, 1 if issues found

## Feature: Configuration
**Source:** specs/config.md

- **R32:** Design-root detection walks up to find the design/ directory
- **R33:** Default paths: design/, src/, crc-*.md, seq-*.md
- **R34:** Optional `.minispec.yaml` config file in a **design root**, for overrides
- **R35:** CLI flags: --design-dir, --src-dir, --quiet, --json
- **R36:** JSON output mode for tooling integration
- **R37:** (deferred) MCP server mode via `minispec serve`
- **R38:** Configurable comment patterns per file extension (map in config)
- **R39:** Default comment patterns for go, js/ts, python, lua, c/h, shell

## Feature: Extended Validation
**Source:** specs/validate.md

- **R40:** Validates all design files (crc-*, seq-*, ui-*, test-*, manifest-*) in design/ are listed in Artifacts
- **R41:** Validates Source fields in requirements.md reference existing spec files
- **R42:** Validates CRC/Seq refs in code traceability comments reference existing design files
- **R43:** Validates files listed in CRC Sequences sections exist

## Feature: Phase Subcommands
**Source:** specs/phase.md

- **R44:** `phase spec` validates spec files exist and are non-empty
- **R45:** `phase requirements` validates requirements.md format and spec sources
- **R46:** `phase design` validates design files, CRC cards, and requirement coverage
- **R47:** `phase implementation` validates code files exist and have traceability comments
- **R48:** `phase gaps` validates gaps section structure
- **R49:** Phase commands show focused output relevant to that phase only
- **R50:** Phase commands exit 0 if phase passes, 1 if issues found

## Feature: New Artifacts Format
**Source:** specs/config.md

- **R51:** Tool parses inline artifact format: `- [x] design.md → code.ts`
- **R52:** Tool parses comma-separated code files after `→` arrow
- **R53:** Tool strips optional backticks from code file paths

## Feature: Version Reporting
**Source:** specs/config.md

- **R54:** `--version` flag displays version and exits
- **R55:** Version is shown in help output header
- **~~R56:~~** (Retired T1 — see R117) `check-version` compares tool version against skill README.md Version: line (project-level then user-level), exits 0 on match, 1 on mismatch or not found

## Feature: Comment Closers
**Source:** specs/config.md

- **R57:** Configurable comment closers per file extension (map in config), merged over defaults
- **R58:** Default comment closers for block-comment languages (.md, .html, .css)
- **R59:** Parser strips config-provided comment closer from traceability refs
- **R60:** `query comment-patterns` shows closers section with WARNING when closers are configured

## Feature: Approved Gaps
**Source:** specs/updates.md

- **R61:** Tool parses gap type "A" (approved) in addition to S/R/D/C/O
- **R62:** `update approve-gap [id]` converts existing gap to approved type with new A-number, preserving description
- **R63:** Approved gaps are reported separately from open/resolved in phase and validate output
- **R64:** Approved gaps do not cause validation failure (not counted as open)
- **R65:** Requirements referenced by approved gaps (via Rn or Rn-Rm ranges in description) are treated as covered and excluded from uncovered-requirements issues

## Feature: Traceability Robustness
**Source:** specs/validate.md

- **R66:** Traceability parser stops each pipe-delimited section (CRC, Seq) at the next `|` delimiter

## Feature: Inline Requirement Traceability
**Source:** specs/validate.md

- **R67:** Traceability parser extracts Rn refs from an optional third pipe-delimited section (e.g. `| R5, R12`) into a ReqRefs field
- **R68:** Validate checks that inline Rn refs in code files exist in requirements.md
- **R69:** Validate checks implementation coverage: every requirement should appear as an inline Rn ref in at least one code file (requirements covered by approved gaps are excluded)
- **R70:** Requirements with design coverage but no implementation coverage are reported as I-type (implementation) gaps
- **R71:** Tool parses gap type "I" (implementation) in addition to S/R/D/C/O/A
- **R72:** Validate output includes an implementation coverage section showing which requirements have code-level refs
- **R104:** Traceability parser also extracts inline Rn refs from a bare annotation — a comment whose first token after the comment leader is a requirement ref (`// R5: desc`, `// R5, R6`, trailing `foo() // R7`) — collecting the leading comma-separated `Rn` refs into ReqRefs. A ref that does not immediately follow the comment leader (a prose mention like `// see R5` or `// computed (R5)`) is not counted.
- **R105:** Inline Rn ref harvesting (both the `// CRC: … | Rn` tail and the bare annotation) expands `Rn-Rm` range syntax into every member, so a range-form annotation (`// R5-R8`) covers the whole span; the second `R` is optional (`R5-8`), comma-separated lists and ranges may be mixed (`// R5-R7, R10`), and a reversed range (`R8-R5`) contributes only the low ref.
- **R106:** The configured comment prefix is wrapped in a non-capturing group when composed into the traceability and bare-annotation regexes, so an alternation pattern (e.g. `<!--\s*|//\s*` for HTML with embedded JS) composes correctly instead of the `|` binding loosely and matching the first alternative without the rest of the comment.

## Feature: Migration Workflow
**Source:** specs/migrations/complete/001-migration-and-retirement.md

- **R73:** Tool parses Gaps section type "T" (retired) in addition to S/R/D/C/I/O/A
- **R74:** Tn entries are written and parsed without a leading checkbox (`- T1: ...`, no `[ ]` or `[x]`)
- **R75:** An (approved) entries are written without a leading checkbox; the parser still accepts the legacy `- [ ] An: ...` form for back-compat
- **R76:** Validate reports A-typed and T-typed gap lines that carry a checkbox marker as a "permanent gaps with checkbox" issue so AIs clean them up
- **R77:** Requirements parser accepts the strikethrough retired form `- **~~Rn:~~** (Retired Tk — see Rxxx) <text>` (or `... no replacement) <text>`) and exposes it via a `Retired` flag on the parsed Requirement
- **R78:** Retired requirements are excluded from coverage uncovered and implementation-coverage uncovered lists; their Rn IDs remain valid for cross-reference resolution
- **R79:** `query migrations` lists in-flight migration spec files (specs/migrations/*.md, non-recursive); empty output and exit 0 when none exist
- **R102:** `query unindexed-specs` lists per-feature specs (specs/*.md, non-recursive, excluding index.md) not referenced in the root index specs/index.md, matched by exact `.md` token; lists every spec when index.md is absent; empty output and exit 0 when all specs are indexed
- **R80:** `update retire R<old> <R<new>|-> "<reason>"` rewrites the R<old> line in requirements.md with the strikethrough/Retired marker AND appends a new Tn entry to the Gaps section of design.md, atomically; outputs the assigned Tn
- **R81:** `update migration-complete <name>` moves specs/migrations/<name>.md to specs/migrations/complete/<NNN>-<name>.md where NNN is the next zero-padded three-digit number, and outputs the new path
- **R82:** `update add-gap T <reason>` adds a checkbox-less T-typed gap line with auto-numbered ID
- **R83:** `update add-gap A <desc>` and `update approve-gap <id>` produce checkbox-less An lines

## Feature: Concise Validate Output
**Source:** specs/migrations/complete/001-migration-and-retirement.md

- **R84:** Validate output prints only the issues categories that contain at least one issue; on success, output is a single `phase: validate OK` line
- **R85:** Validate output groups requirement IDs into ranges (e.g. `R44-72, R83-85`) wherever a list of Rn appears (uncovered, missing impl coverage, duplicates, etc.)
- **R86:** Validate output deduplicates identical issue messages (a code file with multiple matches against a missing design ref reports the broken ref once)
- **R87:** Phase subcommand output uses the same ranging and dedup rules as validate everywhere Rn lists appear, including findings sections (`found:`, per-source listings, covered/uncovered, etc.). Successful phase output stays brief (one summary line plus any sparse findings) and skips full-list enumerations
- **R88:** Validate and phase output category labels are stable, lowercase, machine-greppable strings (e.g. `uncovered requirements:`, `missing impl coverage:`, `permanent gaps with checkbox:`)

## Feature: Source Line Diagnostics
**Source:** specs/validate.md

- **R90:** Requirements parser accepts a comma-separated list of paths on a `**Source:**` line and exposes them via a `Sources []string` field; validate iterates each path independently
- **R91:** Validate detects malformed Source values (entries that don't match a clean relative `.md` path) and near-miss Source-like lines that don't match the canonical pattern, reporting them in `malformed Source values:` and `suspicious Source lines:` categories
- **R92:** Validate output appends a `fix instructions:` block at the bottom describing the canonical `**Source:**` format whenever malformed values or suspicious lines are detected (crank-handle pattern)
- **R93:** Spec-source resolver recognizes the migration-completion convention: a Source value `specs/migrations/X.md` resolves to `specs/migrations/complete/<NNN>-X.md` if the literal path does not exist but a digit-prefixed match does, so requirements pointing at migrated specs do not flag as missing after `update migration-complete`


## Feature: Sequence Anchors
**Source:** specs/overview.md, specs/validate.md

- **R94:** Traceability parser splits each `Seq:` reference into a filename and an optional `#K.x.y` fragment; file-only references remain valid
- **R95:** Tool parses numbered items in sequence-diagram files — lines whose first non-whitespace, non-lane content is a dotted-number token (`1.`, `1.1`, `1.1.1.`), allowing the lane characters `│ ├ └ ─ |` and ASCII tree connectors before the number
- **R96:** Sequence-doc parser groups dotted IDs by their first segment K (diagram index) and exposes a `Has(id)` lookup, the set of K values present, and the tree under each K
- **R97:** Validate resolves each `Seq:` fragment against the parsed sequence file and reports unresolved fragments per code file
- **R98:** Validate checks per-K tree contiguity in every numbered sequence file: at each tree level under K, children must be 1..N with no holes
- **R99:** Validate checks K-sequence contiguity within each numbered sequence file: the set of K values present must be contiguous starting at 1
- **R100:** Validate checks intra-file dotted-ID uniqueness: every dotted ID appears at most once per sequence file
- **R101:** Unnumbered sequence files (no dotted items found) are silently skipped by sequence-numbering validation — numbering is opt-in per file

## Feature: Repository Root
**Source:** specs/repository-root.md

- **R107:** Tool distinguishes two roots: the **design root**, the directory containing `design/`, which owns `specs/`, `design/` and `src/`; and the **repository root**, the top of the version-controlled working tree, which owns `.claude/`, `carves/` and the trajectory files
- **R108:** Repository-root detection searches upward from the current directory **collecting** markers rather than accepting the first one met, so a strong marker wins from any depth below it
- **R109:** `.git`, `.minispec/`, `carves/`, and a trajectory file (`PENDING.md`, `CURRENT.md`, `DONE.md`) are equal strong markers: the deepest directory carrying any of them is the repository root, and no ordering among them is defined
- **R110:** When no strong marker is found, the deepest `.claude` directory is the repository root
- **R111:** When neither a strong marker nor a `.claude` directory is found, the deepest `.minispec.yaml` is the repository root — the weak final fallback, meaningful only because a design root is sometimes also the repository root
- **R112:** Repository-root detection never considers the user's home directory or any directory above it
- **R113:** Repository-root detection failure is an error naming the directory the search started from
- **R114:** Only `.git` counts as a version-control marker; a tree managed by another VCS falls through to the next marker rather than being detected
- **R115:** `query project` reports the repository root alongside the design root
- **R116:** `query project` states when the repository root and the design root are the same directory, rather than printing the same path twice unlabeled
- **R117:** `check-version` finds the skill's `README.md` under the repository root first, then under the user's home directory

## Feature: Config Scopes
**Source:** specs/config.md

- **R118:** Repository configuration lives at `<repository root>/.minispec/config.yaml`
- **R119:** `.minispec/` is a tool-managed directory at the repository root, holding the repository configuration and the tool's machine-local working files
- **R120:** Where the repository root is also a design root, `.minispec/config.yaml` is that design root's configuration too — there is no second file
- **R121:** A design root's `.minispec.yaml` states only what differs from the repository configuration; a design root whose settings match it needs no file at all
- **R122:** A `.minispec.yaml` at the repository root is an error with no exception, because it would have to inherit from `.minispec/config.yaml` — a file inside its own directory
- **R123:** The repository-root `.minispec.yaml` check is a single existence test that detects a symlink as readily as a regular file, so a link to the new location is rejected rather than treated as a supported alias
- **R124:** Settings resolve in three layers, each applied over the previous: built-in defaults, then the repository configuration, then the design root's own file
- **R125:** Scalar settings (`design_dir`, `src_dir`) replace the inherited value when present
- **R126:** Map settings (`comment_patterns`, `comment_closers`) merge per key, so a design root adding one entry keeps every other entry the repository set
- **R127:** List settings (`code_extensions`) merge as a union **between configuration layers**: repository entries are kept, design-root additions appended, duplicates dropped, and repository order preserved so the result is deterministic
- **R130:** The first configuration layer to set a list **replaces** the built-in defaults rather than unioning onto them, so a project can still narrow a shipped list. The remedy in R128 works by moving a setting down a level, and nothing sits below the defaults to move it to; defaults are also not a layer anyone authored, so "state only what you add" cannot apply to them
- **R128:** (inferred) A design root can add to an inherited list but cannot remove from one; removal is achieved by dropping the setting from the repository configuration and stating it in each design root that needs it, so no removal syntax exists
- **R129:** The tool can report which file each effective setting came from, so a value's origin does not require reading two files and knowing the precedence

## Feature: Initialization
**Source:** specs/initialization.md

- **R131:** `track` is a repository-scoped setting in `.minispec/config.yaml` with exactly three values: `none`, `private-trajectory`, `all`
- **R132:** `track: none` asserts the project is not version-controlled; `private-trajectory` and `all` both assert it is git-managed, differing only in whether the trajectory files are ignored
- **R133:** The paths whose ignore state `track` governs are `PENDING.md`, `CURRENT.md` and `DONE.md` at the repository root, plus `.carves/` whenever it exists
- **R134:** `.carves/` must be ignored under every git-managed `track` value, including `all`; public `carves/` is never required to be ignored
- **R135:** A design root's `.minispec.yaml` cannot set or override `track`, because it describes the repository rather than one design root
- **R136:** `minispec init` requires exactly one `--track-<style>` flag and never infers or defaults the value
- **R137:** `minispec init` is the sole creator of `.minispec/config.yaml`; no other command brings it into existence
- **R138:** `init` writes `<repository root>/.minispec/config.yaml` recording the chosen `track` value
- **R139:** `init` adds an ignore line for `.minispec/backup` to the top-level `.gitignore` under any git-managed `track` value
- **R140:** `init` adds ignore lines for the trajectory files under `track: private-trajectory`
- **R141:** `init` reports every file it created or edited, in full, so an agent never has to infer what changed
- **R142:** Plain `init` refuses when `.minispec/` already exists, naming `--repair` as the way to change a `track` value and saying to confirm with the user first
- **R143:** The tool never runs `init` implicitly
- **R144:** `init --track-<style> --repair` sets `track` on an existing configuration *and* brings `.gitignore` into agreement with that value
- **R145:** `--repair` requires `.minispec/` to exist — the inverse of plain `init`'s precondition — so neither form has to guess the caller's intent
- **R146:** The `--repair` instruction tells the agent to confirm the `track` value with the user before running it, because the two repair directions mean opposite things
- **R147:** Every command doing more than reporting its version verifies `track` against whether the repository is git-managed
- **R148:** The same check verifies `track` against the actual ignore state of the paths in R133
- **R149:** A `track` mismatch gripes and exits, naming which fact disagrees and `minispec init --track-<style> --repair` as the repair
- **R150:** The mismatch gripe fires on every run until repaired, rather than once per session
- **R151:** In a tree with no git at all, the `track` check and the git preferences are silent
- **R152:** With no `.minispec/config.yaml`, the only commands that run are `init`, `--version`, `help`, and `check-version`
- **R153:** `check-version` runs without a configuration deliberately, so an agent whose first instruction is to run it can establish tool/skill agreement before being told to initialize
- **R154:** Every other command refuses with a crank handle when no configuration exists
- **R155:** The no-configuration refusal names the absolute path it would make the repository root, so the user's assent lands on a stated location
- **R156:** The refusal instructs the agent to establish whether the directory is a code project — a fact the agent can check — rather than asking the user that question
- **R157:** The refusal instructs the agent to report and wait, never to run `init` on its own conclusion
- **R158:** The refusal asks the agent to recognise a directory of project directories and report that finding, since that is the case that would otherwise plant a repository root in the wrong place
- **R159:** The refusal supplies the gist for the agent to compose from, not verbatim user-facing copy, so the message can adapt to what the agent found
- **R160:** When the user declines, the refusal says to run from inside the project
- **R161:** `--repair` validates the configuration's well-formedness before acting on it
- **R162:** A malformed configuration is the one case where the tool explicitly authorises the agent to edit the configuration directly, stating every problem found and pointing at the skill's configuration documentation
- **R163:** The agent backs up the configuration before editing it, skipping the backup when it would be byte-identical to one already present
- **R164:** The tool gripes while `.minispec/config.yaml` is not tracked by git
- **R165:** The tool gripes while `.minispec/backup` is not ignored by git
- **R166:** The tool does not stage, commit, reset, or otherwise alter git state; editing `.gitignore` is not excluded by this, being a file edit rather than a change to git's state
- **R167:** Version-control checks shell out to the `git` command line only — no second VCS is supported and no VCS library is linked in
- **R168:** A project managed by a version-control system other than git is told its ignore state cannot be checked, rather than passing silently
- **R169:** (inferred) A project that predates `track` must run `init` once before non-version commands work; the one-time refusal is how it is asked the question
- **R170:** The tool adds no ignore line for a path git already ignores, whatever rule is doing it. Whether a path is covered is asked of git rather than decided by matching lines, because an anchored `/PENDING.md`, a bare `PENDING.md` and a `*.md` wildcard are the same intent written three ways
- **R171:** Ignore lines the tool writes are anchored to the repository root (`/PENDING.md`), because every path `track` governs is mandated there and an unanchored pattern also matches nested ones — and, since the last matching pattern wins, would silently widen a rule the project had written narrowly
- **R172:** When a `track` value requires a path to be tracked but a rule outside the top-level `.gitignore` still ignores it, the tool names that path rather than reporting a success it did not achieve
