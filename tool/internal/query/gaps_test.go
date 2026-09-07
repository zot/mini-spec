// CRC: crc-Query.md | R317, R318, R319, R320, R321, R322
package query

import (
	"strings"
	"testing"

	"github.com/zot/minispec/internal/parser"
)

var pool = []parser.Gap{
	{ID: "O22", Type: "O", HasCheckbox: true},
	{ID: "O23", Type: "O", HasCheckbox: true, Resolved: true},
	{ID: "O26", Type: "O", HasCheckbox: true},
	{ID: "A1", Type: "A"},
	{ID: "T2", Type: "T"},
	{ID: "R5", Type: "R", HasCheckbox: true},
}

func ids(gs []parser.Gap) string {
	out := make([]string, len(gs))
	for i, g := range gs {
		out[i] = g.ID
	}
	return strings.Join(out, " ")
}

// R317, R318 — the grammar, and the letter as namespace.
func TestExpandGapRefsGrammar(t *testing.T) {
	got, err := ExpandGapRefs([]string{"O22-O24,A1", "R5-7"})
	if err != nil || strings.Join(got, " ") != "O22 O23 O24 A1 R5 R6 R7" {
		t.Errorf("got %v (%v)", got, err)
	}
	if got, _ := ExpandGapRefs([]string{"O28-O22"}); strings.Join(got, " ") != "O28" {
		t.Errorf("a reversed range contributes only its low end; got %v", got)
	}
	if _, err := ExpandGapRefs([]string{"O22-R5"}); err == nil || !strings.Contains(err.Error(), "crosses") {
		t.Errorf("a range across types was not refused: %v", err)
	}
	if _, err := ExpandGapRefs([]string{"gap22"}); err == nil {
		t.Error("a non-ID was accepted")
	}
}

// R319, R320 — nothing matched is an error naming the ask; partly unassigned is not.
func TestSelectGapsNamesAnEmptyAskAndToleratesGaps(t *testing.T) {
	if _, err := SelectGaps(pool, GapSelection{IDs: []string{"O99"}}); err == nil || !strings.Contains(err.Error(), "O99") {
		t.Errorf("an unmatched selection did not refuse naming what it looked for: %v", err)
	}
	got, err := SelectGaps(pool, GapSelection{IDs: []string{"O22", "O23", "O24", "O25", "O26"}})
	if err != nil || ids(got) != "O22 O23 O26" {
		t.Errorf("partly unassigned range: got %q (%v), want O22 O23 O26 in document order", ids(got), err)
	}
}

// R321, R322 — permanent gaps are claimed by neither state flag; both flags mean every checkbox.
func TestSelectGapsStateFlagsNeverClaimPermanent(t *testing.T) {
	open, _ := SelectGaps(pool, GapSelection{Open: true})
	if ids(open) != "O22 O26 R5" {
		t.Errorf("--open: got %q, want O22 O26 R5 — a permanent gap read as work to do", ids(open))
	}
	closed, _ := SelectGaps(pool, GapSelection{Closed: true})
	if ids(closed) != "O23" {
		t.Errorf("--closed: got %q, want O23", ids(closed))
	}
	both, _ := SelectGaps(pool, GapSelection{Open: true, Closed: true})
	if ids(both) != "O22 O23 O26 R5" {
		t.Errorf("both flags: got %q, want every gap with a checkbox", ids(both))
	}
	all, _ := SelectGaps(pool, GapSelection{})
	if len(all) != len(pool) {
		t.Errorf("the zero selection selects everything; got %d of %d", len(all), len(pool))
	}
	composed, err := SelectGaps(pool, GapSelection{IDs: []string{"O22", "O23"}, Closed: true})
	if err != nil || ids(composed) != "O23" {
		t.Errorf("range and flag compose: got %q (%v)", ids(composed), err)
	}
}
