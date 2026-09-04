// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#2.6 | R199
package query

import (
	"github.com/zot/minispec/internal/alarm"
	"github.com/zot/minispec/internal/parser"
)

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#2.6 | R199
// SelectAlarms narrows the census to the states that carry a decision.
//
// **The list narrows and the census does not.** Callers compute the closing count over the
// whole population either way, so `--unverified` is the full census minus the repetitions
// of *nothing to do here* and minus nothing else. A count filtered along with its list
// answers *how many are wrong* and silently drops *out of how many* — the shape of the
// `| tail -2` workaround this flag replaces, one level in.
func SelectAlarms(assessments []alarm.Assessment, unverified bool) []alarm.Assessment {
	if !unverified {
		return assessments
	}
	out := make([]alarm.Assessment, 0, len(assessments))
	for _, a := range assessments {
		if a.State != alarm.Verified {
			out = append(out, a)
		}
	}
	return out
}

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#3 | R204
// BriefedAlarm is one assessment carrying the brief a delegated re-pull is spawned with.
//
// The assessment is embedded rather than copied field by field, so under `--json` the brief
// marshals as one more field beside the state it belongs to (R204) — the same object,
// answering one more question, rather than a second shape to reconcile.
type BriefedAlarm struct {
	alarm.Assessment
	Brief string
}

// CRC: crc-Query.md | Seq: seq-alarm-freshness.md#3.2 | R200
// AlarmBriefs mints one brief per assessment.
//
// Query holds the manifest facts and Alarm holds the wording: this resolves each alarm's
// document to the test files `design.md` maps it to and hands them over, and composes no
// prose of its own. A document the manifest does not mention resolves to no files, which
// the brief states rather than hides (R203).
func AlarmBriefs(assessments []alarm.Assessment, designRoot string, artifacts []parser.Artifact) []BriefedAlarm {
	byDoc := make(map[string][]string, len(artifacts))
	for _, art := range artifacts {
		paths := make([]string, 0, len(art.CodeFiles))
		for _, cf := range art.CodeFiles {
			paths = append(paths, cf.Path)
		}
		byDoc[art.DesignFile] = paths
	}
	out := make([]BriefedAlarm, 0, len(assessments))
	for _, a := range assessments {
		out = append(out, BriefedAlarm{
			Assessment: a,
			Brief:      alarm.Brief(a, designRoot, byDoc[a.Alarm.Doc]),
		})
	}
	return out
}
