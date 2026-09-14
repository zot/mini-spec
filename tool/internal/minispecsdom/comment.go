// Package minispecsdom holds mini-spec's readers over sdom — the first code that
// knows what a CRC card is. sdom supplies lists, stencils and each language's
// comment style; the keywords live here.
package minispecsdom

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/zot/simple-dom/sdom"
)

// CRC: crc-TraceabilityComment.md | R338, R345
//
// segRe classifies ONE |-segment by which group participates. Anchored at both
// ends: a segment with prose before its keyword does not match, which is what
// makes recognition a matter of consumption rather than of leading with a field.
var segRe = regexp.MustCompile(`(?s)^\s*(?:CRC:(?P<crc>.*)|Seq:(?P<seq>.*)|Test:(?P<test>.*)|(?P<refs>R.*))$`)

// CRC: crc-TraceabilityComment.md | R337, R348
//
// TraceabilityComment is mini-spec's anchor comment as ONE node over the whole
// comment, opener through closer — one kind for every language. Its typed fields
// point at children; nothing is stored twice.
type TraceabilityComment struct {
	sdom.Compound

	opener *sdom.Opener
	closer *sdom.Closer
	inner  *sdom.Text // the interior node Parse read, for the pass to remove

	crc, seq, test *sdom.List
	refs           *sdom.RequirementList
	desc           *sdom.Text
}

// CRC: crc-TraceabilityComment.md | R348
func (c *TraceabilityComment) CRC() *sdom.List { return c.crc }

// CRC: crc-TraceabilityComment.md | R348
func (c *TraceabilityComment) Seq() *sdom.List { return c.seq }

// CRC: crc-TraceabilityComment.md | R348
func (c *TraceabilityComment) Test() *sdom.List { return c.test }

// CRC: crc-TraceabilityComment.md | R348
func (c *TraceabilityComment) Refs() *sdom.RequirementList { return c.refs }

// CRC: crc-TraceabilityComment.md | R341, R348
func (c *TraceabilityComment) Description() *sdom.Text { return c.desc }

// CRC: crc-TraceabilityComment.md | R340
//
// SeqStep splits a Seq item into its path and step: "seq-crud.md#1.4" is
// ("seq-crud.md", "1.4"); an item with no step has step "".
func SeqStep(item string) (path, step string) {
	path, step, _ = strings.Cut(item, "#")
	return
}

// CRC: crc-TraceabilityComment.md
func (c *TraceabilityComment) Equals(other sdom.Node) bool {
	x, ok := other.(*TraceabilityComment)
	return ok && c.Compound.Equals(&x.Compound)
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#2 | R342, R343, R344
//
// Parse reads the comment cmt opens through the context and fills c OFF TO THE
// SIDE, touching no document. False means "not a traceability comment": the
// interior is not a single text node, or the walk did not consume all of it. On
// true the children are the ORIGINAL opener, the interior's glue and fields, and
// the ORIGINAL closer — reused, so a consumer that had either still has it.
func (c *TraceabilityComment) Parse(cmt *sdom.Opener, ctx *sdom.BracketContext) bool {
	d := ctx.Doc()
	if d == nil {
		return false
	}
	i := d.IndexOf(cmt)
	if i < 0 {
		return false
	}
	nodes := d.Nodes()
	cl := ctx.Closer(cmt)
	j := len(nodes)
	if cl != nil {
		j = d.IndexOf(cl)
	}
	if j-i != 2 {
		return false // an empty interior has no field; a re-granulated one is not ours
	}
	t, ok := nodes[i+1].(*sdom.Text)
	if !ok {
		return false
	}
	text, _ := t.Render()
	kids, ok := c.walk(text, t.Location())
	if !ok {
		return false
	}
	c.opener, c.closer, c.inner = cmt, cl, t
	c.assemble(kids)
	return true
}

// assemble tiles the node: opener, the interior's children, closer.
func (c *TraceabilityComment) assemble(interior []sdom.Node) {
	kids := append([]sdom.Node{c.opener}, interior...)
	if c.closer != nil {
		kids = append(kids, c.closer)
	}
	c.Compound = *sdom.NewCompound(c.opener.Location(), kids...)
}

// fieldNames are the segment groups, in the order they are consulted. A segment
// matches at most one of them, since segRe's branches are alternatives.
var fieldNames = []string{"crc", "seq", "test", "refs"}

// CRC: crc-TraceabilityComment.md | R338, R345
//
// parseField parses one group's text into its field node and points the matching
// typed field at it. False means the group's bytes are not wholly a list.
func (c *TraceabilityComment) parseField(name, str string, loc sdom.Loc) (sdom.Node, bool) {
	if name == "refs" {
		l, rest, ok := sdom.ParseRequirementList(str, loc)
		c.refs = l
		return l, ok && rest == ""
	}
	l, rest, ok := sdom.ParseList(str, loc)
	switch name {
	case "crc":
		c.crc = l
	case "seq":
		c.seq = l
	case "test":
		c.test = l
	}
	return l, ok && rest == ""
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#2.2 | R338, R339, R341, R342, R345
//
// walk is the segment walk over an interior — the parsing device order-independence
// forces, since one regex cannot express arbitrary order. It returns the children
// that tile text, or false when any byte would be left uncovered.
func (c *TraceabilityComment) walk(text string, loc sdom.Loc) ([]sdom.Node, bool) {
	sub := func(s, e int) sdom.Loc {
		if loc.Offset() < 0 {
			return sdom.Synthetic(e - s).In(loc.Origin())
		}
		return sdom.Source(loc.Offset()+s, e-s).In(loc.Origin())
	}
	at, sepLen := descSep(text)
	body := text[:at]

	var kids []sdom.Node
	seen := map[string]bool{}
	start := 0
	for start <= len(body) {
		end := strings.Index(body[start:], "|")
		if end < 0 {
			end = len(body)
		} else {
			end += start
		}
		b, ok := sdom.NewStencilBuilder(segRe, body[start:end], sub(start, end))
		if !ok {
			return nil, false
		}
		var field sdom.Node
		for _, name := range fieldNames {
			str, gl := b.Group(name)
			if str == "" {
				continue
			}
			if seen[name] {
				return nil, false
			}
			seen[name] = true
			f, ok := c.parseField(name, str, gl)
			if !ok {
				return nil, false
			}
			b.Put(name, f)
			field = f
		}
		if field == nil {
			return nil, false
		}
		seg, _ := b.Done()
		kids = append(kids, seg...)
		if end == len(body) {
			break
		}
		kids = append(kids, sdom.NewText("|", sub(end, end+1)))
		start = end + 1
	}
	if sepLen > 0 {
		kids = append(kids, sdom.NewText(text[at:at+sepLen], sub(at, at+sepLen)))
		c.desc = sdom.NewText(text[at+sepLen:], sub(at+sepLen, len(text)))
		kids = append(kids, c.desc)
	}
	return kids, true
}

// CRC: crc-TraceabilityComment.md | R339
//
// descSep finds the description separator: the first "--" or "—", or the first
// ":" that does not immediately follow a field keyword. Returns len(text), 0 when
// there is none.
func descSep(text string) (int, int) {
	for i := 0; i < len(text); i++ {
		switch {
		case strings.HasPrefix(text[i:], "--"):
			return i, 2
		case strings.HasPrefix(text[i:], "—"):
			return i, len("—")
		case text[i] == ':':
			before := text[:i]
			if strings.HasSuffix(before, "CRC") || strings.HasSuffix(before, "Seq") || strings.HasSuffix(before, "Test") {
				continue
			}
			return i, 1
		}
	}
	return len(text), 0
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#1 | R346
//
// Comments is the second pass. Every opener whose group kind equals the language's
// Comment.Kind — two configured strings compared, no word spelled here — is tried;
// each success replaces its run from opener to closer inside its own mutation
// window. Nothing else in the document is touched.
func Comments(d *sdom.Doc, ctx *sdom.BracketContext) ([]*TraceabilityComment, error) {
	lang := ctx.Language()
	kind := lang.Comment.Kind
	if kind == "" {
		return nil, nil
	}
	var cands []*sdom.Opener
	for _, n := range d.Nodes() {
		if o, ok := n.(*sdom.Opener); ok {
			s, _ := o.Render()
			if g := lang.GroupFor(s); g != nil && g.Kind == kind {
				cands = append(cands, o)
			}
		}
	}
	var out []*TraceabilityComment
	for _, o := range cands {
		c := &TraceabilityComment{}
		if !c.Parse(o, ctx) {
			continue
		}
		err := d.Mutate(func() error {
			if err := d.Replace(o, c); err != nil {
				return err
			}
			if err := d.Remove(c.inner); err != nil {
				return err
			}
			if c.closer != nil {
				return d.Remove(c.closer)
			}
			return nil
		})
		if err != nil {
			return out, err
		}
		out = append(out, c)
	}
	return out, nil
}

// CRC: crc-TraceabilityComment.md | R347
//
// Fields is what a constructed comment carries. Nil slices and an empty
// Description are absent fields.
type Fields struct {
	CRC, Seq, Test []string
	Refs           []int
	Description    string
}

// CRC: crc-TraceabilityComment.md | R347
//
// New assembles the canonical interior — CRC, Seq, Test, refs, "--" description —
// and runs the SAME walk Parse runs over it, at a synthetic location inside
// synthetic markers from lang.Comment. No child list is hand-built, so nothing can
// construct a comment that disagrees with how one is read. The node has no
// origin, which is what lets it enter any document.
func New(lang *sdom.BracketLang, f Fields) *TraceabilityComment {
	var parts []string
	if len(f.CRC) > 0 {
		parts = append(parts, "CRC: "+strings.Join(f.CRC, ", "))
	}
	if len(f.Seq) > 0 {
		parts = append(parts, "Seq: "+strings.Join(f.Seq, ", "))
	}
	if len(f.Test) > 0 {
		parts = append(parts, "Test: "+strings.Join(f.Test, ", "))
	}
	if len(f.Refs) > 0 {
		parts = append(parts, sdom.RequirementText(f.Refs))
	}
	body := strings.Join(parts, " | ")
	if f.Description != "" {
		body += " -- " + f.Description
	}

	cs := lang.Comment
	opener, closer := markers(lang)
	interior := cs.Prefix[len(opener):] + body + cs.Suffix[:len(cs.Suffix)-len(closer)]

	c := &TraceabilityComment{}
	kids, ok := c.walk(interior, sdom.Synthetic(len(interior)))
	if !ok {
		panic(fmt.Sprintf("minispecsdom: a canonical comment did not parse: %q", interior))
	}
	c.opener = sdom.NewOpener(opener, sdom.Synthetic(len(opener)))
	if closer != "" {
		c.closer = sdom.NewCloser(closer, sdom.Synthetic(len(closer)))
	}
	c.assemble(kids)
	return c
}

// markers finds the opener the style's Prefix starts with and the closer its Suffix
// ends with, in the group of the style's kind. The longest opener wins, and the
// closer always comes from that opener's own group.
func markers(lang *sdom.BracketLang) (opener, closer string) {
	cs := lang.Comment
	for _, g := range lang.Brackets {
		if g.Kind != cs.Kind {
			continue
		}
		for _, o := range g.Open {
			if !strings.HasPrefix(cs.Prefix, o) || len(o) <= len(opener) {
				continue
			}
			opener, closer = o, ""
			if strings.HasSuffix(cs.Suffix, g.Close) {
				closer = g.Close
			}
		}
	}
	return
}
