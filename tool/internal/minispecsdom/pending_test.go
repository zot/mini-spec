package minispecsdom

import (
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
)

func pendingFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/pending-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func ids(p *Pending) []int {
	var out []int
	for _, e := range p.Entries() {
		out = append(out, e.ID)
	}
	return out
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1 | R370, R371, R372, R376
func TestTheFixturesEntriesReadBack(t *testing.T) {
	src := pendingFixture(t)
	p := ParsePending(src)
	if r, _ := p.Render(); r != src {
		t.Fatal("render differs from source")
	}
	if got := ids(p); !slices.Equal(got, []int{8, 12, 3, 14, 15}) {
		t.Fatalf("ids %v, want [8 12 3 14 15]", got)
	}
	e := p.Entry(8)
	if e.Title != "The reader" || e.Skill != "mini-spec" || e.Status != "Design settled; ready to build" ||
		e.SourceDoc != "carves/x.md" || e.SourceKey != "7" || e.Next != "write the spec." {
		t.Errorf("entry 8: %+v", *e)
	}
	if e := p.Entry(12); e.Skill != "" || e.SourceKey != "2.1" || e.Next != "" || e.Status != "Parked with context" {
		t.Errorf("entry 12: %+v", *e)
	}
	if p.MaxID() != 15 {
		t.Errorf("MaxID %d", p.MaxID())
	}
	if u := p.Unread(); len(u) != 2 || u[0] != (Unread{25, "Notes"}) ||
		u[1] != (Unread{33, "Source: [design/design.md](design/design.md), gap `O1-O3`."}) {
		t.Errorf("unread %+v", u)
	}
	// R395: every entry knows its line, 1-based.
	if p.Entry(8).Line() != 8 || p.Entry(12).Line() != 12 || p.Entry(3).Line() != 19 {
		t.Errorf("lines %d %d %d", p.Entry(8).Line(), p.Entry(12).Line(), p.Entry(3).Line())
	}
	// Entry 12's region holds its ### sub-item; entry 3's holds its fence.
	if !strings.Contains(runText(p.Entry(12)), "### Parked context") {
		t.Errorf("the sub-item is not inside entry 12's region")
	}
	if !strings.Contains(runText(p.Entry(3)), "## 9. **not an entry**") {
		t.Errorf("the fenced heading is not inside entry 3's region")
	}
}

// runText renders an entry's region: its run, cut where a rule ends it.
func runText(e *Entry) string {
	var b strings.Builder
	for i, n := range e.run {
		s, _ := n.Render()
		if i == len(e.run)-1 && e.tail != nil {
			s = s[:e.cut]
		}
		b.WriteString(s)
	}
	return b.String()
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2 | R373, R374, R377
func TestPlaceByPositionRefusedNotClamped(t *testing.T) {
	p := ParsePending(pendingFixture(t))
	e := EntryText{ID: 20, Title: "New", Status: "Fresh.", SourceDoc: "carves/z.md", SourceKey: "1"}
	if err := p.Place(e, 1); err != nil {
		t.Fatal(err)
	}
	if got := ids(p); !slices.Equal(got, []int{20, 8, 12, 3, 14, 15}) {
		t.Fatalf("after Place at 1: %v, want [20 8 12 3 14 15]", got)
	}
	r, _ := p.Render()
	if !strings.Contains(r, "---\n\n## 20. **New**. Fresh.\nSource: [carves/z.md](carves/z.md), part `#1`.\n\n## 8. ") {
		t.Errorf("placed at 1:\n%s", r)
	}
	pos, err := p.After(12)
	if err != nil || pos != 4 {
		t.Fatalf("After(12) = %d, %v", pos, err)
	}
	if err := p.Place(EntryText{ID: 21, Title: "After twelve", Status: "S.", SourceDoc: "d", SourceKey: "2"}, pos); err != nil {
		t.Fatal(err)
	}
	if err := p.Place(EntryText{ID: 22, Title: "Last", Status: "S.", SourceDoc: "d", SourceKey: "3"}, len(p.Entries())+1); err != nil {
		t.Fatal(err)
	}
	if got := ids(p); !slices.Equal(got, []int{20, 8, 12, 21, 3, 14, 15, 22}) {
		t.Fatalf("ids %v, want [20 8 12 21 3 14 15 22]", got)
	}
	r, _ = p.Render()
	if !strings.Contains(r, "kept here until resumed.\n\n## 21. **After twelve**") || !strings.HasSuffix(r, "gap `O1-O3`.\n\n## 22. **Last**. S.\nSource: [d](d), part `#3`.\n") {
		t.Errorf("after and last:\n%s", r)
	}
	before, _ := p.Render()
	if err := p.Place(e, 0); err == nil {
		t.Errorf("position 0 accepted")
	}
	if err := p.Place(e, len(p.Entries())+2); err == nil {
		t.Errorf("position len+2 accepted")
	}
	if after, _ := p.Render(); after != before {
		t.Errorf("a refused Place changed the document")
	}
	if _, err := p.After(99); err == nil {
		t.Errorf("After(99) resolved")
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#3 | R375
func TestRemoveDropsExactlyTheRun(t *testing.T) {
	src := pendingFixture(t)
	p := ParsePending(src)
	if err := p.Remove(12); err != nil {
		t.Fatal(err)
	}
	i := strings.Index(src, "## 12.")
	j := strings.Index(src, "## 3.")
	want := src[:i] + src[j:]
	if r, _ := p.Render(); r != want {
		t.Errorf("after Remove(12):\n%s\nwant:\n%s", r, want)
	}
	if got := ids(p); !slices.Equal(got, []int{8, 3, 14, 15}) {
		t.Errorf("ids %v, want [8 3 14 15]", got)
	}
	if err := p.Remove(99); err == nil {
		t.Errorf("Remove(99) accepted")
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.3 | R371, R375
//
// Found by injecting past the alarm list: the fixture's only rule precedes its
// entries, so nothing proved a `---` after an entry ends its region.
func TestARuleEndsARegion(t *testing.T) {
	src := "# Pending\n\n---\n\n## 4. **Only**. S.\nSource: [d](d), part `#1`.\n\n---\n\nTrailing prose after the rule.\n"
	p := ParsePending(src)
	e := p.Entry(4)
	if e == nil || strings.Contains(runText(e), "Trailing prose") {
		t.Fatalf("the region ran past the rule: %q", runText(e))
	}
	if err := p.Remove(4); err != nil {
		t.Fatal(err)
	}
	if r, _ := p.Render(); r != "# Pending\n\n---\n\n---\n\nTrailing prose after the rule.\n" {
		t.Errorf("after Remove:\n%q", r)
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.4.1 | R399, R400, R401
func TestASourceIsAPartOrAGap(t *testing.T) {
	p := ParsePending(pendingFixture(t))
	if e := p.Entry(8); e.Kind != SourcePart || e.SourceKey != "7" {
		t.Errorf("entry 8: kind %d key %q", e.Kind, e.SourceKey)
	}
	if e := p.Entry(14); e.Kind != SourceGap || e.SourceKey != "O136" || e.SourceDoc != "design/design.md" {
		t.Errorf("entry 14: %+v", *e)
	}
	// A range is not a gap source: the entry stands, its source does not, and it is unread.
	if e := p.Entry(15); e.Kind != SourceNone || e.SourceKey != "" || e.SourceDoc != "design/design.md" {
		t.Errorf("entry 15: %+v", *e)
	}
	// An entry with no Source line is SourceNone and not unread (entry 3).
	if e := p.Entry(3); e.Kind != SourceNone || e.SourceDoc != "" {
		t.Errorf("entry 3: %+v", *e)
	}
	// The writer emits one form for each kind, and refuses what would not read back.
	g := EntryText{ID: 30, Title: "Fix", Status: "S.", SourceDoc: "d.md", SourceKey: "O7", Kind: SourceGap}
	if got, want := g.Text(), "## 30. **Fix**. S.\nSource: [d.md](d.md), gap `O7`.\n\n"; got != want {
		t.Errorf("gap form:\n%q\nwant\n%q", got, want)
	}
	if err := p.Place(EntryText{ID: 31, Title: "Bad", Status: "S.", SourceDoc: "d.md", SourceKey: "O1, O2", Kind: SourceGap}, 1); !errors.Is(err, ErrBadGapSource) {
		t.Errorf("a list as a gap source: %v", err)
	}
	if err := p.Place(g, len(p.Entries())+1); err != nil {
		t.Fatal(err)
	}
	if e := p.Entry(30); e == nil || e.Kind != SourceGap || e.SourceKey != "O7" {
		t.Errorf("placed gap entry read back: %+v", e)
	}
}

// CRC: crc-PendingSdom.md | R403
func TestAGroupOpenAtEndOfInputIsUnreadInPending(t *testing.T) {
	src := "# Pending\n\n---\n\n## 1. **T**. s\nSource: [x](x.md), part `#1`.\n\n``oops`\n## 2. **U**. s\n"
	p := ParsePending(src)
	// R445: the two-run is demoted at end of input and the entry after it reads; the lone
	// backtick it had enclosed then opens a span of its own, demoted in turn.
	if len(p.Entries()) != 2 {
		t.Errorf("%d entries, want both — the span was text", len(p.Entries()))
	}
	u := p.Unread()
	if len(u) != 2 || u[0].Line != 8 || u[1].Line != 8 || !strings.HasSuffix(u[0].Text, "read as text") || !strings.HasSuffix(u[1].Text, "read as text") {
		t.Errorf("unread %+v, want the two-run and the lone backtick at line 8, read as text", u)
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.3.1 | R407
func TestPlaceAtTheLastPositionLandsBeforeTheRule(t *testing.T) {
	e := EntryText{ID: 20, Title: "New", Status: "Fresh.", SourceDoc: "carves/z.md", SourceKey: "1"}
	placed := "## 20. **New**. Fresh.\nSource: [carves/z.md](carves/z.md), part `#1`.\n"

	p := ParsePending(pendingFixture(t) + "\n---\n\nprose after the entries.\n")
	if err := p.Place(e, len(p.Entries())+1); err != nil {
		t.Fatal(err)
	}
	if got, _ := p.Render(); !strings.Contains(got, "gap `O1-O3`.\n\n"+placed+"\n---\n\nprose") {
		t.Errorf("entry not between the last entry and the rule:\n%s", got[len(got)-200:])
	}

	// No entries: after the header rule — before the commentary rule when one follows,
	// and ending the file in one newline when nothing does.
	for name, c := range map[string]struct{ src, want string }{
		"commentary follows": {"# Pending\n\nqueue.\n\n---\n\n---\n\nprose after the entries.\n",
			"# Pending\n\nqueue.\n\n---\n\n" + placed + "\n---\n\nprose after the entries.\n"},
		"nothing follows": {"# Pending\n\n---\n", "# Pending\n\n---\n\n" + placed},
	} {
		p := ParsePending(c.src)
		if err := p.Place(e, 1); err != nil {
			t.Fatal(name, err)
		}
		if got, _ := p.Render(); got != c.want {
			t.Errorf("empty queue, %s:\n%q", name, got)
		}
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#3.2.1 | R407, R408
func TestPlaceThenRemoveIsTheIdentity(t *testing.T) {
	e := EntryText{ID: 20, Title: "New", Status: "Fresh.", SourceDoc: "carves/z.md", SourceKey: "1"}
	fixture := pendingFixture(t)
	withRule := fixture + "\n---\n\nprose after the entries.\n"
	for name, c := range map[string]struct{ src, want string }{
		"plain":            {fixture, fixture},
		"rule+commentary":  {withRule, withRule},
		"no final newline": {strings.TrimRight(fixture, "\n"), fixture},
	} {
		p := ParsePending(c.src)
		if err := p.Place(e, len(p.Entries())+1); err != nil {
			t.Fatal(name, err)
		}
		if err := p.Remove(20); err != nil {
			t.Fatal(name, err)
		}
		if got, _ := p.Render(); got != c.want {
			t.Errorf("%s: not the identity; tail %q", name, got[len(got)-60:])
		}
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.2 | R412
func TestATitleWithEmphasisInsideReadsWhole(t *testing.T) {
	p := ParsePending("# P\n\n---\n\n## 5. **A **b** c** (skill). status here.\nSource: [x](x.md), part `#1`.\n")
	e := p.Entry(5)
	if e == nil {
		t.Fatal("entry 5 not read")
	}
	if e.Title != "A **b** c" || e.Skill != "skill" || e.Status != "status here" {
		t.Errorf("title %q skill %q status %q", e.Title, e.Skill, e.Status)
	}
}

// CRC: crc-PendingSdom.md | R413
func TestAWriteThatDoesNotReadBackPanics(t *testing.T) {
	mustReadBack("Pending", "Place", "7", true, "x", "x") // silent
	defer func() {
		r := recover()
		e, ok := r.(*ReadBackError)
		if !ok {
			t.Fatalf("panic %v, want a *ReadBackError", r)
		}
		if e.Reader != "Pending" || e.Write != "Place" || e.Key != "7" || !strings.Contains(e.Error(), "did not read back: want x, got y") {
			t.Errorf("error %q", e.Error())
		}
	}()
	mustReadBack("Pending", "Place", "7", false, "x", "y")
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#1.3 | R447
// A three-dash code span mid-sentence and a `---` line inside a fence are not rules:
// the region runs on to the next entry, and Remove leaves no tail behind.
func TestARuleIsALineOfItsOwnOutsideCode(t *testing.T) {
	head := "# Pending\n\n---\n\n"
	two := "## 6. **two**. s\nSource: [c.md](c.md), part `#2`.\n"
	for name, body := range map[string]string{
		"span":  "body from its `---` rule the way it does.\n\nmore tail.\n\n",
		"fence": "```\n---\n```\n\nmore tail.\n\n",
	} {
		src := head + "## 5. **one**. s\nSource: [c.md](c.md), part `#1`.\n\n" + body + two
		p := ParsePending(src)
		if n := len(p.Entries()); n != 2 {
			t.Fatalf("%s: %d entries, want 2", name, n)
		}
		if err := p.Remove(5); err != nil {
			t.Fatal(err)
		}
		if got, _ := p.Render(); got != head+two {
			t.Errorf("%s: the region ended at the dashes, leaving a tail:\n%s", name, got)
		}
	}
}

// CRC: crc-PendingSdom.md | Seq: seq-pending.md#2.2 | R501
func TestTheCanonicalEntryIsFlushLeftAndAnIndentedOneStillReads(t *testing.T) {
	p := ParsePending("# P\n\n---\n\n## 5. **one**. s\nSource: [c.md](c.md), part `#1`.\n")
	if err := p.Place(EntryText{ID: 9, Title: "nine", Status: "S", SourceDoc: "c.md", SourceKey: "2", Kind: SourcePart, Next: "go"}, 2); err != nil {
		t.Fatal(err)
	}
	r, _ := p.Render()
	if !strings.Contains(r, "\n## 9. **nine**. S\nSource: [c.md](c.md), part `#2`.\nNext: go\n") {
		t.Errorf("not flush left:\n%s", r)
	}
	// The earlier stencil's three-space indent is deliberate here: the readers accept
	// leading whitespace, so entries written under it still read.
	old := ParsePending("# P\n\n---\n\n## 5. **one**. s\n   Source: [c.md](c.md), part `#1`.\n   Next: write it.\n")
	if e := old.Entry(5); e == nil || e.SourceDoc != "c.md" || e.SourceKey != "1" || e.Next != "write it." {
		t.Errorf("indented entry misread: %+v", e)
	}
}
