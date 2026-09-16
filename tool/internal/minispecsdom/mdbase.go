package minispecsdom

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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
	src      string
	links    []Link
	pointers []Pointer
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
	destSpan [2]int // the bytes between `(` and `)`, at parse time
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
		link.destSpan = [2]int{j + 1, j + 1 + k}
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

// ErrNoLink is SetDest on an index no link carries.
var ErrNoLink = errors.New("minispecsdom: no link at that index")

// CRC: crc-Markdown.md | Seq: seq-links.md#3.5 | R462
//
// SetDest replaces the destination bytes of link i — everything between its `(` and `)`,
// title included — with dest, and nothing else. The reader's one write, for the move
// repair: the text, the parentheses and every byte around the link stay where they were.
// After the write the render is re-read and link i is read back with dest as its raw
// destination, or the reader panics with a ReadBackError.
func (m *Markdown) SetDest(i int, dest string) error {
	if i < 0 || i >= len(m.links) {
		return ErrNoLink
	}
	span := m.links[i].destSpan
	if err := m.doc.Mutate(func() error { return m.replaceSpan(span[0], span[1], dest) }); err != nil {
		return err
	}
	out, err := m.Render()
	if err != nil {
		return err
	}
	fresh := ParseMarkdown(out)
	got := ""
	if i < len(fresh.links) {
		l := fresh.links[i]
		got = out[l.destSpan[0]:l.destSpan[1]]
	}
	mustReadBack("Markdown", "SetDest", strconv.Itoa(i), got == dest, dest, got)
	*m = *fresh
	return nil
}

// CRC: crc-Markdown.md | Seq: seq-links.md#5.1 | R493
//
// Pointer is a code span whose content is a markdown path, optionally `#` and a key: the
// tool's own reference form, which no link machinery reads.
type Pointer struct {
	Raw     string // the span as written, backticks included
	Doc     string // the bytes before `#`, ending in .md
	Key     string // the bytes after `#`, "" when none
	line    int
	offset  int
	docSpan [2]int // the document bytes, at parse time
}

func (p Pointer) Line() int   { return p.line }
func (p Pointer) Offset() int { return p.offset }

// ErrNoPointer is SetPointerDoc on an index no pointer carries.
var ErrNoPointer = errors.New("minispecsdom: no pointer at that index")

var pointerSpanRe = regexp.MustCompile("`([^`\n]+)`")

// Pointers is every pointer, in document order.
func (m *Markdown) Pointers() []Pointer {
	if m.pointers == nil {
		m.scanPointers()
	}
	return m.pointers
}

// CRC: crc-Markdown.md | Seq: seq-links.md#5.1 | R493
//
// scanPointers finds single-backtick spans whose content, before any `#`, ends in `.md`. A
// span inside a fenced block is an example: the base places the fence, and a span opener
// that sits inside a code group is skipped. `#7` is a queue ID and `R5` a requirement, and
// neither ends in `.md`, so neither is a pointer.
func (m *Markdown) scanPointers() {
	m.pointers = []Pointer{}
	for _, loc := range pointerSpanRe.FindAllStringSubmatchIndex(m.src, -1) {
		start, end, cStart, cEnd := loc[0], loc[1], loc[2], loc[3]
		if m.inCode(start) && m.enclosingIsFence(start) {
			continue
		}
		doc, key, _ := strings.Cut(m.src[cStart:cEnd], "#")
		if !strings.HasSuffix(doc, ".md") || strings.ContainsAny(doc, " \t") {
			continue
		}
		m.pointers = append(m.pointers, Pointer{
			Raw: m.src[start:end], Doc: doc, Key: key,
			line: m.doc.Line(start), offset: start, docSpan: [2]int{cStart, cStart + len(doc)},
		})
	}
}

// enclosingIsFence reports whether the code group holding off is a fence rather than the
// span itself: a span's own opener is inside its own group, so inCode alone cannot tell a
// span in prose from a span quoted inside a fenced block.
func (m *Markdown) enclosingIsFence(off int) bool {
	n := nodeAt(m.doc.Nodes(), off)
	if n == nil {
		return false
	}
	lang := m.ctx.Language()
	for enc := m.ctx.Enclosing(n); enc != nil; enc = m.ctx.Enclosing(enc) {
		s, _ := enc.Render()
		if g := lang.GroupFor(s); g != nil && g.Kind == "code" && len(s) >= 3 {
			return true
		}
	}
	return false
}

// CRC: crc-Markdown.md | Seq: seq-links.md#5.2 | R494
//
// SetPointerDoc replaces pointer i's document bytes — before the `#`, or the whole content
// when there is no key — with doc, the key and the backticks untouched, and reads the
// pointer back at the same index or panics with a ReadBackError.
func (m *Markdown) SetPointerDoc(i int, doc string) error {
	ps := m.Pointers()
	if i < 0 || i >= len(ps) {
		return ErrNoPointer
	}
	span := ps[i].docSpan
	if err := m.doc.Mutate(func() error { return m.replaceSpan(span[0], span[1], doc) }); err != nil {
		return err
	}
	out, err := m.Render()
	if err != nil {
		return err
	}
	fresh := ParseMarkdown(out)
	got := ""
	if fp := fresh.Pointers(); i < len(fp) {
		got = fp[i].Doc
	}
	mustReadBack("Markdown", "SetPointerDoc", strconv.Itoa(i), got == doc, doc, got)
	*m = *fresh
	return nil
}
