// CRC: crc-Backup.md | R224, R227, R228, R229, R230, R231, R232, R233, R234, R236, R239
package backup

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// write replaces a file with an mtime strictly later than anything on disk, since the drift
// check is an mtime comparison and a filesystem's resolution is coarser than a test's runtime.
func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(path, later, later); err != nil {
		t.Fatal(err)
	}
}

func recorded(t *testing.T, root, body string) *Slot {
	t.Helper()
	s := New(root)
	if err := s.Record(func() error {
		return os.WriteFile(filepath.Join(root, "PENDING.md"), []byte(body), 0o644)
	}); err != nil {
		t.Fatal(err)
	}
	return s
}

func gitTree(t *testing.T, files map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH: the anchor is a git object and cannot be faked")
	}
	root := tree(t, files)
	for _, args := range [][]string{{"init"}, {"add", "-A"}, {"commit", "-m", "seed"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	return root
}

// R236 — the wiring, which fails silently: every anchor test keeps passing if the call in
// swap is deleted. The file is deliberately not a trajectory file, and the anchor must hold
// its pre-mutation content.
func TestTheSwapAnchorsTheWorktreeBeforeTheTransition(t *testing.T) {
	root := gitTree(t, map[string]string{"tracked.md": "before the mutation\n"})
	s := New(root)
	if err := s.Record(func() error {
		return os.WriteFile(filepath.Join(root, "tracked.md"), []byte("after the mutation\n"), 0o644)
	}); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "show", project.SnapshotRef+":tracked.md")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("no anchor was written by the swap: %v", err)
	}
	if string(out) != "before the mutation\n" {
		t.Errorf("the anchor holds %q, so it was taken after the transition rather than before", out)
	}
}

// R239 — no git is not a failure of the slot.
func TestATreeWithNoGitStillRestores(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\noriginal\n"})
	s := recorded(t, root, "# Pending\n\nafter\n")
	if err := s.Revert(); err != nil {
		t.Fatalf("revert without git: %v", err)
	}
	if got := read(t, filepath.Join(root, "PENDING.md")); got != "# Pending\n\noriginal\n" {
		t.Errorf("not restored: %q", got)
	}
}

// R227 — drift refuses, and leaves the live files alone.
func TestDriftRefusesWithoutClobbering(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\noriginal\n"})
	s := recorded(t, root, "# Pending\n\nafter the mutation\n")
	edited := "# Pending\n\nhand-edited since the backup\n"
	write(t, filepath.Join(root, "PENDING.md"), edited)

	err := s.Revert()
	var drift *DriftError
	if !errors.As(err, &drift) {
		t.Fatalf("err = %v, want a DriftError", err)
	}
	if got := read(t, filepath.Join(root, "PENDING.md")); got != edited {
		t.Errorf("the hand edit was clobbered:\n%s", got)
	}
	if msg := drift.Error(); !strings.Contains(msg, "PENDING.md") || !strings.Contains(msg, s.Dir()) {
		t.Errorf("refusal does not name the file and its backup location: %s", msg)
	}
}

// R224 — a failure mid-operation leaves the previous backup intact.
func TestAFailedOperationLeavesThePreviousBackupIntact(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\nfirst\n"})
	s := recorded(t, root, "# Pending\n\nsecond\n")
	before := read(t, filepath.Join(s.Dir(), "PENDING.md"))
	boom := errors.New("the operation failed")
	if err := s.Record(func() error { return boom }); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the operation's own error", err)
	}
	if after := read(t, filepath.Join(s.Dir(), "PENDING.md")); after != before {
		t.Errorf("a failed operation replaced the backup:\ngot  %q\nwant %q", after, before)
	}
}

// R228 — from every state exactly one of revert / replay is legal.
func TestExactlyOneOperationIsLegalFromEachState(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\nfirst\n"})
	s := recorded(t, root, "# Pending\n\nsecond\n")
	if err := s.Replay(); !isIllegal(err) {
		t.Errorf("replay from changed: err = %v, want IllegalError", err)
	}
	if err := s.Revert(); err != nil {
		t.Fatalf("revert from changed: %v", err)
	}
	var ill *IllegalError
	if err := s.Revert(); !errors.As(err, &ill) {
		t.Fatalf("revert from reverted: err = %v, want IllegalError", err)
	}
	if msg := ill.Error(); !strings.Contains(msg, "reverted") || !strings.Contains(msg, "replay") {
		t.Errorf("refusal names neither the state nor what it accepts: %s", msg)
	}
	if err := s.Replay(); err != nil {
		t.Fatalf("replay from reverted: %v", err)
	}
	if err := s.Replay(); !isIllegal(err) {
		t.Errorf("replay from replayed: err = %v, want IllegalError", err)
	}
}

// R228 — changed and replayed are distinct states.
func TestAFreshMutationAlwaysStampsChanged(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\nfirst\n"})
	s := recorded(t, root, "# Pending\n\nsecond\n")
	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	if err := s.Replay(); err != nil {
		t.Fatal(err)
	}
	if st, _, _ := s.State(); st != Replayed {
		t.Fatalf("setup: state = %v, want replayed", st)
	}
	write(t, filepath.Join(root, "PENDING.md"), "# Pending\n\nthird\n")
	if err := s.Record(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if st, _, _ := s.State(); st != Changed {
		t.Errorf("state after a fresh mutation from replayed = %v, want changed", st)
	}
}

// R229 — the goldfish rule.
func TestANewMutationDiscardsWhatWasRevertable(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\nfirst\n"})
	s := recorded(t, root, "# Pending\n\nsecond\n")
	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "PENDING.md"), "# Pending\n\nthird\n")
	if err := s.Record(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if err := s.Replay(); !isIllegal(err) {
		t.Errorf("replay after a new mutation: err = %v, want IllegalError", err)
	}
	if got, want := read(t, filepath.Join(s.Dir(), "PENDING.md")), "# Pending\n\nthird\n"; got != want {
		t.Errorf("backup = %q, want the state before the second mutation (%q)", got, want)
	}
}

// R227 — the drift check must not fire on the tool's own restore.
func TestARestoredFileIsNotReadAsAHandEdit(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n\nfirst\n"})
	s := recorded(t, root, "# Pending\n\nsecond\n")
	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	if err := s.Replay(); err != nil {
		t.Errorf("replay read the restore as a hand edit: %v", err)
	}
}

func TestAnEmptySlotHasNothingToUndo(t *testing.T) {
	s := New(tree(t, map[string]string{"PENDING.md": "# Pending\n"}))
	if err := s.Revert(); !errors.Is(err, ErrNoSlot) {
		t.Errorf("err = %v, want ErrNoSlot", err)
	}
}

func isIllegal(err error) bool {
	var ill *IllegalError
	return errors.As(err, &ill)
}

const openPart = "# Carve: x\n\n## Status\n\n- [ ] **Item 5 — a part.** **OPEN (not queued.)**\n"

func carveWithOpenPart(t *testing.T, root string) (path, body string) {
	t.Helper()
	path = filepath.Join(root, "carves", "x.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(openPart), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, openPart
}

const entry16 = "# Pending\n\n## 16. **the backup slot**. Active.\n" +
	"   Source: [carves/x.md](carves/x.md), part `#Item 5`.\n"

// queuePart records the mutation a vend performs: the entry pointing at the part, and the
// part's marker carrying the number — both in one Record, the pair a revert rolls back in
// opposite directions.
func queuePart(t *testing.T, s *Slot, root, carve string) {
	t.Helper()
	if err := s.Record(func() error {
		if err := os.WriteFile(filepath.Join(root, "PENDING.md"), []byte(entry16), 0o644); err != nil {
			return err
		}
		return parser.SetMarker(carve, "Item 5", "OPEN", "#16.")
	}); err != nil {
		t.Fatal(err)
	}
}

// R222, R230 — the queue rolls backward and the carve moves forward.
func TestACarveIsWrittenNeverRestored(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n"})
	carve, before := carveWithOpenPart(t, root)
	s := New(root)
	queuePart(t, s, root, carve)

	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	if got := read(t, filepath.Join(root, "PENDING.md")); got != "# Pending\n" {
		t.Errorf("the pending file was not restored:\n%s", got)
	}
	got := read(t, carve)
	if !strings.Contains(got, "**REVERTED (#16.)**") {
		t.Errorf("the carve does not carry the revert trace:\n%s", got)
	}
	if got == before {
		t.Error("the carve was restored, which un-vends the number it had been given")
	}
	if err := s.Replay(); err != nil {
		t.Fatal(err)
	}
	if got := read(t, carve); !strings.Contains(got, "**OPEN (#16.)**") {
		t.Errorf("replay did not reopen the part:\n%s", got)
	}
}

// R231, R234 — an ended attempt reopens the part, frees the number, writes no ledger entry.
func TestAnEndedAttemptReopensThePartAndFreesTheNumber(t *testing.T) {
	root := tree(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #1: something.** (`aaa`)\n",
	})
	carve, _ := carveWithOpenPart(t, root)
	ledgerBefore := read(t, filepath.Join(root, "DONE.md"))
	s := New(root)
	queuePart(t, s, root, carve)
	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	freed, err := s.Released()
	if err != nil {
		t.Fatal(err)
	}
	if len(freed) != 1 || freed[0] != 16 {
		t.Fatalf("Released() = %v, want [16]", freed)
	}
	write(t, filepath.Join(root, "PENDING.md"), "# Pending\n\nsomething else entirely\n")
	if err := s.Record(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if got := read(t, carve); !strings.Contains(got, "**OPEN (not queued.)**") {
		t.Errorf("the part was not reopened — aborting an attempt is not aborting the part:\n%s", got)
	}
	if got := read(t, filepath.Join(root, "DONE.md")); got != ledgerBefore {
		t.Errorf("an unfinished attempt was written to the completion ledger:\n%s", got)
	}
	if freed, _ := s.Released(); len(freed) != 0 {
		t.Errorf("Released() = %v after the attempt ended, want none", freed)
	}
}

// R231 — a live item is not released.
func TestALiveItemIsNotReleased(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n"})
	s := recorded(t, root, "# Pending\n\n## 16. **live**. Active.\n")
	if freed, _ := s.Released(); len(freed) != 0 {
		t.Errorf("Released() from changed = %v, want none", freed)
	}
	if err := s.Revert(); err != nil {
		t.Fatal(err)
	}
	if err := s.Replay(); err != nil {
		t.Fatal(err)
	}
	if freed, _ := s.Released(); len(freed) != 0 {
		t.Errorf("Released() from replayed = %v, want none — the entry is live again", freed)
	}
}

// R232, R233 — a completion produces the same entry diff as a revert, so the release must
// not act on the diff alone: the slot is `changed`, the backup holds the entry, the live
// file does not, and the part is landed.
func TestACompletedPartIsNotReopenedByTheNextMutation(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n"})
	carve, _ := carveWithOpenPart(t, root)
	s := New(root)
	queuePart(t, s, root, carve)
	// The completion: entry leaves the live file, the part lands. One Record, like finish.
	if err := s.Record(func() error {
		if err := os.WriteFile(filepath.Join(root, "PENDING.md"), []byte("# Pending\n"), 0o644); err != nil {
			return err
		}
		return parser.SetPartLanded(carve, "Item 5", "`abc`, 2026-09-04 — `#16`.")
	}); err != nil {
		t.Fatal(err)
	}
	landedText := read(t, carve)
	write(t, filepath.Join(root, "PENDING.md"), "# Pending\n\n## 17. **next**. Active.\n")
	if err := s.Record(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	got := read(t, carve)
	if got != landedText || strings.Contains(got, "not queued") || !strings.Contains(got, "LANDED") {
		t.Errorf("the completed part was reopened:\nbefore:\n%s\nafter:\n%s", landedText, got)
	}
}

// R232 alone — a hand-removed entry is not an abandoned attempt.
func TestAHandRemovedEntryIsNotAnAbandonedAttempt(t *testing.T) {
	root := tree(t, map[string]string{"PENDING.md": "# Pending\n"})
	carve, _ := carveWithOpenPart(t, root)
	s := New(root)
	queuePart(t, s, root, carve)
	// A second mutation, so the backup holds the entry.
	if err := s.Record(func() error {
		return os.WriteFile(filepath.Join(root, "PENDING.md"), []byte(entry16+"\n## 17. **another**. Active.\n"), 0o644)
	}); err != nil {
		t.Fatal(err)
	}
	// The hand edit removes #16 while its part stays open.
	write(t, filepath.Join(root, "PENDING.md"), "# Pending\n\n## 17. **another**. Active.\n")
	if err := s.Record(func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if got := read(t, carve); !strings.Contains(got, "**OPEN (#16.)**") {
		t.Errorf("a diff the tool did not cause was read as an abandoned attempt:\n%s", got)
	}
}
