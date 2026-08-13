// CRC: crc-Git.md | Seq: seq-bootstrap.md#1.7 | R151, R164, R166, R167, R168
package project

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These are the tests that keep watch on a dependency this tool does not control.
// Every other component is verified against a fake that assumes Git is right, so if
// `git check-ignore` ever answered differently the mistake would be invisible
// everywhere else. See test-Git.md.
//
// A real repository is built with `git init` — cheap enough to run always, and skipped
// with a clear message rather than silently passing when git is absent.

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH: these tests verify real git behavior and cannot be faked")
	}
}

// newRepo builds a repository with an initial commit, so HEAD exists.
func newRepo(t *testing.T) string {
	t.Helper()
	requireGit(t)
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init")
	write(t, dir, "seed.txt", "seed\n")
	run("add", "seed.txt")
	run("commit", "-m", "seed")
	return dir
}

func write(t *testing.T, dir, rel, body string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// R167 — the baseline the whole consistency check rests on.
func TestFreshRepoIsAWorkingTree(t *testing.T) {
	dir := newRepo(t)
	if !NewGit(dir).IsRepo() {
		t.Error("IsRepo() = false in a fresh repository")
	}
}

// R151, R168 — absence is an answer, not an error: a project with no version control
// is a supported shape.
func TestBareDirectoryIsNotAWorkingTree(t *testing.T) {
	requireGit(t)
	if NewGit(t.TempDir()).IsRepo() {
		t.Error("IsRepo() = true outside any repository")
	}
}

// R167 — the tool runs from wherever the user is, not from the root.
func TestSubdirectoryIsStillAWorkingTree(t *testing.T) {
	dir := newRepo(t)
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if !NewGit(sub).IsRepo() {
		t.Error("IsRepo() = false in a subdirectory of a repository")
	}
}

// R148 — the batching that makes checking on every run affordable. The answers must
// correspond to the paths asked about, not merely come out at the right count.
func TestIgnoredAnswersPerPathInOneInvocation(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, ".gitignore", "PENDING.md\nDONE.md\n")

	got, err := NewGit(dir).Ignored([]string{"PENDING.md", "CURRENT.md", "DONE.md", "README.md"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"PENDING.md": true, "CURRENT.md": false, "DONE.md": true, "README.md": false,
	}
	for path, expected := range want {
		if got[path] != expected {
			t.Errorf("Ignored[%q] = %v, want %v", path, got[path], expected)
		}
	}
}

// R167 — the reason to ask git rather than parse .gitignore ourselves: ignore rules
// compose across directories in ways a naive reader gets wrong.
func TestNestedGitignoreIsHonored(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, ".gitignore", "")
	write(t, dir, "sub/.gitignore", "x.md\n")

	got, err := NewGit(dir).Ignored([]string{"sub/x.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !got["sub/x.md"] {
		t.Error("a path ignored by a nested .gitignore was reported as not ignored")
	}
}

// R167 — the sentry case. `!` reverses an earlier rule, and a hand-rolled matcher is
// exactly where that gets missed.
func TestNegatedPatternIsNotIgnored(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, ".gitignore", "*.md\n!KEEP.md\n")

	got, err := NewGit(dir).Ignored([]string{"x.md", "KEEP.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !got["x.md"] {
		t.Error("x.md should be ignored by *.md")
	}
	if got["KEEP.md"] {
		t.Error("KEEP.md should be un-ignored by !KEEP.md")
	}
}

// R164 — the fact behind the .minispec/config.yaml preference.
func TestTrackedDistinguishesAddedFromUntracked(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "loose.txt", "loose\n")
	g := NewGit(dir)

	if tracked, err := g.Tracked("seed.txt"); err != nil || !tracked {
		t.Errorf("Tracked(seed.txt) = (%v, %v), want (true, nil)", tracked, err)
	}
	if tracked, err := g.Tracked("loose.txt"); err != nil || tracked {
		t.Errorf("Tracked(loose.txt) = (%v, %v), want (false, nil)", tracked, err)
	}
}

// R168 — the narrowing is stated rather than passed silently. This is the failure mode
// the project is most alert to: a check that could not look still returning a clean
// answer. An error cannot be mistaken for "none of these are ignored"; a map of falses
// can.
func TestNonGitTreeReportsIgnoreStateUnavailable(t *testing.T) {
	requireGit(t)
	g := NewGit(t.TempDir())

	got, err := g.Ignored([]string{"PENDING.md"})
	if !errors.Is(err, ErrNoGit) {
		t.Errorf("Ignored() error = %v, want ErrNoGit", err)
	}
	if got != nil {
		t.Errorf("Ignored() = %v, want nil so no caller can read it as a verdict", got)
	}
	if _, err := g.Tracked("anything"); !errors.Is(err, ErrNoGit) {
		t.Errorf("Tracked() error = %v, want ErrNoGit", err)
	}
}

// R166 — codifies the boundary: the tool reads git and never drives it.
func TestNoGitOperationMutatesTheRepository(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, ".gitignore", "PENDING.md\n")
	write(t, dir, "loose.txt", "loose\n")

	snapshot := func() string {
		t.Helper()
		var parts []string
		for _, args := range [][]string{{"status", "--porcelain"}, {"rev-parse", "HEAD"}} {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("git %s: %v", strings.Join(args, " "), err)
			}
			parts = append(parts, string(out))
		}
		return strings.Join(parts, "\x00")
	}

	before := snapshot()
	g := NewGit(dir)
	g.IsRepo()
	if _, err := g.Ignored([]string{"PENDING.md", "loose.txt"}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Tracked("seed.txt"); err != nil {
		t.Fatal(err)
	}
	if after := snapshot(); after != before {
		t.Errorf("a Git method changed repository state:\nbefore %q\nafter  %q", before, after)
	}
}

// R180 — LastChanged must read any object format git produces.
//
// Written after the fact: the first version gated on `len(hash) != 40`, so in a
// SHA-256 repository every format line failed to match, the scan fell through to "no
// changes", and every alarm reported verified. A check that could not look returning a
// clean result is the precise failure the freshness feature exists to prevent, so this
// test costs a real repository rather than being deferred to a fake.
func TestLastChangedReadsAnyObjectFormat(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		dir := t.TempDir()
		run := func(args ...string) {
			t.Helper()
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
				"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Skipf("%s unsupported here (%v): %s", format, err, out)
			}
		}
		run("init", "-q", "--object-format="+format, ".")
		if err := os.WriteFile(filepath.Join(dir, "a.go"),
			[]byte("package p\n\nfunc Foo() int { return 1 }\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "a.go")
		run("commit", "-qm", "one")

		when, err := NewGit(dir).LastChanged("a.go", "Foo")
		if err != nil {
			t.Errorf("%s: LastChanged returned %v, want a date", format, err)
			continue
		}
		if when.IsZero() {
			t.Errorf("%s: LastChanged reported no change for a function that was just committed — "+
				"a format this cannot read silently reports every alarm verified", format)
		}
	}
}

// R182, R184 — a function added since the last commit is *new*, not gone.
//
// `git log -L` searches the file as committed, so a symbol present on disk but absent
// from history makes it fail. Reporting that as a rotted anchor is false and fires on
// every newly written function — found by running the freshness check against this
// feature's own new code, where `LastChanged` itself reported unresolvable.
func TestLastChangedTellsANewFunctionFromAGoneOne(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	path := filepath.Join(dir, "a.go")
	run("init", "-q", ".")
	if err := os.WriteFile(path, []byte("package p\n\nfunc Old() int { return 1 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "a.go")
	run("commit", "-qm", "one")
	// Added on disk, never committed.
	if err := os.WriteFile(path,
		[]byte("package p\n\nfunc Old() int { return 1 }\n\nfunc Fresh() int { return 2 }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	g := NewGit(dir)
	if _, err := g.LastChanged("a.go", "Fresh"); !errors.Is(err, ErrNoHistory) {
		t.Errorf("a newly written function = %v, want ErrNoHistory — it is new, not gone", err)
	}
	if _, err := g.LastChanged("a.go", "Vanished"); !errors.Is(err, ErrUnresolvedSite) {
		t.Errorf("a symbol absent from disk and history = %v, want ErrUnresolvedSite", err)
	}
}
