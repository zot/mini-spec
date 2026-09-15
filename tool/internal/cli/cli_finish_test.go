// CRC: crc-CLI.md | R469, R475
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/update"
)

// R469, R475 — the verb needs no design root, reports the move and its rewrites, and
// refuses a second run because the carve is no longer in carves/.
func TestFinishedCarveNeedsNoDesignRootAndReports(t *testing.T) {
	dir := t.TempDir()
	status := "# Carve: x\n\n## Status\n\n- [x] ~~**Item 1 — done.**~~ **LANDED (`abc`, 2026-09-01 — `#5`.)**\n\n"
	for name, body := range map[string]string{
		"tool/a.md":       "a\n",
		"carves/x.md":     status + "[a](../tool/a.md) [n](nowhere.md)\n",
		"carves/other.md": status + "[x](x.md)\n",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	prev, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(dir, "tool")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	c := &CLI{}
	if code := c.runUpdate([]string{"finished-carve", "../carves/x.md"}); code != 0 {
		t.Errorf("first run exited %d; want 0 with no design root", code)
	}
	if code := c.runUpdate([]string{"finished-carve", "../carves/x.md"}); code != 1 {
		t.Errorf("second run exited %d; the carve is no longer in carves/", code)
	}
	r := &update.FinishReport{
		From: "carves/x.md",
		To:   "carves/done/x.md",
		Considered: []update.Considered{
			{File: "carves/x.md", Line: 7, Old: "../tool/a.md", New: "../../tool/a.md", Outcome: update.Rewritten},
			{File: "carves/x.md", Line: 7, Old: "nowhere.md", Outcome: update.Left},
		},
		Counts:  map[update.Outcome]int{update.Rewritten: 1, update.Left: 1},
		Written: []string{"carves/done/x.md"},
	}
	var out strings.Builder
	printFinish(&out, r)
	wants := []string{
		"moved carves/x.md → carves/done/x.md",
		"carves/x.md:7  ../tool/a.md → ../../tool/a.md  rewritten",
		"carves/x.md:7  nowhere.md  left: does not resolve",
		"rewritten 1, left 1; 1 files written",
	}
	for _, want := range wants {
		if !strings.Contains(out.String(), want) {
			t.Errorf("report lacks %q:\n%s", want, out.String())
		}
	}
	moved, _ := os.ReadFile(filepath.Join(dir, "carves", "done", "x.md"))
	if !strings.HasSuffix(string(moved), "[a](../../tool/a.md) [n](nowhere.md)\n") {
		t.Errorf("moved carve:\n%s", moved)
	}
}
