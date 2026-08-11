// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1 | R152
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/zot/minispec/internal/project"
)

// knownCommands is every command the dispatcher accepts. It sits beside the exempt
// list deliberately: adding a command means touching this file, where forgetting to
// decide whether it is exempt is visible rather than silent.
var knownCommands = map[string]bool{
	"init": true, "check-version": true, "query": true, "update": true,
	"validate": true, "phase": true, "help": true, "-h": true, "--help": true,
}

// exemptCommands need no mini-spec files at all and so run before a project exists.
// Everything not listed here is gated, so a command added later fails safe.
//
// check-version is here deliberately: a skill's first instruction is to run it, and a
// version check that refused in an uninitialized project would leave an agent unable to
// establish whether its tool matches its skill *before* being told what to do about
// that. R152, R153
//
// `--version` and `-v` are not listed, and their absence is not an oversight: Run
// intercepts them before the dispatcher, so they never reach this map. Listing them
// would read as the mechanism that makes them work, and a reader relying on that would
// be looking in the wrong place. There is no bare `version` command.
var exemptCommands = map[string]bool{
	"help":          true,
	"-h":            true,
	"--help":        true,
	"check-version": true, // R153 — a skill's first instruction must work uninitialized
	"init":          true,
}

// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1 | R147, R148, R149, R154
// gate runs before any command that does more than report a version: it requires a
// repository configuration and a track value that agrees with what git reports.
//
// Returns an exit code and true when the command must not proceed.
func (c *CLI) gate(cmd string) (int, bool) {
	// step 1.1
	if exemptCommands[cmd] {
		return 0, false
	}

	// step 1.2
	repoRoot, err := project.RepoRoot()
	if err != nil {
		fmt.Fprint(os.Stderr, noRootMessage(err))
		return 1, true
	}

	// steps 1.3 and 1.4
	cfgPath := project.RepoConfigPath(repoRoot)
	if _, statErr := os.Stat(cfgPath); statErr != nil {
		fmt.Fprint(os.Stderr, noConfigMessage(repoRoot))
		return 1, true
	}

	// steps 1.5 and 1.6
	track, err := project.LoadTrack(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1, true
	}

	// steps 1.7 through 1.11
	git := project.NewGit(repoRoot)
	if ms := track.Verify(git, repoRoot); len(ms) > 0 {
		fmt.Fprint(os.Stderr, project.MismatchReport(track, ms))
		return 1, true
	}

	// step 1.12 — preferences are reported, never refused over.
	if pref := project.PreferenceReport(git, repoRoot); pref != "" {
		fmt.Fprint(os.Stderr, pref)
	}
	return 0, false
}

// trackFlags maps each mandatory flag to the value it records. Three flags rather than
// one flag taking an argument, so a misspelled value is an unknown flag rather than a
// silently-accepted string. R136
var trackFlags = map[string]project.TrackValue{
	"--track-none":               project.TrackNone,
	"--track-private-trajectory": project.TrackPrivateTrajectory,
	"--track-all":                project.TrackAll,
}

// CRC: crc-CLI.md | Seq: seq-bootstrap.md#2 | R136, R141, R143
// runInit parses the creation and repair verbs and prints what changed.
//
// Nothing calls this implicitly. The tool never invokes init itself and no skill tells
// an agent to run it, so the only paths here are a direct user request or the
// no-configuration refusal — whose whole purpose is obtaining that assent.
func (c *CLI) runInit(args []string) int {
	var track project.TrackValue
	var repair bool
	var chosen []string

	for _, a := range args {
		switch v, isTrackFlag := trackFlags[a]; {
		case isTrackFlag:
			track = v
			chosen = append(chosen, a)
		case a == "--repair":
			repair = true
		default:
			fmt.Fprintf(os.Stderr, "init: unrecognized argument %q\n", a)
			return 1
		}
	}
	// Two flags are as wrong as none: both mean the caller has not decided.
	if len(chosen) > 1 {
		fmt.Fprintf(os.Stderr,
			"init: %s conflict — choose exactly one.\n", strings.Join(chosen, " and "))
		return 1
	}

	repoRoot, err := project.RepoRoot()
	if err != nil {
		fmt.Fprint(os.Stderr, noRootMessage(err))
		return 1
	}

	result, err := project.RunInit(
		project.InitOptions{RepoRoot: repoRoot, Track: track, Repair: repair}, project.NewGit(repoRoot))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	// R141 — the agent did not write these files, so it is told exactly what changed
	// rather than left to infer it.
	fmt.Printf("minispec init --track-%s%s\n", track, repairSuffix(repair))
	fmt.Print(result.Report())
	return 0
}

func repairSuffix(repair bool) string {
	if repair {
		return " --repair"
	}
	return ""
}

// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1.4 | R155, R156, R157, R158, R159, R160, R169
// noConfigMessage is the no-configuration crank handle.
//
// It is written for the agent, and the agent writes for the human: a verbatim script
// cannot adapt to what the agent found, and would still say "if this is a code project"
// after the agent had established that it plainly is, or plainly is not.
func noConfigMessage(repoRoot string) string {
	var b strings.Builder
	b.WriteString("It doesn't look like this is a minispec project.\n\n")
	// R155 — the assent has to land on a stated location, not on "here".
	fmt.Fprintf(&b, "There is no %s under %s.\n\n",
		project.ConfigDirName+"/"+project.RepoConfigName, repoRoot)
	b.WriteString("AGENT: do these three things, in order.\n\n")
	// R156 — whether this holds code is a fact, so the agent checks rather than asking.
	fmt.Fprintf(&b, "  1. Look at %s and decide whether it is a code project.\n", repoRoot)
	// R158 — the negative finding is the valuable one.
	b.WriteString("  2. If it instead looks like a FOLDER OF PROJECTS, say so and name the\n" +
		"     likely candidates. That is the more useful answer: it means the command\n" +
		"     was run from above the project, and initializing here would plant a\n" +
		"     repository root in the wrong place.\n")
	// R159 — the gist, composed by the agent.
	b.WriteString("  3. Ask the user, in plain words, whether they want this to be a\n" +
		"     mini-spec project, and whether the work queue should stay private or ship\n" +
		"     with the repository. Compose that question yourself from what you found in\n" +
		"     step 1 — do not repeat this text at them.\n\n")
	// R157 — the stop. Without it a weaker agent concludes yes and builds the road
	// that was deliberately left unbuilt.
	b.WriteString("THEN REPORT AND WAIT. Do not run `minispec init` on your own conclusion:\n" +
		"the user's answer is the only thing that authorises it.\n\n")
	b.WriteString("On yes, the command is:\n\n")
	b.WriteString("    minispec init --track-<none|private-trajectory|all>\n\n")
	// R160
	b.WriteString("If the answer is no, they are most likely outside their project — the\n" +
		"search only walks up, so a project below would not have been found. Run the\n" +
		"command again from inside it.\n\n")
	// R169
	b.WriteString("(A project that already uses mini-spec meets this once: `track` is new, and\n" +
		"this refusal is how it gets asked.)\n")
	return b.String()
}

// CRC: crc-CLI.md | Seq: seq-bootstrap.md#1.2 | R155
// noRootMessage is the refusal when no repository root could be found at all. It is
// the same shape as the no-configuration refusal and for the same reason: the tool
// detects, the agent inspects, the human decides.
func noRootMessage(err error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%v\n\n", err)
	b.WriteString("AGENT: report this and wait. Do not create anything to make the search\n" +
		"succeed — that would plant a repository root wherever the command happened to\n" +
		"be run. If the user meant to work in a project, run the command from inside it.\n")
	return b.String()
}
