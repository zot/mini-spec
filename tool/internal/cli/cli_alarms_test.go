// CRC: crc-CLI.md | Seq: seq-alarm-freshness.md#2.6 | R199, R204
package cli

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/query"
)

// alarmProject builds a real git repository holding one alarm that resolves to `verified`
// and one that resolves to `unrecorded`.
//
// Git is not scaffolding here, it is the point: without a repository every alarm assesses as
// `unchecked` and the filter under test becomes a no-op, so a fixture that skipped git would
// pass whether or not the code worked. See test-Alarm.md.
func alarmProject(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("specs/overview.md", "# Overview\n")
	write("subject.go", "package subject\n\nfunc Guard() bool {\n\treturn true\n}\n")
	write("design/requirements.md", "# Requirements\n")
	write("design/design.md", "# Design\n\n## Artifacts\n\n### Test Designs\n- [ ] test-Fixture.md → `subject_test.go`\n\n## Gaps\n")
	write("design/test-Fixture.md", `# Test Design: Fixture
**Source:** crc-Fixture.md

## Test: the proven one
**Fire alarm:** invert the guard and confirm this goes red
**Inject:** subject.go:Guard
**Pulled:** 2099-01-01 — rang

## Test: the prescribed one
**Fire alarm:** delete the guard entirely
**Inject:** subject.go:Guard
`)
	for _, args := range [][]string{
		{"init", "-q"}, {"add", "-A"},
		{"-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-qm", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v: %s", err, out)
		}
	}
	// os.Chdir rather than t.Chdir, for the reason cli_next_id_test.go records: t.Chdir
	// needs go1.24 and this module declares go1.21. Never parallel.
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })
}

func runAlarms(t *testing.T, args ...string) string {
	t.Helper()
	return captureStdout(t, func() {
		c := &CLI{}
		if code := c.runQuery(append([]string{"alarms"}, args...)); code != 0 {
			t.Fatalf("runQuery alarms %v exited %d", args, code)
		}
	})
}

// R199. The list narrows and the census does not. A census filtered along with its list
// answers *how many are wrong* and silently drops *out of how many* — the `| tail -2`
// workaround this flag replaces, one level in.
//
// Fire alarm: pass `selected` to alarmCensusLine in queryAlarms instead of `assessments`.
// The listing is byte-identical either way, which is exactly why this assertion is on the
// closing line. See test-Alarm.md.
func TestUnverifiedNarrowsTheListNotTheCensus(t *testing.T) {
	alarmProject(t)

	full := runAlarms(t)
	filtered := runAlarms(t, "--unverified")

	if !strings.Contains(full, "the proven one") {
		t.Fatalf("fixture did not produce a verified alarm:\n%s", full)
	}
	if strings.Contains(filtered, "the proven one") {
		t.Errorf("--unverified listed a verified alarm:\n%s", filtered)
	}
	if !strings.Contains(filtered, "the prescribed one") {
		t.Errorf("--unverified dropped an unrecorded alarm:\n%s", filtered)
	}
	if !strings.Contains(filtered, "2 alarms:") {
		t.Errorf("the census under --unverified does not cover the whole population:\n%s", filtered)
	}
	if !strings.Contains(filtered, "1 verified") {
		t.Errorf("the census under --unverified dropped the verified count:\n%s", filtered)
	}
}

// R204. Filtering inside the text branch renders everything under `--json` while every
// markdown assertion stays green — O28's silent ignore one level in, and the reason the
// selection sits before the output form rather than inside one arm of it.
//
// Fire alarm: move the query.SelectAlarms call below the `if c.JSON` return in queryAlarms.
// See test-Alarm.md.
func TestAlarmsJSONHonoursSelection(t *testing.T) {
	alarmProject(t)

	out := runAlarms(t, "--unverified", "--json")
	var got []map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON (%v): %q", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("--json rendered %d assessments; want 1 — the selection was dropped on the JSON path\n%s", len(got), out)
	}
	if got[0]["State"] != "unrecorded" {
		t.Errorf("--json kept the wrong assessment: %v", got[0])
	}
}

// R204. `--brief` is an output form and `--unverified` a filter, so under `--json` the brief
// rides on the assessment it belongs to rather than arriving as a second shape.
func TestBriefRidesOnTheAssessmentUnderJSON(t *testing.T) {
	alarmProject(t)

	out := runAlarms(t, "--unverified", "--brief", "--json")
	var got []query.BriefedAlarm
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("output is not JSON (%v): %q", err, out)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 briefed alarm, got %d\n%s", len(got), out)
	}
	if got[0].State != "unrecorded" {
		t.Errorf("state lost beside the brief: %q", got[0].State)
	}
	if !strings.Contains(got[0].Brief, "delete the guard entirely") {
		t.Errorf("brief does not carry its alarm's prose: %q", got[0].Brief)
	}
	if !strings.Contains(got[0].Brief, "subject_test.go") {
		t.Errorf("brief does not carry the Artifacts manifest's test file: %q", got[0].Brief)
	}
}

// R200. The brief's test files come from `design.md`'s Artifacts manifest, which is the link
// this project already maintains — not from a `**Code:**` field, which 7 of this repository's
// 20 test documents do not carry at all.
func TestBriefResolvesTestFilesThroughTheManifest(t *testing.T) {
	alarmProject(t)

	out := runAlarms(t, "--brief")
	if !strings.Contains(out, "Tests:       subject_test.go") {
		t.Errorf("manifest row not reflected in the brief:\n%s", out)
	}
	if !strings.Contains(out, "Design root: .") {
		t.Errorf("design root not expressed relative to the repository root:\n%s", out)
	}
	if strings.Count(out, "### test-Fixture.md") != 2 {
		t.Errorf("want one brief per selected alarm, got %d:\n%s", strings.Count(out, "### test-Fixture.md"), out)
	}
}

// R200. `..foo/design` is a legitimate descendant whose first component merely begins with two
// dots. Reporting it as an escape would make the brief say the root is unknown when it is
// perfectly locatable — a silent wrong answer, and the reason the test is not a prefix check.
func TestEscapesRootIsNotAPrefixCheck(t *testing.T) {
	for rel, want := range map[string]bool{
		"..":           true,
		"../sibling":   true,
		"..foo/design": false,
		"tool":         false,
		".":            false,
		"a/../b":       false,
	} {
		if got := escapesRoot(rel); got != want {
			t.Errorf("escapesRoot(%q) = %v, want %v", rel, got, want)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return capture(t, &os.Stdout, fn)
}

func capture(t *testing.T, stream **os.File, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prev := *stream
	*stream = w
	defer func() { *stream = prev }()
	fn()
	w.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
