// CRC: crc-Harvest.md | Seq: seq-harvest.md | R513, R514, R515
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/simple-dom/sdom"
)

// HarvestComment is one traceability comment as the checks consume it: where it is, and
// its fields. Refs are expanded — `R5-R8` is four entries — and spelled `Rn`, the way every
// other check spells them.
type HarvestComment struct {
	Line int
	Text string // the comment as written, on one line (R502)
	CRC  []string
	Seq  []string
	Refs []string
}

// FileHarvest is one code file's traceability comments, in document order.
type FileHarvest struct {
	Path     string
	Comments []HarvestComment
}

// UnreadFile is a code file the harvest could not read, or read only part of: Line is 0
// when none of it was read.
type UnreadFile struct {
	Path   string
	Line   int
	Reason string
}

// Where is an unread file's location: the file, or the line where reading stopped.
func (u UnreadFile) Where() string {
	if u.Line > 0 {
		return fmt.Sprintf("%s:%d", u.Path, u.Line)
	}
	return u.Path
}

// Harvest is every code file's traceability, and every file it could not read.
type Harvest struct {
	Files  []FileHarvest
	Unread []UnreadFile
}

// CRC: crc-Harvest.md | Seq: seq-harvest.md#1 | R513
// HarvestArtifacts reads every code file the manifest lists, once each, in manifest order.
// A listed file that does not exist is left to validate, which reports it as a missing
// artifact.
func HarvestArtifacts(root string, artifacts []Artifact, configured minispecsdom.Configured) (Harvest, error) {
	var h Harvest
	seen := map[string]bool{}
	for _, art := range artifacts {
		for _, cf := range art.CodeFiles {
			// step 1.1
			if seen[cf.Path] {
				continue
			}
			seen[cf.Path] = true
			// step 1.2
			if _, err := os.Stat(filepath.Join(root, cf.Path)); err != nil {
				continue
			}
			// step 1.3
			fh, unread, err := HarvestFile(root, cf.Path, configured)
			if err != nil {
				return h, err
			}
			if unread != nil {
				h.Unread = append(h.Unread, *unread)
			}
			// A file read up to a group left open still counts what came before it.
			if unread == nil || unread.Line > 0 {
				h.Files = append(h.Files, fh)
			}
		}
	}
	return h, nil
}

// CRC: crc-Harvest.md | Seq: seq-harvest.md#2 | R513, R514, R515
// HarvestFile reads one code file, path relative to root, through the table its extension
// names. The unread entry is non-nil when the file has no table (nothing read, line 0) or
// when its parse left a string or comment open at end of input — inside one nothing is
// recognized, so everything after its opener went unsearched. A stray closer or an open
// code bracket hides no comment and is not reported: comments are recognized inside code
// brackets, and an HTML page's text is full of unmatched `)`.
func HarvestFile(root, path string, configured minispecsdom.Configured) (FileHarvest, *UnreadFile, error) {
	fh := FileHarvest{Path: path}
	// step 2.1
	ext := filepath.Ext(path)
	lang, ok := minispecsdom.LanguageFor(ext, configured)
	if !ok {
		return fh, &UnreadFile{Path: path, Reason: fmt.Sprintf("no language for %s", ext)}, nil
	}
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return fh, nil, err
	}
	src := string(data)
	// step 2.2
	bp := sdom.NewBracketParser(lang)
	doc := sdom.Parse(src, 0, bp)
	ctx := bp.Context()
	// step 2.3
	comments, err := minispecsdom.Comments(doc, ctx)
	if err != nil {
		return fh, nil, fmt.Errorf("%s: %w", path, err)
	}
	// step 2.4
	for _, c := range comments {
		text, _ := c.Render()
		hc := HarvestComment{Line: lineAt(src, c.Location().Offset()), Text: strings.Join(strings.Fields(text), " ")}
		if c.CRC() != nil {
			hc.CRC = c.CRC().Items()
		}
		if c.Seq() != nil {
			hc.Seq = c.Seq().Items()
		}
		if c.Refs() != nil {
			for _, n := range c.Refs().Items() {
				hc.Refs = append(hc.Refs, "R"+strconv.Itoa(n))
			}
		}
		fh.Comments = append(fh.Comments, hc)
	}
	// step 2.5
	for _, o := range ctx.Unclosed() {
		text, _ := o.Render()
		if g := lang.GroupFor(text); g != nil && g.Restricted() {
			return fh, &UnreadFile{Path: path, Line: lineAt(src, o.Location().Offset()),
				Reason: fmt.Sprintf("%q opened and never closed, so nothing after it was read", text)}, nil
		}
	}
	return fh, nil, nil
}

// lineAt is the 1-based line holding byte offset off.
func lineAt(src string, off int) int {
	if off < 0 || off > len(src) {
		return 0
	}
	return strings.Count(src[:off], "\n") + 1
}
