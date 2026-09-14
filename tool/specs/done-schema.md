# The done schema

The third file schema: it embeds the markdown base, **owns the done file's DOM**, and adds
the completion entry. The done file is a ledger, most-recent first, and the tool reads it
for the joins — which queue IDs a commit discharged, which part it closed — and writes it
by prepending.

```go
type Done struct { /* the document, its contexts, the entries */ }

func ParseDone(src string) *Done
func (d *Done) Doc() *sdom.Doc
func (d *Done) Render() (string, error)

func (d *Done) Entries() []*DoneEntry   // in file order: most recent first
func (d *Done) MaxID() int              // the largest queue ID in any identifier slot
func (d *Done) Unread() []Unread        // `- ` lines at column 0 that did not read as entries, and groups open at end of input; each with its line

func (d *Done) Prepend(header, body string) error

type DoneEntry struct {
    Date    string
    IDs     []int      // every #N in the identifier slot, and nothing outside it
    HasSlot bool       // the header carried an identifier slot at all
    Title   string
    Commit  string     // the first backquoted run after the header's bold
    PartDoc, PartKey string  // the backquoted doc#key, from the header or the body
}
func (e *DoneEntry) Line() int   // 1-based, the bullet's line at parse time
```

**An entry begins at `- **` at column 0.** A list item there whose text opens with bold is
an entry; a list item at column 0 that does not is *entry-like* and listed by `Unread` with its line and text,
because a shape-based reader reports clean over what it never recognized unless it says
so. Both are read from the base's `ListItem` node and the text after it.

**A region runs to the next entry-like bullet at column 0, or to a heading of level 2 or
higher.** A body may quote an entry inside a fence; the base emits no list item there, so
it cannot begin or end an entry.

**The identifier slot is the run between the header's em dash and the colon that opens
the title.** Every `#N` in it is a queue ID this entry discharged; a `#N` anywhere else in
the entry is prose — measured, five body lines in one live ledger would have contributed a
queue ID if bodies were read. The slot is flexible on purpose: `#8`, `O201`, `R3392–R3397`,
several joined by `/`, or nothing.

**The part pointer is a backquoted `doc#key`**, taken from the header first and the body
second — the reverse join from a landed change to the reasoning behind it.

**`Prepend(header, body)`** writes a new entry as one synthetic text just after the rule
that ends the preamble — before the first entry, or at the end of the file when there is
none — separated by blank lines. The header is written as given; the schema does not compose
it. The document is re-read after the write.

**Values are derived, never stored**, from the entry's rendered bytes at those positions.
An entry is a view over its run, as a queue entry is; nothing in it is a field a tool
writes into.

**A group open at end of input is unread.** The base's context answers `Unclosed`; every such
opener is listed by `Unread` at its line, with the text *`<marker>` never closed*, after
the entry-like lines — which is file order, since nothing structured can follow a group still
open at the end. A fence or span that runs to end of file takes every later
entry with it and leaves nothing entry-like to list, which is why the reader must say this
itself rather than wait to notice.

**Every write reads itself back, or panics.** After the re-read, the reader checks that it sees
the write it was asked for — `Prepend` finds the new first entry with its header line — and panics with a `ReadBackError` naming the reader, the
write, the key, what was wanted and what came back. A panic and not an error, because the
writer producing bytes its own reader cannot read is a library invariant, not caller input; a
tool recovers it into a refusal naming the file, and the file stays untouched since nothing is
written until render returns. This proves the tree describes the bytes as the write intended;
it does not prove the write addressed the right region, which no re-parse can. DECIDED (Bill,
2026-09-05): read-back, in place of comparing node kinds against a fresh parse, which a
synthetic placement fails by design.