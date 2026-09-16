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
- **R317:** `minispec query gaps` accepts **RANGE** arguments selecting gaps by ID, in the grammar inline requirement refs already use: a bare ID (`O22`), an inclusive range with the second letter optional (`O22-O28`, `O22-28`), a comma list (`O22,O25`), and mixtures; a reversed range contributes only its low end
- **R318:** A range whose endpoints name **different gap types** is refused, naming both: the letter is the namespace, so `O22-R5` spans nothing, and the gap `R1` is never the requirement `R1`
- **R319:** A range selecting **no gap at all** is refused, naming what it looked for — empty output with a success exit would read as *there are no gaps*, and a report of absence is indistinguishable from a report of nothing wrong
- **R320:** A range whose members are **partly unassigned** selects those that exist and is not an error: gaps in an ID sequence are expected by the ID rule, so `O22-O28` with no `O25` ever assigned is an ordinary request, answered in document order
- **R321:** `--open` selects gaps whose checkbox is unchecked and `--closed` those whose checkbox is checked; a **permanent gap (`A`, `T`) carries no checkbox by rule and is selected by neither**, since it records an approval or a retirement rather than work, and stays reachable by naming it in a range
- **R322:** Passing **both** `--open` and `--closed` selects every gap that has a checkbox — the natural reading of the pair, and the only way to say *work items only* without naming them
- **R323:** `--open` and `--closed` are **subcommand flags** parsed wherever they sit among the positional arguments, never global ones, and the selection is applied **before the output form is chosen** so `--json` renders exactly the selected set; a valid selection that matches nothing prints what was asked and that nothing came back (`no open gaps in O22-O28`), with exit 0

## Feature: Updates
**Source:** specs/updates.md

- **R18:** `update check [file] [item]` checks a checkbox
- **R19:** `update uncheck [file] [item]` unchecks a checkbox
- **R20:** `update add-ref [crc-file] [Rn]` adds requirement to CRC card
- **R21:** `update remove-ref [crc-file] [Rn]` removes requirement from CRC card
- **R22:** `update add-gap [type] [desc]` adds new gap with auto-numbered ID
- **R23:** `update resolve-gap [id]` marks gap as resolved (checks checkbox)
- **~~R103:~~** (Retired T4 — see R331) after `update retire` rewrites the Rold line and appends the Tn (R80), it prints a supersede-at-source reconcile reminder to **stderr** — naming Rold's feature `**Source:**` spec(s) from requirements.md (or saying none is recorded), prompting a grep of design/ for stale directives, and stating the completion test — while **stdout** carries only the Tn; the reminder is advisory and is suppressed by `--quiet`
- **R313:** `update number-alarms [file...]` **assigns and writes** `**Alarm:**` numbers to every alarm lacking one, all of `design/test-*.md` by default: **append-only** — an existing number is kept whatever order it now sits in and a freed one is never handed out again — **idempotent**, and it changes nothing but the added field lines; it reports what it assigned per file and refuses a document whose entry carries a deviation before any byte moves
- **R314:** `update pulled <doc>#<n> --body-file <f>` writes an alarm's `**Pulled:**` line as **today's date from the system clock**, then the body read byte for byte from a file; a previous line is folded after it as ` *Earlier —* …` so the leading date moves and the history is kept; a first pull is inserted directly after the `**Inject:**` line. No commit is written after the date (Bill, 2026-09-04)
- **R315:** `update inject <doc>#<n> <file:symbol>...` rewrites an alarm's `**Inject:**` line and **voids its `**Pulled:**` when the sites resolve to different code** — the old sites in HEAD, where a renamed symbol still exists, and the new sites on disk — never by comparing the field's text: a rename, a re-formatting or a disambiguation to the declaration already watched keeps the record, a move demotes it to a history sentence naming the old sites. A site resolving to nothing counts as different; rewriting the sites to what they already say changes nothing; an empty site list is refused
- **R324:** `minispec update add-req --section <heading> --req <text>...` (or `--req-file <path>...`, never mixed, since the interleaved order is what assigns the numbers) mints the next free `Rn` for each text — counting retired numbers — and appends `- **Rn:** <text>` to the addressed section **in one invocation**, so no command hands out a number without recording it; a batch is appended in the order given and reported as a range. The section is addressed by its heading's literal text at whatever level it lives, with or without a leading `Feature: `; entries land at the end of the section's **own** content, before its first sub-heading; an unknown heading is refused (the tool owns IDs, not prose) and so is a title several headings carry
- **R325:** `update add-req` **refuses a body that opens with its own `**Rn:**` marker** and names the reason, rather than stripping it: the verb mints both the identifier and the label, so a caller writing one is duplicating rather than choosing, and stripping would have to guess whether the written number matches the one about to be assigned
- **R326:** `design.md`'s Gaps section and `requirements.md` are read **through `minispecsdom`'s gaps and requirements readers** — a gap is a keyed bullet at any depth with its text folded, a requirement is a keyed column-0 bullet in a section at any level whose `**Source:**` is inherited from the nearest ancestor that has one, a fenced example is body, and a deviation is listed and refused — and `add-gap`, `resolve-gap`, `approve-gap`, `retire` and `add-req` write through those readers' node-addressed writes, which decide every refusal before a byte moves and read themselves back; minting stays this tool's
- **R331:** An `update` verb that **mints** a value — `retire`'s `Tn`, `migration-complete`'s path, `add-req`'s `Rn` or range — reports it the way every other `update` verb reports: a prose sentence naming the value on stdout, suppressed by `--quiet`, with `--json` carrying the value for a caller that reads it back; a minted value is a report, not a return value, and `--quiet` hides it like any other sentence because `--json` is the channel for a caller that needs it

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
- **R327:** **The two readers of every design document `minispecsdom` owns must agree, and this check is the second opinion**: for the Gaps section, requirements.md and every test design, an independent line scan — a regex over lines, bounded for the gaps section by its heading and the next level-2 heading, built on nothing the reader is built on — is compared with what the reader returned, IDs for gaps and requirements and entry counts for test designs, and every difference is a finding listed first, because every check below it reads through the document reader alone. A reader that lost a file's tail to one unclosed span agrees with itself forever; only a scan that shares none of its blind spots can say how much of the file it actually saw
- **R328:** `validate` prints what the design-document readers could not read — an entry-like line outside the shape, a group never closed — as a **coverage note** naming each file and its count, whether or not anything else fired, because a reader takes silence about coverage as a claim of completeness; it is never an issue in itself

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

- **R40:** Validates all design files (`crc-*`, `seq-*`, `ui-*`, `test-*`, `manifest-*`) in design/ are listed in Artifacts
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
- **R79:** `query migrations` lists in-flight migration spec files (`specs/migrations/*.md`, non-recursive); empty output and exit 0 when none exist
- **R102:** `query unindexed-specs` lists per-feature specs (`specs/*.md`, non-recursive, excluding index.md) not referenced in the root index specs/index.md, matched by exact `.md` token; lists every spec when index.md is absent; empty output and exit 0 when all specs are indexed
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
- **R188:** Validate appends a `fix instructions:` block whenever a Source path is missing, naming the three legitimate repairs — rewrite the Source (renamed), repoint it at the absorbing spec (merged), or retire the requirements and regroup them under `specs/deleted.md` (deleted) — and stating that requirement numbers are never renumbered or reused, because deleting the orphaned requirements opens a numbering gap whose renumbering repair silently repoints every design and code anchor
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
- **R169:** (inferred) A project that predates `track` must run `init` once before non-version commands work — plain `init` where no configuration exists, `--repair` where one exists without a `track` value; in either case a one-time refusal is how it is asked the question
- **R170:** The tool adds no ignore line for a path git already ignores, whatever rule is doing it. Whether a path is covered is asked of git rather than decided by matching lines, because an anchored `/PENDING.md`, a bare `PENDING.md` and a `*.md` wildcard are the same intent written three ways
- **R171:** Ignore lines the tool writes are anchored to the repository root (`/PENDING.md`), because every path `track` governs is mandated there and an unanchored pattern also matches nested ones — and, since the last matching pattern wins, would silently widen a rule the project had written narrowly
- **R172:** When a `track` value requires a path to be tracked but a rule outside the top-level `.gitignore` still ignores it, the tool names that path rather than reporting a success it did not achieve
- **R173:** A repository configuration that parses but sets no `track` predates the setting rather than being malformed, and `--repair` is the verb that fixes it: absence is a version difference, a value outside the closed set is damage
- **R174:** The startup check refuses a pre-`track` configuration with its own crank handle naming `minispec init --track-<style> --repair`, distinct from the malformed-configuration refusal, which would otherwise send the agent to hand-edit a file a flag can repair — and hand-editing sets `track` without reconciling `.gitignore`
- **R175:** The pre-`track` refusal asks the agent only for the user's intent — whether the work queue stays private or ships with the repository — because the configuration's existence already settles that this is a mini-spec project and where its repository root is
- **R176:** The pre-`track` refusal instructs the agent to report and wait, never to choose a `track` value on the user's behalf
- **R177:** Writing `track` into an existing configuration preserves its comments, the order of its keys, and every setting the running binary does not model, rather than re-serialising the file from the settings it knows about
- **R334:** `minispec init carve <name>` scaffolds `carves/<name>.md` in the shape the format mandates — the title, a `## Status` block holding one open unqueued part, a `## Decisions` section, and the part's elaboration stub — creating `carves/` when absent, refusing an existing file and a name that is a path or carries an extension, and reporting the write with its git kind and the `add-item` that queues its first part

## Feature: Fire Alarm Freshness
**Source:** specs/validate.md, specs/queries.md

- **R178:** A `test-*.md` test entry may carry `**Fire alarm:**` (the injection in prose), `**Inject:**` (the `file:symbol` sites the injection edits, comma-separated), `**Pulled:**` (the date it was run and what happened), and `**Code:**` (the test file)
- **R179:** `validate` reports an alarm as **stale** when it carries both `**Inject:**` and `**Pulled:**` and git shows any named symbol changed on or after the pulled date
- **R180:** The change question is asked of the **function**, not the file it lives in — a file-level answer marks every alarm in a busy file stale and so discriminates nothing
- **R181:** A change must be dated **strictly after** the pull to count as stale. Same-day is the normal workflow — fix the code, pull the alarm, commit both together — so counting it stale would mark every freshly-pulled alarm stale on arrival, and a check that always fires is ignored. The cost is a blind spot: a change made later the same day is missed until the next change on any later day
- **R182:** An `**Inject:**` site whose symbol git cannot find is reported as **unresolvable** rather than skipped, since that is the anchor rotting — the failure the field exists to prevent
- **R183:** `validate` reports stale alarms only. Alarms lacking `**Pulled:**` or `**Inject:**` are reported by `query alarms` instead, because a count that stays non-zero for months is a nag rather than a closable gripe
- **R184:** The freshness check is silent in a tree with no git, since a check that could not look must not return a clean result
- **~~R205:~~** (Retired T2 — see R304) An `**Inject:**` symbol may name a method by its receiver — `Type.Method` — and that form resolves to exactly that method's declaration: git is handed a declaration-shaped pattern (`func (<name> *Type) Method`, name and star optional) rather than the anchor as written, which matches no line of Go
- **~~R206:~~** (Retired T3 — see R303) A bare `**Inject:**` symbol is resolved on word boundaries, never as a substring, so `Lookup` does not resolve to `LookupPath`; whether the first bounded match is a declaration rather than a use or a comment is not settled here (see gaps O10, O11)
- **R303:** **The reader computes an `**Inject:**` site's line range and git is asked only when those lines changed** — `git log -L <start>,<end>:<file>` — never handed a pattern to find the symbol with, so the range can land on nothing but the declaration: not a use, not a comment, not the successor's doc block
- **R304:** A `Type.Method` symbol resolves to the method whose **receiver group** names that type — `(d *Doc)` and `(Doc)` both answer `Doc`, a type parameter list is dropped — so same-named methods on different types are different symbols; a bare name resolves to any declaration of it, receiver or none
- **R305:** A declaration's extent runs **from its declaring line through the line on which every bracket group opened inside it has closed**; a member of a grouped `const`, `var` or `type` starts at its own line rather than the keyword's
- **R306:** **No comment is in the range** — not the successor's, which is git's error, and not the declaration's own, which old-sdom measured wrong: three verified alarms went stale over traceability lines rewritten inside their doc blocks. A comment is not the code an injection proves
- **R307:** A symbol **declared more than once** in its file is reported as unresolvable with the count and the repair (name the receiver), never resolved to the first: the anchor has been watching an arbitrary one since it was written, and every answer about it was confident and unfounded
- **R309:** An anchored alarm with no `**Pulled:**` has its sites **resolved before it reads `unrecorded`** — at the cheap half of the cost, the extent without the history walk — so a prescription pointing at a symbol that does not exist, or at two, reads `unresolvable` and names the site; with no git or no history yet it stays `unrecorded`, which is what it was
- **R310:** An alarm is **named `<document>#<n>`**, where `n` comes from an `**Alarm:**` field on its test entry; numbers are local to the file, as sequence-step anchors are, so the path disambiguates and no cross-document uniqueness is implied
- **R311:** An alarm's number is an **identifier, never a position**: assigned once, stored in the document, never reused or renumbered. A new alarm takes the maximum ever assigned in its file plus one, so a gap in a file's sequence is expected rather than closed
- **R312:** An alarm carrying no `**Alarm:**` field is **unnumbered, and the census says so** with its repair (`update number-alarms`), listing it by title, never numbering it by position
- **R316:** A test design is read **through `minispecsdom`'s test-document reader**: an entry is a `## Test:` heading's region, a field is `**Name:**` at a line head outside any code group, the two prose fields fold across wrapped lines, a fenced example is body, and a doubled field is a deviation the reader names and every write refuses; this tool's adapter keeps only the judgments that are its own — a site with no file or symbol is dropped, a `**Pulled:**` whose date does not parse records nothing — and the census prints what the reader could not read as a coverage note beneath its count, never dropping it, since a group never closed takes every later entry with it
- **R308:** The range is computed from the file **as committed at HEAD**, read once per file per invocation, because `-L <start>,<end>` resolves against HEAD and an uncommitted edit higher in the file moves every later declaration; a symbol absent from HEAD's file and present on disk is *no history yet*, not a rotted anchor. Only Go sources have an extent today; a site in any other file is unresolvable (gap O21)
- **R185:** `minispec query alarms` lists every recorded alarm with its state — `verified`, `stale`, `unrecorded`, `unanchored` — and closes with a census of the four counts
- **R186:** `unrecorded` states that the repository does not record a verification, never that the injection was not run — the documents cannot answer the second question
- **R187:** Without git, `query alarms` reports `verified` and `stale` alarms as `unchecked` rather than assuming either
- **R199:** `query alarms --unverified` lists only the alarms whose state carries a decision — everything that is not `verified` — while the closing census still counts the **whole** population, so the filtered form is the full census minus the repetitions of *nothing to do here* and minus nothing else
- **R200:** `query alarms --brief` replaces each selected alarm's line with the spawn prompt for a delegated re-pull: the design root relative to the repository root, the document and test title, the state, the `**Inject:**` sites, the test files `design.md`'s Artifacts manifest maps that document to together with their directories, and the `**Fire alarm:**` prose verbatim
- **R201:** A brief names test **files and directories and never a command**, because `minispec` knows document structure and not build systems, and a guessed command that fails to build produces output a hurried reader scores as *rang*
- **R202:** A brief closes with the evidence contract — return the command, its output before the injection, the diff applied, the output after, and the diff after restoring, and never a verdict on whether the alarm rang — and carries no protocol, which belongs to the agent definition that runs it
- **R203:** A brief is emitted for **every** alarm the filters selected, and where the material for one of its parts is missing — no `**Inject:**` sites, no Artifacts row for the document — it states the absence in place of that part rather than dropping the line, so the brief count never disagrees with the census
- **R204:** `--brief` is an output form and `--unverified` a filter, so the two compose; under `--json` the brief is a field on each assessment rather than a separate output

## Feature: Next Free Identifier
**Source:** specs/queries.md

- **R189:** `minispec query next-id <item|gap|req>` prints the next free identifier for the named class. The class argument is required — it selects both what is counted and which root the files are resolved under
- **R190:** `next-id item` is the maximum item ID across **both** the pending file and the done file. Either file alone yields a number that collides with an existing ID: the pending file's maximum is too low right after items complete, the done file's while the highest IDs are still live. In the done file the IDs are read from an entry **header's leading identifier slot** and from nowhere else — an entry may discharge several, and its body routinely quotes other items, so a body scan raises the maximum from a citation
- **R191:** `next-id item` resolves its files at the repository root; `next-id gap` and `next-id req` resolve theirs at the design root
- **R192:** `next-id gap` reports the next free number for every gap type, because gap numbering runs a separate sequence per type. A type with no gaps reports 1
- **R193:** `next-id req` counts retired requirements, since a retired `Rn` keeps its number permanently and skipping it would hand out a number already taken
- **R194:** `next-id item` reports that it cannot answer when neither trajectory file is present, rather than returning 1 — a project running no trajectory layer has no next ID, and 1 is a confident wrong answer rather than an absent one
- **R195:** `next-id item` answers when exactly one of the two files is present, and names the file it could not read so the reader can judge whether the number is trustworthy
- **R196:** `next-id` writes markdown to stdout and honors the global `--json` flag
- **R197:** `next-id` reports which files it read and how many identifiers each contributed. The number alone is uncheckable: a regex that matches nothing and a genuinely empty queue both answer 1, and the counts are what distinguish a broken parser from a correct answer
- **R198:** `next-id item` answers in a repository with **no design root at all**. The queue is repository-scoped, so requiring a project would make the command unusable exactly where it is meant to run — this repository is the proving case, with design roots at `tool/` and `example/` and the queue above both

## Feature: Carve Status View
**Source:** specs/queries.md

- **R207:** `minispec query carves` prints one line per live carve — its repository-relative path, how many of its parts are open and how many have landed — followed by a census over every carve read, ordered `carves/` before `.carves/` and alphabetically within
- **R208:** Carves are found by **location, not by name**: every `*.md` directly in `<repo root>/carves/` and `<repo root>/.carves/`; `carves/done/` is not read
- **R209:** Counts come from the **status block and nowhere else** in the document; the block is bounded on heading nodes, so a fenced `## Status` or a fenced part line contributes nothing
- **R210:** Subparts, indented one level under their parent, are counted as parts
- **R211:** A document in a carve directory with **no status block is reported as such**, never dropped
- **R212:** A part is listed under its carve when it carries any deviation, or when `--open` is given and it is open; a non-conforming part is **never** behind the flag
- **R213:** `query carves` resolves at the **repository root** and answers in a repository with no design root
- **R214:** When **neither carve directory exists**, the query reports that it cannot answer rather than printing an empty census
- **R215:** `query carves` writes markdown to stdout and honours `--json` wherever it appears among the arguments
- **R216:** A status-block list item with **no checkbox is `stateless`**: counted on its carve's line and in the census always, reported as existence, a line number and a reason and never a state, listed under `--open` — unless it carries a format deviation, in which case it lists always
- **R217:** The per-carve count and the census total for such lines are labelled **`stateless`**, naming the line and never the reader's action
- **R218:** The census states **every count, zeros included**: carves with a status block, open, landed, stateless, non-conforming, documents with no status block
- **R301:** What the carve reader **could not read** — every bracket group still open at end of input — is counted as `unread` on the carve's line and in the census always, and listed under `--open` as the opener's line and marker; the word names the reader's action, never a state of a part
- **R219:** The part line's rules — key form, separator, checkbox interior, `OPEN` attribution, marker verb case — are **`minispecsdom`'s**, and this tool surfaces the reader's `Deviations()` with their targets rather than re-deriving any rule
- **R220:** `SetMarker(path, key, verb, attribution)` and `SetPartLanded(path, key, attribution)` read the carve, apply the reader's `SetMarker` or `Land`, and write the rendered bytes by **temp-file-and-rename**; a refusal from the reader — deviations on the line, `OPEN` over a checked part, a landing of a landed part, a key no part carries — is passed through and **leaves the file byte-identical**

## Feature: Backup Slot
**Source:** specs/backup.md

- **R221:** The backup slot gives **one level of undo and one of redo** over the trajectory files — the most recent change is revertable and replayable and nothing older is recoverable; it is not an undo stack
- **R222:** The slot covers the **trajectory files only** — the pending, current and done files at the repository root — and never a carve, which is written on a revert rather than restored
- **R223:** The copies and the stamp live together in `.minispec/backup/`, the path `init` already has git ignoring
- **R224:** Every operation — change, revert, replay — is the **same swap**: copy the live files to a temporary location, perform the operation, write the stamp, then **move** the temporary copies over the old backups; the move is last and atomic, so a crash leaves the old backup intact or the new one complete
- **R225:** **One backup set suffices**, because it always holds *the other state*; revert and replay are one mechanism in two directions
- **R226:** The stamp records the **state and nothing else** — not the item ID, since the snapshot is the pending file and holds the entry with its number
- **R227:** Revert and replay first verify that **no covered file changed since the stamp**, by modification time against the single stamp; if any did, the operation **refuses** and names the files and where their backups are
- **R228:** The slot has **three states** — `changed`, `reverted`, `replayed` — and from every state **exactly one** of revert and replay is legal; a refusal names the state and what it accepts
- **R229:** A new mutation from **any** state resets the slot to `changed` and replaces the backup; whatever was revertable is gone
- **R230:** Revert **restores** the trajectory files and marks the departed item's carve part `REVERTED (#N.)`; replay returns it to `OPEN (#N.)`; the item is derived from the pending file's two sides, never stored
- **R231:** A release — a new mutation arriving while the slot holds a **reverted** attempt — returns the part to `OPEN (not queued.)`, returns the number to the pool, and writes **no done entry**
- **R232:** A release happens **only from the reverted state**: a completion produces the same entry diff as a revert, so the diff alone must never trigger one
- **R233:** A release **never touches a part whose checkbox is `[x]`**, a second guard independent of the state guard
- **R234:** The released item IDs are **queryable before the mutation** that releases them, so the vend can announce a reuse
- **R235:** Every change the slot makes is **cranked out in full** by the verb that drives it, and `/mini-spec` documents the slot and revert especially
- **R236:** The **worktree anchor** records the working tree as it stood immediately before every transition, at a ref the tool owns outside the stash, replacing what it held in one act; it is reference, never undo
- **R237:** The anchor holds untracked file **contents** and **excludes ignored paths by construction**, built from a scratch index so the repository's real index and working tree are never touched
- **R238:** The anchor's **first parent is the commit that was checked out**, so what it was taken from is recoverable from the anchor itself; a repository with no commits gets an anchor with no parent
- **R239:** A tree with no git is not a failure of the slot: the anchor is skipped and the trajectory files are still restored
- **R240:** Pending entries are read through `minispecsdom.Pending` as `QueueEntry{ID, Title, SourceDoc, SourceKey, Kind, Line}`; the slot reads them on both sides of the swap and needs nothing else from the pending file
- **R335:** `minispec pending changes` reports what has moved in the working tree since the anchor the last transition wrote: a tree of the tree as it stands is built by the anchor's own scratch-index method and diffed against the anchor's tree, and every path is listed as `M` changed, `A` new — tracked or not — or `D` deleted and recoverable, three categories from one tree-vs-tree diff with no side-car list; it reads and builds objects and alters nothing, and with no anchor it refuses rather than reporting a clean tree it never compared
- **R336:** The report names the question it answers — everything that moved since the queue last transitioned, the user's unrelated work included, never what the last item did — and carries the two repair commands, `git checkout refs/minispec/snapshot -- <path>` for the tree as at the transition and `…^1 -- <path>` for the commit beneath it, because the private ref surfaces nowhere else; the tool reports, the agent proposes, the user decides

## Feature: Queue Items
**Source:** specs/queue-items.md

- **R241:** `minispec pending add-item --from <doc>#<part> "<title>"` **mints the item ID and writes both sides of the link in one invocation** — the queue entry carrying the part pointer, and the part line's `**OPEN (#N.)**` marker written by `minispecsdom`'s marker rule (R219). Assignment and the write that records it are one act, which is what keeps R190's `max()` the whole truth: a number is in a document the moment it exists. A command that reserved a number for the agent to write later would be a second copy of the numbering state by construction
- **R242:** The link is **asymmetric, and structurally so**: the carve carries a **bare key** and the queue side holds the recorded pointer `<doc>#<part>`. The public→private direction is never a markdown link, because trajectory files are private in every project and a carve cannot point at a file a cloner does not have
- **R243:** `add-item` **refuses rather than guessing**, naming what it looked for and where: an unresolvable `<doc>#<part>` — no such document, no such key, a document with no status block, a key the reader lists as non-conforming — and a part that **already carries a queue ID**, since a part records exactly one item. Nothing is written on the way out. The reuse announcement on a returned number is R234's and is satisfied here
- **~~R244:~~** (Retired T8 — see R476) `minispec pending finish <N> --commit <hash>` performs the completion **in the mandated order: the source first**, then the current file, then the queue entry. Each discharged part is checked off in its carve; the current file's `## Active` section is reset to its placeholder; the entry moves from the pending file to the done file. Source first because the carve is the copy a future reader trusts, and the one nobody thinks to check
- **~~R245:~~** (Retired T9 — see R477) **Checking off a part is all three markings at once** — checkbox `[x]`, title struck through, and a `` **LANDED (`<hash>`, <date> — `#N`.)** `` record appended — performed by `minispecsdom`'s `Land` through `SetPartLanded` (R220). A landed part's queue ID lives in that record and nowhere else
- **R246:** Completion checks **the parts the item recorded and nothing else**. A parent part is never checked directly — it completes when its subparts do, derived rather than stored. An item may discharge parts in several documents and each is checked: a part records one item, an entry records a list of parts
- **~~R247:~~** (Retired T10 — see R478) `finish` composes the done entry's **header** — date, identifiers, title, commit and part pointer — and **never its body**; a body it is handed (R262) is placed, not authored. *Enough to reconstruct the change without re-reading the code* is a judgment about a future reader, and the tool owns IDs, not prose
- **R248:** All three verbs run **inside the backup slot** (R221–R240), one `Slot.Record` per invocation, so one level of undo covers the whole invocation rather than any single file and a mis-typed part pointer is a `pending revert` rather than a repair
- **R249:** `finish` clears the current file's **`## Active` section and nothing else**. The section is `minispecsdom`'s `Current` region — a heading node ending at the next heading of level 2 or higher — so the active item's context may use `###` and below freely, and standing context in the file's other `##` sections is outside the verb's reach **by construction**: the write removes the region's nodes and inserts one, and has no way to name anything else
- **R250:** **Two refusals, each naming the repair**, both the reader's (`ParseCurrent`) and surfaced by every verb that touches the current file: no `## Active` heading — the file is legacy or damaged; more than one — the region to clear is ambiguous and is refused rather than picked. Absence is reported as an error, never as silence
- **R251:** A part pointer requires **both** halves. An entry whose `Source:` names a document but no part key, or names a gap, records **no part**, and completion has nothing to mark in a carve — the ordinary case, since most queue items point at a spec or a plain document. `QueueEntry.Parts()` and `Gap()` read the same `Kind` the slot's release path reads (R240)
- **R252:** `add-item` writes the **whole** queue entry in the shape `trajectory-format.md` mandates, through `minispecsdom`'s `EntryText` — the `##` heading carrying the number, the title, the skill and the one-line status; the `Source:` line with the part pointer or gap ID; and the `Next:` line when one is given
- **R253:** The **status sentence is required and the `Next:` line is not**: the status lives inside the heading the tool mints while `Next:` is a whole line of its own. A tool may decline to write a line; it may not write two thirds of one
- **R254:** Each of `add-item`'s three prose slots has a **file form** read byte for byte — the title positionally or `--title-file`, `--status`/`--status-file`, `--next-action`/`--next-action-file` — and giving one slot **both** ways is refused rather than resolved by precedence, because a backtick inside a shell argument is command substitution that vanishes without a word. Only the trailing newline a heredoc adds comes off
- **R255:** The item title is **plain text and the verb supplies the emphasis**: a title wrapped end to end in a single `**…**` run is refused, naming the repair, while emphasis *inside* a title is untouched. Whether the reader reads an interior run back whole is `minispecsdom`'s — see gap `O13`
- **R256:** `add-item` **places** the entry at the position the caller states: `--next`, `--nth N`, `--after N`, or `--last` — the default, spelled out. The flags are mutually exclusive and giving two is refused. Which position an item deserves stays the caller's judgment; moving the block there is mechanics
- **R257:** `--next` means **next to be worked**, not position 1: with nothing in progress the entry goes to the top, with a step in progress immediately after the item being worked — position 2, because the pending file's own rule is that the top item is active. "In progress" is the current file's `## Active` holding something other than its placeholder (`Current.Occupied`)
- **R258:** `--nth 1` is **refused while a step is in progress**, since position 1 is the active item's slot, and the message names both repairs: `--next` if *next* was meant, or park the active item first if the caller means to preempt it
- **R259:** An `--after N` naming **no live entry** is refused, and an `--nth N` outside `1 … entries+1` is refused **rather than clamped** — a clamp is a silent reinterpretation of an instruction the caller was specific about. `--after` addresses by item ID and `--nth` by position; `--nth` survives the IDs-are-addresses rule only because the position is consumed in the same invocation and never stored
- **R260:** The pending file's entries are **heading nodes as the reader sees them**, so a `## 5.` quoted inside a fenced example cannot be counted, targeted or split, and the last position stops where the entries stop — a `---` rule found in a text run's own content, a heading of level 2 or higher, or the end of the document. This is `minispecsdom.Pending`'s guarantee (`Place`, `After`, `Remove`), consumed through the adapters
- **R261:** `add-item`'s crank handle **names the placement flags** rather than instructing a hand edit into the file it just wrote; a placement that defaulted to last says so and names the alternatives
- **R262:** `finish` accepts the done entry's body as `--body <text>` or `--body-file <path>`, refused together, the file form read byte for byte
- **R263:** The body is written in the **same operation as the header** (`Done.Prepend(header, body)`), never as a second edit that finds it again: an edit that does not happen cannot change the wrong region
- **R264:** With no body given, `finish` completes the item and **cranks the hand edit, naming the flag that avoids it**. The crank handle is the fallback and never the design
- **R265:** `pending start <N>` writes the current file's `## Active` section: the tool composes the identity line `` `#N` — <title> `` from the queue entry, so it and a later done header cannot disagree, and places the caller's context beneath it byte for byte through `--context` or `--context-file`, refused together
- **R266:** `start` **refuses a `## Active` that already holds an item** (`Current.SetActive` → `ErrOccupied`), naming the repair — park the active item as a sub-item in the pending file — rather than overwriting the one file where nothing reports a loss
- **R267:** `start` addresses the region through the **same reader and write path** `finish` uses to clear it, and runs inside the backup slot like the other two: it is the only verb that writes before any other record of the work exists, so a loss under it is the loss of the sole copy
- **R268:** `finish` accepts `--discharged <text>` or `--discharged-file <path>` and writes the done header's identifier slot as `#N / <text>`, joined with the ` / ` the format mandates. The tool writes the `#N` because it owns IDs; it never infers a requirement range from `requirements.md`, because a generated identifier list reads exactly like an authored one. An absent flag leaves the slot as `#N`
- **R269:** A `--discharged` value containing a **colon** is refused, naming why: the reader takes the slot as the run between the date's em dash and the colon that opens the title, so a colon inside it silently ends the slot early. A verb that owns a format refuses the input that would make its own reader wrong
- **R270:** The entry is removed as **the heading node that opens it and the nodes beneath it**, split at the `---` rule when the boundary falls inside a text run, and the done entry is placed before the ledger's first entry or after its preamble — both `minispecsdom`'s (`Pending.Remove`, `Done.Prepend`), so a `## N.` or a `---` quoted inside a fence can neither bound a removal nor receive a record. `add-item` then `finish` leaves the pending file **byte-identical**
- **R271:** `pending add-item --from` accepts a **gap ID** as well as a part pointer, distinguished by shape and needing no flag: a gap ID is a capital letter and digits with no `/`, no `#` and no `.md`
- **R272:** A `--from` naming a **range or a list** of gaps (`O5-O6`, `O5,O6`) is refused naming the rule that an entry carries one pointer, rather than reported as a syntax error; a gap-shaped token that is not one gap ID gets that refusal and never the part pointer's
- **R273:** A gap source **requires a design root**, since gap IDs are scoped to a `design.md` and a repository may hold several roots. When none resolves, the verb refuses naming what it needed and where it looked; an ID the design root holds no gap for is refused too. A part pointer is unaffected: it names its own document
- **R274:** `add-item` writes **nothing on the gap side**. A part gets its `**OPEN (#N.)**` marker because the carve is public; a gap has no such marker and gains none. The consequence is stated: there is no gap analogue of the carve→queue cross-check
- **R275:** A gap-sourced entry's `Source:` line reads `` Source: [<doc>](<doc>), gap `<ID>`. `` — `minispecsdom`'s gap form (`EntryText{Kind: SourceGap}`), the document as a markdown link and the gap ID backticked without a `#`
- **R276:** An entry names **one gap**, held in the same scalar pointer a part fills; further gaps an item addresses are named in its prose
- **R277:** `pending finish --resolve` resolves the gap a gap-sourced item names, in the design root **the entry named** — a resolver bound to a different design root refuses, naming the gap and both roots. Resolving is never automatic: an item may address a gap only partly, and a gap wrongly marked resolved is work that silently never happens
- **R278:** A gap-sourced completion records `` Gap `<doc>#<gap>` `` in the done entry header, in the shape a part-sourced one records `` Part `<doc>#<key>` ``, so the ledger keeps the link the pending file held
- **R279:** Completing a **gap-sourced** item requires an explicit decision: `--resolve` or `--no-resolve`. Neither given is a refusal naming both spellings, raised **before anything is written** — a notice mitigates forgetting, a required choice makes it impossible
- **R280:** `--resolve` and `--no-resolve` are refused together, before the item is looked up; and a resolve flag on an item that names no gap is reported as having done nothing, since a flag that silently does nothing reads exactly like one that worked
- **R281:** `--no-resolve` leaves the gap open and the completion **records that it was a decision** (`left open … by decision`), because a gap left open deliberately and one left open by oversight are identical in `design.md`
- **R282:** Every change the three verbs make is **cranked out in full** — the number minted, the position taken, the parts checked, the gap resolved or left open, every file written — markdown on stdout by default and `--json` for the machine-readable form
- **R283:** `pending revert` and `pending replay` drive the slot's two operations and report the state before and after, the files restored, the backup directory, and the worktree anchor with the commands that read it; a refusal is the slot's message and a non-zero exit. `pending` resolves the repository root and never a design root, since the trajectory layer sits above every design root
- **R329:** Every queue verb's `written` line says what kind of file each write landed in — `tracked, uncommitted` for a file git tracks, `ignored` for one it ignores, `untracked` otherwise, and nothing where there is no repository — so the one tracked public document among a completion's four writes (the carve flip) is marked as the write that still needs a commit, beside the three gitignored ones; the tool reports and never stages or commits
- **R330:** Re-adding the part a reverted attempt released is not a collision: when the slot holds a reverted attempt and the part's queue ID sits in a `REVERTED` marker, `add-item` proceeds — the release that turns the marker back to `OPEN (not queued.)` runs inside the same mutation, ahead of the new marker — so the common case, re-queueing the very part just rolled back, is not refused while a sibling's mutation would have released it; a `REVERTED` marker outside the reverted state, or an `OPEN` marker naming a live item, is still the refusal it was
- **R332:** A queue verb — `add-item`, `start`, `finish` — run where any of the three trajectory files is missing **refuses before anything is minted and names the missing files**, and the CLI cranks out how to create them: the `add-item` again with `--create`, and what git will do with the files under the project's `track` — the `.gitignore` question folded in, since `init` answered it; never a raw open error
- **R333:** `pending add-item --create` writes each **missing** trajectory file with the lifecycle preamble `trajectory-format.md` mandates — the pending file's ordering and ID rule above its rule, the current file with its `## Active` placeholder, the done file — reports them like any other write, and then proceeds; a file that exists is never touched, and a second `--create` writes nothing
- **R476:** `minispec pending finish <N>` performs the completion **in the mandated order: the source first**, then the current file, then the queue entry, and needs no commit to exist: the item number is the identifier (Bill, 2026-09-15), so the completion runs before the commit and the carve flip lands in its own item's commit. Source first because the carve is the copy a future reader trusts.
- **R477:** **Checking off a part is all three markings at once** — checkbox `[x]`, title struck through, and a `` **LANDED (<date> — `#N`.)** `` record appended, **carrying no commit hash** — performed by `minispecsdom`'s `Land` through `SetPartLanded` (R220). A landed part's queue ID lives in that record and nowhere else; the commit is found by `git log --grep` on that ID, which is why every commit names the items it lands. Older records carrying a hash are read as they stand and never written again.
- **R478:** `finish` composes the done entry's **header** — date, identifiers, title and part pointer, and no commit — and **never its body**; a body it is handed (R262) is placed, not authored. The ledger is private and never squashed away, so a hash may be added there after the fact by hand; the tool records none.
- **R479:** `minispec pending commit-message` composes the commit message for the uncommitted items: the newest done entries down to, and excluding, the first one a commit names — by a hash in its slot, or by every identifier appearing in a message on `HEAD`'s history, `#N` bounded by a non-digit so `#40` never names `#4`; everything older than a named entry is history, whatever its slot says; a repository with no commits names nothing.
- **R480:** The message's subject is the item numbers and titles — `#85, #86: <title>; <title>` — and its body opens with `Items #85, #86.` on its own line, then each entry in the order finished as `#N — <title>` followed by the entry's body from the done file; the tool adds no sign-off.
- **R481:** With every done entry named by a commit the verb refuses, naming the newest entry and the commit that names it, rather than composing an empty message.
- **R482:** `--amend` returns `HEAD`'s message unchanged with `Also lands #N.` and the new entries appended after it, never rewriting a line of the existing message.
- **R483:** `--amend` refuses when `HEAD` is contained in any remote branch, naming the branch, because an amended shared commit rewrites history someone else holds.
- **R484:** The message goes to stdout, or with `--out <file>` to that file byte for byte, for `git commit -F`; the verb never stages or commits, resolves at the repository root and needs no design root.

## Feature: Trajectory Validation
**Source:** specs/validate.md

- **R284:** `minispec validate trajectory` reports the consistency of the trajectory layer — the queue files at the repository root and the carves that point at them. Read-only, like every other validation
- **R285:** It is a **separate subcommand and not part of bare `validate`**, and it resolves the **repository root** rather than a design root: `validate` is design-scoped and runs per design root, while one repository holds one queue and may hold several design roots — this one holds `tool/` and `example/`. Folding it in would report the same drift once per root. The gate belongs to the Makefile, which runs both
- **R286:** A repository running **no trajectory layer passes**: neither queue file and no carve directory is nothing that could be inconsistent, so it says so and exits 0 — a different report from *could not check*, never collapsed into it (R194 draws the same line for `next-id item`)
- **R287:** **Carve → queue:** every `#N` a carve's status block cites resolves to an entry in the pending file or the done file, both read through `minispecsdom`'s readers. A citation to an item that never existed, or whose number was reused, is a pointer into nothing
- **R288:** **Queue → carve:** every pending entry whose `Source:` names a carve part appears in that carve's status block under the key it claims — the direction R287 cannot see. A gap-sourced entry names no part and is not checked here
- **R289:** Citations are ingested **by position** — a part line inside the status block and the marker on it, as `minispecsdom`'s carve reader hands them over (`Part.QueueID`) — never from a pattern swept over prose. An extractor sweeping `**VERB (…)**` across a status block once read a backticked prose example as a live citation and reported a dangling `#121` in a repository that never had one
- **R290:** An item ID held by **both** the pending and the done file is reported. **Repetition within the done file is not a collision**: an item that lands in stages is legitimately recorded across several entries, and from the number alone that is indistinguishable from a reuse — the first draft reported three correct staged records in ark as reused, which is the shape of check that gets muted
- **R291:** **Orphans** are reported in both forms: a carve part marked landed against a queue ID with no done entry, and a done entry naming a part that no carve records
- **R292:** A done entry whose header carries **no identifier slot** is reported as *unmigrated* rather than skipped, from the reader's `HasSlot` — a header with no slot and an entry that legitimately discharged no ID both yield no IDs, so the shape and not the count is what distinguishes them
- **R293:** Conformance is **named with its migration target**, never merely complained about, and the report states what it therefore could not see: unkeyed part lines (the superseded bare-`#N` scheme) carry their queue references where the position rule cannot read them, and their count is printed beside the findings, never instead of them
- **R294:** **Checkbox agreement:** a status line states its state three ways — checkbox, strikethrough, marker — and they must agree, the checkbox authoritative. `OPEN`, `REVERTED` and `DEFERRED` are open-class verbs; `LANDED`, `MIGRATED`, `DISCHARGED` and `SENT` are done-class; a verb outside both is unclassified and silent, since asserting a disagreement about an undefined word would be the tool inventing intent. A sentry over a corpus normalised by hand
- **R295:** Item numbers appearing in **no readable entry** are reported: an ID is assigned at creation, so every number from 1 to the maximum assigned should be accounted for, the maximum read from both files (R190). When R297 also fires the report names the unread count beside the gaps, because it very likely explains them. An unrecognized entry holding the *highest* ID leaves no gap, so only R297 can see it
- **R296:** `CURRENT.md` carries **exactly one `## Active`**, reported before the reference-level findings; both the missing heading and a duplicated one are findings, through the same reader and repair text the write path uses (R250). A missing current file is not this finding
- **R297:** **Entry-like lines the readers did not recognize** are reported as *coverage*, per file, from `minispecsdom`'s `Unread()` on both queue files — printed whether or not anything else fired, and never an issue in themselves. Every shape-based check is blind by construction to a line outside the shape
- **R302:** The unread coverage note counts per file across **every document a reader touched** — the two queue files, the current file, and each carve — so a bracket group never closed in any of them is named at that file, at its opener's line
- **R298:** A **status-block line the reader could not read as a part and lists as deviating** (`Stateless()` with deviations) is an issue, named with file, line and reason, and listed first because every other check reads through the parse it reports on; a checkbox-less `SPLIT` or `MOVED` parent carries no deviation and is not one. The August tree's independent flat-scan cross-check of the status block is not carried — see gap `O18`
- **R299:** `validate trajectory` writes markdown to stdout, honours the global `--json` flag with one key convention, and exits 0 when consistent and 1 when it finds issues
- **R300:** **The two readers of the queue files must agree, and this check is the second opinion.** `ScanTrajectory` reads item IDs line by line; `minispecsdom`'s document readers return entries. Every ID one saw and the other did not is a finding, listed first, per file. Measured 2026-09-05 on this repository: the line scan read 58 IDs from the done file and the document reader returned 17 entries with nothing unread — one unclosed backtick in an entry body absorbed the remaining 41 entries — and every check downstream reported four landed parts as orphans. A check that reads through one parser cannot see what that parser swallowed
- **~~R485:~~** (Retired T14 — see R491) `validate trajectory` classifies every markdown link in the live carves and `carves/done/` as `query links` does; `ignored`, `missing` and `outside` links fail the phase, each reported with the citing file, line, link as written and class; the trajectory files are not checked.
- **R486:** An `untracked` link is reported as a note beside the findings and never fails the phase; `tracked`, `external` and `local` links are not reported.
- **R487:** Outside a git working tree the link check reports that it could not classify, as a note, and the trajectory checks still run; it never reads clean over links it could not see.
- **R491:** `validate trajectory` classifies every markdown link in every markdown file git tracks under the repository root as `query links` does; `ignored`, `missing` and `outside` links fail the phase, each reported with the citing file, line, link as written and class; untracked and ignored documents, the trajectory ledgers among them, are not checked.
- **R499:** `validate trajectory` reports a done entry whose `Part` pointer — a code span ending in `.md` before its `#` — names a document that does not exist, and the trajectory files' unresolvable links, as notes beside the findings, never failures: the ledger is private, and entries naming documents dropped at a restart or belonging to another project are history nobody will repair.

## Feature: traceability comment
**Source:** specs/traceability-comment.md

- **R337:** `TraceabilityComment` is one node kind for every language, tiling the whole comment
  from opener through closer.
- **R338:** The interior is fields separated by `|` — `CRC:`, `Seq:`, `Test:` each with a plain
  list, and a keyword-less requirement list — at most one of each, in any order, followed by an
  optional description after a separator.
- **R339:** A `:` is a field-key colon only immediately after `CRC`, `Seq` or `Test`; anywhere else
  it is the description separator.
- **R340:** A `Seq` item may carry `#step`, and the typed view splits path from step.
- **R341:** Whitespace, keywords, `|` and the separator are computed glue; an unedited comment
  renders back byte-exact; the separator is preserved when unedited and written as `--` on a fresh
  write; the description is bound and writable.
- **R342:** Recognition is a parse that consumes the whole interior; a comment that leads with a
  field but leaves bytes uncovered is not a traceability comment.
- **R343:** `Parse(cmt *Opener, ctx *BracketContext) bool` fills the node off to the side, touching
  no document, and returns false when the interior is not a single text node or is not consumed.
- **R344:** On success the children are the original `*Opener`, the interior's glue and fields, and
  the original `*Closer` — reused, not recreated.
- **R345:** The interior is parsed by a segment walk with one stencil per `|`-segment, and the
  results splice flat into the node.
- **R346:** `Comments(d, ctx)` is the second pass: every opener whose group kind equals the
  language's `Comment.Kind` is a candidate, each success replaces its run from opener to closer
  inside its own mutation window, and nothing else is touched.
- **R347:** `New(lang, Fields)` assembles the canonical interior — CRC, Seq, Test, refs, `--`
  description — and runs the same interior walk at a synthetic location inside synthetic markers,
  so no child list is hand-built and the node has no origin.
- **R348:** The node exposes typed accessors `CRC`, `Seq`, `Test`, `Refs` and `Description`, nil
  when the field is absent.

## Feature: part line
**Source:** specs/part-line.md

- **R349:** `PartLine` is one node over a list item line, from the `- ` marker to the byte before
  the newline, reusing the `ListItem`, `Checkbox` and every bracket marker the base emitted.
- **R350:** Every list item line parses; `Parse` returns false only when the item is not in the
  document, and belonging to a status block is the carve schema's business.
- **R351:** The head is the first bold run after the checkbox; the line is keyed when its interior
  begins `Item N — ` or `N.M — ` — the word required without a dot and forbidden with one, the em
  dash and nothing else — and the key is a bound `Text`, the separator glue, the title the rest of
  that text.
- **R352:** A later bold run whose interior reads `VERB (attribution)`, the verb in capitals as one
  or more words joined by spaces or hyphens, is a `MarkerSpan`; anything else is interspersed text.
- **R353:** `Checkbox()` is the base's own `Checkbox` node, nil when the line has none.
- **R354:** `IsStruck()` derives from whether the head sits inside a `~~` group; `Strike(bool)`
  inserts or removes the `~~` pair around the head's bold run among the node's own children — the
  flat array is untouched, so no mutation window is involved; no consumer touches a `~~` node.
- **R355:** Deviations are reported, each naming its rule and target shape — unkeyed head,
  non-em-dash separator, non-conforming checkbox interior, verb not in capitals, an `OPEN`
  attribution off its shape (R397), the comma-form marker (R398) — and the line still parses.
- **R356:** `MarkerSpan` tiles `**` through `**` reusing both; its verb is a bound `Text`; its
  attribution and queue ID are derived by rendering the nodes between the parentheses.
- **R357:** `MarkerSpan.Set` rewrites the interior canonically as one text under the guarded write:
  the render re-parsed as a marker must yield the same verb and attribution, or the write is
  refused and the literal unchanged.
- **R358:** `Splice` replaces the line's run with the node inside a mutation window; `PartLines`
  parses and splices every list item in document order.
- **R359:** `PartLine` and `MarkerSpan` each declare their own `Equals`, comparing children.
- **R360:** Bound texts and glue are re-cut from the interior texts; no bytes are lost or
  normalised on read, and an unedited line renders back byte-exact.
- **R410:** A `REVERTED` attribution that is not `#N` is the `REVERTED attribution` deviation.
- **R417:** `Key()` returns the fragment — `4` for a part, `2.2` for a subpart — the one key form,
  which `Part`, `SetMarker`, `Land` and the pending `Source:` line all use; `Item ` is the head's
  display word, not part of the key.

## Feature: carve schema
**Source:** specs/carve-schema.md

- **R361:** `Carve` embeds the markdown base and owns a carve file's DOM; `ParseCarve` is the
  only way one is made and `Render` re-emits it.
- **R362:** The status block is the region from the level-2 heading `Status` to the next heading
  of level 2 or higher or the end of file; `HasStatus` is false when there is none.
- **R363:** Only list items inside the status region become part lines; other lists in the body
  are left as the base parsed them, and a fenced sample is invisible by construction.
- **R364:** A status line with a checkbox is a `Part`; one without is `Stateless`, recorded and
  never given a state.
- **R365:** `Depth` is the bullet's leading whitespace, and `Parent` is the nearest preceding
  part with a smaller depth, or nil.
- **R366:** `Part(key)` finds a part by its key; a write to a key no part carries is an error.
- **~~R367:~~** (Retired T7 — see R409) `SetMarker` replaces the first transient marker — verb `OPEN` — removes any other
  transient, and appends a marker when the line carries none; it selects by what it replaces.
- **R368:** `Land` checks the box, strikes the head, and sets `LANDED (attribution)` through the
  marker rule, in one act.
- **R369:** Every part carries its line's deviations.
- **R395:** `Part`, `Entry` and `DoneEntry` report `Line()`: the 1-based line of the part line,
  heading or bullet as the document stood at parse time.
- **R396:** `Pending.Unread` and `Done.Unread` return `[]Unread`, each with the line and text of
  what was not read.
- **R397:** An `OPEN` attribution reads `#N` or `not queued` in any case, with any inner spacing and
  an optional full stop; anything else is the `OPEN attribution` deviation.
- **R398:** A bold run after the head whose interior opens with a verb and a comma — the
  superseded `OPEN, not queued` form — is not a marker and is reported as the `marker scheme` deviation.
- **R391:** `SetMarker` and `Land` decide refusal before any marking, so a refused write leaves the
  line byte-identical.
- **R392:** A write over a line carrying deviations is refused with a `DeviationError` naming the
  key and every deviation's rule and target.
- **R393:** `SetMarker` refuses `OPEN` over a part whose checkbox is checked, with `ErrReopen`.
- **R394:** `Land` over a part whose checkbox is checked is refused with `ErrLanded`, not made
  idempotent.
- **R404:** `Carve.Unread` lists every group open at end of input at its opener's line, with the
  text *`<marker>` never closed*.
- **R409:** `SetMarker` replaces the first transient marker — verb `OPEN` or `REVERTED` — removes any
  other transient, and when the line carries none inserts the marker after the head and any markers
  and before the trailing prose; it selects by what it replaces.
- **R415:** After the write, a fresh parse of the render finds the marker on the keyed part — for `Land`
  its checked, struck box too — or the write panics with a `ReadBackError`.

## Feature: pending schema
**Source:** specs/pending-schema.md

- **R370:** `Pending` embeds the markdown base and owns the pending file's DOM; `ParsePending`
  makes one and `Render` re-emits it.
- **R371:** An entry is a level-2 heading whose text opens `N.`, and its region runs to the next
  heading of level 2 or higher or a `---` line outside a fence; a fenced heading cannot end it.
- **R372:** An entry is a view over its run of flat nodes, its values derived from the rendered
  bytes at the format's positions — number, title, skill, status, Source document and part key,
  Next — and nothing in the run is re-cut.
- **~~R373:~~** (Retired T5 — see R407) `Place(e, pos)` inserts the canonical entry text as one synthetic node before the entry
  at `pos`, or at the end when `pos` is one past the last; a position outside that range is
  refused, not clamped.
- **R374:** `After(id)` resolves the position following a live entry and refuses an unknown id.
- **~~R375:~~** (Retired T6 — see R408) `Remove(id)` drops the entry's run inside one mutation window, splitting a shared tail
  text at the region's end.
- **R376:** `Unread` lists level-2 headings that are not entries; `MaxID` is the largest entry id.
- **R399:** A `Source:` line names a part or a gap, told apart by the word and by shape; `Entry.Kind`
  says which was read and `SourceKey` carries the key for either, the part key without its `#`.
- **R400:** A gap source names exactly one gap ID; a range, a list or a `#` there reads
  as `SourceNone`, and a `Source:` line that read as neither form is listed by `Unread` with its line.
- **R401:** `EntryText.Text()` writes the part or gap form by `Kind`; `Place` refuses a gap key that is
  not one gap ID with `ErrBadGapSource`.
- **R377:** After each write the document is re-read from its bytes and the entries re-derived, since
  a placed entry is one synthetic text until it is parsed.
- **R403:** `Pending.Unread` lists every group open at end of input at its opener's line, with the
  text *`<marker>` never closed*, after the unread headings — file order, for the same
  reason as the done schema's.
- **R407:** `Place(e, pos)` inserts the canonical entry text as one synthetic node before the entry
  at `pos`; at one past the last it lands where the entries end — before the rule that closes the
  region when one follows, else at end of file with the separator adjusted so the file ends in one
  newline, and with no entries at all after the header's rule; a position outside `1 … len+1` is
  refused, not clamped.
- **R408:** `Remove(id)` drops the entry's run inside one mutation window, splitting a shared tail
  text at the region's end, and when the entry was the last thing in the file drops the blank line
  its placement opened, so that `Place` then `Remove` is byte-identical.
- **R412:** An entry's title is the interior of the heading's first emphasis run, read to its own
  close, with the ID taken from the bytes before it and the skill and status from the bytes after.
- **R413:** After the re-read, `Place` finds the entry at its position with every field and `Remove` no
  longer finds it, or the write panics with a `ReadBackError` naming reader, write, key, want and got.
- **R447:** A `---` counts as the rule only when it is a whole document line outside a code group;
  a code span's interior reading `---` and a `---` line inside a fence are body text, and the header
  rule is found before a placement's mutation window opens.
- **R501:** The canonical entry text `Place` renders begins every line at column zero — no
  hanging indent under the heading — while the `Source:` and `Next:` readers accept leading
  whitespace, so entries written under the earlier indented stencil still read.

## Feature: done schema
**Source:** specs/done-schema.md

- **R378:** `Done` embeds the markdown base and owns the done file's DOM; `ParseDone` makes one
  and `Render` re-emits it.
- **~~R379:~~** (Retired T4 — see R396) An entry begins at a column-0 list item whose text opens with bold; a column-0 list item
  that does not is entry-like and counted by `Unread`.
- **R380:** A region runs to the next entry-like bullet at column 0 or a heading of level 2 or
  higher; a fenced quotation can neither begin nor end one.
- **R381:** The identifier slot is the run between the header's em dash and the colon opening the
  title; every `#N` in it is a discharged queue ID, and a `#N` anywhere else is prose.
- **R382:** The part pointer is a backquoted `doc#key`, taken from the header first and the body
  second; the date, title and first backquoted commit are derived from the header.
- **R383:** `Prepend(header, body)` inserts the entry as one synthetic text just after the
  preamble's rule — before the first entry, or at the end when there is none — and the document is
  re-read after the write.
- **R384:** `MaxID` is the largest queue ID in any identifier slot; values are derived from the run's
  rendered bytes and never stored.
- **R402:** `Done.Unread` lists every group open at end of input at its opener's line, with the
  text *`<marker>` never closed*, after the entry-like lines — file order, since nothing
  structured can follow a group still open at the end.
- **R411:** Every reader's `Unread` also lists each closer the context reports paired with nothing,
  at its line, with the text *`<marker>` closes nothing*, in line order with the rest.
- **R445:** Every reader's `Unread` also lists each opener the context reports demoted, at the line
  of its marker, with the text *`<marker>` never closed, read as text*, in line order with the rest.
- **R414:** After the re-read, `Prepend` finds the new first entry with its header line, or panics
  with a `ReadBackError`.

## Feature: current schema
**Source:** specs/current-schema.md

- **R385:** `Current` embeds the markdown base and owns the current file's DOM; `ParseCurrent`
  refuses a document with no `## Active` heading or more than one, matched on `Heading` nodes.
- **R386:** The active region runs from the heading to the next heading of level 2 or higher, or
  the end of the file.
- **R387:** `Active` is the region's body trimmed, `""` when it holds only the
  placeholder `_No active item._`; `Occupied` is the region holding anything else; `Standing` lists the other
  level-2 headings.
- **R388:** `SetActive` and `Reset` replace the region's body as one synthetic text and address no
  node outside the region, so every byte outside it is unchanged by construction.
- **R389:** `SetActive` refuses when the region is occupied.
- **R390:** The document is re-read after each write.
- **R405:** `Current.Unread` lists every group open at end of input at its opener's line, with the
  text *`<marker>` never closed*.
- **R406:** `ParseCurrent`'s two refusals are the sentinels `ErrNoActive` and `ErrManyActive`, told
  apart with `errors.Is`.
- **R416:** After the re-read, `SetActive` reads its body back and `Reset` the placeholder, or the write
  panics with a `ReadBackError`.

## Feature: test-design schema
**Source:** specs/testdoc-schema.md

- **R418:** `ParseTestDoc` parses with the markdown base; an entry is a level-2 heading whose text
  begins `Test:` together with its region, which runs to the next heading of level 2 or higher or
  the end of the file; the title is the text after `Test:`. `Tests` lists entries in order.
- **R419:** A level-2 heading that is not a test is listed in `Unread` rather than guessed at; a
  fenced heading is no heading to the base, so it neither opens an entry nor ends one.
- **R420:** A field is `**Name:**` at the head of a line inside the entry and outside any code
  group. Five names are read — `Fire alarm`, `Inject`, `Pulled`, `Code`, `Alarm`; any other name at
  a line head is body but still ends the field above it. The two prose fields, `Fire alarm`
  and `Pulled`, fold across continuation lines to the next field line or the end of the entry; the
  three list fields, `Inject`, `Code` and `Alarm`, are one line each, and a line following one is body.
- **R421:** A line the base's context places inside a code group is body on the read side and
  unreachable on the write side: a fenced `**Alarm:** 1` names nothing.
- **R422:** `Fire alarm` is prose and its presence is `HasAlarm`; `Inject` splits on commas
  into `file:symbol` sites, trimmed; `Pulled` is a leading `YYYY-MM-DD` date and everything after it as
  the body, and a `**Pulled:**` line with no leading date is a deviation; no commit is read after the
  date; `Code` is a comma-separated list as written; `Alarm` is an integer, and a non-integer is a
  deviation; an `**Alarm:**` on an entry with no `**Fire alarm:**` is a deviation.
- **R423:** A field written twice in one entry is a deviation naming the field: the first occurrence
  is read, and every write to that entry refuses with a `DeviationError`.
- **R424:** Every entry reports its 1-based line at parse time; `Unread` is ordered by line and
  lists the non-test headings, each entry carrying a deviation, and every group open at end of input
  or closer that closes nothing.
- **R425:** Every write addresses an entry by alarm number — `Alarm(n)` returns it or nil, and a
  write to an absent number is `ErrNoAlarm` — edits inside that entry's region only, decides its
  refusal before any byte moves, and after the re-read reads its own write back or panics with
  a `ReadBackError`.
- **R426:** `SetPulled(n, date, body)` writes `**Pulled:** <date> — <body>`; over an existing line
  the old content is folded after the body as ` *Earlier —* <old>` so the leading date moves and the
  history is kept; with none, the line is inserted after the `**Inject:**` field's last line, or
  after the `**Fire alarm:**` field's last line when there is no `**Inject:**`.
- **R427:** `SetInject(n, sites, void)` rewrites the `**Inject:**` line with the sites joined
  by `, `; no line is `ErrNoInject` and no sites is `ErrEmptyInject`. With `void` set, an
  existing `**Pulled:**` line is demoted in the same write to *Pulled at `<old sites>` — <old content> — and
  the site has since moved, so this is history rather than a record.*; with `void` unset it stands.
- **R428:** `NumberAlarms` inserts `**Alarm:** <n>` directly above the `**Fire alarm:**` line of
  every alarm entry lacking one, numbering from one above the highest number in the document;
  append-only, idempotent, no number for an entry with no `**Fire alarm:**`, refused whole with
  a `DeviationError` when an unnumbered alarm entry carries deviations, and it returns the numbers
  assigned in document order.
- **R446:** The title is every byte after `Test:` to the end of the heading's line, read from the
  source, so a code span or emphasis in the heading is part of the title as its own bytes.

## Feature: gaps schema
**Source:** specs/gaps-schema.md

- **R429:** `ParseGaps` parses with the markdown base; the section is the region from the level-2
  heading `Gaps` to the next heading of level 2 or higher or the end of the file; `HasGaps` is false
  without one, and a fenced heading opens nothing.
- **R430:** A gap is a bullet inside the region, at any depth, whose head is a type letter
  from `S R D C I O A T`, a number and a colon, with or without a checkbox; depth is the bullet's leading
  whitespace and `Parent` the nearest preceding gap with a smaller depth; a column-0 bullet of any
  other shape is listed in `Unread`; a line inside a code group is body.
- **R431:** A gap's text is its head text with following lines folded on single spaces to the next
  bullet, a blank line, or the region's end; an indented un-keyed bullet beneath it is one of
  its `Sub` lines, folded the same way.
- **R432:** A permanent gap (`A`, `T`) with a checkbox, a tracked gap without one, and an ID the
  section already carries are each a deviation on that entry, listed in `Unread` with the
  rule; `Gap(id)` returns the first entry with that ID; every write to a deviant entry refuses with
  a `DeviationError`; every gap reports its 1-based line at parse time, and `Unread` is ordered by line
  and carries every group open at end of input or closer that closes nothing.
- **R433:** `Add(id, text)` appends one line — `- [ ] <id>: <text>` for a tracked letter, `- <id>: <text>`
  for a permanent one — directly after the last gap's span or after the heading's line when the
  section is empty; an ID not of the shape `X<n>` is `ErrBadGapID`, one already present
  is `ErrGapExists`, no section is `ErrNoSection`; the text is written unwrapped.
- **R434:** `Resolve(id)` turns the head line's `[ ]` into `[x]` and touches nothing else; a permanent
  gap is `ErrPermanent` and a checked one `ErrResolved`.
- **R435:** `Approve(id, newID)` rewrites the head line as `- <newID>: <head text>` at the entry's
  depth with no checkbox and leaves every line beneath it as written; `newID` must be an
  unused `A<n>` (`ErrBadGapID`, `ErrGapExists`) and a permanent target is `ErrPermanent`.
- **R436:** Every write edits inside the region only, decides its refusal before any byte moves, and
  after the re-read reads its own write back or panics with a `ReadBackError`.

## Feature: requirements schema
**Source:** specs/requirements-schema.md

- **R437:** `ParseRequirements` parses with the markdown base; a section is a heading at any level
  with its own content running to the next heading of any level or the end of the file, its `Parent`
  the nearest preceding section with a smaller level; a fenced heading opens nothing; `Section(title)`
  returns every section with that exact title.
- **R438:** A requirement is a column-0 bullet inside a section whose head is `**Rn:**` or the
  retired `**~~Rn:~~**`; its text folds following lines on single spaces to the next bullet, blank line,
  heading or section end; a column-0 bullet of any other shape is listed in `Unread`; a line inside a
  code group is body.
- **R439:** A retired entry's clause `(Retired Tn — see Rm)` or `(Retired Tn — no replacement)` is read
  into `RetiredBy` and `Replacement` and removed from `Text`; a struck head with no clause is a
  deviation.
- **R440:** `**Source:**` at column 0 inside a section is the section's source, the first when there
  are several, with later ones listed in `Unread`.
- **R441:** A repeated ID is a deviation on the later entry and `Requirement(id)` returns the first;
  a deviant entry is listed in `Unread` with its rule and refuses every write with a `DeviationError`;
  every section and entry reports its 1-based line at parse time; `Unread` is ordered by line and
  carries every group open at end of input or closer that closes nothing.
- **R442:** `Add(title, id, text)` appends `- **<id>:** <text>` on one line after the named section's
  last non-blank line of own content, before any blank lines and before its first sub-heading; an ID
  not of the shape `R<n>` is `ErrBadReqID`, one present is `ErrReqExists`, a title no section carries
  is `ErrNoSection`, one several carry is `ErrManySections`.
- **R443:** `Retire(id, tn, clause)` rewrites a live entry's head line
  as `- **~~<id>:~~** (Retired <tn> — <clause>) <head text>` and leaves every continuation line as
  written; `clause` is `see R<m>` or `no replacement` and `tn` is `T<n>`, else `ErrBadClause`; an
  absent ID is `ErrNoRequirement` and a retired one `ErrRetired`.
- **R444:** Every write edits inside one section's own content only, decides its refusal before any
  byte moves, and after the re-read reads its own write back or panics with a `ReadBackError`.

## Feature: links schema
**Source:** specs/links-schema.md
- **R448:** The links reader parses any markdown document with the markdown base; `Links()` returns every inline link, `[text](dest)` and `![alt](dest)`, in document order, each with its text, destination as written, path, fragment, image flag, 1-based line and byte offset at parse time.
- **R449:** A link whose opening bracket sits inside a code group — a fenced block or a code span — is an example, not a reference, and is never listed.
- **R450:** Only the inline form is read: reference-style links, autolinks and bare URLs are not links, and an unclosed `[` or `(` is text.
- **R451:** A destination wrapped in `<…>` is unwrapped and a trailing quoted title is stripped; `Path` and `Fragment` split at the first `#`, and a fragment-only link has an empty `Path`.
- **R452:** `Render` reproduces the source byte for byte, and the property is tested over the real corpus — every markdown document at the repository root, under `carves/` recursively, under `tool/specs/` and `tool/design/`, and under a sibling `ark` checkout when present — with the test reporting the count it read.
- **R453:** `Unread` lists every bracket group open at end of input or closer that closes nothing, ordered by line.
- **R454:** The reader resolves no destination and knows neither the file system nor git; classification is the query's.
- **R462:** `SetDest(i, dest)` replaces link `i`'s destination bytes — between `(` and `)`, title included — with `dest` and nothing else; an index no link carries is `ErrNoLink`; the reader reads the link back at the same index with the new destination or panics with `ReadBackError`.
- **R493:** A pointer is a code span whose content, before any `#`, ends in `.md`: `Pointers()` returns every one in document order with the span as written, its document, its key and its line and offset; a span inside a fenced block, or one not ending in `.md`, is not a pointer.
- **R494:** `SetPointerDoc(i, doc)` replaces pointer `i`'s document bytes — before the `#`, or the whole content when there is no key — with `doc`, keeping the key and the backticks; an index no pointer carries is `ErrNoPointer`; the write is read back like every other.


## Feature: Reference Links View
**Source:** specs/queries.md
- **~~R455:~~** (Retired T11 — see R488) `query links [file...]` reads the named documents, or with none every `*.md` directly in `<repo root>/carves/` and `.carves/` (never `carves/done/`); it resolves at the repository root and needs no design root.
- **R456:** Each link is resolved relative to its citing file's directory with the fragment removed, and classified as exactly one of `tracked`, `untracked`, `ignored`, `missing`, `outside` (above the repository root or absolute), `external` (a URL scheme) or `local` (fragment only).
- **R457:** `ignored`, `missing` and `outside` are errors; `untracked` is a warning; `tracked`, `external` and `local` carry no decision.
- **R458:** The report prints one line per link carrying a decision — citing file, line, the link as written, and its class — `--all` prints every link, and the closing count states every class, zeros included.
- **R459:** A directory target resolves when it exists and is `tracked` when git holds any file under it.
- **R460:** Classification uses the `git` command line only — `ls-files --error-unmatch` and `check-ignore`, batched per citing file — and outside a git working tree the query refuses to classify and says so rather than passing silently.
- **R461:** The exit status is 1 when any link is an error, and `--json` is honoured anywhere among the arguments.
- **R488:** `query links [file...]` reads the named documents, or with none every markdown file git tracks under the repository root — the public documents, a staged file counting as tracked (Bill, 2026-09-15) — resolves at the repository root, needs no design root, and consults no git for the population beyond the index.
- **R492:** A document under a `testdata/` directory is a fixture and never a public document — Go's own convention, applied mechanically — so it is outside every population `PublicDocuments` supplies, however git tracks it.
- **R495:** `query refs [--to <path>] [file...]` lists every link and every pointer in the named files, or in the union population, one line each with the citing file, line, the reference as written, its kind and the repository-relative path it resolves to, or `—`; it passes no judgment.
- **R496:** `--to <path>` lists every reference in the population that resolves to that path; a pointer resolves relative to the citing file's directory except in a trajectory file, where it resolves from the repository root.
- **R497:** The population of `repair-links`, `finished-carve`'s incoming rewrite and `query refs` is the union of every tracked markdown file and every document the trajectory layer sites — the three trajectory files, `carves/`, `.carves/` and each of their `done/` — whether git tracks them or not.


## Feature: Link Repair
**Source:** specs/updates.md
- **~~R463:~~** (Retired T12 — see R489) `update repair-links [file...]` reads the named documents, or with none every live carve and every `*.md` directly under `carves/done/` (and `.carves/`, `.carves/done/`); it resolves at the repository root, needs no design root, and consults no git.
- **R464:** A link is rewritten only when it is `missing` now and exactly one of the four sibling relocations resolves on disk: the citing file re-based from `carves/` to `carves/done/` or back, or the target moved into or out of `done/`; every other class is untouched.
- **R465:** A missing link that no relocation resolves, or that more than one resolves, is reported and left unchanged, each with its reason.
- **R466:** The new destination is the relocated target relative to the citing file's directory, slash-separated, with the fragment kept as written and `<…>` wrapping kept only if it was there; the rewrite goes through `Markdown.SetDest` so every other byte of the document stays.
- **R467:** The report lists every link considered per file — rewritten with old and new destination, or left with its reason — and closes with the counts of rewritten, unresolvable and ambiguous, zeros included; a run that repaired nothing writes no file and says so.
- **R468:** The exit status is 1 when any considered link was left unrepaired, and a second run over repaired documents finds nothing to do and changes no byte.
- **R489:** `update repair-links [file...]` reads the named documents, or with none every markdown file git tracks under the repository root; it resolves at the repository root, needs no design root, and consults no git beyond the index.
- **R500:** A pointer that does not resolve is repaired by the same rule as a link — rewritten only when exactly one sibling relocation across `done/` resolves it, through `SetPointerDoc` with the key kept — and reported as rewritten, unresolvable or ambiguous alongside the links.


## Feature: Finished Carve
**Source:** specs/updates.md
- **R469:** `update finished-carve <carve>` moves a carve that sits directly in `carves/` or `.carves/` to that directory's `done/`, one direction; a path elsewhere, a carve already under `done/`, or a file already at the destination is refused.
- **R470:** The verb refuses a carve with no status block or whose status block has an open part, naming each open part, before any byte moves.
- **R471:** Every link in the carve that resolves now is rewritten to reach the same target from `done/`; a link that does not resolve is reported and left.
- **~~R472:~~** (Retired T13 — see R490) Every link in the incoming population — the live carves, the done carves, and the pending, current and done files at the repository root — that resolves to the carve is rewritten to reach it at its new path, fragment kept.
- **R473:** Every rewrite is computed through `Markdown.SetDest` and read back before any file is written; a read-back failure refuses the whole move and no file changes.
- **R474:** The move is a plain rename with nothing staged: the rewritten carve is written at its new path and the old file removed, and the rewritten incoming documents are written in place.
- **R475:** The report names the move, each rewrite per file with old and new destination, each link left, and the counts of rewritten and left, zeros included; the verb resolves at the repository root and needs no design root.
- **R490:** Every link in the incoming population — every markdown file git tracks under the repository root, the moved carve excluded — that resolves to the carve is rewritten to reach it at its new path, fragment kept; an untracked or ignored document is outside the population, private by the same rule that makes a tracked one public.
- **R498:** `finished-carve` rewrites both kinds of reference to the moved carve — links through `SetDest` and pointers through `SetPointerDoc`, keys kept — so that it rewrites exactly what `query refs --to <carve>` listed before the move.

