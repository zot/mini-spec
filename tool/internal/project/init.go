// CRC: crc-Init.md | Seq: seq-bootstrap.md | R137
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// GitIgnoreName is the top-level ignore file init writes. Editing it is a file edit,
// not a change to git's state, which is what keeps this consistent with the rule that
// the tool never manipulates the repository. R166
const GitIgnoreName = ".gitignore"

// CRC: crc-Init.md | R136, R144
// InitOptions is one invocation of the creation or repair verb. Track has no zero value
// that means anything: the flag is mandatory precisely because the value cannot be
// inferred, and a default would make the choice for someone without their noticing.
type InitOptions struct {
	RepoRoot string
	Track    TrackValue
	Repair   bool
}

// InitResult records what happened to which file, so the report can be complete and
// can distinguish an edit from a no-op. R141
type InitResult struct {
	Created   []string
	Edited    []string
	Unchanged []string
	// Unresolved names paths the chosen value wants public that some rule outside the
	// top-level .gitignore still ignores. Reported rather than swallowed: a repair that
	// half-worked and said nothing is the silent-partial-success this tool exists to
	// prevent. R172
	Unresolved []string
}

// record files one written path as an edit or a creation, so both writers below draw
// the distinction the same way. R141
func (r *InitResult) record(path string, existed bool) {
	if existed {
		r.Edited = append(r.Edited, path)
	} else {
		r.Created = append(r.Created, path)
	}
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#2 | R137, R138, R142, R144, R145, R161
// RunInit creates a repository configuration, or repairs an existing one.
//
// The two forms are mutually exclusive on one precondition, inverted rather than
// supplemented: plain init requires `.minispec/` to be absent, `--repair` requires it
// to be present. Because the precondition is inverted, neither form ever has to guess
// which the caller meant, and no mode flag has to be reconciled.
func RunInit(opts InitOptions, g GitFacts) (*InitResult, error) {
	if opts.Track == "" {
		// step 2.1 — refused rather than defaulted; see InitOptions.
		return nil, fmt.Errorf(
			"minispec init requires one of --track-none, --track-private-trajectory, or --track-all.\n" +
				"The value cannot be inferred: whether a repository is git-managed is checkable,\n" +
				"but whether you want your work queue to ship with it is not.")
	}
	cfgDir := filepath.Join(opts.RepoRoot, ConfigDirName)
	cfgPath := RepoConfigPath(opts.RepoRoot)
	_, err := os.Stat(cfgDir)
	exists := err == nil

	if opts.Repair {
		// step 3.2
		if !exists {
			return nil, fmt.Errorf(
				"%s does not exist, so there is nothing to repair.\n"+
					"Create it with: minispec init --track-%s", cfgDir, opts.Track)
		}
		// step 3.3 — a flag sets a value; it cannot undo arbitrary damage, so a config
		// that will not parse is refused here and repaired by hand.
		if err := validateWellFormed(cfgPath); err != nil {
			return nil, err
		}
	} else if exists {
		// step 2.3
		return nil, fmt.Errorf(
			"%s already exists, and plain init would be stomping on it.\n"+
				"To change the track value, run:\n\n"+
				"    minispec init --track-<none|private-trajectory|all> --repair\n\n"+
				"The two directions mean opposite things — confirm the value with the user first.",
			cfgDir)
	}

	result := &InitResult{}

	// steps 2.4 and 3.4
	if err := writeConfig(cfgPath, opts.Track, exists, result); err != nil {
		return nil, err
	}

	// steps 2.5 and 3.5 — a project with no repository is left alone rather than told
	// about a file it does not have.
	if g.IsRepo() {
		if err := reconcileIgnore(opts, g, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#2.7 | R171
// ignoreLine renders a governed path as an **anchored** .gitignore pattern.
//
// Anchored because every path `track` governs is mandated at the repository root. A
// bare `PENDING.md` also ignores `docs/PENDING.md` and any nested `PENDING.md`, which
// is broader than intended — and broader than what a project that wrote `/PENDING.md`
// by hand asked for. Since the last matching pattern wins, writing the bare form into
// a file that already had the anchored one would silently widen the project's rule.
func ignoreLine(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return "/" + p
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#3.3 | R161, R162
// validateWellFormed is --repair's precondition. Its failure is the one place the tool
// authorises an agent to edit the configuration directly: the damage is arbitrary —
// most likely a stray character typed while the file was open in an editor — and at
// that point the agent is the only actor left who can act.
func validateWellFormed(cfgPath string) error {
	probe, err := readRepoConfig(cfgPath)
	if err != nil {
		return err
	}
	if probe.Track == "" {
		return nil
	}
	if _, err := ParseTrack(probe.Track); err != nil {
		return malformedConfigError(cfgPath, err)
	}
	return nil
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#2.4 | R137, R138
// writeConfig sets track, preserving every other setting an existing file holds. It
// writes a basic file rather than a commented template of every available setting:
// a layer states only what it means, which is the same rule the inheritance model asks
// of every configuration.
func writeConfig(cfgPath string, track TrackValue, exists bool, result *InitResult) error {
	var cfg Config
	if exists {
		if data, err := os.ReadFile(cfgPath); err == nil {
			_ = yaml.Unmarshal(data, &cfg)
		}
	}
	if cfg.Track == string(track) {
		result.Unchanged = append(result.Unchanged, cfgPath)
		return nil
	}
	cfg.Track = string(track)
	out, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return err
	}
	result.record(cfgPath, exists)
	return nil
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#2.6 | R139, R140, R144, R170, R172
// reconcileIgnore brings the top-level .gitignore into agreement with the track value,
// in both directions.
//
// Both directions are what "symmetric" means, and it is the whole reason one command
// can resolve a mismatch either way: a one-directional repair could only ever make
// files more private, so a project that decided its queue should ship would have no
// path back.
func reconcileIgnore(opts InitOptions, g GitFacts, result *InitResult) error {
	mustIgnore, mustNotIgnore := opts.Track.Expectations(opts.RepoRoot)
	// The backup directory is ignored under every git-managed value: one line then
	// covers everything machine-local.
	want := append([]string{BackupIgnorePath}, mustIgnore...)

	// **Ask git what is already ignored rather than matching lines textually.** An
	// existing `/PENDING.md` and a proposed `PENDING.md` are the same intent written
	// two ways, and only git knows that. Matching text instead appends a duplicate —
	// and because the last matching pattern wins, the duplicate silently *replaces*
	// the project's narrower rule with a broader one. Measured on this repository the
	// first time init ran on it. Same reason the ignore check shells out rather than
	// parsing this file: ignore rules compose in ways a naive reader gets wrong.
	already, err := g.Ignored(append(slices.Clone(want), mustNotIgnore...))
	if err != nil {
		return err
	}

	path := filepath.Join(opts.RepoRoot, GitIgnoreName)
	original, readErr := os.ReadFile(path)
	existed := readErr == nil
	lines := splitLines(string(original))

	present := make(map[string]bool, len(lines))
	for _, l := range lines {
		present[strings.TrimSpace(l)] = true
	}

	// step 2.6/2.7 and 3.6 — add only what no existing rule already covers.
	var added []string
	for _, w := range want {
		line := ignoreLine(w)
		if already[w] || present[line] {
			continue
		}
		added = append(added, line)
		present[line] = true
	}

	// step 3.7 — remove what this value forbids, leaving every unrelated line alone.
	// Both spellings are removed, because a project may have written either.
	forbid := make(map[string]bool, len(mustNotIgnore)*2)
	for _, p := range mustNotIgnore {
		forbid[p] = true
		forbid[ignoreLine(p)] = true
	}
	removedFor := make(map[string]bool, len(mustNotIgnore))
	kept := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if forbid[trimmed] {
			removedFor[strings.TrimPrefix(trimmed, "/")] = true
			continue
		}
		kept = append(kept, l)
	}
	removedAny := len(kept) < len(lines)

	// A path git reports as ignored, for which we found no line to delete, is being
	// ignored by a rule this file does not own — a nested .gitignore, `.git/info/exclude`,
	// or the user's global excludes. Deleting our lines cannot make it public, so say so
	// rather than reporting a success that did not happen.
	for _, p := range mustNotIgnore {
		if already[p] && !removedFor[p] {
			result.Unresolved = append(result.Unresolved, p)
		}
	}

	if len(added) == 0 && !removedAny {
		if existed {
			result.Unchanged = append(result.Unchanged, path)
		}
		return nil
	}
	kept = append(kept, added...)

	body := strings.Join(kept, "\n")
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return err
	}
	result.record(path, existed)
	return nil
}

// splitLines splits a file into lines without inventing a trailing empty one, so a
// rewrite of an unchanged region is byte-stable.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	return strings.Split(s, "\n")
}

// CRC: crc-Init.md | Seq: seq-bootstrap.md#2.8 | R141
// Report prints every file created or edited, in full. Not a courtesy: these are files
// the agent did not write, and an agent that cannot see what changed will reason from a
// stale picture and eventually assert it. Unchanged files are named too, so "nothing to
// do" is distinguishable from "did not look".
func (r *InitResult) Report() string {
	var b strings.Builder
	section := func(label string, paths []string) {
		for _, p := range paths {
			fmt.Fprintf(&b, "  %s: %s\n", label, p)
		}
	}
	section("created", r.Created)
	section("edited", r.Edited)
	section("already correct", r.Unchanged)
	if b.Len() == 0 {
		b.WriteString("  (nothing to do)\n")
	}
	for _, p := range r.Unresolved {
		fmt.Fprintf(&b, "  STILL IGNORED: %s — by a rule outside the top-level %s\n"+
			"    (a nested %s, .git/info/exclude, or your global excludes). This value\n"+
			"    wants it tracked, and removing lines here cannot reach that rule.\n",
			p, GitIgnoreName, GitIgnoreName)
	}
	return b.String()
}
