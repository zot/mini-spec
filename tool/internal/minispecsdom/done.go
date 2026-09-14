package minispecsdom

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// CRC: crc-Done.md | R381, R382
//
// The positions trajectory-format.md names, read from an entry's rendered header.
// A backtick is written \x60, which keeps every pattern a raw string.
var (
	doneHeadRe   = regexp.MustCompile(`^- \*\*(\S+)(?:\s*—\s*([^:]*))?:\s*(.*?)\*\*`)
	doneCommitRe = regexp.MustCompile(`\*\*\s*\(\x60([^\x60]*)\x60\)`)
	partPtrRe    = regexp.MustCompile(`\x60([^#\x60]+)#([^\x60]+)\x60`)
)

// CRC: crc-Done.md | R378
//
// Done is the done file schema: it embeds the markdown base, owns the ledger, and adds
// the completion entry as a VIEW over a list-item region.
type Done struct {
	doc    *sdom.Doc
	parser *schema.MarkdownParser
	ctx    *sdom.BracketContext

	entries []*DoneEntry
	unread  []Unread
}

// CRC: crc-Done.md | R382, R384
type DoneEntry struct {
	Date             string
	IDs              []int
	HasSlot          bool
	Title            string
	Commit           string
	PartDoc, PartKey string

	item *schema.ListItem
	line int // 1-based, at parse time
	run  []sdom.Node
}

// CRC: crc-Done.md | R395
func (e *DoneEntry) Line() int { return e.line }

// CRC: crc-Done.md | Seq: seq-done.md#1 | R378, R379, R380
func ParseDone(src string) *Done {
	d := &Done{}
	d.parse(src)
	return d
}

// parse reads src with the markdown base and derives the view over the result.
func (d *Done) parse(src string) {
	d.parser = schema.NewMarkdownParser()
	d.doc = sdom.Parse(src, 0, d.parser)
	d.ctx = d.parser.Indent().Brackets().Context()
	d.scan()
}

// scan re-derives the entries and the entry-like lines.
func (d *Done) scan() {
	d.entries, d.unread = nil, nil
	nodes := d.doc.Nodes()
	for i, n := range nodes {
		it, ok := n.(*schema.ListItem)
		if !ok || !d.atColumnZero(it) {
			continue
		}
		off := it.Location().Offset()
		if !d.opensBold(i) {
			d.unread = append(d.unread, Unread{d.doc.Line(off), d.lineText(off)})
			continue
		}
		// Cloned: a sub-slice of the document's own array shifts under a mutation.
		e := &DoneEntry{item: it, line: d.doc.Line(off), run: slices.Clone(nodes[i:d.regionEnd(i)])}
		e.derive()
		d.entries = append(d.entries, e)
	}
	d.unread = append(d.unread, unbalanced(d.ctx)...) // R402
	byLine(d.unread)
}

// lineText is the source line beginning at off, without its newline.
func (d *Done) lineText(off int) string {
	s, _, _ := strings.Cut(d.doc.Source()[off:], "\n")
	return s
}

// atColumnZero reports whether the bullet begins its line.
func (d *Done) atColumnZero(it *schema.ListItem) bool {
	off := it.Location().Offset()
	return off == 0 || d.doc.Source()[off-1] == '\n'
}

// opensBold reports whether the list item at i is followed by a `**` opener.
func (d *Done) opensBold(i int) bool {
	nodes := d.doc.Nodes()
	if i+1 >= len(nodes) {
		return false
	}
	o, ok := nodes[i+1].(*sdom.Opener)
	if !ok {
		return false
	}
	s, _ := o.Render()
	return s == "**"
}

// CRC: crc-Done.md | Seq: seq-done.md#1.3 | R380
//
// regionEnd is the index one past the entry's last node: the next column-0 list item,
// or a heading of level 2 or higher. A fenced quotation is text to the base.
func (d *Done) regionEnd(start int) int {
	nodes := d.doc.Nodes()
	for i := start + 1; i < len(nodes); i++ {
		switch m := nodes[i].(type) {
		case *schema.ListItem:
			if d.atColumnZero(m) {
				return i
			}
		case *schema.Heading:
			if m.Level() <= 2 {
				return i
			}
		}
	}
	return len(nodes)
}

// CRC: crc-Done.md | Seq: seq-done.md#1.4 | R381, R382
//
// derive reads the values from the run's rendered bytes: the header line's date, the
// identifier slot between the em dash and the colon — the ONLY source of queue IDs —
// the title, the first backquoted commit after the bold, and the part pointer from the
// header first and the body second.
func (e *DoneEntry) derive() {
	var b strings.Builder
	for _, n := range e.run {
		s, _ := n.Render()
		b.WriteString(s)
	}
	text := b.String()
	head, body, _ := strings.Cut(text, "\n")
	if m := doneHeadRe.FindStringSubmatchIndex(head); m != nil {
		e.Date, e.Title = head[m[2]:m[3]], head[m[6]:m[7]]
		// The slot is present when its group participated — which also tells an
		// empty slot `— :` from no slot at all, and never reads a colon in the date.
		if m[4] >= 0 {
			e.HasSlot = true
			for _, id := range queueIDRe.FindAllStringSubmatch(head[m[4]:m[5]], -1) {
				n, _ := strconv.Atoi(id[1])
				e.IDs = append(e.IDs, n)
			}
		}
	}
	if m := doneCommitRe.FindStringSubmatch(head); m != nil {
		e.Commit = m[1]
	}
	for _, s := range []string{head, body} {
		if m := partPtrRe.FindStringSubmatch(s); m != nil {
			e.PartDoc, e.PartKey = m[1], m[2]
			break
		}
	}
}

// CRC: crc-Done.md | R378
func (d *Done) Doc() *sdom.Doc { return d.doc }

// CRC: crc-Done.md | R378
func (d *Done) Render() (string, error) { return d.doc.Render() }

// CRC: crc-Done.md | R379
func (d *Done) Entries() []*DoneEntry { return d.entries }

// CRC: crc-Done.md | R396
func (d *Done) Unread() []Unread { return d.unread }

// CRC: crc-Done.md | R384
func (d *Done) MaxID() int {
	m := 0
	for _, e := range d.entries {
		for _, id := range e.IDs {
			m = max(m, id)
		}
	}
	return m
}

// CRC: crc-Done.md | Seq: seq-done.md#2 | R383
//
// Prepend writes the entry as one synthetic text just after the rule: before the first
// entry's list item, or at the end when there is none. The header is written as given.
func (d *Done) Prepend(header, body string) error {
	text := header + "\n"
	if body != "" {
		text += strings.TrimSuffix(body, "\n") + "\n"
	}
	text += "\n"
	var before sdom.Node
	if len(d.entries) > 0 {
		before = d.entries[0].item
	} else if !strings.HasSuffix(d.doc.Source(), "\n\n") {
		text = "\n" + text
	}
	if err := d.doc.Mutate(func() error {
		return d.doc.Insert(before, sdom.NewText(text, sdom.Synthetic(len(text))))
	}); err != nil {
		return err
	}
	n := len(d.entries)
	src, _ := d.doc.Render()
	d.parse(src)
	// R414: the prepended entry reads back as the first entry, with its header line.
	got := "no new first entry"
	if len(d.entries) > 0 {
		got = d.lineText(d.entries[0].item.Location().Offset())
	}
	mustReadBack("Done", "Prepend", header, len(d.entries) == n+1 && got == header, header, got)
	return nil
}
