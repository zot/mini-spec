// CRC: crc-CLI.md | R489, R467, R468
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/update"
)

// R489, R467, R468 — the default population is every tracked document, the report names
// every link with its outcome, and the exit status follows what was left.
func TestRepairLinksPopulationReportAndExitStatus(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"tool/x.md":            "x\n",
		"carves/done/moved.md": "[x](../tool/x.md) [n](nowhere.md)\n",
		"carves/live.md":       "[m](moved.md)\n",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, dir, ".")
	prev, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(dir, "tool")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	c := &CLI{}
	if code := c.runUpdate([]string{"repair-links"}); code != 1 {
		t.Errorf("first run exited %d; want 1 with a link left", code)
	}
	if code := c.runUpdate([]string{"repair-links"}); code != 1 {
		t.Errorf("second run exited %d; the unresolvable link is still left", code)
	}
	r, err := update.RepairLinks(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Files) != 3 { // every tracked markdown file, tool/x.md among them
		t.Errorf("population: %v", r.Files)
	}
	var out strings.Builder
	printRepair(&out, r)
	if !strings.Contains(out.String(), "carves/done/moved.md:1  nowhere.md  left: unresolvable") ||
		!strings.HasSuffix(strings.TrimSpace(out.String()), "1 links considered in 3 files: rewritten 0, unresolvable 1, ambiguous 0; 0 files written") {
		t.Errorf("report:\n%s", out.String())
	}
	moved, _ := os.ReadFile(filepath.Join(dir, "carves", "done", "moved.md"))
	live, _ := os.ReadFile(filepath.Join(dir, "carves", "live.md"))
	if string(moved) != "[x](../../tool/x.md) [n](nowhere.md)\n" || string(live) != "[m](done/moved.md)\n" {
		t.Errorf("files after the first run:\n%s%s", moved, live)
	}
}
