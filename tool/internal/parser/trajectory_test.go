// CRC: crc-Trajectory.md | R190, R194, R195, R197
package parser

import (
	"github.com/zot/simple-dom/minispecsdom"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// writeTrajectory lays out a repository root with whichever files are named. A file
// absent from the map is absent from the tree, which is the case the missing-file
// requirements are about.
func writeTrajectory(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func nextItem(t *testing.T, dir string) int {
	t.Helper()
	scan, err := ScanTrajectory(dir)
	if err != nil {
		t.Fatalf("ScanTrajectory: %v", err)
	}
	return scan.MaxItemID() + 1
}

// R190. The alarm: make MaxItemID read the pending file alone and the first case goes
// red while the mirror still passes -- which is why both directions are here. One
// direction cannot detect a one-file read.
func TestMaxItemIDSpansBothFiles(t *testing.T) {
	pendingHigh := "# Pending\n\n## 9. **live thing**. Active.\n"
	pendingLow := "# Pending\n\n## 3. **live thing**. Active.\n"
	doneHigh := "# Done\n\n- **2026-08-14 — #9: a thing.** (`abc1234`) Part `carves/x.md#2`.\n"
	doneLow := "# Done\n\n- **2026-08-14 — #3: a thing.** (`abc1234`) Part `carves/x.md#2`.\n"

	for _, tc := range []struct {
		name          string
		pending, done string
	}{
		{"highest is in the done file", pendingLow, doneHigh},
		{"highest is in the pending file", pendingHigh, doneLow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := writeTrajectory(t, map[string]string{"PENDING.md": tc.pending, "DONE.md": tc.done})
			if got := nextItem(t, dir); got != 10 {
				t.Errorf("next = %d, want 10", got)
			}
		})
	}
}

// R190. The alarm: drop the header guard in parseDoneIDs and this goes red with 100.
// Not hypothetical -- the body line below is ark's real shape, an older pending entry
// quoted verbatim inside a later done entry, and five such lines sit in its ledger.
func TestCitationInDoneEntryBodyDoesNotRaiseTheMaximum(t *testing.T) {
	done := "# Done\n\n" +
		"- **2026-08-14 — #4: a thing.** (`abc1234`) Part `carves/x.md#2`.\n" +
		"  Supersedes the queue entry it grew out of:\n" +
		"  - 2026-07-06 — **PENDING #99 — the older shape: seeds and scope.**\n"
	dir := writeTrajectory(t, map[string]string{"PENDING.md": "# Pending\n", "DONE.md": done})
	if got := nextItem(t, dir); got != 5 {
		t.Errorf("next = %d, want 5 (a citation in entry prose must not count)", got)
	}
}

// R190. The identifier slot holds whatever the entry discharged -- a queue ID, a gap ID,
// a requirement range, several separated by `/`, or nothing at all -- so only the `#N`s
// in it count. Cases 1 and 2 are verbatim from ark's ledger. The rest pass by default
// under any parser that scans the whole header: an entry with no identifiers must
// contribute none rather than a number scavenged out of its title, and the `Part` pointer
// the adopted format appends is spelled exactly like a queue ID while naming a part key.
//
// The alarm: scan the whole header line rather than the slot. Goes red with
// `[117 84 83 4 13 500]`, where 13 is a part key and 500 is prose.
func TestDoneIdentifierSlotVariants(t *testing.T) {
	done := "# Done\n\n" +
		"- **2026-08-02 — O201 / R3399: a bad search request is 400, not 500.** (`037ce42`)\n" +
		"- **2026-08-02 — #117 / R3398: a failed proxy is not an absent server.** (`037ce42`)\n" +
		"- **2026-07-28 — #84 / #83: two items in one landing.** (`bbb`)\n" +
		"- **2026-08-16 — #4: a part key is not a queue ID.** (`ccc`) Part `carves/x.md#13`.\n" +
		"- **2026-08-16 — an incident that discharged no ID.** (`ddd`)\n" +
		"- **2026-08-16 — a title carrying a colon: and #500 after it.** (`eee`)\n"
	dir := writeTrajectory(t, map[string]string{"PENDING.md": "# Pending\n", "DONE.md": done})
	scan, err := ScanTrajectory(dir)
	if err != nil {
		t.Fatalf("ScanTrajectory: %v", err)
	}
	want := []int{117, 84, 83, 4}
	for _, f := range scan.Files {
		if f.Name != "DONE.md" {
			continue
		}
		if !slices.Equal(f.IDs, want) {
			t.Errorf("DONE.md contributed %v, want %v", f.IDs, want)
		}
	}
	if got := scan.MaxItemID(); got != 117 {
		t.Errorf("max = %d, want 117", got)
	}
}

// R194. Neither file present has no answer, and must not be reported as 1.
func TestNoTrajectoryFilesIsUnanswerable(t *testing.T) {
	dir := writeTrajectory(t, nil)
	scan, err := ScanTrajectory(dir)
	if err != nil {
		t.Fatalf("ScanTrajectory: %v", err)
	}
	if scan.AnyPresent() {
		t.Fatal("AnyPresent = true with no files written")
	}
	if got := scan.Missing(); len(got) != 2 {
		t.Errorf("Missing() = %v, want both files", got)
	}
}

// R195, R197. One file missing still answers, and names what it could not read.
func TestOneFileMissingAnswersAndSaysSo(t *testing.T) {
	done := "# Done\n\n- **2026-08-14 — #6: a thing.** (`abc1234`) Part `carves/x.md#2`.\n"
	dir := writeTrajectory(t, map[string]string{"DONE.md": done})
	scan, err := ScanTrajectory(dir)
	if err != nil {
		t.Fatalf("ScanTrajectory: %v", err)
	}
	if !scan.AnyPresent() {
		t.Fatal("AnyPresent = false with DONE.md written")
	}
	if got := scan.MaxItemID() + 1; got != 7 {
		t.Errorf("next = %d, want 7", got)
	}
	missing := scan.Missing()
	if len(missing) != 1 || missing[0] != "PENDING.md" {
		t.Errorf("Missing() = %v, want [PENDING.md]", missing)
	}
}

// R197. The counts are the evidence that separates a correct answer from a silent parse
// failure, so they are asserted per file rather than in aggregate.
func TestPerFileCountsAccompanyTheAnswer(t *testing.T) {
	pending := "# Pending\n\n## 8. **one**. Active.\n\n## 12. **two**. Queued.\n"
	done := "# Done\n\n" +
		"- **2026-08-14 — #1: a.** (`aaa`) Part `carves/x.md#1`.\n" +
		"- **2026-08-13 — #2: b.** (`bbb`) Part `carves/x.md#2`.\n" +
		"- **2026-08-12 — #3: c.** (`ccc`) Part `carves/x.md#3`.\n"
	dir := writeTrajectory(t, map[string]string{"PENDING.md": pending, "DONE.md": done})
	scan, err := ScanTrajectory(dir)
	if err != nil {
		t.Fatalf("ScanTrajectory: %v", err)
	}
	want := map[string]int{"PENDING.md": 2, "DONE.md": 3}
	for _, f := range scan.Files {
		if !f.Present {
			t.Errorf("%s reported absent", f.Name)
			continue
		}
		if len(f.IDs) != want[f.Name] {
			t.Errorf("%s contributed %d IDs, want %d", f.Name, len(f.IDs), want[f.Name])
		}
	}
}

// R192. NextGapNum moved out of update so a read-only caller could use it; this pins
// that the per-type sequences stay independent.
func TestNextGapNumIsPerType(t *testing.T) {
	gaps := []Gap{
		{ID: "O1", Type: "O"}, {ID: "O7", Type: "O"},
		{ID: "S2", Type: "S"},
	}
	if got := NextGapNum(gaps, "O"); got != 8 {
		t.Errorf("O = %d, want 8", got)
	}
	if got := NextGapNum(gaps, "S"); got != 3 {
		t.Errorf("S = %d, want 3", got)
	}
	if got := NextGapNum(gaps, "T"); got != 1 {
		t.Errorf("T = %d, want 1 (a type with no gaps starts at 1)", got)
	}
}

// R240 — the pending-entry adapter maps the dependency's fields, including a gap source.
func TestPendingEntriesReadThroughTheDependency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "PENDING.md")
	src := "# Pending\n\n---\n\n" +
		"## 3. **a part-sourced item** (mini-spec). Active.\n" +
		"   Source: [carves/x.md](carves/x.md), part `#Item 5`.\n\n" +
		"## 9. **a gap-sourced item**. Waiting.\n" +
		"   Source: [tool/design/design.md](tool/design/design.md), gap `O12`.\n"
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := PendingEntries(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d entries, want 2: %+v", len(got), got)
	}
	if got[0].ID != 3 || got[0].SourceDoc != "carves/x.md" || got[0].SourceKey != "Item 5" || got[0].Kind != minispecsdom.SourcePart || got[0].Line != 5 {
		t.Errorf("part entry = %+v", got[0])
	}
	if got[1].ID != 9 || got[1].SourceKey != "O12" || got[1].Kind != minispecsdom.SourceGap {
		t.Errorf("gap entry = %+v", got[1])
	}
	if none, err := PendingEntries(filepath.Join(dir, "absent.md")); err != nil || none != nil {
		t.Errorf("a missing file should be no entries and no error; got %v, %v", none, err)
	}
}
