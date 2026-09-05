package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/simple-dom/minispecsdom"
	"github.com/zot/simple-dom/sdom"
)

// CRC: crc-Carve.md | R208
// The two carve directories the format sites at the repository root. carves/done/ is a
// subdirectory of the first and is never entered.
const (
	PublicCarvesDir  = "carves"
	PrivateCarvesDir = ".carves"
)

// ErrNoCarveDirs reports a repository with neither carve directory. R214
var ErrNoCarveDirs = errors.New("no carves/ or .carves/ directory at the repository root")

// CRC: crc-Carve.md | R207, R214
// CarveScan is every carve read from the repository root, and which directories existed.
type CarveScan struct {
	Dirs   []CarveDir `json:"dirs"`
	Carves []Carve    `json:"carves"`
}

// CarveDir reports a carve directory present or absent — no directory at all has no
// answer, and an empty census is a confident wrong one. R214
type CarveDir struct {
	Name    string `json:"name"`
	Present bool   `json:"present"`
}

// CRC: crc-Carve.md | R211, R216
// Carve is one document as this tool sees it: the reader's parts and stateless lines,
// under a repository-relative path.
type Carve struct {
	Path      string      `json:"path"` // repository-relative, slash-separated
	HasStatus bool        `json:"has_status"`
	Parts     []Part      `json:"parts"`
	Stateless []Stateless `json:"stateless"`
}

// PartState is what the checkbox says. A stateless line has none and is not a Part.
type PartState string

const (
	PartOpen   PartState = "open"
	PartLanded PartState = "landed"
)

// CRC: crc-Carve.md | R219
//
// Part is the reader's part with the census's derived view over it. The rules that read the
// line are the dependency's; this type only names what the census prints.
type Part struct {
	part *minispecsdom.Part
}

// Key is the key as written, "" when the line does not key.
func (p Part) Key() string { return p.part.Key() }

// Keyed reports whether the line keys in the format's form.
func (p Part) Keyed() bool { return p.part.Key() != "" }

// State is open or landed, from the checkbox.
func (p Part) State() PartState {
	if p.part.Checkbox().Checked() {
		return PartLanded
	}
	return PartOpen
}

// Title is the head's whole title — from the reader's title node to the head's closing
// `**` — so a title holding a code span is not cut at its first backtick. The reader's
// Title() is the first text alone, which is right for a bound field and short for a census.
func (p Part) Title() string {
	t := p.part.Title()
	if t == nil {
		return ""
	}
	var b strings.Builder
	started := false
	for _, k := range p.part.Kids() {
		if k == sdom.Node(t) {
			started = true
		}
		if !started {
			continue
		}
		if c, ok := k.(*sdom.Closer); ok && rendered(c) == "**" {
			break
		}
		b.WriteString(rendered(k))
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(b.String()), "."))
}

// QueueID is the #N inside the first marker attribution that carries one — `OPEN (#N.)`,
// `REVERTED (#N.)` or a `LANDED (…— `#N`.)` record — and 0 when none does. Scoped to the
// markers so a superseded bare-#N key is never read as a queue reference; widened past
// `OPEN` on 2026-09-05 because a landed part's queue ID lives in its record and nowhere
// else, and the orphan check (R291) reads it from there. R287
func (p Part) QueueID() int {
	for _, m := range p.part.Markers() {
		if id, ok := m.QueueID(); ok {
			return id
		}
	}
	return 0
}

// Line is the part's 1-based line in its document.
func (p Part) Line() int { return p.part.Line() }

// Depth is the bullet's leading whitespace; 0 is top level.
func (p Part) Depth() int { return p.part.Depth }

// Deviations are the rules the line breaks, each with its target, as the reader reports
// them. R219
func (p Part) Deviations() []minispecsdom.Deviation { return p.part.Deviations() }

// Struck reports whether the head's bold run is wrapped in `~~`, read from the line's own
// nodes by the dependency. R288
func (p Part) Struck() bool { return p.part.IsStruck() }

// Verbs are the marker verbs on the line, upper-cased, in order. R288
func (p Part) Verbs() []string {
	var out []string
	for _, m := range p.part.Markers() {
		v, _ := m.Verb().Render()
		out = append(out, strings.ToUpper(strings.TrimSpace(v)))
	}
	return out
}

// Conforms is the absence of deviations.
func (p Part) Conforms() bool { return len(p.part.Deviations()) == 0 }

// MarshalJSON names the derived view so a reader without the Go source open can read it.
func (p Part) MarshalJSON() ([]byte, error) {
	type view struct {
		Key        string                   `json:"key"`
		Keyed      bool                     `json:"keyed"`
		Depth      int                      `json:"depth"`
		Line       int                      `json:"line"`
		State      PartState                `json:"state"`
		Title      string                   `json:"title"`
		QueueID    int                      `json:"queue_id,omitempty"`
		Deviations []minispecsdom.Deviation `json:"deviations,omitempty"`
	}
	return json.Marshal(view{p.Key(), p.Keyed(), p.Depth(), p.Line(), p.State(), p.Title(), p.QueueID(), p.Deviations()})
}

// CRC: crc-Carve.md | R216
//
// Stateless is a status-block line with no checkbox: existence, a line and a reason, never
// a state. The dependency numbers parts and not these, so the line is derived from the
// node's offset in the source.
type Stateless struct {
	Line       int                      `json:"line"`
	Reason     string                   `json:"reason"`
	Text       string                   `json:"text"`
	Deviations []minispecsdom.Deviation `json:"deviations,omitempty"`
}

// CRC: crc-Carve.md | Seq: seq-carve-status.md#1.3 | R208, R214
// ScanCarves reads every *.md directly in carves/ and .carves/; carves/done/ is not
// entered; both directories absent is ErrNoCarveDirs.
func ScanCarves(repoRoot string) (CarveScan, error) {
	var scan CarveScan
	anyPresent := false
	for _, name := range []string{PublicCarvesDir, PrivateCarvesDir} {
		// os.ReadDir sorts by name, so the report's order is stable without sorting here.
		entries, err := os.ReadDir(filepath.Join(repoRoot, name))
		if errors.Is(err, os.ErrNotExist) {
			scan.Dirs = append(scan.Dirs, CarveDir{Name: name})
			continue
		}
		if err != nil {
			return scan, err
		}
		anyPresent = true
		scan.Dirs = append(scan.Dirs, CarveDir{Name: name, Present: true})
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			c, err := ReadCarve(filepath.Join(repoRoot, name, e.Name()), name+"/"+e.Name())
			if err != nil {
				return scan, err
			}
			scan.Carves = append(scan.Carves, c)
		}
	}
	if !anyPresent {
		return scan, ErrNoCarveDirs
	}
	return scan, nil
}

// CRC: crc-Carve.md | Seq: seq-carve-status.md#2 | R209, R210, R211, R216
// ReadCarve reads one document through the dependency's reader and takes its parts and
// stateless lines as handed — the status region, the fence rule and the subpart depth are
// the reader's, not re-scanned here.
func ReadCarve(path, rel string) (Carve, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return Carve{}, err
	}
	return parseCarve(rel, string(src)), nil
}

func parseCarve(rel, src string) Carve {
	rc := minispecsdom.ParseCarve(src)
	c := Carve{Path: rel, HasStatus: rc.HasStatus()}
	for _, p := range rc.Parts() {
		c.Parts = append(c.Parts, Part{part: p})
	}
	for _, l := range rc.Stateless() {
		text, _, _ := strings.Cut(rendered(l), "\n")
		c.Stateless = append(c.Stateless, Stateless{
			Line:       lineOf(src, l.Location().Offset()),
			Reason:     "no checkbox",
			Text:       strings.TrimSpace(text),
			Deviations: l.Deviations(),
		})
	}
	return c
}

// rendered is a node's source text. Render's error is the document's to report, not a
// census field's, and every read path here drops it the same way.
func rendered(n sdom.Node) string {
	s, _ := n.Render()
	return s
}

// lineOf is the 1-based line holding byte offset off. R216
func lineOf(src string, off int) int {
	if off > len(src) {
		off = len(src)
	}
	return 1 + strings.Count(src[:off], "\n")
}

// CRC: crc-Carve.md | R218
func (c Carve) Open() int    { return c.count(func(p Part) bool { return p.State() == PartOpen }) }
func (c Carve) Landed() int  { return c.count(func(p Part) bool { return p.State() == PartLanded }) }
func (c Carve) Unkeyed() int { return c.count(func(p Part) bool { return !p.Keyed() }) }

// NonConforming counts every status-block line carrying a deviation, on either side of the
// checkbox. R212, R216
func (c Carve) NonConforming() int {
	n := c.count(func(p Part) bool { return !p.Conforms() })
	for _, s := range c.Stateless {
		if len(s.Deviations) > 0 {
			n++
		}
	}
	return n
}

func (c Carve) count(pred func(Part) bool) int {
	n := 0
	for _, p := range c.Parts {
		if pred(p) {
			n++
		}
	}
	return n
}

// CRC: crc-Carve.md | R218
func (s CarveScan) Open() int          { return s.total(Carve.Open) }
func (s CarveScan) Landed() int        { return s.total(Carve.Landed) }
func (s CarveScan) Unkeyed() int       { return s.total(Carve.Unkeyed) }
func (s CarveScan) NonConforming() int { return s.total(Carve.NonConforming) }
func (s CarveScan) Stateless() int     { return s.total(func(c Carve) int { return len(c.Stateless) }) }
func (s CarveScan) NoStatus() int {
	n := 0
	for _, c := range s.Carves {
		if !c.HasStatus {
			n++
		}
	}
	return n
}
func (s CarveScan) WithStatus() int { return len(s.Carves) - s.NoStatus() }

func (s CarveScan) total(f func(Carve) int) int {
	n := 0
	for _, c := range s.Carves {
		n += f(c)
	}
	return n
}

// CRC: crc-Carve.md | Seq: seq-carve-status.md#3 | R220
// SetMarker writes one part's marker through the reader's rule. The reader's refusals pass
// through and the file is untouched by them.
func SetMarker(path, key, verb, attribution string) error {
	return editCarve(path, func(c *minispecsdom.Carve) error {
		return c.SetMarker(key, verb, attribution)
	})
}

// CRC: crc-Carve.md | Seq: seq-carve-status.md#3 | R220
// SetPartLanded is the completion write — box, strike and LANDED record in one act — through
// the reader's Land.
func SetPartLanded(path, key, attribution string) error {
	return editCarve(path, func(c *minispecsdom.Carve) error {
		return c.Land(key, attribution)
	})
}

// CRC: crc-Carve.md | R220
// PartIsLanded reports the keyed part's checkbox; false and no error when no part keys so.
func PartIsLanded(path, key string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	p := minispecsdom.ParseCarve(string(src)).Part(key)
	return p != nil && p.Checkbox().Checked(), nil
}

// editCarve reads, applies, and writes by temp-file-and-rename. The write happens only
// after the edit returned nil, so a refusal leaves the file byte-identical. R220
func editCarve(path string, edit func(*minispecsdom.Carve) error) error {
	return editFile(path, func(src string) (string, error) {
		c := minispecsdom.ParseCarve(src)
		if err := edit(c); err != nil {
			return "", err
		}
		return c.Render()
	})
}

// editFile is the one atomic write every adapter shares: read, hand the bytes to render, and
// replace the file by temp-file-and-rename only when render returned nil. R220
func editFile(path string, render func(src string) (string, error)) (err error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// A write the reader cannot read back panics inside the dependency with a ReadBackError:
	// a library invariant, not caller input, so it is not an error a caller could swallow. Here
	// it becomes a refusal naming the file, and the file stays untouched because nothing has
	// been written yet. Any other panic is still a panic. R220
	defer func() {
		if r := recover(); r != nil {
			if rb, ok := r.(*minispecsdom.ReadBackError); ok {
				err = fmt.Errorf("%s: %w — nothing was written", path, rb)
				return
			}
			panic(r)
		}
	}()
	out, err := render(string(src))
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	// Cleans up every failure below; a no-op once the rename has moved the file away.
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(out); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}
