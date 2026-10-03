// CRC: crc-Query.md | R538, R539
package query

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
)

// CoverageEntry is one requirement `query coverage` prints: the design files that reference it,
// by name, or Unknown when no requirement carries the ID.
type CoverageEntry struct {
	ID      string   `json:"id"`
	Files   []string `json:"files"`
	Unknown bool     `json:"unknown,omitempty"`
}

// CRC: crc-Query.md | R538
// CoverageIDs reads `query coverage` arguments: requirement refs in the number-mode grammar of
// `query implementation`, or none for every requirement. A coverage lookup names requirements,
// so an argument that would be a text pattern there is an error here.
func CoverageIDs(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	sel, err := ClassifyImplArgs(args, false)
	if err != nil || !sel.NumberMode() {
		return nil, fmt.Errorf("%q is not a list of requirement refs (R5, R5-8, R5,R9)", args)
	}
	return sel.IDs, nil
}

// CRC: crc-Query.md | R538, R539
// SelectCoverage is what `query coverage` prints: every requirement in cov, or only ids, in
// ascending order either way, with design files by base name. A selected ID cov does not hold
// is no requirement at all, and says so rather than reading as uncovered.
func SelectCoverage(cov map[string][]string, ids []string) []CoverageEntry {
	if ids == nil {
		ids = slices.Collect(maps.Keys(cov))
	}
	ids = sortedReqIDs(ids)
	out := make([]CoverageEntry, 0, len(ids))
	for _, id := range ids {
		paths, known := cov[id]
		files := make([]string, 0, len(paths))
		for _, p := range paths {
			files = append(files, filepath.Base(p))
		}
		out = append(out, CoverageEntry{ID: id, Files: slices.Compact(files), Unknown: !known})
	}
	return out
}
