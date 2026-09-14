# The part line

The line a carve's status block is made of, as one node in `minispecsdom` over a
document the markdown base parsed:

    - [x] ~~**Item 1 — record and resolve.**~~ **LANDED (`4c6e974`, 2026-08-04 — `#3`.)**

The tool reads five things from it and writes three, and today does both with a
hand-rolled reader of six regexes. Here it is a `PartLine` node whose **checkbox** and
**key** are bound values, whose **markers** are a list of `MarkerSpan` stencils, and whose
strike is behind two methods. The shapes are `trajectory-format.md`'s; this spec says how
they are modelled, not what they are.

## The node

```go
type PartLine struct { sdom.Compound /* … */ }

func (p *PartLine) Parse(item *schema.ListItem, ctx *sdom.BracketContext) bool
func (p *PartLine) Splice(d *sdom.Doc) error        // inside a mutation window

func (p *PartLine) Checkbox() *schema.Checkbox       // nil when the line has none
func (p *PartLine) Key() string                      // the fragment: "1", "2.2"; "" when unkeyed
func (p *PartLine) Title() *sdom.Text                // nil when there is no keyed head
func (p *PartLine) Markers() []*MarkerSpan
func (p *PartLine) IsStruck() bool
func (p *PartLine) Strike(on bool)
func (p *PartLine) Deviations() []Deviation          // each names the rule and its target shape
```

**It tiles the line from the `- ` marker to the byte before the newline**, reusing the
`ListItem` and `Checkbox` nodes the base emitted and every `**`, `~~` and code-span marker
inside; the interior texts are re-cut into bound and glue pieces. **Every list item line
parses.** Recognition is not the question here — a list item *is* a part line — so
`Parse` returns false only when `item` is not in the document. Whether a line belongs to a
status block is the carve schema's business.

**Which bold run is which is decided by content.** The first bold run after the checkbox is
the **head**: its interior must begin `Item N — ` or `N.M — ` (the word required without a
dot and forbidden with one; the em dash and nothing else) for the line to be keyed. The key
is a bound `Text`, the separator glue, and the **title** the rest of the run's first text.
A later bold run whose interior reads `VERB (attribution)` — the verb in capitals, one or
more words joined by spaces or hyphens — is a **marker**. Anything else is interspersed
text and stays as it is.

**Strike is a structural edit behind an accessor.** `IsStruck` derives from whether the head
sits inside a `~~` group; `Strike(true)` inserts a `~~` opener and closer around the head's
bold run, `Strike(false)` removes them. It rearranges the node's own children — the
document's flat array and its indices never see inside a compound — so no mutation window
is involved. A consumer never touches a `~~` node. It is `Bool.Value` / `Bool.Set` one
level up.

**Deviations are the contract, not an error path.** An unkeyed head, a separator that is
not the em dash, a checkbox interior that is neither blank nor `x` (which the base leaves
as text after the `- `), a verb not in capitals, an `OPEN` attribution that is neither
`#N` nor `not queued`, a `REVERTED` attribution that is not `#N`, a bold run in the superseded
comma form (`**OPEN, not queued.**`) —
each is reported with the shape it must take, and the line still parses and its checkbox
still counts. A read path lists them; a write path refuses on them, as the carve schema says.

**Flexible on input, rigid on output.** The reader accepts what it is handed in any case,
with any inner spacing, with or without a full stop — `OPEN (not queued)`, `OPEN (Not
queued.)`, `OPEN (#3)` all read clean, since the format's own table writes `OPEN (#N)` — and
`Set` writes one canonical form. A superseded *scheme* is different from a loose spelling
and stays a deviation: the bare-number key, and the comma form, which is not a marker at
all and would otherwise go unreported — the silent case the reader exists to name.

## The marker span

```go
type MarkerSpan struct { sdom.Compound /* … */ }

func (m *MarkerSpan) Verb() *sdom.Text        // bound: "LANDED"
func (m *MarkerSpan) Attribution() string     // derived: "`4c6e974`, 2026-08-04 — `#3`."
func (m *MarkerSpan) QueueID() (int, bool)    // derived: the #N inside the attribution
func (m *MarkerSpan) Set(verb, attribution string) error
```

It tiles `**` through `**`, reusing both. The **verb is bound**; the **attribution is
derived** by rendering the nodes between the parentheses, because it crosses code spans —
`` `4c6e974` `` and `` `#3` `` are groups the base emitted, and a value that spans nodes is
read, not stored. `Set` rewrites the whole interior canonically as one text — `VERB
(attribution)` with the verb in capitals — under the guarded write: the render is re-parsed
as a marker and must yield the same verb and attribution, or the write is refused and the
literal unchanged. After a write the interior is one text where a fresh parse would give
code spans; the bytes are identical, and that is the corner a canonical write accepts.

## The pass

```go
func PartLines(d *sdom.Doc, ctx *sdom.BracketContext) ([]*PartLine, error)
```

Every `ListItem` in document order, parsed and spliced in its own mutation window — the
convenience the carve schema builds its status reader on, and what the tests drive.

**One key form: the fragment.** `Key()` returns what `carves/x.md#<key>` carries — `4` for a
part, `2.2` for a subpart — and `Item ` is the head's display word, bound in the node but not
part of the key. Before this the part key was `Item 4` and the subpart key `2.2`, two shapes
for one identity and a third in the fragment; mini-spec's `finish` wrote pointers in the
fragment form and its validator compared them to `Key()`, so every landing read as an orphan.
DECIDED (Bill, 2026-09-05): the fragment, everywhere — `Part(key)`, `SetMarker`, `Land`, the
pending `Source:` line — and no reader accepts both shapes, since two forms was the defect.