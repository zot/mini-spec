// CRC: crc-Project.md | Seq: seq-config.md | R118, R119
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Origin labels for settings that no configuration file supplied.
const OriginDefaults = "(built-in defaults)"

// CRC: crc-Project.md | Seq: seq-config.md#2.4 | R129
// Origins records which layer supplied each effective setting, keyed by setting
// name — maps are keyed per entry (`comment_patterns[.go]`) because that is the
// granularity at which they merge.
type Origins map[string]string

// Setting is one resolved setting: its name, its effective value, and the file that
// supplied it. R129
type Setting struct {
	Name   string `json:"setting"`
	Value  string `json:"value"`
	Origin string `json:"origin"`
}

// CRC: crc-Project.md | Seq: seq-config.md#1.6 | R129
// EffectiveSettings lists every resolved setting with its value and the file that
// supplied it, sorted by name. Settings no layer set report the built-in defaults, so
// the answer to "why is this value what it is" never requires reading two files and
// knowing the precedence by heart.
func (p *Project) EffectiveSettings() []Setting {
	var out []Setting
	add := func(name, value string) {
		origin := p.Origins[name]
		if origin == "" {
			origin = OriginDefaults
		}
		out = append(out, Setting{Name: name, Value: value, Origin: origin})
	}
	// A map setting is listed per key, the granularity at which it merges.
	addMap := func(setting string, values map[string]string) {
		keys := make([]string, 0, len(values))
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			add(fmt.Sprintf("%s[%s]", setting, k), values[k])
		}
	}

	add("design_dir", p.Config.DesignDir)
	add("src_dir", p.Config.SrcDir)
	add("code_extensions", strings.Join(p.Config.CodeExtensions, ", "))
	addMap("comment_patterns", p.Config.CommentPatterns)
	addMap("comment_closers", p.Config.CommentClosers)
	return out
}

// ConfigDirName is the tool-managed directory at the repository root. It holds the
// repository configuration and the tool's machine-local working files, which is why
// the repository config lives here rather than beside `.git`. R119
const ConfigDirName = ".minispec"

// RepoConfigName is the repository configuration inside ConfigDirName. R118
const RepoConfigName = "config.yaml"

// DesignConfigName is a design root's own configuration. R34
const DesignConfigName = ".minispec.yaml"

// repoConfigPath is where the repository configuration sits under a repository root.
func repoConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, ConfigDirName, RepoConfigName)
}

// CRC: crc-Project.md | Seq: seq-config.md#1.1 | R118
// resolveConfig produces the effective configuration for a design root, locating the
// repository layer itself.
func resolveConfig(designRoot string) (Config, Origins, error) {
	// step 1.1 — a tree with no markers has no repository layer, which is the
	// behavior that existed before repository configuration did
	repoRoot, hasRepo := repoRootFor(designRoot)
	return resolveConfigFrom(designRoot, repoRoot, hasRepo)
}

// CRC: crc-Project.md | Seq: seq-config.md#1 | R120-R124
// resolveConfigFrom is the resolution proper, with the repository layer supplied
// rather than discovered. Separated for the same reason RepoRootFrom is: a test can
// then state the whole world it is resolving against instead of inheriting the
// machine's.
func resolveConfigFrom(designRoot, repoRoot string, hasRepo bool) (Config, Origins, error) {
	// step 1.2 — before reading anything, so the error is about the file's presence
	// and never about its contents
	if hasRepo {
		if err := rejectRepoRootDesignConfig(repoRoot); err != nil {
			return Config{}, nil, err
		}
	}

	// step 1.3
	cfg := DefaultConfig()
	origins := Origins{}

	// step 1.4
	if hasRepo {
		if err := applyLayerFile(&cfg, origins, repoConfigPath(repoRoot)); err != nil {
			return Config{}, nil, err
		}
	}

	// step 1.5 — unconditional. Where the repository root and the design root are the
	// same directory, this file *is* the one step 1.2 rejected, so there is no
	// "unless the roots coincide" case to write.
	if err := applyLayerFile(&cfg, origins, filepath.Join(designRoot, DesignConfigName)); err != nil {
		return Config{}, nil, err
	}

	return cfg, origins, nil
}

// CRC: crc-Project.md | Seq: seq-config.md#1.2 | R122, R123
// rejectRepoRootDesignConfig fails when a `.minispec.yaml` sits at the repository
// root. It would have to inherit from `.minispec/config.yaml` — a file inside its own
// directory — so the shape is incoherent in every layout, including the one where the
// repository root is also a design root. Lstat rather than Stat: a symlink to the new
// location is still the forbidden shape, not a supported alias.
func rejectRepoRootDesignConfig(repoRoot string) error {
	path := filepath.Join(repoRoot, DesignConfigName)
	if _, err := os.Lstat(path); err != nil {
		return nil
	}
	repoCfg := repoConfigPath(repoRoot)
	return fmt.Errorf(
		"%s is not a valid location for a mini-spec config.\n"+
			"The repository configuration lives at %s.\n"+
			"Remove %s (a symlink counts) and put any settings it holds in %s.",
		path, repoCfg, path, repoCfg,
	)
}

// applyLayerFile reads one configuration layer and applies it. A missing file is not
// an error — a layer a project does not use simply contributes nothing.
func applyLayerFile(cfg *Config, origins Origins, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var layer Config
	if err := yaml.Unmarshal(data, &layer); err != nil {
		return fmt.Errorf("invalid config %s: %w", path, err)
	}
	applyLayer(cfg, layer, path, origins)
	return nil
}

// CRC: crc-Project.md | Seq: seq-config.md#2 | R125-R128, R130
// applyLayer applies one layer over what is already resolved. The three rules are one
// rule seen through three types: a layer states only what it adds or changes, never
// what it keeps. A scalar cannot merge, so replacement is the only form that rule can
// take for it — which is also why a list can be added to but never trimmed. Removing
// an inherited entry means dropping the setting from the layer above and stating it
// here instead.
func applyLayer(cfg *Config, layer Config, origin string, origins Origins) {
	// step 2.1
	if layer.DesignDir != "" {
		cfg.DesignDir = layer.DesignDir
		origins["design_dir"] = origin
	}
	if layer.SrcDir != "" {
		cfg.SrcDir = layer.SrcDir
		origins["src_dir"] = origin
	}
	// step 2.3 — union between configuration layers, preserving inherited order so the
	// result is deterministic and a reader can see which entries came from where.
	//
	// The first configuration layer *replaces* the built-in defaults rather than
	// adding to them, and that exception is load-bearing. Defaults are not a layer
	// anyone authored, so "state only what you add" cannot apply to them — you cannot
	// have added to a list you never wrote. Without the exception the shipped
	// `code_extensions` would be permanently un-narrowable: unioning onto it can only
	// ever grow it, and the remedy for an unwanted entry (drop the setting from the
	// layer above and state it here) has nowhere to reach, since nothing sits below
	// the defaults. Whether a config layer has spoken yet is exactly what `origins`
	// already records.
	if len(layer.CodeExtensions) > 0 {
		inherited := cfg.CodeExtensions
		if _, spokenFor := origins["code_extensions"]; !spokenFor {
			// nothing but the built-in defaults sit below: replace rather than union
			inherited = nil
		}
		cfg.CodeExtensions = union(inherited, layer.CodeExtensions)
		origins["code_extensions"] = origin
	}
	// step 2.2 — per key, so a design root adding one entry keeps every other
	mergeMap(cfg.CommentPatterns, layer.CommentPatterns, "comment_patterns", origin, origins)
	mergeMap(cfg.CommentClosers, layer.CommentClosers, "comment_closers", origin, origins)
}

// mergeMap merges one map setting per key, recording an origin per entry. R126
func mergeMap(dst, src map[string]string, setting, origin string, origins Origins) {
	for k, v := range src {
		dst[k] = v
		origins[fmt.Sprintf("%s[%s]", setting, k)] = origin
	}
}

// union appends entries not already present, keeping inherited order. R127
func union(inherited, added []string) []string {
	seen := make(map[string]bool, len(inherited))
	out := make([]string, 0, len(inherited)+len(added))
	for _, v := range inherited {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	for _, v := range added {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

// repoRootFor locates the repository root for a design root. Searching from the
// design root rather than the current directory is deliberate: the design root was
// itself found by walking up, so it is the stable anchor, and the repository root is
// at or above it either way. R118
func repoRootFor(designRoot string) (string, bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	root, err := RepoRootFrom(designRoot, home)
	if err != nil {
		return "", false
	}
	return root, true
}
