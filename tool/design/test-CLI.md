# Test Design: CLI
**Source:** crc-CLI.md

## Test: minted values are reported, not returned
**Purpose:** validates R331 — `retire`, `migration-complete` and `add-req` print a sentence naming the minted value, `--quiet` hides it, `--json` carries it
**Input:** a design root with two requirements and an empty Gaps section; `retire R1 R2`, `add-req` quiet, `add-req --json` with two texts, `add-req` plain
**Expected:** `Retired R1 as T1 (see R2)`; nothing; JSON carrying `"range": "R4-R5"`; `Added R6 to "A"`
**Refs:** crc-CLI.md — R331
**Code:** internal/cli/cli_minted_test.go
**Alarm:** 1
**Fire alarm:** print the bare `Tn` again — `fmt.Println(tn)` in place of the sentence — and confirm the retire assertion goes red with `T1` alone on stdout
**Inject:** internal/cli/cli.go:runUpdate
**Pulled:** 2026-09-15 — rang again after `runUpdate` gained the `repair-links` dispatch for `#84`, same injection: `TestMintedValuesAreReportedNotReturned` red with the bare `T1` on stdout; restore checksummed clean *Earlier —* 2026-09-13 — rang: `retire stdout = "T1\n", want the sentence`; restore byte-clean by copy
