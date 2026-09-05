// CRC: crc-Backup.md | R221, R222, R223
// Package backup is one level of undo and one of redo over the trajectory files.
//
// Explicitly not an undo stack — the 80s `vi` u with goldfish memory. The most recent
// change is revertable and replayable and nothing older is recoverable, which is enough
// for the real emergency (a command that did the wrong thing thirty seconds ago) and
// avoids owning a history the VCS already owns better.
package backup

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

// State is what the stamp records. R228
//
// Three, not two. Two would count *configurations* — of which there are indeed two,
// live-holds-the-change and live-holds-the-original — and conclude that Changed and
// Replayed are one thing. They are the same configuration and different **nodes**:
// Changed is the entry, reachable only by a fresh mutation and never by toggling.
// Collapsing them would make a fresh mutation stamp Replayed, which is the tool asserting
// something untrue. A stamp that lies is worse than a stamp with one more value in it.
type State uint8

const (
	// Changed is what a fresh mutation leaves behind: a backup taken, neither a revert
	// nor a replay having happened.
	Changed State = iota
	Reverted
	Replayed
)

var stateNames = map[State]string{Changed: "changed", Reverted: "reverted", Replayed: "replayed"}

func (s State) String() string {
	if n, ok := stateNames[s]; ok {
		return n
	}
	return fmt.Sprintf("state(%d)", uint8(s))
}

func parseState(s string) (State, bool) {
	s = strings.TrimSpace(s)
	for st, name := range stateNames {
		if name == s {
			return st, true
		}
	}
	return 0, false
}

// Covered names the files the slot restores. R222
//
// The three trajectory files, and deliberately **not** a carve. A carve is not a
// trajectory file — the format sites these three as those and lists carves separately —
// so it is written on a revert rather than restored, which is what lets a part's vended
// number survive one.
//
// Taken from Project rather than restated: `track` already governs the ignore state of
// exactly this set, and a second spelling of it here is a copy that can disagree.
func Covered() []string { return slices.Clone(project.TrajectoryFiles) }

// ErrNoSlot reports that nothing has been recorded yet, so there is nothing to undo.
var ErrNoSlot = errors.New("the backup slot is empty: nothing has been recorded yet")

// DriftError reports files changed since the stamp. R227
//
// A refusal rather than a reconciliation: the operator is better placed than the tool to
// reconcile a hand edit with a pending revert, and a revert that silently clobbers a later
// edit is worse than no revert.
type DriftError struct {
	Files     []string
	BackupDir string
}

func (e *DriftError) Error() string {
	return fmt.Sprintf("refusing: %s changed since the backup was taken; "+
		"their backups are in %s — reconcile by hand, then re-run",
		strings.Join(e.Files, ", "), e.BackupDir)
}

// IllegalError reports an operation the current state does not accept. R228
//
// It names the state and what that state will accept, because a refusal that does not is a
// crank handle that fails to crank — the caller is left with nothing to do next.
type IllegalError struct {
	State State
	Op    string
	Legal string
}

func (e *IllegalError) Error() string {
	return fmt.Sprintf("refusing: the backup slot is %s, which accepts %s, not %s",
		e.State, e.Legal, e.Op)
}

// Slot is the backup directory and the stamp inside it. R223
//
// Repository-scoped and knows nothing of a design root: the trajectory layer sits above
// every design root, the same reason Carves and NextItemID take a repository root.
type Slot struct {
	root string
	dir  string
}

// New opens the slot beneath a repository root, at the path Project already names and
// `init` already has git ignoring.
func New(repoRoot string) *Slot {
	return &Slot{
		root: repoRoot,
		dir:  filepath.Join(repoRoot, project.ConfigDirName, project.BackupDirName),
	}
}

// Dir is where the copies and the stamp live, together. R223
func (s *Slot) Dir() string { return s.dir }

func (s *Slot) stampPath() string { return filepath.Join(s.dir, "stamp") }

// R226
// State reports the slot's state, and whether it holds anything at all.
//
// The stamp carries the state and **nothing else** — in particular not the item ID, which
// needs no field because the snapshot *is* the pending file: after a revert the backed-up
// copy holds the entry with its number in it.
func (s *Slot) State() (State, bool, error) {
	body, err := os.ReadFile(s.stampPath())
	if errors.Is(err, os.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	st, ok := parseState(string(body))
	if !ok {
		return 0, false, fmt.Errorf("unreadable stamp in %s: %q", s.dir, strings.TrimSpace(string(body)))
	}
	return st, true, nil
}

// CRC: crc-Backup.md | Seq: seq-backup.md#2 | R224, R229
// Record takes a backup, runs the caller's mutation, and stamps Changed.
//
// It resets the slot from **any** state and discards whatever was revertable, without
// ceremony. That is the whole memory model — the goldfish rule — and it is why the slot
// never needs a history.
func (s *Slot) Record(perform func() error) error {
	// A mutation from `reverted` ends an attempt for good: its entry is already gone from
	// the live file and replay is about to become impossible. The part goes back to open
	// and unqueued while both sides of the pair are still readable. R231
	//
	// Inside `perform` rather than before the swap: staging only *copies* the live files
	// and the old backup is untouched until the move at the end, so both sides read the
	// same as they did — and the release writes a marker into a **tracked** carve, which
	// must land *after* the worktree anchor taken at the top of the swap. R236
	//
	// Seq: seq-backup.md#2.2 | R232
	// **Only from `reverted`.** The entry diff alone cannot tell an abandoned attempt from a
	// finished one — a completion removes the entry from the live file exactly as a revert
	// does — so without this the first mutation after any completion reopens the landed part.
	// Measured 2026-08-18 on a tracked public carve. A guard here was once deleted as dead
	// after an injection removing it changed nothing; it changed nothing because the
	// completion verb did not exist yet. An injection proves only what it could reach.
	cur, present, err := s.State()
	if err != nil {
		return err
	}
	return s.swap(Changed, func() error {
		if present && cur == Reverted {
			if err := s.releaseAttempt(); err != nil {
				return err
			}
		}
		return perform()
	})
}

// CRC: crc-Backup.md | Seq: seq-backup.md#3 | R225, R227, R228
// Revert restores the covered files from the backup.
//
// The body is on its own line deliberately: the alarm on `toggle` resolves by the first line
// git's pattern matches, and a one-line body would put the call in column one ahead of the
// declaration.
func (s *Slot) Revert() error {
	return s.toggle(Reverted, "revert")
}

// Replay is Revert in the other direction, and the same mechanism. R225
func (s *Slot) Replay() error {
	return s.toggle(Replayed, "replay")
}

// Seq: seq-backup.md#3.1 | R228
// toggle runs the legality check, the drift check, and then the swap.
func (s *Slot) toggle(next State, op string) error {
	cur, present, err := s.State()
	if err != nil {
		return err
	}
	if !present {
		return ErrNoSlot
	}
	if want := legal(cur); want != op {
		return &IllegalError{State: cur, Op: op, Legal: want}
	}
	drifted, err := s.drift()
	if err != nil {
		return err
	}
	if len(drifted) > 0 {
		return &DriftError{Files: drifted, BackupDir: s.dir}
	}

	// Read both sides *before* the swap, because the swap exchanges them. Which item is
	// involved is derived from that pair rather than stored anywhere: revert diffs
	// live→backup, replay diffs backup→live, and the same code serves both. R230
	live, backedUp, err := s.pendingBothSides()
	if err != nil {
		return err
	}
	if err := s.swap(next, s.restore); err != nil {
		return err
	}
	// The entry departs from opposite sides in each direction, so the pair is ordered by
	// which side holds it now: a revert removes it from the live file, a replay puts it
	// back. Passing the same order for both would silently mark nothing on replay.
	before, after, verb := live, backedUp, "REVERTED"
	if next == Replayed {
		before, after, verb = backedUp, live, "OPEN"
	}
	return s.mark(before, after, verb, queuedAttribution, anyPart)
}

// pendingBothSides reads the pending file as it stands live and as the backup holds it.
func (s *Slot) pendingBothSides() (live, backedUp []parser.QueueEntry, err error) {
	live, err = parser.PendingEntries(filepath.Join(s.root, "PENDING.md"))
	if err != nil {
		return nil, nil, err
	}
	backedUp, err = parser.PendingEntries(filepath.Join(s.dir, "PENDING.md"))
	if err != nil {
		return nil, nil, err
	}
	return live, backedUp, nil
}

// Seq: seq-backup.md#3.1 | R228
// legal names the single operation the given state accepts.
//
// From every state **exactly one** of revert / replay is legal, so there is never a choice
// to disambiguate and never a "reverting twice" case to define.
func legal(s State) string {
	if s == Reverted {
		return "replay"
	}
	return "revert"
}

// Seq: seq-backup.md#3.2 | R227
// drift names covered files modified since the stamp was written.
//
// An mtime comparison against the **single** stamp: no content hashing, and no per-file
// snapshots. That is make's dependency model, right for the same reason it was then — the
// cheap check is sufficient. Comparing against one stamp rather than per file is also what
// makes the check atomic across the files at once.
func (s *Slot) drift() ([]string, error) {
	stamp, err := os.Stat(s.stampPath())
	if err != nil {
		return nil, err
	}
	var out []string
	for _, name := range Covered() {
		info, err := os.Stat(filepath.Join(s.root, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.ModTime().After(stamp.ModTime()) {
			out = append(out, name)
		}
	}
	return out, nil
}

// Seq: seq-backup.md#1 | R224
// swap is the one operation every path shares: copy live aside, perform, stamp, move.
//
// **The move is last and it is a move, not a copy.** A rename within a filesystem is
// atomic, so a crash leaves either the old backup intact or the new one complete — never a
// half-written backup. Same reason one writes a temp file and renames rather than
// truncating in place. Do not simplify this into a direct write to the backup path: the
// shorter form passes every test that does not interrupt it.
func (s *Slot) swap(next State, perform func() error) error {
	// Seq: seq-backup.md#1.5 | R236, R239
	// **First, before anything this transition writes.** Here rather than in the two callers
	// because the swap is the one operation `changed`, `reverted` and `replayed` all pass
	// through, so a future transition cannot be added that forgets it.
	//
	// A repository-less tree is not a failure: the anchor is belt and suspenders over git,
	// and the slot's own job — restoring the trajectory files — does not depend on it.
	if err := project.NewGit(s.root).Snapshot(); err != nil && !errors.Is(err, project.ErrNoGit) {
		return err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(s.dir, ".staging-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	staged, err := s.stage(tmp)
	if err != nil {
		return err
	}
	if err := perform(); err != nil {
		return err
	}
	if err := os.WriteFile(s.stampPath(), []byte(next.String()+"\n"), 0o644); err != nil {
		return err
	}
	for _, name := range staged {
		if err := os.Rename(filepath.Join(tmp, name), filepath.Join(s.dir, name)); err != nil {
			return err
		}
	}
	// A file that was absent when staged must not keep a stale backup, or a revert would
	// resurrect a file the mutation deleted.
	return s.pruneAbsent(staged)
}

// stage copies the live covered files into tmp, returning the ones that existed.
func (s *Slot) stage(tmp string) ([]string, error) {
	var staged []string
	for _, name := range Covered() {
		src := filepath.Join(s.root, name)
		_, err := os.Stat(src)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if err := copyFile(src, filepath.Join(tmp, name)); err != nil {
			return nil, err
		}
		staged = append(staged, name)
	}
	return staged, nil
}

func (s *Slot) pruneAbsent(staged []string) error {
	for _, name := range Covered() {
		if slices.Contains(staged, name) {
			continue
		}
		if err := removeIfPresent(filepath.Join(s.dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// Seq: seq-backup.md#3.3 | R225
// restore copies the backups back over the live files.
func (s *Slot) restore() error {
	for _, name := range Covered() {
		src := filepath.Join(s.dir, name)
		dst := filepath.Join(s.root, name)
		_, err := os.Stat(src)
		if errors.Is(err, os.ErrNotExist) {
			// The backup holds no such file, so the other state had none either.
			if err := removeIfPresent(dst); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// removeIfPresent deletes path, treating an already-absent file as success.
func removeIfPresent(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	// **Deliberately no Chtimes here**, and the omission is load-bearing. An earlier draft
	// normalised the copy's modification time with `time.Now()`, which mixes two clocks:
	// the kernel timestamps ordinary writes from a *coarse* clock — millisecond
	// granularity — while `Chtimes` sets an exact time from a fine one. A restored file
	// then lands a few milliseconds *ahead* of the stamp written after it, and the drift
	// check reads its own restore as a hand edit. Measured at a 3ms gap, one coarse tick.
	return out.Close()
}

// The two rules mark writes under, named because a bare `true` at a call site says nothing
// about which of them is in force. R233
const (
	anyPart       = false // a revert or a replay marks every departed entry
	openPartsOnly = true  // a release never marks a part already recorded as done
)

// queuedAttribution is the attribution of a transient marker that names its queue item;
// unqueuedAttribution is the other transient form. The verb is the caller's. R230, R231
func queuedAttribution(id int) string { return fmt.Sprintf("#%d.", id) }
func unqueuedAttribution(int) string  { return "not queued." }

// CRC: crc-Backup.md | Seq: seq-backup.md#3.4 | R230, R231, R233
// mark records an attempt's fate in the carve that owns its part, through the Carve adapter.
//
// The queue rolls **backward** and the carve moves **forward**, so this is a write and
// never a restore. It is also why the part's vended number survives a revert at all: a
// carve is not a trajectory file, so it is not in the covered set.
//
// Which item is involved is not stored anywhere — it is *derived*: the item being moved
// away from is the one whose entry is in the pending file on one side of the swap and
// absent on the other.
func (s *Slot) mark(before, after []parser.QueueEntry, verb string, attribution func(id int) string, onlyOpen bool) error {
	for _, e := range departed(before, after) {
		if e.SourceDoc == "" || e.SourceKey == "" {
			continue
		}
		path := filepath.Join(s.root, filepath.FromSlash(e.SourceDoc))
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			// A pointer at a carve that is not there is a finding for `validate
			// trajectory`, not a reason to fail an undo.
			continue
		}
		// R233: a release never touches a part the tool has recorded as done — a second and
		// independent reason not to write. Below the existence check, and that order is
		// measured: reading the carve first turned the skip above into a hard failure.
		if onlyOpen {
			landed, err := parser.PartIsLanded(path, e.SourceKey)
			if err != nil {
				return err
			}
			if landed {
				continue
			}
		}
		if err := parser.SetMarker(path, e.SourceKey, verb, attribution(e.ID)); err != nil {
			return err
		}
	}
	return nil
}

// departed names entries present in before and absent from after, by item ID.
func departed(before, after []parser.QueueEntry) []parser.QueueEntry {
	held := map[int]bool{}
	for _, e := range after {
		held[e.ID] = true
	}
	var out []parser.QueueEntry
	for _, e := range before {
		if !held[e.ID] {
			out = append(out, e)
		}
	}
	return out
}

// CRC: crc-Backup.md | Seq: seq-backup.md#2.2 | R231, R234
// Released names the item IDs a new mutation would abandon.
//
// **The diff and nothing else**: an entry is reported when the backup holds it and the live
// file does not. The diff does not imply `reverted` — a completion removes the entry from
// the live file exactly as a revert does — which is harmless *here*, because the ID is only
// compared against the one being minted and a completed number stays in the done file. It
// was not harmless in Record, which is why R232 lives there.
//
// It is a query rather than a side effect because the announcement has to happen at the
// vend, and the vend is *inside* the mutation that does the releasing.
func (s *Slot) Released() ([]int, error) {
	live, backedUp, err := s.pendingBothSides()
	if err != nil {
		return nil, err
	}
	var ids []int
	for _, e := range departed(backedUp, live) {
		ids = append(ids, e.ID)
	}
	return ids, nil
}

// Seq: seq-backup.md#2.2 | R231
// releaseAttempt returns an abandoned attempt's part to the open-and-unqueued state.
//
// **Aborting an attempt is not aborting the part.** The queue item was an attempt at it;
// the part is still open and still to be completed, so it reads `OPEN (not queued.)` and
// not some verb implying it was given up on. **No done entry is written**: the done file is
// the completion ledger, and an unfinished attempt reconstructs nothing.
func (s *Slot) releaseAttempt() error {
	live, backedUp, err := s.pendingBothSides()
	if err != nil {
		return err
	}
	return s.mark(backedUp, live, "OPEN", unqueuedAttribution, openPartsOnly)
}
