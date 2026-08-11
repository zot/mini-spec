// CRC: crc-Track.md | Seq: seq-bootstrap.md | R131
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// CRC: crc-Track.md | R131
// TrackValue is the one thing about a repository a tool cannot infer: whether it is
// version-controlled, and whether its work queue ships with it. A closed set of three,
// so an unrecognised value is a malformed configuration rather than a fourth behavior.
type TrackValue string

const (
	// TrackNone asserts the project is not under version control. R132
	TrackNone TrackValue = "none"
	// TrackPrivateTrajectory asserts git, with the trajectory files ignored. R132
	TrackPrivateTrajectory TrackValue = "private-trajectory"
	// TrackAll asserts git, with the trajectory files tracked like everything else. R132
	TrackAll TrackValue = "all"
)

// TrajectoryFiles are the queue files whose ignore state `track` governs. Their
// location at the repository root is mandated rather than configured: the trajectory
// crosses design roots, so it is repository-scoped and there is nothing to site. R133
var TrajectoryFiles = []string{"PENDING.md", "CURRENT.md", "DONE.md"}

// PrivateCarvesDir is the private carve directory, checked on demand because it is
// created only when a project wants one. R133
const PrivateCarvesDir = ".carves"

// BackupDirName holds the tool's machine-local working files, including the revert
// slot's stamp. One ignore line then covers everything machine-local, leaving
// config.yaml the only visible thing in `.minispec/`. R165
const BackupDirName = "backup"

// BackupIgnorePath is the .gitignore entry for that directory.
var BackupIgnorePath = ConfigDirName + "/" + BackupDirName

// CRC: crc-Track.md | R131
// ParseTrack resolves a configuration value, rejecting anything outside the closed set.
func ParseTrack(s string) (TrackValue, error) {
	switch v := TrackValue(strings.TrimSpace(s)); v {
	case TrackNone, TrackPrivateTrajectory, TrackAll:
		return v, nil
	default:
		return "", fmt.Errorf(
			"track: %q is not a recognized value (expected %s, %s, or %s)",
			s, TrackNone, TrackPrivateTrajectory, TrackAll)
	}
}

// IsGitStyle reports whether this value asserts a git-managed repository. R132
func (v TrackValue) IsGitStyle() bool {
	return v == TrackPrivateTrajectory || v == TrackAll
}

// CRC: crc-Track.md | Seq: seq-bootstrap.md#1.9 | R133, R134
// Expectations returns the paths this value requires git to ignore and the paths it
// requires git *not* to ignore, relative to the repository root.
//
// `.carves/` is required ignored under every git-managed value, `all` included: a
// directory whose whole purpose is privacy does not become public because the queue
// did. Public `carves/` is never mentioned, because nothing about it is in question.
func (v TrackValue) Expectations(repoRoot string) (mustIgnore, mustNotIgnore []string) {
	if !v.IsGitStyle() {
		return nil, nil
	}
	if v == TrackPrivateTrajectory {
		mustIgnore = append(mustIgnore, TrajectoryFiles...)
	} else {
		mustNotIgnore = append(mustNotIgnore, TrajectoryFiles...)
	}
	// Checked on demand: the directory need not exist, and demanding it would turn an
	// optional convention into a mandate.
	if _, err := os.Lstat(filepath.Join(repoRoot, PrivateCarvesDir)); err == nil {
		mustIgnore = append(mustIgnore, PrivateCarvesDir)
	}
	return mustIgnore, mustNotIgnore
}

// Mismatch is one disagreement between the declared value and what git reports. R149
type Mismatch struct {
	// Fact names which of the two checks disagreed, so the repair message can say
	// which rather than leaving the reader to work it out.
	Fact   string
	Detail string
}

// CRC: crc-Track.md | Seq: seq-bootstrap.md#1.10 | R147, R148, R151
// Verify compares the declared value against the two facts git can supply — whether
// the tree is a working tree, and the actual ignore state of the governed paths — and
// returns every disagreement.
//
// The two facts fail independently: a project can gain git after `init`, or have its
// ignore lines hand-edited. Collecting rather than returning the first means a repair
// message never sends the user round twice.
//
// A tree with no git at all and `track: none` is in agreement and silent — an
// instruction naming a file the project does not have is worse than no instruction.
func (v TrackValue) Verify(g GitFacts, repoRoot string) []Mismatch {
	var out []Mismatch
	present := g.IsRepo()

	switch {
	case v == TrackNone && present:
		out = append(out, Mismatch{
			Fact: "git presence",
			Detail: "track is none, but this is a git working tree. " +
				"Change it to private-trajectory or all.",
		})
	case v.IsGitStyle() && !present:
		out = append(out, Mismatch{
			Fact: "git presence",
			Detail: fmt.Sprintf("track is %s, but this is not a git working tree. "+
				"Change it to none.", v),
		})
	}

	// Nothing below is answerable without a working tree, and saying so is the point:
	// a check that could not look must not return a clean result.
	if !present {
		return out
	}

	mustIgnore, mustNotIgnore := v.Expectations(repoRoot)
	ignored, err := g.Ignored(append(slices.Clone(mustIgnore), mustNotIgnore...))
	if err != nil {
		return append(out, Mismatch{
			Fact:   "ignore state",
			Detail: "could not be determined: " + err.Error(),
		})
	}
	for _, p := range mustIgnore {
		if !ignored[p] {
			out = append(out, Mismatch{
				Fact:   "ignore state",
				Detail: fmt.Sprintf("%s should be ignored under track: %s, and is not", p, v),
			})
		}
	}
	for _, p := range mustNotIgnore {
		if ignored[p] {
			out = append(out, Mismatch{
				Fact:   "ignore state",
				Detail: fmt.Sprintf("%s should not be ignored under track: %s, and is", p, v),
			})
		}
	}
	return out
}

// CRC: crc-Track.md | Seq: seq-bootstrap.md#1.6 | R131, R152
// LoadTrack reads the declared value from the repository configuration.
//
// A configuration with no `track` is malformed rather than defaulted. `init` is the
// sole creator and always writes one, so its absence means the file was hand-made or
// damaged — and guessing on its behalf would silently make the choice the mandatory
// flag exists to prevent anyone making by accident.
func LoadTrack(cfgPath string) (TrackValue, error) {
	cfg, err := readRepoConfig(cfgPath)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(cfg.Track) == "" {
		return "", malformedConfigError(cfgPath, fmt.Errorf(
			"no `track` setting. Every configuration has one, because `minispec init`"+
				" cannot be run without choosing it"))
	}
	v, err := ParseTrack(cfg.Track)
	if err != nil {
		return "", malformedConfigError(cfgPath, err)
	}
	return v, nil
}

// CRC: crc-Track.md | Seq: seq-bootstrap.md#1.12 | R164, R165, R151
// PreferenceReport names the git preferences this repository does not meet.
//
// Reported on every run and never refused over: these are preferences, not invariants.
// Both are computed from git rather than stored, so there is nothing to assert and
// nothing to go stale.
//
// Silent in a tree with no git, where neither preference means anything.
func PreferenceReport(g GitFacts, repoRoot string) string {
	if !g.IsRepo() {
		return ""
	}
	cfgRel := ConfigDirName + "/" + RepoConfigName
	var notes []string
	if tracked, err := g.Tracked(cfgRel); err == nil && !tracked {
		notes = append(notes, fmt.Sprintf(
			"%s should be tracked by git, and is not. It is the one part of %s meant to be shared.",
			cfgRel, ConfigDirName))
	}
	if ignored, err := g.Ignored([]string{BackupIgnorePath}); err == nil && !ignored[BackupIgnorePath] {
		notes = append(notes, fmt.Sprintf(
			"%s should be ignored by git, and is not. It holds machine-local working files.",
			BackupIgnorePath))
	}
	if len(notes) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("note: unmet git preferences\n")
	for _, n := range notes {
		fmt.Fprintf(&b, "  - %s\n", n)
	}
	return b.String()
}

// CRC: crc-Track.md | Seq: seq-bootstrap.md#1.11 | R146, R149, R150
// MismatchReport is the crank handle for a failed Verify. It names every disagreement
// and the exact command that repairs it, because a gripe that does not say what to run
// is a nag.
//
// It fires on every run until repaired rather than once per session: a one-shot
// reminder decays to nothing — measured in this project's sibling repository, where the
// one category carrying only a reminder sat at 19% stale while every category carrying
// a forcing function sat at zero. It cannot become wallpaper either, because it is
// closable.
func MismatchReport(v TrackValue, ms []Mismatch) string {
	var b strings.Builder
	fmt.Fprintf(&b, "track: %s does not match this repository.\n\n", v)
	for _, m := range ms {
		fmt.Fprintf(&b, "  - %s: %s\n", m.Fact, m.Detail)
	}
	b.WriteString("\nRepair it with:\n\n")
	b.WriteString("    minispec init --track-<none|private-trajectory|all> --repair\n\n")
	b.WriteString("The flag value decides which way the mismatch is resolved, and the two\n")
	b.WriteString("directions mean opposite things — confirm the value with the user before\n")
	b.WriteString("running it. Do not hand-edit the configuration.\n")
	return b.String()
}
