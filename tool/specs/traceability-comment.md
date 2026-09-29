# The traceability comment

Package `minispecsdom`, a sibling of `sdom` and the first code that knows what a
CRC card is. Mini-spec's anchor between code and design is a comment:

    // CRC: crc-Store.md | Seq: seq-crud.md#1.4 | Test: test-Store.md | R4, R5-7 -- note

`TraceabilityComment` is a stenciled node over the **whole** comment, opener through
closer, and it is **one kind for every language** — only the wrapping is
language-specific, through `BracketLang.Comment`. A consumer skims the flat array
for the kind and has the complete comment as one node.

## The grammar

The interior is the bytes between the comment's opener and closer.

```
interior ::= WS? field ( WS? "|" WS? field )* ( WS? descsep description )?
field    ::= "CRC:"  WS? list  |  "Seq:" WS? list  |  "Test:" WS? list  |  reqlist
descsep  ::= "--" | "—" | ":"      accepted on read;  "--" written
           | "."                  accepted on read, directly after a requirement list only
```

- Four field kinds, **at most one of each, in any order**; a reader classifies each
  `|`-segment by its lead. `CRC:`, `Seq:` and `Test:` carry a plain list
  (simple-dom’s `lists.md`); the keyword-less segment is the requirement list.
- A `:` is a field-key colon only immediately after `CRC`, `Seq` or `Test`;
  anywhere else it is the description separator, so `R5: desc` reads `R5` then a
  description.
- A `.` directly after the requirement list is also a description separator, so
  `R271. A gap is a source` reads `R271` then a description. Anywhere else a `.` is
  ordinary text: file names in `CRC:`, `Seq:` and `Test:` lists are full of them.
- A `Seq` item may carry `#step`; the typed view splits path from step.
- **Whitespace is glue everywhere.** `//CRC:x|R7` and `// CRC:  x  |  R7` both parse
  and both render back byte-exact. The keywords, `|` and the separator are computed
  glue, so the pattern cannot silently eat bytes.
- The description is bound and writable. Its separator is preserved byte-exact when
  unedited — an existing `—` or `:` stays — and written as `--` on a fresh write.

**Recognition is a parse that consumes the whole interior.** `// Test: a repaint
frame round-trips (R3136).` leads with a keyword and leaves prose uncovered, so it
is not a traceability comment; nor are `// see R5` or `// (R5)`.

## The node

```go
type TraceabilityComment struct {
    sdom.Compound
    // typed views of the children; nil when the field is absent
}

func (c *TraceabilityComment) CRC()  *sdom.List
func (c *TraceabilityComment) Seq()  *sdom.List
func (c *TraceabilityComment) Test() *sdom.List
func (c *TraceabilityComment) Refs() *sdom.RequirementList
func (c *TraceabilityComment) Description() *sdom.Text   // nil when absent

func (c *TraceabilityComment) Parse(cmt *sdom.Opener, ctx *sdom.BracketContext) bool
```

`Parse` reads the comment that `cmt` opens through the context — its closer, its
inner text — and fills the node **off to the side**, touching no document. It returns
false when the interior is not a single text node or the walk does not consume it;
the caller discards the node and moves on. On true the node's children are the
original `*Opener`, the interior's glue and fields, and the original `*Closer` —
**reused, not recreated**, so the context's identity-keyed links stay valid after the
splice. Order-independence rules out one regex for the interior, so `Parse` walks
`|`-segments, one stencil per segment, and splices the results flat: the segment
walk is a parsing device, not tree structure.

## The pass

```go
func Comments(d *sdom.Doc, ctx *sdom.BracketContext) ([]*TraceabilityComment, error)
```

The second pass over a bracket-parsed document: every opener whose group's kind is
the language's `Comment.Kind` is a candidate; each gets a `Parse`, and each success
is spliced in — the run from opener to closer replaced by the one node — inside its
own mutation window. Anything else in the document is untouched. Associating a
comment with the declaration it sits above is the consumer's job.

## Languages

A code file is read through the bracket table its **extension** names:

| Extensions | Table | From |
|---|---|---|
| `.go` | `LangGo` | sdom |
| `.js`, `.ts` | `LangJavaScript` | sdom |
| `.lua` | `LangLua` | sdom |
| `.sh`, `.bash` | `LangShell` | sdom |
| `.py` | `LangPython`'s bracket table | sdom |
| `.pas`, `.dpr` | `LangPascal` | sdom |
| `.c`, `.h` | `LangC` | here |
| `.cpp`, `.hpp`, `.cc` | `LangCPP` | here |
| `.java` | `LangJava` | here |
| `.el` | `LangElisp` | here |
| `.html` | `LangHTML` | here |
| `.md` | `LangMarkdown` | here |
| `.css` | `LangCSS` | here |

sdom ships the tables for the languages it was built against, and says a consumer
needing another constructs its own. The tables built here:

- **`LangC`** is `LangGo` without the backtick raw string, which C does not have.
- **`LangCPP`** is `LangC` with raw strings, `R"delim( … )delim"` with an optional `u8`,
  `u`, `U` or `L` prefix: an opener pattern and a closer pattern that name the same capture
  group, so the string closes only where the delimiter matches the one it opened with.
- **`LangJava`** is `LangC` with text blocks, `"""` to `"""`, listed ahead of `"` so a text
  block is not read as an empty string followed by another.
- **`LangElisp`** recognizes `;` line comments, `"` strings with `\` escapes, `( )` and
  `[ ]`, and character literals: a `?` opening only at the start of a token, closed by the
  one character after it — escaped or with modifiers, as in `?\(` and `?\C-x`. Without
  them `?(` reads as an unbalanced bracket; with the token-start test, the `?` ending a
  name like `f-exists?` stays part of the name.
- **`LangMarkdown`** recognizes one group, `<!--` to `-->`, raw inside. Nothing else in a
  markdown file is code.
- **`LangCSS`** recognizes `/*` to `*/`, raw inside, and both quoted strings, so a `/*`
  inside `content: "…"` is text.
- **`LangHTML`** reads a page and the script and style inside it. `<!--` to `-->` is raw
  and recognized everywhere. `<script` to `</script>` is code mode, and every JavaScript
  group lists `<script` and the JavaScript brackets (`{`, `(`, `[`, `${`) as its allowed
  parents, since a group's allowed parent is checked against the group immediately
  enclosing it: a `//` inside a function body sits under `{`. `<style` to `</style>` is
  restricted to `/*` and the two quotes; a stylesheet's comments are all that is read from
  it, so its braces stay text. Outside a script nothing JavaScript is live, so the
  apostrophe in *don't* opens no string and the `//` in a URL opens no comment, in page
  text or in a stylesheet's `url(//…)`.

**The form to write is the table's `Comment` style**, which every table sets: `// ` for
Go, `<!-- ` … ` -->` for HTML and Markdown, `/* ` … ` */` for CSS. Where a language has
several comment forms, the reader accepts them all and the writer uses that one. The
order of a table's brackets is not a preference: it belongs to matching, where a longer
marker must precede its prefix, so Lua lists `--[[` before the `--` it writes.

A project defines further languages in its configuration, or replaces a built-in one for
the extensions it names (see [config.md](config.md)). A file whose extension has no table,
built in or configured, is **not read, and says so** (see below).

## The harvest

The harvest is the one reader of code files' traceability, and `validate`, `query
traceability` and `query implementation` all consume it. For every code file the
`design.md` Artifacts manifest lists, it parses the file through its table, runs the pass,
and keeps each comment with its file and line: its `CRC` and `Seq` items, and its
requirement list **expanded**, so `R5-R8` yields R5, R6, R7 and R8.

- **Every traceability comment counts, whatever fields it carries.** A comment led by
  `CRC:`, one with only `Seq:` and refs, and a bare `// R5: note` all implement their refs.
  What does not count is what the grammar does not read: `// see R5`, `// (R5)`, and a
  comment that leads with a field but leaves prose uncovered.
- **Nothing is skipped silently.** A file whose extension has no table is listed unread,
  and so is a file whose parse leaves a **string or comment** open at end of input, at the
  line it opened: inside one, nothing is recognized, so everything after it went unsearched
  and a coverage answer that omitted it would read clean over code it never saw. A stray
  closer or an unclosed code bracket is not reported, because it hides nothing: comments
  are recognized inside code brackets, and page text in HTML is full of unmatched `)`.

## Construction

```go
type Fields struct {
    CRC, Seq, Test []string
    Refs           []int
    Description    string
}

func New(lang *sdom.BracketLang, f Fields) *TraceabilityComment
```

`New` assembles the canonical interior — `CRC`, `Seq`, `Test`, refs, in that order,
single spaces, `--` before a description — and runs **the same interior walk** over it
that `Parse` runs, at a synthetic location, wrapped in a synthetic opener and closer
from `lang.Comment`. There is no hand-built child list, so nothing can construct a
comment that disagrees with how one is read. The node has no origin, which is what
lets a consumer insert it into any document: a node from a separate parse carries
that parse's origin and can enter no other document.

## Tests

- A **fuzzed source string** over the whole grammar — any field order, whitespace,
  all three separators, ranged refs — asserting `render(dom) == src` and
  `parse(render(dom)).Equals(dom)`. The DOM compare fails on under-modelling,
  which a byte round-trip cannot see. Lua's opener is `--`, the same token as the
  written separator, and is a case the generator must reach.
- `New` produces a node that `Parse` reads back `Equals`.
- Every shipped language's `Comment` constructs a comment that parses back as its
  `Kind`.
