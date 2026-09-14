# The current schema

The fourth file schema: it embeds the markdown base, **owns the current file's DOM**, and
adds the one region that matters — `## Active`. Everything else in the file is standing
context, and the whole point of the schema is that a write reaches the region and
nothing else.

```go
type Current struct { /* the document, its contexts, the region */ }

func ParseCurrent(src string) (*Current, error)   // refuses no Active (ErrNoActive), or more than one (ErrManyActive)
func (c *Current) Doc() *sdom.Doc
func (c *Current) Render() (string, error)

func (c *Current) Active() string         // the region's body, trimmed; "" when only the placeholder
func (c *Current) Occupied() bool         // the region holds something other than the placeholder
func (c *Current) Standing() []string     // the other level-2 headings, in order
func (c *Current) Unread() []Unread       // groups open at end of input, each with its opener's line

func (c *Current) SetActive(body string) error   // refuses when Occupied
func (c *Current) Reset() error                  // writes the placeholder
```

**Exactly one `## Active`.** Two make the region ambiguous, and a reader that cannot tell
which it was asked to clear refuses rather than picks; none means the active item cannot be
told from the standing context around it. Both are errors from `ParseCurrent`, matched on
`Heading` nodes so a fenced `## Active` counts for nothing.

**The region runs from the heading to the next heading of level 2 or higher**, so the
active item's context may use `###` and below freely, and a standing `##` section is never
inside it. That boundary is what the 2026-08-18 incident lacked, when a reset defined as
*everything after the rule* deleted 320 lines of standing context.

**A write replaces the region's body as one unit and touches nothing else.** `SetActive`
and `Reset` split the heading's text after its title line, remove the body's nodes, and
insert the new body — as one synthetic text, blank-line separated — before the next
heading or at the end. Every byte outside the region is where it was, by construction:
the nodes outside it are not addressed. `SetActive` **refuses over a held item**: opening
another would discard it, and the pending file is the stack to park it on first.

**The placeholder is `_No active item._`**, and `Occupied` is the region holding anything
else. The document is re-read after each write.

**A group open at end of input is unread.** `Unread` lists every opener the base's context
reports as closed by end of input, at its line, with the text *`<marker>` never closed*;
a fenced `## Active` that never closes is how the region goes missing without a refusal.

**The two refusals are sentinels.** `ParseCurrent` returns `ErrNoActive` or `ErrManyActive`,
so a tool that must name the repair — *add the heading beneath the rule, holding
`_No active item._`* — tells the cases apart with `errors.Is`, never by matching the message.

**Every write reads itself back, or panics.** After the re-read, the reader checks that it sees
the write it was asked for — `SetActive` reads the body back and `Reset` reads the placeholder — and panics with a `ReadBackError` naming the reader, the
write, the key, what was wanted and what came back. A panic and not an error, because the
writer producing bytes its own reader cannot read is a library invariant, not caller input; a
tool recovers it into a refusal naming the file, and the file stays untouched since nothing is
written until render returns. This proves the tree describes the bytes as the write intended;
it does not prove the write addressed the right region, which no re-parse can. DECIDED (Bill,
2026-09-05): read-back, in place of comparing node kinds against a fresh parse, which a
synthetic placement fails by design.