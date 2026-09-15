// CRC: crc-CLI.md | R488, R458, R461
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/query"
)

// R488, R458, R461 — the default population is every tracked document, an untracked one is
// never read, and the exit status follows the errors.
func TestQueryLinksPopulationAndExitStatus(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"target.md":        "x\n",
		"carves/a.md":      "see [t](../target.md)\n",
		"carves/done/b.md": "see [m](../../missing.md)\n",
		"notes.md":         "private [n](nowhere.md)\n",
		"x/testdata/s.md":  "fixture [f](nowhere.md)\n",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, dir, "target.md", "carves", "x")
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	// a tracked file deleted on disk, the deletion not yet staged, is no longer a document
	os.WriteFile(filepath.Join(dir, "gone.md"), []byte("[g](nowhere.md)\n"), 0o644)
	gitInit(t, dir, "gone.md")
	os.Remove(filepath.Join(dir, "gone.md"))

	c := &CLI{}
	if code := c.runQuery([]string{"links"}); code != 1 {
		t.Errorf("default population exited %d; the tracked done carve has a missing link", code)
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
	if strings.Contains(all.String(), "notes.md") || strings.Contains(all.String(), "testdata") || strings.Contains(all.String(), "gone.md") {
		t.Errorf("an untracked document or a fixture was read:\n%s", all.String())
	}
	want := "2 links in 3 files: tracked 1, untracked 0, ignored 0, missing 1, outside 0, external 0, local 0"
	if !strings.HasSuffix(strings.TrimSpace(plain.String()), want) {
		t.Errorf("summary: want %q in:\n%s", want, plain.String())
	}
}

// gitInit makes dir a repository and stages the named pathspecs: the link population is what
// the index holds, so a fixture without one has no public documents. Shared by the repair and
// finished-carve tests.
func gitInit(t *testing.T, dir string, add ...string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	for _, args := range [][]string{{"init", "-q"}, append([]string{"add", "--"}, add...)} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}
