package minispecsdom

import (
	"os"
	"strings"
	"testing"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

func parseMarkdown(t *testing.T, src string) (*sdom.Doc, *sdom.BracketContext) {
	t.Helper()
	p := schema.NewMarkdownParser()
	d := sdom.Parse(src, 0, p)
	return d, p.Indent().Brackets().Context()
}

func fixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/trajectory-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func lines(t *testing.T, src string) (*sdom.Doc, []*PartLine) {
	t.Helper()
	d, ctx := parseMarkdown(t, src)
	ps, err := PartLines(d, ctx)
	if err != nil {
		t.Fatalf("PartLines: %v", err)
	}
	return d, ps
}

func byKey(ps []*PartLine) map[string]*PartLine {
	m := map[string]*PartLine{}
	for _, p := range ps {
		m[p.Key()] = p
	}
	return m
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#1 | R349, R351, R352, R353, R356, R360
func TestTheStatusBlocksLinesReadBack(t *testing.T) {
	src := fixture(t)
	d, ps := lines(t, src)
	if r, _ := d.Render(); r != src {
		t.Fatalf("render differs from source")
	}
	var keys []string
	for _, p := range ps {
		keys = append(keys, p.Key())
	}
	if got := strings.Join(keys, ","); got != "1,2,2.1,2.2,4" {
		t.Fatalf("keys %q", got)
	}
	m := byKey(ps)
	for k, want := range map[string]string{"1": "x", "2": "-", "2.1": "x", "2.2": " ", "4": " "} {
		cb := m[k].Checkbox()
		switch {
		case want == "-" && cb != nil:
			t.Errorf("%s: has a checkbox", k)
		case want != "-" && cb == nil:
			t.Errorf("%s: no checkbox", k)
		case want != "-" && cb.Checked() != (want == "x"):
			t.Errorf("%s: checked=%v", k, cb.Checked())
		}
	}
	one := m["1"]
	if len(one.Markers()) != 1 {
		t.Fatalf("Item 1: %d markers", len(one.Markers()))
	}
	mk := one.Markers()[0]
	if v, _ := mk.Verb().Render(); v != "LANDED" {
		t.Errorf("verb %q", v)
	}
	if id, ok := mk.QueueID(); !ok || id != 3 {
		t.Errorf("queue id %d %v", id, ok)
	}
	if a := mk.Attribution(); !strings.Contains(a, "4c6e974") {
		t.Errorf("attribution %q", a)
	}
	if title, _ := one.Title().Render(); title != "record and resolve." {
		t.Errorf("title %q", title)
	}
	two := m["2"]
	if two.IsStruck() || len(two.Markers()) != 1 {
		t.Errorf("Item 2: struck=%v markers=%d", two.IsStruck(), len(two.Markers()))
	} else if v, _ := two.Markers()[0].Verb().Render(); v != "SPLIT" {
		t.Errorf("Item 2 verb %q", v)
	}
	if id, ok := m["2.2"].Markers()[0].QueueID(); !ok || id != 8 {
		t.Errorf("2.2 queue id %d %v", id, ok)
	}
	if !one.IsStruck() || !m["2.1"].IsStruck() || m["2.2"].IsStruck() {
		t.Errorf("strike derivation wrong")
	}
	if len(one.Deviations()) != 0 {
		t.Errorf("Item 1 deviations %v", one.Deviations())
	}
}

// CRC: crc-PartLine.md | Seq: seq-partline.md#2 | R354
func TestStrikeIsDerivedAndHidden(t *testing.T) {
	src := fixture(t)
	d, ps := lines(t, src)
	m := byKey(ps)
	m["2.2"].Strike(true)
	m["1"].Strike(false)
	if !m["2.2"].IsStruck() || m["1"].IsStruck() {
		t.Fatalf("IsStruck did not follow Strike")
	}
	r, _ := d.Render()
	if !strings.Contains(r, "- [ ] ~~**2.2 — the config move.**~~ **OPEN (#8.)**") {
		t.Errorf("2.2 not struck around the head:\n%s", r)
	}
	if !strings.Contains(r, "- [x] **Item 1 — record and resolve.** **LANDED") {
		t.Errorf("Item 1 not unstruck:\n%s", r)
	}
	// Every other byte is where it was: undo both and compare.
	m["2.2"].Strike(false)
	m["1"].Strike(true)
	if r, _ := d.Render(); r != src {
		t.Errorf("strike/unstrike moved other bytes")
	}
}

// CRC: crc-PartLine.md | R350, R355
func TestDeviationsNameTheTarget(t *testing.T) {
	src := "- [ ] **Part A — old scheme.** **OPEN (#8.)**\n- [X] **Item 3 - hyphen.**\n- [ ] **Item 5 — ok.** **open (soon.)**\n- a plain bullet\n- [ ] **Item 6 — r.** **REVERTED (Bill)**\n"
	d, ps := lines(t, src)
	if len(ps) != 5 {
		t.Fatalf("%d lines, want 5", len(ps))
	}
	rules := func(p *PartLine) string {
		var r []string
		for _, dv := range p.Deviations() {
			if dv.Target == "" {
				t.Errorf("deviation %q has no target", dv.Rule)
			}
			r = append(r, dv.Rule)
		}
		return strings.Join(r, ",")
	}
	if got := rules(ps[0]); got != "key form" {
		t.Errorf("line 1 deviations %q", got)
	}
	if got := rules(ps[1]); got != "checkbox interior,separator,key form" {
		t.Errorf("line 2 deviations %q", got)
	}
	if got := rules(ps[2]); got != "verb case,OPEN attribution" {
		t.Errorf("line 3 deviations %q", got)
	}
	if got := rules(ps[4]); got != "REVERTED attribution" { // R410
		t.Errorf("line 5 deviations %q", got)
	}
	if ps[0].Checkbox() == nil || ps[0].Checkbox().Checked() {
		t.Errorf("an unkeyed line's checkbox must still count")
	}
	// A headless line: nothing to strike, and asking must not fail.
	plain := ps[3]
	if got := rules(plain); got != "key form" || plain.IsStruck() {
		t.Errorf("plain bullet: deviations %q struck %v", got, plain.IsStruck())
	}
	plain.Strike(true)
	if r, _ := d.Render(); r != src || plain.IsStruck() {
		t.Errorf("Strike on a headless line changed something")
	}
}

// CRC: crc-MarkerSpan.md | Seq: seq-partline.md#3 | R357
func TestAMarkerWriteIsCanonicalAndGuarded(t *testing.T) {
	d, ps := lines(t, fixture(t))
	mk := byKey(ps)["2.2"].Markers()[0]
	if err := mk.Set("landed", "`abc`, 2026-09-03 — `#8`."); err != nil {
		t.Fatal(err)
	}
	if r, _ := d.Render(); !strings.Contains(r, "**LANDED (`abc`, 2026-09-03 — `#8`.)**") {
		t.Errorf("render after Set:\n%s", r)
	}
	if id, ok := mk.QueueID(); !ok || id != 8 {
		t.Errorf("queue id %d %v", id, ok)
	}
	before, _ := d.Render()
	if err := mk.Set("BAD)", "x"); err == nil {
		t.Errorf("a verb with a parenthesis was accepted")
	}
	if after, _ := d.Render(); after != before {
		t.Errorf("a refused write changed the document")
	}
}

// CRC: crc-PartLine.md | R397, R398
func TestFlexibleOnInputRigidOnOutput(t *testing.T) {
	// Loose spellings of the two OPEN attributions read clean.
	for _, a := range []string{"#3.", "#3", "not queued.", "not queued", "Not  queued.", "NOT QUEUED"} {
		_, ps := lines(t, "- [ ] **Item 1 — a.** **OPEN ("+a+")**\n")
		if d := ps[0].Deviations(); len(d) != 0 {
			t.Errorf("OPEN (%s): %v", a, d)
		}
	}
	// A superseded scheme stays a deviation, and so does anything off the shape.
	for src, rule := range map[string]string{
		"- [ ] **Item 1 — a.** **OPEN, not queued.**\n": "marker scheme",
		"- [ ] **Item 1 — a.** **OPEN (soon)**\n":       "OPEN attribution",
	} {
		_, ps := lines(t, src)
		d := ps[0].Deviations()
		if len(d) != 1 || d[0].Rule != rule || d[0].Target == "" {
			t.Errorf("%q: %v", src, d)
		}
	}
	// The write side is still one form: Set writes the canonical stop-less attribution as given.
	_, ps := lines(t, "- [ ] **Item 1 — a.** **open (Not queued)**\n")
	if err := ps[0].Markers()[0].Set("OPEN", "#4."); err != nil {
		t.Fatal(err)
	}
	if r, _ := ps[0].Render(); r != "- [ ] **Item 1 — a.** **OPEN (#4.)**" {
		t.Errorf("after Set: %q", r)
	}
}
