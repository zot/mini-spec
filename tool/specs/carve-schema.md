# The carve schema

The first of the four file schemas: it embeds the markdown base, **owns a carve file's
DOM**, and adds what a carve has that markdown does not — the `## Status` block and the
part lines inside it. Nothing parses a carve with the base alone.

```go
type Carve struct { /* the document, its contexts, the parts */ }

func ParseCarve(src string) *Carve
func (c *Carve) Doc() *sdom.Doc
func (c *Carve) Render() (string, error)

func (c *Carve) HasStatus() bool
func (c *Carve) Parts() []*Part          // every part line with a checkbox, in order
func (c *Carve) Stateless() []*PartLine  // status lines with no checkbox: split parents, standing constraints
func (c *Carve) Part(key string) *Part   // by the fragment — `4`, `2.2`; nil when no part keys so
func (c *Carve) Unread() []Unread        // groups open at end of input, each with its opener's line

func (c *Carve) SetMarker(key, verb, attribution string) error
func (c *Carve) Land(key, attribution string) error

var ErrNoPart, ErrLanded, ErrReopen error
type DeviationError struct { Key string; Deviations []Deviation }   // names each rule and target

type Part struct {
    *PartLine
    Depth  int    // the bullet's leading whitespace; 0 is top level
    Parent *Part  // the nearest preceding part with a smaller depth, or nil
}
func (p *Part) Line() int   // 1-based, at parse time
```

## What it reads

**The status block is a region**: from the level-2 heading whose text is `Status` to the
next heading of level 2 or higher, or the end of the file. A carve with no such heading has
no status block, and `HasStatus` says so rather than the reader guessing.

**Only list items inside that region become part lines.** A carve's body carries other
checkbox lists — open questions, breakdowns inside an elaboration — that track different
things, and they are left as the base parsed them. Inside a fence the base emits no list
item, so a fenced sample of a status block is invisible here by construction.

**A line with a checkbox is a part; one without is stateless.** No checkbox means
something — a split parent, a standing constraint — so it is recorded rather than dropped,
and never given a state the document declined to state.

**Depth is the bullet's leading whitespace, and a subpart is an indented sibling.** The
flat array has no children; a part's `Parent` is derived as the nearest preceding part
with a smaller depth, the rule the indent frames already use.

Every part carries its line's deviations, so the shape a line must take travels with it.

**Every part reports its line**, 1-based, as the document stood when it was parsed: the
tool's diagnostics print `file:line`, and a count without a line is a report nobody can act
on. A line is read at parse time, not tracked through writes.

## What it writes

**`SetMarker(key, verb, attribution)`** applies the tool's rule to the keyed line: replace
the first *transient* marker — one whose verb is `OPEN` or `REVERTED`, the two a revert and
a replay write over each other — with the new one, remove any other transient, and, when the
line carries none, insert a marker where the grammar `Head Marker* Text?` puts it: after the
head and any markers and **before the trailing prose**, never after it, so the written line is
one this reader lists as conforming. It selects by the kind of
what it replaces, never by what it writes, so setting a record over a line that also
carries `NOT VERIFIED` leaves that standing.

**`Land(key, attribution)`** is the completion write, three markings in one act because the
format requires them to agree: the checkbox becomes `[x]`, the head is struck, and
`LANDED (attribution)` goes on through the marker rule.

Both edit the node's own children — the flat array is untouched — and a write to a key
no part carries is an error, never a silent no-op.

**Both refuse before they mark, and a refused write leaves the line byte-identical.** A read
path lists a deviation; a write path refuses on it. Three refusals:

- **A line carrying deviations** takes no write. The error is a `DeviationError` naming the
  key and every deviation's rule and target, so the caller can print the migration the line
  needs rather than write a canonical marker onto a shape it does not conform to.
- **`OPEN` over a checked part** is refused with `ErrReopen`, whatever else the line says. A
  landed part is never returned to `OPEN` by a write; this is a second guard, not the caller's
  remembering.
- **`Land` over a checked part** is refused with `ErrLanded` rather than made idempotent: a
  second landing carrying a different commit and date would otherwise keep the first record
  silently, and a double landing is the caller's to see where it happens.

Refusal is decided before any of `Land`'s three markings, so the box, the strike and the
marker still agree after a refusal because none of them moved.

**A group open at end of input is unread.** `Unread` lists every opener the base's context
reports as closed by end of input, at its line, with the text *`<marker>` never closed*.
A carve has no other unread kind yet; this is the first entry of the report sdomification
Item 1 will grow.

**Every write reads itself back, or panics.** After the re-read, the reader checks that it sees
the write it was asked for — `SetMarker` and `Land` find the marker on the keyed part in a fresh parse of the render, `Land` its checked, struck box too — and panics with a `ReadBackError` naming the reader, the
write, the key, what was wanted and what came back. A panic and not an error, because the
writer producing bytes its own reader cannot read is a library invariant, not caller input; a
tool recovers it into a refusal naming the file, and the file stays untouched since nothing is
written until render returns. This proves the tree describes the bytes as the write intended;
it does not prove the write addressed the right region, which no re-parse can. DECIDED (Bill,
2026-09-05): read-back, in place of comparing node kinds against a fresh parse, which a
synthetic placement fails by design.