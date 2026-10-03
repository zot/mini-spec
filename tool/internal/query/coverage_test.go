// CRC: crc-Query.md | Test: test-Query.md | R538, R539
package query

import (
	"strings"
	"testing"
)

// renderCoverage is one line per entry: `R6 crc-A.md crc-B.md`, `R99 ?` when unknown.
func renderCoverage(entries []CoverageEntry) string {
	var lines []string
	for _, e := range entries {
		line := e.ID
		if e.Unknown {
			line += " ?"
		}
		for _, f := range e.Files {
			line += " " + f
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "; ")
}

// R538, R539 — named IDs only, ascending, files by name; unknown marked; none means all.
func TestCoverageSelectsAscendingByFileName(t *testing.T) {
	cov := map[string][]string{
		"R10": {},
		"R6":  {"/d/crc-A.md", "/d/crc-B.md"},
		"R5":  {"/d/crc-A.md"},
	}
	ids, err := CoverageIDs([]string{"R10,R5-6", "R99"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := renderCoverage(SelectCoverage(cov, ids)), "R5 crc-A.md; R6 crc-A.md crc-B.md; R10; R99 ?"; got != want {
		t.Errorf("selected: got %q, want %q", got, want)
	}
	// SelectCoverage orders by itself, not by trusting its caller: unsorted in, ascending out.
	if got, want := renderCoverage(SelectCoverage(cov, []string{"R99", "R10", "R5"})), "R5 crc-A.md; R10; R99 ?"; got != want {
		t.Errorf("unsorted selection: got %q, want %q — ascending whatever order it was named in", got, want)
	}
	if got, want := renderCoverage(SelectCoverage(cov, nil)), "R5 crc-A.md; R6 crc-A.md crc-B.md; R10"; got != want {
		t.Errorf("all: got %q, want %q — every requirement, ascending", got, want)
	}
}

// R538 — a coverage lookup names requirements; a pattern is refused.
func TestCoverageRefusesWhatIsNotARef(t *testing.T) {
	for _, arg := range []string{"@status", "D3"} {
		if ids, err := CoverageIDs([]string{arg}); err == nil {
			t.Errorf("%q: want a refusal, got %v", arg, ids)
		}
	}
}
