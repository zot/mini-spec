# Test Design: Markdown
**Source:** crc-Markdown.md

## Test: links are read outside code groups only
**Purpose:** R448, R449
**Input:** a document with a link in prose, an image link, a link inside a code span, one inside a fenced block, and a fenced block that itself contains a heading
**Expected:** two links, in order, with text, dest, image flag, line and offset; the two inside code groups absent; the render byte-exact
**Refs:** crc-Markdown.md, seq-links.md#1.2
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 1
**Fire alarm:** drop the `inCode` test in the scan. Red: four links read, the code-span one at the line the carve's decision quotes.
**Inject:** internal/minispecsdom/mdbase.go:Markdown.scanLinks
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass: `want 2 links, got 4`, the code-span `[text](path)` and the fenced `[fenced](b.md)` both read; restore checksummed clean

## Test: only the inline form, and malformed brackets are text
**Purpose:** R450
**Input:** `[text][ref]`, `<https://x>`, a bare URL, an unclosed `[text`, `[text](unclosed`, and a checkbox `- [ ] [real](a.md)`
**Expected:** exactly one link, `real` → `a.md`
**Refs:** crc-Markdown.md, seq-links.md#1.2.2
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 2
**Fire alarm:** accept `[text][ref]` as a link with dest `ref`. Red: two links.
**Inject:** internal/minispecsdom/mdbase.go:Markdown.scanLinks
**Pulled:** 2026-09-15 — rang, by hand: `[text][ref]` read as a link with dest `ref`, two links where one exists. *First attempt did not distinguish:* an injection that accepted the `[` but still looked for `)` never reached the property on a line with no `)`, so the injection was rewritten to take `]` as the closer after a `[` — the fixture was right, the injection was not; restore checksummed clean

## Test: destination forms
**Purpose:** R451
**Input:** `(<a b.md>)`, `(a.md "title")`, `(a.md#frag)`, `(#frag)`
**Expected:** dest `a b.md`; dest `a.md` with the title gone; path `a.md` fragment `frag`; empty path, fragment `frag`
**Refs:** crc-Markdown.md, seq-links.md#1.2.3
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 3
**Fire alarm:** split path and fragment at the last `#` instead of the first. Red: `a.md#x#y` reads path `a.md#x`.
**Inject:** internal/minispecsdom/mdbase.go:splitDest
**Pulled:** 2026-09-15 — rang, by hand: `"a.md#x#y": got path "a.md#x" frag "y"`; restore checksummed clean

## Test: the render is byte-exact over the real corpus
**Purpose:** R452, R453
**Input:** every `*.md` at the repository root, under `carves/` recursively, `tool/specs/`, `tool/design/`, and under `../ark` when present
**Expected:** every document renders to its own bytes and the node lengths sum to the source length; the count read is logged and is at least 100; every `Unread` line is logged with its file and line (measured 2026-09-15: 955 documents, the unread lines all demoted emphasis — a glob's asterisk — and healthy)
**Refs:** crc-Markdown.md
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 4
**Fire alarm:** retain the source in `parseBase` and make `Render` hand it back rather than the nodes. Red: not the round trip — it stays green, which is the R216 lesson from the first landing; the alarm is the companion check that after a `Split` at a link offset and a `Remove` of the right half, the render shrinks by that node's length.
**Inject:** internal/minispecsdom/mdbase.go:markdownDoc.parseBase, internal/minispecsdom/mdbase.go:markdownDoc.Render
**Pulled:** 2026-09-15 — rang, by hand, sited on `parseBase` retaining the source and `Render` handing it back: the plain round trip stayed green over all 955 documents, as R216 predicted, and the removal check went red on every document holding a link — `render does not follow the nodes after a removal at offset 80976 (264551 bytes, want 264482)` on ark's done file among them; restore checksummed clean

## Test: SetDest rewrites the destination bytes alone
**Purpose:** R462
**Input:** the links fixture; `SetDest(0, "../x/a.md")`, then `SetDest(1, "<i m.png>")`; `SetDest(9, …)`
**Expected:** the first link reads `[one](../x/a.md)` and the second `![pic](<i m.png>)` with dest `i m.png`; the render equals the fixture with exactly those two substitutions and nothing else; `ErrNoLink`
**Refs:** crc-Markdown.md, seq-links.md#3.5
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 5
**Fire alarm:** replace from the `[` rather than from after the `(`, so the text goes with the destination. Red: the read-back panics with a `ReadBackError` — link 0 no longer reads back with the destination written — and the test binary dies on it; a file write through `EditFile` turns the same panic into a refusal.
**Inject:** internal/minispecsdom/mdbase.go:Markdown.SetDest
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the span starting at the `[`: the test binary died on the `ReadBackError` panic from `SetDest` — the link no longer read back with the destination written; restore checksummed clean

## Test: pointers are read outside fences and rewritten with the key kept
**Purpose:** R493, R494
**Input:** a line with `` `carves/x.md#3` ``, `` `specs/a.md` ``, `` `#7` ``, `` `R5` `` and `` `a b.md` ``, then a fenced block quoting `` `carves/fenced.md#1` ``; `SetPointerDoc(0, "carves/done/x.md")`; `SetPointerDoc(9, …)`
**Expected:** two pointers — `carves/x.md` with key `3` and `specs/a.md` with none — at line 1; the fenced one, the queue ID, the requirement and the path with a space are not pointers; after the write the render is the source with exactly that span's document replaced and the key kept; `ErrNoPointer`
**Refs:** crc-Markdown.md, seq-links.md#5.2
**Code:** internal/minispecsdom/mdbase_test.go
**Alarm:** 6
**Fire alarm:** drop the `.md` test so every code span is a pointer. Red: `#7` and `R5` are read as pointers and the count is five.
**Inject:** internal/minispecsdom/mdbase.go:Markdown.scanPointers
**Pulled:** 2026-09-16 — rang, by hand after the simplification pass, with the `.md` test dropped: `#7` and `R5` read as pointers; restore checksummed clean
