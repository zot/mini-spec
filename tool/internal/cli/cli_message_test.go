// CRC: crc-CLI.md | R484
package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/backup"
)

// R484 — the verb writes the file, never commits, and never enters the backup slot.
func TestCommitMessageWritesTheFileAndNeverCommits(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"PENDING.md":            "# Pending\n\n---\n",
		"CURRENT.md":            "# Current\n\n---\n\n## Active\n\n_No active item._\n",
		"DONE.md":               "# Done\n\n---\n\n- **2026-09-15 — #4: a part to queue.** Part `carves/x.md#7`.\n  the body.\n",
		".minispec/config.yaml": "track: private-trajectory\n",
		".gitignore":            "PENDING.md\nCURRENT.md\nDONE.md\n.minispec/backup\n",
		"carves/x.md":           "# c\n",
		"tool/.keep":            "",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-qm", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v: %s", err, out)
		}
	}
	prev, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(dir, "tool")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	before, heldBefore, _ := backup.New(dir).State()
	out := filepath.Join(dir, "msg.txt")
	c := &CLI{}
	if code := c.runPending([]string{"commit-message", "--out", out}); code != 0 {
		t.Fatalf("exited %d", code)
	}
	b, err := os.ReadFile(out)
	if err != nil || !strings.HasPrefix(string(b), "#4: a part to queue\n\nItems #4.\n") {
		t.Errorf("the file: %v\n%s", err, b)
	}
	log := exec.Command("git", "log", "--oneline")
	log.Dir = dir
	if l, _ := log.Output(); strings.Count(string(l), "\n") != 1 {
		t.Errorf("the verb committed:\n%s", l)
	}
	// The stamp's presence is the fact: the recorded state is the zero value, so comparing
	// states alone reads a fresh record as unchanged.
	if after, heldAfter, _ := backup.New(dir).State(); heldAfter != heldBefore || after != before {
		t.Errorf("the verb entered the backup slot: %v/%v → %v/%v", before, heldBefore, after, heldAfter)
	}
}
