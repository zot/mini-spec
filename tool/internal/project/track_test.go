// CRC: crc-Track.md | Seq: seq-bootstrap.md | R131, R133, R134, R147, R148, R149, R151, R164, R165, R173
package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// fakeGit states the world a consistency check is verified against, so no repository
// has to exist and no `git` has to run. This double is the reason Git is a separate
// component: the seam is what makes the table below possible. See test-Track.md.
type fakeGit struct {
	repo    bool
	ignored map[string]bool
	tracked map[string]bool
	// changed states when a "file:symbol" site last changed. A site absent from the
	// map has never changed; a site listed in unresolved is one git cannot find, which
	// is a different answer and must stay distinguishable from both.
	changed    map[string]time.Time
	unresolved map[string]bool
	untracked  map[string]bool
}

func (f *fakeGit) IsRepo() bool { return f.repo }

func (f *fakeGit) Ignored(paths []string) (map[string]bool, error) {
	if !f.repo {
		return nil, ErrNoGit
	}
	out := make(map[string]bool, len(paths))
	for _, p := range paths {
		out[p] = f.ignored[p]
	}
	return out, nil
}

func (f *fakeGit) Tracked(path string) (bool, error) {
	if !f.repo {
		return false, ErrNoGit
	}
	return f.tracked[path], nil
}

func (f *fakeGit) LastChanged(file, symbol string) (time.Time, error) {
	if !f.repo {
		return time.Time{}, ErrNoGit
	}
	// An untracked file has no history to search, which is "could not look" rather
	// than a rotted anchor. The fake models it so a test cannot assert a world the
	// real Git could never produce.
	if f.untracked[file] {
		return time.Time{}, ErrNoHistory
	}
	site := file + ":" + symbol
	if f.unresolved[site] {
		return time.Time{}, ErrUnresolvedSite
	}
	return f.changed[site], nil
}

func (f *fakeGit) SiteResolves(file, symbol string) error {
	_, err := f.LastChanged(file, symbol)
	return err
}

func gitWith(ignored ...string) *fakeGit {
	f := &fakeGit{repo: true, ignored: map[string]bool{}, tracked: map[string]bool{}}
	for _, p := range ignored {
		f.ignored[p] = true
	}
	return f
}

// R131 — the closed set. An unrecognised value is a malformed configuration rather
// than a fourth behavior.
func TestParseTrackAcceptsOnlyTheClosedSet(t *testing.T) {
	for _, ok := range []string{"none", "private-trajectory", "all"} {
		if _, err := ParseTrack(ok); err != nil {
			t.Errorf("ParseTrack(%q) = error %v, want accepted", ok, err)
		}
	}
	for _, bad := range []string{"None", "private", "", "yes", "ALL", "private_trajectory"} {
		if v, err := ParseTrack(bad); err == nil {
			t.Errorf("ParseTrack(%q) = %q, want rejected", bad, v)
		}
	}
}

// R132, R147 — the git-presence half, in the direction that catches a project gaining
// git after init.
func TestTrackNoneAgreesOnlyWhenGitAbsent(t *testing.T) {
	dir := t.TempDir()

	ms := TrackNone.Verify(&fakeGit{repo: true}, dir)
	if len(ms) != 1 || ms[0].Fact != "git presence" {
		t.Fatalf("none in a git tree = %+v, want one git-presence mismatch", ms)
	}

	if ms := TrackNone.Verify(&fakeGit{repo: false}, dir); len(ms) != 0 {
		t.Fatalf("none outside git = %+v, want agreement", ms)
	}
}

// R147 — the mirror direction: a declared git style in a tree with no repository.
func TestGitStylesRequireAGitTree(t *testing.T) {
	dir := t.TempDir()
	for _, v := range []TrackValue{TrackPrivateTrajectory, TrackAll} {
		ms := v.Verify(&fakeGit{repo: false}, dir)
		if len(ms) != 1 || ms[0].Fact != "git presence" {
			t.Errorf("%s outside git = %+v, want one git-presence mismatch", v, ms)
		}
	}
}

// R133, R148 — the ignore-state half, which fails independently of git presence.
func TestPrivateTrajectoryRequiresTheQueueIgnored(t *testing.T) {
	dir := t.TempDir()
	ms := TrackPrivateTrajectory.Verify(gitWith(), dir)
	if len(ms) != len(TrajectoryFiles) {
		t.Fatalf("got %d mismatches, want %d: %+v", len(ms), len(TrajectoryFiles), ms)
	}
	joined := mismatchText(ms)
	for _, f := range TrajectoryFiles {
		if !strings.Contains(joined, f) {
			t.Errorf("mismatch report does not name %s: %s", f, joined)
		}
	}
}

// R132, R148 — the opposite expectation from the same fact. This asymmetry is what
// makes repair need to run in both directions.
func TestAllRequiresTheQueueNotIgnored(t *testing.T) {
	dir := t.TempDir()
	ms := TrackAll.Verify(gitWith(TrajectoryFiles...), dir)
	if len(ms) != len(TrajectoryFiles) {
		t.Fatalf("got %d mismatches, want %d: %+v", len(ms), len(TrajectoryFiles), ms)
	}
	if !strings.Contains(mismatchText(ms), "should not be ignored") {
		t.Errorf("report does not say the files should not be ignored: %s", mismatchText(ms))
	}
}

// R134 — the one path whose expectation does not follow the queue's. A directory whose
// whole purpose is privacy does not become public because the queue did.
func TestPrivateCarvesMustBeIgnoredEvenUnderAll(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, PrivateCarvesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	// The trajectory files are correctly not ignored under `all`, so `.carves/` is the
	// only thing left that can disagree.
	ms := TrackAll.Verify(gitWith(), dir)
	if len(ms) != 1 {
		t.Fatalf("got %+v, want exactly the .carves mismatch", ms)
	}
	if !strings.Contains(ms[0].Detail, PrivateCarvesDir) {
		t.Errorf("mismatch does not name %s: %s", PrivateCarvesDir, ms[0].Detail)
	}
}

// R133 — checked on demand. Requiring the directory would turn an optional convention
// into a mandate.
func TestPrivateCarvesIsNotRequiredToExist(t *testing.T) {
	dir := t.TempDir()
	if ms := TrackPrivateTrajectory.Verify(gitWith(TrajectoryFiles...), dir); len(ms) != 0 {
		t.Fatalf("got %+v, want agreement when .carves/ is absent", ms)
	}
}

// R149 — collecting rather than returning the first, so a repair message never sends
// the user round twice.
func TestVerifyCollectsEveryDisagreement(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, PrivateCarvesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	// private-trajectory in a git tree that ignores nothing: three queue files plus
	// .carves/ all disagree at once.
	ms := TrackPrivateTrajectory.Verify(gitWith(), dir)
	if len(ms) != len(TrajectoryFiles)+1 {
		t.Fatalf("got %d mismatches, want %d: %+v", len(ms), len(TrajectoryFiles)+1, ms)
	}
}

// R146, R149 — the crank handle. A gripe that does not say what to run is a nag.
func TestMismatchReportNamesTheRepair(t *testing.T) {
	report := MismatchReport(TrackNone, []Mismatch{{Fact: "git presence", Detail: "x"}})
	for _, want := range []string{"init --track-", "--repair", "confirm the value with the user"} {
		if !strings.Contains(report, want) {
			t.Errorf("report does not contain %q:\n%s", want, report)
		}
	}
}

// R151 — a tree with no repository is a supported shape, not a misconfiguration.
func TestNoGitIsSilent(t *testing.T) {
	dir := t.TempDir()
	g := &fakeGit{repo: false}
	if ms := TrackNone.Verify(g, dir); len(ms) != 0 {
		t.Errorf("Verify = %+v, want silence", ms)
	}
	if pref := PreferenceReport(g, dir); pref != "" {
		t.Errorf("PreferenceReport = %q, want silence", pref)
	}
}

// R164, R165 — both preferences are reported, and only while unmet.
func TestPreferenceReportNamesOnlyUnmetPreferences(t *testing.T) {
	dir := t.TempDir()
	cfgRel := ConfigDirName + "/" + RepoConfigName

	unmet := PreferenceReport(gitWith(), dir)
	if !strings.Contains(unmet, cfgRel) || !strings.Contains(unmet, BackupIgnorePath) {
		t.Errorf("both preferences unmet but report is:\n%s", unmet)
	}

	g := gitWith(BackupIgnorePath)
	g.tracked[cfgRel] = true
	if met := PreferenceReport(g, dir); met != "" {
		t.Errorf("both preferences met but report is:\n%s", met)
	}
}

// R131, R152, R173 — neither absence nor a bad value is defaulted, and the two are
// reported *apart* so the caller can name the repair that actually applies.
//
// Asserting only "an error" is what let the two collapse: the gate then sent every
// pre-`track` configuration to the hand-edit refusal, which `--repair` would have
// accepted. So each case names the error it must produce.
func TestLoadTrackRejectsAMissingOrBadValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	for _, tc := range []struct {
		name, body  string
		wantNoTrack bool
	}{
		{name: "empty", body: "", wantNoTrack: true},
		{name: "no track key", body: "design_dir: design\n", wantNoTrack: true},
		{name: "blank track", body: "track: \"\"\n", wantNoTrack: true},
		{name: "unknown value", body: "track: sometimes\n"},
		{name: "unparseable", body: "track: [\n"},
	} {
		if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
			t.Fatal(err)
		}
		v, err := LoadTrack(path)
		if err == nil {
			t.Errorf("%s: LoadTrack = %q, want an error", tc.name, v)
			continue
		}
		if got := errors.Is(err, ErrNoTrack); got != tc.wantNoTrack {
			t.Errorf("%s: errors.Is(err, ErrNoTrack) = %v, want %v (err: %v)",
				tc.name, got, tc.wantNoTrack, err)
		}
		// Damage authorises a hand edit; absence must not, because a flag repairs it.
		if authorised := strings.Contains(err.Error(), "authorised to edit"); authorised == tc.wantNoTrack {
			t.Errorf("%s: hand-edit authorisation present = %v, want %v: %v",
				tc.name, authorised, !tc.wantNoTrack, err)
		}
	}

	if err := os.WriteFile(path, []byte("track: all\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if v, err := LoadTrack(path); err != nil || v != TrackAll {
		t.Errorf("LoadTrack = (%q, %v), want (all, nil)", v, err)
	}
}

// R173 — the pre-`track` error names the verb that fixes it. A refusal that does not
// say what to run is a nag, and this one is the only prompt the case ever gets.
func TestNoTrackErrorNamesRepair(t *testing.T) {
	if !strings.Contains(ErrNoTrack.Error(), "--repair") {
		t.Errorf("ErrNoTrack does not name --repair: %v", ErrNoTrack)
	}
}

func mismatchText(ms []Mismatch) string {
	parts := make([]string, 0, len(ms))
	for _, m := range ms {
		parts = append(parts, m.Fact+": "+m.Detail)
	}
	return strings.Join(parts, "\n")
}
