// CRC: crc-Init.md | Seq: seq-bootstrap.md#2 | R136, R137, R138, R139, R140, R141, R142, R144, R145, R161, R170, R171, R172
package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
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
	broken := "track = \"all\"\n[unclosed\n"
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

// R173 — a configuration with no `track` predates the setting rather than being
// damaged, so --repair supplies what is absent instead of refusing.
func TestRepairAcceptsAConfigWithNoTrack(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ConfigDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(RepoConfigPath(dir), []byte("design_dir = \"design\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := initIn(t, dir, TrackAll, true, gitWith()); err != nil {
		t.Fatalf("--repair on a pre-track config failed: %v", err)
	}
	v, err := LoadTrack(RepoConfigPath(dir))
	if err != nil || v != TrackAll {
		t.Errorf("after repair LoadTrack = (%q, %v), want (all, nil)", v, err)
	}
}

// repairIn writes original as the repository configuration, runs --repair to track, and
// returns the file afterwards.
func repairIn(t *testing.T, original string, track TrackValue) (string, *InitResult) {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, RepoConfigPath(dir), original)
	result, err := initIn(t, dir, track, true, gitWith())
	if err != nil {
		t.Fatalf("--repair failed: %v", err)
	}
	return readFile(t, RepoConfigPath(dir)), result
}

// R524 — the writer edits one line and leaves every other byte as written.
//
// The YAML-era writer lost comments to the obvious implementation (decode into Config,
// encode back), and nothing caught it: the suite asserted what the file *gained* and
// never what it kept. Measured on this tool's own reference repository, where a repair
// deleted ten lines of comment explaining a non-obvious setting. So this compares bytes.
func TestRepairPreservesEveryByteOutsideTheTrackLine(t *testing.T) {
	original := `# why this project reads only Go
# (the reasoning a human left for the next reader)
code_extensions = [".go"]


design_dir = "design"   # spacing a person chose
`
	after, _ := repairIn(t, original, TrackPrivateTrajectory)
	if want := original + "track = \"private-trajectory\"\n"; after != want {
		t.Errorf("repair changed more than one line:\ngot:\n%s\nwant:\n%s", after, want)
	}
}

// R524 — the *replace* branch, which the insert case above does not reach. Only the
// quoted value changes: the comment above the line, the one trailing it, and the
// neighbouring setting stay as written.
func TestRepairChangingAnExistingValueKeepsItsComments(t *testing.T) {
	original := "# the queue ships with this repository on purpose\ntrack = \"all\"  # why this value\ndesign_dir = \"design\"\n"
	after, _ := repairIn(t, original, TrackPrivateTrajectory)
	want := "# the queue ships with this repository on purpose\ntrack = \"private-trajectory\"  # why this value\ndesign_dir = \"design\"\n"
	if after != want {
		t.Errorf("repair changed more than the value:\ngot:\n%s\nwant:\n%s", after, want)
	}
}

// R524 — a value that is already correct leaves the file byte-for-byte alone, so a
// no-op repair cannot reformat what it did not need to touch.
func TestRepairOfACorrectValueRewritesNothing(t *testing.T) {
	original := "# a note\ntrack   =   \"all\"\ndesign_dir = \"design\"\n"
	after, result := repairIn(t, original, TrackAll)
	if after != original {
		t.Errorf("no-op repair rewrote the config:\ngot:\n%s\nwant:\n%s", after, original)
	}
	if len(result.Unchanged) == 0 {
		t.Error("no-op repair did not report the config as unchanged")
	}
}

// R524 — every key after a table header belongs to that table, so an inserted `track`
// has to land above the first one. Appended at the end it would still be *in the file*,
// which is all a text check asks, and decode as the table's key.
func TestInsertedTrackGoesAboveTheFirstTable(t *testing.T) {
	original := "design_dir = \"design\"\n\n[[languages]]\nname = \"x\"\n"
	out, err := setTrack([]byte(original), TrackAll)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if _, err := toml.Decode(string(out), &doc); err != nil {
		t.Fatalf("result does not decode: %v\n%s", err, out)
	}
	if doc["track"] != "all" {
		t.Errorf("top-level track = %v, want all:\n%s", doc["track"], out)
	}
	if langs, _ := doc["languages"].([]map[string]any); len(langs) != 1 || langs[0]["track"] != nil {
		t.Errorf("the table gained a track, or was lost: %v\n%s", doc["languages"], out)
	}
}

// R524 — the degenerate documents have no line to replace and no table to insert before,
// so the line goes after whatever they hold, and nothing they hold moves.
func TestSetTrackHandlesDegenerateDocuments(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"absent", "", "track = \"none\"\n"},
		{"blank lines", "\n\n", "\n\ntrack = \"none\"\n"},
		{"comments only", "# nothing but a comment\n", "# nothing but a comment\ntrack = \"none\"\n"},
		{"no final newline", "design_dir = \"d\"", "design_dir = \"d\"\ntrack = \"none\"\n"},
	} {
		out, err := setTrack([]byte(tc.body), TrackNone)
		if err != nil {
			t.Errorf("%s: setTrack failed: %v", tc.name, err)
			continue
		}
		if string(out) != tc.want {
			t.Errorf("%s: setTrack = %q, want %q", tc.name, out, tc.want)
		}
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

// init is exempt from the gate, so it reports a YAML configuration itself — before the
// "nothing to repair" and "already exists" refusals, which would otherwise send the user
// the wrong way over a file they already have. R522
func TestInitReportsAYAMLConfigFirst(t *testing.T) {
	for _, repair := range []bool{false, true} {
		dir := t.TempDir()
		legacy := filepath.Join(dir, ConfigDirName, LegacyRepoConfigName)
		writeFile(t, legacy, "track: all\n")
		_, err := initIn(t, dir, TrackAll, repair, gitWith())
		if err == nil || !strings.Contains(err.Error(), legacy) || !strings.Contains(err.Error(), "TOML") {
			t.Errorf("repair=%v: error %v does not report the YAML file", repair, err)
		}
		if _, statErr := os.Stat(RepoConfigPath(dir)); statErr == nil {
			t.Errorf("repair=%v: init wrote %s beside the YAML file", repair, RepoConfigName)
		}
	}
}
