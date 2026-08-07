// CRC: crc-Project.md | Seq: seq-config.md | R118
package project

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeFile writes content at a path inside a tree, creating parent directories.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// repoConfig writes the repository configuration inside a tree and returns its path,
// so a test asserting on provenance names the layer once rather than rebuilding it.
func repoConfig(t *testing.T, root, content string) string {
	t.Helper()
	path := repoConfigPath(root)
	writeFile(t, path, content)
	return path
}

// designConfig writes a design root's own configuration at <root>/<sub> and returns
// its path. An empty sub puts it at the root itself — the shape resolution rejects.
func designConfig(t *testing.T, root, sub, content string) string {
	t.Helper()
	path := filepath.Join(root, sub, DesignConfigName)
	writeFile(t, path, content)
	return path
}

// resolveIn resolves a design root against an explicit repository root, so no test
// depends on what sits above the temporary tree.
func resolveIn(t *testing.T, root, sub string) (Config, Origins, error) {
	t.Helper()
	return resolveConfigFrom(filepath.Join(root, sub), root, true)
}

// mustResolveIn is resolveIn for the cases that expect success: only the rejection
// tests have anything to say about the error, so the rest assert on the resolved
// configuration alone.
func mustResolveIn(t *testing.T, root, sub string) (Config, Origins) {
	t.Helper()
	cfg, origins, err := resolveIn(t, root, sub)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return cfg, origins
}

// mustResolveNoRepo resolves a design root with no repository layer above it — the
// pre-existing single-project shape, where the design root's file applies straight
// over the built-in defaults.
func mustResolveNoRepo(t *testing.T, designRoot string) (Config, Origins) {
	t.Helper()
	cfg, origins, err := resolveConfigFrom(designRoot, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return cfg, origins
}

// mustRejectIn resolves expecting the repository-root rejection, checks that the error
// names the forbidden path, and returns it so a test with more to ask can go on asking.
func mustRejectIn(t *testing.T, root, sub, forbidden string) error {
	t.Helper()
	_, _, err := resolveIn(t, root, sub)
	if err == nil {
		t.Fatalf("expected a rejection naming %s", forbidden)
	}
	if !strings.Contains(err.Error(), forbidden) {
		t.Errorf("error %q does not name the forbidden path %s", err, forbidden)
	}
	return err
}

// The repository configuration reaches a design root beneath it. R118, R121
func TestRepoConfigAppliesToDesignRootBelow(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "src_dir: lib\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	if cfg.SrcDir != "lib" {
		t.Errorf("src_dir = %q, want %q", cfg.SrcDir, "lib")
	}
}

// A design root's own file overrides the repository layer. R124, R125
func TestDesignRootOverridesRepo(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "src_dir: lib\n")
	proj := designConfig(t, root, "tool", "src_dir: source\n")

	cfg, origins := mustResolveIn(t, root, "tool")
	if cfg.SrcDir != "source" {
		t.Errorf("src_dir = %q, want %q", cfg.SrcDir, "source")
	}
	if origins["src_dir"] != proj {
		t.Errorf("origin = %q, want %q", origins["src_dir"], proj)
	}
}

// A design root whose settings match the repository needs no file at all. R121
func TestDesignRootWithNoFileInheritsEverything(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "src_dir: lib\ncomment_patterns:\n  .zig: \"//\\\\s*\"\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	if cfg.SrcDir != "lib" {
		t.Errorf("src_dir = %q, want %q", cfg.SrcDir, "lib")
	}
	if cfg.CommentPatterns[".zig"] == "" {
		t.Error("inherited comment_patterns entry missing")
	}
}

// Maps merge per key, so a project adds one pattern without restating the rest. R126
func TestMapsMergePerKey(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repo := repoConfig(t, root, "comment_patterns:\n  .zig: \"ZIG\"\n  .nim: \"NIM\"\n")
	designConfig(t, root, "tool", "comment_patterns:\n  .odin: \"ODIN\"\n")

	cfg, origins := mustResolveIn(t, root, "tool")
	for ext, want := range map[string]string{".zig": "ZIG", ".nim": "NIM", ".odin": "ODIN"} {
		if got := cfg.CommentPatterns[ext]; got != want {
			t.Errorf("comment_patterns[%s] = %q, want %q", ext, got, want)
		}
	}
	if origins["comment_patterns[.zig]"] != repo {
		t.Errorf("inherited key lost its origin: %q", origins["comment_patterns[.zig]"])
	}
}

// Merging per key still overrides at the key. R126
func TestMapKeySetByBothTakesDesignRoot(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "comment_patterns:\n  .zig: \"REPO\"\n  .nim: \"NIM\"\n")
	designConfig(t, root, "tool", "comment_patterns:\n  .zig: \"PROJECT\"\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	if cfg.CommentPatterns[".zig"] != "PROJECT" {
		t.Errorf("comment_patterns[.zig] = %q, want PROJECT", cfg.CommentPatterns[".zig"])
	}
	if cfg.CommentPatterns[".nim"] != "NIM" {
		t.Errorf("untouched key changed: %q", cfg.CommentPatterns[".nim"])
	}
}

// Lists union rather than replace, so a project states additions. R127
func TestListsUnion(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "code_extensions: [.go, .ts]\n")
	designConfig(t, root, "tool", "code_extensions: [.lua]\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	want := []string{".go", ".ts", ".lua"}
	if !slices.Equal(cfg.CodeExtensions, want) {
		t.Errorf("code_extensions = %v, want %v (inherited order, addition appended)", cfg.CodeExtensions, want)
	}
}

// Repeating an inherited entry is harmless rather than doubling it. R127
func TestListDuplicatesDropped(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "code_extensions: [.go, .ts]\n")
	designConfig(t, root, "tool", "code_extensions: [.ts, .lua]\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	want := []string{".go", ".ts", ".lua"}
	if !slices.Equal(cfg.CodeExtensions, want) {
		t.Errorf("code_extensions = %v, want %v (.ts once, in its inherited position)", cfg.CodeExtensions, want)
	}
}

// The flat rule, on the shape it exists to forbid. R122
func TestRepoRootDesignConfigRejected(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "src_dir: lib\n")
	forbidden := designConfig(t, root, "", "src_dir: other\n")

	mustRejectIn(t, root, "tool", forbidden)
}

// A link to the new location is the forbidden shape, not a supported alias — the
// exact transition scaffolding left in the reference project. R123
func TestRejectionCatchesSymlink(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "src_dir: lib\n")
	link := filepath.Join(root, DesignConfigName)
	if err := os.Symlink(filepath.Join(ConfigDirName, RepoConfigName), link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	mustRejectIn(t, root, "tool", link)
}

// The error is about presence, not parseability, because rejection precedes reading. R122
func TestRejectionPrecedesReading(t *testing.T) {
	root := mkTree(t, "tool/design/")
	forbidden := designConfig(t, root, "", "this: [is not: valid yaml\n")

	err := mustRejectIn(t, root, "tool", forbidden)
	if strings.Contains(err.Error(), "invalid config") {
		t.Errorf("reported a parse failure, not the forbidden path: %q", err)
	}
}

// The "no exception" half, on the layout where an exception would be most expected:
// repository root and design root are the same directory. R120, R122
func TestRejectionAppliesWhenRootsCoincide(t *testing.T) {
	root := mkTree(t, "design/")
	repoConfig(t, root, "src_dir: lib\n")
	designConfig(t, root, "", "src_dir: other\n")

	_, _, err := resolveConfigFrom(root, root, true)
	if err == nil {
		t.Fatal("expected an error: the coincident-roots layout is not privileged")
	}
}

// Pre-existing single-project behavior survives: no repository root, no repository
// layer, and the design root's own file applies over the defaults. R124
func TestNoRepoRootMeansNoRepoLayer(t *testing.T) {
	root := mkTree(t, "design/")
	designConfig(t, root, "", "src_dir: source\n")

	cfg, origins := mustResolveNoRepo(t, root)
	if cfg.SrcDir != "source" {
		t.Errorf("src_dir = %q, want source", cfg.SrcDir)
	}
	if cfg.DesignDir != DefaultConfig().DesignDir {
		t.Errorf("design_dir = %q, want the built-in default", cfg.DesignDir)
	}
	if origins["design_dir"] != "" {
		t.Errorf("a setting no layer supplied claims an origin: %q", origins["design_dir"])
	}
}

// A value's origin is answerable without reading two files. R129
func TestProvenanceNamesTheSourceFile(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repo := repoConfig(t, root, "src_dir: lib\ncomment_patterns:\n  .zig: \"ZIG\"\n")
	proj := designConfig(t, root, "tool", "design_dir: spec\ncomment_patterns:\n  .odin: \"ODIN\"\n")

	_, origins := mustResolveIn(t, root, "tool")
	for setting, want := range map[string]string{
		"src_dir":                 repo,
		"comment_patterns[.zig]":  repo,
		"design_dir":              proj,
		"comment_patterns[.odin]": proj,
	} {
		if origins[setting] != want {
			t.Errorf("origin of %s = %q, want %q", setting, origins[setting], want)
		}
	}
}

// The first configuration layer replaces the built-in defaults rather than adding to
// them, so a project can still narrow the extension list. Without this, the shipped
// defaults would be permanently un-narrowable — union can only grow a list, and
// nothing sits below the defaults to move a setting down to. R127
func TestFirstLayerReplacesDefaults(t *testing.T) {
	root := mkTree(t, "tool/design/")
	repoConfig(t, root, "code_extensions: [.go]\n")

	cfg, _ := mustResolveIn(t, root, "tool")
	if want := []string{".go"}; !slices.Equal(cfg.CodeExtensions, want) {
		t.Errorf("code_extensions = %v, want %v — defaults must be replaced, not added to", cfg.CodeExtensions, want)
	}
}

// A design root alone, with no repository layer, also replaces the defaults. R127
func TestDesignRootAloneReplacesDefaults(t *testing.T) {
	root := mkTree(t, "design/")
	designConfig(t, root, "", "code_extensions: [.lua]\n")

	cfg, _ := mustResolveNoRepo(t, root)
	if want := []string{".lua"}; !slices.Equal(cfg.CodeExtensions, want) {
		t.Errorf("code_extensions = %v, want %v", cfg.CodeExtensions, want)
	}
}
