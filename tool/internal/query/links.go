// CRC: crc-Links.md | Seq: seq-links.md#2 | R455, R456, R457, R459, R460
package query

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

// LinkClass is where a link was found to point. Exactly one per link. R456
type LinkClass string

const (
	ClassTracked   LinkClass = "tracked"
	ClassUntracked LinkClass = "untracked"
	ClassIgnored   LinkClass = "ignored"
	ClassMissing   LinkClass = "missing"
	ClassOutside   LinkClass = "outside"
	ClassExternal  LinkClass = "external"
	ClassLocal     LinkClass = "local"
)

// LinkClasses is every class, in the order the report states them.
var LinkClasses = []LinkClass{ClassTracked, ClassUntracked, ClassIgnored, ClassMissing, ClassOutside, ClassExternal, ClassLocal}

// CRC: crc-Links.md | R457
// IsError and IsWarning are the 2026-08-04 decision: error on ignored (and the two states
// that can never resolve for a cloner), warn on untracked — usually just early.
func (c LinkClass) IsError() bool   { return c == ClassIgnored || c == ClassMissing || c == ClassOutside }
func (c LinkClass) IsWarning() bool { return c == ClassUntracked }

// Decided reports whether the class carries a decision — an error or a warning.
func (c LinkClass) Decided() bool { return c.IsError() || c.IsWarning() }

// CheckedLink is one link, resolved and classified.
type CheckedLink struct {
	File     string    `json:"file"` // the citing file, repository-relative
	Line     int       `json:"line"`
	Link     string    `json:"link"`               // as written
	Resolved string    `json:"resolved,omitempty"` // repository-relative, for the disk classes
	Class    LinkClass `json:"class"`
}

// LinkReport is every link in the population with the count per class, zeros included.
type LinkReport struct {
	Files  []string          `json:"files"`
	Links  []CheckedLink     `json:"links"`
	Counts map[LinkClass]int `json:"counts"`
}

// Errors reports whether any link is an error. R461's exit status.
func (r *LinkReport) Errors() bool {
	for _, l := range r.Links {
		if l.Class.IsError() {
			return true
		}
	}
	return false
}

// ErrNoGitTree is the refusal outside a working tree: a fossil-only project's references
// go unchecked, and the tool says so rather than passing silently. R460
var ErrNoGitTree = errors.New("not a git working tree, so links cannot be classified")

// CRC: crc-Links.md | Seq: seq-links.md#2.1 | R455, R463
// LiveCarves is the live-carve population, shared with the move repair.
func LiveCarves(root string) ([]string, error) { return defaultPopulation(root) }

// CRC: crc-Links.md | Seq: seq-links.md#2.1 | R463, R485
// PublicCarves is every live carve and every `*.md` directly under each carve directory's
// `done/`: the public documents, shared by the move repair and the validation phase.
func PublicCarves(root string) ([]string, error) {
	files, err := defaultPopulation(root)
	if err != nil {
		return nil, err
	}
	for _, dir := range []string{"carves", ".carves"} {
		done, err := filepath.Glob(filepath.Join(root, dir, "done", "*.md"))
		if err != nil {
			return nil, err
		}
		for _, p := range done {
			files = append(files, filepath.ToSlash(filepath.Join(dir, "done", filepath.Base(p))))
		}
	}
	return files, nil
}

// CRC: crc-Links.md | Seq: seq-links.md#2.3.1 | R456, R464
// ClassifyLink is classify for the move repair: the class a link has before git is asked,
// "" with its repository-relative path when it exists inside the tree.
func ClassifyLink(root, file string, l minispecsdom.Link) (LinkClass, string) {
	return classify(root, file, l)
}

// CRC: crc-Links.md | Seq: seq-links.md#2.1 | R455
// defaultPopulation is the live carves — every `*.md` directly in carves/ and .carves/,
// never carves/done/ — as `query carves` reads them.
func defaultPopulation(root string) ([]string, error) {
	scan, err := parser.ScanCarves(root)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(scan.Carves))
	for _, cv := range scan.Carves {
		files = append(files, cv.Path)
	}
	return files, nil
}

// CRC: crc-Links.md | Seq: seq-links.md#2 | R455, R456, R459, R460
//
// CheckLinks resolves and classifies every link in files (repository-relative or absolute;
// the live carves when none) against root's working tree. Git is asked once per citing
// file, after the classes that need no disk and no git are decided, so a URL is never asked
// of the disk and a missing file never of git.
func CheckLinks(root string, files []string) (*LinkReport, error) {
	git := project.NewGit(root)
	if !git.IsRepo() {
		return nil, ErrNoGitTree
	}
	if len(files) == 0 {
		population, err := defaultPopulation(root)
		if err != nil {
			return nil, err
		}
		files = population
	}
	report := &LinkReport{Counts: map[LinkClass]int{}}
	for _, c := range LinkClasses {
		report.Counts[c] = 0
	}
	for _, f := range files {
		rel, abs := relAbs(root, f)
		report.Files = append(report.Files, rel)
		src, err := os.ReadFile(abs)
		if err != nil {
			return nil, err
		}
		checked, err := classifyFile(root, rel, git, minispecsdom.ParseMarkdown(string(src)).Links())
		if err != nil {
			return nil, err
		}
		for _, l := range checked {
			report.Counts[l.Class]++
		}
		report.Links = append(report.Links, checked...)
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

// CRC: crc-Links.md | Seq: seq-links.md#2.3 | R456, R459, R460
// classifyFile decides every link in one citing file, then asks git once about the ones
// that reached the disk.
func classifyFile(root, file string, git *project.Git, links []minispecsdom.Link) ([]CheckedLink, error) {
	out := make([]CheckedLink, len(links))
	var ask []int
	for i, l := range links {
		cls, resolved := classify(root, file, l)
		out[i] = CheckedLink{File: file, Line: l.Line(), Link: l.Raw, Resolved: resolved, Class: cls}
		if cls == "" {
			ask = append(ask, i)
		}
	}
	if len(ask) == 0 {
		return out, nil
	}
	paths := make([]string, len(ask))
	for j, i := range ask {
		paths[j] = out[i].Resolved
	}
	ignored, err := git.Ignored(paths)
	if err != nil {
		return nil, err
	}
	for _, i := range ask {
		if ignored[out[i].Resolved] {
			out[i].Class = ClassIgnored
			continue
		}
		tracked, err := git.Tracked(out[i].Resolved)
		if err != nil {
			return nil, err
		}
		if tracked {
			out[i].Class = ClassTracked
		} else {
			out[i].Class = ClassUntracked
		}
	}
	return out, nil
}

// CRC: crc-Links.md | Seq: seq-links.md#2.3.1 | R456, R459
//
// classify decides the classes that need no git — external, local, outside, missing — and
// returns "" with the resolved repository-relative path for a link that exists inside the
// tree, which git then places. The order is the constraint: resolution happens before the
// disk is asked, so `../escape.md` is `outside` even when the parent happens to hold it.
func classify(root, file string, l minispecsdom.Link) (LinkClass, string) {
	if hasScheme(l.Dest) {
		return ClassExternal, ""
	}
	if l.Path == "" {
		return ClassLocal, ""
	}
	if filepath.IsAbs(l.Path) {
		return ClassOutside, l.Path
	}
	abs := filepath.Join(root, filepath.Dir(file), filepath.FromSlash(l.Path))
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ClassOutside, filepath.ToSlash(rel)
	}
	rel = filepath.ToSlash(rel)
	if _, err := os.Stat(abs); err != nil {
		return ClassMissing, rel
	}
	return "", rel
}

// hasScheme is a URL: a scheme, then `:`, before any `/` or `#`.
func hasScheme(dest string) bool {
	i := strings.IndexByte(dest, ':')
	if i <= 0 {
		return false
	}
	if j := strings.IndexAny(dest, "/#"); j >= 0 && j < i {
		return false
	}
	for _, r := range dest[:i] {
		if !isSchemeRune(r) {
			return false
		}
	}
	return true
}

// isSchemeRune is a character a URL scheme may hold: RFC 3986's ALPHA / DIGIT / "+" / "-" / ".".
func isSchemeRune(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '+' || r == '-' || r == '.'
}

// CRC: crc-Links.md | Seq: seq-links.md#2.5 | R458
// Summary is the closing count: every class, zeros included, in a fixed order.
func (r *LinkReport) Summary() string {
	parts := make([]string, 0, len(LinkClasses))
	for _, c := range LinkClasses {
		parts = append(parts, fmt.Sprintf("%s %d", c, r.Counts[c]))
	}
	return fmt.Sprintf("%d links in %d files: %s", len(r.Links), len(r.Files), strings.Join(parts, ", "))
}
