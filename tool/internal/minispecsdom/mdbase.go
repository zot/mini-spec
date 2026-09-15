package minispecsdom

import (
	"fmt"
	"strings"

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

// CRC: crc-Markdown.md | Seq: seq-links.md#1 | R448, R454
//
// Markdown is the base as a document of its own: any markdown file, read for the one thing
// no schema reader asks — where it points. It writes nothing and resolves nothing; the file
// system and git are the classifier's.
type Markdown struct {
	markdownDoc
	src   string
	links []Link
}

// CRC: crc-Markdown.md | R448, R451
//
// Link is one inline link, `[text](dest)` or `![alt](dest)`, as the document was parsed.
type Link struct {
	Raw      string // the link as written, `[` (or `!`) through `)`
	Text     string // the bytes between the brackets
	Dest     string // the destination, `<…>` unwrapped and any title stripped
	Path     string // Dest before the first `#`; "" for a fragment-only link
	Fragment string // the bytes after the first `#`; "" when none
	Image    bool   // opened by `!`
	line     int
	offset   int
}

// Line is the link's 1-based line at parse time.
func (l Link) Line() int { return l.line }

// Offset is the byte offset of the opening `[`, or of the `!` before it.
func (l Link) Offset() int { return l.offset }

// CRC: crc-Markdown.md | Seq: seq-links.md#1.1 | R448
// ParseMarkdown parses src with the base and reads its links.
func ParseMarkdown(src string) *Markdown {
	m := &Markdown{src: src}
	m.parseBase(src)
	m.scanLinks()
	return m
}

// Links is every link, in document order.
func (m *Markdown) Links() []Link { return m.links }

// CRC: crc-Markdown.md | Seq: seq-links.md#1.4 | R453
// Unread is every bracket group open at end of input or closer that closes nothing.
func (m *Markdown) Unread() []Unread {
	u := unbalanced(m.ctx)
	byLine(u)
	return u
}

// CRC: crc-Markdown.md | Seq: seq-links.md#1.2 | R449, R450
//
// scanLinks walks the source for the inline form. It scans the source rather than each
// Text node because link text may hold emphasis, whose markers are nodes of their own; what
// the base contributes is `inCode`, the structural test a regex cannot make. A `[` inside a
// code group is skipped; so is one whose brackets do not balance, one not followed directly
// by `(`, and one whose destination has no `)` on its line — each is text. Nested brackets
// balance one level, so `[a [b] c](x)` reads whole while a checkbox's `[ ]` beside a link
// is passed over and the link after it found.
func (m *Markdown) scanLinks() {
	src := m.src
	for i := 0; i < len(src); i++ {
		if src[i] != '[' || m.inCode(i) {
			continue
		}
		depth, j := 1, i+1
		for ; j < len(src) && depth > 0; j++ {
			switch src[j] {
			case '[':
				depth++
			case ']':
				depth--
			}
		}
		if depth != 0 || j >= len(src) || src[j] != '(' {
			continue
		}
		k := strings.IndexAny(src[j+1:], ")\n")
		if k < 0 || src[j+1+k] != ')' {
			continue
		}
		start := i
		image := i > 0 && src[i-1] == '!'
		if image {
			start--
		}
		end := j + 2 + k
		link := splitDest(src[j+1 : j+1+k])
		link.Raw = src[start:end]
		link.Text = src[i+1 : j-1]
		link.Image = image
		link.offset = start
		link.line = m.doc.Line(start)
		m.links = append(m.links, link)
		i = end - 1
	}
}

// CRC: crc-Markdown.md | Seq: seq-links.md#1.2.3 | R451
// splitDest reads a raw destination: `<…>` unwrapped, a trailing quoted title stripped,
// and path and fragment split at the FIRST `#`.
func splitDest(raw string) Link {
	d := strings.TrimSpace(raw)
	if n := len(d); n > 1 && (d[n-1] == '"' || d[n-1] == '\'') {
		if sp := strings.LastIndexAny(d[:n-1], " \t"); sp >= 0 {
			d = strings.TrimSpace(d[:sp])
		}
	}
	if len(d) >= 2 && d[0] == '<' && d[len(d)-1] == '>' {
		d = d[1 : len(d)-1]
	}
	l := Link{Dest: d, Path: d}
	if i := strings.IndexByte(d, '#'); i >= 0 {
		l.Path, l.Fragment = d[:i], d[i+1:]
	}
	return l
}
