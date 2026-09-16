// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4 | R469, R470, R471, R490, R473, R474
package update

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/query"
)

// Left is a link in the carve that does not resolve today: reported, not the move's to fix.
const Left Outcome = "left"

// FinishReport is the move, every rewrite, every link left, and the counts.
type FinishReport struct {
	From       string          `json:"from"`
	To         string          `json:"to"`
	Considered []Considered    `json:"considered"`
	Counts     map[Outcome]int `json:"counts"`
	Written    []string        `json:"written"`
}

// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4.6 | R475
// Summary is the closing count.
func (r *FinishReport) Summary() string {
	return fmt.Sprintf("moved %s → %s: rewritten %d, left %d; %d files written",
		r.From, r.To, r.Counts[Rewritten], r.Counts[Left], len(r.Written))
}

var (
	ErrNotACarve   = errors.New("not directly in a carve directory (carves/ or .carves/ at the repository root)")
	ErrDestination = errors.New("a file is already at the destination")
	ErrNoStatus    = errors.New("the carve has no status block, so it cannot be known finished")
)

// OpenPartsError is the refusal on a carve that is not finished, naming what is open.
type OpenPartsError struct{ Parts []string }

func (e *OpenPartsError) Error() string {
	return fmt.Sprintf("the carve still has %d open part(s): %s", len(e.Parts), strings.Join(e.Parts, ", "))
}

// plan is one document's rewrites, applied in memory and written only after every plan held.
type plan struct {
	rel, abs string
	doc      *minispecsdom.Markdown
	render   string
	changed  bool
}

// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4 | R469, R470, R473, R474
//
// FinishCarve moves carve to its directory's done/ and rewrites every link the move would
// break, both directions, from where each link resolves now. Every refusal comes before
// any byte moves: the path, the destination, the status block, and every rewrite's
// read-back are all settled in memory first; only then are the incoming documents written
// in place, the carve written at its destination, and the old file removed.
func FinishCarve(root, carve string) (rep *FinishReport, err error) {
	rel, abs := relAbs(root, carve)
	dir := path.Dir(rel)
	if dir != "carves" && dir != ".carves" {
		return nil, fmt.Errorf("%s: %w", rel, ErrNotACarve)
	}
	dest := path.Join(dir, "done", path.Base(rel))
	destAbs := filepath.Join(root, filepath.FromSlash(dest))
	if _, statErr := os.Stat(destAbs); statErr == nil {
		return nil, fmt.Errorf("%s: %w", dest, ErrDestination)
	}
	cv, err := parser.ReadCarve(abs, rel)
	if err != nil {
		return nil, err
	}
	if !cv.HasStatus {
		return nil, fmt.Errorf("%s: %w", rel, ErrNoStatus)
	}
	var open []string
	for _, p := range cv.Parts {
		if p.State() == parser.PartOpen {
			open = append(open, p.Key())
		}
	}
	if len(open) > 0 {
		return nil, &OpenPartsError{Parts: open}
	}
	// A rewrite the reader cannot read back panics inside SetDest; here that is a refusal
	// with nothing written, because no file has been touched yet. R473
	defer func() {
		if r := recover(); r != nil {
			if rb, ok := r.(*minispecsdom.ReadBackError); ok {
				rep, err = nil, fmt.Errorf("%w — nothing was written", rb)
				return
			}
			panic(r)
		}
	}()
	report := &FinishReport{From: rel, To: dest, Counts: map[Outcome]int{Rewritten: 0, Left: 0}}
	self, err := outgoingPlan(root, rel, dest, report)
	if err != nil {
		return nil, err
	}
	population, err := finishPopulation(root)
	if err != nil {
		return nil, err
	}
	var incoming []*plan
	for _, f := range population {
		if f == rel {
			continue
		}
		p, err := incomingPlan(root, f, rel, dest, report)
		if err != nil {
			return nil, err
		}
		if p.changed {
			incoming = append(incoming, p)
		}
	}
	// Every plan held. Write the incoming documents in place, the carve at its destination,
	// and remove the old file — a plain rename, nothing staged. R474
	for _, p := range incoming {
		if err := parser.EditFile(p.abs, func(string) (string, error) { return p.render, nil }); err != nil {
			return nil, err
		}
		report.Written = append(report.Written, p.rel)
	}
	if err := os.MkdirAll(filepath.Dir(destAbs), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(destAbs, []byte(self.render), 0o644); err != nil {
		return nil, err
	}
	if err := os.Remove(abs); err != nil {
		return nil, err
	}
	report.Written = append(report.Written, dest)
	return report, nil
}

// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4.4 | R490
// finishPopulation is the owned documents — every tracked markdown file plus every sited
// one — since a move breaks the reference in every document the tool wrote. R497
func finishPopulation(root string) ([]string, error) { return query.OwnedDocuments(root) }

// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4.3 | R471
//
// outgoingPlan rewrites each of the carve's links that resolves now so it reaches the same
// target from dest; a link that does not resolve is recorded as left. Computed from the
// current target, never searched for after the move: a search could land on a different
// file that happens to sit at the relocated path.
func outgoingPlan(root, rel, dest string, report *FinishReport) (*plan, error) {
	p, err := readPlan(root, rel)
	if err != nil {
		return nil, err
	}
	for i, l := range p.doc.Links() {
		cls, target := query.ClassifyLink(root, rel, l)
		c := Considered{File: rel, Line: l.Line(), Old: l.Dest}
		switch cls {
		case "":
			c.Outcome = Rewritten
			c.New = newDest(dest, target, l)
			if err := p.doc.SetDest(i, c.New); err != nil {
				return nil, err
			}
			p.changed = true
		case query.ClassMissing, query.ClassOutside:
			c.Outcome = Left
		default:
			continue // external and local links do not break
		}
		report.Counts[c.Outcome]++
		report.Considered = append(report.Considered, c)
	}
	return p.finish()
}

// CRC: crc-FinishedCarve.md | Seq: seq-links.md#4.4 | R490, R498
// incomingPlan rewrites each link and each pointer in file that resolves to the carve so it
// reaches dest: exactly what `query refs --to carve` lists for the file.
func incomingPlan(root, file, carve, dest string, report *FinishReport) (*plan, error) {
	p, err := readPlan(root, file)
	if err != nil {
		return nil, err
	}
	// R498 — pointers: the document bytes rewritten, the key kept. A pointer in a trajectory
	// file is written from the repository root, so its new form is dest itself.
	for _, r := range query.RefsIn(root, file, p.doc) {
		if r.Kind != "pointer" || r.Resolved != carve {
			continue
		}
		c := Considered{File: file, Line: r.Line, Old: r.Text, Outcome: Rewritten, New: pointerDoc(file, dest)}
		if err := p.doc.SetPointerDoc(r.Index, c.New); err != nil {
			return nil, err
		}
		p.changed = true
		report.Counts[Rewritten]++
		report.Considered = append(report.Considered, c)
	}
	for i, l := range p.doc.Links() {
		if cls, target := query.ClassifyLink(root, file, l); cls != "" || target != carve {
			continue
		}
		c := Considered{File: file, Line: l.Line(), Old: l.Dest, Outcome: Rewritten, New: newDest(file, dest, l)}
		if err := p.doc.SetDest(i, c.New); err != nil {
			return nil, err
		}
		p.changed = true
		report.Counts[Rewritten]++
		report.Considered = append(report.Considered, c)
	}
	return p.finish()
}

func readPlan(root, rel string) (*plan, error) {
	_, abs := relAbs(root, rel)
	src, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	return &plan{rel: rel, abs: abs, doc: minispecsdom.ParseMarkdown(string(src))}, nil
}

func (p *plan) finish() (*plan, error) {
	out, err := p.doc.Render()
	if err != nil {
		return nil, err
	}
	p.render = out
	return p, nil
}

// relDoc is target relative to file's directory, slash-separated.
func relDoc(file, target string) string {
	rel, err := filepath.Rel(filepath.FromSlash(path.Dir(file)), filepath.FromSlash(target))
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

// pointerDoc is target as a pointer in file writes it: dest itself in a trajectory file,
// where a pointer is written from the repository root, relative to the citing file's
// directory otherwise. R498, R500
func pointerDoc(file, target string) string {
	if query.IsTrajectoryFile(file) {
		return target
	}
	return relDoc(file, target)
}
