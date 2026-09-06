// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md | R179
package alarm

import (
	"errors"
	"sort"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

// CRC: crc-Alarm.md | R185
// State is what a recorded fault injection amounts to today. A closed set, so an
// unhandled case is a compile-time gap rather than a fifth behavior.
type State string

const (
	// Verified: pulled, and no injection site has changed since. R185
	Verified State = "verified"
	// Stale: pulled, but a site changed strictly after that date — the proof is
	// void. R179
	Stale State = "stale"
	// Unrecorded: anchored, but no verification is recorded. **Not** a claim that the
	// injection was never run — the documents cannot answer that, and asserting the
	// stronger version would be the tool manufacturing a finding. R186
	Unrecorded State = "unrecorded"
	// Unanchored: no `**Inject:**`, so the site lives only in prose and nothing can
	// check it. What a backfill cannot fix, because the person who pulled it knew the
	// file and six weeks later nobody does. R185
	Unanchored State = "unanchored"
	// Unresolvable: git cannot find a named symbol — the anchor itself has rotted,
	// which is the failure the field exists to prevent. R182
	Unresolvable State = "unresolvable"
	// Unchecked: there is no working tree, so the question could not be asked. Never
	// collapsed into Verified: a check that could not look must not return a clean
	// result. R184, R187
	Unchecked State = "unchecked"
)

// CRC: crc-Alarm.md | R185
// Assessment is one alarm resolved to a state, with the evidence for it.
type Assessment struct {
	Alarm parser.Alarm
	State State
	// Site names the injection site responsible for a Stale or Unresolvable verdict,
	// so the report says which one rather than leaving the reader to find it.
	Site string
	// Changed is the commit date that voided a Stale assessment.
	Changed string
}

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#1 | R179, R181, R182, R184, R187
// Assess resolves every alarm against what git reports.
//
// Pure over its two inputs — the alarms and a GitFacts — which is what lets the whole
// state table be pinned against a fake with no repository anywhere.
//
// The order of the checks is the cheap-first order, and it is load-bearing rather than
// an optimisation: the two states needing no git at all are decided before git is
// consulted, so cost is proportional to the *verified* population rather than to every
// alarm. A project adopting the convention has many unrecorded alarms and few pulled
// ones, so the expensive path is the small one.
func Assess(alarms []parser.Alarm, g project.GitFacts) []Assessment {
	out := make([]Assessment, 0, len(alarms))
	for _, a := range alarms {
		out = append(out, assessOne(a, g))
	}
	return out
}

func assessOne(a parser.Alarm, g project.GitFacts) Assessment {
	// step 1.2 — nothing to check against.
	if len(a.Sites) == 0 {
		return Assessment{Alarm: a, State: Unanchored}
	}
	// step 1.3 — anchored but never recorded as run.
	if !a.HasPulled() {
		return Assessment{Alarm: a, State: Unrecorded}
	}
	// step 1.4 — could not look.
	if !g.IsRepo() {
		return Assessment{Alarm: a, State: Unchecked}
	}
	// A site git could not answer for is remembered rather than returned at once: a
	// later site may be definitely stale, and a definite finding must not be masked by
	// an inconclusive one. Only if nothing worse turns up does the alarm report as
	// unchecked.
	var unchecked *Assessment
	for _, site := range a.Sites {
		when, err := g.LastChanged(site.File, site.Symbol)
		switch {
		case errors.Is(err, project.ErrUnresolvedSite):
			// step 1.6 — the anchor no longer points at anything.
			return Assessment{Alarm: a, State: Unresolvable, Site: site.String()}
		case errors.Is(err, project.ErrAmbiguousSite):
			// step 1.6 — the anchor points at more than one thing, which is the same
			// failure from the other side: nothing has been watched. The site carries
			// the count and the repair. R307
			return Assessment{Alarm: a, State: Unresolvable, Site: site.String() + " (" + err.Error() + ")"}
		case err != nil:
			if unchecked == nil {
				unchecked = &Assessment{Alarm: a, State: Unchecked, Site: site.String()}
			}
			continue
		}
		// step 1.7 — **strictly after**, and the strictness is the whole rule. The
		// normal workflow is to fix the code, pull the alarm and commit both
		// together, so the code's last change and the pull share a date for every
		// freshly recorded alarm. An inclusive comparison was written first and
		// marked two of three verified alarms stale the day they were written; a
		// check that fires on arrival is ignored, which is worse than the blind spot
		// it avoids — a change made later the same day, caught by the next change on
		// any later day. R181
		if when.After(a.Pulled) {
			return Assessment{Alarm: a, State: Stale, Site: site.String(),
				Changed: when.Format("2006-01-02")}
		}
	}
	if unchecked != nil {
		return *unchecked
	}
	return Assessment{Alarm: a, State: Verified}
}

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#2.5 | R185
// Census counts the assessments by state. The counts partition the input — every alarm
// lands in exactly one — because a census that double-counts or drops reads as a
// measurement while being an error.
func Census(as []Assessment) map[State]int {
	out := make(map[State]int, 6)
	for _, a := range as {
		out[a.State]++
	}
	return out
}

// CensusOrder is the order a census reads in: what is wrong first, what is merely
// unfinished after.
var CensusOrder = []State{Stale, Unresolvable, Unrecorded, Unanchored, Unchecked, Verified}

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#2.1 | R183
// Voided returns the assessments `validate` reports: the proofs that have expired and
// the anchors that no longer resolve.
//
// Unrecorded and unanchored alarms are deliberately absent. Both are worth knowing and
// neither is closable soon — a project adopting the convention carries many, the counts
// fall slowly, and a line reporting a non-zero number every run for months is the
// recurring nag this project distinguishes from a gripe you can discharge. They belong
// to `query alarms`, which is asked rather than emitted.
func Voided(as []Assessment) []Assessment {
	var out []Assessment
	for _, a := range as {
		if a.State == Stale || a.State == Unresolvable {
			out = append(out, a)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Alarm.Doc != out[j].Alarm.Doc {
			return out[i].Alarm.Doc < out[j].Alarm.Doc
		}
		return out[i].Alarm.Test < out[j].Alarm.Test
	})
	return out
}
