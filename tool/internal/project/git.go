// CRC: crc-Git.md | Seq: seq-bootstrap.md | R167
package project

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNoGit reports that no ignore or tracking state could be determined — there is no
// usable `git`, or the directory is not inside a working tree. Returned rather than a
// map of falses because a caller can mistake "none of these are ignored" for a
// verified answer, and cannot mistake an error for one. R168
var ErrNoGit = errors.New("not a git working tree: ignore state cannot be determined")

// CRC: crc-Git.md | R167
// GitFacts is what a consistency check needs to know about a working tree. It exists
// as an interface so Track can be verified against a stated world instead of the
// machine's: the seam is what makes the check testable without a repository.
type GitFacts interface {
	IsRepo() bool
	Ignored(paths []string) (map[string]bool, error)
	Tracked(path string) (bool, error)
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

// run invokes git in the working directory. Shelling out rather than linking a library
// is deliberate: it gets tracked and ignored status for free and stays correct as git
// changes, at the stated cost that no other VCS is supported. R167
func (g *Git) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.workDir
	out, err := cmd.Output()
	return string(out), err
}
