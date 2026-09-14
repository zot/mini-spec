package minispecsdom

import (
	"errors"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// CRC: crc-PartLine.md | R351
//
// headRe cuts a keyed head's first text: the key, the em-dash separator, the title.
// `Item` is required without a dot and forbidden with one; the separator is the em dash
// and nothing else. Anything else leaves the line unkeyed, reported rather than refused.
var headRe = regexp.MustCompile(`(?s)^(?P<key>Item \d+|\d+(?:\.\d+)+)(?P<sep>\s*—\s*)(?P<title>.*)$`)

// looseHeadRe recognizes a head that keys in some OTHER scheme or separator, so the
// deviation can name the target rather than merely say "unkeyed".
var looseHeadRe = regexp.MustCompile(`^(?:Item\s+)?\d+(?:\.\d+)*\s*[—–-]`)

// CRC: crc-MarkerSpan.md | R352
//
// markerRe cuts a marker's first text: the verb — words joined by spaces or hyphens,
// matched without regard to case so a mis-cased one is read and reported — then either
// the opening parenthesis of an attribution or the end.
var markerRe = regexp.MustCompile(`(?is)^(?P<verb>[a-z][a-z -]*[a-z]|[a-z])\b(?P<rest>\s*\(.*|\.?)$`)

// canonicalRe is what a written marker interior must read back as.
var canonicalRe = regexp.MustCompile(`(?s)^([A-Z][A-Z -]*[A-Z]|[A-Z]) \((.*)\)$`)

var queueIDRe = regexp.MustCompile(`#(\d+)`)

// checkboxTextRe recognizes a bracketed interior the base did not take as a Checkbox,
// so an off-form one is reported rather than passed over.
var checkboxTextRe = regexp.MustCompile(`^\[.\]`)

// CRC: crc-MarkerSpan.md | R397
//
// queuedRe and notQueuedRe are the two `OPEN` attributions, read flexibly — any case,
// any inner spacing, an optional full stop — while Set writes one canonical form.
var (
	queuedRe    = regexp.MustCompile(`^#\d+\.?$`)
	notQueuedRe = regexp.MustCompile(`(?i)^not\s+queued\.?$`)
)

// CRC: crc-PartLine.md | R398
//
// commaSchemeRe recognizes the superseded comma form of a marker — `OPEN, not queued.` —
// which is not a marker at all and would otherwise go unreported.
var commaSchemeRe = regexp.MustCompile(`(?is)^[a-z][a-z -]*,`)

// ErrMarkerRefused reports a MarkerSpan.Set whose render would not read back as the
// verb and attribution that were set. The literal is unchanged.
var ErrMarkerRefused = errors.New("minispecsdom: the marker would not read back; unchanged")

// CRC: crc-PartLine.md | R355
//
// Deviation names a rule a line breaks and the shape it must take, per
// trajectory-format.md: a read path lists it and a write path refuses on it.
type Deviation struct{ Rule, Target string }

const (
	keyTarget      = "a part is keyed `Item N` and a subpart `N.M`, separated from its title by an em dash"
	checkboxTarget = "`[ ]` for open, `[x]` for landed"
	verbTarget     = "a marker verb is written in capitals"
	markerTarget   = "an `OPEN` attribution is `#N` or `not queued`"
	revertedTarget = "a `REVERTED` attribution is `#N`"
	schemeTarget   = "a marker is `**VERB (attribution)**`; the comma form `**OPEN, not queued.**` is superseded"
)

// CRC: crc-PartLine.md | R349, R353, R359
//
// PartLine is a carve status line as one node: bound checkbox and key, a list of
// marker spans, the strike behind an accessor. It tiles from the `- ` to the byte
// before the newline, reusing every node the base emitted.
type PartLine struct {
	sdom.Compound

	item     *schema.ListItem
	checkbox *schema.Checkbox
	key      *sdom.Text
	title    *sdom.Text
	headOpen *sdom.Opener // the head's `**`, so Strike can find it
	markers  []*MarkerSpan
	devs     []Deviation

	run  []sdom.Node // the flat nodes this replaces, item first
	tail *sdom.Text  // the run's last text, which holds the newline
	cut  int         // where the line ends inside tail; 0 when tail is all next line
}

// CRC: crc-PartLine.md | R353
func (p *PartLine) Checkbox() *schema.Checkbox { return p.checkbox }

// CRC: crc-PartLine.md | R351
func (p *PartLine) Key() string {
	if p.key == nil {
		return ""
	}
	s, _ := p.key.Render()
	// R417: one key form, the fragment — `4` for a part, `2.2` for a subpart — which is what
	// `carves/x.md#<key>` carries. `Item ` is the head's display word, not part of the key.
	return strings.TrimPrefix(s, "Item ")
}

// CRC: crc-PartLine.md | R351
func (p *PartLine) Title() *sdom.Text { return p.title }

// CRC: crc-PartLine.md | R352
func (p *PartLine) Markers() []*MarkerSpan { return p.markers }

// CRC: crc-PartLine.md | R355
func (p *PartLine) Deviations() []Deviation { return p.devs }

// CRC: crc-PartLine.md | R359
func (p *PartLine) Equals(o sdom.Node) bool {
	x, ok := o.(*PartLine)
	return ok && p.Compound.Equals(&x.Compound)
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#1 | R349, R350, R351, R352, R355, R360
//
// Parse walks the line from item to the newline, classifies each bold run by content
// — the first is the head, a later `VERB (…)` run a marker — re-cuts the head's first
// text into key, separator and title, and keeps everything else as it was. Every list
// item parses; false only when item is not in the document.
func (p *PartLine) Parse(item *schema.ListItem, ctx *sdom.BracketContext) bool {
	d := ctx.Doc()
	if d == nil {
		return false
	}
	i := d.IndexOf(item)
	if i < 0 {
		return false
	}
	nodes := d.Nodes()
	p.item = item
	p.run = []sdom.Node{item}
	kids := []sdom.Node{item}

	j := i + 1
	if j < len(nodes) {
		if cb, ok := nodes[j].(*schema.Checkbox); ok {
			p.checkbox = cb
			kids = append(kids, cb)
			p.run = append(p.run, cb)
			j++
		} else if t, ok := nodes[j].(*sdom.Text); ok {
			if s, _ := t.Render(); checkboxTextRe.MatchString(s) {
				p.devs = append(p.devs, Deviation{"checkbox interior", checkboxTarget})
			}
		}
	}

	for ; j < len(nodes); j++ {
		n := nodes[j]
		if t, ok := n.(*sdom.Text); ok {
			s, _ := t.Render()
			if at := strings.IndexByte(s, '\n'); at >= 0 {
				p.tail, p.cut = t, at
				if at > 0 {
					kids = append(kids, cutText(t, 0, at))
				}
				break
			}
			kids = append(kids, t)
			p.run = append(p.run, t)
			continue
		}
		p.run = append(p.run, n)
		o, cl := boldRun(n, ctx)
		if o == nil {
			kids = append(kids, n)
			continue
		}
		// A bold run: opener, interior nodes, closer — all reused.
		end := d.IndexOf(cl)
		interior := nodes[j+1 : end]
		p.run = append(p.run, interior...)
		p.run = append(p.run, cl)
		if p.headOpen == nil {
			// The first bold run is the head.
			p.headOpen = o
			kids = append(kids, o)
			kids = append(kids, p.cutHead(interior)...)
			kids = append(kids, cl)
		} else if m := newMarkerSpan(o, interior, cl); m != nil {
			p.markers = append(p.markers, m)
			p.devs = append(p.devs, m.deviations()...)
			kids = append(kids, m)
		} else {
			if commaSchemeRe.MatchString(firstText(interior)) {
				p.devs = append(p.devs, Deviation{"marker scheme", schemeTarget})
			}
			kids = append(kids, o)
			kids = append(kids, interior...)
			kids = append(kids, cl)
		}
		j = end
	}
	if p.key == nil {
		p.devs = append(p.devs, Deviation{"key form", keyTarget})
	}
	p.Compound = *sdom.NewCompound(item.Location(), kids...)
	return true
}

// boldRun reports the bold run a node begins: the `**` opener and the closer that
// pairs with it. Anything else — another marker, an unpaired `**`, a node that is not
// an opener at all — is not a run, and both are nil.
func boldRun(n sdom.Node, ctx *sdom.BracketContext) (*sdom.Opener, *sdom.Closer) {
	o, ok := n.(*sdom.Opener)
	if !ok {
		return nil, nil
	}
	if s, _ := o.Render(); s != "**" {
		return nil, nil
	}
	cl := ctx.Closer(o)
	if cl == nil {
		return nil, nil
	}
	return o, cl
}

// firstText renders the first node of a run when it is a Text, else "".
func firstText(nodes []sdom.Node) string {
	if len(nodes) == 0 {
		return ""
	}
	t, ok := nodes[0].(*sdom.Text)
	if !ok {
		return ""
	}
	s, _ := t.Render()
	return s
}

// cutHead re-cuts the head's first text into key, separator glue and title, or leaves
// the interior as glue when it does not key.
func (p *PartLine) cutHead(interior []sdom.Node) []sdom.Node {
	if len(interior) == 0 {
		return nil
	}
	t, ok := interior[0].(*sdom.Text)
	if !ok {
		return interior
	}
	s, _ := t.Render()
	b, ok := sdom.NewStencilBuilder(headRe, s, t.Location())
	if !ok {
		if looseHeadRe.MatchString(s) {
			p.devs = append(p.devs, Deviation{"separator", keyTarget})
		}
		return interior
	}
	p.key = sdom.NewText(b.Group("key"))
	b.Put("key", p.key)
	b.Omit("sep")
	p.title = sdom.NewText(b.Group("title"))
	b.Put("title", p.title)
	kids, _ := b.Done()
	return append(kids, interior[1:]...)
}

// headIndex is where the head's bold opener sits among kids, or -1.
func (p *PartLine) headIndex(kids []sdom.Node) int {
	for i, k := range kids {
		if k == sdom.Node(p.headOpen) {
			return i
		}
	}
	return -1
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#2 | R354
// IsStruck derives from the children: the head's bold opener sits right after a `~~`.
func (p *PartLine) IsStruck() bool {
	kids := p.Kids()
	i := p.headIndex(kids)
	if i <= 0 {
		return false
	}
	o, ok := kids[i-1].(*sdom.Opener)
	if !ok {
		return false
	}
	s, _ := o.Render()
	return s == "~~"
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#2 | R354
//
// Strike wraps or unwraps the head's bold run in `~~`, among this node's own children.
// The flat array never sees inside a compound, so no mutation window is involved. A
// line with no head has nothing to strike and is left alone.
func (p *PartLine) Strike(on bool) {
	if p.headOpen == nil || p.IsStruck() == on {
		return
	}
	kids := p.Kids()
	open := p.headIndex(kids)
	shut := open + 1
	for ; shut < len(kids); shut++ {
		if c, ok := kids[shut].(*sdom.Closer); ok {
			if s, _ := c.Render(); s == "**" {
				break
			}
		}
	}
	head := kids[open : shut+1] // the head's bold run, opener through closer
	var out []sdom.Node
	if on {
		out = append(out, kids[:open]...)
		out = append(out, sdom.NewOpener("~~", sdom.Synthetic(2)))
		out = append(out, head...)
		out = append(out, sdom.NewCloser("~~", sdom.Synthetic(2)))
		out = append(out, kids[shut+1:]...)
	} else {
		out = append(out, kids[:open-1]...)
		out = append(out, head...)
		out = append(out, kids[shut+2:]...)
	}
	p.Compound = *sdom.NewCompound(p.item.Location(), out...)
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#1.5 | R358
//
// Splice replaces the line's run with this node, inside a mutation window. The run's
// last text holds the newline, so it is split first and its left half moves inside.
func (p *PartLine) Splice(d *sdom.Doc) error {
	if p.tail != nil && p.cut > 0 {
		left, _, err := d.Split(p.tail, p.cut)
		if err != nil {
			return err
		}
		kids := p.Kids()
		kids[len(kids)-1] = left
		p.Compound = *sdom.NewCompound(p.item.Location(), kids...)
		p.run = append(p.run, left)
	}
	if err := d.Replace(p.item, p); err != nil {
		return err
	}
	for _, n := range p.run[1:] {
		if err := d.Remove(n); err != nil {
			return err
		}
	}
	return nil
}

// CRC: crc-PartLine.md | R358
//
// PartLines parses and splices every list item in document order, each in its own
// mutation window.
func PartLines(d *sdom.Doc, ctx *sdom.BracketContext) ([]*PartLine, error) {
	return partLines(d, ctx, func(*schema.ListItem) bool { return true })
}

// partLines is PartLines over the list items keep admits — the carve schema keeps
// only those inside its status region.
func partLines(d *sdom.Doc, ctx *sdom.BracketContext, keep func(*schema.ListItem) bool) ([]*PartLine, error) {
	var items []*schema.ListItem
	for _, n := range d.Nodes() {
		if it, ok := n.(*schema.ListItem); ok && keep(it) {
			items = append(items, it)
		}
	}
	var out []*PartLine
	for _, it := range items {
		p := &PartLine{}
		if !p.Parse(it, ctx) {
			continue
		}
		if err := d.Mutate(func() error { return p.Splice(d) }); err != nil {
			return out, err
		}
		out = append(out, p)
	}
	return out, nil
}

// CRC: crc-PartLine.md | Seq: seq-carve.md#2.3.1 | R391, R392
//
// refuse is the deviation guard every write path runs first: a line carrying deviations
// takes no write, and the error names each rule and its target.
func (p *PartLine) refuse() error {
	if len(p.devs) == 0 {
		return nil
	}
	return &DeviationError{Key: p.Key(), Deviations: p.devs}
}

// CRC: crc-PartLine.md | Seq: seq-carve.md#2.3 | R409, R391, R392, R393
//
// SetMarker applies the tool's rule among this node's children: replace the first
// TRANSIENT marker — verb OPEN or REVERTED — with the new one, remove any other
// transient, and insert a marker before the trailing prose when the line carries none.
// It selects by the kind of what it replaces, never by what it writes, so a record set
// over a line that also carries NOT VERIFIED leaves that standing. It refuses before it
// touches anything: over deviations, and OPEN over a checked box.
func (p *PartLine) SetMarker(verb, attribution string) error {
	if err := p.refuse(); err != nil {
		return err
	}
	if strings.EqualFold(verb, "OPEN") && p.checkbox != nil && p.checkbox.Checked() { // Seq: seq-carve.md#2.3.1
		return ErrReopen
	}
	var first *MarkerSpan
	var drop []*MarkerSpan
	for _, m := range p.markers {
		if !m.isTransient() { // R409: OPEN and REVERTED are what a write replaces
			continue
		}
		if first == nil {
			first = m
			continue
		}
		drop = append(drop, m)
	}
	if first != nil {
		if err := first.Set(verb, attribution); err != nil {
			return err
		}
	}
	var out []sdom.Node
	for _, k := range p.Kids() {
		if m, ok := k.(*MarkerSpan); ok && slices.Contains(drop, m) {
			out = dropGlue(out) // so the markers either side do not run together
			continue
		}
		out = append(out, k)
	}
	if first == nil {
		m := newMarker(verb, attribution)
		if m == nil {
			return ErrMarkerRefused
		}
		out = p.insertMarker(out, m) // R409: before trailing prose, never after it
		p.markers = append(p.markers, m)
	}
	p.markers = slices.DeleteFunc(p.markers, func(m *MarkerSpan) bool { return slices.Contains(drop, m) })
	p.Compound = *sdom.NewCompound(p.item.Location(), out...)
	return nil
}

// CRC: crc-PartLine.md | Seq: seq-carve.md#2.3 | R409
//
// insertMarker places a new marker where the grammar `Head Marker* Text?` puts it: after
// the head and any markers, before the trailing prose — which is the first text past the
// head's closer with anything in it but whitespace. With no such text the marker ends the
// line, as before.
func (p *PartLine) insertMarker(kids []sdom.Node, m *MarkerSpan) []sdom.Node {
	ins := []sdom.Node{sdom.NewText(" ", sdom.Synthetic(1)), m}
	at := p.proseAt(kids)
	if at < 0 {
		return append(kids, ins...)
	}
	if s, _ := kids[at].Render(); !strings.HasPrefix(s, " ") {
		ins = append(ins, sdom.NewText(" ", sdom.Synthetic(1))) // the prose brings none of its own
	}
	return slices.Insert(kids, at, ins...)
}

// proseAt is the index of the trailing prose among kids — the first text after the head's
// bold run that is not all whitespace — or -1 when the line ends in its head or markers.
func (p *PartLine) proseAt(kids []sdom.Node) int {
	pastHead := false
	for i, k := range kids {
		switch n := k.(type) {
		case *sdom.Closer:
			if s, _ := n.Render(); s == "**" {
				pastHead = true
			}
		case *sdom.Text:
			if s, _ := n.Render(); pastHead && strings.TrimSpace(s) != "" {
				return i
			}
		}
	}
	return -1
}

// newMarker builds a synthetic marker span the way Set writes one, or nil when the
// verb and attribution would not read back.
func newMarker(verb, attribution string) *MarkerSpan {
	open := sdom.NewOpener("**", sdom.Synthetic(2))
	cl := sdom.NewCloser("**", sdom.Synthetic(2))
	m := &MarkerSpan{Compound: *sdom.NewCompound(open.Location(), open, cl)}
	if err := m.Set(verb, attribution); err != nil {
		return nil
	}
	return m
}

// dropGlue drops a trailing whitespace text: the glue that separated a marker now gone.
func dropGlue(nodes []sdom.Node) []sdom.Node {
	n := len(nodes)
	if n == 0 {
		return nodes
	}
	t, ok := nodes[n-1].(*sdom.Text)
	if !ok {
		return nodes
	}
	if s, _ := t.Render(); strings.TrimSpace(s) != "" {
		return nodes
	}
	return nodes[:n-1]
}

// CRC: crc-MarkerSpan.md | R356, R359
//
// MarkerSpan is `**VERB (attribution)**` as a stencil over the bold run the base
// emitted: the verb bound, the attribution derived across whatever nodes it spans.
type MarkerSpan struct {
	sdom.Compound

	verb     *sdom.Text
	attrFrom int // index in kids of the first attribution node
	attrTo   int // index in kids just past the last attribution node
}

// newMarkerSpan classifies a bold run by content and builds the span, or nil.
//
// Only the first text and the last are re-cut — into the verb, the glue that carries
// the parentheses, and the attribution's ends. Everything between them, code spans
// included, is the run's own nodes, reused.
func newMarkerSpan(o *sdom.Opener, interior []sdom.Node, cl *sdom.Closer) *MarkerSpan {
	if len(interior) == 0 {
		return nil
	}
	first, ok := interior[0].(*sdom.Text)
	if !ok {
		return nil
	}
	head, _ := first.Render()
	m := markerRe.FindStringSubmatch(head)
	if m == nil {
		return nil
	}
	verb := m[1]

	ms := &MarkerSpan{verb: cutText(first, 0, len(verb))}
	kids := []sdom.Node{o, ms.verb}
	at := strings.IndexByte(head[len(verb):], '(')
	if at < 0 {
		// No attribution: the rest of the run is glue and the span is empty.
		if len(verb) < len(head) {
			kids = append(kids, cutText(first, len(verb), len(head)))
		}
		kids = append(kids, interior[1:]...)
		ms.attrFrom, ms.attrTo = len(kids), len(kids)
	} else {
		last, ok := interior[len(interior)-1].(*sdom.Text)
		if !ok {
			return nil // the run ends inside something else: not a marker
		}
		open := len(verb) + at
		kids = append(kids, cutText(first, len(verb), open+1)) // glue through "("
		ms.attrFrom = len(kids)
		// The attribution starts just past "(" and ends at the last ")" in the run's
		// last text — the same text when the run is one, otherwise past the middle.
		from := open + 1
		if len(interior) > 1 {
			if from < len(head) {
				kids = append(kids, cutText(first, from, len(head)))
			}
			kids = append(kids, interior[1:len(interior)-1]...)
			from = 0
		}
		lastText, _ := last.Render()
		shut := strings.LastIndexByte(lastText[from:], ')')
		if shut < 0 {
			return nil
		}
		shut += from
		if shut > from {
			kids = append(kids, cutText(last, from, shut))
		}
		ms.attrTo = len(kids)
		kids = append(kids, cutText(last, shut, len(lastText))) // ")" and any glue after
	}
	kids = append(kids, cl)
	ms.Compound = *sdom.NewCompound(o.Location(), kids...)
	return ms
}

// CRC: crc-MarkerSpan.md | R356
func (m *MarkerSpan) Verb() *sdom.Text { return m.verb }

// CRC: crc-MarkerSpan.md | R356
// Attribution renders the nodes between the parentheses; "" when there are none.
func (m *MarkerSpan) Attribution() string {
	var b strings.Builder
	for _, k := range m.Kids()[m.attrFrom:m.attrTo] {
		s, _ := k.Render()
		b.WriteString(s)
	}
	return b.String()
}

// CRC: crc-MarkerSpan.md | R356
func (m *MarkerSpan) QueueID() (int, bool) {
	sm := queueIDRe.FindStringSubmatch(m.Attribution())
	if sm == nil {
		return 0, false
	}
	n, _ := strconv.Atoi(sm[1])
	return n, true
}

// isOpen reports the OPEN marker, whatever its case.
func (m *MarkerSpan) isOpen() bool {
	v, _ := m.verb.Render()
	return strings.EqualFold(v, "OPEN")
}

// isReverted reports the REVERTED marker, whatever its case.
func (m *MarkerSpan) isReverted() bool {
	v, _ := m.verb.Render()
	return strings.EqualFold(v, "REVERTED")
}

// CRC: crc-PartLine.md | R409
// isTransient reports a marker a write replaces: OPEN, and REVERTED — what a revert writes
// over `OPEN (#N)` and what a replay writes `OPEN (#N)` back over.
func (m *MarkerSpan) isTransient() bool { return m.isOpen() || m.isReverted() }

// deviations reports the marker's own rules: verb case, and the OPEN and REVERTED
// attributions.
func (m *MarkerSpan) deviations() []Deviation {
	var out []Deviation
	v, _ := m.verb.Render()
	if v != strings.ToUpper(v) {
		out = append(out, Deviation{"verb case", verbTarget})
	}
	if m.isReverted() && !queuedRe.MatchString(m.Attribution()) {
		out = append(out, Deviation{"REVERTED attribution", revertedTarget}) // R410
	}
	if m.isOpen() {
		if a := m.Attribution(); !queuedRe.MatchString(a) && !notQueuedRe.MatchString(a) {
			out = append(out, Deviation{"OPEN attribution", markerTarget})
		}
	}
	return out
}

// CRC: crc-MarkerSpan.md | Seq: seq-partline.md#3 | R357
//
// Set rewrites the interior canonically — VERB (attribution), verb in capitals — under
// the guarded write: the render must read back as the same verb and attribution, or
// the write is refused and the literal unchanged. The result is four texts where a
// fresh parse might give code spans; bytes identical, structure not — the corner a
// canonical write accepts.
func (m *MarkerSpan) Set(verb, attribution string) error {
	verb = strings.ToUpper(verb)
	canon := verb + " (" + attribution + ")"
	sm := canonicalRe.FindStringSubmatch(canon)
	if sm == nil || sm[1] != verb || sm[2] != attribution {
		return ErrMarkerRefused
	}
	kids := m.Kids()
	open, cl := kids[0], kids[len(kids)-1]
	m.verb = sdom.NewText(verb, sdom.Synthetic(len(verb)))
	attr := sdom.NewText(attribution, sdom.Synthetic(len(attribution)))
	out := []sdom.Node{open, m.verb, sdom.NewText(" (", sdom.Synthetic(2)), attr, sdom.NewText(")", sdom.Synthetic(1)), cl}
	m.attrFrom, m.attrTo = 3, 4 // attr's place in out
	m.Compound = *sdom.NewCompound(open.Location(), out...)
	return nil
}

// CRC: crc-MarkerSpan.md | R359
func (m *MarkerSpan) Equals(o sdom.Node) bool {
	x, ok := o.(*MarkerSpan)
	return ok && m.Compound.Equals(&x.Compound)
}

// cutText returns a new Text over t's bytes [s, e), with the matching location.
func cutText(t *sdom.Text, s, e int) *sdom.Text {
	str, _ := t.Render()
	loc := t.Location()
	if loc.Offset() < 0 {
		return sdom.NewText(str[s:e], sdom.Synthetic(e-s).In(loc.Origin()))
	}
	return sdom.NewText(str[s:e], sdom.Source(loc.Offset()+s, e-s).In(loc.Origin()))
}
