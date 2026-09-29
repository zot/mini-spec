// CRC: crc-Query.md | Seq: seq-query.md | R502, R503, R504, R505, R506, R507
package query

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/parser"
)

// ImplSelection is what `query implementation` was asked: requirement IDs (number mode) or
// a pattern over requirement text (text mode), and whether text mode keeps retired ones.
type ImplSelection struct {
	IDs     []string       // number mode: ascending, each once
	Pattern *regexp.Regexp // text mode
	Retired bool
}

// NumberMode reports whether the caller named the requirements.
func (sel ImplSelection) NumberMode() bool { return sel.Pattern == nil }

// ImplLocation is one comment implementing a requirement.
type ImplLocation struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Comment string `json:"comment"`
}

// ImplEntry is one selected requirement and where it is implemented. Text is empty in
// number mode, where the caller already knows the requirement.
type ImplEntry struct {
	ID        string         `json:"id"`
	Text      string         `json:"text,omitempty"`
	Retired   bool           `json:"retired,omitempty"`
	Locations []ImplLocation `json:"locations"`
}

// ImplResult is the whole answer, the files the harvest could not read included, since a
// requirement implemented only inside one would otherwise read as unimplemented.
type ImplResult struct {
	Mode    string              `json:"mode"`
	Entries []ImplEntry         `json:"entries"`
	Unread  []parser.UnreadFile `json:"unread,omitempty"`
}

// CRC: crc-Query.md | Seq: seq-query.md | R503
// ClassifyImplArgs reads the args as requirement refs in the `query gaps` grammar, R-only,
// and falls back to a single regexp over requirement text. A gap ID that is not an R is a
// pattern rather than a lookup: `D3` names no requirement, and answering it as one would
// print nothing where the caller meant a search.
func ClassifyImplArgs(args []string, retired bool) (ImplSelection, error) {
	sel := ImplSelection{Retired: retired}
	if len(args) == 0 {
		return sel, fmt.Errorf("usage: minispec query implementation <Rn... | pattern> [--retired]")
	}
	if ids, err := ExpandGapRefs(args); err == nil && allRequirements(ids) {
		sel.IDs = sortedReqIDs(ids)
		return sel, nil
	}
	if len(args) > 1 {
		return sel, fmt.Errorf("%q is not a list of requirement refs, and a text pattern is one argument", strings.Join(args, " "))
	}
	re, err := regexp.Compile(args[0])
	if err != nil {
		return sel, fmt.Errorf("%q is neither requirement refs nor a valid pattern: %w", args[0], err)
	}
	sel.Pattern = re
	return sel, nil
}

// allRequirements reports whether ids names at least one ID and every one is an R.
func allRequirements(ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		if !strings.HasPrefix(id, "R") {
			return false
		}
	}
	return true
}

// sortedReqIDs orders IDs by number and drops repeats.
func sortedReqIDs(ids []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Slice(out, func(i, j int) bool { return reqNum(out[i]) < reqNum(out[j]) })
	return out
}

func reqNum(id string) int {
	n, _ := strconv.Atoi(strings.TrimPrefix(id, "R"))
	return n
}

// CRC: crc-Query.md | Seq: seq-query.md | R502, R504, R505, R506
// SelectImplementation answers the selection from the harvest: one entry per selected
// requirement, ascending, its locations in manifest then line order. An entry with no
// locations is kept, empty — that is the "no impl refs" the output must state, and dropping
// it would read as every match being implemented.
func SelectImplementation(reqs []parser.Requirement, h parser.Harvest, sel ImplSelection) ImplResult {
	index := map[string][]ImplLocation{}
	for _, fh := range h.Files {
		for _, c := range fh.Comments {
			for _, r := range c.Refs {
				index[r] = append(index[r], ImplLocation{Path: fh.Path, Line: c.Line, Comment: c.Text})
			}
		}
	}
	res := ImplResult{Mode: "text", Unread: h.Unread}
	if sel.NumberMode() {
		res.Mode = "number"
		retired := map[string]bool{}
		for _, r := range reqs {
			retired[r.ID] = r.Retired
		}
		for _, id := range sel.IDs {
			res.Entries = append(res.Entries, ImplEntry{ID: id, Retired: retired[id], Locations: index[id]})
		}
		return res
	}
	for _, r := range reqs {
		if (sel.Retired || !r.Retired) && sel.Pattern.MatchString(r.Text) {
			res.Entries = append(res.Entries, ImplEntry{ID: r.ID, Text: r.Text, Retired: r.Retired, Locations: index[r.ID]})
		}
	}
	sort.SliceStable(res.Entries, func(i, j int) bool { return reqNum(res.Entries[i].ID) < reqNum(res.Entries[j].ID) })
	return res
}

// CRC: crc-Query.md | Seq: seq-query.md | R502, R506
// Implementation reads the requirements and the one harvest validate reads, and answers the
// selection from them.
func (q *Query) Implementation(sel ImplSelection) (ImplResult, error) {
	reqs, err := q.Requirements()
	if err != nil {
		return ImplResult{}, err
	}
	h, err := q.TraceabilityAll()
	if err != nil {
		return ImplResult{}, err
	}
	return SelectImplementation(reqs, h, sel), nil
}
