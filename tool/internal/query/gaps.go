// CRC: crc-Query.md | Seq: seq-query.md | R317, R318, R319, R320, R321, R322
package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/parser"
)

// GapSelection is what the caller asked for: which IDs, and which checkbox states. The zero
// value selects everything, which is what `query gaps` has always printed.
type GapSelection struct {
	IDs    []string // expanded IDs; empty means every gap
	Open   bool
	Closed bool
}

var gapRefRe = regexp.MustCompile(`^([SRDCIOAT])(\d+)(?:-([SRDCIOAT])?(\d+))?$`)

// CRC: crc-Query.md | R317, R318
// ExpandGapRefs reads RANGE arguments in the grammar inline requirement refs already use —
// a bare ID, an inclusive range with the second letter optional, comma lists, mixtures —
// into the IDs they name. A range whose ends name different types is refused: the letter is
// the namespace, so `O22-R5` spans nothing. A reversed range contributes only its low end.
func ExpandGapRefs(args []string) ([]string, error) {
	var out []string
	for _, arg := range args {
		for _, ref := range strings.Split(arg, ",") {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				continue
			}
			m := gapRefRe.FindStringSubmatch(ref)
			if m == nil {
				return nil, fmt.Errorf("%q is not a gap ID or range (O22, O22-O28, O22-28, O22,O25)", ref)
			}
			typ, lo := m[1], m[2]
			if m[4] == "" {
				out = append(out, typ+lo)
				continue
			}
			if m[3] != "" && m[3] != typ {
				return nil, fmt.Errorf("%s crosses gap types %s and %s; the letter is the namespace, so a range spans one type", ref, typ, m[3])
			}
			a, _ := strconv.Atoi(lo)
			b, _ := strconv.Atoi(m[4])
			if b < a {
				out = append(out, typ+lo)
				continue
			}
			for n := a; n <= b; n++ {
				out = append(out, typ+strconv.Itoa(n))
			}
		}
	}
	return out, nil
}

// CRC: crc-Query.md | R319, R320, R321, R322
// SelectGaps narrows gaps to the selection, in document order rather than argument order so
// two runs of one question can be diffed.
//
// An ID selection matching nothing is an error naming what it looked for: printing nothing
// and succeeding would read as *there are no gaps*, and a report of absence is
// indistinguishable from a report of nothing wrong. A selection whose members are merely
// partly unassigned is not an error — gaps in an ID sequence are expected by the ID rule.
// A permanent gap (A, T) carries no checkbox by rule, so neither state filter claims it;
// both flags together mean every gap that has a checkbox.
func SelectGaps(gaps []parser.Gap, sel GapSelection) ([]parser.Gap, error) {
	out := gaps
	if len(sel.IDs) > 0 {
		want := make(map[string]bool, len(sel.IDs))
		for _, id := range sel.IDs {
			want[id] = true
		}
		var kept []parser.Gap
		for _, g := range gaps {
			if want[g.ID] {
				kept = append(kept, g)
			}
		}
		if len(kept) == 0 {
			return nil, fmt.Errorf("no such gap: looked for %s", strings.Join(sel.IDs, ", "))
		}
		out = kept
	}
	if !sel.Open && !sel.Closed {
		return out, nil
	}
	var kept []parser.Gap
	for _, g := range out {
		// HasCheckbox first is load-bearing: Resolved is false by zero value, so `!Resolved`
		// alone would sweep every permanent entry into --open, reading exactly like work.
		if !g.HasCheckbox {
			continue
		}
		if (sel.Open && !g.Resolved) || (sel.Closed && g.Resolved) {
			kept = append(kept, g)
		}
	}
	return kept, nil
}

// Describe names a selection for a report that matched nothing: `open gaps in O22-O28`.
func (sel GapSelection) Describe(args []string) string {
	state := "gaps"
	switch {
	case sel.Open && sel.Closed:
		state = "gaps with a checkbox"
	case sel.Open:
		state = "open gaps"
	case sel.Closed:
		state = "closed gaps"
	}
	if len(args) == 0 {
		return state
	}
	return state + " in " + strings.Join(args, " ")
}
