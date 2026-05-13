// CRC: crc-Validate.md | Seq: seq-validate.md | R97, R98, R99, R100
package validate

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/zot/minispec/internal/project"
)

func mkSeqProject(t *testing.T, seqFiles map[string]string) *Validate {
	t.Helper()
	root := t.TempDir()
	designDir := filepath.Join(root, "design")
	if err := os.Mkdir(designDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range seqFiles {
		if err := os.WriteFile(filepath.Join(designDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	p := &project.Project{
		RootPath:  root,
		DesignDir: designDir,
		SrcDir:    filepath.Join(root, "src"),
		Config:    project.DefaultConfig(),
	}
	return New(p)
}

func TestValidateSeqNumbering_DetectsTreeGap(t *testing.T) {
	v := mkSeqProject(t, map[string]string{
		"seq-foo.md": "```\n1.1. one\n1.3. three\n```\n",
	})
	result := newResult()
	v.validateSeqNumbering(result)
	want := map[string][]string{"seq-foo.md": {"1.2"}}
	if !reflect.DeepEqual(result.SeqNumberingGaps, want) {
		t.Errorf("SeqNumberingGaps = %v, want %v", result.SeqNumberingGaps, want)
	}
}

func TestValidateSeqNumbering_DetectsKGap(t *testing.T) {
	v := mkSeqProject(t, map[string]string{
		"seq-foo.md": "```\n1.1. one\n3.1. three-one\n```\n",
	})
	result := newResult()
	v.validateSeqNumbering(result)
	want := map[string][]string{"seq-foo.md": {"2"}}
	if !reflect.DeepEqual(result.SeqNumberingGaps, want) {
		t.Errorf("SeqNumberingGaps = %v, want %v", result.SeqNumberingGaps, want)
	}
}

func TestValidateSeqNumbering_DetectsDuplicates(t *testing.T) {
	v := mkSeqProject(t, map[string]string{
		"seq-foo.md": "```\n1.1. first\n1.2. second\n1.1. duplicate\n```\n",
	})
	result := newResult()
	v.validateSeqNumbering(result)
	want := map[string][]string{"seq-foo.md": {"1.1"}}
	if !reflect.DeepEqual(result.SeqDuplicateIDs, want) {
		t.Errorf("SeqDuplicateIDs = %v, want %v", result.SeqDuplicateIDs, want)
	}
}

func TestValidateSeqNumbering_SkipsUnnumbered(t *testing.T) {
	v := mkSeqProject(t, map[string]string{
		"seq-uml.md":  "Agent -> Server: do thing\n",
		"seq-tree.md": "```\n1.1. one\n1.2. two\n```\n",
	})
	result := newResult()
	v.validateSeqNumbering(result)
	if len(result.SeqNumberingGaps) != 0 {
		t.Errorf("expected no gaps, got %v", result.SeqNumberingGaps)
	}
	if len(result.SeqDuplicateIDs) != 0 {
		t.Errorf("expected no dupes, got %v", result.SeqDuplicateIDs)
	}
}

func newResult() *ValidationResult {
	return &ValidationResult{
		UnknownCRCRefs:      make(map[string][]string),
		MissingDesignRefs:   make(map[string][]string),
		MissingCRCSequences: make(map[string][]string),
		MissingSeqFragments: make(map[string][]string),
		SeqNumberingGaps:    make(map[string][]string),
		SeqDuplicateIDs:     make(map[string][]string),
	}
}
