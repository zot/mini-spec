# Test Design: TraceabilityComment
**Source:** crc-TraceabilityComment.md

## Test: the fuzzed DOM round-trip
**Purpose:** R337, R338, R339, R341, R342, R345 — the whole grammar, and under-modelling caught where a byte round-trip cannot see it
**Input:** Go native fuzzing over a seeded corpus: every field order, `//CRC:x|R7` with no spaces, wide spacing, all three description separators, ranged refs, `Seq` steps, and Lua's `-- CRC: x | R4 -- desc`; each seed wrapped in each shipped language's `Comment`
**Expected:** for every input that `Comments` recognizes, `render(doc) == src`, the node carries at least one field, and `Comments(parse(render)).Equals(node)`
**Refs:** crc-TraceabilityComment.md, seq-anchor.md#2
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 1
**Fire alarm:** make the segment walk keep the whole interior as one glue `Text` and set no fields (in `Parse`, replace the `c.walk` call with a single `NewText` over the interior). Bytes round-trip perfectly, and **the DOM compare stays green** — both sides run the same parser, so symmetric under-modelling is invisible to it. Red comes from the *recognized with no field* assertion. (The first pull believed the DOM compare would object; it did not, and the red that day was a nil dereference in `TestFieldsReadBack` — plumbing, not an assertion — which is why this assertion exists.)
**Inject:** internal/minispecsdom/comment.go:TraceabilityComment.Parse
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang, on the named test run alone (`-run FuzzDOMRoundTrip`): every seed `recognized with no field`; over the whole package the nil dereference in `TestFieldsReadBack` ends the run first, as the record foresaw; restore clean. *Earlier —* 2026-09-03 — rang, on the second attempt. The delegated pull went red only by a nil dereference in `TestFieldsReadBack`, with the fuzz test green; the no-field assertion was added and the alarm re-pulled by hand the same day: `go: "// CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5\n" recognized with no field` on the first seed.
## Test: recognition is consumption
**Purpose:** R342, R343, R346 — leading with a keyword is not enough, and nothing else is touched
**Input:** a Go file with `// CRC: crc-A.md | R1`, `// Test: a repaint frame round-trips (R3136).`, `// see R5`, `// (R5)`, and `x := 1 // R7`
**Expected:** `Comments` returns two nodes — the first and the trailing `// R7`; the other three remain ordinary groups; `Parse` on the `Test:` comment returns false and the document is unchanged by the attempt
**Refs:** crc-TraceabilityComment.md, seq-anchor.md#1
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 2
**Fire alarm:** return true whenever a segment matched, ignoring the remainder — the `Test:` prose comment is recognized and the count goes to three.
**Inject:** internal/minispecsdom/comment.go:TraceabilityComment.walk
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang on the named test, through simple-dom's stencil guard rather than the count: `the node put for group "test" spans 30+3, but the group is 30+50`; restore clean. *Earlier —* 2026-09-03 — rang: `4 comments recognized, want 2` — not three as predicted, since `see R5` and `(R5)` both become glue-only comments under the injection. Site is `walk`, the helper `Parse` calls.
## Test: the splice reuses the markers and the context survives it
**Purpose:** R344, R346
**Input:** `func f() {}\n// CRC: crc-A.md | R1\nfunc g() {}\n` under Go; take the comment opener and closer before the pass
**Expected:** after `Comments`, the node's first and last children are those very pointers; the flat array holds one node where three were; the context still pairs the two `{ }` groups around it, since the splice bumped the generation and the next read rebuilt cleanly
**Refs:** crc-TraceabilityComment.md, seq-anchor.md#1.2.2
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 3
**Fire alarm:** build fresh `NewOpener` / `NewCloser` from the originals' text — `Equals` still holds and the identity check fails.
**Inject:** internal/minispecsdom/comment.go:TraceabilityComment.assemble
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `the markers were recreated rather than reused`; restore clean. *Earlier —* 2026-09-03 — rang: `the markers were recreated rather than reused`, only that test. Site is `assemble`.
## Test: New parses back Equals, in every language
**Purpose:** R347, R348 (with simple-dom’s R208) — construction agrees with reading, and the comment style's kind holds per language
**Input:** `New(lang, Fields{CRC: [crc-A.md], Seq: [seq-b.md#1.2], Refs: [4 5 6 9], Description: "note"})` for every shipped `BracketLang` with a non-empty `Comment.Prefix`
**Expected:** the render is `<Prefix>CRC: crc-A.md | Seq: seq-b.md#1.2 | R4-6, R9 -- note<Suffix>`; parsing that source with the language and running `Comments` yields one node `Equals` to the constructed one; the constructed node's location has no origin
**Refs:** crc-TraceabilityComment.md, crc-BracketLang.md
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 4
**Fire alarm:** write refs before `Seq` in `New` — the render changes, and the parse-back still `Equals`, which is why the render is asserted literally.
**Inject:** internal/minispecsdom/comment.go:New
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: the render with refs before `Seq` failed the literal assertion in all six languages; restore clean. *Earlier —* 2026-09-03 — rang in all six languages on the literal render, `CRC … | R4-6, R9 | Seq …`, while the parse-back stayed `Equals` — exactly why the render is asserted literally.
## Test: a field write through the node
**Purpose:** R341 (with simple-dom’s R201, R203) — the write reaches one literal and the descsep is preserved
**Input:** parse `// R5: desc` and `// CRC: a.md — note`; `Refs().SetItems([5 6 7])` on the first
**Expected:** the first renders `// R5-7: desc` — the `:` separator untouched; the second's `Description` reads ` note` and renders back with the em dash
**Refs:** crc-TraceabilityComment.md, seq-anchor.md#3
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 5
**Fire alarm:** normalise the separator to `--` on read — the second renders `-- note`, the byte round-trip fails.
**Inject:** internal/minispecsdom/comment.go:TraceabilityComment.walk
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `-- note` where the em dash was, and the fuzz round-trip seeds failed with it; restore clean. *Earlier —* 2026-09-03 — rang: `the em dash was not preserved: "// CRC: a.md -- note"`, `after the write: "// R5-7-- desc"`, and two fuzz seeds on the byte round-trip. Site is `walk`.
## Test: only comment groups are candidates
**Purpose:** R346 — found by injecting past the alarm list, not by design
**Input:** Go source with `f(R7)`, a string `"CRC: crc-A.md | R1"`, and `[R4, R5]` — three non-comment groups whose interiors parse as fields
**Expected:** `Comments` returns nothing and the node count is unchanged
**Refs:** crc-TraceabilityComment.md, seq-anchor.md#1.1
**Code:** internal/minispecsdom/comment_test.go
**Alarm:** 6
**Fire alarm:** drop the `g.Kind == kind` half of the candidate filter in `Comments`, so every opener is tried. Red: three comments recognized in a file with none. Every other test stayed green under this injection when it was tried on 2026-09-03, because no fixture had a bracket interior shaped like a field.
**Inject:** internal/minispecsdom/comment.go:Comments
**Pulled:** 2026-09-14 — re-pulled by delegation at `9c6c796`'s tree after the package moved from mini-spec-tool (the census read every ported alarm stale, the files being new to git); rang: `3 comments recognized in a file with none`; restore clean. *Earlier —* 2026-09-03 — rang: `3 comments recognized in a file with none`, only this test.
