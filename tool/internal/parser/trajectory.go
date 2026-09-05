// CRC: crc-Trajectory.md | R190, R194, R195, R197
package parser

import (
	"bufio"
	"errors"
	"github.com/zot/simple-dom/minispecsdom"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// The shapes below are defined by the skill's trajectory-format.md, which is normative
// for them. They are consumed here and deliberately not redefined: a second normative
// statement of a format is the failure the trajectory layer exists to prevent.
var (
	// A pending-file item entry is a `##` heading opening with its number.
	pendingItemRe = regexp.MustCompile(`^##\s+(\d+)\.`)
	// A done-file entry header opens in bold at column 0; its body lines are indented.
	doneEntryRe = regexp.MustCompile(`^- \*\*`)
	// The identifier slot of a done entry: the run between the date's em dash and the
	// colon that opens the title. Deliberately unanchored, so that removing the header
	// guard above is a real injection rather than a no-op.
	doneSlotRe = regexp.MustCompile(`—\s*([^:]*):`)
	// A queue ID inside that slot, written bare.
	doneIDRe = regexp.MustCompile(`#(\d+)`)
)

// ErrNoTrajectoryFiles marks a tree that carries neither queue file. R194
//
// Distinguished from "the queue is empty" because the two are different claims and only
// one of them has an answer. A project running no trajectory layer has no next item ID;
// replying 1 would be a confident wrong answer rather than an absent one.
var ErrNoTrajectoryFiles = errors.New("no trajectory files found at the repository root")

// TrajectoryFileScan is one file's contribution. R197
//
// Present is kept separate from len(IDs) because they answer different questions: a file
// that is absent and a file that parsed nothing both yield no identifiers, and only the
// first is a missing file. Collapsing them is how a broken regex comes to look like an
// empty queue.
type TrajectoryFileScan struct {
	Name    string
	Present bool
	IDs     []int
}

// TrajectoryScan is what the queue files contributed, per file. R197
type TrajectoryScan struct {
	Files []TrajectoryFileScan
}

// CRC: crc-Trajectory.md | R190, R195, R197
// ScanTrajectory reads the item IDs from the pending and done files beneath repoRoot.
//
// Both files are always reported, present or not, so the caller can name what it could
// not read rather than quietly answering from half the evidence.
func ScanTrajectory(repoRoot string) (TrajectoryScan, error) {
	var scan TrajectoryScan
	for _, f := range []struct {
		name  string
		parse func(string) ([]int, error)
	}{
		{"PENDING.md", parsePendingIDs},
		{"DONE.md", parseDoneIDs},
	} {
		ids, err := f.parse(filepath.Join(repoRoot, f.name))
		switch {
		case errors.Is(err, os.ErrNotExist):
			scan.Files = append(scan.Files, TrajectoryFileScan{Name: f.name})
		case err != nil:
			return scan, err
		default:
			scan.Files = append(scan.Files, TrajectoryFileScan{Name: f.name, Present: true, IDs: ids})
		}
	}
	return scan, nil
}

// MaxItemID is the highest ID across **both** files. R190
//
// Either file alone returns a plausible number that collides with an existing ID: the
// pending file's maximum is too low right after items complete, the done file's while
// the highest IDs are still live. That is why this takes the whole scan rather than a
// file.
func (s TrajectoryScan) MaxItemID() int {
	maxID := 0
	for _, f := range s.Files {
		for _, id := range f.IDs {
			maxID = max(maxID, id)
		}
	}
	return maxID
}

// Missing names the files that were not there, in scan order. R195
func (s TrajectoryScan) Missing() []string {
	var out []string
	for _, f := range s.Files {
		if !f.Present {
			out = append(out, f.Name)
		}
	}
	return out
}

// AnyPresent reports whether either file existed. R194
func (s TrajectoryScan) AnyPresent() bool {
	for _, f := range s.Files {
		if f.Present {
			return true
		}
	}
	return false
}

// parsePendingIDs collects the numbers of the pending file's item headings.
func parsePendingIDs(path string) ([]int, error) {
	return scanIDs(path, func(line string) []int {
		return submatchInts(pendingItemRe, line)
	})
}

// CRC: crc-Trajectory.md | R190
// parseDoneIDs collects queue IDs from the identifier slot of done-entry **headers**.
//
// The slot holds whatever the entry discharged — a queue ID, a gap ID, a requirement
// range, or several separated by `/` — so every `#N` in it counts and nothing outside it
// does.
//
// Headers only, deliberately. Entry bodies are prose several lines long and routinely
// quote other items: measured 2026-08-16, five body lines in ark's ledger would
// contribute a queue ID if the header were not required. Letting one in is the same
// class of error as reading one file instead of two — a number that is plausible and
// wrong.
func parseDoneIDs(path string) ([]int, error) {
	return scanIDs(path, func(line string) []int {
		if !doneEntryRe.MatchString(line) {
			return nil
		}
		slot := doneSlotRe.FindStringSubmatch(line)
		if slot == nil {
			return nil
		}
		return submatchInts(doneIDRe, slot[1])
	})
}

// scanIDs walks a file line by line, collecting whatever match reports.
func scanIDs(path string, match func(string) []int) ([]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ids []int
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ids = append(ids, match(scanner.Text())...)
	}
	return ids, scanner.Err()
}

// submatchInts pulls the first capture group of every match as an int. A done entry's
// slot may name more than one item, so the count per line is not bounded at one.
func submatchInts(re *regexp.Regexp, s string) []int {
	var ids []int
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		if n, err := strconv.Atoi(m[1]); err == nil {
			ids = append(ids, n)
		}
	}
	return ids
}

// CRC: crc-Trajectory.md | R240
//
// QueueEntry is one pending-file entry as the backup slot and the queue verbs consume it,
// read through the dependency's Pending reader. The source is a carve part or a gap, told
// apart by Kind; SourceKey is the part key or the gap ID.
type QueueEntry struct {
	ID        int                     `json:"id"`
	Title     string                  `json:"title"`
	SourceDoc string                  `json:"source_doc,omitempty"`
	SourceKey string                  `json:"source_key,omitempty"`
	Kind      minispecsdom.SourceKind `json:"kind"`
	Line      int                     `json:"line"`
}

// CRC: crc-Trajectory.md | R240
// PendingEntries reads the pending file at path through minispecsdom.Pending. A missing file
// is no entries and no error: the slot legitimately reads a side that has none.
func PendingEntries(path string) ([]QueueEntry, error) {
	src, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []QueueEntry
	for _, e := range minispecsdom.ParsePending(string(src)).Entries() {
		out = append(out, QueueEntry{
			ID:        e.ID,
			Title:     e.Title,
			SourceDoc: e.SourceDoc,
			SourceKey: e.SourceKey,
			Kind:      e.Kind,
			Line:      e.Line(),
		})
	}
	return out, nil
}
