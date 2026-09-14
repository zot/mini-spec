# The requirements schema

The reader over a mini-spec `design/requirements.md`: it embeds the markdown base, **owns
the document's DOM**, and adds what the file has that markdown does not — sections at any
heading level, each with its own content, the `**Source:**` line, and the numbered
requirement entry in its live and retired forms. It is what `update add-req` and
`update retire` write through, and what `query requirements`, `next-id req` and the
coverage checks read.

```go
type Requirements struct { /* the document, its context, the sections and entries */ }

func ParseRequirements(src string) *Requirements
func (r *Requirements) Doc() *sdom.Doc
func (r *Requirements) Render() (string, error)

func (r *Requirements) Sections() []*Section              // every heading, any level, in order
func (r *Requirements) Section(title string) []*Section   // by exact title; several when the title repeats
func (r *Requirements) Requirements() []*Requirement      // every entry in order, retired ones included
func (r *Requirements) Requirement(id string) *Requirement  // by `R5`; nil when none
func (r *Requirements) Unread() []Unread                  // unkeyed column-0 bullets, deviant entries, groups never closed

func (r *Requirements) Add(title, id, text string) error
func (r *Requirements) Retire(id, tn, clause string) error

var ErrNoRequirement, ErrBadReqID, ErrReqExists, ErrRetired, ErrBadClause, ErrManySections error
// ErrNoSection and DeviationError are shared with the gaps and carve readers

type Section struct {
    Title  string
    Level  int
    Source string    // the `**Source:**` line's text, "" when none
    Parent *Section  // the nearest preceding section with a smaller level, or nil
}
func (s *Section) Line() int

type Requirement struct {
    ID          string   // `R5`
    Number      int
    Text        string   // the entry's text with continuation lines folded; the retired clause removed
    Retired     bool
    RetiredBy   string   // `T3` when retired
    Replacement string   // `R10` when retired with one, "" for no replacement
    Section     *Section
}
func (q *Requirement) Line() int
func (q *Requirement) Deviations() []Deviation
```

## What it reads

**A section is a heading at any level and its own content**, which runs to the next
heading of *any* level or the end of the file. A `## Feature:` section with a `### Notes`
beneath it therefore owns only the lines above the `###`; the sub-heading is a section of
its own, with the feature as its `Parent`. A fenced heading is no heading at all to the
base, so it neither opens a section nor ends one.

**A requirement is a column-0 bullet inside a section whose head is `**Rn:**` or its retired
form `**~~Rn:~~**`.** The text folds across following lines on single spaces to the next
bullet, a blank line, a heading, or the end of the section, because requirement bodies
wrap. A retired entry's head text begins with the clause `(Retired Tn — see Rm)` or
`(Retired Tn — no replacement)`, which is read into `RetiredBy` and `Replacement` and
removed from `Text`, so the text is the original requirement in both forms. A retired head
with no such clause is a deviation: the strikethrough says the requirement is dead and
nothing says where it went. A column-0 bullet of any other shape is listed in `Unread`. A
line inside a code group is body.

**`**Source:**` at column 0 inside a section is that section's source**, the first one when
there are several, with any later one listed in `Unread`.

**A repeated ID is a deviation on the later entry**, and `Requirement(id)` returns the
first. A deviant entry is listed in `Unread` with its rule and refuses every write.

**Every section and entry reports its line**, 1-based, at parse time. `Unread` is ordered by
line and also carries every group open at end of input or closer that closes nothing.

## What it writes

Every write edits inside one section's own content and nothing outside it, decides its
refusal before any byte moves, and re-reads the document afterwards, reading its own write
back or panicking with a `ReadBackError`.

**`Add(title, id, text)`** appends `- **<id>:** <text>` on one line at the end of the named
section's own content — after its last non-blank line, before any blank lines that separate
it from the next heading — which is the add-req rule: a new requirement goes before the
section's first sub-heading, never after it. `id` is the caller's (`query next-id req` mints
it, counting retired numbers): not of the shape `R<n>` is `ErrBadReqID`, already present is
`ErrReqExists`. A title no section carries is `ErrNoSection`; one that several carry is
`ErrManySections`, because a section title is a judgment and the reader does not pick.

**`Retire(id, tn, clause)`** rewrites the head line of a live entry as
`- **~~<id>:~~** (Retired <tn> — <clause>) <head text>`, leaving every continuation line as
written. `clause` is `see R<m>` or `no replacement` and anything else is `ErrBadClause`;
`tn` is `T<n>`; an entry already retired is `ErrRetired`, so a second retirement is visible
rather than absorbed. The `Tn` entry in the gaps section is the gaps reader's write, made by
the verb that calls both.

## What it does not do

It does not mint numbers, compute coverage, or judge whether a Source resolves. It does not
wrap text, delete or renumber an entry, or write the `Tn` gap. It reads `(inferred)` as
ordinary text.
