// CRC: crc-CLI.md | R331
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mintProject builds a design root with two requirements and an empty Gaps section, and
// chdirs into it so getProject finds it.
func mintProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("specs/a.md", "# A\n")
	write("design/requirements.md", "# Requirements\n\n## Feature: A\n**Source:** specs/a.md\n\n- **R1:** one\n- **R2:** two\n")
	write("design/design.md", "# Design\n\n## Artifacts\n\n## Gaps\n\n- [ ] O1: open\n")
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })
	return root
}

// R331 — a minted value is reported as a sentence, hidden by --quiet, carried by --json.
func TestMintedValuesAreReportedNotReturned(t *testing.T) {
	mintProject(t)
	out := captureStdout(t, func() {
		if code := (&CLI{}).runUpdate([]string{"retire", "R1", "R2", "folded"}); code != 0 {
			t.Fatalf("retire exited %d", code)
		}
	})
	if strings.TrimSpace(out) != "Retired R1 as T1 (see R2)" {
		t.Errorf("retire stdout = %q, want the sentence", out)
	}
	quiet := captureStdout(t, func() {
		if code := (&CLI{Quiet: true}).runUpdate([]string{"add-req", "--section", "A", "--req", "three"}); code != 0 {
			t.Fatalf("add-req exited %d", code)
		}
	})
	if quiet != "" {
		t.Errorf("--quiet still printed %q", quiet)
	}
	js := captureStdout(t, func() {
		if code := (&CLI{JSON: true}).runUpdate([]string{"add-req", "--section", "A", "--req", "four", "--req", "five"}); code != 0 {
			t.Fatalf("add-req --json exited %d", code)
		}
	})
	if !strings.Contains(js, `"range": "R4-R5"`) && !strings.Contains(js, `"range":"R4-R5"`) {
		t.Errorf("--json does not carry the minted range:\n%s", js)
	}
	plain := captureStdout(t, func() {
		if code := (&CLI{}).runUpdate([]string{"add-req", "--section", "A", "--req", "six"}); code != 0 {
			t.Fatalf("add-req exited %d", code)
		}
	})
	if strings.TrimSpace(plain) != `Added R6 to "A"` {
		t.Errorf("add-req stdout = %q, want the sentence", plain)
	}
}
