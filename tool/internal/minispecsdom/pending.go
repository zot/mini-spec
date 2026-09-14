package minispecsdom

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// CRC: crc-PendingSdom.md | R372
//
// The positions trajectory-format.md names, read from an entry's rendered bytes.
// A backtick is written \x60, which keeps every pattern a raw string.
var (
	entryIDRe    = regexp.MustCompile(`^## (\d+)\. $`)
	entryRestRe  = regexp.MustCompile(`^(?: \(([^)]*)\))?\.?\s*(.*)$`)
	sourceLineRe = regexp.MustCompile(`^\s*Source:\s*\[[^\]]*\]\(([^)]*)\)(?:,\s*(part|gap)\s+\x60([^\x60]+)\x60)?`)
	nextLineRe   = regexp.MustCompile(`^\s*Next:\s*(.*)$`)
)

// entryOpenRe recognizes the heading text an entry opens with: `N.`.
var entryOpenRe = regexp.MustCompile(`^\d+\.`)

// ErrNoEntry reports an id no entry carries.
var ErrNoEntry = errors.New("minispecsdom: no entry with that id")

// CRC: crc-PendingSdom.md | R400, R401
//
// gapIDRe is one gap ID — O136, R42, T7. A range, a list or a `#` is not a gap source,
// because an entry discharges one thing.
var gapIDRe = regexp.MustCompile(`^[A-Z]\d+$`)

// ErrBadGapSource reports a Place whose gap key is not one gap ID.
var ErrBadGapSource = errors.New("minispecsdom: a gap source names exactly one gap ID")

// CRC: crc-PendingSdom.md | R399
//
// SourceKind says what a Source: line read as: a carve part, a gap, or neither.
type SourceKind int

const (
	SourceNone SourceKind = iota // no Source: line, or one that named neither form
	SourcePart
	SourceGap
)

// CRC: crc-PendingSdom.md | R370
//
// Pending is the pending file schema: it embeds the markdown base, owns the document,
// and adds the queue entry as a VIEW over a heading region.
type Pending struct {
	doc    *sdom.Doc
	parser *schema.MarkdownParser
	ctx    *sdom.BracketContext

	entries []*Entry
	unread  []Unread
}

// CRC: crc-PendingSdom.md | R396
//
// Unread is one thing a reader did not read as an entry: its 1-based line and its
// text, so the report is one somebody can act on. Shared with the done schema.
type Unread struct {
	Line int
	Text string
}

// CRC: crc-PendingSdom.md | R372
//
// Entry is one queue entry: its derived values and the run of flat nodes it spans.
type Entry struct {
	ID                   int
	Title, Skill, Status string
	SourceDoc, SourceKey string
	Kind                 SourceKind
	Next                 string

	head *schema.Heading
	line int         // 1-based, at parse time
	run  []sdom.Node // head first, through the last node of the region
	tail *sdom.Text  // the run's last text when the region ends inside it
	cut  int         // where the region ends inside tail
}

// CRC: crc-PendingSdom.md | R395
func (e *Entry) Line() int { return e.line }

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.2 | R412
//
// deriveHead reads the heading line by its nodes rather than by a regex over its bytes:
// the title is the first emphasis run's interior, read to ITS OWN close — so a title with
// emphasis inside it comes back whole — and the ID and the skill and status are the bytes
// either side of that run.
func (e *Entry) deriveHead(line string, ctx *sdom.BracketContext) {
	var title *sdom.Opener
	for _, n := range e.run[1:] {
		o, ok := n.(*sdom.Opener)
		if !ok {
			continue
		}
		// The first opener on the line is the title's, or there is no title.
		if s, _ := o.Render(); strings.HasPrefix(s, "*") {
			title = o
		}
		break
	}
	// The bytes either side of the title run, and with no title a heading that must be
	// the ID and nothing else, however many blanks it trails.
	head, rest := strings.TrimRight(line, " ")+" ", ""
	if title != nil {
		outer := ctx.OuterText(title)
		at := strings.Index(line, outer)
		if at < 0 {
			return
		}
		e.Title = ctx.InnerText(title)
		head, rest = line[:at], line[at+len(outer):]
	}
	if m := entryIDRe.FindStringSubmatch(head); m != nil {
		e.ID, _ = strconv.Atoi(m[1])
	}
	if m := entryRestRe.FindStringSubmatch(rest); m != nil {
		e.Skill, e.Status = m[1], strings.TrimSuffix(m[2], ".")
	}
}

// CRC: crc-PendingSdom.md | R373, R401
// EntryText is what Place writes.
type EntryText struct {
	ID                                               int
	Title, Skill, Status, SourceDoc, SourceKey, Next string
	Kind                                             SourceKind // SourceGap writes the gap form; anything else the part form
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1 | R370, R371, R372, R376
func ParsePending(src string) *Pending {
	p := &Pending{}
	p.parse(src)
	return p
}

// parse reads src with the markdown base and derives the view over the result.
func (p *Pending) parse(src string) {
	p.parser = schema.NewMarkdownParser()
	p.doc = sdom.Parse(src, 0, p.parser)
	p.ctx = p.parser.Indent().Brackets().Context()
	p.scan()
}

// reload re-reads the document from its bytes. A placed entry is one synthetic text
// until it is parsed, so after a write the array and the view are rebuilt together.
func (p *Pending) reload() {
	src, _ := p.doc.Render()
	p.parse(src)
}

// scan re-derives the entries and the unread headings from the document.
func (p *Pending) scan() {
	p.entries, p.unread = nil, nil
	nodes := p.doc.Nodes()
	for i, n := range nodes {
		h, ok := n.(*schema.Heading)
		if !ok || h.Level() != 2 {
			continue
		}
		var title string
		if i+1 < len(nodes) {
			title = headingText(nodes[i+1])
		}
		line := p.doc.Line(h.Location().Offset())
		if !entryOpenRe.MatchString(title) {
			p.unread = append(p.unread, Unread{line, title})
			continue
		}
		end, cut := p.regionEnd(i)
		// Cloned: a sub-slice of the document's own array shifts under a Remove.
		e := &Entry{head: h, line: line, run: slices.Clone(nodes[i:end])}
		if cut >= 0 {
			e.tail, e.cut = nodes[end-1].(*sdom.Text), cut
		}
		if bad := e.derive(p.ctx); bad != nil {
			p.unread = append(p.unread, *bad)
		}
		p.entries = append(p.entries, e)
	}
	p.unread = append(p.unread, unbalanced(p.ctx)...) // R403
	byLine(p.unread)
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.3 | R371
//
// regionEnd returns the index one past the region's last node, and where the region
// ends inside that last node when a `---` line cuts it (-1 otherwise). A region ends
// at the next heading of level 2 or higher; a fenced heading is no heading at all to
// the base, so it cannot end one.
func (p *Pending) regionEnd(start int) (end, cut int) {
	nodes := p.doc.Nodes()
	for i := start + 1; i < len(nodes); i++ {
		if h, ok := nodes[i].(*schema.Heading); ok && h.Level() <= 2 {
			return i, -1
		}
		if t, ok := nodes[i].(*sdom.Text); ok {
			if at := p.ruleAt(t); at >= 0 {
				return i + 1, at
			}
		}
	}
	return len(nodes), -1
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.3 | R371, R447
//
// ruleAt is the byte offset within t's text of the first `---` that is a whole document
// line outside a code group, or -1. A code span's interior is its own text node reading
// `---`, and a fence's interior is a text node holding a `---` line: neither is a rule
// (measured 2026-09-12, both ended an entry's region). The node's boundaries are checked
// against the source, and its enclosure against the code group. SplitAfter keeps each
// newline with the line it ends, so every element begins one.
func (p *Pending) ruleAt(t *sdom.Text) int {
	if p.inCode(t) {
		return -1
	}
	s, _ := t.Render()
	src, base := p.doc.Source(), t.Location().Offset()
	off := 0
	for _, line := range strings.SplitAfter(s, "\n") {
		if strings.TrimRight(line, " \t\n") == "---" {
			start, end := base+off, base+off+len(line)
			atHead := start == 0 || src[start-1] == '\n'
			atEnd := strings.HasSuffix(line, "\n") || end >= len(src) || src[end] == '\n'
			if atHead && atEnd {
				return off
			}
		}
		off += len(line)
	}
	return -1
}

// CRC: crc-PendingSdom.md | R447
// inCode reports whether n sits inside a code group — a span or a fence.
func (p *Pending) inCode(n sdom.Node) bool {
	lang := p.ctx.Language()
	for enc := p.ctx.Enclosing(n); enc != nil; enc = p.ctx.Enclosing(enc) {
		s, _ := enc.Render()
		if g := lang.GroupFor(s); g != nil && g.Kind == "code" {
			return true
		}
	}
	return false
}

// derive reads the values from the run's rendered bytes, and returns the `Source:` line
// it could not read as either form, or nil.
func (e *Entry) derive(ctx *sdom.BracketContext) *Unread {
	var b strings.Builder
	for i, n := range e.run {
		s, _ := n.Render()
		if e.tail != nil && i == len(e.run)-1 {
			s = s[:e.cut]
		}
		b.WriteString(s)
	}
	lines := strings.Split(b.String(), "\n")
	e.deriveHead(lines[0], ctx)
	var bad *Unread
	for i, l := range lines[1:] {
		if m := sourceLineRe.FindStringSubmatch(l); m != nil && e.SourceDoc == "" {
			e.SourceDoc = m[1]
			word, key := m[2], m[3]
			switch { // Seq: seq-pending.md#1.4.1
			case word == "part":
				e.Kind, e.SourceKey = SourcePart, strings.TrimPrefix(key, "#")
			case word == "gap" && gapIDRe.MatchString(key):
				e.Kind, e.SourceKey = SourceGap, key
			default: // Seq: seq-pending.md#1.4.2
				bad = &Unread{e.line + i + 1, l}
			}
		} else if m := nextLineRe.FindStringSubmatch(l); m != nil && e.Next == "" {
			e.Next = m[1]
		}
	}
	return bad
}

// CRC: crc-PendingSdom.md | R370
func (p *Pending) Doc() *sdom.Doc { return p.doc }

// CRC: crc-PendingSdom.md | R370
func (p *Pending) Render() (string, error) { return p.doc.Render() }

// CRC: crc-PendingSdom.md | R372
func (p *Pending) Entries() []*Entry { return p.entries }

// CRC: crc-PendingSdom.md | R376
func (p *Pending) Unread() []Unread { return p.unread }

// CRC: crc-PendingSdom.md | R372
func (p *Pending) Entry(id int) *Entry {
	for _, e := range p.entries {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// CRC: crc-PendingSdom.md | R376
func (p *Pending) MaxID() int {
	m := 0
	for _, e := range p.entries {
		m = max(m, e.ID)
	}
	return m
}

// CRC: crc-PendingSdom.md | R374
func (p *Pending) After(id int) (int, error) {
	for i, e := range p.entries {
		if e.ID == id {
			return i + 2, nil
		}
	}
	return 0, fmt.Errorf("%w: --after %d", ErrNoEntry, id)
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.2 | R373, R401
// Text renders the canonical entry: heading line, Source line in the part or gap form by
// Kind, Next line, blank line.
func (e EntryText) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "## %d. **%s**", e.ID, e.Title)
	if e.Skill != "" {
		fmt.Fprintf(&b, " (%s)", e.Skill)
	}
	word, key := "part", "#"+e.SourceKey
	if e.Kind == SourceGap {
		word, key = "gap", e.SourceKey
	}
	fmt.Fprintf(&b, ". %s\n   Source: [%s](%s), %s `%s`.\n", e.Status, e.SourceDoc, e.SourceDoc, word, key)
	if e.Next != "" {
		fmt.Fprintf(&b, "   Next: %s\n", e.Next)
	}
	b.WriteString("\n")
	return b.String()
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2 | R407, R377
//
// Place inserts the canonical entry as ONE synthetic text before the entry at pos.
// At one past the last it lands where the entries END, not where the file does: before
// the rule that closes the region when one follows, else at end of file — where the
// separator is adjusted so the file still ends in one newline — and, with no entries at
// all, after the header's rule. Refused rather than clamped outside 1 … len+1: a clamp
// silently reinterprets an instruction the caller was specific about.
func (p *Pending) Place(e EntryText, pos int) error {
	n := len(p.entries)
	if pos < 1 || pos > n+1 {
		return fmt.Errorf("minispecsdom: position %d is outside 1 … %d; refused rather than clamped", pos, n+1)
	}
	if e.Kind == SourceGap && !gapIDRe.MatchString(e.SourceKey) { // Seq: seq-pending.md#2.1.1
		return ErrBadGapSource
	}
	text := e.Text()
	var rule *sdom.Text
	var at int
	var last bool
	if n == 0 {
		rule, at, last = p.headerRule()
	}
	if err := p.doc.Mutate(func() error {
		if pos <= n {
			return p.insert(text, p.entries[pos-1].head)
		}
		if n == 0 {
			return p.placeFirst(text, rule, at, last)
		}
		last := p.entries[n-1]
		if last.tail == nil {
			return p.appendEntry(text)
		}
		// Seq: seq-pending.md#2.3.1 — before the rule that ends the region.
		_, right, err := p.doc.Split(last.tail, last.cut)
		if err != nil {
			return err
		}
		return p.insert(text, right)
	}); err != nil {
		return err
	}
	p.reload()
	// R413: the placed entry reads back, at its position, with every field.
	got := p.Entry(e.ID)
	ok := got != nil && e.matches(got) && slices.Index(p.entries, got) == pos-1
	mustReadBack("Pending", "Place", strconv.Itoa(e.ID), ok, fmt.Sprintf("%+v at %d", e, pos), fmt.Sprintf("%+v", got))
	return nil
}

// matches reports whether got is the entry Text wrote: every field as given, with the
// status shorn of the period Text writes after it, and the source kind Text chose.
func (e EntryText) matches(got *Entry) bool {
	return got.Title == e.Title && got.Skill == e.Skill && got.Status == strings.TrimSuffix(e.Status, ".") &&
		got.SourceDoc == e.SourceDoc && got.SourceKey == e.SourceKey && got.Kind == e.kind() && got.Next == e.Next
}

// kind is the source kind Text writes: SourceGap for a gap, SourcePart otherwise.
func (e EntryText) kind() SourceKind {
	if e.Kind == SourceGap {
		return SourceGap
	}
	return SourcePart
}

// insert puts text before a node as one synthetic text.
func (p *Pending) insert(text string, before sdom.Node) error {
	return p.doc.Insert(before, sdom.NewText(text, sdom.Synthetic(len(text))))
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.3.2 | R407
//
// appendEntry lands an entry at end of file: a blank line separates it from what precedes
// it, and the canonical text's own trailing blank line is dropped so the file ends in one
// newline — the shape Remove restores.
func (p *Pending) appendEntry(text string) error {
	src := p.doc.Source()
	switch {
	case strings.HasSuffix(src, "\n\n"): // the blank line is already there
	case strings.HasSuffix(src, "\n"):
		text = "\n" + text
	default:
		text = "\n\n" + text
	}
	return p.insert(strings.TrimSuffix(text, "\n"), nil)
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.3.3 | R407
//
// headerRule finds the header's rule — the first `---` line in the file — as the text
// holding it, the rule's offset within that text, and whether the text is the file's
// last node. Read-only, and called before the mutation window opens, since deciding
// what is a rule asks the context (R447) and the context does not answer mid-mutation.
func (p *Pending) headerRule() (t *sdom.Text, at int, last bool) {
	nodes := p.doc.Nodes()
	for i, n := range nodes {
		t, ok := n.(*sdom.Text)
		if !ok {
			continue
		}
		if at := p.ruleAt(t); at >= 0 {
			return t, at, i == len(nodes)-1
		}
	}
	return nil, -1, false
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.3.3 | R407
//
// placeFirst lands the only entry after the header's rule, found by headerRule, with one
// blank line between, whether the rule is followed by commentary or by nothing; with no
// rule at all the entry goes at end of file.
func (p *Pending) placeFirst(text string, t *sdom.Text, at int, last bool) error {
	if t == nil {
		return p.appendEntry(text)
	}
	s, _ := t.Render()
	cut := at + len("---\n")
	if cut > len(s) || (cut == len(s) && last) {
		// The rule is the file's last line, with no newline or with nothing after it.
		return p.appendEntry(text)
	}
	if strings.HasPrefix(s[cut:], "\n") {
		cut++ // the blank line after the rule already exists
	} else {
		text = "\n" + text
	}
	_, right, err := p.doc.Split(t, cut)
	if err != nil {
		return err
	}
	return p.insert(text, right)
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#3 | R408, R377
//
// Remove drops the entry's run inside one window, splitting the shared tail text at
// the region's end so the next entry's bytes stay. When the entry was the last thing in
// the file, the blank line its placement opened is dropped too, so that Place then Remove
// is the identity on the bytes — add-item and finish are inverses or they are not.
func (p *Pending) Remove(id int) error {
	e := p.Entry(id)
	if e == nil {
		return ErrNoEntry
	}
	if err := p.doc.Mutate(func() error {
		run := slices.Clone(e.run)
		if e.tail != nil {
			left, _, err := p.doc.Split(e.tail, e.cut)
			if err != nil {
				return err
			}
			run[len(run)-1] = left // the right half is the next entry's
		}
		tail := p.tailBefore(e.head, run) // resolved before the array moves under the removal
		for _, n := range run {
			if err := p.doc.Remove(n); err != nil {
				return err
			}
		}
		closeTail(tail) // Seq: seq-pending.md#3.2.1
		return nil
	}); err != nil {
		return err
	}
	p.reload()
	mustReadBack("Pending", "Remove", strconv.Itoa(id), p.Entry(id) == nil, "no entry", "the entry still reads") // R413
	return nil
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#3.2.1 | R408
// tailBefore is the text that will end the file once run is removed — the nearest text
// before head, past any Indent where the entry's indented lines return to column zero —
// or nil when the run is not the last thing in the file. Resolved before the removal,
// because the node array is live and moves under it.
func (p *Pending) tailBefore(head sdom.Node, run []sdom.Node) *sdom.Text {
	nodes := p.doc.Nodes()
	if run[len(run)-1] != nodes[len(nodes)-1] {
		return nil
	}
	for i := slices.Index(nodes, head) - 1; i >= 0; i-- {
		if t, ok := nodes[i].(*sdom.Text); ok {
			return t
		}
	}
	return nil
}

// closeTail trims the file back to one trailing newline when removing the last entry
// left it ending in a blank line — the separator appendEntry opened.
func closeTail(t *sdom.Text) {
	if t == nil {
		return
	}
	if s, _ := t.Render(); strings.HasSuffix(s, "\n\n") {
		t.SetText(strings.TrimSuffix(s, "\n"))
	}
}
