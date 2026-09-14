package minispecsdom

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func keysOf(ps []*Part) string {
	var ks []string
	for _, p := range ps {
		ks = append(ks, p.Key())
	}
	return strings.Join(ks, ",")
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#1 | R361, R362, R363, R364, R365, R369
func TestTheFixturesStatusBlockReadsAsParts(t *testing.T) {
	src := fixture(t)
	c := ParseCarve(src)
	if !c.HasStatus() {
		t.Fatal("no status block found")
	}
	if got := keysOf(c.Parts()); got != "1,2.1,2.2,4" {
		t.Fatalf("parts %q", got)
	}
	depths, parents := []int{}, []string{}
	for _, p := range c.Parts() {
		depths = append(depths, p.Depth)
		if p.Parent == nil {
			parents = append(parents, "-")
		} else {
			parents = append(parents, p.Parent.Key())
		}
	}
	if !slices.Equal(depths, []int{0, 2, 2, 0}) {
		t.Errorf("depths %v, want [0 2 2 0]", depths)
	}
	if got := strings.Join(parents, ","); got != "-,1,1,-" {
		t.Errorf("parents %q", got)
	}
	if len(c.Stateless()) != 1 || c.Stateless()[0].Key() != "2" {
		t.Errorf("stateless %d", len(c.Stateless()))
	}
	if r, _ := c.Render(); r != src {
		t.Errorf("render differs")
	}
	// A bullet outside the region is not a part; a level-3 heading does not end the
	// region, a level-2 one does.
	outside := ParseCarve("## Status\n\n- [ ] **Item 1 — a.** **OPEN (not queued.)**\n\n### Sub\n\n- [ ] **Item 2 — still inside.**\n\n## Notes\n\n- [ ] **Item 9 — not a part.**\n")
	if got := keysOf(outside.Parts()); got != "1,2" {
		t.Errorf("a bullet outside the region became a part: %q", got)
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#1.2 | R362, R363
func TestNoStatusBlockAndAFencedOne(t *testing.T) {
	for _, src := range []string{"# Carve\n\nprose\n\n## Decisions\n\n- [ ] **Item 1 — a.**\n", "# Carve\n\n```\n## Status\n\n- [ ] **Item 1 — a.**\n```\n"} {
		c := ParseCarve(src)
		if c.HasStatus() || len(c.Parts()) != 0 {
			t.Errorf("%q: status=%v parts=%d", src, c.HasStatus(), len(c.Parts()))
		}
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#2.3 | R366, R367
func TestSetMarkerFollowsTheToolsRule(t *testing.T) {
	c := ParseCarve(fixture(t))
	if err := c.SetMarker("4", "OPEN", "#9."); err != nil {
		t.Fatal(err)
	}
	r, _ := c.Render()
	if !strings.Contains(r, "- [ ] **Item 4 — fail fast.** **OPEN (#9.)**\n") {
		t.Errorf("Item 4 after SetMarker:\n%s", r)
	}
	src := "## Status\n\n- [ ] **Item 1 — a.** **NOT VERIFIED.** **OPEN (#8.)**\n- [ ] **Item 2 — b.**\n- [ ] **Item 3 — c.** **OPEN (#4.)** **OPEN (#5.)**\n- [ ] **Item 4 — d.** **REVERTED (#70.)** Needs Item 1.\n- [ ] **Item 5 — e.** Needs Item 1.\n"
	c = ParseCarve(src)
	if err := c.SetMarker("1", "LANDED", "x"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetMarker("2", "DEFERRED", "Bill, 2026-09-03"); err != nil {
		t.Fatal(err)
	}
	// Two transients on one line: the first is replaced, the second removed.
	if err := c.SetMarker("3", "OPEN", "#6."); err != nil {
		t.Fatal(err)
	}
	// R409: REVERTED is a transient, replaced by the replay's OPEN rather than joined by it;
	// and a marker appended to a line with trailing prose goes before the prose.
	if err := c.SetMarker("4", "OPEN", "#70."); err != nil {
		t.Fatal(err)
	}
	if err := c.SetMarker("5", "OPEN", "#3."); err != nil {
		t.Fatal(err)
	}
	if err := c.SetMarker("7", "OPEN", "#1."); err == nil {
		t.Errorf("an unknown key was accepted")
	}
	want := "## Status\n\n- [ ] **Item 1 — a.** **NOT VERIFIED.** **LANDED (x)**\n- [ ] **Item 2 — b.** **DEFERRED (Bill, 2026-09-03)**\n- [ ] **Item 3 — c.** **OPEN (#6.)**\n- [ ] **Item 4 — d.** **OPEN (#70.)** Needs Item 1.\n- [ ] **Item 5 — e.** **OPEN (#3.)** Needs Item 1.\n"
	if r, _ := c.Render(); r != want {
		t.Errorf("got:\n%s\nwant:\n%s", r, want)
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#2 | R368
func TestLandIsThreeMarkingsInOneAct(t *testing.T) {
	c := ParseCarve(fixture(t))
	if err := c.Land("2.2", "`abc`, 2026-09-03 — `#8`."); err != nil {
		t.Fatal(err)
	}
	r, _ := c.Render()
	if !strings.Contains(r, "  - [x] ~~**2.2 — the config move.**~~ **LANDED (`abc`, 2026-09-03 — `#8`.)**\n") {
		t.Errorf("after Land:\n%s", r)
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#2.1.1 | R391, R392
func TestWritesRefuseOverDeviations(t *testing.T) {
	src := "## Status\n\n- [ ] **Item 1 — a.** **OPEN (soon)**\n"
	c := ParseCarve(src)
	for _, tc := range []struct {
		name  string
		write func() error
	}{
		{"SetMarker", func() error { return c.SetMarker("1", "LANDED", "x") }},
		{"Land", func() error { return c.Land("1", "x") }},
	} {
		err := tc.write()
		var dev *DeviationError
		if !errors.As(err, &dev) {
			t.Errorf("%s over a deviating line: %v", tc.name, err)
			continue
		}
		if dev.Key != "1" || len(dev.Deviations) != 1 || dev.Deviations[0].Rule != "OPEN attribution" {
			t.Errorf("%s: %+v", tc.name, dev)
		}
		if !strings.Contains(err.Error(), "OPEN attribution: "+markerTarget) {
			t.Errorf("%s: the error does not name the rule and target:\n%s", tc.name, err)
		}
		if r, _ := c.Render(); r != src {
			t.Errorf("%s changed a refused line:\n%s", tc.name, r)
		}
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#2.3.1 | R391, R393
func TestOpenNeverReopensALandedPart(t *testing.T) {
	src := fixture(t)
	c := ParseCarve(src)
	if err := c.SetMarker("1", "open", "not queued."); !errors.Is(err, ErrReopen) {
		t.Errorf("OPEN over a landed part: %v", err)
	}
	if r, _ := c.Render(); r != src {
		t.Errorf("a refused reopen changed the line:\n%s", r)
	}
	// The guard is narrow: a record other than OPEN still goes on a landed part.
	if err := c.SetMarker("1", "NOT VERIFIED", "Bill, 2026-09-04"); err != nil {
		t.Errorf("a record over a landed part: %v", err)
	}
}

// CRC: crc-CarveSdom.md | Seq: seq-carve.md#2.1.2 | R391, R394
func TestLandRefusesOverALandedPart(t *testing.T) {
	src := fixture(t)
	c := ParseCarve(src)
	if err := c.Land("1", "`fff`, 2026-09-04 — `#9`."); !errors.Is(err, ErrLanded) {
		t.Errorf("Land over a landed part: %v", err)
	}
	if r, _ := c.Render(); r != src {
		t.Errorf("a refused Land changed the line:\n%s", r)
	}
}

// CRC: crc-CarveSdom.md | R395
func TestEveryPartKnowsItsLine(t *testing.T) {
	c := ParseCarve(fixture(t))
	want := map[string]int{"1": 7, "2.1": 9, "2.2": 10, "4": 11}
	for _, p := range c.Parts() {
		if p.Line() != want[p.Key()] {
			t.Errorf("%s: line %d, want %d", p.Key(), p.Line(), want[p.Key()])
		}
	}
}

// CRC: crc-CarveSdom.md | R404
func TestAGroupOpenAtEndOfInputIsUnreadInCarve(t *testing.T) {
	src := "# C\n\n## Status\n\n- [ ] **Item 1 — t.** **OPEN (not queued.)**\n\n`x\n"
	c := ParseCarve(src)
	if len(c.Parts()) != 1 {
		t.Errorf("%d parts, want 1", len(c.Parts()))
	}
	u := c.Unread()
	if len(u) != 1 || u[0].Line != 7 || !strings.Contains(u[0].Text, "never closed") {
		t.Errorf("unread %+v, want the span at line 7", u)
	}
}

// CRC: crc-CarveSdom.md | R411
func TestACloserThatClosesNothingIsUnread(t *testing.T) {
	c := ParseCarve("# C\n\n## Status\n\n- [ ] **Item 1 — t.** **OPEN (not queued.)**\n\nsee ``x```y``\n")
	u := c.Unread()
	// Three reports on one line: the span the three-run ended (never closed), the three-run
	// itself (closes nothing), and the span the trailing two-run opened, demoted at end of
	// input (never closed, read as text) — R445.
	if len(u) != 3 || !strings.HasSuffix(u[0].Text, "never closed") || !strings.Contains(u[1].Text, "```` closes nothing") || !strings.HasSuffix(u[2].Text, "read as text") {
		t.Errorf("unread %+v, want the ended span, the rejected three-run and the demoted span", u)
	}
}

// CRC: crc-CarveSdom.md | R417
func TestTheKeyIsTheFragment(t *testing.T) {
	c := ParseCarve(fixture(t))
	if c.Part("4") == nil || c.Part("Item 4") != nil {
		t.Errorf("Part resolves by the fragment alone: %v %v", c.Part("4") != nil, c.Part("Item 4") != nil)
	}
	if k := c.Stateless()[0].Key(); k != "2" {
		t.Errorf("stateless key %q, want 2", k)
	}
}
