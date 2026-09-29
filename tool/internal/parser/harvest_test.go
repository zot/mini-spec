// CRC: crc-Harvest.md | Seq: seq-harvest.md | R508, R509, R513, R514, R515, R527
package parser

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/simple-dom/sdom"
)

// gc writes a Go traceability comment from the table's own style, so no line in this file
// reads as a real one to a harvester scanning it.
func gc(interior string) string { return sdom.LangGo.Comment.Prefix + interior + "\n" }

// tree writes files under a fresh root and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for p, body := range files {
		full := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func harvestOne(t *testing.T, name, body string, configured minispecsdom.Configured) (FileHarvest, *UnreadFile) {
	t.Helper()
	root := tree(t, map[string]string{name: body})
	fh, unread, err := HarvestFile(root, name, configured)
	if err != nil {
		t.Fatal(err)
	}
	return fh, unread
}

func allRefs(fh FileHarvest) []string {
	var out []string
	for _, c := range fh.Comments {
		out = append(out, c.Refs...)
	}
	return out
}

// R513, R502 — a comment is found with its line, text and fields, ranges expanded.
func TestHarvestFindsACommentWithItsLineAndFields(t *testing.T) {
	fh, unread := harvestOne(t, "a.go", "package a\n\n"+gc("CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5-7")+"func A() {}\n", nil)
	if unread != nil || len(fh.Comments) != 1 {
		t.Fatalf("unread %v, %d comments", unread, len(fh.Comments))
	}
	c := fh.Comments[0]
	if c.Line != 3 || !slices.Equal(c.CRC, []string{"crc-Store.md"}) || !slices.Equal(c.Seq, []string{"seq-crud.md#1.4"}) ||
		!slices.Equal(c.Refs, []string{"R4", "R5", "R6", "R7"}) ||
		c.Text != "// CRC: crc-Store.md | Seq: seq-crud.md#1.4 | R4, R5-7" {
		t.Errorf("got %+v", c)
	}
	// R502 — a block comment over two lines prints on one.
	fh, _ = harvestOne(t, "b.go", "package b\n/* CRC: crc-Store.md\n   | R4 */\n", nil)
	if len(fh.Comments) != 1 || fh.Comments[0].Text != "/* CRC: crc-Store.md | R4 */" {
		t.Errorf("block comment text: %+v", fh.Comments)
	}
}

// R513 — no traceability comment is a finding for validate, not a failure to read.
func TestHarvestOfAFileWithNoCommentIsNotUnread(t *testing.T) {
	fh, unread := harvestOne(t, "a.go", "package a\n\n// ordinary prose\nfunc A() {}\n", nil)
	if unread != nil || len(fh.Comments) != 0 {
		t.Errorf("unread %v, comments %+v", unread, fh.Comments)
	}
}

// R508, R514 — the three classes measured 2026-09-25: a Seq-only comment, a bare one and
// the `Rn.` form count; prose mentions, parentheticals and a quoted leader do not.
func TestHarvestCountsEveryShapeTheGrammarReads(t *testing.T) {
	body := "package a\n" +
		gc("Seq: seq-x.md#2.2 | R11") +
		gc("R12: note") +
		gc("R13. Prose follows the full stop") +
		gc("see R14") +
		gc("computed lazily (R15)") +
		gc("the leader (e.g. `// R16: desc`) in prose")
	fh, _ := harvestOne(t, "a.go", body, nil)
	if got := allRefs(fh); !slices.Equal(got, []string{"R11", "R12", "R13"}) {
		t.Errorf("refs %v, want [R11 R12 R13]", got)
	}
}

// R515 — an extension with no table is unread, with its reason.
func TestHarvestOfAnUnmappedExtensionIsUnread(t *testing.T) {
	_, unread := harvestOne(t, "x.zig", "// "+"R1\n", nil)
	if unread == nil || unread.Line != 0 || unread.Reason != "no language for .zig" {
		t.Errorf("unread %+v", unread)
	}
}

// R515 — a string left open swallows what follows; what came before still counts.
func TestHarvestOfAnUnclosedStringKeepsWhatWasRead(t *testing.T) {
	body := gc("R1") + "package a\n" + "var s = \"never closed\n" + gc("R2")
	fh, unread := harvestOne(t, "a.go", body, nil)
	if unread == nil || unread.Line != 3 {
		t.Fatalf("unread %+v, want line 3", unread)
	}
	if got := allRefs(fh); !slices.Equal(got, []string{"R1"}) {
		t.Errorf("refs %v, want [R1]", got)
	}
}

// R515 — stray closers and open code brackets hide nothing and are not reported.
func TestHarvestIgnoresStrayClosersAndOpenCodeBrackets(t *testing.T) {
	if _, unread := harvestOne(t, "p.html", "<p>(see note) and a stray } here</p>\n", nil); unread != nil {
		t.Errorf("html page text reported unread: %+v", unread)
	}
	fh, unread := harvestOne(t, "a.go", "package a\nfunc A() {\n"+gc("R3"), nil)
	if unread != nil {
		t.Errorf("an open code bracket reported unread: %+v", unread)
	}
	if got := allRefs(fh); !slices.Equal(got, []string{"R3"}) {
		t.Errorf("refs %v, want [R3]", got)
	}
}

// R527 — a configured language wins for the extensions it names.
func TestHarvestUsesAConfiguredLanguageFirst(t *testing.T) {
	toy := &sdom.BracketLang{Comment: sdom.CommentStyle{Prefix: "## ", Suffix: "\n", Kind: "comment"},
		Brackets: []sdom.BracketGroup{{Open: []string{"##"}, Close: "\n", AllowedInner: []string{}, Kind: "comment"}}}
	fh, _ := harvestOne(t, "a.go", "## R1\n"+gc("R2"), minispecsdom.Configured{".go": toy})
	if got := allRefs(fh); !slices.Equal(got, []string{"R1"}) {
		t.Errorf("refs %v, want [R1] — the configured table, not Go's", got)
	}
}

// R513 — manifest order, each file once, a missing file left to validate.
func TestHarvestReadsTheManifestInOrderOnce(t *testing.T) {
	root := tree(t, map[string]string{"a.go": gc("R1"), "b.go": gc("R2")})
	arts := []Artifact{
		{DesignFile: "crc-A.md", CodeFiles: []CodeFile{{Path: "a.go"}, {Path: "b.go"}}},
		{DesignFile: "crc-B.md", CodeFiles: []CodeFile{{Path: "a.go"}, {Path: "gone.go"}}},
	}
	h, err := HarvestArtifacts(root, arts, nil)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, f := range h.Files {
		paths = append(paths, f.Path)
	}
	if !slices.Equal(paths, []string{"a.go", "b.go"}) || len(h.Unread) != 0 {
		t.Errorf("files %v unread %+v", paths, h.Unread)
	}
}

// R509 — another language's comment form is read by its table, not by a pattern.
func TestHarvestReadsPythonThroughItsTable(t *testing.T) {
	fh, unread := harvestOne(t, "a.py", "import os\n# CRC"+": crc-Store.md | Seq: seq-crud.md\ndef add(): pass\n", nil)
	if unread != nil || len(fh.Comments) != 1 || !slices.Equal(fh.Comments[0].CRC, []string{"crc-Store.md"}) {
		t.Errorf("unread %v comments %+v", unread, fh.Comments)
	}
}
