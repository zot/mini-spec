package minispecsdom

import (
	"fmt"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// CRC: crc-TestDoc.md | R418, R421, R425
//
// markdownDoc is what the design-document readers share: the base parse, the region end,
// the code-group test, and span replacement. A reader embeds it and adds its own scan.
type markdownDoc struct {
	doc    *sdom.Doc
	parser *schema.MarkdownParser
	ctx    *sdom.BracketContext
	total  int // the rendered length at parse time
}

// parseBase parses src, keeps the bracket context the code-group test reads, and measures
// the document.
func (m *markdownDoc) parseBase(src string) {
	m.parser = schema.NewMarkdownParser()
	m.doc = sdom.Parse(src, 0, m.parser)
	m.ctx = m.parser.Indent().Brackets().Context()
	m.total = docLength(m.doc.Nodes())
}

func (m *markdownDoc) Doc() *sdom.Doc          { return m.doc }
func (m *markdownDoc) Render() (string, error) { return m.doc.Render() }

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.3 | R418
// regionEnd is the index of the next heading of level 2 or higher after start, or len.
func (m *markdownDoc) regionEnd(start int) int {
	nodes := m.doc.Nodes()
	for i := start + 1; i < len(nodes); i++ {
		if h, ok := nodes[i].(*schema.Heading); ok && h.Level() <= 2 {
			return i
		}
	}
	return len(nodes)
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.4 | R421
// inCode reports whether the node holding offset off sits inside a code group.
func (m *markdownDoc) inCode(off int) bool {
	n := nodeAt(m.doc.Nodes(), off)
	if n == nil {
		return false
	}
	lang := m.ctx.Language()
	for enc := m.ctx.Enclosing(n); enc != nil; enc = m.ctx.Enclosing(enc) {
		s, _ := enc.Render()
		if g := lang.GroupFor(s); g != nil && g.Kind == "code" {
			return true
		}
	}
	return false
}

// docLength is the rendered length of the parsed nodes, which is where a region ends
// when no heading follows it.
func docLength(nodes []sdom.Node) int {
	total := 0
	for _, n := range nodes {
		total += n.Location().Length()
	}
	return total
}

// nodeAt is the parsed node whose span holds off; synthetic nodes have no span.
func nodeAt(nodes []sdom.Node, off int) sdom.Node {
	for _, n := range nodes {
		l := n.Location()
		if l.Offset() >= 0 && l.Offset() <= off && off < l.Offset()+l.Length() {
			return n
		}
	}
	return nil
}

// replacement is one span of parsed bytes and the text that takes its place.
type replacement struct {
	start, end int
	text       string
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#2.3 | R425
//
// replaceSpan replaces the parsed bytes [start, end) with text, inside a mutation window.
// Boundaries fall at node edges or inside a Text, which is split; the nodes wholly inside
// the span are removed and one synthetic text takes their place. Offsets are those of the
// parse, and a synthetic node has none, so several spans may be replaced in one window as
// long as no two overlap.
func (m *markdownDoc) replaceSpan(start, end int, text string) error {
	right, err := m.boundary(end)
	if err != nil {
		return err
	}
	left, err := m.boundary(start)
	if err != nil {
		return err
	}
	var gone []sdom.Node
	inside := false
	for _, n := range m.doc.Nodes() {
		if n == left {
			inside = true
		}
		if n == right {
			break
		}
		if inside {
			gone = append(gone, n)
		}
	}
	for _, n := range gone {
		if err := m.doc.Remove(n); err != nil {
			return err
		}
	}
	return m.doc.Insert(right, sdom.NewText(text, sdom.Synthetic(len(text))))
}

// boundary returns the node beginning exactly at off, splitting a Text when off falls
// inside one; nil when off is the end of the document.
func (m *markdownDoc) boundary(off int) (sdom.Node, error) {
	nodes := m.doc.Nodes()
	n := nodeAt(nodes, off)
	if n == nil {
		if off >= m.total {
			return nil, nil
		}
		return nil, fmt.Errorf("minispecsdom: no parsed node at offset %d; a span was already replaced there", off)
	}
	if n.Location().Offset() == off {
		return n, nil
	}
	_, right, err := m.doc.Split(n, off-n.Location().Offset())
	return right, err
}
