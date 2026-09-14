// CRC: crc-Trajectory.md | R190, R194, R195, R197
package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
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
	out, _, err := pendingEntriesUnread(path)
	return out, err
}

// CRC: crc-Trajectory.md | R242, R251
//
// PartRef is one side of the item↔part link as the queue verbs carry it: the document, the
// key inside it, and whether that key is a part key or a gap ID. The document side is written
// as a pointer only here, on the **queue** side — a carve carries a bare key, because
// trajectory files are private in every project and a carve cannot point at a file a cloner
// does not have.
type PartRef struct {
	Doc  string                  `json:"doc"`
	Key  string                  `json:"key"`
	Kind minispecsdom.SourceKind `json:"kind,omitempty"`
}

// CRC: crc-Trajectory.md | R246, R251
// Parts is what this entry discharges in a carve, as a list.
//
// The model permits several and the reader records at most one; returning a list keeps the
// caller's loop honest when the entry shape grows. **Both halves, and the part kind**: a gap
// source is not a part, and an entry naming a document with no key records nothing, which is
// the ordinary case since most items point at a spec or a plain document.
func (e QueueEntry) Parts() []PartRef {
	if e.Kind != minispecsdom.SourcePart || e.SourceDoc == "" || e.SourceKey == "" {
		return nil
	}
	return []PartRef{{Doc: e.SourceDoc, Key: e.SourceKey, Kind: minispecsdom.SourcePart}}
}

// CRC: crc-Trajectory.md | R251, R276
// Gap is the gap this entry repairs, when its pointer names one. Scalar, exactly as the part
// pointer is; kept apart from Parts because the two complete by different acts.
func (e QueueEntry) Gap() (PartRef, bool) {
	if e.Kind != minispecsdom.SourceGap || e.SourceDoc == "" || e.SourceKey == "" {
		return PartRef{}, false
	}
	return PartRef{Doc: e.SourceDoc, Key: e.SourceKey, Kind: minispecsdom.SourceGap}, true
}

// CRC: crc-Trajectory.md | R252
//
// Entry is what add-item places: the caller composes title, skill, status and next action;
// the shape they go into is the dependency's EntryText.
type Entry struct {
	ID     int
	Title  string
	Skill  string
	Status string // the one-line status that follows the skill on the heading line
	Next   string // the `Next:` line's text, or empty for an entry that has none
	Part   PartRef
}

// PlaceKind names the four placement intents. R256
type PlaceKind uint8

const (
	PlaceLast  PlaceKind = iota // the end of the queue — the default, spelled out
	PlaceNext                   // *next to be worked*, resolved against the current file
	PlaceNth                    // position N, 1-based
	PlaceAfter                  // immediately after the entry whose item **ID** is N
)

// Place is where an entry goes: the caller's intent, not yet a position. R256
type Place struct {
	Kind PlaceKind
	N    int
}

// readPending parses the pending file at path through the dependency's reader.
func readPending(path string) (*minispecsdom.Pending, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return minispecsdom.ParsePending(string(src)), nil
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#1.10 | R256, R259, R260
// ResolvePlace turns an intent into a 1-based position among the entries the reader sees,
// and reports how many positions there are. `--next` is the orchestrator's to resolve
// against the current file before it reaches here.
func ResolvePlace(pendingPath string, p Place) (pos, total int, err error) {
	pend, err := readPending(pendingPath)
	if err != nil {
		return 0, 0, err
	}
	n := len(pend.Entries())
	switch p.Kind {
	case PlaceLast:
		return n + 1, n + 1, nil
	case PlaceNth:
		// Refused rather than clamped: a clamp is a silent reinterpretation of an
		// instruction the caller was specific about. R259
		if p.N < 1 || p.N > n+1 {
			return 0, 0, fmt.Errorf("--nth %d is outside 1 … %d: the queue holds %d entries, and a position is refused rather than clamped", p.N, n+1, n)
		}
		return p.N, n + 1, nil
	case PlaceAfter:
		pos, err := pend.After(p.N)
		if err != nil {
			return 0, 0, fmt.Errorf("--after %d names no live entry; the queue holds %s", p.N, itemList(pend.Entries()))
		}
		return pos, n + 1, nil
	}
	return 0, 0, fmt.Errorf("--next is resolved against the current file before it reaches here")
}

// itemList renders IDs the way the caller wrote them, for a refusal that hands back what
// would have worked.
func itemList(entries []*minispecsdom.Entry) string {
	if len(entries) == 0 {
		return "no entries"
	}
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, fmt.Sprintf("#%d", e.ID))
	}
	return strings.Join(parts, ", ")
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#1.6.1 | R252, R260
// PlaceItem writes the whole entry at pos through the dependency's Place, which inserts one
// synthetic text beside the entry it precedes or after the last one and refuses a position
// outside 1 … entries+1.
func PlaceItem(pendingPath string, e Entry, pos int) error {
	return editFile(pendingPath, func(src string) (string, error) {
		pend := minispecsdom.ParsePending(src)
		text := minispecsdom.EntryText{
			ID: e.ID, Title: e.Title, Skill: e.Skill, Status: e.Status, Next: e.Next,
			SourceDoc: e.Part.Doc, SourceKey: e.Part.Key, Kind: e.Part.Kind,
		}
		// R275 — the gap form is the dependency's; anything else writes the part form.
		if text.Kind != minispecsdom.SourceGap {
			text.Kind = minispecsdom.SourcePart
		}
		if err := pend.Place(text, pos); err != nil {
			return "", err
		}
		return pend.Render()
	})
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#2.3.4 | R244, R263, R270
// CompleteItem removes the entry from the pending file and prepends the done entry — header
// and body in one write. Two files, and the pending side first, so a failure on the done side
// leaves an ID absent from both rather than present in both.
func CompleteItem(pendingPath, donePath string, id int, header, body string) error {
	err := editFile(pendingPath, func(src string) (string, error) {
		pend := minispecsdom.ParsePending(src)
		if err := pend.Remove(id); err != nil {
			return "", fmt.Errorf("no entry for #%d in %s: %w", id, pendingPath, err)
		}
		return pend.Render()
	})
	if err != nil {
		return err
	}
	return editFile(donePath, func(src string) (string, error) {
		done := minispecsdom.ParseDone(src)
		if err := done.Prepend(header, body); err != nil {
			return "", err
		}
		return done.Render()
	})
}

// editCurrent parses the current file, refusing as the reader does when `## Active` cannot
// be told apart, and writes the edit back atomically. R249, R250
func editCurrent(path string, edit func(*minispecsdom.Current) error) error {
	return editFile(path, func(src string) (string, error) {
		cur, err := parseCurrent(path, src)
		if err != nil {
			return "", err
		}
		if err := edit(cur); err != nil {
			return "", fmt.Errorf("%s: %w", path, err)
		}
		return cur.Render()
	})
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#2.3.3 | R249, R250
// ResetCurrent clears the current file's `## Active` region and **nothing else**: the reader
// enumerates the region's nodes and the write removes those and inserts the placeholder, so
// standing context outside it is unreachable by construction.
func ResetCurrent(path string) error {
	return editCurrent(path, (*minispecsdom.Current).Reset)
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#3.3.2 | R265, R266, R267
// SetActive opens an item: the identity line the orchestrator composed, then the caller's
// context. The same reader and write path that clears the region; the reader refuses over a
// held item.
func SetActive(path, line, context string) error {
	body := line
	if context != "" {
		body += "\n\n" + context
	}
	return editCurrent(path, func(c *minispecsdom.Current) error {
		err := c.SetActive(body)
		if errors.Is(err, minispecsdom.ErrOccupied) {
			return fmt.Errorf("%w — opening another here would discard it; park it as a sub-item in the pending file, which is a stack you can push onto", err)
		}
		return err
	})
}

// CRC: crc-Trajectory.md | Seq: seq-queue-item.md#1.10.2 | R257
// ActiveInProgress asks the current file whether a step is being worked: `## Active` holding
// something other than its placeholder.
func ActiveInProgress(path string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	cur, err := parseCurrent(path, string(src))
	if err != nil {
		return false, err
	}
	return cur.Occupied(), nil
}

// parseCurrent is ParseCurrent with the two refusals naming their repair. R250
//
// The reader says what it could not tell apart; the repair is this tool's to name, because the
// shape it names is the skill's format. Told apart by the reader's sentinels (`ErrNoActive`,
// `ErrManyActive`), which arrived the day they were asked for.
func parseCurrent(path, src string) (*minispecsdom.Current, error) {
	cur, err := minispecsdom.ParseCurrent(src)
	if err == nil {
		return cur, nil
	}
	switch {
	case errors.Is(err, minispecsdom.ErrNoActive):
		return nil, fmt.Errorf("%s carries no `## Active` heading, so the active item cannot be told from the standing context; add the heading beneath the rule, holding `_No active item._`", path)
	case errors.Is(err, minispecsdom.ErrManyActive):
		return nil, fmt.Errorf("%s carries more than one `## Active` heading, so the region to write is ambiguous; keep one", path)
	}
	return nil, fmt.Errorf("%s: %w", path, err)
}

// CRC: crc-Trajectory.md | R287, R292
//
// DoneEntry is one done-file entry as the dependency's Done reader reads it: the IDs in the
// header's identifier slot — the only source of queue IDs — whether the header carried a
// slot at all, and the part pointer from the header or the body.
type DoneEntry struct {
	Date    string `json:"date"`
	IDs     []int  `json:"ids"`
	HasSlot bool   `json:"has_slot"`
	Title   string `json:"title"`
	Commit  string `json:"commit"`
	PartDoc string `json:"part_doc,omitempty"`
	PartKey string `json:"part_key,omitempty"`
	Line    int    `json:"line"`
}

// CRC: crc-Trajectory.md | R287
// DoneEntries reads the done file at path through minispecsdom.Done, with the count of
// entry-like lines the reader did not recognize. A missing file is no entries and no error.
func DoneEntries(path string) ([]DoneEntry, []minispecsdom.Unread, error) {
	src, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	d := minispecsdom.ParseDone(string(src))
	var out []DoneEntry
	for _, e := range d.Entries() {
		out = append(out, DoneEntry{
			Date: e.Date, IDs: e.IDs, HasSlot: e.HasSlot, Title: e.Title, Commit: e.Commit,
			PartDoc: e.PartDoc, PartKey: e.PartKey, Line: e.Line(),
		})
	}
	return out, d.Unread(), nil
}

// CRC: crc-Trajectory.md | R287, R297
//
// QueueScan is both queue files read through the dependency's readers — the entries and
// what each reader could not recognize — for the checks that need more than IDs.
// ScanTrajectory's regex ID scan stays beside it for `next-id`; see gap O19.
type QueueScan struct {
	Pending       []QueueEntry
	Done          []DoneEntry
	PendingUnread []minispecsdom.Unread
	DoneUnread    []minispecsdom.Unread
	CurrentUnread []minispecsdom.Unread // R302
}

// CRC: crc-Trajectory.md | Seq: seq-validate-trajectory.md#1.2.1 | R287
// ScanQueue reads both queue files beneath repoRoot through the dependency.
func ScanQueue(repoRoot string) (QueueScan, error) {
	var q QueueScan
	var err error
	q.Pending, q.PendingUnread, err = pendingEntriesUnread(filepath.Join(repoRoot, "PENDING.md"))
	if err != nil {
		return q, err
	}
	q.Done, q.DoneUnread, err = DoneEntries(filepath.Join(repoRoot, "DONE.md"))
	if err != nil {
		return q, err
	}
	q.CurrentUnread = currentUnread(filepath.Join(repoRoot, "CURRENT.md"))
	return q, nil
}

// R302
// currentUnread is the current file's unread list. A file that is missing or that the reader
// refuses contributes nothing here: absence and shape are each answered by their own check.
func currentUnread(path string) []minispecsdom.Unread {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	cur, err := minispecsdom.ParseCurrent(string(src))
	if err != nil {
		return nil
	}
	return cur.Unread()
}

// pendingEntriesUnread is PendingEntries with the reader's unread lines beside it.
func pendingEntriesUnread(path string) ([]QueueEntry, []minispecsdom.Unread, error) {
	src, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	p := minispecsdom.ParsePending(string(src))
	var out []QueueEntry
	for _, e := range p.Entries() {
		out = append(out, QueueEntry{
			ID: e.ID, Title: e.Title, SourceDoc: e.SourceDoc, SourceKey: e.SourceKey,
			Kind: e.Kind, Line: e.Line(),
		})
	}
	return out, p.Unread(), nil
}

// CRC: crc-Trajectory.md | R296
// CheckActive reports the current file's shape: exactly one `## Active`, with the repair
// named. A file that is not there is not this finding — absence is answered elsewhere.
func CheckActive(path string) error {
	src, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = parseCurrent(path, string(src))
	return err
}
