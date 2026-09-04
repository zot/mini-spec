// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3 | R200
package alarm

import (
	"fmt"
	"path"
	"strings"

	"github.com/zot/minispec/internal/parser"
)

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3.8 | R202
// contract is the one invariant a brief restates, and it is the load-bearing half.
//
// The protocol — baseline first, inject, diff, re-run, restore, prove the restore clean —
// is invariant too, and lives in the agent definition that runs it rather than here:
// minting it per alarm would repeat some 350 words against a census whose whole purpose is
// to be small. This does not get that treatment, because a puller that returns a verdict
// has pre-empted the judgment the delegation exists to keep. Measured: three times inside
// one item the naive verdict would have been wrong.
const contract = `Return five things and nothing else: the command you ran, its output before the
injection, the diff you applied, its output after, and the diff after you restored.
Do not report whether the alarm rang, whether the test is adequate, or what should
happen next — that judgment belongs to the reader of your evidence, and a verdict
pre-empts it.`

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3 | R200, R201, R202, R203
// Brief renders the spawn prompt for one delegated re-pull.
//
// Pure over its three inputs, so the whole shape is testable with no repository, no
// manifest and no agent. designRoot is expressed **relative to the repository root** by
// the caller — a delegated puller runs in its own worktree, where this checkout's
// absolute path resolves to nothing, or worse, resolves to the tree the worktree exists
// to protect. testFiles come from `design.md`'s Artifacts manifest, which Query holds.
//
// Every part is emitted even when its material is missing (R203): a dropped line reads as
// an oversight in this function, while a stated absence reads as what it is, and the count
// of briefs stays equal to the count of selected alarms.
func Brief(a Assessment, designRoot string, testFiles []string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "### %s — %s   [%s]\n\n", a.Alarm.Doc, a.Alarm.Test, a.State)
	b.WriteString("Re-pull this fire alarm and report evidence, never a verdict.\n\n")

	fmt.Fprintf(&b, "Design root: %s\n", designRootLine(designRoot))

	// R201 — files and directories, never a command. This tool knows document structure
	// and not build systems, and a guessed command that fails to build produces output a
	// hurried reader scores as *rang*.
	fmt.Fprintf(&b, "Sites:       %s\n", sitesLine(a.Alarm.Sites))
	fmt.Fprintf(&b, "Tests:       %s\n", filesLine(a.Alarm.Doc, testFiles))
	fmt.Fprintf(&b, "Directories: %s\n\n", dirsLine(testFiles))

	// R200 — verbatim. The prose *is* the injection, so a paraphrase or a first line is a
	// different injection. ParseTestDoc already folded the continuation lines into one body.
	fmt.Fprintf(&b, "Fire alarm, verbatim from %s:\n", a.Alarm.Doc)
	for _, line := range strings.Split(a.Alarm.Prose, "\n") {
		fmt.Fprintf(&b, "> %s\n", line)
	}

	b.WriteString("\n" + contract + "\n")
	return b.String()
}

// R203 — an unanchored alarm has no sites, and saying so tells the reader the next move is
// to write the anchor rather than to spawn anyone.
func sitesLine(sites []parser.AlarmSite) string {
	if len(sites) == 0 {
		return "none — this alarm is unanchored, so there is nothing to inject. It needs an\n" +
			"             **Inject:** field written before it can be delegated at all."
	}
	parts := make([]string, 0, len(sites))
	for _, s := range sites {
		parts = append(parts, s.String())
	}
	return strings.Join(parts, ", ") + "   (edit only these)"
}

// R203 — a document the Artifacts manifest maps to nothing has no test files, and the
// repair is a manifest row rather than a guess made here.
func filesLine(doc string, files []string) string {
	if len(files) == 0 {
		return "none — design.md's Artifacts manifest maps " + doc + " to no code file, so\n" +
			"             there is no test to run until that row is written."
	}
	return strings.Join(files, ", ")
}

func dirsLine(files []string) string {
	dirs := dirsOf(files)
	if len(dirs) == 0 {
		return "none — no test files to take them from."
	}
	return strings.Join(dirs, ", ")
}

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3.3 | R200
// dirsOf reduces test files to the distinct directories holding them, in first-seen order.
//
// Deterministic order rather than map order, so two runs over the same manifest produce the
// same brief and a diff between them means something. Slash-separated throughout: these are
// manifest paths, which are written with `/` whatever the host does.
func dirsOf(files []string) []string {
	seen := make(map[string]bool, len(files))
	out := make([]string, 0, len(files))
	for _, f := range files {
		d := path.Dir(f)
		if seen[d] {
			continue
		}
		seen[d] = true
		out = append(out, d)
	}
	return out
}

// CRC: crc-Alarm.md | Seq: seq-alarm-freshness.md#3.1 | R200, R203
// designRootLine states an unresolvable root rather than inventing one.
//
// An absolute path would send a delegated puller at the original checkout, where its
// injection lands in the working copy the worktree exists to protect; a confident `.` would
// be a guess wearing the same clothes as a fact.
func designRootLine(root string) string {
	if root == "" {
		return "unknown — the repository root could not be resolved, so nothing here can say\n" +
			"             where these paths sit in your checkout. Locate them before injecting."
	}
	return root + " — every path below is relative to it, and it is relative to the\n" +
		"             repository root. Resolve them inside your own checkout."
}
