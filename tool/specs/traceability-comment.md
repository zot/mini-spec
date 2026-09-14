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
```

- Four field kinds, **at most one of each, in any order**; a reader classifies each
  `|`-segment by its lead. `CRC:`, `Seq:` and `Test:` carry a plain list
  (simple-dom’s `lists.md`); the keyword-less segment is the requirement list.
- A `:` is a field-key colon only immediately after `CRC`, `Seq` or `Test`;
  anywhere else it is the description separator, so `R5: desc` reads `R5` then a
  description.
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
