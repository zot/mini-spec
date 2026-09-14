package minispecsdom

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// CRC: crc-TestDoc.md | R420, R422
//
// The shapes a test design's alarm fields take. A field is `**Name:**` at the head of a
// line; the five named here are read, and any other name is body that still ends the field
// above it (R420).
var (
	fieldHeadRe = regexp.MustCompile(`^\*\*([A-Z][A-Za-z ]*):\*\*[ \t]*(.*)$`)
	pulledRe    = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})\s*(?:[—–-]\s*)?(.*)$`)
)

const (
	fieldFireAlarm = "Fire alarm"
	fieldInject    = "Inject"
	fieldPulled    = "Pulled"
	fieldCode      = "Code"
	fieldAlarm     = "Alarm"
)

// CRC: crc-TestDoc.md | R425, R427
var (
	ErrNoAlarm     = errors.New("minispecsdom: no entry carries that alarm number")
	ErrNoInject    = errors.New("minispecsdom: the entry carries no **Inject:** line to rewrite")
	ErrEmptyInject = errors.New("minispecsdom: refusing an empty **Inject:** — an alarm with no site is a state to record, not a value to write")
)

// CRC: crc-TestDoc.md | R422
// Site is one `file:symbol` an injection edits.
type Site struct{ File, Symbol string }

func (s Site) String() string {
	if s.Symbol == "" {
		return s.File
	}
	return s.File + ":" + s.Symbol
}

// CRC: crc-TestDoc.md | R422
// Pulled is a `**Pulled:**` record: the leading date the census reads, and everything after it.
type Pulled struct{ Date, Body string }

// CRC: crc-TestDoc.md | R418
//
// TestDoc is the test-design schema: it embeds the markdown base, owns the document, and
// reads each `## Test:` heading's region as an entry with its alarm fields.
type TestDoc struct {
	markdownDoc

	entries []*TestEntry
	unread  []Unread
}

// CRC: crc-TestDoc.md | R418, R422, R423, R424
type TestEntry struct {
	Title     string
	FireAlarm string
	Inject    []Site
	Pulled    *Pulled
	Code      []string
	Alarm     int

	line       int
	start, end int // the region's byte span in the document, [start, end)
	fields     []field
	deviations []Deviation
}

// field is one recognized field's line span inside the entry: the head line's start to the
// offset one past its last continuation line.
type field struct {
	name       string
	start, end int
	content    string // after `**Name:**`, continuation lines folded with single spaces
}

// CRC: crc-TestDoc.md | R424
func (e *TestEntry) Line() int { return e.line }

// CRC: crc-TestDoc.md | R422
func (e *TestEntry) HasAlarm() bool { return e.field(fieldFireAlarm) != nil }

// CRC: crc-TestDoc.md | R423
func (e *TestEntry) Deviations() []Deviation { return e.deviations }

func (e *TestEntry) field(name string) *field {
	for i := range e.fields {
		if e.fields[i].name == name {
			return &e.fields[i]
		}
	}
	return nil
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1 | R418
func ParseTestDoc(src string) *TestDoc {
	t := &TestDoc{}
	t.parse(src)
	return t
}

// parse reads src with the markdown base and derives the view over the result.
func (t *TestDoc) parse(src string) {
	t.parseBase(src)
	t.scan()
}

// reload re-reads the document from its bytes. A written line is one synthetic text
// until it is parsed, so after a write the array and the view are rebuilt together.
func (t *TestDoc) reload() {
	src, _ := t.doc.Render()
	t.parse(src)
}

func (t *TestDoc) Tests() []*TestEntry { return t.entries }
func (t *TestDoc) Unread() []Unread    { return t.unread }

// CRC: crc-TestDoc.md | R425
func (t *TestDoc) Alarm(n int) *TestEntry {
	if n == 0 {
		return nil // an unnumbered entry answers to no number
	}
	for _, e := range t.entries {
		if e.Alarm == n {
			return e
		}
	}
	return nil
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.2 | R418, R419, R424
// scan derives the entries and the unread list from the document.
func (t *TestDoc) scan() {
	t.entries, t.unread = nil, nil
	nodes := t.doc.Nodes()
	for i, n := range nodes {
		h, ok := n.(*schema.Heading)
		if !ok || h.Level() != 2 {
			continue
		}
		title := headingLine(t.doc, h) // R446: the whole line after the marker, as source
		line := t.doc.Line(h.Location().Offset())
		rest, isTest := strings.CutPrefix(title, "Test:")
		if !isTest {
			t.unread = append(t.unread, Unread{line, title})
			continue
		}
		stop := t.regionEnd(i) // a node index; the entry's own end is a byte offset
		e := &TestEntry{Title: strings.TrimSpace(rest), line: line, start: h.Location().Offset(), end: t.total}
		if stop < len(nodes) {
			e.end = nodes[stop].Location().Offset()
		}
		t.readFields(e, nodes[i:stop])
		for _, d := range e.deviations {
			t.unread = append(t.unread, Unread{line, fmt.Sprintf("`## Test: %s` — %s: %s", e.Title, d.Rule, d.Target)})
		}
		t.entries = append(t.entries, e)
	}
	t.unread = append(t.unread, unbalanced(t.ctx)...)
	byLine(t.unread)
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.4 | R420, R421, R422, R423
//
// readFields renders the region to lines and reads each field head outside a code group,
// folding its body to the next head. The heading's own line is skipped.
func (t *TestDoc) readFields(e *TestEntry, run []sdom.Node) {
	var b strings.Builder
	for _, n := range run {
		s, _ := n.Render()
		b.WriteString(s)
	}
	off := e.start
	var cur *field
	closeField := func(at int) {
		if cur != nil {
			cur.end = at
			e.fields = append(e.fields, *cur)
			cur = nil
		}
	}
	for i, l := range strings.SplitAfter(b.String(), "\n") {
		lineStart := off
		off += len(l)
		if i == 0 || l == "" {
			continue
		}
		m := fieldHeadRe.FindStringSubmatch(strings.TrimRight(l, "\n"))
		if m != nil && !t.inCode(lineStart) {
			closeField(lineStart)
			if named(m[1]) {
				cur = &field{name: m[1], start: lineStart, content: strings.TrimSpace(m[2])}
			}
			continue
		}
		if cur == nil {
			continue
		}
		if !folds(cur.name) {
			closeField(lineStart) // a list field is one line; what follows is body
			continue
		}
		cur.content = strings.TrimSpace(cur.content + " " + strings.TrimSpace(l))
	}
	closeField(e.end)
	t.derive(e)
}

// folds reports whether a field's body continues onto following lines: the two prose
// fields do; the three list-valued fields are one line each (measured 2026-09-06 over
// mini-spec's 20 test designs: 40 and 5 continuations against 0, 0 and 1).
func folds(name string) bool { return name == fieldFireAlarm || name == fieldPulled }

// named reports whether a field head is one of the five the schema reads.
func named(name string) bool {
	switch name {
	case fieldFireAlarm, fieldInject, fieldPulled, fieldCode, fieldAlarm:
		return true
	}
	return false
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.5 | R422, R423
// derive reads the values from the fields, and the deviations a malformed or doubled field is.
func (t *TestDoc) derive(e *TestEntry) {
	seen := map[string]bool{}
	var kept []field
	for _, f := range e.fields {
		if seen[f.name] {
			e.deviations = append(e.deviations, Deviation{"a field appears once in an entry", fmt.Sprintf("`**%s:**` again at line %d", f.name, t.doc.Line(f.start))})
			continue
		}
		seen[f.name] = true
		kept = append(kept, f)
		switch f.name {
		case fieldFireAlarm:
			e.FireAlarm = f.content
		case fieldInject:
			e.Inject = parseSites(f.content)
		case fieldCode:
			e.Code = splitList(f.content)
		case fieldPulled:
			m := pulledRe.FindStringSubmatch(f.content)
			if m == nil {
				e.deviations = append(e.deviations, Deviation{"`**Pulled:**` leads with a YYYY-MM-DD date", f.content})
				continue
			}
			e.Pulled = &Pulled{Date: m[1], Body: strings.TrimSpace(m[2])}
		case fieldAlarm:
			n, err := strconv.Atoi(f.content)
			if err != nil || n <= 0 {
				e.deviations = append(e.deviations, Deviation{"`**Alarm:**` is a positive integer", f.content})
				continue
			}
			e.Alarm = n
		}
	}
	e.fields = kept
	if e.Alarm != 0 && !e.HasAlarm() {
		e.deviations = append(e.deviations, Deviation{"an `**Alarm:**` numbers a `**Fire alarm:**`", fmt.Sprintf("`**Alarm:** %d` with no `**Fire alarm:**`", e.Alarm)})
	}
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseSites(s string) []Site {
	var out []Site
	for _, p := range splitList(s) {
		file, sym, _ := strings.Cut(p, ":")
		out = append(out, Site{File: file, Symbol: sym})
	}
	return out
}

func joinSites(sites []Site) string {
	parts := make([]string, len(sites))
	for i, s := range sites {
		parts[i] = s.String()
	}
	return strings.Join(parts, ", ")
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#2.1 | R423, R425
// writable finds the entry a write addresses and decides its refusal before any byte moves.
func (t *TestDoc) writable(n int) (*TestEntry, error) {
	e := t.Alarm(n)
	if e == nil {
		return nil, ErrNoAlarm
	}
	if len(e.deviations) > 0 {
		return nil, &DeviationError{Key: strconv.Itoa(n), Deviations: e.deviations}
	}
	return e, nil
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#2.2.1 | R426
func (t *TestDoc) SetPulled(n int, date, body string) error {
	e, err := t.writable(n)
	if err != nil {
		return err
	}
	body = strings.TrimSpace(body)
	line := "**" + fieldPulled + ":** " + date + " — " + body
	start := e.insertionPoint()
	end := start // an empty span, so a first record is an insertion
	if old := e.field(fieldPulled); old != nil {
		line += " *Earlier —* " + old.content
		start, end = old.start, old.end
	}
	if err := t.doc.Mutate(func() error { return t.replaceSpan(start, end, line+"\n") }); err != nil {
		return err
	}
	t.reload()
	got := t.Alarm(n)
	ok := got != nil && got.Pulled != nil && got.Pulled.Date == date && strings.HasPrefix(got.Pulled.Body, body)
	mustReadBack("TestDoc", "SetPulled", strconv.Itoa(n), ok, fmt.Sprintf("%s — %s", date, body), fmt.Sprintf("%+v", got))
	return nil
}

// insertionPoint is where a new `**Pulled:**` line goes: after `Inject`, else after `Fire alarm`.
func (e *TestEntry) insertionPoint() int {
	if f := e.field(fieldInject); f != nil {
		return f.end
	}
	return e.field(fieldFireAlarm).end
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#2.2.2 | R427
func (t *TestDoc) SetInject(n int, sites []Site, void bool) error {
	e, err := t.writable(n)
	if err != nil {
		return err
	}
	if len(sites) == 0 {
		return ErrEmptyInject
	}
	inj := e.field(fieldInject)
	if inj == nil {
		return ErrNoInject
	}
	joined := joinSites(sites)
	pulled := e.field(fieldPulled)
	writes := []replacement{{inj.start, inj.end, "**" + fieldInject + ":** " + joined + "\n"}}
	if void && pulled != nil {
		demoted := "*Pulled at `" + inj.content + "` — " + pulled.content + " — and the site has since moved, so this is history rather than a record.*\n"
		writes = append(writes, replacement{pulled.start, pulled.end, demoted})
		// The earlier span first: its end boundary is a parsed node the later span still owns.
		slices.SortFunc(writes, func(a, b replacement) int { return cmp.Compare(a.start, b.start) })
	}
	err = t.doc.Mutate(func() error {
		for _, w := range writes {
			if err := t.replaceSpan(w.start, w.end, w.text); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	t.reload()
	got := t.Alarm(n)
	ok := got != nil && joinSites(got.Inject) == joined && (!void || got.Pulled == nil)
	mustReadBack("TestDoc", "SetInject", strconv.Itoa(n), ok, joined, fmt.Sprintf("%+v", got))
	return nil
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#2.2.3 | R428
func (t *TestDoc) NumberAlarms() ([]int, error) {
	high := 0
	var todo []*TestEntry
	for _, e := range t.entries {
		high = max(high, e.Alarm)
		if e.HasAlarm() && e.Alarm == 0 {
			if len(e.deviations) > 0 {
				return nil, &DeviationError{Key: e.Title, Deviations: e.deviations}
			}
			todo = append(todo, e)
		}
	}
	if len(todo) == 0 {
		return nil, nil
	}
	var assigned []int
	err := t.doc.Mutate(func() error {
		for _, e := range todo {
			high++
			assigned = append(assigned, high)
			at := e.field(fieldFireAlarm).start
			if err := t.replaceSpan(at, at, "**"+fieldAlarm+":** "+strconv.Itoa(high)+"\n"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	t.reload()
	ok := true
	for _, n := range assigned {
		if e := t.Alarm(n); e == nil || !e.HasAlarm() {
			ok = false
			break
		}
	}
	mustReadBack("TestDoc", "NumberAlarms", fmt.Sprint(assigned), ok, "every assigned number on an alarm", fmt.Sprintf("%d entries", len(t.entries)))
	return assigned, nil
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.2 | R446
// headingLine is every byte of the heading's line after its marker, trimmed, read from
// the source rather than the node after the marker — so a code span or emphasis in the
// heading rides along as source instead of ending the title at its opener.
func headingLine(doc *sdom.Doc, h *schema.Heading) string {
	src := doc.Source()
	loc := h.Location()
	from := loc.Offset() + loc.Length()
	to := strings.IndexByte(src[from:], '\n')
	if to < 0 {
		to = len(src) - from
	}
	return strings.TrimSpace(src[from : from+to])
}
