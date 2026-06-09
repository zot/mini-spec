// CRC: crc-Update.md | Seq: seq-update.md | R103
package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zot/minispec/internal/project"
)

// mkRetireProject builds a throwaway project with a requirements.md (one
// feature with a **Source:**) and a design.md carrying an empty Gaps section,
// the minimum Retire needs.
func mkRetireProject(t *testing.T, reqs, design string) *Update {
	t.Helper()
	dir := t.TempDir()
	designDir := filepath.Join(dir, "design")
	if err := os.MkdirAll(designDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(designDir, "requirements.md"), []byte(reqs), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(designDir, "design.md"), []byte(design), 0o644); err != nil {
		t.Fatal(err)
	}
	return New(&project.Project{
		RootPath:  dir,
		DesignDir: designDir,
		SrcDir:    filepath.Join(dir, "src"),
		Config:    project.DefaultConfig(),
	})
}

// TestRetire_ReturnsSources is the Sleeping-Sentry guard for the
// supersede-at-source reminder (R103): Retire must surface the retired
// requirement's **Source:** spec(s) so the CLI can name the prose to
// reconcile. If a future change drops Sources from the return, the reminder
// goes blind and this test fails. R103
func TestRetire_ReturnsSources(t *testing.T) {
	reqs := `# Requirements

## Feature: Storage
**Source:** specs/storage.md

- **R1:** the old behavior
- **R2:** the new behavior
`
	u := mkRetireProject(t, reqs, "# Design\n\n## Gaps\n")

	tn, sources, err := u.Retire("R1", "R2", "moved behavior")
	if err != nil {
		t.Fatal(err)
	}
	if tn != "T1" {
		t.Errorf("tn = %q, want T1", tn)
	}
	if len(sources) != 1 || sources[0] != "specs/storage.md" {
		t.Errorf("sources = %v, want [specs/storage.md]", sources)
	}
}

// TestRetire_NoSource covers the reminder's fallback branch: a requirement
// whose feature has no **Source:** line yields no sources, and the CLI says so
// rather than naming a file. R103
func TestRetire_NoSource(t *testing.T) {
	reqs := `# Requirements

## Feature: Orphan

- **R1:** the old behavior
`
	u := mkRetireProject(t, reqs, "# Design\n\n## Gaps\n")

	_, sources, err := u.Retire("R1", "-", "removed outright")
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 {
		t.Errorf("sources = %v, want empty", sources)
	}
}
