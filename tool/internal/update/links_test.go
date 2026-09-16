// CRC: crc-LinkRepair.md | R489, R464, R465, R466, R468
package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repairRoot lays out both ends of a move: one link for each relocation, one that nothing
// resolves, one that two do, and links of other classes that must stay untouched.
func repairRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"tool/x.md":            "x\n",
		"carves/done/old.md":   "o\n",
		"carves/back.md":       "b\n",
		"carves/both.md":       "1\n",
		"carves/done/both.md":  "2\n",
		"carves/done/moved.md": "out [x](../tool/x.md#sec) and [o](done/old.md) and [n](nowhere.md) and [w](<done/old.md#w>)\n",
		"carves/live.md": "in [m](moved.md) and [b](done/back.md) and [two](sub/both.md)\n" +
			"kept [t](../tool/x.md) [e](https://x.example) [l](#here) and ` [f](done/moved.md) `\n",
		// `sub/both.md` from carves/ resolves two ways once `sub` is re-based: carves/both.md via
		// the citing dir moved into done/ (done/sub/both.md — absent) is not it; the target moved
		// across done/ gives carves/sub/done/both.md (absent) — so make the ambiguity real:
		"carves/sub/done/both.md": "3\n",
		"carves/done/sub/both.md": "4\n",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, root, "carves", "tool")
	return root
}

// gitInit makes root a repository and stages the named paths: the population is what the
// index holds, so a fixture without one has no public documents.
func gitInit(t *testing.T, root string, add ...string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	for _, args := range [][]string{{"init", "-q"}, append([]string{"add", "--"}, add...)} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// R464, R465, R466 — the four relocations, one each, and the two refusals.
func TestTheFourRelocationsAndTheTwoRefusals(t *testing.T) {
	root := repairRoot(t)
	r, err := RepairLinks(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]Considered{
		"../tool/x.md#sec": {Outcome: Rewritten, New: "../../tool/x.md#sec"},
		"done/old.md":      {Outcome: Rewritten, New: "old.md"},
		"done/old.md#w":    {Outcome: Rewritten, New: "<old.md#w>"},
		"moved.md":         {Outcome: Rewritten, New: "done/moved.md"},
		"done/back.md":     {Outcome: Rewritten, New: "back.md"},
		"nowhere.md":       {Outcome: Unresolvable},
		"sub/both.md":      {Outcome: Ambiguous},
	}
	if len(r.Considered) != len(want) {
		t.Fatalf("want %d considered, got %+v", len(want), r.Considered)
	}
	for _, c := range r.Considered {
		w := want[c.Old]
		if c.Outcome != w.Outcome || c.New != w.New {
			t.Errorf("%s: got %s %q, want %s %q", c.Old, c.Outcome, c.New, w.Outcome, w.New)
		}
	}
	if r.Counts[Rewritten] != 5 || r.Counts[Unresolvable] != 1 || r.Counts[Ambiguous] != 1 || !r.Unrepaired() {
		t.Errorf("counts %v unrepaired %v", r.Counts, r.Unrepaired())
	}
	moved := readFile(t, filepath.Join(root, "carves", "done", "moved.md"))
	if moved != "out [x](../../tool/x.md#sec) and [o](old.md) and [n](nowhere.md) and [w](<old.md#w>)\n" {
		t.Errorf("moved.md after repair:\n%s", moved)
	}
	live := readFile(t, filepath.Join(root, "carves", "live.md"))
	if !strings.HasPrefix(live, "in [m](done/moved.md) and [b](back.md) and [two](sub/both.md)\n") ||
		!strings.HasSuffix(live, "kept [t](../tool/x.md) [e](https://x.example) [l](#here) and ` [f](done/moved.md) `\n") {
		t.Errorf("live.md after repair:\n%s", live)
	}
}

// R464, R468 — only missing links are considered; a second run is a no-op byte for byte.
func TestOnlyMissingLinksAndASecondRunIsANoOp(t *testing.T) {
	root := repairRoot(t)
	if _, err := RepairLinks(root, nil); err != nil {
		t.Fatal(err)
	}
	live := filepath.Join(root, "carves", "live.md")
	before := readFile(t, live)
	st, _ := os.Stat(live)
	r, err := RepairLinks(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Counts[Rewritten] != 0 || len(r.Written) != 0 || len(r.Considered) != 2 { // only the two left are still missing
		t.Errorf("second run: %+v", r)
	}
	st2, _ := os.Stat(live)
	if readFile(t, live) != before || !st2.ModTime().Equal(st.ModTime()) {
		t.Error("the second run rewrote a file")
	}
	clean := filepath.Join(root, "carves", "clean.md")
	os.WriteFile(clean, []byte("[t](../tool/x.md) [e](https://x) [l](#a)\n"), 0o644)
	r, err = RepairLinks(root, []string{"carves/clean.md"})
	if err != nil || len(r.Considered) != 0 || len(r.Written) != 0 {
		t.Errorf("a clean document: %+v %v", r, err)
	}
}

// R500 — a pointer is repaired by the same rule, the key kept.
func TestAPointerIsRepairedByTheSameRule(t *testing.T) {
	root := t.TempDir()
	for name, body := range map[string]string{
		"carves/done/moved.md": "m\n",
		"DONE.md":              "# Done\n\n---\n\n- **2026-09-01 — #5: x.** Part `carves/moved.md#1`. and `carves/nowhere.md#2`\n",
	} {
		p := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, root, "carves")
	r, err := RepairLinks(root, []string{"DONE.md"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Counts[Rewritten] != 1 || r.Counts[Unresolvable] != 1 {
		t.Errorf("counts: %v (%+v)", r.Counts, r.Considered)
	}
	got := readFile(t, filepath.Join(root, "DONE.md"))
	if !strings.Contains(got, "Part `carves/done/moved.md#1`.") || !strings.Contains(got, "`carves/nowhere.md#2`") {
		t.Errorf("ledger after repair:\n%s", got)
	}
}
