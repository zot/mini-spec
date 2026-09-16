// CRC: crc-Refs.md | Seq: seq-links.md#5 | R495, R496, R497
package query

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/parser"
)

// Ref is one reference in the inventory: a link or a pointer, where it stands, and where it
// resolves.
type Ref struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Text     string `json:"text"`               // as written
	Kind     string `json:"kind"`               // "link" or "pointer"
	Resolved string `json:"resolved,omitempty"` // repository-relative, "" when it does not resolve
	Index    int    `json:"-"`                  // the reader's index of the link or pointer, for a rewrite
	Fragment string `json:"-"`                  // the link's fragment or the pointer's key
}

// trajectoryFiles are the sited files at the repository root; a pointer in one resolves
// from the root, which is how every `Part` pointer is written.
var trajectoryFiles = []string{"PENDING.md", "CURRENT.md", "DONE.md"}

// CRC: crc-Refs.md | Seq: seq-links.md#5.3 | R497
//
// OwnedDocuments is the union the move and the repair work over: every tracked markdown
// file, plus every document the trajectory layer sites — the trajectory files, the carve
// directories and their done/ — whether git tracks them or not. Two questions: whom a
// pointer is for, which git answers, and who wrote it, which for a sited document is the
// tool (Bill, 2026-09-16).
func OwnedDocuments(root string) ([]string, error) {
	files, err := PublicDocuments(root)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, f := range files {
		seen[f] = true
	}
	add := func(rel string) {
		if seen[rel] {
			return
		}
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			return
		}
		seen[rel] = true
		files = append(files, rel)
	}
	for _, f := range trajectoryFiles {
		add(f)
	}
	for _, dir := range []string{parser.PublicCarvesDir, parser.PrivateCarvesDir} {
		for _, pat := range []string{"*.md", "done/*.md"} {
			matches, _ := filepath.Glob(filepath.Join(root, dir, pat))
			for _, abs := range matches {
				rel, _ := filepath.Rel(root, abs)
				add(filepath.ToSlash(rel))
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

// IsTrajectoryFile reports a sited file at the repository root.
func IsTrajectoryFile(rel string) bool { return slices.Contains(trajectoryFiles, rel) }

// CRC: crc-Refs.md | Seq: seq-links.md#5.1 | R495, R496
//
// RefsIn is every link and pointer in one document, resolved. A link resolves relative to
// the citing file's directory; a pointer likewise, except in a trajectory file, where it
// resolves from the repository root.
func RefsIn(root, file string, m *minispecsdom.Markdown) []Ref {
	var out []Ref
	for i, l := range m.Links() {
		r := Ref{File: file, Line: l.Line(), Text: l.Raw, Kind: "link", Index: i, Fragment: l.Fragment}
		if cls, target := classify(root, file, l); cls == "" {
			r.Resolved = target
		}
		out = append(out, r)
	}
	for i, p := range m.Pointers() {
		out = append(out, Ref{
			File: file, Line: p.Line(), Text: p.Raw, Kind: "pointer", Index: i, Fragment: p.Key,
			Resolved: resolvePointer(root, file, p.Doc),
		})
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Line < out[b].Line })
	return out
}

// resolvePointer is the document a pointer in file names, repository-relative, or "" when it
// leaves the tree or nothing is there.
func resolvePointer(root, file, doc string) string {
	base := path.Dir(file)
	if IsTrajectoryFile(file) {
		base = "."
	}
	rel := path.Clean(path.Join(base, doc))
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return ""
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		return ""
	}
	return rel
}

// CRC: crc-Refs.md | Seq: seq-links.md#5 | R495, R496
// Refs is the inventory over files (the owned documents when none), narrowed by to.
func Refs(root string, files []string, to string) ([]Ref, error) {
	if len(files) == 0 {
		owned, err := OwnedDocuments(root)
		if err != nil {
			return nil, err
		}
		files = owned
	}
	if to != "" {
		to = filepath.ToSlash(filepath.Clean(to))
	}
	var out []Ref
	for _, f := range files {
		rel, abs := relAbs(root, f)
		src, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		for _, r := range RefsIn(root, rel, minispecsdom.ParseMarkdown(string(src))) {
			if to == "" || r.Resolved == to {
				out = append(out, r)
			}
		}
	}
	return out, nil
}

// FormatRefs renders one line per reference.
func FormatRefs(refs []Ref) string {
	var b strings.Builder
	for _, r := range refs {
		res := r.Resolved
		if res == "" {
			res = "—"
		}
		fmt.Fprintf(&b, "%s:%d  %s  %s  %s\n", r.File, r.Line, r.Text, r.Kind, res)
	}
	fmt.Fprintf(&b, "%d references\n", len(refs))
	return b.String()
}
