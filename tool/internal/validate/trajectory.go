// CRC: crc-TrajectoryValidate.md | R284, R285, R286
package validate

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/simple-dom/minispecsdom"
)

// The two queue files, each check asking for one or the other by name. trajectory-format.md
// remains normative for the names; these are a consumer of them.
const (
	pendingFile = "PENDING.md"
	doneFile    = "DONE.md"
	currentFile = "CURRENT.md"
)

// doneVerbs and openVerbs are the marker classes a checkbox has to agree with. The
// vocabulary is open by design, so this is a *known* subset rather than an exhaustive one:
// a verb coined next month is unclassified and silent, which is the right failure — a
// disagreement asserted about a word nobody has defined would be the tool inventing intent.
var (
	doneVerbs = map[string]bool{"LANDED": true, "MIGRATED": true, "DISCHARGED": true, "SENT": true}
	openVerbs = map[string]bool{"OPEN": true, "REVERTED": true, "DEFERRED": true}
)

// TrajectoryIssues is what the scan found, bucketed by class. R284
//
// Buckets rather than a flat list because the classes have different repairs: a dangling
// citation is edited in the carve, an unmigrated ledger entry in the done file.
type TrajectoryIssues struct {
	// Absent means no trajectory layer at all — clean, and distinct from clean-after-checking. R286
	Absent        bool     `json:"absent"`
	Dangling      []string `json:"dangling,omitempty"`
	MissingParts  []string `json:"missing_parts,omitempty"`
	Duplicates    []string `json:"duplicates,omitempty"`
	Orphans       []string `json:"orphans,omitempty"`
	Unmigrated    []string `json:"unmigrated,omitempty"`
	Disagreements []string `json:"disagreements,omitempty"`
	// Unreachable counts citations the position rule cannot see because the document keys
	// on a superseded scheme. Reported beside the findings, never instead of them. R293
	Unreachable     int      `json:"unreachable"`
	UnreachableDocs []string `json:"unreachable_docs,omitempty"`
	// MissingIDs are numbers below the maximum assigned that appear in no readable entry. R295
	MissingIDs []int `json:"missing_ids,omitempty"`
	// Unread is the second coverage dimension: entry-like lines the reader did not recognize,
	// per file. Every shape-based check above is blind to them by construction. R297
	Unread map[string]int `json:"unread,omitempty"`
	// Structure holds findings about a trajectory file's own shape — `CURRENT.md` not
	// carrying exactly one `## Active`. R296
	Structure []string `json:"structure,omitempty"`
	// Disagreements between the two readers of the queue files: an item ID the line scan
	// found that the document reader returned no entry for, or the reverse. The document
	// reader can lose a whole tail of a file to one unclosed span and report nothing unread,
	// so the line scan is the second opinion that says how much of the file it actually saw. R300
	ReaderDisagreement []string `json:"reader_disagreement,omitempty"`
	// Stateless names lines in a carve's status block the reader could not read as parts.
	// A carve that loses one loses it from `query carves` and from every check here at once,
	// so it is an issue rather than a note. R298
	Stateless []string `json:"stateless,omitempty"`
}

// HasIssues reports whether anything needs repair. Unreachable citations and unread entries
// are statements about coverage, not issues in themselves. R293, R297
func (t *TrajectoryIssues) HasIssues() bool {
	return len(t.Dangling)+len(t.MissingParts)+len(t.Duplicates)+
		len(t.Orphans)+len(t.Unmigrated)+len(t.Disagreements)+len(t.MissingIDs)+
		len(t.Structure)+len(t.Stateless)+len(t.ReaderDisagreement) > 0
}

// CRC: crc-TrajectoryValidate.md | Seq: seq-validate-trajectory.md#1.2 | R284, R286
// RunTrajectory checks the queue files and the carves against each other.
//
// Repository-scoped: it takes a root and never a Project. Validate resolves a design root
// and runs once per one, and a repository may hold several, so folding this in would report
// the same drift once per root. R285
func RunTrajectory(repoRoot string) (*TrajectoryIssues, error) {
	traj, err := parser.ScanTrajectory(repoRoot)
	if err != nil {
		return nil, err
	}
	carves, carveErr := parser.ScanCarves(repoRoot)
	noCarves := errors.Is(carveErr, parser.ErrNoCarveDirs)
	if carveErr != nil && !noCarves {
		return nil, carveErr
	}
	// Nothing that could be inconsistent is a clean result, and it is not "could not
	// check" either. R286
	if !traj.AnyPresent() && noCarves {
		return &TrajectoryIssues{Absent: true}, nil
	}
	q, err := parser.ScanQueue(repoRoot)
	if err != nil {
		return nil, err
	}

	t := &TrajectoryIssues{}
	// R296 — the region the completion verb clears must exist and be unique. Checked here
	// because the write path checks it only when someone runs the verb, and the damage that
	// prompted this sat green in between.
	if err := parser.CheckActive(filepath.Join(repoRoot, "CURRENT.md")); err != nil {
		t.Structure = append(t.Structure, err.Error())
	}
	t.checkReaderAgreement(traj, q)
	t.checkCarveToQueue(traj, carves)
	t.checkQueueToCarve(q, carves)
	t.checkDuplicateIDs(q)
	t.checkOrphans(q, carves)
	t.checkLedgerConformance(q)
	t.checkLineAgreement(carves)
	t.checkNumbering(traj)
	t.checkStateless(carves)
	t.countUnreachable(carves)
	t.countUnread(q, carves)
	return t, nil
}

// Seq: seq-validate-trajectory.md#2.10 | R300
// checkReaderAgreement compares the IDs the line scan read with the IDs the document reader
// returned entries for, per file, and names every ID one saw and the other did not.
//
// **Measured 2026-09-05 on this repository, the day this check was written:** the line scan
// read 58 IDs from the done file and the document reader returned 17 entries, reporting
// nothing unread — one unclosed backtick in an entry body absorbed the remaining 41 entries
// into a single text node. Every check downstream of the document reader then reported
// four landed parts as orphans. A check that reads through one parser cannot see what that
// parser swallowed; only a second reader can, which is why this runs first.
func (t *TrajectoryIssues) checkReaderAgreement(traj parser.TrajectoryScan, q parser.QueueScan) {
	docIDs := map[string]map[int]bool{pendingFile: {}, doneFile: {}}
	for _, e := range q.Pending {
		docIDs[pendingFile][e.ID] = true
	}
	for _, e := range q.Done {
		for _, id := range e.IDs {
			docIDs[doneFile][id] = true
		}
	}
	for _, f := range traj.Files {
		lineIDs := map[int]bool{}
		for _, id := range f.IDs {
			lineIDs[id] = true
		}
		t.disagree(f.Name, "the line scan read but the document reader returned no entry for", lineIDs, docIDs[f.Name])
		t.disagree(f.Name, "the document reader returned but the line scan did not read", docIDs[f.Name], lineIDs)
	}
}

// disagree names the IDs in a and not in b, sorted, as one finding.
func (t *TrajectoryIssues) disagree(file, what string, a, b map[int]bool) {
	var ids []int
	for id := range a {
		if !b[id] {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return
	}
	sort.Ints(ids)
	t.ReaderDisagreement = append(t.ReaderDisagreement, fmt.Sprintf(
		"%s: %d item ID(s) %s — %s. An unclosed span or fence above them is the usual cause; every finding below reads through the document reader",
		file, len(ids), what, idList(ids)))
}

// Seq: seq-validate-trajectory.md#2.1 | R287
// checkCarveToQueue resolves every citation a carve makes against both queue files.
//
// Both, not either: most cited items have completed, so resolving against the pending file
// alone reports a dangling citation for every landed part.
func (t *TrajectoryIssues) checkCarveToQueue(traj parser.TrajectoryScan, carves parser.CarveScan) {
	known := idSet(traj)
	for _, c := range carves.Carves {
		for _, p := range c.Parts {
			id := p.QueueID()
			if id != 0 && !known[id] {
				t.Dangling = append(t.Dangling, fmt.Sprintf(
					"%s part %s cites #%d, which is in neither queue file", c.Path, keyOf(p), id))
			}
		}
	}
}

// Seq: seq-validate-trajectory.md#2.2 | R288
// checkQueueToCarve is the direction the carve side structurally cannot see.
func (t *TrajectoryIssues) checkQueueToCarve(q parser.QueueScan, carves parser.CarveScan) {
	for _, e := range q.Pending {
		if e.Kind != minispecsdom.SourcePart {
			continue
		}
		if msg := missingPart(e.SourceDoc, e.SourceKey, e.Line, carves, pendingFile); msg != "" {
			t.MissingParts = append(t.MissingParts, msg)
		}
	}
}

// Seq: seq-validate-trajectory.md#2.3 | R290
// checkDuplicateIDs reports an ID held by **both** files — a live item and a completed one
// sharing a number. Repetition *within* the done file is deliberately not a finding: an item
// that lands in stages is legitimately recorded across several entries, and from the number
// alone that is indistinguishable from a reuse.
func (t *TrajectoryIssues) checkDuplicateIDs(q parser.QueueScan) {
	pending := map[int]int{}
	for _, e := range q.Pending {
		if _, seen := pending[e.ID]; !seen {
			pending[e.ID] = e.Line
		}
	}
	done := firstLines(q.Done)
	var ids []int
	for id := range pending {
		if _, both := done[id]; both {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	for _, id := range ids {
		t.Duplicates = append(t.Duplicates, fmt.Sprintf(
			"#%d is live in %s:%d and completed in %s:%d — an ID is assigned once "+
				"and never reused", id, pendingFile, pending[id], doneFile, done[id]))
	}
}

// Seq: seq-validate-trajectory.md#2.4 | R291
// checkOrphans reports both forms. The second is the one nobody catches by eye.
func (t *TrajectoryIssues) checkOrphans(q parser.QueueScan, carves parser.CarveScan) {
	done := map[int]bool{}
	for _, e := range q.Done {
		for _, id := range e.IDs {
			done[id] = true
		}
	}
	for _, c := range carves.Carves {
		for _, p := range c.Parts {
			id := p.QueueID()
			if p.State() == parser.PartLanded && id != 0 && !done[id] {
				t.Orphans = append(t.Orphans, fmt.Sprintf(
					"%s part %s is landed against #%d, which has no done entry", c.Path, keyOf(p), id))
			}
		}
	}
	for _, e := range q.Done {
		if msg := missingPart(e.PartDoc, e.PartKey, e.Line, carves, doneFile); msg != "" {
			t.Orphans = append(t.Orphans, msg)
		}
	}
}

// Seq: seq-validate-trajectory.md#2.5 | R292
// checkLedgerConformance names an entry whose header carries no identifier slot — not one
// that discharged no ID, which is legitimate. The distinction is the *shape* of the header.
func (t *TrajectoryIssues) checkLedgerConformance(q parser.QueueScan) {
	for _, e := range q.Done {
		if !e.HasSlot {
			t.Unmigrated = append(t.Unmigrated, fmt.Sprintf(
				"%s:%d has no identifier slot — migrate the header to "+
					"`- **YYYY-MM-DD — <ids>: <title>.**`", doneFile, e.Line))
		}
	}
}

// Seq: seq-validate-trajectory.md#2.6 | R294
// checkLineAgreement is a sentry over a corpus normalised by hand: the checkbox is
// authoritative and the strikethrough and marker must agree with it.
func (t *TrajectoryIssues) checkLineAgreement(carves parser.CarveScan) {
	for _, c := range carves.Carves {
		for _, p := range c.Parts {
			landed := p.State() == parser.PartLanded
			var why string
			switch {
			case landed && !p.Struck():
				why = "checked but not struck through"
			case !landed && p.Struck():
				why = "struck through but unchecked"
			case landed && hasVerb(p.Verbs(), openVerbs):
				why = "checked, but its marker says open"
			case !landed && hasVerb(p.Verbs(), doneVerbs):
				why = "unchecked, but its marker says landed"
			}
			if why != "" {
				t.Disagreements = append(t.Disagreements, fmt.Sprintf(
					"%s part %s is %s — the checkbox is authoritative", c.Path, keyOf(p), why))
			}
		}
	}
}

// Seq: seq-validate-trajectory.md#2.8 | R298
// checkStateless names every status-block line the reader could not read as a part **and**
// lists as deviating. The reader's own report rather than a second opinion — the independent
// cross-check is gap O18.
func (t *TrajectoryIssues) checkStateless(carves parser.CarveScan) {
	for _, c := range carves.Carves {
		for _, s := range c.Stateless {
			// A checkbox-less line carrying no deviation is the format's own shape — a SPLIT
			// or MOVED parent — and is not a finding; one carrying a deviation is a line the
			// reader could neither place nor pass, the census's listing rule (R216) applied
			// to a check.
			if len(s.Deviations) == 0 {
				continue
			}
			t.Stateless = append(t.Stateless, fmt.Sprintf("%s:%d — %s: %s", c.Path, s.Line, s.Reason, strings.TrimSpace(s.Text)))
		}
	}
}

// Seq: seq-validate-trajectory.md#1.4.2 | R293
// countUnreachable records what the integrity check could not see: a document keyed on the
// superseded bare-`#N` scheme carries its queue references in key position.
func (t *TrajectoryIssues) countUnreachable(carves parser.CarveScan) {
	for _, c := range carves.Carves {
		n := c.Unkeyed()
		if n == 0 {
			continue
		}
		t.Unreachable += n
		t.UnreachableDocs = append(t.UnreachableDocs, fmt.Sprintf("%s (%d)", c.Path, n))
	}
}

// Seq: seq-validate-trajectory.md#2.7 | R295
// checkNumbering reports numbers below the maximum that appear in no readable entry. The
// maximum comes from the scan's own accessor, which reads **both** files (R190).
func (t *TrajectoryIssues) checkNumbering(traj parser.TrajectoryScan) {
	known := idSet(traj)
	maxID := traj.MaxItemID()
	for id := 1; id <= maxID; id++ {
		if !known[id] {
			t.MissingIDs = append(t.MissingIDs, id)
		}
	}
}

// Seq: seq-validate-trajectory.md#2.9 | R297, R302
// countUnread records what each reader could not recognize, per file: the two queue files,
// the current file, and every carve.
func (t *TrajectoryIssues) countUnread(q parser.QueueScan, carves parser.CarveScan) {
	record := func(name string, n int) {
		if n == 0 {
			return
		}
		if t.Unread == nil {
			t.Unread = map[string]int{}
		}
		t.Unread[name] = n
	}
	record(pendingFile, len(q.PendingUnread))
	record(doneFile, len(q.DoneUnread))
	record(currentFile, len(q.CurrentUnread))
	for _, c := range carves.Carves {
		record(c.Path, len(c.Unread))
	}
}

// missingPart checks one entry's part pointer against the carve it names. A pointer at a
// document outside carves/ names no part; a pointer at a carve the scan did not list is left
// to the carve directory's own absence.
func missingPart(doc, key string, line int, carves parser.CarveScan, file string) string {
	if doc == "" || key == "" || !strings.Contains(doc, "carves/") {
		return ""
	}
	for _, c := range carves.Carves {
		if !strings.HasSuffix(doc, c.Path) && !strings.HasSuffix(c.Path, doc) {
			continue
		}
		for _, p := range c.Parts {
			if p.Keyed() && p.Key() == key {
				return ""
			}
		}
		return fmt.Sprintf("%s:%d claims %s part %s, which that carve's status block does not hold",
			file, line, c.Path, key)
	}
	return ""
}

// firstLines maps each ID a file's entries carry to the line of the first entry carrying
// it — an item that lands in stages is recorded across several, and the finding names where
// to start looking.
func firstLines(entries []parser.DoneEntry) map[int]int {
	out := map[int]int{}
	for _, e := range entries {
		for _, id := range e.IDs {
			if _, seen := out[id]; !seen {
				out[id] = e.Line
			}
		}
	}
	return out
}

func idSet(traj parser.TrajectoryScan) map[int]bool {
	out := map[int]bool{}
	for _, f := range traj.Files {
		for _, id := range f.IDs {
			out[id] = true
		}
	}
	return out
}

// idList renders item IDs as `#1 #2 #3`, the form every finding that names a set of them uses.
func idList(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = fmt.Sprintf("#%d", id)
	}
	return strings.Join(parts, " ")
}

func hasVerb(verbs []string, set map[string]bool) bool {
	for _, v := range verbs {
		if set[v] {
			return true
		}
	}
	return false
}

func keyOf(p parser.Part) string {
	if !p.Keyed() {
		return "(unkeyed)"
	}
	return p.Key()
}

// CRC: crc-TrajectoryValidate.md | Seq: seq-validate-trajectory.md#1.4 | R286, R293, R299
// FormatText emits issues only, and always says what it could not read. The closing notes
// are not decoration: a report that names a migration target *looks* complete, and a reader
// takes its silence about integrity as a pass.
func (t *TrajectoryIssues) FormatText() string {
	if t.Absent {
		return "no trajectory layer here: no queue files and no carve directory\n" +
			"phase: validate trajectory OK\n"
	}
	var body strings.Builder
	for _, sec := range []struct {
		label string
		items []string
	}{
		// R298 — first, because every check below reads through the parse this one reports on.
		// R300 — first of all: when the two readers disagree, every finding below is a
		// statement about the part of the file the document reader saw.
		{"the two readers of the queue files disagree", t.ReaderDisagreement},
		{"status-block lines not read as parts", t.Stateless},
		// R296 — a file whose own shape is wrong makes every reference-level finding below it
		// a statement about a document nobody can trust.
		{"trajectory file structure", t.Structure},
		{"dangling citations", t.Dangling},
		{"queue entries whose part is missing", t.MissingParts},
		{"reused item IDs", t.Duplicates},
		{"orphans", t.Orphans},
		{"unmigrated done entries", t.Unmigrated},
		{"lines whose markings disagree", t.Disagreements},
		{"item numbers in no readable entry", t.missingIDLines()},
	} {
		if len(sec.items) == 0 {
			continue
		}
		fmt.Fprintf(&body, "  %s:\n", sec.label)
		for _, it := range sec.items {
			fmt.Fprintf(&body, "    %s\n", it)
		}
	}
	notes := t.unreadNote() + t.unreachableNote()
	if body.Len() > 0 {
		return "issues:\n" + body.String() + notes + "phase: validate trajectory FAILED\n"
	}
	return notes + "phase: validate trajectory OK\n"
}

// missingIDLines renders R295's finding, naming the unread count beside it when there is
// one: a gap in a fully-readable ledger is a different finding from a gap in one the reader
// could only partly see, and the report must not let them look alike.
func (t *TrajectoryIssues) missingIDLines() []string {
	if len(t.MissingIDs) == 0 {
		return nil
	}
	line := idList(t.MissingIDs)
	if unread := t.unreadTotal(); unread > 0 {
		line += fmt.Sprintf("\n      %d entry-like line(s) went unrecognized, which very likely "+
			"explains these — see the coverage note below", unread)
	}
	return []string{line}
}

func (t *TrajectoryIssues) unreadTotal() int {
	n := 0
	for _, v := range t.Unread {
		n += v
	}
	return n
}

// unreadNote is R297's coverage statement, printed whether or not anything else fired.
func (t *TrajectoryIssues) unreadNote() string {
	if len(t.Unread) == 0 {
		return ""
	}
	names := make([]string, 0, len(t.Unread))
	for name := range t.Unread {
		names = append(names, name)
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = fmt.Sprintf("%s (%d)", name, t.Unread[name])
	}
	return fmt.Sprintf(
		"note: %d line(s) were not read — %s. An entry-like line outside the recognized\n"+
			"      shape, or a bracket group never closed, which takes the rest of its file with it;\n"+
			"      every shape-based check above is blind to them by construction.\n",
		t.unreadTotal(), strings.Join(parts, ", "))
}

func (t *TrajectoryIssues) unreachableNote() string {
	if t.Unreachable == 0 {
		return ""
	}
	return fmt.Sprintf(
		"note: %d citation(s) unreachable — %s key on a superseded scheme, so their queue\n"+
			"      references sit in key position where the position rule cannot read them.\n"+
			"      Migrate to `Item N` and they become checkable.\n",
		t.Unreachable, strings.Join(t.UnreachableDocs, ", "))
}
