// CRC: crc-Parser.md | R73, R74, R75, R77
package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempNamed(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseRequirements_Retired(t *testing.T) {
	path := writeTempNamed(t, "requirements.md", `# Requirements

## Feature: Test
**Source:** specs/test.md

- **R1:** active requirement
- **~~R2:~~** (Retired T1 — see R5) original retired text
- **~~R3:~~** (Retired T2 — no replacement) gone for good
- **R4:** another active
- **R5:** replacement requirement
`)
	reqs, err := ParseRequirements(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reqs) != 5 {
		t.Fatalf("got %d reqs, want 5", len(reqs))
	}
	if reqs[1].ID != "R2" || !reqs[1].Retired {
		t.Errorf("R2 not retired: %+v", reqs[1])
	}
	if reqs[1].Text != "original retired text" {
		t.Errorf("R2 text: got %q", reqs[1].Text)
	}
	if !reqs[2].Retired || reqs[2].Text != "gone for good" {
		t.Errorf("R3 retired/text wrong: %+v", reqs[2])
	}
	if reqs[0].Retired || reqs[3].Retired {
		t.Errorf("non-retired marked retired")
	}
}

func TestParseGaps_TypeT_NoCheckbox(t *testing.T) {
	path := writeTempNamed(t, "design.md", `# Design

## Gaps

- [ ] D1: missing thing
- A1: approved without checkbox
- [ ] A2: legacy approved with checkbox
- T1: R5 retired by R10 (rewrite)
- T2: R7 retired (no replacement)
`)
	gaps, err := ParseGaps(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(gaps) != 5 {
		t.Fatalf("got %d gaps, want 5", len(gaps))
	}

	want := []struct {
		id          string
		typ         string
		hasCheckbox bool
	}{
		{"D1", "D", true},
		{"A1", "A", false},
		{"A2", "A", true},
		{"T1", "T", false},
		{"T2", "T", false},
	}
	for i, w := range want {
		got := gaps[i]
		if got.ID != w.id || got.Type != w.typ || got.HasCheckbox != w.hasCheckbox {
			t.Errorf("gap[%d] = {%s %s checkbox=%v}, want {%s %s checkbox=%v}",
				i, got.ID, got.Type, got.HasCheckbox, w.id, w.typ, w.hasCheckbox)
		}
	}
}
