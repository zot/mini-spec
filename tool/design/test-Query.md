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

## Test: the arguments classify into number mode or text mode
**Purpose:** validates R503 — a clean list of requirement refs is number mode, IDs ascending and deduplicated whatever order they were given in; anything else is one regexp over requirement text; a gap ID that is not an `R` is text, not number
**Input:** `R8,R5-6` `R5`; `@status`; `D3`; `foo bar`; `(`
**Expected:** number `R5 R6 R8`; text `@status`; text `D3`; a refusal (two args, not refs); a refusal (the pattern does not compile)
**Refs:** crc-Query.md — R503
**Code:** internal/query/implementation_test.go
**Alarm:** 3
**Fire alarm:** drop the every-ID-is-an-R check in `ClassifyImplArgs`, so `D3` reads as number mode — a lookup for a gap that silently answers nothing where the caller meant a text search
**Inject:** internal/query/implementation.go:ClassifyImplArgs
**Pulled:** 2026-09-29 — rang: `D3: want text mode, got {IDs:[D3] Pattern:<nil> …}`; restore byte-clean by copy

## Test: each selected requirement is answered, in ascending order, from the harvest
**Purpose:** validates R502, R504 and R506 — a range in a comment answers each member; locations come in manifest then line order with their file, line and comment; requirements come ascending; a selected requirement with no implementing comment is kept with no locations rather than dropped; number mode carries no requirement text and text mode does
**Input:** a harvest of `a.go` (line 3 `// R5-R7`) and `b.go` (line 9 `// CRC: x.md | R6`); requirements in document order R8, R6, R7 (retired), R5 — not numeric, as in a real `requirements.md`; number `R8,R6`; text `.`
**Expected:** number `R6 → a.go:3, b.go:9` then `R8 → none`, no text; text over `.` gives R5, R6, R8 with text, R8 with no locations
**Refs:** crc-Query.md — R502, R504, R506
**Code:** internal/query/implementation_test.go
**Alarm:** 4
**Fire alarm:** skip a requirement with no locations in `SelectImplementation` — R8 vanishes from the answer, and the output reads exactly like *every match is implemented*, which is the absence the spec says must be stated
**Inject:** internal/query/implementation.go:SelectImplementation
**Pulled:** 2026-09-29 — rang: `text mode: got "R5[five] a.go:3; R6[six] a.go:3 b.go:9"` — R8 dropped; restore byte-clean by copy

## Test: text mode answers ascending when the document is not
**Purpose:** validates R504 — `requirements.md` holds requirements in feature order, not numeric order, and text mode reorders them ascending
**Input:** the fixture above, whose requirements are in document order R8, R6, R7, R5; text `.`
**Expected:** R5, R6, R8
**Refs:** crc-Query.md — R504
**Code:** internal/query/implementation_test.go
**Alarm:** 6
**Fire alarm:** drop the ascending sort in text mode in `SelectImplementation` — entries come in `requirements.md` document order, which reads as ordered and is not; silent on a fixture already ascending, which is how it was found
**Inject:** internal/query/implementation.go:SelectImplementation
**Pulled:** 2026-09-29 — rang: text mode came back `R8; R6; R5` in document order, three assertions red; restore byte-clean by copy

## Test: retired requirements answer by number, and by text only with --retired
**Purpose:** validates R505 — a retired requirement selected by number is answered with its locations; text mode leaves it out unless the flag is set; the flag changes nothing in number mode
**Input:** the harvest and requirements above; number `R7` with and without the flag; text `.` with and without the flag
**Expected:** number `R7 → a.go:3` both ways; text without the flag omits R7, with it includes R7 retired
**Refs:** crc-Query.md — R505
**Code:** internal/query/implementation_test.go
**Alarm:** 5
**Fire alarm:** drop the retired filter in text mode in `SelectImplementation` — R7 appears in a survey of live intent, looking like a requirement that still holds
**Inject:** internal/query/implementation.go:SelectImplementation
**Pulled:** 2026-09-29 — rang: `text mode: got "R5[five] a.go:3; R7[seven](retired) a.go:3" — a retired requirement surveyed as live intent`; restore byte-clean by copy
