// CRC: crc-Query.md | Test: test-Query.md | R502, R503, R504, R505, R506
package query

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/parser"
)

// R503 — clean requirement refs are number mode, ascending and each once; anything else is
// one pattern, and a gap ID that is not an R is a pattern too.
func TestClassifyImplArgs(t *testing.T) {
	sel, err := ClassifyImplArgs([]string{"R8,R5-6", "R5"}, false)
	if err != nil || !sel.NumberMode() || !slices.Equal(sel.IDs, []string{"R5", "R6", "R8"}) {
		t.Errorf("refs: %+v, %v", sel, err)
	}
	for _, arg := range []string{"@status", "D3"} {
		sel, err := ClassifyImplArgs([]string{arg}, false)
		if err != nil || sel.NumberMode() || sel.Pattern.String() != arg {
			t.Errorf("%s: want text mode, got %+v, %v", arg, sel, err)
		}
	}
	for _, args := range [][]string{{"foo", "bar"}, {"("}, nil} {
		if _, err := ClassifyImplArgs(args, false); err == nil {
			t.Errorf("%q: want a refusal", args)
		}
	}
}

// implFixture: a.go line 3 writes a range, b.go line 9 a header; R7 is retired and R8 has
// no implementing comment. Document order is not numeric order, as in a real requirements.md.
func implFixture() ([]parser.Requirement, parser.Harvest) {
	reqs := []parser.Requirement{
		{ID: "R8", Text: "eight"}, {ID: "R6", Text: "six"},
		{ID: "R7", Text: "seven", Retired: true}, {ID: "R5", Text: "five"},
	}
	h := parser.Harvest{Files: []parser.FileHarvest{
		{Path: "a.go", Comments: []parser.HarvestComment{{Line: 3, Text: "// R5-R7", Refs: []string{"R5", "R6", "R7"}}}},
		{Path: "b.go", Comments: []parser.HarvestComment{{Line: 9, Text: "// CRC: x.md | R6", CRC: []string{"x.md"}, Refs: []string{"R6"}}}},
	}}
	return reqs, h
}

// render is one line per entry: `R6 a.go:3 b.go:9`, text and retired marked when present.
func render(res ImplResult) string {
	var lines []string
	for _, e := range res.Entries {
		line := e.ID
		if e.Text != "" {
			line += "[" + e.Text + "]"
		}
		if e.Retired {
			line += "(retired)"
		}
		for _, l := range e.Locations {
			line += fmt.Sprintf(" %s:%d", l.Path, l.Line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "; ")
}

func mustClassify(t *testing.T, retired bool, args ...string) ImplSelection {
	t.Helper()
	sel, err := ClassifyImplArgs(args, retired)
	if err != nil {
		t.Fatal(err)
	}
	return sel
}

// R502, R504, R506 — a range answers each member; locations in manifest then line order;
// requirements ascending; an unimplemented one kept, empty; text only in text mode.
func TestSelectImplementationAnswersEachRequirement(t *testing.T) {
	reqs, h := implFixture()
	got := render(SelectImplementation(reqs, h, mustClassify(t, false, "R8,R6")))
	if want := "R6 a.go:3 b.go:9; R8"; got != want {
		t.Errorf("number mode: got %q, want %q", got, want)
	}
	res := SelectImplementation(reqs, h, mustClassify(t, false, "."))
	if got, want := render(res), "R5[five] a.go:3; R6[six] a.go:3 b.go:9; R8[eight]"; got != want {
		t.Errorf("text mode: got %q, want %q — ascending, R8 kept with no refs, retired R7 left out", got, want)
	}
	if e := res.Entries[0]; e.ID != "R5" || len(e.Locations) == 0 || e.Locations[0].Comment != "// R5-R7" {
		t.Errorf("comment not carried: %+v", e)
	}
}

// R505 — retired answers by number always, and by text only with --retired.
func TestSelectImplementationRetired(t *testing.T) {
	reqs, h := implFixture()
	for _, flag := range []bool{false, true} {
		if got := render(SelectImplementation(reqs, h, mustClassify(t, flag, "R7"))); got != "R7(retired) a.go:3" {
			t.Errorf("number mode, --retired=%v: got %q", flag, got)
		}
	}
	if got := render(SelectImplementation(reqs, h, mustClassify(t, false, "seven|five"))); got != "R5[five] a.go:3" {
		t.Errorf("text mode: got %q — a retired requirement surveyed as live intent", got)
	}
	if got := render(SelectImplementation(reqs, h, mustClassify(t, true, "seven|five"))); got != "R5[five] a.go:3; R7[seven](retired) a.go:3" {
		t.Errorf("text mode --retired: got %q", got)
	}
}
