// CRC: crc-Pending.md | Seq: seq-queue-item.md#1 | R241, R242, R243, R246, R248
// Package pending writes the trajectory files: `pending add-item` creates a queue entry and the
// part line it points at, `pending start` opens the item in the current file, and `pending
// finish` moves the entry to the done file and checks off the parts the item discharged.
//
// An **orchestrator** rather than a reader. It owns the *order* of the writes and delegates
// every shape to whoever already owns it: the queue entry, the `## Active` region and the done
// entry to internal/parser's adapters over the dependency's readers, the part line to the carve
// adapter, the format itself to the skill.
//
// **It is the backup slot's first production caller.** R248
package pending

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/zot/minispec/internal/backup"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/simple-dom/minispecsdom"
)

// The queue files, mandated at the repository root by the skill's format reference.
const (
	pendingFile = "PENDING.md"
	currentFile = "CURRENT.md"
	doneFile    = "DONE.md"
)

// Created is what add-item did, for the crank handle. R282
type Created struct {
	ID       int            `json:"id"`
	Part     parser.PartRef `json:"part"`
	Title    string         `json:"title"`
	Files    []string       `json:"files"`
	Position int            `json:"position"` // 1-based, among the entries. R256
	Entries  int            `json:"entries"`  // how many positions the queue offered. R261
	ReusedID bool           `json:"reused_id,omitempty"`
}

// CRC: crc-Pending.md | Seq: seq-queue-item.md#1.10 | R256, R257, R258
// resolvePlacement turns the caller's intent into a position.
//
// **`--next` is the only intent that reads a second file**, and it reads one boolean out of it:
// `## Active` holding something other than the placeholder. The pending file's own rule — *the
// top item is active* — turns that boolean into position 1 or position 2.
func resolvePlacement(repoRoot, pendingPath string, place parser.Place) (pos, total int, err error) {
	switch place.Kind {
	case parser.PlaceNext:
		n, err := resolveNext(repoRoot)
		if err != nil {
			return 0, 0, err
		}
		pos, total, err := parser.ResolvePlace(pendingPath, parser.Place{Kind: parser.PlaceNth, N: n})
		if err != nil {
			return 0, 0, fmt.Errorf("--next resolved to position %d: %w", n, err)
		}
		return pos, total, nil
	case parser.PlaceNth:
		// The same refusal `--next` embodies from the other side: position 1 is the active
		// item's slot. Both repairs are named, because preempting the active item is a real
		// thing to want and must not read as forbidden. R258
		busy, err := inProgress(repoRoot)
		if err != nil {
			return 0, 0, err
		}
		if place.N == 1 && busy {
			return 0, 0, fmt.Errorf("--nth 1 while a step is in progress would displace the active item; use --next to queue behind it, or park the active item first if you mean to preempt it")
		}
	}
	return parser.ResolvePlace(pendingPath, place)
}

// CRC: crc-Pending.md | Seq: seq-queue-item.md#1.10.2 | R257
// resolveNext is *next to be worked* rather than *position 1*: with a step in progress the
// top item is the active one, so next is position 2.
func resolveNext(repoRoot string) (int, error) {
	busy, err := inProgress(repoRoot)
	if err != nil {
		return 0, err
	}
	if busy {
		return 2, nil
	}
	return 1, nil
}

// inProgress asks the current file whether a step is being worked. R257
func inProgress(repoRoot string) (bool, error) {
	return parser.ActiveInProgress(filepath.Join(repoRoot, currentFile))
}

// Finished is what finish did, for the crank handle. R282
type Finished struct {
	ID    int              `json:"id"`
	Parts []parser.PartRef `json:"parts"`
	Files []string         `json:"files"`
	// Gap is the gap a gap-sourced item named, and HasGap whether it named one. Reported
	// whether or not it was resolved, because the report names the gap either way. R278
	Gap    parser.PartRef `json:"gap"`
	HasGap bool           `json:"has_gap"`
	// GapResolved records that `--resolve` was given and the resolve succeeded. R277
	GapResolved bool `json:"gap_resolved"`
}

// Started is what an opening wrote. R265
type Started struct {
	ID    int      `json:"id"`
	Title string   `json:"title"`
	Files []string `json:"files"`
}

// findEntry reads the pending file and returns the entry for id, refusing an absent one.
func findEntry(repoRoot string, id int) (parser.QueueEntry, error) {
	entries, err := parser.PendingEntries(filepath.Join(repoRoot, pendingFile))
	if err != nil {
		return parser.QueueEntry{}, err
	}
	for _, e := range entries {
		if e.ID == id {
			return e, nil
		}
	}
	return parser.QueueEntry{}, fmt.Errorf("no entry for #%d in %s", id, pendingFile)
}

// CRC: crc-Pending.md | Seq: seq-queue-item.md#3 | R265, R267
// Start opens an item: the identity line composed here, the caller's context placed beneath it.
//
// **The ID and title are read from the queue entry rather than taken from the caller**, so the
// active block and the done header a completion later mints cannot disagree about what was
// worked. It runs inside the slot like the other two and has the strongest claim of the three
// to be there: it writes before any record of the work exists anywhere else.
func Start(repoRoot string, id int, context string) (Started, error) {
	if err := missingTrajectory(repoRoot); err != nil { // R332
		return Started{}, err
	}
	out := Started{ID: id}
	entry, err := findEntry(repoRoot, id)
	if err != nil {
		return out, err
	}
	out.Title = entry.Title
	line := fmt.Sprintf("`#%d` — %s", id, entry.Title)
	err = backup.New(repoRoot).Record(func() error {
		// Seq: seq-queue-item.md#3.3.2
		return parser.SetActive(filepath.Join(repoRoot, currentFile), line, context)
	})
	if err != nil {
		return out, err
	}
	out.Files = []string{currentFile}
	return out, nil
}

// CRC: crc-Pending.md | Seq: seq-queue-item.md#1 | R241, R243
// AddItem mints the next item ID and writes **both sides of the link in one invocation**.
//
// Assignment and the write that records it are one act, which is what keeps R190's `max()`
// the whole truth: a number is in a document the moment it exists.
func AddItem(repoRoot, gapsPath, from string, text parser.Entry, place parser.Place) (Created, error) {
	var out Created
	// R332 — the layer must be there before anything is minted; the refusal is typed so the
	// caller can say how to create it rather than relay a raw open error.
	if err := missingTrajectory(repoRoot); err != nil {
		return out, err
	}
	// Seq: seq-queue-item.md#1.12 | R271, R272, R273
	// A gap ID or a part pointer, decided by shape and needing no flag: a letter and digits
	// carries no `/`, no `#` and no `.md`, so the two cannot collide.
	gap, err := gapSource(repoRoot, gapsPath, from)
	if err != nil {
		return out, err
	}
	if gap.Kind == minispecsdom.SourceGap {
		return mintAndPlace(repoRoot, gap, text, place)
	}
	part, err := splitPointer(from)
	if err != nil {
		return out, err
	}
	carvePath := filepath.Join(repoRoot, filepath.FromSlash(part.Doc))
	if _, err := os.Stat(carvePath); err != nil {
		return out, fmt.Errorf("no document %s beneath %s: %w", part.Doc, repoRoot, err)
	}
	// Seq: seq-queue-item.md#1.3 | R243
	// Refuse rather than guess, naming what was looked for and where.
	carve, err := parser.ReadCarve(carvePath, part.Doc)
	if err != nil {
		return out, err
	}
	if !carve.HasStatus {
		return out, fmt.Errorf("%s has no status block, so it holds no part %s", part.Doc, part.Key)
	}
	found := false
	for _, p := range carve.Parts {
		if !p.Keyed() || p.Key() != part.Key {
			continue
		}
		found = true
		// Seq: seq-queue-item.md#1.4 | R243, R330
		// A part records exactly one item. Re-queuing it silently would leave the older
		// pointer resolving to work it never described, and both entries would be
		// individually well-formed, so nothing downstream could detect it.
		//
		// **Unless the ID is a reverted attempt's.** While the slot holds a reverted attempt
		// the part still reads `REVERTED (#N.)`, and the release that returns it to open and
		// unqueued runs inside the *next* mutation — this one — ahead of the new marker. So
		// the very part just rolled back, the common case, was refused while a sibling's
		// mutation would have released it. Measured by mini-spec-tool 2026-09-06.
		if id := p.QueueID(); id != 0 && !releasable(repoRoot, p) {
			return out, fmt.Errorf("%s part %s already carries queue ID #%d; a part records exactly one item", part.Doc, part.Key, id)
		}
	}
	if !found {
		return out, fmt.Errorf("no part keyed %s in %s", part.Key, part.Doc)
	}
	return mintAndPlace(repoRoot, part, text, place)
}

// R330
// releasable reports whether a part's queue ID belongs to the attempt the slot holds as
// reverted — the one state in which the next mutation releases the part before it marks it.
// A slot that cannot be read is not the reverted state, so the refusal stands; understating
// what may be re-queued costs one sibling mutation, overstating it queues a live part twice.
func releasable(repoRoot string, p parser.Part) bool {
	if !slices.Contains(p.Verbs(), "REVERTED") {
		return false
	}
	state, present, err := backup.New(repoRoot).State()
	return err == nil && present && state == backup.Reverted
}

// mintAndPlace mints the ID, places the entry, and writes the source side when there is one.
//
// **The gap case writes nothing on the source side** (R274), so the marker write is the one
// step that branches; everything else — the ID, the placement, the slot — is the same act for
// both kinds and is written once rather than twice.
func mintAndPlace(repoRoot string, part parser.PartRef, text parser.Entry, place parser.Place) (Created, error) {
	var out Created
	// Seq: seq-queue-item.md#1.5 | R241
	// max() across **both** queue files. Either alone collides: the pending file's maximum is
	// too low right after items complete, the done file's while the highest IDs are still live.
	scan, err := parser.ScanTrajectory(repoRoot)
	if err != nil {
		return out, err
	}
	id := scan.MaxItemID() + 1

	slot := backup.New(repoRoot)
	released, _ := slot.Released()
	text.ID, text.Part = id, part
	out = Created{ID: id, Part: part, Title: text.Title, ReusedID: slices.Contains(released, id)}

	// Seq: seq-queue-item.md#1.10 | R256, R257
	// Resolved **before** the slot opens, so a refusal costs no copy and leaves no state to
	// undo. The intent is the caller's; the position is this verb's to compute.
	pendingPath := filepath.Join(repoRoot, pendingFile)
	pos, total, err := resolvePlacement(repoRoot, pendingPath, place)
	if err != nil {
		return out, err
	}
	out.Position, out.Entries = pos, total

	// Seq: seq-queue-item.md#1.6 | R248
	// One Record for the whole invocation: the copy, the worktree anchor and the state
	// transition cover the change rather than a fragment of it, so a mis-typed pointer is a
	// `pending revert` instead of a repair.
	err = slot.Record(func() error {
		if err := parser.PlaceItem(pendingPath, text, pos); err != nil {
			return err
		}
		// R274 — nothing is written on the gap side. A gap has no marker and gains none.
		if part.Kind == minispecsdom.SourceGap {
			return nil
		}
		// Seq: seq-queue-item.md#1.6.2 | R241
		return parser.SetMarker(filepath.Join(repoRoot, filepath.FromSlash(part.Doc)), part.Key, "OPEN", fmt.Sprintf("#%d.", id))
	})
	if err != nil {
		return out, err
	}
	// The files it reports are the files it wrote, which is why the gap case names one.
	out.Files = []string{pendingFile}
	if part.Kind != minispecsdom.SourceGap {
		out.Files = append(out.Files, part.Doc)
	}
	return out, nil
}

// A gap ID is one capital letter and digits; a gap-shaped token is that, or one followed by a
// range or list separator — the forms the ID grammar accepts and this verb does not. R271, R272
var (
	// R271, R272
	gapIDRe     = regexp.MustCompile(`^[A-Z]\d+$`)
	gapShapedRe = regexp.MustCompile(`^[A-Z]\d+\s*[-,–]`)
)

// gapSource reads `--from` as a gap ID, answering a zero PartRef when it is not one. R271, R272, R273
//
// A **range or list** is refused *by the rule that an entry carries one pointer* rather than as
// a parse error: naming the decision is worth more to the caller than naming the syntax.
func gapSource(repoRoot, gapsPath, from string) (parser.PartRef, error) {
	if !gapIDRe.MatchString(from) {
		if gapShapedRe.MatchString(from) {
			// R272 — one gap per source, held in the scalar a part pointer fills.
			return parser.PartRef{}, fmt.Errorf("--from %q names several gaps, and an entry carries one pointer.\n"+
				"Name the gap this item repairs; any others belong in the entry's prose.", from)
		}
		if !strings.ContainsAny(from, "#/.") {
			// R272 — a token carrying none of `#`, `/` or `.` was **meant** as a gap ID, so
			// the gap grammar's message is the useful one, not the part pointer's.
			return parser.PartRef{}, fmt.Errorf("--from %q is not a gap ID: a gap ID is one capital letter and digits, such as O136", from)
		}
		return parser.PartRef{}, nil // not a gap ID; the part pointer path owns it
	}
	// R273 — gap IDs are scoped to a design.md and a repository may hold several roots, so
	// the root is resolved as every other gap verb resolves it and its absence is a refusal
	// naming what was looked for. A part pointer names its own document and needs none.
	if gapsPath == "" {
		return parser.PartRef{}, fmt.Errorf("--from %s names a gap, which needs a design root, and none was found.\n"+
			"Run from inside the project whose design.md holds %s, or name a carve part as <doc>#<part>.", from, from)
	}
	gaps, err := parser.ParseGaps(gapsPath)
	if err != nil {
		return parser.PartRef{}, err
	}
	doc := filepath.ToSlash(gapsPath)
	if rel, err := filepath.Rel(repoRoot, gapsPath); err == nil {
		doc = filepath.ToSlash(rel)
	}
	for _, g := range gaps {
		if g.ID == from {
			return parser.PartRef{Doc: doc, Key: g.ID, Kind: minispecsdom.SourceGap}, nil
		}
	}
	return parser.PartRef{}, fmt.Errorf("no gap %s in %s", from, doc)
}

// FinishOpts carries what the caller composes and this package only places. R262, R268
type FinishOpts struct {
	// Body is the done entry's body, placed beneath the header in the same write. R262, R263
	Body string
	// Discharged is whatever the item discharged besides its queue ID — a requirement range, a
	// gap ID, several separated by `/`. The `#N` is the tool's; this is the caller's. R268
	Discharged string
	// ResolveGap resolves the gap a gap-sourced entry names. **It is handed the whole ref, not
	// the ID**: the entry recorded which document held that gap when the item was created, and
	// a caller resolving by ID alone would resolve it in whichever design root it happens to be
	// standing in. Nil means the caller did not ask, and resolving is never inferred. R277
	//
	// *A callback rather than a dependency on the update package*, so this one keeps its
	// one-way path to parser and two verb layers do not reach into each other.
	ResolveGap func(gap parser.PartRef) error
	// DeclineResolve is `--no-resolve`: the caller decided the item does **not** close its
	// gap. It exists so that decision has a spelling, which is what makes the question
	// unskippable — a gap-sourced completion supplying neither is refused (R279). R281
	DeclineResolve bool
}

// CRC: crc-Pending.md | Seq: seq-queue-item.md#2 | R244, R246, R247, R248
// Finish completes an item **in the mandated order: the source first.**
//
// Each discharged part is checked off in its carve; then the current file is reset; then the
// entry moves from the pending file to the done file. Source first because the carve is the
// copy a future reader trusts, and the one nobody thinks to check — and because the queue
// files are the ones the slot can put back.
func Finish(repoRoot string, id int, commit string, opt FinishOpts) (Finished, error) {
	if err := missingTrajectory(repoRoot); err != nil { // R332
		return Finished{}, err
	}
	out := Finished{ID: id}
	entry, err := findEntry(repoRoot, id)
	if err != nil {
		return out, err
	}
	// R269. Refused **before** anything is written, because the reader takes the slot as the
	// run between the em dash and the colon that opens the title — a colon inside it ends the
	// slot early and every identifier after it stops being read, silently.
	if strings.Contains(opt.Discharged, ":") {
		return out, fmt.Errorf("--discharged %q carries a colon, and the identifier slot ends at the first one — everything after it would stop being read while the header still parsed; write the identifiers without a colon", opt.Discharged)
	}
	out.Parts = entry.Parts()
	out.Gap, out.HasGap = entry.Gap()
	// Seq: seq-queue-item.md#2.2.1 | R279
	// **Before anything is written**, so a caller who has not decided pays a retry rather
	// than a half-completed item.
	if out.HasGap && opt.ResolveGap == nil && !opt.DeclineResolve {
		return out, fmt.Errorf("#%d repairs gap %s, so completing it needs a decision: --resolve if this closed it,\n"+
			"--no-resolve if it did not. There is no default, because a default would guess which of those happened.", id, out.Gap.Key)
	}

	attribution := fmt.Sprintf("`%s`, %s — `#%d`.", commit, time.Now().Format("2006-01-02"), id)
	err = backup.New(repoRoot).Record(func() error {
		// R245, R246 — Seq: seq-queue-item.md#2.3.1
		// The parts the item recorded, and nothing else. A parent completes when its
		// subparts do — derived, never stored — so no parent box is set and no sibling is
		// consulted.
		for _, p := range out.Parts {
			if err := parser.SetPartLanded(filepath.Join(repoRoot, filepath.FromSlash(p.Doc)), p.Key, attribution); err != nil {
				return err
			}
			out.Files = append(out.Files, p.Doc)
		}
		// Seq: seq-queue-item.md#2.3.5 | R277
		// The gap's answer to the part's `LANDED` marker, and **only when asked**. Source
		// first, exactly as the part case is, and for the same reason.
		if out.HasGap && opt.ResolveGap != nil {
			if err := opt.ResolveGap(out.Gap); err != nil {
				return err
			}
			out.GapResolved = true
			out.Files = append(out.Files, out.Gap.Doc)
		}
		// Seq: seq-queue-item.md#2.3.3 | R249
		// The `## Active` section and nothing else. How that region is found is the reader's
		// — this package owns only *when* it happens, which is after the source and before
		// the queue.
		if err := parser.ResetCurrent(filepath.Join(repoRoot, currentFile)); err != nil {
			return err
		}
		out.Files = append(out.Files, currentFile)
		// Seq: seq-queue-item.md#2.3.4 | R244, R247
		out.Files = append(out.Files, pendingFile, doneFile)
		return parser.CompleteItem(filepath.Join(repoRoot, pendingFile), filepath.Join(repoRoot, doneFile), id,
			doneHeader(id, entry.Title, commit, opt.Discharged, out.Parts, out.Gap), opt.Body)
	})
	return out, err
}

// R247, R268, R278
// doneHeader composes the entry **header** and never the body.
//
// Identifiers, date, title, commit and part pointer are facts the caller was handed. The body
// — *enough to reconstruct the change without re-reading the code* — is a judgment about what
// a future reader will need, which is authoring. The tool owns IDs, not prose.
func doneHeader(id int, title, commit, discharged string, parts []parser.PartRef, gap parser.PartRef) string {
	// R268. The `#N` is the tool's because the tool owns IDs; whatever else the item
	// discharged is the caller's, and the two join with the ` / ` the format mandates.
	slot := fmt.Sprintf("#%d", id)
	if discharged != "" {
		slot += " / " + discharged
	}
	var b strings.Builder
	fmt.Fprintf(&b, "- **%s — %s: %s.** (`%s`)", time.Now().Format("2006-01-02"), slot, strings.TrimSuffix(title, "."), commit)
	for _, p := range parts {
		fmt.Fprintf(&b, " Part `%s#%s`.", p.Doc, p.Key)
	}
	// R278 — the ledger keeps the link the pending file held. The kind is the discriminator
	// here as it is everywhere else: a zero ref is not a gap.
	if gap.Kind == minispecsdom.SourceGap {
		fmt.Fprintf(&b, " Gap `%s#%s`.", gap.Doc, gap.Key)
	}
	return b.String()
}

// R242, R243
// splitPointer reads `<doc>#<key>`, the queue side of the link and the only side written as a
// pointer at all.
func splitPointer(from string) (parser.PartRef, error) {
	doc, key, ok := strings.Cut(from, "#")
	if !ok || doc == "" || key == "" {
		return parser.PartRef{}, fmt.Errorf("part pointer %q is not <doc>#<part>", from)
	}
	// R243. The key is the fragment — `4`, `2.2` — and `Item ` is the head's display word
	// (Bill, 2026-09-05, in the dependency). Flexible on input, rigid on output: a caller who
	// types the head's word is understood, and the entry is written in the one form.
	return parser.PartRef{Doc: doc, Key: strings.TrimPrefix(key, "Item "), Kind: minispecsdom.SourcePart}, nil
}
