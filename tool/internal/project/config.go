// CRC: crc-Project.md | Seq: seq-config.md | R118, R119
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/simple-dom/sdom"
)

// Origin labels for settings that no configuration file supplied.
const OriginDefaults = "(built-in defaults)"

// CRC: crc-Project.md | Seq: seq-config.md#2.4 | R129
// Origins records which layer supplied each effective setting, keyed by setting
// name.
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
	if p.Config.Track != "" {
		add("track", p.Config.Track)
	}
	add("design_dir", p.Config.DesignDir)
	add("src_dir", p.Config.SrcDir)
	add("code_extensions", strings.Join(p.Config.CodeExtensions, ", "))
	for _, def := range p.Config.Languages {
		add("languages["+def.Name+"]", strings.Join(def.Extensions, ", "))
	}
	return out
}

// ConfigDirName is the tool-managed directory at the repository root. It holds the
// repository configuration and the tool's machine-local working files, which is why
// the repository config lives here rather than beside `.git`. R119
const ConfigDirName = ".minispec"

// RepoConfigName is the repository configuration inside ConfigDirName. R118, R520
const RepoConfigName = "config.toml"

// DesignConfigName is a design root's own configuration. R34, R520
const DesignConfigName = ".minispec.toml"

// The configuration's names before it moved to TOML, kept only so a file still under
// one of them is reported rather than passed over as no configuration at all. R522
const (
	LegacyRepoConfigName   = "config.yaml"
	LegacyDesignConfigName = ".minispec.yaml"
)

// retiredKeys are keys a configuration once set and the tool no longer reads, with
// what replaced each. R521
var retiredKeys = map[string]string{
	"comment_patterns": "code files are read through a language table chosen by extension",
	"comment_closers":  "code files are read through a language table chosen by extension",
}

// CRC: crc-Project.md | Seq: seq-config.md#1.2.1 | R522
// LegacyConfigError names the first configuration still in YAML among the places one
// would be read — the repository's `.minispec/config.yaml`, a `.minispec.yaml` at the
// repository root, and a design root's — or returns nil. Nothing reads YAML, so the
// tool never guesses at an old file's meaning: the person converts it by hand, and the
// converted file is read under the same strict rules as any other. Either root may be
// empty.
func LegacyConfigError(repoRoot, designRoot string) error {
	var candidates []string
	if repoRoot != "" {
		candidates = append(candidates,
			filepath.Join(repoRoot, ConfigDirName, LegacyRepoConfigName),
			filepath.Join(repoRoot, LegacyDesignConfigName))
	}
	if designRoot != "" {
		candidates = append(candidates, filepath.Join(designRoot, LegacyDesignConfigName))
	}
	for _, path := range candidates {
		if _, err := os.Lstat(path); err != nil {
			continue
		}
		converted := strings.TrimSuffix(path, ".yaml") + ".toml"
		return fmt.Errorf(
			"%s is a YAML configuration, and mini-spec configuration is TOML as of 2026-09-25.\n"+
				"Nothing reads YAML any more, so this file is not being applied.\n"+
				"Convert it by hand to %s and remove the YAML file; the keys keep their names,\n"+
				"and the format is documented in .claude/skills/mini-spec/config-reference.md.",
			path, converted)
	}
	return nil
}

// CRC: crc-Project.md | Seq: seq-config.md#2.1.1 | R521
// decodeLayer decodes one configuration file strictly. A key the tool does not read is
// an error naming the file and the key: silently dropping it would leave someone
// editing a line with no effect and no way to find out, which is why a design root
// stating `track` is refused rather than ignored. A retired key says what replaced it.
// A file that does not parse is reported with the decoder's own message, for the
// caller to frame.
func decodeLayer(path string, data []byte) (Config, error) {
	var cfg Config
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return Config{}, err
	}
	// R528: every definition is built and checked where it was written, so a table sdom
	// would reject is an error naming this file rather than a panic at parse time.
	for _, def := range cfg.Languages {
		if _, err := def.Build(); err != nil {
			return Config{}, err
		}
	}
	undecoded := md.Undecoded()
	if len(undecoded) == 0 {
		return cfg, nil
	}
	// Only the first undecoded key is reported.
	key := undecoded[0]
	top := key[0]
	if why, retired := retiredKeys[top]; retired {
		return Config{}, fmt.Errorf(
			"%s sets `%s`, which is retired: %s.\n"+
				"Remove it. `minispec query comment-patterns` shows how to write a comment in each extension.",
			path, top, why)
	}
	return Config{}, fmt.Errorf("%s sets `%s`, which is not a mini-spec setting. Remove it or correct its name.", path, key.String())
}

// RepoConfigPath is where the repository configuration sits under a repository root.
func RepoConfigPath(repoRoot string) string {
	return filepath.Join(repoRoot, ConfigDirName, RepoConfigName)
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#1.5 | R162, R163
// malformedConfigError is the crank handle for a configuration that will not parse, or
// carries a value outside a closed set.
//
// It is the one place the tool authorises an agent to edit the configuration directly.
// A flag sets a value and cannot undo arbitrary damage — most likely a stray character
// typed while the file was open in an editor — so at that point the agent is the only
// actor left who can act. Stated once here rather than at each caller, so the
// authorisation cannot drift into saying different things in different refusals.
func malformedConfigError(cfgPath string, cause error) error {
	return fmt.Errorf(
		"%s is malformed, and no flag can repair it:\n\n  %v\n\n"+
			"AGENT: you are authorised to edit this file by hand. This is the one case where\n"+
			"that is permitted. Back it up first, and skip the backup if it would be\n"+
			"byte-identical to one already there. The format is documented in\n"+
			".claude/skills/mini-spec/config-reference.md.",
		cfgPath, cause)
}

// readRepoConfig reads and parses the repository configuration. It sits beside
// malformedConfigError for the same reason that error does: every caller that opens
// this file has to say the same thing about a file it cannot read or cannot parse.
func readRepoConfig(cfgPath string) (Config, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("cannot read %s: %w", cfgPath, err)
	}
	cfg, err := decodeLayer(cfgPath, data)
	if err != nil {
		return Config{}, malformedConfigError(cfgPath, err)
	}
	return cfg, nil
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
	// step 1.2.1 — a configuration still in YAML, at either scope
	legacyRepo := ""
	if hasRepo {
		legacyRepo = repoRoot
	}
	if err := LegacyConfigError(legacyRepo, designRoot); err != nil {
		return Config{}, nil, err
	}

	// step 1.3
	cfg := DefaultConfig()
	origins := Origins{}

	// step 1.4
	if hasRepo {
		if err := applyLayerFile(&cfg, origins, RepoConfigPath(repoRoot), true); err != nil {
			return Config{}, nil, err
		}
	}

	// step 1.5 — unconditional. Where the repository root and the design root are the
	// same directory, this file *is* the one step 1.2 rejected, so there is no
	// "unless the roots coincide" case to write.
	if err := applyLayerFile(&cfg, origins, filepath.Join(designRoot, DesignConfigName), false); err != nil {
		return Config{}, nil, err
	}

	return cfg, origins, nil
}

// CRC: crc-Project.md | Seq: seq-config.md#1.2 | R122, R123
// rejectRepoRootDesignConfig fails when a `.minispec.toml` sits at the repository
// root. It would have to inherit from `.minispec/config.toml` — a file inside its own
// directory — so the shape is incoherent in every layout, including the one where the
// repository root is also a design root. Lstat rather than Stat: a symlink to the new
// location is still the forbidden shape, not a supported alias.
func rejectRepoRootDesignConfig(repoRoot string) error {
	path := filepath.Join(repoRoot, DesignConfigName)
	if _, err := os.Lstat(path); err != nil {
		return nil
	}
	repoCfg := RepoConfigPath(repoRoot)
	return fmt.Errorf(
		"%s is not a valid location for a mini-spec config.\n"+
			"The repository configuration lives at %s.\n"+
			"Remove %s (a symlink counts) and put any settings it holds in %s.",
		path, repoCfg, path, repoCfg,
	)
}

// CRC: crc-Project.md | Seq: seq-config.md#1.4 | R135
// applyLayerFile reads one configuration layer and applies it. A missing file is not
// an error — a layer a project does not use simply contributes nothing.
//
// isRepoLayer gates the one repository-scoped setting. A design root stating `track`
// is refused rather than ignored: silently dropping it would leave someone editing a
// line that has no effect and no way to discover that, which is the failure this
// project treats as worse than an error.
func applyLayerFile(cfg *Config, origins Origins, path string, isRepoLayer bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// step 2.1.1
	layer, err := decodeLayer(path, data)
	if err != nil {
		return fmt.Errorf("invalid config %s: %w", path, err)
	}
	if layer.Track != "" && !isRepoLayer {
		return fmt.Errorf(
			"%s sets `track`, which is repository-scoped and belongs only in %s.\n"+
				"It describes the repository, and a repository may hold several design roots,\n"+
				"so one of them cannot answer for the whole. Remove it here.",
			path, ConfigDirName+"/"+RepoConfigName)
	}
	applyLayer(cfg, layer, path, origins)
	return nil
}

// CRC: crc-Project.md | Seq: seq-config.md#2 | R125, R127, R128, R130
// applyLayer applies one layer over what is already resolved. The three rules are one
// rule seen through three types: a layer states only what it adds or changes, never
// what it keeps. A scalar cannot merge, so replacement is the only form that rule can
// take for it — which is also why a list can be added to but never trimmed. Removing
// an inherited entry means dropping the setting from the layer above and stating it
// here instead.
func applyLayer(cfg *Config, layer Config, origin string, origins Origins) {
	// step 2.1 — track is a scalar like any other, but only the repository layer can
	// have supplied one: applyLayerFile refuses it from anywhere else. R135
	if layer.Track != "" {
		cfg.Track = layer.Track
		origins["track"] = origin
	}
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
	// step 2.2 retired with R126: no map settings remain.

	// step 2.5 — a definition is replaced whole by name, or added under a new one; never
	// merged field by field, since half of one table and half of another is neither. R527
	for _, def := range layer.Languages {
		i := slices.IndexFunc(cfg.Languages, func(d minispecsdom.LanguageDef) bool { return d.Name == def.Name })
		if i >= 0 {
			cfg.Languages[i] = def
		} else {
			cfg.Languages = append(cfg.Languages, def)
		}
		origins["languages["+def.Name+"]"] = origin
	}
}

// CRC: crc-Project.md | R527, R555, R556
// Languages is the extension map and the `files` rules the project's configured languages
// make, for the harvest to consult before the built-in tables. Each definition was checked when its file loaded,
// so building it again cannot fail on a configuration that resolved.
func (p *Project) Languages() (minispecsdom.Configured, error) {
	out := minispecsdom.Configured{Ext: map[string]*sdom.BracketLang{}}
	for _, def := range p.Config.Languages {
		lang, err := def.Build()
		if err != nil {
			return minispecsdom.Configured{}, err
		}
		for _, ext := range def.Extensions {
			out.Ext[ext] = lang
		}
		for _, pat := range def.Files { // R555, R556 — in configuration order
			out.Files = append(out.Files, minispecsdom.FileRule{Pattern: pat, Lang: lang})
		}
	}
	return out, nil
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
