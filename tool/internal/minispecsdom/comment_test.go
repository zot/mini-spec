package minispecsdom

import (
	"slices"
	"strings"
	"testing"

	"github.com/zot/simple-dom/sdom"
)

func parse(src string, lang *sdom.BracketLang) (*sdom.Doc, *sdom.BracketContext) {
	bp := sdom.NewBracketParser(lang)
	return sdom.Parse(src, 0, bp), bp.Context()
}

func allLangs() map[string]*sdom.BracketLang {
	return map[string]*sdom.BracketLang{
		"go": &sdom.LangGo, "shell": &sdom.LangShell, "pascal": &sdom.LangPascal,
		"js": &sdom.LangJavaScript, "lua": &sdom.LangLua, "python": &sdom.LangPython.BracketLang,
	}
}

// goc wraps an interior as a Go line comment. Fixtures are built from the style
// rather than written as literals, so no test string looks like a traceability
// comment to a line-based harvester reading this file.
func goc(interior string) string { return sdom.LangGo.Comment.Prefix + interior + "\n" }

// one parses src, runs the pass, and requires exactly one comment.
func one(t *testing.T, src string, lang *sdom.BracketLang) (*sdom.Doc, *TraceabilityComment) {
	t.Helper()
	d, ctx := parse(src, lang)
	cs, err := Comments(d, ctx)
	if err != nil || len(cs) != 1 {
		t.Fatalf("%q: %d comments, err %v", src, len(cs), err)
	}
	return d, cs[0]
}

// seeds is the grammar's corpus: every field order, no spaces, wide spaces, all
// three separators, ranges, steps. Each is an INTERIOR; the fuzz wraps it per
// language.
var seeds = []string{
	"CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5",
	"CRC:crc-Index.md|Seq:seq-x.md#2|R7",
	"R5 | CRC: crc-x.md | Seq: s.md#1.2 -- note",
	"CRC:   crc-Odd.md   |   R7",
	"R5, R6",
	"R5: desc",
	"R7",
	"Test: test-BibleRender.md | CRC: crc-A.md, crc-B.md | R3181, R3182-3190",
	"CRC: a.md — with an em dash",
	"R4-6, R9 -- CRC: x | R4 -- desc",
	"Seq: seq-a.md#1, seq-b.md#2.3 | Test: t.md",
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#2 | R337, R338, R339, R341, R342, R345
//
// The DOM round-trip: render(dom) == src is what a byte round-trip proves, and it
// goes green even if nothing was parsed into fields. Comparing DOMs — the reparse
// against the reference — is what fails on under-modelling.
func FuzzDOMRoundTrip(f *testing.F) {
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, interior string) {
		if strings.ContainsAny(interior, "\n") {
			return
		}
		for name, lang := range allLangs() {
			cs := lang.Comment
			src := cs.Prefix + interior + cs.Suffix
			d, ctx := parse(src, lang)
			got, _ := Comments(d, ctx)
			r, err := d.Render()
			if err != nil || r != src {
				t.Fatalf("%s: render %q != %q (%v)", name, r, src, err)
			}
			if len(got) != 1 {
				continue // not recognized: nothing more to check for this input
			}
			// The DOM compare below runs the same parser on both sides, so it cannot
			// see SYMMETRIC under-modelling; this can. Recognition means a field.
			if c := got[0]; c.CRC() == nil && c.Seq() == nil && c.Test() == nil && c.Refs() == nil {
				t.Fatalf("%s: %q recognized with no field", name, src)
			}
			d2, ctx2 := parse(r, lang)
			again, _ := Comments(d2, ctx2)
			if len(again) != 1 || !again[0].Equals(got[0]) {
				t.Fatalf("%s: %q does not reparse to an Equals DOM", name, src)
			}
		}
	})
}

// CRC: crc-TraceabilityComment.md | R338, R340, R348
func TestFieldsReadBack(t *testing.T) {
	_, c := one(t, goc("CRC: crc-Store.md, crc-Index.md | Seq: seq-crud.md#1.4 | Test: t.md | R4, R5-7 -- note"), &sdom.LangGo)
	if got := c.CRC().Items(); !slices.Equal(got, []string{"crc-Store.md", "crc-Index.md"}) {
		t.Errorf("CRC %q", got)
	}
	if got := c.Seq().Items(); !slices.Equal(got, []string{"seq-crud.md#1.4"}) {
		t.Errorf("Seq %q", got)
	}
	if p, s := SeqStep(c.Seq().Items()[0]); p != "seq-crud.md" || s != "1.4" {
		t.Errorf("SeqStep %q %q", p, s)
	}
	if got := c.Test().Items(); !slices.Equal(got, []string{"t.md"}) {
		t.Errorf("Test %q", got)
	}
	if got := c.Refs().Items(); !slices.Equal(got, []int{4, 5, 6, 7}) {
		t.Errorf("Refs %v", got)
	}
	if d, _ := c.Description().Render(); d != " note" {
		t.Errorf("Description %q", d)
	}
	_, c = one(t, goc("R7"), &sdom.LangGo)
	if c.CRC() != nil || c.Description() != nil || !slices.Equal(c.Refs().Items(), []int{7}) {
		t.Errorf("a refs-only comment: CRC=%v desc=%v", c.CRC(), c.Description())
	}
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#1 | R342, R343, R346
//
// Leading with a keyword is not enough: recognition is consumption.
func TestRecognitionIsConsumption(t *testing.T) {
	src := goc("CRC: crc-A.md | R1") +
		goc("Test: a repaint frame round-trips as kind-only (R3136).") +
		goc("see R5") +
		goc("(R5)") +
		"x := 1 " + goc("R7")
	d, ctx := parse(src, &sdom.LangGo)
	before := len(d.Nodes())
	var prose *sdom.Opener
	for _, n := range d.Nodes() {
		if o, ok := n.(*sdom.Opener); ok && strings.HasPrefix(ctx.InnerText(o), " Test: a repaint") {
			prose = o
		}
	}
	if c := new(TraceabilityComment); c.Parse(prose, ctx) {
		t.Errorf("the prose Test: comment was recognized")
	}
	if len(d.Nodes()) != before {
		t.Errorf("a failed Parse changed the document")
	}
	cs, err := Comments(d, ctx)
	if err != nil || len(cs) != 2 {
		t.Fatalf("%d comments recognized, want 2 (%v)", len(cs), err)
	}
	if got := cs[1].Refs().Items(); !slices.Equal(got, []int{7}) {
		t.Errorf("the trailing comment reads %v", got)
	}
	if r, _ := d.Render(); r != src {
		t.Errorf("render changed:\n%q", r)
	}
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#1.2.2 | R344, R346
func TestSpliceReusesTheMarkers(t *testing.T) {
	src := "func f() {}\n" + goc("CRC: crc-A.md | R1") + "func g() {}\n"
	d, ctx := parse(src, &sdom.LangGo)
	var opener *sdom.Opener
	for _, n := range d.Nodes() {
		if o, ok := n.(*sdom.Opener); ok {
			if s, _ := o.Render(); s == "//" {
				opener = o
			}
		}
	}
	closer := ctx.Closer(opener)
	before := len(d.Nodes())
	cs, _ := Comments(d, ctx)
	if len(cs) != 1 {
		t.Fatal("expected one comment")
	}
	kids := cs[0].Kids()
	if kids[0] != sdom.Node(opener) || kids[len(kids)-1] != sdom.Node(closer) {
		t.Errorf("the markers were recreated rather than reused")
	}
	if len(d.Nodes()) != before-2 {
		t.Errorf("%d nodes, want %d: three became one", len(d.Nodes()), before-2)
	}
	braces := 0
	for _, n := range d.Nodes() {
		if o, ok := n.(*sdom.Opener); ok {
			if s, _ := o.Render(); s == "{" && ctx.Closer(o) != nil {
				braces++
			}
		}
	}
	if braces != 2 {
		t.Errorf("the context pairs %d brace groups after the splice, want 2", braces)
	}
}

// CRC: crc-TraceabilityComment.md | R347, R348
func TestNewParsesBackEqualsInEveryLanguage(t *testing.T) {
	f := Fields{CRC: []string{"crc-A.md"}, Seq: []string{"seq-b.md#1.2"}, Refs: []int{4, 5, 6, 9}, Description: "note"}
	for name, lang := range allLangs() {
		cs := lang.Comment
		if cs.Prefix == "" {
			continue
		}
		c := New(lang, f)
		r, _ := c.Render()
		want := cs.Prefix + "CRC: crc-A.md | Seq: seq-b.md#1.2 | R4-6, R9 -- note" + cs.Suffix
		if r != want {
			t.Errorf("%s: New renders %q, want %q", name, r, want)
		}
		if c.Location().Origin() != nil {
			t.Errorf("%s: a constructed node carries an origin", name)
		}
		_, back := one(t, r, lang)
		if !back.Equals(c) {
			t.Errorf("%s: the parse-back is not Equals to the constructed node", name)
		}
	}
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#3 | R341
func TestAFieldWriteThroughTheNode(t *testing.T) {
	d, c := one(t, goc("R5: desc"), &sdom.LangGo)
	c.Refs().SetItems([]int{5, 6, 7})
	if r, _ := d.Render(); r != goc("R5-7: desc") {
		t.Errorf("after the write: %q", r)
	}
	d, c = one(t, goc("CRC: a.md — note"), &sdom.LangGo)
	if s, _ := c.Description().Render(); s != " note" {
		t.Errorf("description %q", s)
	}
	if r, _ := d.Render(); r != goc("CRC: a.md — note") {
		t.Errorf("the em dash was not preserved: %q", r)
	}
	if err := c.CRC().SetItems([]string{"a b"}); err == nil {
		t.Errorf("an item with whitespace was accepted")
	}
}

// CRC: crc-TraceabilityComment.md | Seq: seq-anchor.md#1.1 | R346
//
// Found by injecting past the alarm list: dropping the kind filter in Comments
// left every test green. A bracket group or a string whose interior looks like
// fields is not a comment, and the pass must not so much as try it.
func TestOnlyCommentGroupsAreCandidates(t *testing.T) {
	src := "x := f(R7)\ny := \"CRC: crc-A.md | R1\"\nz := [R4, R5]\n"
	d, ctx := parse(src, &sdom.LangGo)
	before := len(d.Nodes())
	cs, err := Comments(d, ctx)
	if err != nil || len(cs) != 0 {
		t.Fatalf("%d comments recognized in a file with none (%v)", len(cs), err)
	}
	if len(d.Nodes()) != before {
		t.Errorf("the pass changed a document it recognized nothing in")
	}
}
