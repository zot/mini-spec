# Test Design: Harvest
**Source:** crc-Harvest.md

Temporary trees and in-memory sources; no fixtures beyond what each test writes. Traceability
comments in fixtures are built from the language's comment style rather than written as
literals, so no test string reads as a traceability comment to a harvester scanning this file.

## Test: a comment is found with its line and fields
**Purpose:** the happy path, replacing the regex harvester's "found" test
**Input:** a Go file whose third line is a comment `CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5-7`
**Expected:** one comment at line 3 with CRC `crc-Store.md`, Seq `seq-crud.md` step `1.4`, refs
R4, R5, R6, R7 in the `Rn` spelling, and its text the comment as written; a `/* … */` comment
spanning two lines has text on one line, its newline collapsed to a space
**Refs:** seq-harvest.md#2.4 — R513, R502

## Test: a file with no traceability comment yields none, and is not unread
**Purpose:** absence of comments is a finding for validate, not a failure to read
**Input:** a Go file with ordinary comments only
**Expected:** zero comments, and the file is not in the unread list
**Refs:** seq-harvest.md#2.3 — R513

## Test: every comment shape the grammar reads supplies its refs, and nothing else does
**Purpose:** the three measured classes of 2026-09-25, pinned
**Input:** one Go file holding a Seq-only comment with R11, a bare `R12: note`, an `R13. Prose`
comment, `see R14`, `(R15)`, and prose quoting a comment leader followed by R16
**Expected:** refs R11, R12, R13 and nothing else
**Code:** internal/parser/harvest_test.go
**Alarm:** 1
**Fire alarm:** drop the full-stop candidate from `descSep`, so `R13. Prose` is prose again, and
confirm the refs come back as R11 and R12 only
**Inject:** internal/minispecsdom/comment.go:descSep
**Pulled:** 2026-09-29 — rang: refs [R11 R12]; restore byte-clean
**Refs:** seq-harvest.md#2.4 — R508, R514

## Test: an extension with no table is unread, with its reason
**Purpose:** a file the harvest cannot read must never read as clean
**Input:** a manifest listing `x.zig`
**Expected:** `x.zig` unread, line 0, reason naming `.zig`; no comments
**Refs:** seq-harvest.md#2.1 — R515

## Test: an unclosed string keeps what was read and reports where
**Purpose:** a swallowing group is visible, and what came before it still counts
**Input:** a Go file with a traceability comment on line 1 and a string opened on line 3 and
never closed, with a traceability comment after it
**Expected:** the line-1 comment is harvested, the later one is not, and the file is unread at
line 3
**Code:** internal/parser/harvest_test.go
**Alarm:** 2
**Fire alarm:** skip step 2.5, so a file whose string runs to end of input returns no unread
entry, and confirm the test sees nil where line 3 was expected
**Inject:** internal/parser/harvest.go:HarvestFile
**Pulled:** 2026-09-29 — rang: "unread <nil>, want line 3"; restore byte-clean
**Refs:** seq-harvest.md#2.5 — R515

## Test: stray closers and open code brackets are not reported
**Purpose:** they hide no comment, and HTML page text is full of them
**Input:** an HTML page whose text holds `(see note)` and a stray `}`, and a Go file ending
inside an unclosed `{` with a traceability comment inside it
**Expected:** neither file is unread, and the Go comment is harvested
**Code:** internal/parser/harvest_test.go
**Alarm:** 3
**Fire alarm:** report the first unclosed group whatever its kind, dropping the restricted test,
and confirm the Go file ending inside `{` is reported unread
**Inject:** internal/parser/harvest.go:HarvestFile
**Pulled:** 2026-09-29 — rang: the open `{` reported unread at line 2; restore byte-clean
**Refs:** seq-harvest.md#2.5 — R515

## Test: a configured language overrides the built-in table
**Purpose:** a project's definition wins for the extensions it names
**Input:** a definition for `.go` whose only comment form is `##`, and a `.go` file holding a
`##` traceability comment and a `//` one
**Expected:** the `##` comment is harvested and the `//` one is not
**Code:** internal/parser/harvest_test.go
**Alarm:** 4
**Fire alarm:** consult the built-in map before the configured one in `LanguageFor`, and confirm
the `//` comment is read instead of the `##` one
**Inject:** internal/minispecsdom/langs.go:LanguageFor
**Pulled:** 2026-09-29 — rang: refs [R2], Go's table read instead; restore byte-clean
**Refs:** seq-harvest.md#2.1 — R527

## Test: the manifest is read in order, each file once, missing files skipped
**Purpose:** a file listed under two artifacts is harvested once; a missing file is validate's
**Input:** a manifest listing `a.go`, `b.go`, `a.go` and `gone.go`, with `gone.go` absent
**Expected:** harvested files `a.go`, `b.go` in that order; `gone.go` neither harvested nor unread
**Code:** internal/parser/harvest_test.go
**Alarm:** 5
**Fire alarm:** drop the `seen` check in `HarvestArtifacts`, and confirm `a.go` is harvested twice
**Inject:** internal/parser/harvest.go:HarvestArtifacts
**Pulled:** 2026-09-29 — rang: files [a.go b.go a.go]; restore byte-clean
**Refs:** seq-harvest.md#1.1, seq-harvest.md#1.2 — R513

## Test: another language's comment form is read by its table
**Purpose:** replaces the regex harvester's configurable-pattern test: the table, not a pattern,
decides the form
**Input:** a Python file with a `#` traceability comment
**Expected:** the comment is harvested with its CRC and Seq fields
**Refs:** seq-harvest.md#2.1 — R509
