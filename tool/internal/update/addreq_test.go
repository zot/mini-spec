// CRC: crc-Update.md | R324, R325, R326
package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/project"
)

const reqsDoc = `# Requirements

## Feature: Alpha
**Source:** specs/alpha.md

- **R1:** first
- **~~R2:~~** (Retired T1 — no replacement) gone

### Notes

- **R4:** a note-level requirement

## Feature: Beta
**Source:** specs/beta.md

- **R3:** third
`

const gapsDoc = `# Design

## Gaps

- [ ] O1: open one
- [x] O2: done
- A1: approved
- T1: R2 retired (reason)

## Other
`

func reqProject(t *testing.T) (*Update, string, string) {
	t.Helper()
	root := t.TempDir()
	design := filepath.Join(root, "design")
	if err := os.MkdirAll(design, 0o755); err != nil {
		t.Fatal(err)
	}
	rp := filepath.Join(design, "requirements.md")
	dp := filepath.Join(design, "design.md")
	for p, body := range map[string]string{rp: reqsDoc, dp: gapsDoc} {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Update{Project: &project.Project{RootPath: root, DesignDir: design}}, rp, dp
}

// R324 — mint and append in one act, numbers counting retired ones, before the first
// sub-heading, in batch order, with or without the Feature: prefix; unknown or ambiguous
// sections refused with nothing written.
func TestAddReqMintsAppendsAndRefuses(t *testing.T) {
	u, rp, _ := reqProject(t)
	ids, err := u.AddReq("Alpha", []string{"fifth", "sixth"})
	if err != nil || strings.Join(ids, " ") != "R5 R6" {
		t.Fatalf("ids %v err %v; want R5 R6 — the retired R2 and the note-level R4 both count", ids, err)
	}
	got := read(t, rp)
	if !strings.Contains(got, "- **~~R2:~~** (Retired T1 — no replacement) gone\n- **R5:** fifth\n- **R6:** sixth\n\n### Notes") {
		t.Errorf("the entries did not land at the end of the section's own content, before its sub-heading:\n%s", got)
	}
	if ids, err := u.AddReq("Feature: Beta", []string{"fourth"}); err != nil || ids[0] != "R7" {
		t.Errorf("Feature: prefix: %v %v", ids, err)
	}
	before := read(t, rp)
	if _, err := u.AddReq("Gamma", []string{"x"}); err == nil || !strings.Contains(err.Error(), "Gamma") {
		t.Errorf("an unknown heading was not refused by name: %v", err)
	}
	if _, err := u.AddReq("Alpha", []string{"**R9:** labelled"}); err == nil || !strings.Contains(err.Error(), "R9") {
		t.Errorf("a body writing its own label was not refused naming it: %v", err)
	}
	if read(t, rp) != before {
		t.Error("a refusal wrote something")
	}
}

// R326 — the gap verbs through the reader: add mints and places, resolve checks, approve
// converts; a permanent or already-resolved gap is refused with nothing written.
func TestGapVerbsWriteThroughTheReader(t *testing.T) {
	u, _, dp := reqProject(t)
	id, err := u.AddGap("O", "a new one")
	if err != nil || id != "O3" {
		t.Fatalf("AddGap: %s %v; want O3 — O2 counts though resolved", id, err)
	}
	if got := read(t, dp); !strings.Contains(got, "- T1: R2 retired (reason)\n- [ ] O3: a new one\n\n## Other") {
		t.Errorf("the gap was not appended after the last entry:\n%s", got)
	}
	if err := u.ResolveGap("O1"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(read(t, dp), "- [x] O1: open one") {
		t.Error("O1 was not checked")
	}
	if err := u.ResolveGap("O2"); err == nil {
		t.Error("resolving a resolved gap was absorbed")
	}
	if err := u.ResolveGap("A1"); err == nil {
		t.Error("resolving a permanent gap was accepted")
	}
	newID, err := u.ApproveGap("O3")
	if err != nil || newID != "A2" {
		t.Fatalf("ApproveGap: %s %v; want A2", newID, err)
	}
	if !strings.Contains(read(t, dp), "- A2: a new one") {
		t.Error("O3 was not rewritten as A2")
	}
	if again, err := u.ApproveGap("A2"); err != nil || again != "A2" {
		t.Errorf("approving an approved gap: %s %v; want its own ID and no write", again, err)
	}
}

// R80, R326 — retire rewrites the head line in one document and adds the Tn in the other.
func TestRetireWritesBothDocumentsThroughTheReaders(t *testing.T) {
	u, rp, dp := reqProject(t)
	tn, sources, err := u.Retire("R3", "R1", "folded")
	if err != nil || tn != "T2" || strings.Join(sources, ",") != "specs/beta.md" {
		t.Fatalf("Retire: %s %v %v", tn, sources, err)
	}
	if !strings.Contains(read(t, rp), "- **~~R3:~~** (Retired T2 — see R1) third") {
		t.Error("the head line was not rewritten")
	}
	if !strings.Contains(read(t, dp), "- T2: R3 retired by R1 (folded)") {
		t.Error("the Tn gap was not added")
	}
	if _, _, err := u.Retire("R3", "-", "again"); err == nil {
		t.Error("a second retirement was absorbed")
	}
}
