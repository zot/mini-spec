# Test Design: Requirements
**Source:** crc-Requirements.md

## Test: the fixture reads as sections and entries
**Purpose:** R437, R438, R439, R440, R441
**Input:** `internal/minispecsdom/testdata/requirements-sample.md`
**Expected:** five sections in order with the `### Notes` parented to its feature; seven entries `R1`–`R7`; the feature's source and line; `R1` folded across two lines; `R3` retired by `T1` with replacement `R7` and its clause removed from the text; `R4` retired with no replacement; `R5` belonging to the sub-heading; the fenced `R99` and `## Feature: quoted` not read; nothing unread; byte-exact render
**Refs:** crc-Requirements.md, seq-requirements.md#1
**Code:** internal/minispecsdom/requirements_test.go
**Alarm:** 1
**Fire alarm:** let a heading of level 3 or deeper neither close the open section nor become the current one, so a feature's content runs to the next `##`. Red: `R5` belongs to the feature rather than the notes, and `Add` to the feature lands after the notes. (A first injection that only skipped the close stayed green: the deep heading still became the current section, so every entry beneath it was still its own — the property has two guards and the injection must remove both.)
**Inject:** internal/minispecsdom/requirements.go:Requirements.scan
**Pulled:** 2026-09-07 — rang, by hand, on the two-guard injection: `R5 does not belong to the sub-heading` and the `Add` placement assertion; the one-guard injection stayed green first; restore byte-clean

## Test: deviations and a second Source
**Purpose:** R438, R439, R440, R441
**Input:** a section with two `Source:` lines, a struck entry with no clause, a repeated `R1`, and a bare bullet
**Expected:** the first source wins and the second is unread; one deviation on the clause-less entry and one on the repeated ID; four unread; `Retire` on the deviant entry refuses with a `DeviationError`
**Refs:** crc-Requirements.md, seq-requirements.md#1.3
**Code:** internal/minispecsdom/requirements_test.go
**Alarm:** 2
**Fire alarm:** accept a struck head without a clause as a plain retirement. Red: `R2` carries no deviation and the write over it succeeds.
**Inject:** internal/minispecsdom/requirements.go:Requirements.newRequirement
**Pulled:** 2026-09-07 — rang, by hand: two deviations where three were expected, and the write over `R2` was refused as `already retired` instead of as a deviation; restore byte-clean

## Test: Add lands at the end of the section's own content
**Purpose:** R442, R444
**Input:** `Add` to the feature with a sub-heading; to an empty section; to the section whose content spans a fence; a taken ID, a bad ID, an unknown title, a repeated title
**Expected:** the new line sits after `R4` and before the blank and `### Notes`, the file longer by that line alone; the empty section's line follows its `Source:`; the fence section's line follows `R7`; `ErrReqExists`, `ErrBadReqID`, `ErrNoSection`, `ErrManySections`
**Refs:** crc-Requirements.md, seq-requirements.md#2.2
**Code:** internal/minispecsdom/requirements_test.go
**Alarm:** 3
**Fire alarm:** insert at the end of the section's own span rather than after its last non-blank line. Red: the new line lands after the blank, directly before `### Notes`.
**Inject:** internal/minispecsdom/requirements.go:Requirements.Add
**Pulled:** 2026-09-07 — rang, by hand: the placement assertions for the feature and the fence section both failed; restore byte-clean

## Test: Retire strikes the head and adds the clause
**Purpose:** R443, R444
**Input:** `Retire("R1", "T3", "see R7")`; the same again; a bad clause; a bad Tn; an absent ID; `Retire("R2", "T4", "no replacement")`
**Expected:** `R1`'s head struck with `(Retired T3 — see R7)` before its text and its continuation line intact, the file longer by exactly the strike and clause; `ErrRetired`; `ErrBadClause` twice; `ErrNoRequirement`; `R2` retired with no replacement and its `(inferred)` text kept
**Refs:** crc-Requirements.md, seq-requirements.md#2.3
**Code:** internal/minispecsdom/requirements_test.go
**Alarm:** 4
**Fire alarm:** write the clause after the head text rather than before it. Red: the retire assertion fails and the read-back does not see the clause.
**Inject:** internal/minispecsdom/requirements.go:Requirements.Retire
**Pulled:** 2026-09-07 — rang, by hand: `Requirements.Retire on "R1" did not read back` — the clause after the text is not the retired form; restore byte-clean
