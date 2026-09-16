# The links schema

The reader over **any markdown document the project owns** for the one thing the other
schemas do not read: its references. A reference is a real markdown link, `[text](dest)`
or `![alt](dest)` (Bill, 2026-08-04: a prose mention is not a reference, because it cannot be
checked). It embeds the markdown base and owns the document's DOM; it writes nothing.

```go
type Markdown struct { /* the document and its context */ }

func ParseMarkdown(src string) *Markdown
func (m *Markdown) Doc() *sdom.Doc
func (m *Markdown) Render() (string, error)
func (m *Markdown) Links() []Link             // every link, in document order
func (m *Markdown) Unread() []Unread          // groups never closed or closing nothing
func (m *Markdown) SetDest(i int, dest string) error   // rewrite link i's destination bytes alone
func (m *Markdown) Pointers() []Pointer                 // every `doc.md` / `doc.md#key` code span, in document order
func (m *Markdown) SetPointerDoc(i int, doc string) error // rewrite pointer i's document bytes alone, key kept

type Pointer struct {
    Raw  string   // the span as written, backticks included
    Doc  string   // the bytes before `#`, ending in .md
    Key  string   // the bytes after `#`, "" when none
}
func (p Pointer) Line() int
func (p Pointer) Offset() int

type Link struct {
    Raw      string   // the link as written, `[` (or `!`) through `)`
    Text     string   // the bytes between the brackets, as written
    Dest     string   // the destination as written, title stripped
    Path     string   // Dest before any `#`; "" for a fragment-only link
    Fragment string   // the bytes after `#`, "" when none
    Image    bool     // opened by `!`
}
func (l Link) Line() int                      // 1-based, at parse time
func (l Link) Offset() int                    // byte offset of the opening `[` or `!`
```

## What it reads

**Links are not nodes of the base.** simple-dom's markdown lexicon opens no group on `[`,
because a line-head marker may share no first byte with an opener, so a link is read by
scanning the source after the parse rather than by the parser — the source, not each `Text`
run, because link text may hold emphasis whose markers are nodes of their own. What the base
supplies is the part a regex cannot: **which bytes sit inside a code group.** A link inside
a fenced block or a code span is an example, not a reference, and is never listed —
`carves/reference-discipline.md` quotes ` [text](path) ` in a code span on the very line
that decides the rule, and a grep flagged it as a missing file on 2026-08-04.

**The inline form only.** `[text](dest)` and `![alt](dest)`, where the text may hold
balanced brackets one level deep (a checkbox `[ ]` inside link text is legal and rare) and
the destination runs to the closing parenthesis. A destination wrapped in `<…>` is unwrapped;
a trailing quoted title is stripped. Reference-style links (`[text][ref]`), autolinks
(`<https://…>`) and bare URLs are not read: measured 2026-08-16 across 425 mini-spec and ark
documents, reference links scored zero, and a form nobody writes is a form nobody checks.
An unclosed `[` or `(` is text.

**Every link is reported where it stands**, with its line and byte offset as the document was
parsed, so a report can name the citing line and a later reader can bind it.

**`Path` and `Fragment` are split at the first `#`.** A fragment-only link (`#section`)
points into its own document and has an empty `Path`; a link with a path keeps its fragment
for the reader's caller to check or ignore.

`Unread` lists every bracket group open at end of input or closer that closes nothing, as
every reader does, because a fence never closed swallows every link after it and nothing
else would say so.

## Pointers

**A pointer is a code span whose content is a markdown path, optionally followed by `#` and
a key**: `` `carves/x.md#3` ``, `` `specs/index.md` ``. It is the tool's own reference form —
the ledgers name a part that way and a carve names a document it does not link — and no
link machinery reads it, which is how a move left every done entry naming a carve's old
path. A code span is a pointer only when the bytes before any `#` end in `.md`: `` `#7` ``
is a queue ID and `` `R5` `` a requirement. Spans inside a fenced block are examples and are
not read. Every pointer reports its line and offset, like a link.

## What it writes

**`SetDest(i, dest)`** replaces the destination bytes of link `i` — the bytes between `(` and
`)`, the title included — with `dest`, and nothing else: the text, the parentheses and every
byte around the link stay where they were. It is the reader's one write and exists for the
move repair ([updates.md](updates.md), `update repair-links`). An index out of range is
`ErrNoLink`. After the write the reader re-reads its render and reads the link back at the
same index with the new destination, or panics with `ReadBackError`, as every reader does.
Byte offsets of the parse are what `Links()` reports, so several links may be rewritten in
one document as long as each is written once.

**`SetPointerDoc(i, doc)`** replaces the document bytes of pointer `i` — before the `#`, or
the whole content when there is no key — with `doc`, the key and the backticks untouched;
`ErrNoPointer` for an index none carries; read back like every write.

## What it does not do

**No judgment.** The reader does not resolve a destination, does not know the file system,
and does not know git. Where a link points and whether that is allowed is the classifier's
question ([queries.md](queries.md), `query links`), and later the document-class model's
(`carves/reference-discipline.md` Item 2).

## Round trip

`Render` reproduces the source byte for byte, and the property is checked **over the real
corpus rather than fixtures**: every markdown document under this repository's `carves/`,
`tool/specs/`, `tool/design/` and the repository root, and every one under a sibling `ark`
checkout when one is present. simple-dom's own round-trip test covers its own tree only; the
documents this tool is trusted with are these, and a fixture holds only what its author
thought to include. The test reports the count it read, so an empty population is visible as
a wrong number rather than a green pass over nothing.
