// CRC: crc-CLI.md | Seq: seq-queue-item.md#1 | R282, R283
package cli

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/backup"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/pending"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/update"
	"github.com/zot/simple-dom/minispecsdom"
)

// CRC: crc-CLI.md | Seq: seq-queue-item.md#1.1 | R283
// runPending dispatches the trajectory verbs. It resolves the **repository root** and never a
// design root: the trajectory layer sits above every design root, the same reason `query
// carves` takes that path.
func (c *CLI) runPending(args []string) int {
	const usage = "Usage: minispec pending <add-item|start|finish|revert|replay>"
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}
	repoRoot, err := project.RepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	slot := backup.New(repoRoot)

	// The three link verbs report what they wrote rather than what the slot did, so they
	// return before the state-transition crank handle below. R282
	var op func() error
	switch args[0] {
	case "add-item":
		return c.runAddItem(repoRoot, args[1:])
	case "start":
		return c.runStart(repoRoot, args[1:])
	case "finish":
		return c.runFinish(repoRoot, args[1:])
	case "revert":
		op = slot.Revert
	case "replay":
		op = slot.Replay
	default:
		fmt.Fprintf(os.Stderr, "Unknown pending subcommand: %s\n", args[0])
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}

	before, _, _ := slot.State()
	if err := op(); err != nil {
		// A refusal is the message, not a stack trace: drift names the files and where
		// their backups are, an illegal operation names the state and what it accepts.
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	after, _, _ := slot.State()
	c.crankPending(args[0], before, after, slot)
	return 0
}

// CRC: crc-CLI.md | Seq: seq-queue-item.md#1 | R241, R253, R254, R256, R261, R282
// runAddItem mints an item, writes both sides of the link, and places the entry.
func (c *CLI) runAddItem(repoRoot string, args []string) int {
	const usage = "Usage: minispec pending add-item <title> --from <doc>#<part>|<gap-id> --status <text>\n" +
		"       [--after N] [--last] [--next] [--next-action <text>] [--nth N] [--skill <name>]\n" +
		"       --next places by intent; --next-action is the entry's `Next:` line\n" +
		"       every prose slot also has a --<slot>-file form — --title-file, --status-file,\n" +
		"       --next-action-file — read byte for byte, and refused alongside its inline twin"
	fs := flag.NewFlagSet("add-item", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	from := fs.String("from", "", "the part this item discharges, as <doc>#<part>, or the gap ID it repairs")
	skill := fs.String("skill", "", "the skill that runs this item, if any")
	status := fs.String("status", "", "the one-line status that follows the skill on the heading")
	statusFile := fs.String("status-file", "", "that status, read from a file byte for byte")
	titleFile := fs.String("title-file", "", "the title, read from a file byte for byte")
	nextAction := fs.String("next-action", "", "the entry's `Next:` line")
	nextActionFile := fs.String("next-action-file", "", "that next action, read from a file byte for byte")
	create := fs.Bool("create", false, "write the missing trajectory files with their preambles first")
	fs.Bool("next", false, "place it next to be worked")
	fs.Bool("last", false, "place it at the end of the queue (the default)")
	nth := fs.Int("nth", 0, "place it at position N")
	after := fs.Int("after", 0, "place it immediately after item #N")
	words, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		return 1
	}
	// **Which flags the caller actually gave, rather than which are non-zero.** A zero value
	// is a value: `--nth 0` and `--status ""` are instructions, and comparing against the zero
	// value makes them indistinguishable from absence. R254, R256
	gave := flagsGiven(fs)
	title := strings.Join(words, " ")
	if title != "" && gave["title-file"] {
		fmt.Fprintln(os.Stderr, "the title was given both as an argument and with --title-file; supply it one way")
		return 1
	}
	if gave["title-file"] {
		if title, err = readSlot(*titleFile); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
	}
	statusText, err := resolveSlot("status", gave["status"], gave["status-file"], *status, *statusFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	nextText, err := resolveSlot("next-action", gave["next-action"], gave["next-action-file"], *nextAction, *nextActionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if *from == "" || title == "" {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}
	// R255. The title is plain text and the verb supplies the emphasis, so a title wrapped end
	// to end in one `**…**` run would come out as `****…**`.
	if wrappedWhole(title) {
		fmt.Fprintln(os.Stderr, "the title is wrapped in **…**; give it as plain text — the verb supplies the emphasis, and a wrapped one would render as ****…**")
		return 1
	}
	// R253. The status lives *inside* the heading this verb mints, so defaulting it to the
	// empty string is how a required field quietly becomes optional.
	if statusText == "" {
		fmt.Fprintln(os.Stderr, "--status (or --status-file) is required: the status sentence sits inside the heading this verb writes, and a tool may decline to write a line but not two thirds of one")
		return 1
	}
	place, err := placeFrom(gave, *nth, *after)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	entry := parser.Entry{Title: title, Skill: *skill, Status: statusText, Next: nextText}
	// R273 — a gap source needs a design root; a part pointer does not. Resolved leniently
	// here and refused *inside* the verb, so a part pointer still works in a tree with no
	// design root at all — which is the layout `pending` exists for.
	gapsPath := ""
	if p, perr := c.getProject(); perr == nil {
		gapsPath = p.DesignMdPath()
	}
	// R333 — the scaffold, reached by being refused: the missing files are written with the
	// preambles the format mandates, and reported like any other write, before the item.
	if *create {
		made, cerr := pending.CreateTrajectory(repoRoot)
		if cerr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", cerr)
			return 1
		}
		if len(made) > 0 {
			fmt.Printf("  created  %s\n", describeWrites(repoRoot, made))
		}
	}
	created, err := pending.AddItem(repoRoot, gapsPath, *from, entry, place)
	if err != nil {
		reportQueueError(repoRoot, err) // R332
		return 1
	}
	if c.JSON {
		c.output(created)
		return 0
	}
	fmt.Printf("Minted #%d: %s\n", created.ID, created.Title)
	// R274 — the gap case writes nothing on the source side, and says so rather than
	// reporting a marker it did not write.
	if created.Part.Kind == minispecsdom.SourceGap {
		fmt.Printf("  gap      %s in %s, unchanged — nothing is written on the gap side\n", created.Part.Key, created.Part.Doc)
	} else {
		fmt.Printf("  part     %s#%s, marked **OPEN (#%d.)**\n", created.Part.Doc, created.Part.Key, created.ID)
	}
	fmt.Printf("  placed   position %d of %d\n", created.Position, created.Entries)
	fmt.Printf("  written  %s\n", describeWrites(repoRoot, created.Files))
	if created.ReusedID {
		// R234. The one person who could have carried this number into a chat or a commit
		// message is the one this reaches.
		fmt.Printf("  note     #%d returned to the pool from an unfinished attempt and is being reused\n", created.ID)
	}
	// R261. Naming the flags is the same information as "order it by intent", with the hand
	// taken out of it.
	if place.Kind == parser.PlaceLast && !gave["last"] {
		fmt.Println("\nPlaced last, which is the default. To place by intent: --next, --nth N, --after N, --last.")
	}
	return 0
}

// flagsGiven names the flags the caller actually passed, which is what every zero-valued flag
// here is read against. R254, R256
func flagsGiven(fs *flag.FlagSet) map[string]bool {
	gave := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { gave[f.Name] = true })
	return gave
}

// wrappedWhole reports a title that is one `**…**` run end to end. Emphasis *inside* a title
// is untouched: a title mentioning `**OPEN**` is legitimate and is not a wrapped one. R255
func wrappedWhole(s string) bool {
	if len(s) < 4 || !strings.HasPrefix(s, "**") || !strings.HasSuffix(s, "**") {
		return false
	}
	inner := s[2 : len(s)-2]
	return !strings.Contains(inner, "**")
}

// CRC: crc-CLI.md | Seq: seq-queue-item.md#1.9 | R254
// resolveSlot takes one prose slot supplied inline **or** from a file, and never both.
//
// **Refused rather than resolved by precedence.** A precedence rule is a silent choice between
// two things the caller believed were one. The file form is why the slot exists at all: a
// backtick inside a shell argument is command substitution, and when it fires the text is
// simply *gone* from what the tool receives with nothing reporting it.
func resolveSlot(name string, gaveInline, gaveFile bool, inline, path string) (string, error) {
	if gaveInline && gaveFile {
		return "", fmt.Errorf("--%s and --%s-file both name the same slot; supply it one way", name, name)
	}
	if !gaveFile {
		return inline, nil
	}
	return readSlot(path)
}

// readSlot reads a slot's file. Only the trailing newline a heredoc adds comes off — nothing is
// trimmed, re-quoted, or normalised, because every other byte is the caller's prose. R254
func readSlot(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(b), "\n"), nil
}

// CRC: crc-CLI.md | Seq: seq-queue-item.md#1.11 | R256
// placeFrom reads the placement intent off the four flags, which are mutually exclusive.
//
// Giving two is refused and the message names both, because a precedence between them would
// silently place the entry somewhere the caller did not ask for — in a gitignored file, where
// nothing would report it.
func placeFrom(gave map[string]bool, nth, after int) (parser.Place, error) {
	place := parser.Place{Kind: parser.PlaceLast}
	var given []string
	if gave["next"] {
		given = append(given, "--next")
		place = parser.Place{Kind: parser.PlaceNext}
	}
	// **Given, not non-zero.** `--nth 0` is an instruction the caller was specific about, and
	// ResolvePlace goes out of its way to refuse it *rather than clamp*.
	if gave["nth"] {
		given = append(given, "--nth")
		place = parser.Place{Kind: parser.PlaceNth, N: nth}
	}
	if gave["after"] {
		given = append(given, "--after")
		place = parser.Place{Kind: parser.PlaceAfter, N: after}
	}
	if gave["last"] {
		given = append(given, "--last")
	}
	if len(given) > 1 {
		return place, fmt.Errorf("%s are mutually exclusive; give one", strings.Join(given, ", "))
	}
	return place, nil
}

// CRC: crc-CLI.md | Seq: seq-queue-item.md#3 | R265, R282
// runStart opens an item, writing the current file's `## Active` section.
func (c *CLI) runStart(repoRoot string, args []string) int {
	const usage = "Usage: minispec pending start <item-number> [--context <text>|--context-file <path>]"
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	context := fs.String("context", "", "the active block's context")
	contextFile := fs.String("context-file", "", "the active block's context, read from a file byte for byte")
	pos, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		return 1
	}
	if len(pos) != 1 {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}
	id, err := strconv.Atoi(strings.TrimPrefix(pos[0], "#"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q is not an item number\n", pos[0])
		return 1
	}
	gave := flagsGiven(fs)
	text, err := resolveSlot("context", gave["context"], gave["context-file"], *context, *contextFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	started, err := pending.Start(repoRoot, id, text)
	if err != nil {
		reportQueueError(repoRoot, err) // R332
		return 1
	}
	if c.JSON {
		c.output(started)
		return 0
	}
	fmt.Printf("Started #%d: %s\n", started.ID, started.Title)
	fmt.Printf("  written  %s\n", describeWrites(repoRoot, started.Files))
	if text == "" {
		fmt.Println("\n  The active block carries its identity line only. Add the item's context there,")
		fmt.Println("  or hand it to the verb next time: --context-file <path> places it for you.")
	}
	return 0
}

// CRC: crc-CLI.md | Seq: seq-queue-item.md#2 | R244, R247, R262, R268, R280, R282
// runFinish completes an item across all four surfaces.
func (c *CLI) runFinish(repoRoot string, args []string) int {
	const usage = "Usage: minispec pending finish <item-number> --commit <hash> [--body <text>|--body-file <path>] [--discharged <text>|--discharged-file <path>] (--resolve|--no-resolve)"
	fs := flag.NewFlagSet("finish", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	commit := fs.String("commit", "", "the commit the work landed in")
	// R277 — the gap's answer to a part's LANDED marker, and never inferred.
	resolve := fs.Bool("resolve", false, "resolve the gap a gap-sourced item names")
	// R279, R281 — the decision needs two spellings or it is not a decision.
	noResolve := fs.Bool("no-resolve", false, "record that the item did not close its gap")
	body := fs.String("body", "", "the done entry's body")
	bodyFile := fs.String("body-file", "", "the done entry's body, read from a file byte for byte")
	discharged := fs.String("discharged", "", "what the item discharged besides its queue ID")
	dischargedFile := fs.String("discharged-file", "", "what the item discharged, read from a file byte for byte")
	pos, err := parseFlagsAnywhere(fs, args)
	if err != nil {
		return 1
	}
	gave := flagsGiven(fs)
	// R262. The pair every prose slot on `add-item` already uses, refused together.
	text, err := resolveSlot("body", gave["body"], gave["body-file"], *body, *bodyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	// R268. The same pair one field along: the tool owns the `#N`, the caller owns the rest.
	slot, err := resolveSlot("discharged", gave["discharged"], gave["discharged-file"], *discharged, *dischargedFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}
	if len(pos) != 1 || *commit == "" {
		fmt.Fprintln(os.Stderr, usage)
		return 1
	}
	id, err := strconv.Atoi(strings.TrimPrefix(pos[0], "#"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%q is not an item number\n", pos[0])
		return 1
	}
	// R280 — refused together, before the item is looked up: nothing preserves an ordering
	// between contradictory intents, and honouring one would be the tool deciding what the
	// caller meant.
	if *resolve && *noResolve {
		fmt.Fprintln(os.Stderr, "--resolve and --no-resolve are contradictory: pass the one that describes what happened.")
		return 1
	}
	opts := pending.FinishOpts{Body: text, Discharged: slot, DeclineResolve: *noResolve}
	if *resolve {
		p, perr := c.getProject()
		if perr != nil {
			fmt.Fprintf(os.Stderr, "--resolve needs a design root: %v\n", perr)
			return 1
		}
		// R277 — resolve the gap **the entry names**, in the document it named it in.
		opts.ResolveGap = gapResolver(update.New(p).ResolveGap, p.DesignMdPath(), repoRoot, id)
	}
	done, err := pending.Finish(repoRoot, id, *commit, opts)
	if err != nil {
		reportQueueError(repoRoot, err) // R332
		return 1
	}
	if c.JSON {
		c.output(done)
		return 0
	}
	reportFinished(repoRoot, done, text, *resolve || *noResolve)
	return 0
}

// CRC: crc-CLI.md | R277
// gapResolver binds the resolve act to the design root the **entry** named, refusing when that
// is not the root this invocation resolved.
//
// The project comes from the working directory; the ref came from wherever the item was
// created, and in a repository with two design roots those are free to differ. Resolving the
// same ID in the wrong `design.md` would close a gap nobody asked about and leave the intended
// one open, and **both documents would still validate**.
func gapResolver(resolve func(string) error, designMd, repoRoot string, id int) func(parser.PartRef) error {
	return func(gap parser.PartRef) error {
		if got := filepath.Join(repoRoot, filepath.FromSlash(gap.Doc)); got != designMd {
			return fmt.Errorf("#%d names %s in %s, but this design root is %s.\n"+
				"Run --resolve from that project, or resolve it there with `minispec update resolve-gap %s`.",
				id, gap.Key, gap.Doc, designMd, gap.Key)
		}
		return resolve(gap.Key)
	}
}

// CRC: crc-CLI.md | R247, R264, R277, R280, R281, R282
// reportFinished renders a completion. **Its own function so a test can reach it**: the
// notices below are decisions about what the caller is told, and a decision nothing can assert
// on is one that drifts silently.
func reportFinished(repoRoot string, done pending.Finished, body string, askedResolve bool) {
	fmt.Printf("Completed #%d\n", done.ID)
	for _, p := range done.Parts {
		fmt.Printf("  checked  %s#%s\n", p.Doc, p.Key)
	}
	if done.GapResolved {
		fmt.Printf("  resolved gap %s in %s\n", done.Gap.Key, done.Gap.Doc)
	}
	fmt.Printf("  written  %s\n", describeWrites(repoRoot, done.Files))
	// R280 — a resolve flag on an item that names no gap resolved nothing and said nothing:
	// a flag that silently does nothing is this project's signature defect wearing the shape
	// of success.
	if askedResolve && !done.HasGap {
		fmt.Printf("  note     the resolve flag did nothing: #%d names no gap, so there was none to close.\n", done.ID)
	}
	// R281 — a gap left open **by decision** and one left open by oversight are identical in
	// design.md, and this completion is the only place that difference is known.
	if done.HasGap && !done.GapResolved {
		fmt.Printf("  left open gap %s, by decision (--no-resolve)\n", done.Gap.Key)
	}
	// R247, R264. The body is authoring and stays the caller's; **placing** it is the tool's.
	// The crank handle below runs when the caller declined the flag — the fallback, never the
	// design.
	if body != "" {
		fmt.Println("  placed   the done entry's body, beneath the header, in the same write")
		return
	}
	fmt.Println("\nThe done entry carries its header only. Write the body beneath it in DONE.md:")
	fmt.Println("  enough to reconstruct the change without re-reading the code — gaps banked,")
	fmt.Println("  sources touched, and whatever a future reader would otherwise re-derive.")
	fmt.Println("\n  Or hand it to the verb instead: --body-file <path> places it for you, and then")
	fmt.Println("  there is no anchor to get wrong and no edit into a file with no diff.")
}

// pendingReport is the machine-readable form of what a slot operation did. R283
type pendingReport struct {
	Operation string   `json:"operation"`
	Before    string   `json:"before"`
	After     string   `json:"after"`
	Restored  []string `json:"restored"`
	BackupDir string   `json:"backup_dir"`
	// Anchor is where the worktree anchor lives — the *only* place it surfaces at all, since
	// it is deliberately outside the stash listing. R283
	Anchor string `json:"anchor"`
}

// CRC: crc-CLI.md | Seq: seq-backup.md#1 | R235, R283
// crankPending prints every change in full, and how to reach the tree as it stood before it.
func (c *CLI) crankPending(op string, before, after backup.State, slot *backup.Slot) {
	if c.JSON {
		c.output(pendingReport{
			Operation: op, Before: before.String(), After: after.String(),
			Restored: backup.Covered(), BackupDir: slot.Dir(), Anchor: project.SnapshotRef,
		})
		return
	}
	fmt.Printf("%sed: the trajectory files now hold the %s state.\n", op, stateWord(after))
	fmt.Printf("  slot: %s -> %s\n", before, after)
	for _, name := range backup.Covered() {
		fmt.Printf("  restored: %s\n", name)
	}
	fmt.Printf("  backups:  %s\n", slot.Dir())
	// **Hand over the commands, not the location.** Naming a ref the reader must work out
	// what to do with leaves them knowing what broke and not how to reach it.
	fmt.Printf("\nthe working tree was anchored before this change, at %[1]s.\n"+
		"  what has moved since:   git diff --name-status %[1]s\n"+
		"  restore as of then:     git checkout %[1]s -- <path>\n"+
		"  restore as of the commit beneath it: git checkout %[1]s^1 -- <path>\n"+
		"nothing in your working tree was altered — the anchor is for reading. Ignored\n"+
		"paths are absent from it by construction; the trajectory files above are the\n"+
		"backup slot's business, not the anchor's.\n", project.SnapshotRef)
}

func stateWord(s backup.State) string {
	if s == backup.Reverted {
		return "pre-change"
	}
	return "post-change"
}

// CRC: crc-CLI.md | R329
// describeWrites names each written file with what kind of file it is: the one tracked public
// document among a completion's four writes — the carve flip — is the write that still needs a
// commit, and until 2026-09-12 nothing in the report told it apart from the three gitignored
// files beside it. Measured 2026-08-18: `#18` was the first part discharged through the verb,
// and its flip sat uncommitted with nothing marking it. The tool reports and never stages
// (Bill, 2026-08-04); with no repository the names print bare.
func describeWrites(repoRoot string, files []string) string {
	g := project.NewGit(repoRoot)
	if !g.IsRepo() {
		return strings.Join(files, ", ")
	}
	// Each probe returns its zero value beside any error — a nil map from Ignored, false from
	// Tracked — and a nil map reads false, so a probe that fails already lands on the default
	// word below. Dropping the errors here is not ignoring them; it is saying where they go.
	ignored, _ := g.Ignored(files)
	out := make([]string, len(files))
	for i, f := range files {
		tracked, _ := g.Tracked(f)
		switch {
		case tracked:
			out[i] = f + " (tracked, uncommitted)"
		case ignored[f]:
			out[i] = f + " (ignored)"
		default:
			out[i] = f + " (untracked)"
		}
	}
	return strings.Join(out, ", ")
}

// CRC: crc-CLI.md | R332
// missingTrajectoryMessage is the crank handle for a queue verb refused because the layer
// is not there: which files, what creates them, and what git will do with them — the
// `.gitignore` question folded in, since `init --track-*` already answered it.
func missingTrajectoryMessage(repoRoot string, e *pending.MissingTrajectoryError) string {
	var b strings.Builder
	fmt.Fprintf(&b, "The trajectory layer is not here: %s missing beneath %s.\n\n", strings.Join(e.Files, ", "), repoRoot)
	b.WriteString("Create it with the lifecycle preambles the format mandates by running the\n")
	b.WriteString("add-item again with --create:\n\n")
	b.WriteString("    minispec pending add-item <title> --from <doc>#<part> --status <text> --create\n\n")
	track, err := project.LoadTrack(filepath.Join(repoRoot, project.ConfigDirName, project.RepoConfigName))
	switch {
	case err != nil:
		b.WriteString("Whether the queue is private or ships with the repository is `track` in\n.minispec/config.yaml, which could not be read; `minispec init --track-<value>` sets it.\n")
	case track == project.TrackPrivateTrajectory:
		b.WriteString("track is private-trajectory: init already wrote the .gitignore lines, so the three\nfiles will be ignored — private to this checkout.\n")
	default:
		fmt.Fprintf(&b, "track is %s: the three files will be tracked and ship with the repository.\n", track)
	}
	b.WriteString("\nNothing was written. A carve that does not exist yet is `minispec init carve <name>`.\n")
	return b.String()
}

// reportQueueError prints the crank handle when err is the missing-layer refusal, and the
// error itself otherwise.
func reportQueueError(repoRoot string, err error) {
	var m *pending.MissingTrajectoryError
	if errors.As(err, &m) {
		fmt.Fprint(os.Stderr, missingTrajectoryMessage(repoRoot, m))
		return
	}
	fmt.Fprintf(os.Stderr, "%v\n", err)
}
