// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1.4 | R152, R153, R155, R156, R157, R158, R160
package cli

import (
	"errors"
	"strings"
	"testing"
)

// The refusal messages are pure string logic, so they cost nothing to pin — and they
// are load-bearing in a way code is not: each clause closes a hole the others do not,
// and a later editor tightening the prose can delete one without breaking anything
// that compiles. These tests are what notices. See test-Bootstrap.md.

// R152, R153 — the exempt set, checked in both directions. The gated side matters more:
// a command added later must fail safe by being absent from the exempt list.
func TestOnlyVersionLikeCommandsAreExempt(t *testing.T) {
	for _, cmd := range []string{"init", "check-version", "help", "--help", "-h"} {
		if !exemptCommands[cmd] {
			t.Errorf("%q should run without a configuration", cmd)
		}
	}
	for _, cmd := range []string{"query", "update", "validate", "phase"} {
		if exemptCommands[cmd] {
			t.Errorf("%q should be gated: it needs a project", cmd)
		}
	}
	// `--version`/`-v` are intercepted by Run before the dispatcher, so listing them
	// here would name the wrong mechanism — and a reader who trusted it would look in
	// the wrong place when it broke. There is no bare `version` command.
	for _, cmd := range []string{"--version", "-v", "version"} {
		if exemptCommands[cmd] {
			t.Errorf("%q is in exemptCommands but never reaches it — dead entry that reads as live", cmd)
		}
	}
}

// Every dispatchable command must be known, or an unrecognised-command typo would be
// reported as a missing project instead.
func TestKnownCommandsCoversEveryDispatchedCommand(t *testing.T) {
	for _, cmd := range []string{"init", "check-version", "query", "update", "validate", "phase", "help"} {
		if !knownCommands[cmd] {
			t.Errorf("%q is dispatched but not in knownCommands", cmd)
		}
	}
	if knownCommands["frobnicate"] {
		t.Error("an unknown command is being treated as known")
	}
}

// R155 — the assent has to land on a stated location, not on "here".
func TestNoConfigMessageNamesTheAbsolutePath(t *testing.T) {
	const root = "/home/someone/work/thing"
	if got := noConfigMessage(root); !strings.Contains(got, root) {
		t.Errorf("refusal does not name the path it would make the repository root:\n%s", got)
	}
}

// R157 — the stop. Without it a weaker agent reads "check whether this is a code
// project", concludes yes, and runs init: exactly the road left deliberately unbuilt.
// This is the single clause most worth guarding, because deleting it breaks nothing
// visible.
func TestNoConfigMessageTellsTheAgentToStop(t *testing.T) {
	got := noConfigMessage("/tmp/x")
	if !strings.Contains(got, "REPORT AND WAIT") {
		t.Errorf("refusal is missing the stop instruction:\n%s", got)
	}
	if !strings.Contains(got, "Do not run `minispec init` on your own conclusion") {
		t.Errorf("refusal does not forbid acting on the agent's own conclusion:\n%s", got)
	}
}

// R156, R158, R160 — the agent checks the fact, hunts the negative finding, and knows
// what to say when the answer is no.
func TestNoConfigMessageCarriesEachClause(t *testing.T) {
	got := noConfigMessage("/tmp/x")
	for _, want := range []struct{ clause, why string }{
		{"FOLDER OF PROJECTS", "the negative finding is the valuable one (R158)"},
		{"is a code project", "the agent establishes the fact rather than asking (R156)"},
		{"from inside it", "what to do when the user declines (R160)"},
		{"minispec init --track-", "the command to run on yes"},
	} {
		if !strings.Contains(got, want.clause) {
			t.Errorf("refusal is missing %q — %s:\n%s", want.clause, want.why, got)
		}
	}
}

// R155 — the no-root refusal is the same shape, and must not invite the agent to
// create a marker just to make the search succeed.
func TestNoRootMessageForbidsCreatingAMarker(t *testing.T) {
	got := noRootMessage(errors.New("no repository root found (searched from /tmp/x)"))
	if !strings.Contains(got, "searched from /tmp/x") {
		t.Errorf("refusal drops the underlying error:\n%s", got)
	}
	if !strings.Contains(got, "Do not create anything") {
		t.Errorf("refusal does not forbid planting a root:\n%s", got)
	}
}

// R136 — three flags rather than one taking an argument, so a misspelled value is an
// unknown flag rather than a silently-accepted string.
func TestTrackFlagsCoverExactlyTheClosedSet(t *testing.T) {
	if len(trackFlags) != 3 {
		t.Fatalf("trackFlags has %d entries, want 3", len(trackFlags))
	}
	for flag, want := range trackFlags {
		if !strings.HasPrefix(flag, "--track-") {
			t.Errorf("%q is not a --track- flag", flag)
		}
		if string(want) != strings.TrimPrefix(flag, "--track-") {
			t.Errorf("%q maps to %q, want the matching value", flag, want)
		}
	}
}
