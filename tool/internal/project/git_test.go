// CRC: crc-Git.md | Seq: seq-bootstrap.md#1.7 | R151, R164, R166, R167, R168, R303, R304, R305, R306, R307
package project

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

// R205 — a method anchor is handed to git as a declaration-shaped pattern and resolves to
// its own declaration, not to the first same-named method in the file. Built against a
// real repository because the pattern is in git's regex dialect, not Go's.
//
// Layout chosen so the wrong answer has a different date: B.Run is first in the file and
// committed on day one with a function after it (so its -L range never grows); A.Run is
// appended on day two. A pattern that resolves "Run" to the first match reports day one
// for A.Run; the right one reports day two.
func TestLastChangedResolvesAMethodByItsReceiver(t *testing.T) {
	dir := newRepo(t)
	run := func(env []string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t"), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	day1 := []string{"GIT_AUTHOR_DATE=2026-01-01T12:00:00", "GIT_COMMITTER_DATE=2026-01-01T12:00:00"}
	day2 := []string{"GIT_AUTHOR_DATE=2026-01-02T12:00:00", "GIT_COMMITTER_DATE=2026-01-02T12:00:00"}
	base := "package p\n\ntype B struct{}\n\nfunc (b B) Run() int { return 2 }\n\nfunc sep() {}\n"
	write(t, dir, "x.go", base)
	run(day1, "add", "x.go")
	run(day1, "commit", "-qm", "b")
	write(t, dir, "x.go", base+"\ntype A struct{}\n\nfunc (a *A) Run() int { return 1 }\n")
	run(day2, "add", "x.go")
	run(day2, "commit", "-qm", "a")

	g := NewGit(dir)
	a, err := g.LastChanged("x.go", "A.Run")
	if err != nil {
		t.Fatalf("A.Run: %v — the method anchor did not resolve", err)
	}
	b, err := g.LastChanged("x.go", "B.Run")
	if err != nil {
		t.Fatalf("B.Run: %v — the method anchor did not resolve", err)
	}
	if got, want := a.Format("2006-01-02"), "2026-01-02"; got != want {
		t.Errorf("A.Run changed %s, want %s — the pattern resolved to the first Run in the file, not to A's", got, want)
	}
	if got, want := b.Format("2006-01-02"), "2026-01-01"; got != want {
		t.Errorf("B.Run changed %s, want %s", got, want)
	}
}

// R206 — a bare anchor is bounded, so it cannot resolve inside a longer name.
func TestLastChangedBoundsABareSymbol(t *testing.T) {
	dir := newRepo(t)
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	write(t, dir, "x.go", "package p\n\nfunc LookupPath() int { return 1 }\n")
	run("add", "x.go")
	run("commit", "-qm", "one")

	g := NewGit(dir)
	if _, err := g.LastChanged("x.go", "LookupPath"); err != nil {
		t.Fatalf("LookupPath: %v, want a date", err)
	}
	if _, err := g.LastChanged("x.go", "Lookup"); err != ErrUnresolvedSite {
		t.Errorf("Lookup resolved (%v); want ErrUnresolvedSite — an unbounded name matched inside LookupPath", err)
	}
}

// R237 — the anchor touches neither the working tree nor the real index.
func TestAnchorTouchesNeitherWorkingTreeNorIndex(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "seed.txt", "modified since the commit\n")
	g := NewGit(dir)
	before, err := g.run("status", "--porcelain")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "seed.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "modified since the commit\n" {
		t.Errorf("the anchor moved the modification out of the working tree: %q", body)
	}
	after, err := g.run("status", "--porcelain")
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("the anchor changed index or worktree state:\n  before %q\n  after  %q", before, after)
	}
}

// R237 — untracked contents in, ignored paths out.
func TestAnchorHoldsUntrackedContentsAndNoIgnoredPaths(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, ".gitignore", "ignored.txt\n")
	write(t, dir, "ignored.txt", "the slot's business, never the anchor's\n")
	write(t, dir, "new-carve.md", "untracked and wanted back\n")
	g := NewGit(dir)
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	listed, err := g.run("ls-tree", "-r", "--name-only", SnapshotRef)
	if err != nil {
		t.Fatal(err)
	}
	names := strings.Fields(listed)
	if !slices.Contains(names, "new-carve.md") {
		t.Errorf("untracked contents are absent from the anchor: %v", names)
	}
	if slices.Contains(names, "ignored.txt") {
		t.Errorf("an ignored path reached the anchor: %v", names)
	}
	body, err := g.run("show", SnapshotRef+":new-carve.md")
	if err != nil {
		t.Fatal(err)
	}
	if body != "untracked and wanted back\n" {
		t.Errorf("the anchor holds the name but not the contents: %q", body)
	}
}

// R236 — exactly one anchor exists.
func TestAnchorIsReplacedRatherThanAccumulated(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "seed.txt", "first\n")
	g := NewGit(dir)
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "seed.txt", "second\n")
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	body, err := g.run("show", SnapshotRef+":seed.txt")
	if err != nil {
		t.Fatal(err)
	}
	if body != "second\n" {
		t.Errorf("the anchor still holds the earlier tree: %q", body)
	}
}

// R238 — a fresh repository still gets an anchor, with no parent.
func TestAnchorIsWrittenInARepositoryWithNoCommits(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	write(t, dir, "a.txt", "untracked, and the only thing here\n")
	g := NewGit(dir)
	if err := g.Snapshot(); err != nil {
		t.Fatalf("no anchor in a repository with no commits: %v", err)
	}
	listed, err := g.run("ls-tree", "-r", "--name-only", SnapshotRef)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(strings.Fields(listed), "a.txt") {
		t.Errorf("the anchor is empty in a fresh repository: %q", listed)
	}
	if _, err := g.run("rev-parse", SnapshotRef+"^1"); err == nil {
		t.Error("the anchor claims a first parent in a repository with no commits")
	}
}

// R238 — the anchor's first parent is the checked-out commit.
func TestAnchorCarriesItsBaseCommitAsFirstParent(t *testing.T) {
	dir := newRepo(t)
	g := NewGit(dir)
	head, err := g.run("rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	base, err := g.run("rev-parse", SnapshotRef+"^1")
	if err != nil {
		t.Fatalf("the anchor has no first parent, so what it was taken from is unrecoverable: %v", err)
	}
	if strings.TrimSpace(base) != strings.TrimSpace(head) {
		t.Errorf("anchor^1 = %s, want the commit that was checked out, %s", base, head)
	}
}

// R303, R306 — the range is the declaration's own lines and nobody's comment. Git's
// funcname range gave a declaration the *successor's* doc comment, and old-sdom's first
// repair gave it its own; measured against the live corpus the second turned three
// verified alarms stale, every one a traceability line rewritten inside the doc block.
func TestLastChangedIgnoresACommentOnlyEdit(t *testing.T) {
	dir := newRepo(t)
	run := dated(t, dir)
	body := "package p\n\n// Foo does a thing.\nfunc Foo() int {\n\treturn 1\n}\n\n// Bar is next.\nfunc Bar() int { return 2 }\n"
	write(t, dir, "x.go", body)
	run(day(1), "add", "x.go")
	run(day(1), "commit", "-qm", "one")
	write(t, dir, "x.go", strings.NewReplacer("// Foo does a thing.", "// Foo does a thing, reworded.",
		"// Bar is next.", "// Bar is next, reworded at length.").Replace(body))
	run(day(2), "add", "x.go")
	run(day(2), "commit", "-qm", "comments only")

	g := NewGit(dir)
	when, err := g.LastChanged("x.go", "Foo")
	if err != nil {
		t.Fatal(err)
	}
	if got := when.Format("2006-01-02"); got != "2026-01-01" {
		t.Errorf("Foo changed %s after a comment-only edit; want 2026-01-01 — a comment is in the range", got)
	}
}

// R305 — the range ends where the declaration's groups close, not at the line before
// the next declaration: appending after the last function in a file must not stale it
// (gap O11, measured 2026-09-04: `func (a *A) Run()` with nothing after it was
// attributed to the commit that appended `B` below it).
func TestLastChangedIgnoresAnAppendAfterTheLastDeclaration(t *testing.T) {
	dir := newRepo(t)
	run := dated(t, dir)
	body := "package p\n\ntype A struct{}\n\nfunc (a *A) Run() int {\n\treturn 1\n}\n"
	write(t, dir, "x.go", body)
	run(day(1), "add", "x.go")
	run(day(1), "commit", "-qm", "one")
	write(t, dir, "x.go", body+"\nfunc B() int { return 2 }\n")
	run(day(2), "add", "x.go")
	run(day(2), "commit", "-qm", "append")

	when, err := NewGit(dir).LastChanged("x.go", "A.Run")
	if err != nil {
		t.Fatal(err)
	}
	if got := when.Format("2006-01-02"); got != "2026-01-01" {
		t.Errorf("A.Run changed %s after an append below it; want 2026-01-01 — the trailing blank line is in the range", got)
	}
}

// R307 — a bare name two declarations answer to is ambiguous, not resolved to the first.
func TestLastChangedRefusesAnAmbiguousBareName(t *testing.T) {
	dir := newRepo(t)
	run := dated(t, dir)
	write(t, dir, "x.go", "package p\n\ntype A struct{}\ntype B struct{}\n\nfunc (a A) Run() int { return 1 }\n\nfunc (b B) Run() int { return 2 }\n")
	run(day(1), "add", "x.go")
	run(day(1), "commit", "-qm", "one")

	g := NewGit(dir)
	_, err := g.LastChanged("x.go", "Run")
	var amb *AmbiguousSiteError
	if !errors.As(err, &amb) || amb.N != 2 || !errors.Is(err, ErrAmbiguousSite) {
		t.Errorf("Run = %v; want AmbiguousSiteError{N: 2} — the anchor watched an arbitrary one of two", err)
	}
	if _, err := g.LastChanged("x.go", "A.Run"); err != nil {
		t.Errorf("A.Run = %v; want a date — the receiver form disambiguates", err)
	}
}

// R303, R304, R305 — the extent itself, over shapes the repository tests do not reach: a
// signature spanning lines, a nested group, a grouped var, a receiver with a type parameter.
func TestSiteExtentShapes(t *testing.T) {
	src := "package p\n" + // 1
		"\n" + // 2
		"var (\n" + // 3
		"\tA = 1\n" + // 4
		"\tB = 2\n" + // 5
		")\n" + // 6
		"\n" + // 7
		"// Doc for Multi.\n" + // 8
		"func Multi(\n" + // 9
		"\ta int,\n" + // 10
		") (int, error) {\n" + // 11
		"\tif a > 0 {\n" + // 12
		"\t\treturn a, nil\n" + // 13
		"\t}\n" + // 14
		"\treturn 0, nil\n" + // 15
		"}\n" + // 16
		"\n" + // 17
		"func (s *Set[T]) Add(v T) {}\n" + // 18
		"\n" + // 19
		"func One() {}\n" // 20
	for _, c := range []struct {
		symbol     string
		start, end int
		count      int
	}{
		{"A", 4, 4, 1}, {"B", 5, 5, 1}, {"Multi", 9, 16, 1}, {"Set.Add", 18, 18, 1},
		{"Add", 18, 18, 1}, {"One", 20, 20, 1}, {"Other.Add", 0, 0, 0}, {"Nope", 0, 0, 0},
	} {
		start, end, n := siteExtent(src, c.symbol)
		if start != c.start || end != c.end || n != c.count {
			t.Errorf("%s: got %d-%d count %d, want %d-%d count %d", c.symbol, start, end, n, c.start, c.end, c.count)
		}
	}
}

// dated returns a git runner for dir taking a date's environment first.
func dated(t *testing.T, dir string) func(env []string, args ...string) {
	return func(env []string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t"), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
}

func day(n int) []string {
	stamp := fmt.Sprintf("2026-01-%02dT12:00:00", n)
	return []string{"GIT_AUTHOR_DATE=" + stamp, "GIT_COMMITTER_DATE=" + stamp}
}

// R335 — three categories from one tree-vs-tree diff, verified on a tree carrying all three
// at once, the D being a deleted untracked file: the case a name list could report and not
// return. Nothing in the working tree is altered by the report.
func TestChangesSinceSnapshotReportsAllThreeCategories(t *testing.T) {
	dir := newRepo(t)
	write(t, dir, "tracked.go", "package p\n")
	write(t, dir, "gone.txt", "untracked and about to be deleted\n")
	run := dated(t, dir)
	run(day(1), "add", "tracked.go")
	run(day(1), "commit", "-qm", "base")
	g := NewGit(dir)
	if _, err := g.ChangesSinceSnapshot(); !errors.Is(err, ErrNoSnapshot) {
		t.Fatalf("with no anchor: %v; want ErrNoSnapshot — a report on nothing reads as clean", err)
	}
	if err := g.Snapshot(); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "tracked.go", "package p\n\nvar changed = true\n")
	write(t, dir, "fresh.md", "new since the transition\n")
	if err := os.Remove(filepath.Join(dir, "gone.txt")); err != nil {
		t.Fatal(err)
	}
	changes, err := g.ChangesSinceSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, c := range changes {
		got[c.Path] = c.Status
	}
	want := map[string]string{"tracked.go": "M", "fresh.md": "A", "gone.txt": "D"}
	for path, status := range want {
		if got[path] != status {
			t.Errorf("%s: got %q, want %q (all: %v)", path, got[path], status, changes)
		}
	}
	if len(changes) != 3 {
		t.Errorf("got %d changes, want 3: %v", len(changes), changes)
	}
	// The report is reading only: the tree is as it was before the call.
	if _, err := os.Stat(filepath.Join(dir, "gone.txt")); !errors.Is(err, os.ErrNotExist) {
		t.Error("the report restored a deleted file — it must alter nothing")
	}
	if out, _ := exec.Command("git", "-C", dir, "status", "--porcelain").Output(); !strings.Contains(string(out), " M tracked.go") || !strings.Contains(string(out), "?? fresh.md") {
		t.Errorf("the report touched the index or the tree:\n%s", out)
	}
}
