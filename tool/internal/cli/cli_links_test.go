// CRC: crc-CLI.md | R455, R458, R461
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/query"
)

// R455, R458, R461 — the default population is the live carves; a done carve's missing link
// counts only when it is named; the exit status follows the errors.
func TestQueryLinksPopulationAndExitStatus(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	dir := t.TempDir()
	for name, body := range map[string]string{
		"target.md":        "x\n",
		"carves/a.md":      "see [t](../target.md)\n",
		"carves/done/b.md": "see [m](../../missing.md)\n",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	c := &CLI{}
	if code := c.runQuery([]string{"links"}); code != 0 {
		t.Errorf("default population exited %d; the done carve must not be read", code)
	}
	if code := c.runQuery([]string{"links", "carves/done/b.md"}); code != 1 {
		t.Errorf("naming the done carve exited %d; want 1 on its missing link", code)
	}

	r, err := query.CheckLinks(dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	var plain, all strings.Builder
	printLinks(&plain, r, false)
	printLinks(&all, r, true)
	if strings.Contains(plain.String(), "[t](../target.md)") {
		t.Errorf("a tracked link listed without --all:\n%s", plain.String())
	}
	if !strings.Contains(all.String(), "carves/a.md:1  [t](../target.md)  tracked") {
		t.Errorf("--all does not list the tracked link:\n%s", all.String())
	}
	done, err := query.CheckLinks(dir, []string{"carves/done/b.md"})
	if err != nil {
		t.Fatal(err)
	}
	var named strings.Builder
	printLinks(&named, done, false)
	if !strings.Contains(named.String(), "carves/done/b.md:1  [m](../../missing.md)  missing  error") {
		t.Errorf("a missing link is not listed with its class and the error mark:\n%s", named.String())
	}
	want := "1 links in 1 files: tracked 1, untracked 0, ignored 0, missing 0, outside 0, external 0, local 0"
	if !strings.HasSuffix(strings.TrimSpace(plain.String()), want) {
		t.Errorf("summary: want %q in:\n%s", want, plain.String())
	}
}
