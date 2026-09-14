# The pending schema

The second file schema: it embeds the markdown base, **owns the pending file's DOM**, and
adds the queue entry. An entry is a `## N.` heading and everything beneath it to the next
heading of level 2 or higher, or a `---` rule outside a fence.

```go
type Pending struct { /* the document, its contexts, the entries */ }

func ParsePending(src string) *Pending
func (p *Pending) Doc() *sdom.Doc
func (p *Pending) Render() (string, error)

func (p *Pending) Entries() []*Entry      // in file order — the top entry is active
func (p *Pending) Entry(id int) *Entry
func (p *Pending) MaxID() int              // 0 when empty
func (p *Pending) Unread() []Unread        // level-2 headings that are not entries, and groups open at end of input; each with its line

func (p *Pending) Place(e EntryText, pos int) error   // 1-based among entries; len+1 lands where the entries end
func (p *Pending) After(id int) (int, error)          // the position that follows a live entry
func (p *Pending) Remove(id int) error

type Entry struct {
    ID                int
    Title, Skill, Status string   // derived from the heading line — the title from its emphasis node, whole
    SourceDoc, SourceKey string   // derived from the Source: line
    Kind              SourceKind  // SourcePart, SourceGap, or SourceNone when the line did not read
    Next              string      // derived from the Next: line, "" when absent
}
type SourceKind int
const ( SourceNone SourceKind = iota; SourcePart; SourceGap )
func (e *Entry) Line() int         // 1-based, the heading's line at parse time

type Unread struct { Line int; Text string }   // shared by every reader here

type EntryText struct { ID int; Title, Skill, Status, SourceDoc, SourceKey, Next string; Kind SourceKind }
var ErrBadGapSource error   // Place refuses a gap key that is not one gap ID
```

**An entry is a view, not a node.** Unlike a part line it has no field a tool writes into:
the tool mints whole entries and removes whole entries, and reads the number, the Source
pointer and the Next line. So the schema keeps each entry as the run of flat nodes from
its heading to the end of its region, derives its values from that run's rendered bytes at
the positions `trajectory-format.md` names — `## N.` at column 0, the `Source:` line
beneath, the key after `part` or `gap` — and never re-cuts them. Everything in the run
stays exactly as the base parsed it.

**A source is a carve part or a gap, told apart by the word and by shape.** After the
document link comes either ``part `#<key>` `` or ``gap `<gap ID>` `` — the key in its one form,
the fragment `4` or `2.2`, never `Item 4`; the reader says which it read in `Kind`, and
`SourceKey` carries the key for both, the part key without its `#`. A gap
source names exactly one ID — `O136`, `R42`, `T7` — because an entry discharges one thing: a
range, a list, or a `#` there does not read, and the entry's `Kind` is `SourceNone`. Such a
`Source:` line is listed by `Unread` with its line, since a pointer the reader could not follow
is the silent case. An entry with no `Source:` line at all is `SourceNone` and not listed.

**The writer emits one form for each.** `EntryText.Text()` writes ``part `#<key>` `` or
``gap `<ID>` `` by `Kind`; `Place` refuses a gap key that is not one gap ID, since a source that
would not read back is not a source.

*A door kept open, not built:* an entry may one day discharge several parts — `Source:` names
one and a done entry's body may name more. Nothing here should make that a breaking change.

**A region ends at the next heading of level 2 or higher, or at a `---` line outside a
code group.** The rule is a whole line of the document — a code span reading `---`
mid-sentence is prose, and a `---` line inside a fence is that fence's. A fence in a body is
that entry's: the base emits no heading inside one, so a quoted `## 5.` cannot end an
entry, by construction.

**Placement is a node placement, never a byte splice.** `Place(e, pos)` renders the
canonical entry — the heading line, the `Source:` line, an optional `Next:` line, a blank
line — as one synthetic text and inserts it before the entry at `pos`. **At one past the
last it lands where the entries end, not where the file does**: before the rule that closes
the region when one follows — the shape `trajectory-format.md` describes, entries then `---`
then commentary — else at end of file, where the separator is adjusted so the file still ends
in one newline; with no entries at all, after the header's rule. A position outside
`1 … len+1` is **refused, not clamped**: a clamp silently reinterprets an instruction the
caller was specific about.

**Removal drops the run, and Place then Remove is the identity.** `Remove(id)` takes the
entry's nodes out of the array — splitting the shared tail text at the region's end so the
next entry's or the rule's bytes stay — and, when the entry was the last thing in the file,
drops the blank line its placement opened. `add-item` and `finish` are inverses or they are
not, and the byte-identical round trip is the test that says which; measured 2026-09-05
before this, every round grew the file by one blank line.

**What it could not read is reported.** A level-2 heading whose text does not open `N.`
is not an entry; it is listed by `Unread`, since a shape-based reader is blind to what is
outside its shape and must say so. So is a `Source:` line that names neither a part nor one
gap ID.

**After each write the document is re-read from its bytes** and the view rebuilt with it.
A placed entry is one synthetic text until it is parsed, and nothing holds a node of this
document across a write, so re-reading costs nothing a consumer can see.

**A group open at end of input is unread**, listed at the opener's line with the text
*`<marker>` never closed*, after the unread headings — the same rule and reason as the
done schema's.

**The title is read from the node, not by a regex over the bytes.** The heading's first
emphasis run is the title, its interior read to its own close, so a title with emphasis
inside it — `**A title mentioning **OPEN** here**` — comes back whole; the ID is the bytes
before that run and the skill and status the bytes after it. A regex that matched the first
`**` lazily read such a title as `A title mentioning `. Measured on mini-spec's live queue
2026-09-05: 0 of 11 titles carried interior emphasis.

**Every write reads itself back, or panics.** After the re-read, the reader checks that it sees
the write it was asked for — `Place` finds the entry at its position with every field, `Remove` no longer finds it — and panics with a `ReadBackError` naming the reader, the
write, the key, what was wanted and what came back. A panic and not an error, because the
writer producing bytes its own reader cannot read is a library invariant, not caller input; a
tool recovers it into a refusal naming the file, and the file stays untouched since nothing is
written until render returns. This proves the tree describes the bytes as the write intended;
it does not prove the write addressed the right region, which no re-parse can. DECIDED (Bill,
2026-09-05): read-back, in place of comparing node kinds against a fresh parse, which a
synthetic placement fails by design.