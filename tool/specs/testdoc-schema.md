# The test-design schema

The first reader over a mini-spec design document rather than a trajectory file: it embeds
the markdown base, **owns a test design's DOM** (`design/test-*.md`), and adds what a test
design has that markdown does not — the `## Test:` entry and the five alarm fields inside
it. It is what the tool's alarm census and its three alarm write verbs (`update pulled`,
`update inject`, `update number-alarms`) read and write through.

```go
type TestDoc struct { /* the document, its context, the entries */ }

func ParseTestDoc(src string) *TestDoc
func (t *TestDoc) Doc() *sdom.Doc
func (t *TestDoc) Render() (string, error)

func (t *TestDoc) Tests() []*TestEntry        // every `## Test:` entry, in order
func (t *TestDoc) Alarm(n int) *TestEntry     // the entry whose `**Alarm:**` is n; nil when none
func (t *TestDoc) Unread() []Unread           // level-2 headings that are not tests, doubled fields, groups never closed

func (t *TestDoc) SetPulled(n int, date, body string) error
func (t *TestDoc) SetInject(n int, sites []Site, void bool) error
func (t *TestDoc) NumberAlarms() ([]int, error) // the numbers assigned, in document order

var ErrNoAlarm, ErrNoInject, ErrEmptyInject error
type DeviationError struct { Key string; Deviations []Deviation }   // shared with the carve

type Site struct{ File, Symbol string }       // one `file:symbol`
type Pulled struct{ Date, Body string }       // the leading date, and everything after it
type TestEntry struct {
    Title     string
    FireAlarm string    // "" when the entry records no alarm
    Inject    []Site
    Pulled    *Pulled   // nil when never pulled: a prescription, not a record
    Code      []string
    Alarm     int       // 0 when unnumbered
}
func (e *TestEntry) Line() int                // 1-based, at parse time
func (e *TestEntry) HasAlarm() bool           // a `**Fire alarm:**` field is present
func (e *TestEntry) Deviations() []Deviation  // each doubled field, by name
```

## What it reads

**An entry is a region**: from a level-2 heading whose text begins `Test:` to the next
heading of level 2 or higher, or the end of the file. The title is every byte after `Test:` to
the end of the heading's line, read as source — a code span or emphasis in the heading rides
along as its own bytes rather than ending the title at its opener.
A level-2 heading that is not a test — `## Notes`, a stray `## Status` — is listed as
unread rather than guessed at, and a fenced `## Test:` is no heading at all to the base, so
it neither opens an entry nor ends one.

**A field is `**Name:**` at the head of a line inside the entry, outside any code group.**
Five names are read: `Fire alarm`, `Inject`, `Pulled`, `Code`, `Alarm`. Any other
`**Name:**` at a line head — `Purpose`, `Input`, `Expected`, `Refs` — is the test design's
own and is body here, though it still ends the field above it. **The two prose fields fold**:
a `Fire alarm` or `Pulled` body continues across following lines to the next field line or
the end of the entry, because these documents wrap and a reader that took one line per
field would see half of every alarm. **The three list fields are one line each** — `Inject`,
`Code`, `Alarm` — and a line following one of them is body, which is what lets a demoted
`Pulled` record sit under `Inject` without being read as a site. Measured 2026-09-06 over
mini-spec's 20 test designs: 40 `Fire alarm` and 5 `Pulled` continuations, 0 `Inject`,
0 `Alarm`, 1 `Code`.

**Inside a code group a line is body, not a field.** A `**Alarm:** 1` quoted in a fenced
example names nothing, on the read side and on the write side alike; the base's context
says which lines a code group encloses, so this is structural rather than a scan for fence
markers.

Per field:

- `**Fire alarm:**` — prose. Its presence is what makes the entry an alarm (`HasAlarm`).
- `**Inject:**` — a comma-separated list of `file:symbol` sites, whitespace trimmed. A
  site with no colon is read as a file and an empty symbol; the reader records what was
  written and leaves judging it to the census.
- `**Pulled:**` — a leading `YYYY-MM-DD` date, then everything after it as the body,
  leading separators and whitespace trimmed. A `**Pulled:**` line with no leading date is
  a deviation. The reader carries no commit: `old-sdom`'s optional `@ <commit>` after the
  date was retired (Bill, 2026-09-04) and no live document writes it.
- `**Code:**` — a comma-separated list of test files, as written.
- `**Alarm:**` — an integer, local to the document. A non-integer is a deviation.

**An `**Alarm:**` on an entry with no `**Fire alarm:**` is a deviation**: the number names an alarm that is not there.

**A doubled field in one entry is a deviation, not a guess.** The first occurrence is read;
the entry's `Deviations` name the field, and every write to that entry refuses.

**Every entry reports its line**, 1-based, as the document stood when it was parsed.

`Unread` lists, ordered by line: every level-2 heading that did not read as a test, every
entry carrying a deviation, and every bracket group open at end of input or closer that
closes nothing, as the trajectory readers do.

## What it writes

Every write addresses an entry by its alarm number, edits inside the entry's region and
nothing outside it, and re-reads the document afterwards, reading its own write back or
panicking with a `ReadBackError`. A number no entry carries is `ErrNoAlarm`; an entry
carrying deviations refuses with a `DeviationError` naming each, and the bytes are
unchanged. A refusal is decided before any byte moves.

**`SetPulled(n, date, body)`** writes the `**Pulled:**` line as `**Pulled:** <date> — <body>`.
When the entry already carries one, the new line replaces it and the old line's content is
folded after the body as ` *Earlier —* <old content>`, so the leading date moves and the
history is kept in one line. When it carries none, the line is inserted directly after the
`**Inject:**` field's last line, or, with no `**Inject:**`, after the `**Fire alarm:**`
field's last line. The date is the caller's: which clock it comes from is the verb's rule.

**`SetInject(n, sites, void)`** rewrites the `**Inject:**` line as `**Inject:** ` followed by
the sites joined by `, `. No `**Inject:**` line is `ErrNoInject`; an empty site list is
`ErrEmptyInject`, because an alarm with no site is a state to record, not a value to write.
When `void` is set and the entry carries a `**Pulled:**` line, that line is demoted out of
field shape in the same write:

    *Pulled at `<old sites>` — <old content> — and the site has since moved, so this is history rather than a record.*

naming the sites the pull was earned at. Whether the sites really moved is the caller's
judgment — resolving symbols is the tool's business — but only the writer holds the old
sites at the moment of the write, so the demotion is a flag on the replace rather than a
second write. With `void` false the `**Pulled:**` line stands.

**`NumberAlarms()`** gives every alarm entry lacking an `**Alarm:**` field the next free
number in the document — one more than the highest number any entry carries, counting from
there — inserting `**Alarm:** <n>` directly above the `**Fire alarm:**` line. Append-only:
an existing number is kept whatever order it now sits in. Idempotent: a second call assigns
nothing and changes no byte. An entry with no `**Fire alarm:**` is not an alarm and gets no
number. An unnumbered alarm entry carrying deviations refuses the whole run with a
`DeviationError` before any byte moves. Returns the numbers assigned, in document order.

## What it does not do

It does not resolve a site to a declaration, ask git anything, or judge freshness; those
are the census's, over what this reader returns. It does not read `**Purpose:**`,
`**Input:**`, `**Expected:**` or `**Refs:**` into fields, since no verb consumes them. It
does not renumber, and it does not fill in an absent `**Pulled:**` or `**Alarm:**`, because
absence is data: a prescription and a record read identically in prose.
