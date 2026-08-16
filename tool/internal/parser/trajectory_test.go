// CRC: crc-Trajectory.md | R190, R194, R195, R197
package parser

import (
	"os"
	"path/filepath"
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
	doneHigh := "# Done\n\n- **2026-08-14 — a thing (`#9`).** `abc1234`.\n"
	doneLow := "# Done\n\n- **2026-08-14 — a thing (`#3`).** `abc1234`.\n"

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

// R190. The alarm: scan every line of a done entry instead of its header and this goes
// red with 100. Not hypothetical -- this project's own `#8` ledger entry cites `#7` in
// its body prose.
func TestCitationInDoneEntryBodyDoesNotRaiseTheMaximum(t *testing.T) {
	done := "# Done\n\n" +
		"- **2026-08-14 — a thing (`#4`).** `abc1234`.\n" +
		"  It reuses the `#99` shape described earlier, scoped smaller.\n"
	dir := writeTrajectory(t, map[string]string{"PENDING.md": "# Pending\n", "DONE.md": done})
	if got := nextItem(t, dir); got != 5 {
		t.Errorf("next = %d, want 5 (a citation in entry prose must not count)", got)
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
	done := "# Done\n\n- **2026-08-14 — a thing (`#6`).** `abc1234`.\n"
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
		"- **2026-08-14 — a (`#1`).** `aaa`.\n" +
		"- **2026-08-13 — b (`#2`).** `bbb`.\n" +
		"- **2026-08-12 — c (`#3`).** `ccc`.\n"
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
