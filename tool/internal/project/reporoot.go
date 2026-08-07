// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md | R107
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// strongMarkers end the search outright: each is repository-level by mandate or by
// definition, so where several coexist they agree and no ordering among them is
// needed. Existence is checked rather than file type — a linked worktree or
// submodule carries `.git` as a regular file, and it marks a repository root just
// as a directory does. R109, R114
var strongMarkers = []string{
	".git",
	".minispec",
	"carves",
	"PENDING.md",
	"CURRENT.md",
	"DONE.md",
}

// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md#1 | R107
// RepoRoot resolves the repository root by walking up from the current directory,
// bounded by the user's home directory.
func RepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	return RepoRootFrom(cwd, home)
}

// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md#1.3 | R108-R113
// RepoRootFrom resolves the repository root from an explicit starting directory and
// home boundary. The boundary is a parameter rather than an environment lookup so
// the exclusion is testable on a temporary tree.
func RepoRootFrom(startDir, homeBoundary string) (string, error) {
	// step 1.1 — absolute, so parent traversal terminates predictably
	start, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	// step 1.2 — an empty boundary means "no exclusion" and is left empty:
	// filepath.Abs("") resolves to the current directory, which would install an
	// accidental boundary rather than none
	home := homeBoundary
	if home != "" {
		if abs, err := filepath.Abs(home); err == nil {
			home = abs
		}
	}

	// Weak markers are recorded rather than returned on sight: a strong marker may
	// still sit above, and `.claude` in particular exists at several levels of an
	// ordinary path.
	var claudeCandidate, yamlCandidate string

	// step 1.3 — the upward walk
	dir := start
	for {
		// step 1.3.1 — precedes every marker test, so nothing at or above home is
		// ever consulted
		if atBoundary(dir, home) {
			break
		}
		// step 1.3.2 — the first strong marker met going up is by definition the
		// deepest, so this returns instead of recording
		if hasStrongMarker(dir) {
			return dir, nil
		}
		// steps 1.3.3 and 1.3.4 — independent, not alternatives: one directory may
		// hold both a .claude and a .minispec.yaml
		if claudeCandidate == "" && isDir(filepath.Join(dir, ".claude")) {
			claudeCandidate = dir
		}
		if yamlCandidate == "" && exists(filepath.Join(dir, ".minispec.yaml")) {
			yamlCandidate = dir
		}
		// step 1.3.5
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// steps 1.4 and 1.5 — a weak candidate is promoted only by exhausting the walk
	if claudeCandidate != "" {
		return claudeCandidate, nil
	}
	if yamlCandidate != "" {
		return yamlCandidate, nil
	}
	// step 1.6 — absence is reported as an error carrying the search origin, never
	// as a bare failure a caller could mistake for a clean result
	return "", fmt.Errorf("no repository root found (searched from %s)", start)
}

// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md#1.3.1 | R112
// atBoundary reports whether dir is the home directory or an ancestor of it.
// ~/.claude exists on a normal installation, so without this any marker-free tree
// would resolve its repository root to the user's home directory — and anything the
// tool later created would be deposited there.
func atBoundary(dir, home string) bool {
	if home == "" {
		return false
	}
	rel, err := filepath.Rel(dir, home)
	if err != nil {
		return false
	}
	outside := rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
	return !outside
}

// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md#1.3.2 | R109, R114
// hasStrongMarker reports whether a directory carries any search-ending marker.
func hasStrongMarker(dir string) bool {
	for _, m := range strongMarkers {
		if exists(filepath.Join(dir, m)) {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
