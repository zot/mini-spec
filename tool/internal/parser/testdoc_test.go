// CRC: crc-Parser.md | Seq: seq-alarm-freshness.md#1.1 | R178
package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// Pure deterministic string parsing over a temp file — the cheap case the
// Implementation phase says has no excuse for going hand-checked. Written after a
// simplification pass restructured ParseTestDoc and a green suite proved nothing,
// because nothing here was tested at all.

func parse(t *testing.T, body string) []Alarm {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test-Fixture.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ParseTestDoc(path)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

// R178 — the fields, and the wrap. A reader assuming one line per field sees half of
// every alarm, which is the failure that bit three separate greps during this feature's
// own measurement.
func TestParseTestDocFoldsWrappedProseAndReadsFields(t *testing.T) {
	got := parse(t, `# Test Design: Fixture

## Test: the first one
**Purpose:** whatever
**Fire alarm:** delete the thing and confirm red. This sentence
wraps across a line, and this one does too.
**Inject:** a.go:Foo, b.go:Bar
**Pulled:** 2026-08-05 — rang
**Refs:** crc-X.md — R1
`)
	if len(got) != 1 {
		t.Fatalf("got %d alarms, want 1", len(got))
	}
	a := got[0]
	if a.Test != "the first one" {
		t.Errorf("Test = %q", a.Test)
	}
	if want := "delete the thing and confirm red. This sentence wraps across a line, and this one does too."; a.Prose != want {
		t.Errorf("Prose = %q\nwant %q", a.Prose, want)
	}
	if len(a.Sites) != 2 || a.Sites[0].String() != "a.go:Foo" || a.Sites[1].String() != "b.go:Bar" {
		t.Errorf("Sites = %v, want a.go:Foo and b.go:Bar", a.Sites)
	}
	if !a.HasPulled() || a.Pulled.Format("2006-01-02") != "2026-08-05" {
		t.Errorf("Pulled = %v", a.Pulled)
	}
}

// R178 — one alarm must never adopt the next test's fields. Getting this wrong would
// silently attribute a verification to an alarm that never had one, which is the
// strongest possible false claim this parser could make.
func TestParseTestDocDoesNotLetAnAlarmAdoptTheNextTestsFields(t *testing.T) {
	got := parse(t, `## Test: unrecorded one
**Fire alarm:** break it
**Refs:** crc-X.md

## Test: recorded one
**Fire alarm:** break the other
**Inject:** b.go:Bar
**Pulled:** 2026-08-05 — rang
`)
	if len(got) != 2 {
		t.Fatalf("got %d alarms, want 2", len(got))
	}
	if got[0].HasPulled() || len(got[0].Sites) != 0 {
		t.Errorf("first alarm adopted the second's fields: sites=%v pulled=%v", got[0].Sites, got[0].Pulled)
	}
	if !got[1].HasPulled() || len(got[1].Sites) != 1 {
		t.Errorf("second alarm lost its own fields: sites=%v pulled=%v", got[1].Sites, got[1].Pulled)
	}
}

// R178 — the shapes that must not be mistaken for fields or sites.
func TestParseTestDocRejectsWhatIsNotAField(t *testing.T) {
	got := parse(t, `## Test: only one
**Fire alarm:** a mention of **Inject:** mid-prose must stay prose
**Inject:** good.go:Fine, garbage-no-colon, :nofile, nosymbol:
**Pulled:** not-a-date
`)
	if len(got) != 1 {
		t.Fatalf("got %d alarms, want 1", len(got))
	}
	a := got[0]
	if len(a.Sites) != 1 || a.Sites[0].String() != "good.go:Fine" {
		t.Errorf("Sites = %v, want only good.go:Fine — a half anchor points somewhere, and somewhere is worse than nowhere", a.Sites)
	}
	if a.HasPulled() {
		t.Errorf("an unparseable date was accepted as a verification: %v", a.Pulled)
	}
}

// R178 — a document with no alarms is empty, not an error, since most test designs
// carry none.
func TestParseTestDocIsEmptyWhenThereAreNoAlarms(t *testing.T) {
	if got := parse(t, "## Test: plain\n**Purpose:** nothing to inject\n"); len(got) != 0 {
		t.Errorf("got %d alarms, want none", len(got))
	}
}
