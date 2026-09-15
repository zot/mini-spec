// CRC: crc-LinkRepair.md | Seq: seq-links.md#3 | R463, R464, R465, R466, R467
package update

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/query"
)

// Outcome is what the repair did with one missing link.
type Outcome string

const (
	Rewritten    Outcome = "rewritten"
	Unresolvable Outcome = "unresolvable" // no relocation resolves: not the move's doing
	Ambiguous    Outcome = "ambiguous"    // more than one does: not the tool's to choose
)

// Considered is one missing link and its outcome.
type Considered struct {
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Old     string  `json:"old"`
	New     string  `json:"new,omitempty"`
	Outcome Outcome `json:"outcome"`
}

// RepairReport is every link considered and the counts, zeros included.
type RepairReport struct {
	Files      []string        `json:"files"`
	Considered []Considered    `json:"considered"`
	Counts     map[Outcome]int `json:"counts"`
	Written    []string        `json:"written"` // the files that changed
}

// Unrepaired reports whether any considered link was left. R468's exit status.
func (r *RepairReport) Unrepaired() bool {
	return r.Counts[Unresolvable]+r.Counts[Ambiguous] > 0
}

// CRC: crc-LinkRepair.md | Seq: seq-links.md#3.6 | R467
// Summary is the closing count.
func (r *RepairReport) Summary() string {
	return fmt.Sprintf("%d links considered in %d files: rewritten %d, unresolvable %d, ambiguous %d; %d files written",
		len(r.Considered), len(r.Files), r.Counts[Rewritten], r.Counts[Unresolvable], r.Counts[Ambiguous], len(r.Written))
}

// CRC: crc-LinkRepair.md | Seq: seq-links.md#3.1 | R463
// repairPopulation is the live carves and every `*.md` directly under each carve directory's
// `done/`: both ends of a move can hold a broken link.
func repairPopulation(root string) ([]string, error) {
	files, err := query.LiveCarves(root)
	if err != nil {
		return nil, err
	}
	for _, dir := range []string{"carves", ".carves"} {
		done, err := filepath.Glob(filepath.Join(root, dir, "done", "*.md"))
		if err != nil {
			return nil, err
		}
		for _, p := range done {
			rel, _ := filepath.Rel(root, p)
			files = append(files, filepath.ToSlash(rel))
		}
	}
	return files, nil
}

// CRC: crc-LinkRepair.md | Seq: seq-links.md#3 | R463, R464, R466
//
// RepairLinks rewrites, in every file of the population, each missing link that exactly one
// sibling relocation resolves. Only missing links are considered: the verb repairs one
// defect, a move, and retargets nothing for any other reason. A file is written only when
// at least one of its links was rewritten, atomically, through the reader's own write.
func RepairLinks(root string, files []string) (*RepairReport, error) {
	if len(files) == 0 {
		population, err := repairPopulation(root)
		if err != nil {
			return nil, err
		}
		files = population
	}
	report := &RepairReport{Counts: map[Outcome]int{Rewritten: 0, Unresolvable: 0, Ambiguous: 0}}
	for _, f := range files {
		rel, abs := relAbs(root, f)
		report.Files = append(report.Files, rel)
		var considered []Considered
		wrote := false
		err := parser.EditFile(abs, func(src string) (string, error) {
			m := minispecsdom.ParseMarkdown(src)
			for i, l := range m.Links() {
				if cls, _ := query.ClassifyLink(root, rel, l); cls != query.ClassMissing {
					continue
				}
				c := Considered{File: rel, Line: l.Line(), Old: l.Dest}
				switch cands := candidates(root, rel, l); len(cands) {
				case 0:
					c.Outcome = Unresolvable
				case 1:
					c.Outcome = Rewritten
					c.New = newDest(rel, cands[0], l)
					if err := m.SetDest(i, c.New); err != nil {
						return "", err
					}
					wrote = true
				default:
					c.Outcome = Ambiguous
				}
				considered = append(considered, c)
			}
			if !wrote {
				return src, nil
			}
			return m.Render()
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", rel, err)
		}
		for _, c := range considered {
			report.Counts[c.Outcome]++
		}
		report.Considered = append(report.Considered, considered...)
		if wrote {
			report.Written = append(report.Written, rel)
		}
	}
	return report, nil
}

// relAbs is a path both ways round root.
func relAbs(root, f string) (rel, abs string) {
	if filepath.IsAbs(f) {
		abs = filepath.Clean(f)
	} else {
		abs = filepath.Join(root, f)
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		rel = abs
	}
	return filepath.ToSlash(rel), abs
}

// CRC: crc-LinkRepair.md | Seq: seq-links.md#3.3 | R464, R465
//
// candidates are the sibling relocations a move between a carve directory and its `done/`
// can produce, kept when something is on disk there inside the root: the citing directory
// re-based across `done/`, and the target's name moved across `done/`. Deduplicated, so a
// relocation reachable both ways counts once.
func candidates(root, file string, l minispecsdom.Link) []string {
	dir := path.Dir(file)
	target := path.Join(dir, l.Path)
	var out []string
	seen := map[string]bool{}
	try := func(p string) {
		p = path.Clean(p)
		if seen[p] || p == ".." || strings.HasPrefix(p, "../") {
			return
		}
		seen[p] = true
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(p))); err == nil {
			out = append(out, p)
		}
	}
	try(path.Join(acrossDone(dir), l.Path))
	try(path.Join(acrossDone(path.Dir(target)), path.Base(target)))
	return out
}

// acrossDone is dir with `done/` removed from its end, or added to it.
func acrossDone(dir string) string {
	if path.Base(dir) == "done" {
		return path.Dir(dir)
	}
	return path.Join(dir, "done")
}

// CRC: crc-LinkRepair.md | Seq: seq-links.md#3.4 | R466
// newDest is the relocated target relative to the citing file's directory, slash-separated,
// the fragment kept as written and the angle wrapping kept if the old destination had it.
func newDest(file, target string, l minispecsdom.Link) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(file)), filepath.FromSlash(target))
	if err != nil {
		rel = target
	}
	d := filepath.ToSlash(rel)
	if l.Fragment != "" {
		d += "#" + l.Fragment
	}
	oldDest := strings.TrimSpace(l.Raw[strings.IndexByte(l.Raw, '(')+1:])
	if strings.HasPrefix(oldDest, "<") {
		d = "<" + d + ">"
	}
	return d
}
