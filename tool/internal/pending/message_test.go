// CRC: crc-CommitMessage.md | R479, R480, R481, R482, R483
package pending

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.email=t@example.com", "-c", "user.name=t"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func finishedFixture(t *testing.T) string {
	t.Helper()
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, FinishOpts{Body: "  what a reader needs to know.\n"}); err != nil {
		t.Fatal(err)
	}
	return root
}

// R479, R480, R481 — the uncommitted items are the ones no message names, bounded.
func TestTheUncommittedItemsAreTheOnesNoMessageNames(t *testing.T) {
	root := finishedFixture(t)
	msg, err := Compose(root, false)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Subject != "#4: a part to queue" || !strings.HasPrefix(msg.Text, "#4: a part to queue\n\nItems #4.\n") ||
		!strings.Contains(msg.Text, "#4 — a part to queue\n") || !strings.Contains(msg.Text, "what a reader needs to know.") {
		t.Errorf("composed:\n%s", msg.Text)
	}
	gitIn(t, root, "commit", "-q", "--allow-empty", "-m", "unrelated: names #40 only")
	if _, err := Compose(root, false); err != nil {
		t.Errorf("#40 was read as naming #4: %v", err)
	}
	// an old-scheme entry, hash in its slot, is committed by definition
	done := read(t, root, "DONE.md") + "\n- **2026-08-01 — #9: an older completion.** (`abc1234`) Part `carves/x.md#2`.\n"
	if err := os.WriteFile(filepath.Join(root, "DONE.md"), []byte(done), 0o644); err != nil {
		t.Fatal(err)
	}
	if msg, err := Compose(root, false); err != nil || strings.Contains(msg.Text, "#9") {
		t.Errorf("an entry with a hash was composed: %v\n%s", err, msg.Text)
	}
	gitIn(t, root, "commit", "-q", "--allow-empty", "-m", "lands #4")
	_, err = Compose(root, false)
	if !errors.Is(err, ErrNothingToCommit) || !strings.Contains(err.Error(), "#4") {
		t.Errorf("want the nothing-to-compose refusal naming #4, got %v", err)
	}
	// an older entry with neither a hash nor a naming commit, below the named #4: history
	done = read(t, root, "DONE.md") + "\n- **2026-08-02 — #3: older, unnamed, no hash.** Part `carves/x.md#1`.\n"
	if err := os.WriteFile(filepath.Join(root, "DONE.md"), []byte(done), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Compose(root, false); !errors.Is(err, ErrNothingToCommit) {
		t.Errorf("an entry older than a named one was composed: %v", err)
	}
}

// R482, R483 — amend appends after the existing message, and refuses on a remote.
func TestAmendAppendsAndRefusesOnARemote(t *testing.T) {
	root := finishedFixture(t)
	gitIn(t, root, "commit", "-q", "--allow-empty", "-m", "the existing subject\n\nthe existing body, byte for byte.")
	msg, err := Compose(root, true)
	if err != nil {
		t.Fatal(err)
	}
	want := "the existing subject\n\nthe existing body, byte for byte.\n\nAlso lands #4.\n\n#4 — a part to queue\n"
	if !strings.HasPrefix(msg.Text, want) || msg.Subject != "the existing subject" {
		t.Errorf("amended text:\n%s", msg.Text)
	}
	bare := t.TempDir()
	gitIn(t, bare, "init", "-q", "--bare")
	gitIn(t, root, "remote", "add", "o", bare)
	gitIn(t, root, "push", "-q", "o", "HEAD:refs/heads/main")
	var onRemote *OnRemoteError
	if _, err := Compose(root, true); !errors.As(err, &onRemote) || !strings.Contains(err.Error(), "o/main") {
		t.Errorf("want the on-remote refusal naming o/main, got %v", err)
	}
}
