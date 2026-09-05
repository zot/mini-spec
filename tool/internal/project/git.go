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

// CRC: crc-Git.md | Seq: seq-alarm-freshness.md#1.5 | R180, R182, R184
// LastChanged reports when the named function in the named file last changed.
//
// **The question is asked of the function, not the file**, because a file-level answer
// marks every alarm in a busy file stale and so discriminates nothing. Measured on this
// project's sibling repository before the check existed: file granularity called 8 of
// 13 alarms stale, function granularity called 3, and hand-checking those left 1. A
// check that reports everything is discarded as noise within a week, which makes the
// coarse version worse than none.
//
// `git log -L :symbol:file` answers it directly. Its output interleaves format lines
// with diff hunks, and commits arrive newest-first, so the first format line is the
// answer.
//
// Four outcomes are kept apart deliberately: no working tree (ErrNoGit), a path git
// holds no history for (ErrNoHistory), a tracked file whose symbol git cannot find
// (ErrUnresolvedSite), and a function that exists and has never changed (zero time, no
// error). Only the last is a clean result, and the three before it must never be
// mistaken for one.
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
	out, err := g.run("log", "-L", ":"+sitePattern(symbol)+":"+file, "--format=%H|%ad", "--date=short")
	if err != nil {
		// `-L` searches the file as committed, so a function added since the last
		// commit is absent from history while being perfectly present on disk. Reporting
		// that as a rotted anchor is false and fires on every newly written function —
		// found by running this check against its own new code. If the working tree
		// still declares the symbol, the honest answer is that there is no history yet.
		if declaresSymbol(g.abs(file), symbol) {
			return time.Time{}, ErrNoHistory
		}
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

// CRC: crc-Git.md | R205, R206
// sitePattern is the regex git is handed for `-L :<pattern>:<file>`, built from an
// `**Inject:**` symbol.
//
// **The anchor is not handed over as written.** `Doc.Render` as a regex matches no line
// of Go — the dot is any character and the shape never occurs on a declaration line — so
// every method-form anchor read *unresolvable*; measured 2026-09-04, 69 alarms in a
// sibling project over nothing else. A `Type.Method` symbol becomes a declaration-shaped
// pattern over its receiver, with the receiver's name and pointer star optional, so it
// resolves to *that* method and not to a same-named method on another type. A bare
// symbol is bounded on both sides, so `Lookup` no longer resolves to `LookupPath` —
// git takes the first line that matches, and an unbounded name matches inside a longer
// one first.
//
// **The dialect is git's, POSIX basic regex, not Go's.** Parentheses are literal, a
// group is `\(…\)`, an optional group is `\{0,1\}`, and `\b` is the GNU boundary. Each of
// those was probed against a real repository before it was relied on, which is why the
// tests for this function build one rather than asserting over strings.
//
// What it does not settle: a bare name's first bounded match may be a use or the doc
// comment above the declaration. Only an extent computed from a parse can, and that is
// the reclaim this is the stopgap for (gaps O10, O11).
func sitePattern(symbol string) string {
	if typ, method, ok := strings.Cut(symbol, "."); ok && typ != "" && method != "" {
		return `func (\([A-Za-z_][A-Za-z0-9_]* \)\{0,1\}\*\{0,1\}` + typ + `) ` + method + `\b`
	}
	return `\b` + symbol + `\b`
}

// declaresSymbol reports whether the file on disk still contains a declaration of the
// symbol. Deliberately loose — it asks "is this plausibly present" rather than parsing
// Go — because it is only ever used to choose between two *failure* reports, and the
// looseness errs toward "cannot tell", never toward a clean result.
func declaresSymbol(path, symbol string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		t := strings.TrimSpace(line)
		if !strings.HasPrefix(t, "func ") && !strings.HasPrefix(t, "var ") &&
			!strings.HasPrefix(t, "const ") && !strings.HasPrefix(t, "type ") {
			continue
		}
		// R206 — bounded, as the pattern handed to git is: `func LookupPath` on disk
		// does not declare `Lookup`, and saying it did turned a rotted anchor into
		// "no history yet" — found by the test for the bounded pattern, which fell
		// through to this fallback and read the wrong error.
		if declaresName(t, symbol) {
			return true
		}
		// R205 — a method's declaration line reads `func (x *Type) Method(`, never
		// `Type.Method`; without this a method written since the last commit reports
		// as a rotted anchor rather than as new.
		if typ, method, ok := strings.Cut(symbol, "."); ok &&
			strings.HasPrefix(t, "func (") && strings.Contains(t, typ+")") &&
			strings.Contains(t, ") "+method+"(") {
			return true
		}
	}
	return false
}

// declaresName reports whether a declaration line names symbol as a whole word.
func declaresName(line, symbol string) bool {
	at := 0
	for {
		i := strings.Index(line[at:], symbol)
		if i < 0 {
			return false
		}
		i += at
		before := i == 0 || !isIdent(line[i-1])
		after := i+len(symbol) == len(line) || !isIdent(line[i+len(symbol)])
		if before && after {
			return true
		}
		at = i + 1
	}
}

func isIdent(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

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
