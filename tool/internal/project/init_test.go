// CRC: crc-Init.md | Seq: seq-bootstrap.md#2 | R136, R137, R138, R139, R140, R141, R142, R144, R145, R161, R170, R171, R172
package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case runs against a temporary directory and the fake Git from track_test.go,
// so no repository is created and no `git` is invoked. What init writes is a function
// of two inputs — the chosen value and whether the tree is git-managed — which makes
// the whole matrix cheap to state. See test-Init.md.

func initIn(t *testing.T, dir string, track TrackValue, repair bool, g GitFacts) (*InitResult, error) {
	t.Helper()
	return RunInit(InitOptions{RepoRoot: dir, Track: track, Repair: repair}, g)
}

func readGitignore(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, GitIgnoreName))
	if err != nil {
		return ""
	}
	return string(data)
}

// R136 — the value is never inferred or defaulted; it is the one decision the tool
// must not make for the user.
func TestInitRequiresATrackValue(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, "", false, gitWith()); err == nil {
		t.Fatal("init with no track value succeeded, want refusal")
	}
	if _, err := os.Stat(filepath.Join(dir, ConfigDirName)); err == nil {
		t.Error("refused init still created .minispec/")
	}
}

// R137, R138 — the happy path, and that the value round-trips.
func TestInitWritesTheChosenValue(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	v, err := LoadTrack(RepoConfigPath(dir))
	if err != nil || v != TrackPrivateTrajectory {
		t.Fatalf("LoadTrack = (%q, %v), want (private-trajectory, nil)", v, err)
	}
}

// R139, R140 — the value that ignores the most.
func TestPrivateTrajectoryWritesBothKindsOfIgnoreLine(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	got := readGitignore(t, dir)
	want := append([]string{BackupIgnorePath}, TrajectoryFiles...)
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf(".gitignore missing %q:\n%s", w, got)
		}
	}
}

// R139 — the discriminating case: `all` still hides the tool's machine-local files.
func TestAllWritesOnlyTheBackupIgnoreLine(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackAll, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	got := readGitignore(t, dir)
	if !strings.Contains(got, BackupIgnorePath) {
		t.Errorf(".gitignore missing %q:\n%s", BackupIgnorePath, got)
	}
	for _, f := range TrajectoryFiles {
		if strings.Contains(got, f) {
			t.Errorf(".gitignore should not mention %q under track: all:\n%s", f, got)
		}
	}
}

// R139, R151 — a project with no repository is left alone rather than told about a
// file it does not have.
func TestNoneWritesNoIgnoreLines(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackNone, false, &fakeGit{repo: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(RepoConfigPath(dir)); err != nil {
		t.Errorf("config was not written: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, GitIgnoreName)); err == nil {
		t.Error(".gitignore was created in a tree with no repository")
	}
}

// R139, R140 — the tool edits a file it does not own, so clobbering it is the failure
// that would cost a user the most.
func TestExistingGitignoreIsAppendedToNotReplaced(t *testing.T) {
	dir := t.TempDir()
	original := "node_modules/\n*.log\n"
	if err := os.WriteFile(filepath.Join(dir, GitIgnoreName), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	got := readGitignore(t, dir)
	for _, line := range []string{"node_modules/", "*.log"} {
		if !strings.Contains(got, line) {
			t.Errorf("original line %q was lost:\n%s", line, got)
		}
	}
	if !strings.Contains(got, BackupIgnorePath) {
		t.Errorf("new line was not added:\n%s", got)
	}
}

// R139, R141 — an already-correct file is left alone, and the report says so rather
// than claiming an edit.
func TestIgnoreLinesAreNotDuplicated(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	first := readGitignore(t, dir)

	// Repair to the same value: everything is already correct.
	result, err := initIn(t, dir, TrackPrivateTrajectory, true, gitWith())
	if err != nil {
		t.Fatal(err)
	}
	if second := readGitignore(t, dir); second != first {
		t.Errorf(".gitignore changed on a no-op repair:\n%q\n%q", first, second)
	}
	if len(result.Edited) != 0 || len(result.Created) != 0 {
		t.Errorf("no-op repair reported changes: %+v", result)
	}
	if len(result.Unchanged) == 0 {
		t.Error("no-op repair reported nothing at all; it should say what it looked at")
	}
}

// R139, R140 — the defect this caught the first time init ran on a real repository.
//
// mini-spec's own .gitignore already carried the **anchored** form (`/PENDING.md`).
// Textual line matching did not recognise it, so init appended the bare form — and
// because the last matching pattern wins, the duplicate silently replaced the
// project's narrower rule with one that also ignores docs/PENDING.md and every nested
// PENDING.md. Two failures in one: a redundant line, and a widened rule nobody asked
// for.
//
// Counting occurrences rather than asserting presence is the whole point: the broken
// version passed a Contains check.
func TestAnAlreadyIgnoredPathIsNotDuplicated(t *testing.T) {
	dir := t.TempDir()
	existing := "/PENDING.md\n/CURRENT.md\n/DONE.md\n"
	if err := os.WriteFile(filepath.Join(dir, GitIgnoreName), []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	// git reports them ignored, because they are — by the anchored lines above.
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith(TrajectoryFiles...)); err != nil {
		t.Fatal(err)
	}

	got := readGitignore(t, dir)
	for _, f := range TrajectoryFiles {
		if n := strings.Count(got, f); n != 1 {
			t.Errorf("%q appears %d times, want 1 — a duplicate widens the rule:\n%s", f, n, got)
		}
	}
	if strings.Contains(got, "\n"+TrajectoryFiles[0]) {
		t.Errorf("the bare form was appended alongside the anchored one:\n%s", got)
	}
}

// R139 — the same rule for a path ignored by a pattern we could never have matched
// textually. Only git can answer this.
func TestAPathIgnoredByAWildcardIsNotDuplicated(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, GitIgnoreName), []byte("*.md\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith(TrajectoryFiles...)); err != nil {
		t.Fatal(err)
	}
	got := readGitignore(t, dir)
	for _, f := range TrajectoryFiles {
		if strings.Contains(got, f) {
			t.Errorf("%q was added although *.md already ignores it:\n%s", f, got)
		}
	}
}

// R140 — new lines are anchored, because every governed path is mandated at the
// repository root and a bare pattern reaches further than intended.
func TestAddedIgnoreLinesAreAnchored(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(readGitignore(t, dir)), "\n") {
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "/") {
			t.Errorf("wrote unanchored pattern %q; it would match nested paths too", line)
		}
	}
}

// R141 — a repair that half-worked and said nothing is the silent partial success this
// tool exists to prevent. Removing our own lines cannot reach a rule in a nested
// .gitignore, .git/info/exclude, or the user's global excludes.
func TestRepairReportsWhatItCouldNotMakePublic(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackAll, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	// git insists the files are ignored, but the top-level .gitignore holds no line
	// naming them — so something outside this file is doing it.
	result, err := initIn(t, dir, TrackAll, true, gitWith(TrajectoryFiles...))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Unresolved) != len(TrajectoryFiles) {
		t.Fatalf("Unresolved = %v, want all three trajectory files", result.Unresolved)
	}
	if !strings.Contains(result.Report(), "STILL IGNORED") {
		t.Errorf("the report does not surface it:\n%s", result.Report())
	}
}

// R142, R146 — the precondition that keeps the two forms apart, and the refusal that
// must not stomp.
func TestPlainInitRefusesWhenConfigExists(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackAll, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(RepoConfigPath(dir))
	if err != nil {
		t.Fatal(err)
	}

	_, err = initIn(t, dir, TrackNone, false, gitWith())
	if err == nil {
		t.Fatal("plain init over an existing config succeeded, want refusal")
	}
	if !strings.Contains(err.Error(), "--repair") {
		t.Errorf("refusal does not name --repair: %v", err)
	}
	if !strings.Contains(err.Error(), "confirm the value with the user") {
		t.Errorf("refusal does not require user confirmation: %v", err)
	}
	after, _ := os.ReadFile(RepoConfigPath(dir))
	if string(after) != string(before) {
		t.Error("refused init modified the existing config")
	}
}

// R145 — the inverted precondition, checked in the other direction.
func TestRepairRefusesWhenConfigIsAbsent(t *testing.T) {
	dir := t.TempDir()
	_, err := initIn(t, dir, TrackAll, true, gitWith())
	if err == nil {
		t.Fatal("--repair with no config succeeded, want refusal")
	}
	if !strings.Contains(err.Error(), "init --track-") {
		t.Errorf("refusal does not point at plain init: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ConfigDirName)); err == nil {
		t.Error("refused --repair still created .minispec/")
	}
}

// R144 — the forward half of symmetry.
func TestRepairAddsWhatTheNewValueRequires(t *testing.T) {
	dir := t.TempDir()
	if _, err := initIn(t, dir, TrackAll, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	if _, err := initIn(t, dir, TrackPrivateTrajectory, true, gitWith()); err != nil {
		t.Fatal(err)
	}
	if v, _ := LoadTrack(RepoConfigPath(dir)); v != TrackPrivateTrajectory {
		t.Errorf("track = %q, want private-trajectory", v)
	}
	got := readGitignore(t, dir)
	for _, f := range TrajectoryFiles {
		if !strings.Contains(got, f) {
			t.Errorf(".gitignore missing %q after repair:\n%s", f, got)
		}
	}
}

// R144 — the reverse half. A one-directional repair could only ever make files more
// private, so a project that decided its queue should ship would have no path back.
// This is the direction that makes the command able to resolve a mismatch either way.
func TestRepairRemovesWhatTheNewValueForbids(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, GitIgnoreName), []byte("keep-me/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith()); err != nil {
		t.Fatal(err)
	}
	if _, err := initIn(t, dir, TrackAll, true, gitWith()); err != nil {
		t.Fatal(err)
	}
	got := readGitignore(t, dir)
	for _, f := range TrajectoryFiles {
		if strings.Contains(got, f) {
			t.Errorf("%q survived a repair to track: all:\n%s", f, got)
		}
	}
	if !strings.Contains(got, "keep-me/") {
		t.Errorf("an unrelated line was removed:\n%s", got)
	}
	if !strings.Contains(got, BackupIgnorePath) {
		t.Errorf("%q should survive a repair to track: all:\n%s", BackupIgnorePath, got)
	}
}

// R161, R162 — a flag sets a value and cannot undo arbitrary damage, so this is the
// precondition behind the one place the agent is authorised to edit the config.
func TestRepairRefusesAMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ConfigDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	broken := "track: all\n  : [unclosed\n"
	if err := os.WriteFile(RepoConfigPath(dir), []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := initIn(t, dir, TrackNone, true, gitWith())
	if err == nil {
		t.Fatal("--repair on a malformed config succeeded, want refusal")
	}
	for _, want := range []string{"malformed", "authorised to edit", "config-reference.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal does not contain %q: %v", want, err)
		}
	}
	after, _ := os.ReadFile(RepoConfigPath(dir))
	if string(after) != broken {
		t.Error("refused --repair modified the malformed config")
	}
}

// R141 — the agent did not write these files and must not have to infer what changed.
func TestReportNamesEveryFileTouched(t *testing.T) {
	dir := t.TempDir()
	result, err := initIn(t, dir, TrackPrivateTrajectory, false, gitWith())
	if err != nil {
		t.Fatal(err)
	}
	report := result.Report()
	if !strings.Contains(report, RepoConfigName) {
		t.Errorf("report does not name the config:\n%s", report)
	}
	if !strings.Contains(report, GitIgnoreName) {
		t.Errorf("report does not name .gitignore:\n%s", report)
	}
}
