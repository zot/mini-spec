// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#2.6 | R199, R200, R203
package query

import (
	"strings"
	"testing"

	"github.com/zot/minispec/internal/alarm"
	"github.com/zot/minispec/internal/parser"
)

func assessed(doc string, state alarm.State) alarm.Assessment {
	return alarm.Assessment{Alarm: parser.Alarm{Doc: doc, Test: string(state) + " one"}, State: state}
}

// R199. `--unverified` selects every state that carries a decision, and the closed set is
// what makes that "everything but verified" rather than a list to keep in step: a sixth state
// added to Alarm is selected here without anyone remembering to come back.
func TestSelectAlarmsKeepsEveryDecidingState(t *testing.T) {
	in := []alarm.Assessment{
		assessed("a.md", alarm.Verified),
		assessed("b.md", alarm.Stale),
		assessed("c.md", alarm.Unrecorded),
		assessed("d.md", alarm.Unanchored),
		assessed("e.md", alarm.Unresolvable),
		assessed("f.md", alarm.Unchecked),
	}
	got := SelectAlarms(in, true)
	if len(got) != len(in)-1 {
		t.Fatalf("selected %d of %d; want everything but the verified one", len(got), len(in))
	}
	for _, a := range got {
		if a.State == alarm.Verified {
			t.Errorf("verified alarm survived the filter: %v", a)
		}
	}
	if len(SelectAlarms(in, false)) != len(in) {
		t.Errorf("the bare form must stay the full census")
	}
}

// R200, R203. Query holds the manifest and Alarm holds the wording. A document the manifest
// does not mention resolves to no files, which the brief states rather than hides — so the
// count of briefs stays equal to the count of selected alarms whatever the manifest says.
func TestAlarmBriefsCoverEverySelectedAlarm(t *testing.T) {
	in := []alarm.Assessment{assessed("test-Known.md", alarm.Stale), assessed("test-Unmapped.md", alarm.Stale)}
	arts := []parser.Artifact{{DesignFile: "test-Known.md", CodeFiles: []parser.CodeFile{{Path: "internal/known/known_test.go"}}}}

	got := AlarmBriefs(in, "tool", arts)
	if len(got) != len(in) {
		t.Fatalf("minted %d briefs for %d alarms", len(got), len(in))
	}
	if !strings.Contains(got[0].Brief, "internal/known/known_test.go") {
		t.Errorf("mapped document lost its test file: %q", got[0].Brief)
	}
	if !strings.Contains(got[1].Brief, "maps test-Unmapped.md to no code file") {
		t.Errorf("unmapped document did not state the absence: %q", got[1].Brief)
	}
}
