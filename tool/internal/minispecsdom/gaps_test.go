package minispecsdom

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func loadGaps(t *testing.T) (*Gaps, string) {
	t.Helper()
	src, err := os.ReadFile("testdata/gaps-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return ParseGaps(string(src)), string(src)
}

func gapIDs(g *Gaps) string {
	var ids []string
	for _, p := range g.Items() {
		ids = append(ids, p.ID)
	}
	return strings.Join(ids, ",")
}

// CRC: crc-Gaps.md | R429, R430, R431, R432
func TestGapsReadsTheFixture(t *testing.T) {
	g, src := loadGaps(t)
	if !g.HasGaps() {
		t.Fatal("no section")
	}
	if ids := gapIDs(g); ids != "A1,T1,I1,O1,O2,O3,O4" {
		t.Errorf("ids = %s", ids)
	}
	a1, t1, i1, o1, o2, o3 := g.Gap("A1"), g.Gap("T1"), g.Gap("I1"), g.Gap("O1"), g.Gap("O2"), g.Gap("O3")
	if !a1.Permanent() || a1.Checkbox || a1.Line() != 13 || a1.Type != "A" || a1.Number != 1 {
		t.Errorf("A1: %+v", a1)
	}
	if len(t1.Sub) != 1 || t1.Sub[0] != "reason: the root moved from the design directory to the repository" {
		t.Errorf("T1 sub = %q", t1.Sub)
	}
	if !i1.Checked || !i1.Checkbox || !strings.HasSuffix(i1.Text, "with a parse context.") {
		t.Errorf("I1: %+v", i1)
	}
	if o1.Checked || !strings.Contains(o1.Text, "its children is; `Compound.Location` also") {
		t.Errorf("O1 did not fold: %q", o1.Text)
	}
	if len(o2.Sub) != 2 || o2.Sub[1] != "[ ] Feature B (3 scenarios)" {
		t.Errorf("O2 sub = %q", o2.Sub)
	}
	if o3.Text != "A document about the format quotes an entry in a fence:" || g.Gap("O99") != nil {
		t.Errorf("O3: %q / O99 %v", o3.Text, g.Gap("O99"))
	}
	if g.Gap("O5") != nil {
		t.Error("a bullet after the section read as a gap")
	}
	if len(g.Unread()) != 0 {
		t.Errorf("unread = %v", g.Unread())
	}
	if out, _ := g.Render(); out != src {
		t.Error("render is not byte-exact")
	}
	if n := ParseGaps("# X\n\nno section\n"); n.HasGaps() || len(n.Items()) != 0 {
		t.Error("a document with no section has gaps")
	}
	if f := ParseGaps("# X\n\n```\n## Gaps\n- [ ] O1: fenced\n```\n"); f.HasGaps() {
		t.Error("a fenced heading opened a section")
	}
}

// CRC: crc-Gaps.md | R430, R432
func TestGapsDeviationsAndNesting(t *testing.T) {
	g := ParseGaps("## Gaps\n\n- [ ] A1: boxed permanent\n- O2: unboxed tracked\n- [ ] O3: fine\n  - [ ] O4: nested\n- [ ] O3: again\n- reason: a bare bullet\n")
	if len(g.Gap("A1").Deviations()) != 1 || len(g.Gap("O2").Deviations()) != 1 {
		t.Errorf("checkbox deviations: %v %v", g.Gap("A1").Deviations(), g.Gap("O2").Deviations())
	}
	o4 := g.Gap("O4")
	if o4 == nil {
		t.Fatal("the nested O4 was not read as a gap")
	}
	if o4.Depth != 2 || o4.Parent != g.Gap("O3") || len(o4.Deviations()) != 0 {
		t.Errorf("nested: %+v", o4)
	}
	if g.Gap("O3").Text != "fine" || len(g.Items()) != 5 || len(g.Items()[4].Deviations()) != 1 {
		t.Errorf("doubled ID: %+v", g.Items()[4])
	}
	if u := g.Unread(); len(u) != 4 || !strings.HasSuffix(u[3].Text, "a bare bullet") {
		t.Errorf("unread = %v", u)
	}
	var de *DeviationError
	if err := g.Resolve("A1"); !errors.As(err, &de) {
		t.Errorf("write over a deviant entry: %v", err)
	}
	if err := g.Resolve("O9"); !errors.Is(err, ErrNoGap) {
		t.Errorf("absent: %v", err)
	}
}

// CRC: crc-Gaps.md | R436, R550
func TestGapsAdd(t *testing.T) {
	g, src := loadGaps(t)
	if err := g.Add("O5", "a new one with `code`"); err != nil {
		t.Fatal(err)
	}
	out, _ := g.Render()
	if !strings.Contains(out, "- [ ] O4: last entry, one line\n- [ ] O5: a new one with `code`\n\n## Notes") {
		t.Errorf("placement:\n%s", out)
	}
	if len(out) != len(src)+len("- [ ] O5: a new one with `code`\n") {
		t.Error("changed more than the added line")
	}
	if err := g.Add("A2", "approved forever"); err != nil {
		t.Fatal(err)
	}
	out, _ = g.Render()
	if !strings.Contains(out, "`code`\n- A2: approved forever\n\n## Notes") || g.Gap("A2").Checkbox {
		t.Errorf("permanent add:\n%s", out)
	}
	if err := g.Add("O5", "x"); !errors.Is(err, ErrGapExists) {
		t.Errorf("exists: %v", err)
	}
	if err := g.Add("Q1", "x"); !errors.Is(err, ErrBadGapID) {
		t.Errorf("bad id: %v", err)
	}
	if err := ParseGaps("# X\n").Add("O1", "x"); !errors.Is(err, ErrNoSection) {
		t.Errorf("no section: %v", err)
	}
	e := ParseGaps("# X\n\n## Gaps\n\n## Notes\n")
	if err := e.Add("O1", "first"); err != nil {
		t.Fatal(err)
	}
	if out, _ = e.Render(); out != "# X\n\n## Gaps\n- [ ] O1: first\n\n## Notes\n" {
		t.Errorf("empty section: %q", out)
	}
}

// CRC: crc-Gaps.md | R434, R435, R436
func TestGapsResolveAndApprove(t *testing.T) {
	g, src := loadGaps(t)
	if err := g.Resolve("O1"); err != nil {
		t.Fatal(err)
	}
	out, _ := g.Render()
	if !strings.Contains(out, "- [x] O1: R28 understates") || len(out) != len(src) || !g.Gap("O1").Checked {
		t.Error("resolve did not flip the head alone")
	}
	if err := g.Resolve("O1"); !errors.Is(err, ErrResolved) {
		t.Errorf("twice: %v", err)
	}
	if err := g.Resolve("A1"); !errors.Is(err, ErrPermanent) {
		t.Errorf("permanent: %v", err)
	}
	if err := g.Approve("O3", "A2"); err != nil {
		t.Fatal(err)
	}
	out, _ = g.Render()
	if !strings.Contains(out, "- A2: A document about the format quotes an entry in a fence:\n\n  ```markdown") {
		t.Errorf("approve:\n%s", out)
	}
	if g.Gap("O3") != nil || g.Gap("A2").Permanent() != true || g.Gap("A2").Text != "A document about the format quotes an entry in a fence:" {
		t.Errorf("after approve: %+v", g.Gap("A2"))
	}
	// A wrapped entry keeps every line beneath its head.
	g, src = loadGaps(t)
	if err := g.Approve("O1", "A3"); err != nil {
		t.Fatal(err)
	}
	out, _ = g.Render()
	if !strings.Contains(out, "- A3: R28 understates the rule the code implements. It says a compound is altered if any of\n  its children is;") || len(out) != len(src)-len("[ ] O1")+len("A3") {
		t.Errorf("approve of a wrapped entry:\n%s", out)
	}
	if err := g.Approve("O2", "O7"); !errors.Is(err, ErrBadGapID) {
		t.Errorf("non-A: %v", err)
	}
	if err := g.Approve("O2", "A1"); !errors.Is(err, ErrGapExists) {
		t.Errorf("taken: %v", err)
	}
	if err := g.Approve("T1", "A4"); !errors.Is(err, ErrPermanent) {
		t.Errorf("permanent: %v", err)
	}
}

// CRC: crc-Gaps.md | Test: test-Gaps.md | R540
func TestGapsAddAfterUnterminatedLine(t *testing.T) {
	src := "# X\n\n## Gaps\n\n- [ ] O1: last, no newline"
	g := ParseGaps(src)
	if err := g.Add("O2", "next"); err != nil {
		t.Fatal(err)
	}
	if out, _ := g.Render(); out != src+"\n- [ ] O2: next\n" {
		t.Errorf("render: %q", out)
	}
	if got := g.Gap("O1"); got == nil || got.Text != "last, no newline" {
		t.Errorf("O1 read back as %+v", got)
	}
}

// subsectionGaps is ui-engine's layout in miniature: T entries at the head, prose under one
// subsection, a D under its own, and an A filed under Oversights as frictionless does.
const subsectionGaps = "# Design\n\n## Gaps\n\n- T1: R1 retired by R2 (x)\n\n" +
	"### Incomplete Implementation\n\nNone identified.\n\n" +
	"### Design → Code Gaps\n\n- [ ] D1: a design gap\n\n" +
	"### Oversights (On)\n\n- [ ] O1: an oversight\n- A1: approved here\n\n## Notes\n"

// CRC: crc-Gaps.md | Test: test-Gaps.md | R550, R551
func TestGapsAddPlacesBySubsection(t *testing.T) {
	g := ParseGaps(subsectionGaps)
	for _, add := range [][2]string{{"T2", "t"}, {"O2", "o"}, {"A2", "a"}, {"I1", "i"}, {"S1", "s"}} {
		if err := g.Add(add[0], add[1]); err != nil {
			t.Fatalf("%s: %v", add[0], err)
		}
	}
	out, _ := g.Render()
	for _, want := range []string{
		"- T1: R1 retired by R2 (x)\n- T2: t\n- [ ] S1: s\n\n### Incomplete Implementation",
		"None identified.\n- [ ] I1: i\n\n### Design",
		"- [ ] O1: an oversight\n- [ ] O2: o\n- A1: approved here\n- A2: a\n\n## Notes",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

// CRC: crc-Gaps.md | Test: test-Gaps.md | R550
func TestASubsectionHeadingNamesAType(t *testing.T) {
	for _, c := range []struct {
		title, letter string
		want          bool
	}{
		{"Oversights (On)", "O", true},
		{"design -> code gaps", "D", true},
		{"Spec→Requirements", "S", true},
		{"Spec → Design Gaps", "S", false},
		{"Spec → Design Gaps", "D", false},
		{"Notes", "O", false},
		{"Approved (An)", "O", false},
	} {
		if got := namesType(c.title, c.letter); got != c.want {
			t.Errorf("namesType(%q, %s) = %v, want %v", c.title, c.letter, got, c.want)
		}
	}
}

// CRC: crc-Gaps.md | Test: test-Gaps.md | R551
func TestAHeadingEndsTheEntryAboveIt(t *testing.T) {
	g := ParseGaps("# D\n\n## Gaps\n\n- [ ] O1: first\n### Later\n- [ ] O2: second\n")
	if got := g.Gap("O1"); got == nil || got.Text != "first" {
		t.Errorf("O1 read as %+v", got)
	}
	if g.Gap("O2") == nil {
		t.Error("O2 under the subsection was not read")
	}
}
