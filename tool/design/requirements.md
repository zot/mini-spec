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
- **R25:** Validates requirements.md format and sequential Rn numbering
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
