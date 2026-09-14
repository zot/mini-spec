package minispecsdom

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func doneFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/done-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// CRC: crc-Done.md | Seq: seq-done.md#1 | R378, R396, R380, R381, R382, R384
func TestTheDoneFixturesEntriesReadBack(t *testing.T) {
	src := doneFixture(t)
	d := ParseDone(src)
	if r, _ := d.Render(); r != src {
		t.Fatal("render differs from source")
	}
	es := d.Entries()
	if len(es) != 3 {
		t.Fatalf("%d entries, want 3", len(es))
	}
	if !slices.Equal(es[0].IDs, []int{8}) || len(es[1].IDs) != 0 || len(es[2].IDs) != 0 {
		t.Errorf("ids %v %v %v", es[0].IDs, es[1].IDs, es[2].IDs)
	}
	if !es[0].HasSlot || !es[1].HasSlot || es[2].HasSlot {
		t.Errorf("slots %v %v %v", es[0].HasSlot, es[1].HasSlot, es[2].HasSlot)
	}
	if es[0].Date != "2026-08-30" || es[0].Title != "the node protocol." || es[0].Commit != "77fa5f4" ||
		es[0].PartDoc != "carves/x.md" || es[0].PartKey != "1" {
		t.Errorf("entry 0: %+v", *es[0])
	}
	if es[1].PartDoc != "carves/x.md" || es[1].PartKey != "3" || es[1].Commit != "abc1234" {
		t.Errorf("entry 1 (pointer from the body): %+v", *es[1])
	}
	if u := d.Unread(); d.MaxID() != 8 || len(u) != 1 || u[0] != (Unread{21, "- not an entry, but entry-like"}) {
		t.Errorf("MaxID %d Unread %+v", d.MaxID(), u)
	}
	// R395: every entry knows its line, 1-based.
	if es[0].Line() != 7 || es[1].Line() != 11 || es[2].Line() != 14 {
		t.Errorf("lines %d %d %d", es[0].Line(), es[1].Line(), es[2].Line())
	}
	// An indented bullet inside a body is the entry's, not a boundary.
	var b strings.Builder
	for _, n := range es[0].run {
		s, _ := n.Render()
		b.WriteString(s)
	}
	if !strings.Contains(b.String(), "- O9, an indented bullet") {
		t.Errorf("the body's indented bullet ended entry 0's region")
	}
}

// CRC: crc-Done.md | Seq: seq-done.md#2 | R383
func TestPrependLandsAfterTheRule(t *testing.T) {
	d := ParseDone(doneFixture(t))
	if err := d.Prepend("- **2026-09-03 — #21: the done schema.** (`deadbee`) Part `carves/t.md#5`.", "One line of body."); err != nil {
		t.Fatal(err)
	}
	r, _ := d.Render()
	if !strings.Contains(r, "---\n\n- **2026-09-03 — #21: the done schema.** (`deadbee`) Part `carves/t.md#5`.\nOne line of body.\n\n- **2026-08-30") {
		t.Errorf("after Prepend:\n%s", r)
	}
	if es := d.Entries(); len(es) != 4 || !slices.Equal(es[0].IDs, []int{21}) || d.MaxID() != 21 {
		t.Errorf("entries after Prepend: %d, max %d", len(es), d.MaxID())
	}
	empty := ParseDone("# Done\n\nLedger.\n\n---\n")
	if err := empty.Prepend("- **2026-09-03 — #1: first.**", ""); err != nil {
		t.Fatal(err)
	}
	if r, _ := empty.Render(); r != "# Done\n\nLedger.\n\n---\n\n- **2026-09-03 — #1: first.**\n\n" {
		t.Errorf("empty ledger after Prepend:\n%q", r)
	}
}

// CRC: crc-Done.md | R402
func TestAGroupOpenAtEndOfInputIsUnreadInDone(t *testing.T) {
	src := "# Done\n\n---\n\n- **2026-09-05 — #1: T.** (`abc`)\nbody\n```\nlost tail\n- **2026-09-04 — #2: U.**\n"
	d := ParseDone(src)
	// R445: the fence is demoted at end of input, so the entry after it reads too.
	if len(d.Entries()) != 2 {
		t.Errorf("%d entries, want both — the fence was text", len(d.Entries()))
	}
	u := d.Unread()
	if len(u) != 1 || u[0].Line != 7 || !strings.Contains(u[0].Text, "never closed, read as text") || !strings.Contains(u[0].Text, "```") {
		t.Errorf("unread %+v, want the fence at line 7, read as text", u)
	}
}
