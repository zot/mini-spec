# The gaps schema

The reader over the `## Gaps` section of a mini-spec `design/design.md`: it embeds the
markdown base, **owns the document's DOM**, and adds what the section has that markdown
does not — the typed, numbered gap entry with its checkbox or its permanence. It is what
`query gaps` and the three gap writes (`add-gap`, `resolve-gap`, `approve-gap`) read and
write through, retiring the tool's pre-sdom line reader of the same section.

```go
type Gaps struct { /* the document, its context, the entries */ }

func ParseGaps(src string) *Gaps
func (g *Gaps) Doc() *sdom.Doc
func (g *Gaps) Render() (string, error)

func (g *Gaps) HasGaps() bool                 // a level-2 heading `Gaps` exists
func (g *Gaps) Items() []*Gap                 // every gap entry in order, nested ones included
func (g *Gaps) Gap(id string) *Gap            // by ID — `O81`, `T3`; nil when none
func (g *Gaps) Unread() []Unread              // column-0 bullets that are not gaps, deviant entries, groups never closed

func (g *Gaps) Add(id, text string) error
func (g *Gaps) Resolve(id string) error
func (g *Gaps) Approve(id, newID string) error

var ErrNoSection, ErrNoGap, ErrBadGapID, ErrGapExists, ErrPermanent, ErrResolved error
type DeviationError struct { Key string; Deviations []Deviation }   // shared with the carve

type Gap struct {
    ID      string   // `O81`
    Type    string   // the letter: S R D C I O A T
    Number  int
    Checkbox, Checked bool   // whether the line carries `[ ]`/`[x]`, and which
    Text    string   // the head's text with its continuation lines folded
    Sub     []string // indented un-keyed bullets beneath it, one per bullet, folded
    Depth   int      // the bullet's leading whitespace; 0 is top level
    Parent  *Gap     // the nearest preceding gap with a smaller depth, or nil
}
func (p *Gap) Line() int
func (p *Gap) Permanent() bool                // Type is A or T
func (p *Gap) Deviations() []Deviation
```

## What it reads

**The section is a region**: from the level-2 heading whose text is `Gaps` to the next
heading of level 2 or higher, or the end of the file. Without such a heading `HasGaps` is
false and there are no entries; a fenced `## Gaps` is no heading at all to the base.

**A gap is a bullet inside the region, at any depth, whose head is `X<n>:`** — a type letter
from `S R D C I O A T`, a number, a colon — with or without a checkbox between the bullet and
the key: `- [ ] O81: …`, `- [x] I1: …`, `- A2: …`, `- T7: …`. Depth is the bullet's leading
whitespace, and a gap's `Parent` is the nearest preceding gap with a smaller depth, the rule
the carve reader uses for subparts. A column-0 bullet of any other shape is listed in
`Unread`: the section is a list, so a bullet the reader cannot key is a bullet the tool
cannot see, and it says so. A line inside a code group is body.

**Text folds.** A gap's text is its head line's text with every following line folded on
single spaces until the next bullet, a blank line, or the region's end — gap bodies wrap,
and a reader that took one line would drop the tail of most of them. An indented bullet
beneath a gap that carries no key is one of its `Sub` lines, folded the same way and kept
as written after its bullet, which is what the skill's nested `- [ ] Feature A (5 scenarios)`
and `- reason: …` sub-items are.

**Deviations, listed and refused.** `A` and `T` are permanent and carry no checkbox; the
others track work and carry one. A permanent gap with a checkbox and a tracked gap without
one are each a deviation on that entry. An ID the section already carries is a deviation on
the later entry, and `Gap(id)` returns the first. A deviant entry is listed in `Unread` with
its rule, and every write to it refuses with a `DeviationError`.

**Every gap reports its line**, 1-based, at parse time. `Unread` is ordered by line and also
carries every group open at end of input or closer that closes nothing.

## What it writes

Every write edits inside the region and nothing outside it, decides its refusal before any
byte moves, and re-reads the document afterwards, reading its own write back or panicking
with a `ReadBackError`.

**`Add(id, text)`** appends one entry: `- [ ] <id>: <text>` for a tracked letter, `- <id>: <text>`
for a permanent one, on one line, with the number the caller minted (`query next-id gap`
owns the numbering). It goes directly after the last gap's span — after its continuation and
its sub-items, before any blank line that follows — or directly after the heading's line when
the section is empty. An ID not of the shape `X<n>` is `ErrBadGapID`; one the section already
carries is `ErrGapExists`; no section is `ErrNoSection`. The text is written as given, unwrapped.

**`Resolve(id)`** turns the head line's `[ ]` into `[x]`. A permanent gap is `ErrPermanent`,
because there is nothing to close; a checked one is `ErrResolved`, so a second resolution is
visible where it happens rather than absorbed.

**`Approve(id, newID)`** converts a tracked gap to an approved one: the head line is rewritten
as `- <newID>: <head text>` — no checkbox — and every continuation line and sub-item beneath
it stands as written. `newID` must be an unused `A<n>` (`ErrBadGapID`, `ErrGapExists`); a
permanent target is `ErrPermanent`. The head text is the head line's own text, not the folded
body, so a wrapped entry keeps its wrapping and nothing beneath the head is touched. A
sub-item that carries a checkbox under a newly permanent gap is left as it is: the format
says the sub-items carry the state where a parent has none, and whether they should follow
the parent into permanence is the tool's call, not the reader's.

## What it does not do

It does not mint numbers, filter by open or closed, or interpret a range — `query gaps` does
those over what this reader returns. It does not renumber or delete, and it never writes a
checkbox on `A` or `T`.
