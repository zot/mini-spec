// CRC: crc-TrajectoryValidate.md | R286, R287, R288, R290, R291, R292, R293, R294, R297, R302
package validate

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// repo builds a repository root from a map of relative path to contents. Directories are
// created as needed; a file absent from the map is absent from the tree, which is how the
// "no trajectory layer" case is expressed.
func repo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func run(t *testing.T, files map[string]string) *TrajectoryIssues {
	t.Helper()
	got, err := RunTrajectory(repo(t, files))
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func joined(items []string) string { return strings.Join(items, "\n") }

// R287. Citations resolve against **both** queue files.
//
// Most cited items have completed, so resolving against the pending file alone reports a
// dangling citation for every landed part — R190's one-file failure through a third door.
func TestDanglingCitationIsFound(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 1. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #2: a thing.** (`abc1234`)\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [ ] **Item 3 — a part.** **OPEN (#99.)**\n" +
			"- [x] ~~**Item 4 — landed.**~~ **LANDED (`abc1234`, 2026-08-16 — `#2`.)**\n",
	})
	if len(got.Dangling) != 1 {
		t.Fatalf("got %d dangling, want 1: %s", len(got.Dangling), joined(got.Dangling))
	}
	if !strings.Contains(got.Dangling[0], "#99") || !strings.Contains(got.Dangling[0], "carves/x.md") {
		t.Errorf("finding does not name the carve and the citation: %s", got.Dangling[0])
	}
	// The done-file citation must resolve, or the check is useless on landed parts.
	if strings.Contains(joined(got.Dangling), "#2") {
		t.Error("a citation resolving to the done file was reported dangling")
	}
}

// R288. The direction the carve side structurally cannot see.
func TestQueueEntryMissingFromItsCarveIsFound(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 1. **live**. Active.\n" +
			"   Source: [carves/x.md](carves/x.md), part `#7`.\n",
		"DONE.md":     "# Done\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n- [ ] **Item 3 — a part.** **OPEN (#1.)**\n",
	})
	if len(got.MissingParts) != 1 {
		t.Fatalf("got %d missing-part findings, want 1: %s", len(got.MissingParts), joined(got.MissingParts))
	}
	if !strings.Contains(got.MissingParts[0], "part 7") {
		t.Errorf("finding does not name the key it claims: %s", got.MissingParts[0])
	}
}

// R289. Citations are ingested by position, never swept from prose.
//
// The fixture is this repository's own line 75, verbatim: a prose sentence quoting ark's
// shape as an example, inside backticks. It is not a strawman — it is the real defect, run
// and believed on 2026-08-16 while scoping this item, and the fifth ad-hoc instrument in
// this layer's history to return a confident wrong answer.
func TestBackquotedProseExampleIsNotACitation(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 1. **live**. Active.\n",
		"DONE.md":    "# Done\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"Ark's live carves already read this way — `**Item 8 — a test harness…** **OPEN (#121.)**`.\n\n" +
			"- [ ] **Item 3 — a real part.** **OPEN (#1.)**\n",
	})
	if len(got.Dangling) != 0 {
		t.Errorf("a backquoted prose example was read as a citation: %s", joined(got.Dangling))
	}
}

// R286. Nothing that could be inconsistent is a clean result — and not "could not check".
func TestNoTrajectoryLayerPasses(t *testing.T) {
	got := run(t, map[string]string{"README.md": "# just a repository\n"})
	if !got.Absent {
		t.Error("a repository with no trajectory layer was not reported absent")
	}
	if got.HasIssues() {
		t.Errorf("an absent layer produced issues: %+v", got)
	}
	// The report must *say* so. A silent clean result over an absent layer reads exactly
	// like a clean result over a checked one.
	if !strings.Contains(got.FormatText(), "no trajectory layer here") {
		t.Errorf("the report does not state that the layer is absent:\n%s", got.FormatText())
	}
}

// R290. An ID held by both files is a collision; repetition within the done file is not.
//
// The second half is a correction, not a simplification. The first draft checked
// within-file too and, run against ark, reported #41 (five staged entries), #65 and #98 as
// reused — all three correct work. A check that fires on correct work gets muted.
func TestOnlyCrossFileIDCollisionIsReported(t *testing.T) {
	staged := run(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md": "# Done\n\n" +
			"- **2026-07-20 — #41: pass 1.** (`aaa`)\n" +
			"- **2026-07-20 — #41: pass 2.** (`bbb`)\n" +
			"- **2026-07-20 — #41: closed.** (`ccc`)\n",
	})
	if len(staged.Duplicates) != 0 {
		t.Errorf("a staged multi-entry record was reported as reuse: %s", joined(staged.Duplicates))
	}

	collided := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 8. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #8: already done.** (`abc1234`)\n",
	})
	if len(collided.Duplicates) != 1 {
		t.Fatalf("got %d collisions, want 1: %s", len(collided.Duplicates), joined(collided.Duplicates))
	}
	for _, want := range []string{"PENDING.md", "DONE.md", "#8"} {
		if !strings.Contains(collided.Duplicates[0], want) {
			t.Errorf("collision does not name %q: %s", want, collided.Duplicates[0])
		}
	}
}

// R291. Both orphan forms. The second is the one nobody catches by eye.
func TestBothOrphanFormsAreFound(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 6. **live**. Active.\n",
		"DONE.md": "# Done\n\n" +
			"- **2026-08-16 — #9: a thing.** (`abc1234`) Part `carves/x.md#42`.\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [x] ~~**Item 3 — landed.**~~ **LANDED (`abc1234`, 2026-08-16 — `#5`.)**\n" +
			// An *open* part citing a live item has no done entry and is not an orphan — the
			// landed guard is what says so. Added 2026-09-05 after a probe past the alarm list
			// dropped that guard and nothing objected.
			"- [ ] **Item 4 — open and live.** **OPEN (#6.)**\n",
	})
	if len(got.Orphans) != 2 {
		t.Fatalf("got %d orphans, want 2 (one of each form): %s", len(got.Orphans), joined(got.Orphans))
	}
	all := joined(got.Orphans)
	if !strings.Contains(all, "#5") {
		t.Error("a part landed against an ID with no done entry was not reported")
	}
	if !strings.Contains(all, "42") {
		t.Error("a done entry naming a part no carve records was not reported")
	}
}

// R292. The *shape* of the header is what distinguishes unmigrated from legitimately empty.
func TestDoneEntryWithNoIdentifierSlotIsNamed(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md": "# Done\n\n" +
			"- **2026-08-16 — #7: conforming.** (`abc1234`)\n" +
			"- **2026-07-19 — PENDING #46 CLOSED (operation-object rollout).** ark's real shape\n" +
			"- **2026-08-16 — O201: a gap closed with no queue ID.** (`bbb`)\n",
	})
	if len(got.Unmigrated) != 1 {
		t.Fatalf("got %d unmigrated, want 1: %s", len(got.Unmigrated), joined(got.Unmigrated))
	}
	// The third entry has a slot and legitimately discharged no queue ID. Reporting it is
	// the false-positive direction, which is the one that gets a check muted.
	if strings.Contains(got.Unmigrated[0], "O201") {
		t.Errorf("an entry that legitimately discharged no queue ID was reported: %s", got.Unmigrated[0])
	}
}

// R294. A sentry over a corpus that is already clean — these fixtures are the only place
// it will ring until something drifts, which is exactly when an untested guard rots.
func TestDisagreeingMarkingsAreFound(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md":    "# Done\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [x] ~~**Item 1 — agrees.**~~ **LANDED (`abc1234`, 2026-08-16 — `#1`.)**\n" +
			"- [x] **Item 2 — checked but not struck.** **LANDED (`abc1234`, 2026-08-16.)**\n" +
			"- [ ] ~~**Item 3 — struck but unchecked.**~~ **OPEN (not queued.)**\n" +
			"- [ ] **Item 4 — unchecked, marker says landed.** **LANDED (`abc1234`, 2026-08-16.)**\n" +
			"- [x] ~~**Item 5 — checked, marker says open.**~~ **OPEN (not queued.)**\n",
	})
	// Four, one per branch. Item 5 was added 2026-08-16 after re-pulling this alarm against
	// the simplified code: killing the landed-with-open-marker branch alone left the test
	// green, because no fixture reached it. The earlier pull had rung only because it killed
	// both marker branches at once and the *other* one was covered — a proof that looked
	// good and was covering half of what it claimed.
	if len(got.Disagreements) != 4 {
		t.Fatalf("got %d disagreements, want 4: %s", len(got.Disagreements), joined(got.Disagreements))
	}
	if strings.Contains(joined(got.Disagreements), "part 1 ") {
		t.Error("the agreeing line was reported")
	}
	if !strings.Contains(joined(got.Disagreements), "marker says landed") {
		t.Error("an unchecked line carrying a done-class marker was not reported")
	}
	if !strings.Contains(joined(got.Disagreements), "marker says open") {
		t.Error("a checked line carrying an open-class marker was not reported")
	}
}

// R293. Conformance is named, and what could not be read is stated rather than passed over.
func TestUnreachableCitationsAreStated(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md":    "# Done\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [ ] **#72 `/mini-spec`** — the TUI itself, keyed on the superseded scheme\n",
	})
	if got.Unreachable != 1 {
		t.Fatalf("Unreachable = %d, want 1", got.Unreachable)
	}
	out := got.FormatText()
	// Both halves. Naming the migration target without stating the coverage gap *looks*
	// like a complete report, and a reader takes the silence about integrity as a pass.
	if !strings.Contains(out, "Item N") {
		t.Errorf("the report does not name the migration target:\n%s", out)
	}
	if !strings.Contains(out, "unreachable") {
		t.Errorf("the report does not state what it could not read:\n%s", out)
	}
}

// R295. Every number up to the maximum should appear in a readable entry.
//
// `trajectory-format.md` says a gap in the sequence is expected, which is true of a
// deliberately abandoned ID — and worth testing rather than assuming. Measured 2026-08-16,
// ark has 17 gaps and 16 of them are mentioned in its own queue files, so they are losses.
func TestItemNumbersInNoReadableEntryAreFound(t *testing.T) {
	contiguous := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 3. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #1: a.** (`aaa`)\n- **2026-08-16 — #2: b.** (`bbb`)\n",
	})
	if len(contiguous.MissingIDs) != 0 {
		t.Errorf("a contiguous ledger reported gaps: %v", contiguous.MissingIDs)
	}

	gapped := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 5. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #1: a.** (`aaa`)\n- **2026-08-16 — #3: c.** (`ccc`)\n",
	})
	if len(gapped.MissingIDs) != 2 || gapped.MissingIDs[0] != 2 || gapped.MissingIDs[1] != 4 {
		t.Errorf("MissingIDs = %v, want [2 4]", gapped.MissingIDs)
	}
	if !gapped.HasIssues() {
		t.Error("a gap in a fully-readable ledger is a finding, not a note")
	}
}

// R297. Coverage, not a defect count — and the honest half of R295.
//
// A shape-based check is blind by construction to a line outside the shape, so it reports
// clean over everything it never saw. Measured 2026-08-16 in ark: 139 entry-like lines in
// the done file and 54 recognized.
func TestUnrecognizedEntryLinesAreCounted(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 2. **live**. Active.\n" +
			"## add to the list\n", // ark's real shape: an item with no queue number
		"DONE.md": "# Done\n\n" +
			"- **2026-08-16 — #1: recognized.** (`aaa`)\n" +
			"- **2026-08-16 — #4: also recognized.** (`ccc`)\n" +
			"- 2026-07-15 — **the older shape, date outside the bold (#3)** (`bbb`)\n",
	})
	if got.Unread["DONE.md"] != 1 || got.Unread["PENDING.md"] != 1 {
		t.Fatalf("Unread = %v, want one per file", got.Unread)
	}

	out := got.FormatText()
	if !strings.Contains(out, "were not read") {
		t.Errorf("the report does not state its coverage:\n%s", out)
	}
	// The pairing is the point. #3 sits in the unrecognized line, so it shows as a gap —
	// and the gap report must say the unread lines very likely explain it, or a reader
	// takes a partial read for a complete one.
	//
	// Note the limit this fixture had to be built around, which is real: an unrecognized
	// entry holding the *highest* ID does not raise the maximum, so it leaves no gap and
	// the numbering check cannot see it. Only the coverage count can. That is why R297 is
	// the honest half and R295 the symptom, rather than the other way round.
	if !strings.Contains(out, "very likely explains these") {
		t.Errorf("the gap finding does not name the unread lines that explain it:\n%s", out)
	}
}

// R299. The machine-readable form uses one key convention.
//
// Written because the sibling command shipped the opposite: `query carves` emitted Go
// field names beside snake_case in one document until gap O18 caught it, and nothing
// validates JSON shape. This is that defect class, guarded at the point it would recur.
func TestJSONKeysAreSnakeCase(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 3. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #1: a.** (`aaa`)\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [ ] **#72 `/mini-spec`** — keyed on the superseded scheme\n",
	})
	blob, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(blob, &keys); err != nil {
		t.Fatal(err)
	}
	for k := range keys {
		if k != strings.ToLower(k) || strings.Contains(k, " ") {
			t.Errorf("key %q is not snake_case — a Go field name reached the JSON", k)
		}
	}
	for _, want := range []string{"missing_ids", "unreachable", "absent"} {
		if _, ok := keys[want]; !ok {
			t.Errorf("key %q is missing from the machine-readable form; got %v", want, keysOf(keys))
		}
	}
}

func keysOf(m map[string]json.RawMessage) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// write puts one trajectory file in place for a case built from nothing.
func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// R296 — `CURRENT.md` must carry exactly one `## Active`, and this is where that is noticed.
//
// **`activeRange` has always refused both shapes — at write time only.** Measured 2026-08-19, the
// check ran green twice in one session over a file a completion verb had just damaged; the
// trajectory files are gitignored, so git cannot diff them, and the backup slot holds one level of
// undo which the next operation spends. A check between operations stands where a diff would.
func TestTrajectoryReportsCurrentWithoutExactlyOneActive(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"no Active at all", "# Current\n\nstanding context only\n", "carries no"},
		{"two Actives", "# Current\n\n## Active\n\na\n\n## Active\n\nb\n", "more than one"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, "PENDING.md", "# Pending\n")
			write(t, dir, "CURRENT.md", tc.body)
			got, err := RunTrajectory(dir)
			if err != nil {
				t.Fatal(err)
			}
			if !got.HasIssues() {
				t.Fatalf("%s reported OK", tc.name)
			}
			if len(got.Structure) != 1 || !strings.Contains(got.Structure[0], tc.want) {
				t.Errorf("Structure = %v, want one finding naming %q", got.Structure, tc.want)
			}
		})
	}

	// Exactly one is clean, and a file that is not there at all is not this finding: absence is
	// already answered, and two checks giving one fact different words is worse than either.
	dir := t.TempDir()
	write(t, dir, "PENDING.md", "# Pending\n")
	write(t, dir, "CURRENT.md", "# Current\n\n## Active\n\nnothing in progress\n")
	got, err := RunTrajectory(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Structure) != 0 {
		t.Errorf("exactly one Active reported %v", got.Structure)
	}
}

// R298. A status-block line the reader could not read as a part **and** lists as deviating is
// an issue, not a note; a checkbox-less SPLIT parent carries no deviation and is not one.
func TestAStatelessLineInTheStatusBlockIsReported(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 1. **live**. Active.\n",
		"DONE.md":    "# Done\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n" +
			"- [ ] **Item 1 — a part.** **OPEN (#1.)**\n" +
			"- **Item 8 — a split parent, checkbox-less by the format's own rule.** **SPLIT (Bill, 2026-08-14.)**\n" +
			"- a bullet with no checkbox and no head, which is a line the reader cannot place\n",
	})
	if len(got.Stateless) != 1 {
		t.Fatalf("got %d stateless findings, want 1: %s", len(got.Stateless), joined(got.Stateless))
	}
	if !got.HasIssues() {
		t.Error("a stateless line did not gate the check")
	}
	if !strings.Contains(got.Stateless[0], "carves/x.md:7") {
		t.Errorf("the finding does not name the file and line: %s", got.Stateless[0])
	}
}

// R300. The line scan is the document reader's second opinion. **The fixture is this
// repository's own ledger shape on 2026-09-05:** an entry whose body carries an unclosed
// backtick, after which the document reader returned nothing and reported nothing unread,
// while the line scan went on reading headers.
func TestTheTwoReadersOfTheQueueFilesMustAgree(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md": "# Pending\n",
		"DONE.md": "# Done\n\n---\n\n" +
			"- **2026-08-24 — #3: a title.** (`aaa`)\n" +
			"  a body with an unclosed `span that runs on\n" +
			"- **2026-08-16 — #2: b.** (`bbb`)\n" +
			"- **2026-08-16 — #1: a.** (`ccc`)\n",
	})
	if len(got.ReaderDisagreement) == 0 {
		// Either the dependency now reads all three — in which case this fixture no longer
		// isolates the case and needs a new one — or the check is gone.
		t.Fatal("the two readers disagreed about the done file and nothing said so")
	}
	if !got.HasIssues() {
		t.Error("a reader disagreement did not gate the check")
	}
	if !strings.Contains(got.ReaderDisagreement[0], "#1") || !strings.Contains(got.ReaderDisagreement[0], "#2") {
		t.Errorf("the finding does not name the swallowed IDs: %s", got.ReaderDisagreement[0])
	}
	agree := run(t, map[string]string{
		"PENDING.md": "# Pending\n\n## 3. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-16 — #1: a.** (`aaa`)\n- **2026-08-16 — #2: b.** (`bbb`)\n",
	})
	if len(agree.ReaderDisagreement) != 0 {
		t.Errorf("readers that agree were reported: %v", agree.ReaderDisagreement)
	}
}

// R302. The coverage note names every file a reader could leave partly unread: the current
// file and each carve, beside the two queue files. A fence never closed in a carve takes
// its later parts out of every check above, and this count is the only one that says so.
func TestCurrentFileAndCarveUnreadAreCounted(t *testing.T) {
	got := run(t, map[string]string{
		"PENDING.md":  "# Pending\n\n---\n\n## 2. **live**. Active.\n   Source: [carves/x.md](carves/x.md), part `#1`.\n",
		"DONE.md":     "# Done\n",
		"CURRENT.md":  "# Current\n\n---\n\n## Active\n\n#2 — live.\n\n```\nnever closed\n",
		"carves/x.md": "# Carve: x\n\n## Status\n\n- [ ] **Item 1 — live.** **OPEN (#2.)**\n\n## Item 1\n\n`never closed\n",
	})
	if got.Unread["CURRENT.md"] != 1 || got.Unread["carves/x.md"] != 1 {
		t.Fatalf("Unread = %v, want one for the current file and one for the carve", got.Unread)
	}
	out := got.FormatText()
	if !strings.Contains(out, "CURRENT.md (1)") || !strings.Contains(out, "carves/x.md (1)") {
		t.Errorf("the note does not name both files:\n%s", out)
	}
}

// R485, R486, R487 — links a cloner cannot follow fail the phase; untracked and no-git are notes.
func TestLinksACloneCannotFollowFailThePhase(t *testing.T) {
	files := map[string]string{
		"PENDING.md":         "# Pending\n\n---\n",
		"DONE.md":            "# Done\n\n---\n",
		"CURRENT.md":         "# Current\n\n---\n\n## Active\n\n_No active item._\n",
		"tracked.md":         "t\n",
		"fresh.md":           "u\n",
		"private/p.md":       "p\n",
		".gitignore":         "private/\nPENDING.md\nDONE.md\nCURRENT.md\n",
		"carves/x.md":        "# Carve: x\n\n## Status\n\n- [ ] **Item 1 — a.** **OPEN (not queued.)**\n\nsee [t](../tracked.md), [u](../fresh.md), [p](../private/p.md), [m](../missing.md) and ` [f](../nope.md) `\n",
		"carves/done/old.md": "# Carve: old\n\n## Status\n\n- [x] ~~**Item 1 — b.**~~ **LANDED (2026-09-01 — `#1`.)**\n\n[x](../tool/x.md)\n",
	}
	root := repo(t, files)
	for _, args := range [][]string{{"init", "-q"}, {"add", "tracked.md", ".gitignore", "carves"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v: %s", err, out)
		}
	}
	got, err := RunTrajectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Links) != 3 || !got.HasIssues() {
		t.Errorf("want three findings (ignored, missing, and the done carve's), got %v", got.Links)
	}
	found := joined(got.Links)
	for _, want := range []string{
		"carves/x.md:7  [p](../private/p.md)  ignored",
		"carves/x.md:7  [m](../missing.md)  missing",
		"carves/done/old.md:7  [x](../tool/x.md)  missing",
	} {
		if !strings.Contains(found, want) {
			t.Errorf("missing finding %q in %v", want, got.Links)
		}
	}
	if len(got.LinkNotes) != 1 || !strings.Contains(got.LinkNotes[0], "[u](../fresh.md)  untracked") {
		t.Errorf("want one untracked note, got %v", got.LinkNotes)
	}
	text := got.FormatText()
	if !strings.Contains(text, "links a cloner cannot follow:") || !strings.HasSuffix(text, "FAILED\n") {
		t.Errorf("report:\n%s", text)
	}

	plain := run(t, files)
	if len(plain.Links) != 0 || len(plain.LinkNotes) != 1 || !strings.Contains(plain.LinkNotes[0], "unclassified") {
		t.Errorf("without git: want no findings and the unclassified note, got %v / %v", plain.Links, plain.LinkNotes)
	}
	if !strings.Contains(plain.FormatText(), "note: links") {
		t.Errorf("the note is not printed:\n%s", plain.FormatText())
	}
}
