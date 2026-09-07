// CRC: crc-Update.md | R310, R313, R314, R315, R316
package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

const alarmsDoc = `# Test Design: X

## Test: one
**Purpose:** p
**Fire alarm:** break it
**Inject:** a.go:F
**Pulled:** 2026-08-05 — rang, the original record
**Alarm:** 1

## Test: three
**Fire alarm:** another, wrapped
across a line
**Inject:** b.go:G
**Alarm:** 3

## Test: unnumbered
**Fire alarm:** a third
**Inject:** c.go:H
`

func alarmProject(t *testing.T, body string) (*Update, string) {
	t.Helper()
	root := t.TempDir()
	design := filepath.Join(root, "design")
	if err := os.MkdirAll(design, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(design, "test-X.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Update{Project: &project.Project{RootPath: root, DesignDir: design}}, path
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// stubRanger states where each site resolves; a site absent from the map resolves nowhere.
type stubRanger struct{ head, disk map[string][2]int }

func (s stubRanger) SiteRange(file, sym string) (int, int, error) {
	if r, ok := s.head[file+":"+sym]; ok {
		return r[0], r[1], nil
	}
	return 0, 0, project.ErrUnresolvedSite
}
func (s stubRanger) DiskRange(file, sym string) (int, int, error) {
	if r, ok := s.disk[file+":"+sym]; ok {
		return r[0], r[1], nil
	}
	return 0, 0, project.ErrUnresolvedSite
}

// R311, R313 — append-only: the freed number 2 is never handed out again, existing numbers
// are untouched, and a second run writes nothing.
func TestNumberAlarmsIsAppendOnlyAndIdempotent(t *testing.T) {
	u, path := alarmProject(t, alarmsDoc)
	docs, err := u.NumberAlarms(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 || len(docs[0].Assigned) != 1 || docs[0].Assigned[0] != 4 {
		t.Fatalf("assigned %v, want [4] — the freed 2 must not be reused", docs)
	}
	after := read(t, path)
	if !strings.Contains(after, "**Alarm:** 4\n**Fire alarm:** a third") {
		t.Errorf("the field was not inserted above the third alarm:\n%s", after)
	}
	stripped := strings.ReplaceAll(after, "**Alarm:** 4\n", "")
	if stripped != alarmsDoc {
		t.Errorf("the migration changed something other than the line it added")
	}
	again, err := u.NumberAlarms(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(again[0].Assigned) != 0 || read(t, path) != after {
		t.Errorf("a second run assigned %v or changed the document; it must find nothing to do", again[0].Assigned)
	}
}

// R314 — the leading date moves and the old record follows as history; the body's
// backticks survive; the neighbour is untouched; a first pull is inserted after Inject.
func TestSetPulledMovesTheLeadingDateAndInsertsAFirstRecord(t *testing.T) {
	u, path := alarmProject(t, alarmsDoc)
	now := time.Date(2026, 9, 7, 9, 0, 0, 0, time.UTC)
	if err := u.SetPulled("test-X.md#1", "rang: `state = \"x\"`", now); err != nil {
		t.Fatal(err)
	}
	got := read(t, path)
	want := "**Pulled:** 2026-09-07 — rang: `state = \"x\"` *Earlier —* 2026-08-05 — rang, the original record\n"
	if !strings.Contains(got, want) {
		t.Errorf("the new date does not lead with the old record folded after it:\n%s", got)
	}
	if err := u.SetPulled("test-X.md#3", "first pull", now); err != nil {
		t.Fatal(err)
	}
	got = read(t, path)
	if !strings.Contains(got, "**Inject:** b.go:G\n**Pulled:** 2026-09-07 — first pull\n**Alarm:** 3") {
		t.Errorf("a first pull was not inserted after the site it vouches for:\n%s", got)
	}
	if err := u.SetPulled("test-X.md#9", "x", now); err == nil {
		t.Error("an alarm number the document does not hold was accepted")
	}
	if err := u.SetPulled("test-X.md", "x", now); err == nil {
		t.Error("a key with no number was accepted")
	}
}

// R315 — a move voids the record and keeps it as history; a rewrite to the same code
// (disambiguation, rename) keeps the record; a rewrite to the same text changes nothing.
func TestSetInjectVoidsOnlyWhenTheCodeMoves(t *testing.T) {
	u, path := alarmProject(t, alarmsDoc)
	same := stubRanger{head: map[string][2]int{"a.go:F": {3, 9}}, disk: map[string][2]int{"a.go:T.F": {3, 9}, "c.go:Moved": {1, 2}}}

	cleared, err := u.SetInject("test-X.md#1", []parser.AlarmSite{{File: "a.go", Symbol: "F"}}, same)
	if err != nil || cleared {
		t.Fatalf("a no-op rewrite: cleared=%v err=%v; want neither", cleared, err)
	}
	if read(t, path) != alarmsDoc {
		t.Error("a no-op rewrite changed the document")
	}

	cleared, err = u.SetInject("test-X.md#1", []parser.AlarmSite{{File: "a.go", Symbol: "T.F"}}, same)
	if err != nil || cleared {
		t.Fatalf("disambiguating to the same lines: cleared=%v err=%v; want the record kept", cleared, err)
	}
	if got := read(t, path); !strings.Contains(got, "**Inject:** a.go:T.F\n**Pulled:** 2026-08-05") {
		t.Errorf("the site was not rewritten with the record standing:\n%s", got)
	}

	cleared, err = u.SetInject("test-X.md#1", []parser.AlarmSite{{File: "c.go", Symbol: "Moved"}}, same)
	if err != nil || !cleared {
		t.Fatalf("a move: cleared=%v err=%v; want the record voided", cleared, err)
	}
	got := read(t, path)
	if strings.Contains(got, "**Pulled:**") || !strings.Contains(got, "*Pulled at `a.go:T.F`") {
		t.Errorf("the record earned at the old site survived the move, or its history was dropped:\n%s", got)
	}
	if _, err := u.SetInject("test-X.md#3", nil, same); err == nil {
		t.Error("an empty site list was written")
	}
}
