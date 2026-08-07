// CRC: crc-RepoRoot.md | Seq: seq-reporoot.md | R107
package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mkTree builds a temporary tree. Each entry is a path relative to the root;
// entries ending in "/" become directories, the rest become empty files.
// Returns the absolute root.
func mkTree(t *testing.T, entries ...string) string {
	t.Helper()
	root := t.TempDir()
	// Resolve symlinks so comparisons hold on platforms where TempDir is linked
	// (macOS /var -> /private/var); filepath.Abs alone would not.
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}
	for _, e := range entries {
		p := filepath.Join(root, strings.TrimSuffix(e, "/"))
		if strings.HasSuffix(e, "/") {
			if err := os.MkdirAll(p, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", p, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, nil, 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return root
}

// resolve runs the resolver from <root>/sub with the home boundary placed above the
// tree, so the exclusion is inert unless a test sets it deliberately.
func resolve(t *testing.T, root, sub string) (string, error) {
	t.Helper()
	return RepoRootFrom(filepath.Join(root, sub), filepath.Dir(root))
}

// mustResolve is resolve for the cases that expect success: only the boundary and
// failure tests have anything to say about the error, so the rest assert on the
// resolved root alone.
func mustResolve(t *testing.T, root, sub string) string {
	t.Helper()
	got, err := resolve(t, root, sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return got
}

// Each strong marker is sufficient on its own, and none is privileged. R109
func TestStrongMarkersEachResolve(t *testing.T) {
	for _, marker := range []string{".git/", ".minispec/", "carves/", "PENDING.md", "CURRENT.md", "DONE.md"} {
		t.Run(strings.TrimSuffix(marker, "/"), func(t *testing.T) {
			root := mkTree(t, "a/b/", marker)
			if got := mustResolve(t, root, "a/b"); got != root {
				t.Errorf("got %s, want %s", got, root)
			}
		})
	}
}

// A `.git` file — the linked-worktree shape — marks a root as a directory does. R114
func TestGitAsFileResolves(t *testing.T) {
	root := mkTree(t, "a/b/", ".git")
	if got := mustResolve(t, root, "a/b"); got != root {
		t.Errorf("got %s, want %s", got, root)
	}
}

// Coexisting strong markers agree, so no ordering among them is needed. R109
func TestCoexistingStrongMarkersAgree(t *testing.T) {
	root := mkTree(t, "a/b/", ".git/", ".minispec/", "carves/")
	if got := mustResolve(t, root, "a/b"); got != root {
		t.Errorf("got %s, want %s", got, root)
	}
}

// The property that makes collecting necessary: a first-match walk accepting the
// `.claude` at a/ would resolve a directory that is not a repository. R108
func TestStrongMarkerAboveBeatsWeakBelow(t *testing.T) {
	root := mkTree(t, "a/b/", ".git/", "a/.claude/")
	if got := mustResolve(t, root, "a/b"); got != root {
		t.Errorf("got %s, want %s (first-match walk would have taken a/)", got, root)
	}
}

// Among strong markers the deepest wins, because the walk returns on the first met. R109
func TestDeepestStrongMarkerWins(t *testing.T) {
	root := mkTree(t, "a/b/", ".git/", "a/carves/")
	want := filepath.Join(root, "a")
	if got := mustResolve(t, root, "a/b"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// `.claude` resolves only when nothing stronger exists anywhere above. R110
func TestClaudeFallback(t *testing.T) {
	root := mkTree(t, "a/b/", ".claude/")
	if got := mustResolve(t, root, "a/b"); got != root {
		t.Errorf("got %s, want %s", got, root)
	}
}

// The recorded candidate is the deepest `.claude`, not the last one seen. R110
func TestDeepestClaudeWins(t *testing.T) {
	root := mkTree(t, "a/b/", ".claude/", "a/.claude/")
	want := filepath.Join(root, "a")
	if got := mustResolve(t, root, "a/b"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// `.minispec.yaml` is the last resort, and ranks below `.claude` regardless of
// depth — presence is not a declaration. R111
func TestMinispecYamlIsLastResort(t *testing.T) {
	t.Run("alone", func(t *testing.T) {
		root := mkTree(t, "a/b/", ".minispec.yaml")
		if got := mustResolve(t, root, "a/b"); got != root {
			t.Errorf("got %s, want %s", got, root)
		}
	})
	t.Run("outranked by a shallower .claude", func(t *testing.T) {
		root := mkTree(t, "a/b/", ".claude/", "a/.minispec.yaml")
		if got := mustResolve(t, root, "a/b"); got != root {
			t.Errorf("got %s, want %s (.claude outranks .minispec.yaml at any depth)", got, root)
		}
	})
}

// Nothing at or above the home boundary is consulted, even a strong marker. R112
func TestHomeBoundaryIsNeverCrossed(t *testing.T) {
	root := mkTree(t, "a/b/", ".git/", ".claude/")
	_, err := RepoRootFrom(filepath.Join(root, "a", "b"), root)
	if err == nil {
		t.Fatal("expected failure: markers sit at the home boundary and must not be consulted")
	}
}

// Absence is reported as an error carrying the search origin, never as a bare
// failure a caller could mistake for a clean result. R113
func TestFailureNamesStartingDirectory(t *testing.T) {
	root := mkTree(t, "a/b/")
	start := filepath.Join(root, "a", "b")
	_, err := resolve(t, root, "a/b")
	if err == nil {
		t.Fatal("expected an error for a marker-free tree")
	}
	if !strings.Contains(err.Error(), start) {
		t.Errorf("error %q does not name the starting directory %s", err, start)
	}
}

// Only `.git` counts as a version-control marker; another VCS falls through. R114
func TestNonGitVCSFallsThrough(t *testing.T) {
	root := mkTree(t, "a/b/", ".claude/", "a/.fslckout")
	if got := mustResolve(t, root, "a/b"); got != root {
		t.Errorf("got %s, want %s (.fslckout must not mark a root)", got, root)
	}
}

// The two roots are resolved independently — the shape that motivated the split:
// one repository, several design roots beneath it. R107
func TestTwoRootsResolveIndependently(t *testing.T) {
	root := mkTree(t, ".git/", "carves/", "tool/design/", "example/design/")

	repo, err := resolve(t, root, "tool")
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	if repo != root {
		t.Errorf("repo root: got %s, want %s", repo, root)
	}

	p, err := DetectFrom(filepath.Join(root, "tool"))
	if err != nil {
		t.Fatalf("design root: %v", err)
	}
	want := filepath.Join(root, "tool")
	if p.RootPath != want {
		t.Errorf("design root: got %s, want %s", p.RootPath, want)
	}
}
