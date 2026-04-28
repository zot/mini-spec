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

## Feature: Updates
**Source:** specs/updates.md

- **R18:** `update check [file] [item]` checks a checkbox
- **R19:** `update uncheck [file] [item]` unchecks a checkbox
- **R20:** `update add-ref [crc-file] [Rn]` adds requirement to CRC card
- **R21:** `update remove-ref [crc-file] [Rn]` removes requirement from CRC card
- **R22:** `update add-gap [type] [desc]` adds new gap with auto-numbered ID
- **R23:** `update resolve-gap [id]` marks gap as resolved (checks checkbox)

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

- **R32:** Project detection walks up to find design/ directory
- **R33:** Default paths: design/, src/, crc-*.md, seq-*.md
- **R34:** Optional .minispec.yaml config file for overrides
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
- **R56:** `check-version` compares tool version against skill README.md Version: line (project-level then user-level), exits 0 on match, 1 on mismatch or not found

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

## Feature: Migration Workflow
**Source:** specs/migrations/complete/001-migration-and-retirement.md

- **R73:** Tool parses Gaps section type "T" (retired) in addition to S/R/D/C/I/O/A
- **R74:** Tn entries are written and parsed without a leading checkbox (`- T1: ...`, no `[ ]` or `[x]`)
- **R75:** An (approved) entries are written without a leading checkbox; the parser still accepts the legacy `- [ ] An: ...` form for back-compat
- **R76:** Validate reports A-typed and T-typed gap lines that carry a checkbox marker as a "permanent gaps with checkbox" issue so AIs clean them up
- **R77:** Requirements parser accepts the strikethrough retired form `- **~~Rn:~~** (Retired Tk — see Rxxx) <text>` (or `... no replacement) <text>`) and exposes it via a `Retired` flag on the parsed Requirement
- **R78:** Retired requirements are excluded from coverage uncovered and implementation-coverage uncovered lists; their Rn IDs remain valid for cross-reference resolution
- **R79:** `query migrations` lists in-flight migration spec files (specs/migrations/*.md, non-recursive); empty output and exit 0 when none exist
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
