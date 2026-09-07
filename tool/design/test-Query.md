# Test Design: Query
**Source:** crc-Query.md

## Test: the RANGE grammar and the letter as namespace
**Purpose:** validates R317 and R318 — bare IDs, ranges with the second letter optional, comma lists and mixtures expand in document-independent order; a reversed range is its low end; a range across two types is refused naming both; a non-ID is refused
**Input:** `O22-O24,A1`, `R5-7`, `O28-O22`, `O22-R5`, `gap22`
**Expected:** `O22 O23 O24 A1 R5 R6 R7`; `O28`; a refusal naming O and R; a refusal
**Refs:** crc-Query.md — R317, R318
**Code:** internal/query/gaps_test.go
**Alarm:** 1
**Fire alarm:** drop the cross-type check in `ExpandGapRefs` and confirm `O22-R5` expands
**Inject:** internal/query/gaps.go:ExpandGapRefs
**Pulled:** 2026-09-07 — rang: `a range across types was not refused: <nil>`; restore byte-clean by copy

## Test: nothing matched is an error, partly unassigned is not
**Purpose:** validates R319 and R320 — an unmatched selection refuses naming what it looked for; a range with unassigned members selects what exists, in document order
**Input:** `O99`; `O22-O26` over a pool holding O22, O23, O26
**Expected:** a refusal naming O99; `O22 O23 O26`
**Refs:** crc-Query.md — R319, R320
**Code:** internal/query/gaps_test.go

## Test: the state flags never claim a permanent gap
**Purpose:** validates R321 and R322 — `--open` and `--closed` select by checkbox and skip `A` and `T`; both flags mean every checkbox; the zero selection selects everything; a range and a flag compose
**Input:** a pool of three tracked O gaps (one resolved), A1, T2, and a tracked R5
**Expected:** open `O22 O26 R5`; closed `O23`; both `O22 O23 O26 R5`; all six; `O22,O23 --closed` gives `O23`
**Refs:** crc-Query.md — R321, R322
**Code:** internal/query/gaps_test.go
**Alarm:** 2
**Fire alarm:** drop the `HasCheckbox` guard in `SelectGaps` so `!Resolved` alone decides — confirm A1 and T2 appear under `--open`, reading exactly like work to do
**Inject:** internal/query/gaps.go:SelectGaps
**Pulled:** 2026-09-07 — rang: `--open: got "O22 O26 A1 T2 R5", want O22 O26 R5 — a permanent gap read as work to do`; restore byte-clean by copy
