// CRC: crc-Parser.md | Seq: seq-parse.md | R94, R95, R96, R98, R99, R100
package parser

import (
	"bufio"
	"cmp"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// SeqDoc holds the numbered items parsed from a sequence-diagram file. R95, R96
type SeqDoc struct {
	Path  string
	Items map[string]int // dotted id (no trailing dot) -> first line number
	Dupes []string       // ids that appeared on more than one line, sorted by id
	Ks    []int          // sorted unique first-segment values present
}

// Numbered reports whether the file contains any dotted-number items.
func (d *SeqDoc) Numbered() bool {
	return len(d.Items) > 0
}

// Has reports whether the given dotted id is present in the file.
func (d *SeqDoc) Has(id string) bool {
	_, ok := d.Items[id]
	return ok
}

// NumberingGaps returns dotted ids that must exist for contiguity but don't,
// covering both the K-sequence (top-level) and per-K tree levels. R98, R99
func (d *SeqDoc) NumberingGaps() []string {
	childrenByParent := make(map[string]map[int]bool)
	for id := range d.Items {
		parts := strings.Split(id, ".")
		for i, part := range parts {
			n, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			parent := strings.Join(parts[:i], ".")
			if childrenByParent[parent] == nil {
				childrenByParent[parent] = make(map[int]bool)
			}
			childrenByParent[parent][n] = true
		}
	}

	var gaps []string
	for parent, kids := range childrenByParent {
		maxN := 0
		for n := range kids {
			maxN = max(maxN, n)
		}
		for i := 1; i <= maxN; i++ {
			if kids[i] {
				continue
			}
			if parent == "" {
				gaps = append(gaps, strconv.Itoa(i))
			} else {
				gaps = append(gaps, parent+"."+strconv.Itoa(i))
			}
		}
	}
	slices.SortFunc(gaps, compareSeqIDs)
	return gaps
}

// SplitSeqRef parses a Seq reference like "seq-foo.md#1.4" into its file
// and fragment parts. Fragment is "" when no '#' is present. R94
func SplitSeqRef(ref string) (file, fragment string) {
	if i := strings.Index(ref, "#"); i >= 0 {
		return ref[:i], ref[i+1:]
	}
	return ref, ""
}

// seqItemRe matches a dotted-number token at the start of a line, allowing
// lane and tree characters before it. A single-segment number requires the
// trailing dot ("1.") to avoid matching prose; multi-segment numbers may
// omit the trailing dot. R95
var seqItemRe = regexp.MustCompile(`^[\s│├└─┌┐┘┤┬┴┼|+\-]*(\d+(?:\.\d+)+\.?|\d+\.)(?:\s|$)`)

// ParseSeqDoc scans a sequence-diagram file for numbered items. R95, R96, R100
func ParseSeqDoc(path string) (*SeqDoc, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc := &SeqDoc{
		Path:  path,
		Items: make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	dupeSeen := make(map[string]bool)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		m := seqItemRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		id := strings.TrimSuffix(m[1], ".")
		if _, ok := doc.Items[id]; ok {
			if !dupeSeen[id] {
				doc.Dupes = append(doc.Dupes, id)
				dupeSeen[id] = true
			}
			continue
		}
		doc.Items[id] = lineNum
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	slices.SortFunc(doc.Dupes, compareSeqIDs)

	kSet := make(map[int]bool)
	for id := range doc.Items {
		k, err := strconv.Atoi(strings.SplitN(id, ".", 2)[0])
		if err != nil {
			continue
		}
		kSet[k] = true
	}
	doc.Ks = make([]int, 0, len(kSet))
	for k := range kSet {
		doc.Ks = append(doc.Ks, k)
	}
	slices.Sort(doc.Ks)

	return doc, nil
}

// compareSeqIDs orders dotted ids numerically segment-by-segment.
func compareSeqIDs(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	for i := 0; i < min(len(aParts), len(bParts)); i++ {
		ai, _ := strconv.Atoi(aParts[i])
		bi, _ := strconv.Atoi(bParts[i])
		if c := cmp.Compare(ai, bi); c != 0 {
			return c
		}
	}
	return cmp.Compare(len(aParts), len(bParts))
}
