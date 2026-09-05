// CRC: crc-CLI.md | R207, R212, R213, R216, R217, R218
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/parser"
)

const censusFixture = "# Carve: x\n\n## Status\n\n" +
	"- [ ] **Item 1 — open and conforming.** **OPEN (#7.)**\n" +
	"- [x] ~~**Item 2 — landed.**~~ **LANDED (`abc`, 2026-09-01 — `#5`.)**\n" +
	"- [ ] **#4 — keyed on a superseded scheme.** **OPEN (not queued.)**\n" +
	"- **Item 3 — a split parent.** **SPLIT (Bill, 2026-09-01.)**\n"

func scanFixture(t *testing.T) *parser.CarveScan {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "carves"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "carves", "x.md"), []byte(censusFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	scan, err := parser.ScanCarves(root)
	if err != nil {
		t.Fatal(err)
	}
	return &scan
}

// R212, R216, R217, R218 — a non-conforming part lists without --open; a clean open part
// and a clean stateless line only with it; the census states zeros and says stateless.
func TestTheRenderedLineAndTheListingRule(t *testing.T) {
	scan := scanFixture(t)
	var plain, open strings.Builder
	printCarves(&plain, scan, false)
	printCarves(&open, scan, true)

	if !strings.Contains(plain.String(), "(unkeyed)") || strings.Contains(plain.String(), "Item 1 ") || strings.Contains(plain.String(), "(stateless)") {
		t.Errorf("without --open, want only the non-conforming row:\n%s", plain.String())
	}
	if !strings.Contains(open.String(), "Item 1    #7") || !strings.Contains(open.String(), "(stateless) L8") {
		t.Errorf("with --open, want the open part and the stateless row:\n%s", open.String())
	}
	want := "1 carve: 2 open, 1 landed, 1 stateless, 1 non-conforming; 0 documents with no status block"
	if !strings.HasSuffix(strings.TrimSpace(plain.String()), want) {
		t.Errorf("census: want %q at the end of:\n%s", want, plain.String())
	}
	if !strings.Contains(plain.String(), "carves/x.md    2 open    1 landed  1 non-conforming  1 stateless") {
		t.Errorf("carve line shape:\n%s", plain.String())
	}
}

// R213 — the command answers with no design root anywhere.
func TestCarvesNeedsNoDesignRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "carves"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "carves", "x.md"), []byte(censusFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".minispec.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if code := (&CLI{}).runQuery([]string{"carves"}); code != 0 {
			t.Errorf("runQuery exited %d in a tree with no design root; want 0", code)
		}
	})
	if !strings.Contains(out, "1 carve:") {
		t.Errorf("no census in:\n%s", out)
	}
}
