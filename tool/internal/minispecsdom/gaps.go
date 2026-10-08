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

// CRC: crc-Gaps.md | R430, R550
//
// The shapes a gap entry takes: a bullet at any depth, an optional checkbox, a type letter,
// a number, a colon. A backtick is written \x60 where one is needed.
var (
	gapHeadRe = regexp.MustCompile(`^(\s*)- (?:\[([ x])\] )?([SRDCIOAT])(\d+):[ \t]*(.*)$`)
	bulletRe  = regexp.MustCompile(`^(\s*)- (.*)$`)
	subHeadRe = regexp.MustCompile(`^###+\s+(.*?)\s*#*\s*$`)
	gapKeyRe  = regexp.MustCompile(`^([SRDCIOAT])(\d+)$`)
)

// CRC: crc-Gaps.md | R434, R435, R550
var (
	ErrNoSection = errors.New("minispecsdom: the document has no `## Gaps` section")
	ErrNoGap     = errors.New("minispecsdom: no gap carries that ID")
	ErrBadGapID  = errors.New("minispecsdom: a gap ID is a type letter and a number, like O81")
	ErrGapExists = errors.New("minispecsdom: the section already carries that gap ID")
	ErrPermanent = errors.New("minispecsdom: a permanent gap (A, T) is neither resolved nor approved")
	ErrResolved  = errors.New("minispecsdom: the gap is already resolved")
)

// CRC: crc-Gaps.md | R429
//
// Gaps is the gaps schema: it embeds the markdown base, owns the document, and reads the
// `## Gaps` region as typed entries.
type Gaps struct {
	markdownDoc

	heading    *schema.Heading
	sectionEnd int // one past the `## Gaps` section's last byte
	items      []*Gap
	unread     []Unread
	// regions are the section's head and then each `###` subsection, in order: where each
	// one's last non-blank line ends, the point `Add` places into. R550
	regions []gapRegion
}

// gapRegion is the section head (no title) or one `###` subsection. R550
type gapRegion struct {
	title string
	end   int // one past its last non-blank line
}

// CRC: crc-Gaps.md | R430, R431, R432
type Gap struct {
	ID                string
	Type              string
	Number            int
	Checkbox, Checked bool
	Text              string
	Sub               []string
	Depth             int
	Parent            *Gap

	line       int
	head       span // the head line
	body       span // the head line through its continuation and sub-items
	headText   string
	deviations []Deviation
}

// span is a byte range in the document, [start, end).
type span struct{ start, end int }

// CRC: crc-Gaps.md | R432
func (gap *Gap) Line() int { return gap.line }

// CRC: crc-Gaps.md | R432
func (gap *Gap) Permanent() bool { return permanentType(gap.Type) }

// CRC: crc-Gaps.md | R432
func (gap *Gap) Deviations() []Deviation { return gap.deviations }

func permanentType(letter string) bool { return letter == "A" || letter == "T" }

// CRC: crc-Gaps.md | Seq: seq-gaps.md#1 | R429
func ParseGaps(src string) *Gaps {
	g := &Gaps{}
	g.parse(src)
	return g
}

// parse reads src with the markdown base and derives the view over the result.
func (g *Gaps) parse(src string) {
	g.parseBase(src)
	g.scan()
}

// reload re-reads the document from its bytes, so the view and the array are rebuilt together.
func (g *Gaps) reload() {
	src, _ := g.doc.Render()
	g.parse(src)
}

func (g *Gaps) HasGaps() bool    { return g.heading != nil }
func (g *Gaps) Items() []*Gap    { return g.items }
func (g *Gaps) Unread() []Unread { return g.unread }

// CRC: crc-Gaps.md | R432
func (g *Gaps) Gap(id string) *Gap {
	for _, gap := range g.items {
		if gap.ID == id {
			return gap
		}
	}
	return nil
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#1.2 | R429, R430, R431, R432
// scan finds the region and reads its bullets.
func (g *Gaps) scan() {
	g.heading, g.items, g.unread, g.regions = nil, nil, nil, nil
	nodes := g.doc.Nodes()
	start := -1
	for i, n := range nodes {
		h, ok := n.(*schema.Heading)
		if ok && h.Level() == 2 && i+1 < len(nodes) && headingText(nodes[i+1]) == "Gaps" {
			g.heading, start = h, i
			break
		}
	}
	if start >= 0 {
		stop := g.regionEnd(start) // a node index; the section's own end is a byte offset
		g.sectionEnd = g.total
		if stop < len(nodes) {
			g.sectionEnd = nodes[stop].Location().Offset()
		}
		g.readItems(nodes[start:stop])
	}
	g.unread = append(g.unread, unbalanced(g.ctx)...)
	byLine(g.unread)
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#1.3 | R430, R431, R432
//
// readItems renders the region to lines and reads each bullet outside a code group: a keyed
// one is a gap, an indented un-keyed one is its parent's sub-item, and any other line folds
// into the open entry until a blank line or the region's end. The heading's own line is
// skipped.
func (g *Gaps) readItems(run []sdom.Node) {
	var b strings.Builder
	for _, n := range run {
		s, _ := n.Render()
		b.WriteString(s)
	}
	src := b.String()
	off := run[0].Location().Offset()
	seen := map[string]bool{}
	var cur *Gap // the entry whose body is open
	inSub := false
	g.regions = []gapRegion{{end: off + len(strings.SplitAfterN(src, "\n", 2)[0])}}
	closeBody := func(at int) {
		if cur != nil {
			cur.body.end = at
			cur = nil
		}
		inSub = false
	}
	for i, l := range strings.SplitAfter(src, "\n") {
		lineStart := off
		off += len(l)
		line := strings.TrimRight(l, "\n")
		if i == 0 || l == "" {
			continue
		}
		if strings.TrimSpace(line) == "" {
			closeBody(lineStart)
			continue
		}
		if g.inCode(lineStart) {
			g.regions[len(g.regions)-1].end = off
			continue
		}
		// Seq: seq-gaps.md#1.3.6 | R551
		// A subsection heading ends the entry above it and opens a region of its own.
		if m := subHeadRe.FindStringSubmatch(line); m != nil {
			closeBody(lineStart)
			g.regions = append(g.regions, gapRegion{title: m[1], end: off})
			continue
		}
		g.regions[len(g.regions)-1].end = off
		if m := gapHeadRe.FindStringSubmatch(line); m != nil {
			closeBody(lineStart)
			cur = newGap(m, g.doc.Line(lineStart), span{lineStart, off})
			g.record(cur, line, seen)
			continue
		}
		if m := bulletRe.FindStringSubmatch(line); m != nil {
			if cur != nil && len(m[1]) > cur.Depth {
				cur.Sub = append(cur.Sub, strings.TrimSpace(m[2]))
				inSub = true
				continue
			}
			closeBody(lineStart)
			g.unread = append(g.unread, Unread{g.doc.Line(lineStart), line})
			continue
		}
		switch {
		case inSub:
			last := len(cur.Sub) - 1
			cur.Sub[last] += " " + strings.TrimSpace(line)
		case cur != nil:
			cur.Text += " " + strings.TrimSpace(line)
		}
	}
	closeBody(g.sectionEnd)
}

// newGap reads a head line's match into an entry, whose body starts out as the head alone.
func newGap(m []string, line int, head span) *Gap {
	indent, box, letter, digits, text := m[1], m[2], m[3], m[4], strings.TrimSpace(m[5])
	number, _ := strconv.Atoi(digits)
	return &Gap{
		ID:       letter + digits,
		Type:     letter,
		Number:   number,
		Checkbox: box != "",
		Checked:  box == "x",
		Text:     text,
		Depth:    len(indent),
		line:     line,
		head:     head,
		body:     head,
		headText: text,
	}
}

// record places a new entry under its parent, notes the deviations its head line carries,
// and lists each of those unread.
func (g *Gaps) record(gap *Gap, line string, seen map[string]bool) {
	gap.Parent = g.parentOf(gap.Depth)
	switch {
	case gap.Permanent() && gap.Checkbox:
		gap.deviations = append(gap.deviations, Deviation{"a permanent gap (A, T) carries no checkbox", line})
	case !gap.Permanent() && !gap.Checkbox:
		gap.deviations = append(gap.deviations, Deviation{"a tracked gap carries `[ ]` or `[x]`", line})
	}
	if seen[gap.ID] {
		gap.deviations = append(gap.deviations, Deviation{"a gap ID appears once in the section", gap.ID})
	}
	seen[gap.ID] = true
	for _, d := range gap.deviations {
		g.unread = append(g.unread, Unread{gap.line, fmt.Sprintf("%s — %s: %s", gap.ID, d.Rule, d.Target)})
	}
	g.items = append(g.items, gap)
}

// parentOf is the nearest entry read so far that is shallower than depth, or nil at the
// top level.
func (g *Gaps) parentOf(depth int) *Gap {
	for i := len(g.items) - 1; i >= 0; i-- {
		if g.items[i].Depth < depth {
			return g.items[i]
		}
	}
	return nil
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#2.1 | R432, R436
// writable finds the gap a write addresses and decides its refusal before any byte moves.
func (g *Gaps) writable(id string) (*Gap, error) {
	gap := g.Gap(id)
	if gap == nil {
		return nil, ErrNoGap
	}
	if len(gap.deviations) > 0 {
		return nil, &DeviationError{Key: id, Deviations: gap.deviations}
	}
	return gap, nil
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#2.2 | R540, R550
func (g *Gaps) Add(id, text string) error {
	m := gapKeyRe.FindStringSubmatch(id)
	if m == nil {
		return ErrBadGapID
	}
	if g.heading == nil {
		return ErrNoSection
	}
	if g.Gap(id) != nil {
		return ErrGapExists
	}
	permanent := permanentType(m[1])
	box := "[ ] "
	if permanent {
		box = ""
	}
	text = strings.TrimSpace(text)
	line := "- " + box + id + ": " + text + "\n"
	at := g.insertionPoint(m[1])
	if err := g.doc.Mutate(func() error { return g.insertLine(at, line) }); err != nil {
		return err
	}
	g.reload()
	got := g.Gap(id)
	ok := got != nil && got.Text == text && got.Checkbox == !permanent && len(got.deviations) == 0
	mustReadBack("Gaps", "Add", id, ok, line, fmt.Sprintf("%+v", got))
	return nil
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#2.2.1 | R550
// insertionPoint is where `Add` places an entry of the letter. With no subsections, after the
// last gap's span or the heading; with them, after the last gap of the same letter, else at
// the end of the subsection whose heading names the type, else at the end of the head.
func (g *Gaps) insertionPoint(letter string) int {
	if len(g.regions) <= 1 {
		at := g.afterHeading()
		for _, gap := range g.items { // the last span in the document, not the last read
			at = max(at, gap.body.end)
		}
		return at
	}
	at := -1
	for _, gap := range g.items {
		if gap.Type == letter {
			at = max(at, gap.body.end)
		}
	}
	if at >= 0 {
		return at
	}
	for _, r := range g.regions[1:] {
		if namesType(r.title, letter) {
			return r.end
		}
	}
	return g.regions[0].end
}

// typeNames are the standard subsection names, in normalised form. R550
var typeNames = map[string]string{
	"S": "spec→requirements", "R": "requirements→design", "D": "design→code",
	"C": "code→design", "I": "incomplete implementation", "O": "oversights",
	"A": "approved", "T": "retired",
}

var typeSuffixRe = regexp.MustCompile(`\(([SRDCIOAT])n\)`)

// CRC: crc-Gaps.md | R550
// namesType reports whether a subsection title names the gap type: a `(Xn)` suffix, or the
// standard name with arrows written `→` or `->`, spaced or not, case ignored, and a trailing
// word such as `Gaps` allowed.
func namesType(title, letter string) bool {
	if m := typeSuffixRe.FindStringSubmatch(title); m != nil {
		return m[1] == letter
	}
	t := strings.ToLower(title)
	t = strings.ReplaceAll(t, "->", "→")
	t = strings.ReplaceAll(t, " →", "→")
	t = strings.ReplaceAll(t, "→ ", "→")
	name := typeNames[letter]
	return t == name || strings.HasPrefix(t, name+" ")
}

// afterHeading is the offset just past the heading's line.
func (g *Gaps) afterHeading() int {
	off := g.heading.Location().Offset()
	src, _ := g.doc.Render()
	if i := strings.IndexByte(src[off:], '\n'); i >= 0 {
		return off + i + 1
	}
	return g.total
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#2.3 | R434
func (g *Gaps) Resolve(id string) error {
	gap, err := g.writable(id)
	if err != nil {
		return err
	}
	if gap.Permanent() {
		return ErrPermanent
	}
	if gap.Checked {
		return ErrResolved
	}
	line := strings.Replace(g.headLine(gap), "- [ ] ", "- [x] ", 1)
	if err := g.doc.Mutate(func() error { return g.replaceSpan(gap.head.start, gap.head.end, line) }); err != nil {
		return err
	}
	g.reload()
	got := g.Gap(id)
	ok := got != nil && got.Checked && got.headText == gap.headText
	mustReadBack("Gaps", "Resolve", id, ok, "[x]", fmt.Sprintf("%+v", got))
	return nil
}

// CRC: crc-Gaps.md | Seq: seq-gaps.md#2.4 | R435
func (g *Gaps) Approve(id, newID string) error {
	m := gapKeyRe.FindStringSubmatch(newID)
	if m == nil || m[1] != "A" {
		return ErrBadGapID
	}
	if g.Gap(newID) != nil {
		return ErrGapExists
	}
	gap, err := g.writable(id)
	if err != nil {
		return err
	}
	if gap.Permanent() {
		return ErrPermanent
	}
	line := strings.Repeat(" ", gap.Depth) + "- " + newID + ": " + gap.headText + "\n"
	if err := g.doc.Mutate(func() error { return g.replaceSpan(gap.head.start, gap.head.end, line) }); err != nil {
		return err
	}
	g.reload()
	got := g.Gap(newID)
	ok := got != nil && !got.Checkbox && got.headText == gap.headText && g.Gap(id) == nil
	mustReadBack("Gaps", "Approve", id, ok, line, fmt.Sprintf("%+v", got))
	return nil
}

// headLine is the head line's bytes, newline included.
func (g *Gaps) headLine(gap *Gap) string {
	src, _ := g.doc.Render()
	return src[gap.head.start:gap.head.end]
}
