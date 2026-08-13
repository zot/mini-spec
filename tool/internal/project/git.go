// CRC: crc-Git.md | Seq: seq-bootstrap.md | R167
package project

import (
	"errors"
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
	out, err := g.run("log", "-L", ":"+symbol+":"+file, "--format=%H|%ad", "--date=short")
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
		if strings.Contains(t, symbol) {
			return true
		}
	}
	return false
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

// run invokes git in the working directory. Shelling out rather than linking a library
// is deliberate: it gets tracked and ignored status for free and stays correct as git
// changes, at the stated cost that no other VCS is supported. R167
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	return string(out), err
}
