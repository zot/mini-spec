// CRC: crc-Git.md | Seq: seq-bootstrap.md | R167
package project

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ErrNoGit reports that no ignore or tracking state could be determined — there is no
// usable `git`, or the directory is not inside a working tree. Returned rather than a
// map of falses because a caller can mistake "none of these are ignored" for a
// verified answer, and cannot mistake an error for one. R168
var ErrNoGit = errors.New("not a git working tree: ignore state cannot be determined")

// ErrUnresolvedSite reports that git could not find the named symbol in the named
// file. Distinguished from "has not changed" because they are opposite findings: one
// says the anchor is intact and idle, the other says the anchor no longer points at
// anything. Collapsing them would make a rotted anchor read as a fresh proof. R182
var ErrUnresolvedSite = errors.New("git cannot resolve that symbol in that file")

// ErrNoHistory reports that git holds no history for a path — it is untracked, or
// newly written and not yet committed. Distinct from ErrNoGit, whose message is about
// a missing working tree and would be untrue here, and distinct from ErrUnresolvedSite,
// which says the file is tracked and the *symbol* is what is missing. Three different
// findings that a caller must be able to tell apart. R182, R184
var ErrNoHistory = errors.New("git holds no history for that path")

// CRC: crc-Git.md | R167
// GitFacts is what a consistency check needs to know about a working tree. It exists
// as an interface so Track can be verified against a stated world instead of the
// machine's: the seam is what makes the check testable without a repository.
type GitFacts interface {
	IsRepo() bool
	Ignored(paths []string) (map[string]bool, error)
	Tracked(path string) (bool, error)
	LastChanged(file, symbol string) (time.Time, error)
	// SiteResolves answers only whether a site names exactly one declaration in the
	// file as committed — the cheap half of LastChanged, without the history walk. R309
	SiteResolves(file, symbol string) error
}

// CRC: crc-Git.md | Seq: seq-bootstrap.md#1.7 | R166, R167
// Git answers questions about the version-controlled working tree by invoking the
// `git` command line. Every answer is a computed property, so there is nothing here to
// store and nothing to go stale.
//
// It reads git and never drives it: no staging, no commits, no resets, no branch
// state. Editing `.gitignore` is an ordinary file edit and lives in init.go.
type Git struct {
	workDir string
	// probed caches the working-tree answer for one invocation: a tree either is
	// git-managed or is not for the whole of a run, so asking twice cannot differ.
	probed bool
	isRepo bool
	// head caches each file as committed at HEAD, nil for a path HEAD does not hold. R308
	head map[string]*string
}

// NewGit returns a Git answering about the given directory.
func NewGit(workDir string) *Git {
	return &Git{workDir: workDir}
}

// CRC: crc-Git.md | Seq: seq-bootstrap.md#1.7 | R151, R168
// IsRepo reports whether workDir is inside a git working tree. False is a legitimate
// answer rather than an error — a project with no version control is a supported shape,
// and `track: none` is the honest declaration for it.
func (g *Git) IsRepo() bool {
	if !g.probed {
		g.probed = true
		out, err := g.run("rev-parse", "--is-inside-work-tree")
		g.isRepo = err == nil && strings.TrimSpace(out) == "true"
	}
	return g.isRepo
}

// CRC: crc-Git.md | Seq: seq-bootstrap.md#1.9 | R148
// Ignored reports, per path, whether git would ignore it. The whole question costs one
// invocation because `check-ignore` takes a list, which is what makes verifying on
// every run affordable rather than something to sample.
//
// Paths are returned keyed exactly as they were passed, so a caller reads back the
// answer to the question it asked rather than matching on count.
func (g *Git) Ignored(paths []string) (map[string]bool, error) {
	if !g.IsRepo() {
		return nil, ErrNoGit
	}
	result := make(map[string]bool, len(paths))
	if len(paths) == 0 {
		return result, nil
	}
	byAbs := make(map[string]string, len(paths))
	args := make([]string, 0, len(paths)+1)
	args = append(args, "check-ignore")
	for _, p := range paths {
		abs := g.abs(p)
		byAbs[abs] = p
		args = append(args, abs)
		result[p] = false
	}
	// check-ignore exits 1 when nothing matched, which is an answer rather than a
	// failure; it prints the matching paths and nothing else.
	out, _ := g.run(args...)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if orig, ok := byAbs[g.abs(line)]; ok {
			result[orig] = true
		}
	}
	return result, nil
}

// abs resolves a path against the working directory, so a caller's spelling of a path
// and git's own spelling of the same path compare equal.
func (g *Git) abs(p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(g.workDir, p)
}

// CRC: crc-Git.md | R164
// Tracked reports whether a path is known to the index — the fact behind the
// `.minispec/config.yaml` preference.
func (g *Git) Tracked(path string) (bool, error) {
	if !g.IsRepo() {
		return false, ErrNoGit
	}
	// --error-unmatch exits non-zero for a path the index does not hold, which is the
	// answer rather than a failure.
	out, err := g.run("ls-files", "--error-unmatch", "--", path)
	if err != nil {
		return false, nil
	}
	return strings.TrimSpace(out) != "", nil
}

// CRC: crc-Git.md | Seq: seq-alarm-freshness.md#1.5 | R180, R182, R184, R303, R307, R308
// LastChanged reports when the named function in the named file last changed.
//
// **The question is asked of the function, not the file**, because a file-level answer
// marks every alarm in a busy file stale and so discriminates nothing. Measured on this
// project's sibling repository before the check existed: file granularity called 8 of
// 13 alarms stale, function granularity called 3, and hand-checking those left 1. A
// check that reports everything is discarded as noise within a week, which makes the
// coarse version worse than none.
//
// **The reader computes the range; git is asked when those lines changed.** The query
// was `git log -L :pattern:file`, which asks git to *find* the symbol as well as bound
// it, and git bounds a declaration at the line before the next one — so the range
// carried the successor's doc comment and the trailing blank line, and the last
// declaration in a file read stale on every append (gaps O10, O11). The bounds now come
// from a parse of the file **as committed at HEAD**, because `-L <start>,<end>` resolves
// those numbers against HEAD and walks them backwards: an uncommitted edit higher in
// the file moves every later declaration, and working-tree numbers would then name
// somebody else's code in history. Measured on old-sdom, 2026-08-21: two alarms stale
// over a function nobody had touched.
//
// `git log` output interleaves format lines with diff hunks, and commits arrive
// newest-first, so the first format line is the answer.
//
// Five outcomes are kept apart deliberately: no working tree (ErrNoGit), a path git
// holds no history for (ErrNoHistory), a tracked file whose symbol the parse cannot
// find (ErrUnresolvedSite), a symbol declared more than once (AmbiguousSiteError), and
// a function that exists and has never changed (zero time, no error). Only the last is
// a clean result, and the four before it must never be mistaken for one.
func (g *Git) LastChanged(file, symbol string) (time.Time, error) {
	if !g.IsRepo() {
		return time.Time{}, ErrNoGit
	}
	// A file git does not track has no history to search, which is "could not look"
	// rather than "the anchor points at nothing". Collapsing the two would make every
	// alarm anchored into a newly written file report as a rotted anchor — found by
	// running this check against its own uncommitted source.
	if tracked, terr := g.Tracked(file); terr != nil || !tracked {
		return time.Time{}, ErrNoHistory
	}
	start, end, err := g.siteRange(file, symbol)
	if err != nil {
		return time.Time{}, err
	}
	out, err := g.run("log", "-L", fmt.Sprintf("%d,%d:%s", start, end, file), "--format=%H|%ad", "--date=short")
	if err != nil {
		return time.Time{}, ErrUnresolvedSite
	}
	for _, line := range strings.Split(out, "\n") {
		hash, date, ok := strings.Cut(strings.TrimSpace(line), "|")
		// Any object-format hash, not just SHA-1's forty characters. Pinning the
		// length to 40 made every line of a SHA-256 repository fail to match, so the
		// loop fell through to "no changes" and reported every alarm verified — a
		// check that could not look returning a clean result, which is the precise
		// failure this whole feature exists to prevent, inside its own implementation.
		if !ok || !isObjectHash(hash) {
			continue
		}
		when, perr := time.Parse("2006-01-02", date)
		if perr != nil {
			continue
		}
		return when, nil
	}
	return time.Time{}, nil
}

// CRC: crc-Git.md | Seq: seq-alarm-freshness.md#1.3 | R309
// SiteResolves reports whether a site names exactly one declaration in the file as
// committed, with the same four failures LastChanged keeps apart and none of its cost:
// no history walk, so a census can ask it of every prescription without the expensive
// path growing past the verified population.
func (g *Git) SiteResolves(file, symbol string) error {
	if !g.IsRepo() {
		return ErrNoGit
	}
	if tracked, terr := g.Tracked(file); terr != nil || !tracked {
		return ErrNoHistory
	}
	_, _, err := g.siteRange(file, symbol)
	return err
}

// CRC: crc-Git.md | R315
// SiteRange is the site's line range in HEAD's copy of the file — what a recorded pull
// was earned against — and DiskRange the same over the working tree, which is what a
// rewritten anchor names now. `update inject` compares the two: the old sites in
// history, where a renamed symbol still exists, and the new sites on disk.
func (g *Git) SiteRange(file, symbol string) (start, end int, err error) {
	if !g.IsRepo() {
		return 0, 0, ErrNoGit
	}
	return g.siteRange(file, symbol)
}

// R315
func (g *Git) DiskRange(file, symbol string) (start, end int, err error) {
	src, rerr := os.ReadFile(g.abs(file))
	if rerr != nil {
		return 0, 0, ErrUnresolvedSite
	}
	start, end, n := siteExtent(string(src), symbol)
	switch {
	case n > 1:
		return 0, 0, &AmbiguousSiteError{Symbol: symbol, N: n}
	case n == 0:
		return 0, 0, ErrUnresolvedSite
	}
	return start, end, nil
}

// Seq: seq-alarm-freshness.md#1.5.1, seq-alarm-freshness.md#1.5.2, seq-alarm-freshness.md#1.5.3 | R307, R308
// siteRange is the site's line range in HEAD's copy of the file, or the failure that
// stands in for it.
func (g *Git) siteRange(file, symbol string) (start, end int, err error) {
	head, ok := g.headFile(file)
	if !ok {
		return 0, 0, ErrNoHistory
	}
	start, end, n := siteExtent(head, symbol)
	switch {
	case n > 1:
		// R307 — counted before the extent is trusted, because the extent cannot report
		// this: the anchor has been watching an arbitrary one of them since it was
		// written, and every answer about it has been confident and unfounded.
		return 0, 0, &AmbiguousSiteError{Symbol: symbol, N: n}
	case n == 0:
		// A symbol absent from HEAD's file and present on disk is one written since the
		// last commit — *no history yet* rather than a rotted anchor, which is the
		// distinction this check would otherwise get backwards on every new function.
		if src, rerr := os.ReadFile(g.abs(file)); rerr == nil {
			if _, _, live := siteExtent(string(src), symbol); live > 0 {
				return 0, 0, ErrNoHistory
			}
		}
		return 0, 0, ErrUnresolvedSite
	}
	return start, end, nil
}

// R308
// headFile is the file's content as committed at HEAD, cached per file for one
// invocation: a census asks about many sites in one file, and the range is computed
// once per site over the same bytes. False when HEAD holds no such path.
func (g *Git) headFile(file string) (string, bool) {
	if g.head == nil {
		g.head = map[string]*string{}
	}
	if cached, seen := g.head[file]; seen {
		if cached == nil {
			return "", false
		}
		return *cached, true
	}
	// `./` makes the path relative to the working directory, as `-L` already treats it;
	// bare `HEAD:path` is relative to the repository root, and a design root beneath
	// it (this repository's `tool/`) read every site as having no history.
	out, err := g.run("show", "HEAD:./"+filepath.ToSlash(file))
	if err != nil {
		g.head[file] = nil
		return "", false
	}
	g.head[file] = &out
	return out, true
}

// CRC: crc-Git.md | R307
// AmbiguousSiteError reports a symbol declared more than once in its file. A state of
// its own rather than ErrUnresolvedSite, because the repair differs: the anchor points
// at something, and the fix is to say which one — `Type.Method`.
type AmbiguousSiteError struct {
	Symbol string
	N      int
}

func (e *AmbiguousSiteError) Error() string {
	return fmt.Sprintf("%s is declared %d times in that file; name the receiver", e.Symbol, e.N)
}

// ErrAmbiguousSite is the sentinel every AmbiguousSiteError matches under errors.Is.
var ErrAmbiguousSite = errors.New("symbol declared more than once")

func (e *AmbiguousSiteError) Is(target error) bool { return target == ErrAmbiguousSite }

// isObjectHash reports whether a token is a full object name in any format git
// supports — 40 hex characters for SHA-1, 64 for SHA-256.
func isObjectHash(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return false
		}
	}
	return true
}

// ErrGitFailed reports a git invocation that failed for a reason other than the ones the
// callers name; the wrapped message carries git's own words.
var ErrGitFailed = errors.New("git failed")

// SnapshotRef is where the worktree anchor lives — a ref the tool owns outright.
//
// **Deliberately not the stash**: a stash entry can be applied by someone reaching for their
// own work, `git stash clear` destroys it at any position in the stack, and replacing it
// would mean finding the previous entry by message and dropping it instead of one atomic
// write. R236
const SnapshotRef = "refs/minispec/snapshot"

// CRC: crc-Git.md | Seq: seq-backup.md#4 | R236, R237, R238
// Snapshot writes the worktree anchor: the working tree as it stands, recorded and reachable,
// with nothing moved, staged or restored.
//
// **Reference, not undo.** Building the object destroys nothing, which is what makes it safe
// to take over work in progress the tool has no business owning.
//
// **The obvious command is the wrong one and it is one word away**: the familiar stash verb
// *relocates* the changes. Worse, the flag that looks like the fix is a lie — `git stash
// create -u` exits 0, hands back a plausible object, and holds no untracked file at all
// (measured 2026-08-17). A check asking only "did it error?" reports it working.
func (g *Git) Snapshot() error {
	if !g.IsRepo() {
		return ErrNoGit
	}
	tree, err := g.scratchTree()
	if err != nil {
		return err
	}
	// The current commit becomes the anchor's first parent, so what it was taken from is
	// recoverable from the anchor itself. A repository with no commits has no parent. R238
	//
	// **Both guards are load-bearing, and the error one is the half that looks droppable.**
	// In a repository with no commits `git rev-parse HEAD` exits 128 *and prints `HEAD` to
	// stdout* — measured 2026-08-17. Keeping only the emptiness check passes `-p HEAD` to
	// `commit-tree` and the anchor cannot be written at all.
	args := []string{"commit-tree", tree, "-m", "minispec worktree anchor"}
	if head, err := g.run("rev-parse", "HEAD"); err == nil {
		if h := strings.TrimSpace(head); h != "" {
			args = append(args, "-p", h)
		}
	}
	out, err := g.run(args...)
	if err != nil {
		return fmt.Errorf("%w: git commit-tree: %v", ErrGitFailed, err)
	}
	// One write, so exactly one anchor exists and nothing has to be found and dropped. R236
	if _, err := g.run("update-ref", SnapshotRef, strings.TrimSpace(out)); err != nil {
		return fmt.Errorf("%w: git update-ref: %v", ErrGitFailed, err)
	}
	return nil
}

// CRC: crc-Git.md | Seq: seq-backup.md#4 | R335, R336
// Change is one path that has moved since the anchor: M changed, A new — tracked or not —
// D deleted, and recoverable, because the anchor holds the contents.
type Change struct {
	Status string `json:"status"` // "M", "A" or "D"
	Path   string `json:"path"`
}

// ErrNoSnapshot reports that no transition has anchored the tree yet, so there is nothing to
// diff against; a report built on nothing would read as clean for exactly one run.
var ErrNoSnapshot = errors.New("no worktree anchor: nothing has transitioned since this repository was initialised")

// CRC: crc-Git.md | Seq: seq-backup.md#4 | R335
// ChangesSinceSnapshot reports what has moved in the working tree since the anchor: a tree of
// the tree as it stands now, built by the same scratch-index method the anchor was, diffed
// against the anchor's tree. Three categories from one tree-vs-tree diff, no side-car list and
// no second source to keep in step; renames are reported as their two halves so the
// categories stay three. It reads and builds objects and alters nothing.
func (g *Git) ChangesSinceSnapshot() ([]Change, error) {
	if !g.IsRepo() {
		return nil, ErrNoGit
	}
	anchorTree := SnapshotRef + "^{tree}"
	if _, err := g.run("rev-parse", "--verify", "--quiet", anchorTree); err != nil {
		return nil, ErrNoSnapshot
	}
	now, err := g.scratchTree()
	if err != nil {
		return nil, err
	}
	out, err := g.run("diff-tree", "-r", "--no-renames", "--name-status", anchorTree, now)
	if err != nil {
		return nil, fmt.Errorf("%w: git diff-tree: %v", ErrGitFailed, err)
	}
	var changes []Change
	for _, line := range splitLines(out) {
		status, path, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		changes = append(changes, Change{Status: status[:1], Path: path})
	}
	return changes, nil
}

// R237
// scratchTree writes a tree object for the whole working tree, using an index of its own.
//
// The scratch index is the entire mechanism, and it buys three properties **structurally**:
// untracked file *contents* are included, ignored paths are excluded — so the trajectory
// files remain the backup slot's business and the two mechanisms cannot overlap — and the
// repository's real index and working tree are never touched.
func (g *Git) scratchTree() (string, error) {
	f, err := os.CreateTemp("", "minispec-index-")
	if err != nil {
		return "", err
	}
	idx := f.Name()
	f.Close()
	// Removed rather than truncated: git wants to create this file itself, and an existing
	// empty file is not a valid index.
	os.Remove(idx)
	defer os.Remove(idx)

	env := []string{"GIT_INDEX_FILE=" + idx}
	if _, err := g.runEnv(env, "add", "-A"); err != nil {
		return "", fmt.Errorf("%w: git add -A: %v", ErrGitFailed, err)
	}
	out, err := g.runEnv(env, "write-tree")
	if err != nil {
		return "", fmt.Errorf("%w: git write-tree: %v", ErrGitFailed, err)
	}
	return strings.TrimSpace(out), nil
}

// runEnv is run with additional environment. Separate rather than a variadic option on run
// because exactly one caller needs it and the whole point of that caller is that its index
// is *not* the repository's.
func (g *Git) runEnv(env []string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.Output()
	return string(out), err
}

// run invokes git in the working directory. Shelling out rather than linking a library
// is deliberate: it gets tracked and ignored status for free and stays correct as git
// changes, at the stated cost that no other VCS is supported. R167
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	return string(out), err
}
