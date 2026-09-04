// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3 | R200, R201, R202, R203
package alarm

import (
	"slices"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/parser"
)

func briefFixture(prose string, sites ...parser.AlarmSite) Assessment {
	return Assessment{
		Alarm: parser.Alarm{Doc: "test-Carve.md", Test: "part lines survive a rewrite", Prose: prose, Sites: sites},
		State: Stale,
	}
}

// R200. The prose *is* the injection, so a paraphrase or a first line is a different
// injection. ParseTestDoc folds an alarm's wrapped continuation lines into one body; this
// pins that the brief carries all of what it folded.
//
// Fire alarm: emit only the first line of `Prose`. The truncated brief still reads as a
// complete, well-formed brief, which is why the assertion is on the last line rather than
// on the shape. See test-Alarm.md.
func TestBriefQuotesProseWhole(t *testing.T) {
	prose := "delete the guard\nand then the second line\nand the third"
	got := Brief(briefFixture(prose), "tool", []string{"internal/parser/carve_test.go"})
	for _, line := range strings.Split(prose, "\n") {
		if !strings.Contains(got, "> "+line) {
			t.Errorf("brief dropped prose line %q\n%s", line, got)
		}
	}
}

// R203. An unanchored alarm has no sites and an unmapped document has no test files. A
// missing line reads as an oversight in the generator; a stated absence reads as what it is,
// and keeps the number of briefs equal to the number of selected alarms.
//
// Fire alarm: skip a line when its list is empty. The brief that results is shorter and
// entirely well-formed, so nothing but this assertion notices. See test-Alarm.md.
func TestBriefStatesAnAbsentPart(t *testing.T) {
	got := Brief(briefFixture("break it"), "tool", nil)
	for _, want := range []string{"Sites:", "Tests:", "Directories:"} {
		if !strings.Contains(got, want) {
			t.Errorf("brief dropped the %q line entirely\n%s", want, got)
		}
	}
	if !strings.Contains(got, "unanchored") {
		t.Errorf("no sites, but the brief does not say the alarm is unanchored\n%s", got)
	}
	if !strings.Contains(got, "maps test-Carve.md to no code file") {
		t.Errorf("no test files, but the brief does not name the empty manifest row\n%s", got)
	}
}

// R200. A delegated puller runs in its own worktree. An absolute design root is not a
// cosmetic flaw: it sends the agent at the *original* checkout, where its injection lands
// in the working copy the worktree exists to protect.
//
// Fire alarm: pass the absolute design root through. See test-Alarm.md.
func TestBriefNamesNoAbsolutePath(t *testing.T) {
	got := Brief(briefFixture("break it", parser.AlarmSite{File: "internal/parser/carve.go", Symbol: "parsePartLine"}),
		"tool", []string{"internal/parser/carve_test.go"})
	if !strings.Contains(got, "Design root: tool ") {
		t.Errorf("design root not named as a relative path\n%s", got)
	}
	for _, line := range strings.Split(got, "\n") {
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "/") && strings.Count(field, "/") > 1 {
				t.Errorf("brief carries an absolute path %q, which resolves outside a worktree\n%s", field, got)
			}
		}
	}
}

// R200, R203. An unresolvable repository root is stated rather than papered over with a
// confident `.`, which would be a guess wearing the same clothes as a fact.
func TestBriefStatesAnUnknownRoot(t *testing.T) {
	got := Brief(briefFixture("break it"), "", nil)
	if !strings.Contains(got, "Design root: unknown") {
		t.Errorf("empty root not stated as unknown\n%s", got)
	}
}

// R201, R202. The contract is the half that decides whether delegating was worth doing, and
// a brief that named a command would invite the failure `SKILL.md` names — an injection that
// breaks the build teaching nothing, because the test never ran.
func TestBriefCarriesTheContractAndNoCommand(t *testing.T) {
	got := Brief(briefFixture("break it", parser.AlarmSite{File: "internal/parser/carve.go", Symbol: "parsePartLine"}),
		"tool", []string{"internal/parser/carve_test.go"})
	for _, want := range []string{"never a verdict", "the command you ran", "before the\ninjection", "after you restored"} {
		if !strings.Contains(got, want) {
			t.Errorf("brief is missing contract clause %q\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"go test", "make ", "npm ", "pytest"} {
		if strings.Contains(got, forbidden) {
			t.Errorf("brief mints the command %q; it may name files and directories only\n%s", forbidden, got)
		}
	}
}

// R200. Directories are distinct and in first-seen order, so two runs over the same manifest
// produce the same brief and a diff between them means something.
func TestBriefDirsAreDistinctAndOrdered(t *testing.T) {
	got := dirsOf([]string{"internal/alarm/a_test.go", "internal/query/q_test.go", "internal/alarm/b_test.go"})
	want := []string{"internal/alarm", "internal/query"}
	if !slices.Equal(got, want) {
		t.Fatalf("dirsOf: got %v, want %v", got, want)
	}
}
