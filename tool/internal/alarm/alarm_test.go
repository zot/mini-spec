// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md | R179, R181, R182, R184, R185, R187
package alarm

import (
	"testing"
	"time"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

// stubGit states a world instead of consulting one, so the whole state table is
// decidable without a repository anywhere. See test-Alarm.md.
type stubGit struct {
	repo       bool
	changed    map[string]time.Time
	unresolved map[string]bool
	nohistory  map[string]bool
}

func (s *stubGit) IsRepo() bool                              { return s.repo }
func (s *stubGit) Ignored([]string) (map[string]bool, error) { return nil, nil }
func (s *stubGit) Tracked(string) (bool, error)              { return false, nil }
func (s *stubGit) LastChanged(file, sym string) (time.Time, error) {
	if !s.repo {
		return time.Time{}, project.ErrNoGit
	}
	site := file + ":" + sym
	if s.unresolved[site] {
		return time.Time{}, project.ErrUnresolvedSite
	}
	if s.nohistory[site] {
		return time.Time{}, project.ErrNoHistory
	}
	return s.changed[site], nil
}

func (s *stubGit) SiteResolves(file, sym string) error {
	_, err := s.LastChanged(file, sym)
	return err
}

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func alarmWith(sites []parser.AlarmSite, pulled string) parser.Alarm {
	a := parser.Alarm{Doc: "test-X.md", Test: "a test", Sites: sites}
	if pulled != "" {
		a.Pulled = day(pulled)
	}
	return a
}

var siteA = []parser.AlarmSite{{File: "a.go", Symbol: "Foo"}}

// R185 — every state is reachable, and only by its own shape.
func TestEachStateIsReachableByItsOwnShape(t *testing.T) {
	git := &stubGit{repo: true, changed: map[string]time.Time{"a.go:Foo": day("2026-08-01")}}
	for _, tc := range []struct {
		name  string
		alarm parser.Alarm
		g     project.GitFacts
		want  State
	}{
		{"verified", alarmWith(siteA, "2026-08-05"), git, Verified},
		{"stale", alarmWith(siteA, "2026-07-01"), git, Stale},
		{"unrecorded", alarmWith(siteA, ""), git, Unrecorded},
		{"unanchored", alarmWith(nil, "2026-08-05"), git, Unanchored},
		{"unchecked", alarmWith(siteA, "2026-08-05"), &stubGit{repo: false}, Unchecked},
		{"unresolvable", alarmWith(siteA, "2026-08-05"),
			&stubGit{repo: true, unresolved: map[string]bool{"a.go:Foo": true}}, Unresolvable},
	} {
		got := Assess([]parser.Alarm{tc.alarm}, tc.g)[0]
		if got.State != tc.want {
			t.Errorf("%s: state = %q, want %q", tc.name, got.State, tc.want)
		}
	}
}

// R181 — the boundary, which is the rule this feature got wrong first.
//
// Same-day must be **verified**, not stale: the normal workflow commits the fix, the
// test and the pull together, so a freshly recorded alarm always shares a date with the
// code it proves. The inclusive version marked two of ark's three verified alarms stale
// the day they were written, and a check that fires on arrival is ignored.
func TestSameDayIsVerifiedAndTheNextDayIsStale(t *testing.T) {
	git := &stubGit{repo: true, changed: map[string]time.Time{"a.go:Foo": day("2026-08-05")}}

	if got := Assess([]parser.Alarm{alarmWith(siteA, "2026-08-05")}, git)[0]; got.State != Verified {
		t.Errorf("same-day = %q, want verified — an inclusive boundary fires on every fresh alarm", got.State)
	}
	if got := Assess([]parser.Alarm{alarmWith(siteA, "2026-08-04")}, git)[0]; got.State != Stale {
		t.Errorf("change one day later = %q, want stale", got.State)
	}
}

// R179 — a stale verdict names the site and the date that voided it, because a report
// that says only "something moved" leaves the reader to redo the search.
func TestStaleNamesTheSiteAndTheDate(t *testing.T) {
	git := &stubGit{repo: true, changed: map[string]time.Time{
		"a.go:Foo": day("2026-08-01"),
		"b.go:Bar": day("2026-09-09"),
	}}
	a := alarmWith([]parser.AlarmSite{{File: "a.go", Symbol: "Foo"}, {File: "b.go", Symbol: "Bar"}},
		"2026-08-05")
	got := Assess([]parser.Alarm{a}, git)[0]
	if got.State != Stale {
		t.Fatalf("state = %q, want stale", got.State)
	}
	if got.Site != "b.go:Bar" || got.Changed != "2026-09-09" {
		t.Errorf("got site %q changed %q, want b.go:Bar / 2026-09-09", got.Site, got.Changed)
	}
}

// R182 — a site git cannot resolve is reported, never treated as unchanged. Silence
// there would be the tool declining to report its own mechanism failing.
func TestUnresolvableIsReportedNotSilentlyPassed(t *testing.T) {
	git := &stubGit{repo: true, unresolved: map[string]bool{"a.go:Foo": true}}
	got := Assess([]parser.Alarm{alarmWith(siteA, "2026-08-05")}, git)[0]
	if got.State != Unresolvable {
		t.Fatalf("state = %q, want unresolvable", got.State)
	}
	if got.Site != "a.go:Foo" {
		t.Errorf("site = %q, want it named", got.Site)
	}
}

// R184, R187 — a check that could not look must not return a clean result. The states
// needing no git still resolve, which is what keeps the cost proportional.
func TestNoGitYieldsUncheckedNeverVerified(t *testing.T) {
	none := &stubGit{repo: false}
	if got := Assess([]parser.Alarm{alarmWith(siteA, "2026-08-05")}, none)[0]; got.State != Unchecked {
		t.Errorf("with no git = %q, want unchecked", got.State)
	}
	if got := Assess([]parser.Alarm{alarmWith(siteA, "")}, none)[0]; got.State != Unrecorded {
		t.Errorf("unrecorded needs no git, got %q", got.State)
	}
	if got := Assess([]parser.Alarm{alarmWith(nil, "")}, none)[0]; got.State != Unanchored {
		t.Errorf("unanchored needs no git, got %q", got.State)
	}
}

// R185 — the counts partition the input. A census that drops or double-counts reads as
// a measurement while being an error.
func TestCensusPartitionsEveryAlarm(t *testing.T) {
	git := &stubGit{repo: true, changed: map[string]time.Time{"a.go:Foo": day("2026-09-09")}}
	alarms := []parser.Alarm{
		alarmWith(siteA, "2026-08-05"), // stale
		alarmWith(siteA, ""),           // unrecorded
		alarmWith(nil, ""),             // unanchored
		alarmWith(nil, "2026-08-05"),   // unanchored
	}
	as := Assess(alarms, git)
	total := 0
	for _, n := range Census(as) {
		total += n
	}
	if total != len(alarms) {
		t.Errorf("census totals %d, want %d — the states must partition", total, len(alarms))
	}
}

// R183 — validate's slice is the closable one. Unrecorded and unanchored are excluded
// deliberately: they stay non-zero for months, and a permanent non-zero line is the nag
// this project distinguishes from a gripe you can discharge.
func TestVoidedCarriesOnlyTheClosableStates(t *testing.T) {
	git := &stubGit{repo: true,
		changed:    map[string]time.Time{"a.go:Foo": day("2026-09-09")},
		unresolved: map[string]bool{"b.go:Bar": true},
	}
	alarms := []parser.Alarm{
		alarmWith(siteA, "2026-08-05"),                                             // stale
		alarmWith([]parser.AlarmSite{{File: "b.go", Symbol: "Bar"}}, "2026-08-05"), // unresolvable
		alarmWith(siteA, ""),                                                       // unrecorded
		alarmWith(nil, ""),                                                         // unanchored
	}
	v := Voided(Assess(alarms, git))
	if len(v) != 2 {
		t.Fatalf("Voided returned %d, want 2 (stale + unresolvable)", len(v))
	}
	for _, a := range v {
		if a.State != Stale && a.State != Unresolvable {
			t.Errorf("Voided included %q, which validate must not report", a.State)
		}
	}
}

// R179 — an inconclusive site must not mask a definite one. A site git cannot answer
// for is remembered; if a later site is definitely stale, stale is the verdict, because
// a finding you have is worth more than one you could not reach.
func TestAnUncheckableSiteDoesNotMaskAStaleOne(t *testing.T) {
	git := &stubGit{
		repo:      true,
		nohistory: map[string]bool{"new.go:Fresh": true},
		changed:   map[string]time.Time{"old.go:Moved": day("2026-09-09")},
	}
	a := alarmWith([]parser.AlarmSite{
		{File: "new.go", Symbol: "Fresh"}, // asked first, and unanswerable
		{File: "old.go", Symbol: "Moved"}, // definitely stale
	}, "2026-08-05")

	got := Assess([]parser.Alarm{a}, git)[0]
	if got.State != Stale {
		t.Errorf("state = %q, want stale — an unanswerable first site hid a definite finding", got.State)
	}
	// With nothing worse to find, the unanswerable site is still reported rather than
	// passing as verified.
	only := alarmWith([]parser.AlarmSite{{File: "new.go", Symbol: "Fresh"}}, "2026-08-05")
	if got := Assess([]parser.Alarm{only}, git)[0]; got.State != Unchecked {
		t.Errorf("lone unanswerable site = %q, want unchecked", got.State)
	}
}

// R309 — a prescription's anchor is checked before it reads as unrecorded: the site that
// nobody has run is the one most likely to have rotted, and until now the only one never
// looked at. Without git it stays unrecorded, which is what it was.
func TestAPrescriptionsRottedAnchorIsUnresolvable(t *testing.T) {
	git := &stubGit{repo: true, unresolved: map[string]bool{"a.go:Foo": true}}
	got := Assess([]parser.Alarm{alarmWith(siteA, "")}, git)[0]
	if got.State != Unresolvable || got.Site != "a.go:Foo" {
		t.Errorf("state = %q site = %q; want unresolvable at a.go:Foo — a prescription pointing at nothing read as an ordinary prescription", got.State, got.Site)
	}
	intact := &stubGit{repo: true, changed: map[string]time.Time{"a.go:Foo": day("2026-08-01")}}
	if got := Assess([]parser.Alarm{alarmWith(siteA, "")}, intact)[0]; got.State != Unrecorded {
		t.Errorf("intact prescription = %q, want unrecorded", got.State)
	}
	if got := Assess([]parser.Alarm{alarmWith(siteA, "")}, &stubGit{repo: false})[0]; got.State != Unrecorded {
		t.Errorf("prescription with no git = %q, want unrecorded — nothing could be checked", got.State)
	}
}
