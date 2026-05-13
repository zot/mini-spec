// CRC: crc-Parser.md | Seq: seq-parse.md | R94, R95, R96, R98, R99, R100
package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeSeqFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "seq-test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSplitSeqRef(t *testing.T) {
	cases := []struct {
		in, wantFile, wantFrag string
	}{
		{"seq-foo.md", "seq-foo.md", ""},
		{"seq-foo.md#1.4", "seq-foo.md", "1.4"},
		{"seq-foo.md#1", "seq-foo.md", "1"},
		{"seq-foo.md#1.4.3.2.1", "seq-foo.md", "1.4.3.2.1"},
	}
	for _, c := range cases {
		f, fr := SplitSeqRef(c.in)
		if f != c.wantFile || fr != c.wantFrag {
			t.Errorf("SplitSeqRef(%q) = (%q, %q), want (%q, %q)", c.in, f, fr, c.wantFile, c.wantFrag)
		}
	}
}

func TestParseSeqDoc_TreeStyle(t *testing.T) {
	content := "# Diagram\n\n```\n" +
		"1. fsnotify event\n" +
		"  ├── 1.1. if new directory\n" +
		"  │     └── 1.1.1. watchDirRecursive\n" +
		"  ├── 1.2. if ark.toml changed\n" +
		"  │     ├── 1.2.1. Config.Load\n" +
		"  │     └── 1.2.2. clearIgnoredPaths\n" +
		"```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Numbered() {
		t.Fatal("expected Numbered=true")
	}
	for _, id := range []string{"1", "1.1", "1.1.1", "1.2", "1.2.1", "1.2.2"} {
		if !doc.Has(id) {
			t.Errorf("missing id %q", id)
		}
	}
	if got, want := doc.Ks, []int{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("Ks = %v, want %v", got, want)
	}
	if gaps := doc.NumberingGaps(); len(gaps) != 0 {
		t.Errorf("unexpected gaps: %v", gaps)
	}
}

func TestParseSeqDoc_UMLStyle(t *testing.T) {
	content := "# UML\n\n```\n" +
		"Agent                    Server\n" +
		"  |                        |\n" +
		"  |   1.1                  |\n" +
		"  |-- POST /api/ui_audit ->|\n" +
		"  |                        |   1.2\n" +
		"  |                        |-- AuditApp() ----->\n" +
		"  |                        |    1.3\n" +
		"  |                        |<-- AuditResult ----\n" +
		"  |    1.4                 |\n" +
		"  |<-- JSON response ------|\n" +
		"```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"1.1", "1.2", "1.3", "1.4"} {
		if !doc.Has(id) {
			t.Errorf("missing id %q", id)
		}
	}
	if gaps := doc.NumberingGaps(); len(gaps) != 0 {
		t.Errorf("unexpected gaps: %v", gaps)
	}
}

func TestParseSeqDoc_MultipleDiagrams(t *testing.T) {
	content := "```\n1. one\n1.1. one-one\n```\n\n```\n2. two\n2.1. two-one\n2.2. two-two\n```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := doc.Ks, []int{1, 2}; !reflect.DeepEqual(got, want) {
		t.Errorf("Ks = %v, want %v", got, want)
	}
	if gaps := doc.NumberingGaps(); len(gaps) != 0 {
		t.Errorf("unexpected gaps: %v", gaps)
	}
}

func TestParseSeqDoc_Unnumbered(t *testing.T) {
	content := "Agent -> Server: do thing\nServer -> Auditor: AuditApp()\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Numbered() {
		t.Errorf("expected Numbered=false, got items=%v", doc.Items)
	}
}

func TestParseSeqDoc_NumberingGapInTree(t *testing.T) {
	content := "```\n1.1. one\n1.3. three\n```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	gaps := doc.NumberingGaps()
	want := []string{"1.2"}
	if !reflect.DeepEqual(gaps, want) {
		t.Errorf("gaps = %v, want %v", gaps, want)
	}
}

func TestParseSeqDoc_KSequenceGap(t *testing.T) {
	content := "```\n1.1. one\n3.1. three-one\n```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	gaps := doc.NumberingGaps()
	want := []string{"2"}
	if !reflect.DeepEqual(gaps, want) {
		t.Errorf("gaps = %v, want %v", gaps, want)
	}
}

func TestParseSeqDoc_DuplicateID(t *testing.T) {
	content := "```\n1.1. first\n1.2. second\n1.1. duplicate\n```\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := doc.Dupes, []string{"1.1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Dupes = %v, want %v", got, want)
	}
}

func TestParseSeqDoc_NoFalsePositiveOnProse(t *testing.T) {
	content := "This file mentions version 2.5 in prose.\nThe rate is 1.5 things per second.\n"
	path := writeSeqFile(t, content)
	doc, err := ParseSeqDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Numbered() {
		t.Errorf("expected unnumbered, got %v", doc.Items)
	}
}
