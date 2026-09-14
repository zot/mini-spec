package minispecsdom

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func currentFixture(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("testdata/current-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// outside returns the fixture's bytes on either side of the Active region's body:
// everything through the `## Active` line, and everything from the next heading on.
func outside(src string) (head, tail string) {
	i := strings.Index(src, "## Active\n") + len("## Active\n")
	j := strings.Index(src, "## How this project works")
	return src[:i], src[j:]
}

// CRC: crc-Current.md | Seq: seq-current.md#1 | R385, R386, R387
func TestTheCurrentFixturesRegionsReadBack(t *testing.T) {
	src := currentFixture(t)
	c, err := ParseCurrent(src)
	if err != nil {
		t.Fatal(err)
	}
	if r, _ := c.Render(); r != src {
		t.Fatal("render differs")
	}
	a := c.Active()
	if !strings.HasPrefix(a, "`#21` — The done schema.") || !strings.Contains(a, "### A sub-heading inside the item") || !strings.HasSuffix(a, "Still the item's.") {
		t.Errorf("Active:\n%s", a)
	}
	if !c.Occupied() {
		t.Errorf("not occupied")
	}
	if s := strings.Join(c.Standing(), "|"); s != "How this project works|Tool wrinkles" {
		t.Errorf("standing %q", s)
	}
}

// CRC: crc-Current.md | Seq: seq-current.md#2 | R388, R389, R390
func TestAWriteReachesTheRegionAndNothingElse(t *testing.T) {
	src := currentFixture(t)
	head, tail := outside(src)
	c, err := ParseCurrent(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.SetActive("x"); !errors.Is(err, ErrOccupied) {
		t.Fatalf("SetActive over a held item: %v, want ErrOccupied", err)
	}
	if r, _ := c.Render(); r != src {
		t.Fatalf("a refused SetActive changed the document")
	}
	if err := c.Reset(); err != nil {
		t.Fatal(err)
	}
	r, _ := c.Render()
	if r != head+"\n"+Placeholder+"\n\n"+tail {
		t.Errorf("after Reset:\n%s", r)
	}
	if c.Occupied() || c.Active() != "" {
		t.Errorf("occupied after Reset")
	}
	if err := c.SetActive("`#22` — new item\n\nits context"); err != nil {
		t.Fatal(err)
	}
	r, _ = c.Render()
	if r != head+"\n`#22` — new item\n\nits context\n\n"+tail {
		t.Errorf("after SetActive:\n%s", r)
	}
	if !c.Occupied() {
		t.Errorf("not occupied after SetActive")
	}
	// A region at the end of the file: no trailing blank line is added.
	end, err := ParseCurrent("# Current\n\n---\n\n## Active\n\n_No active item._\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := end.SetActive("last"); err != nil {
		t.Fatal(err)
	}
	if r, _ := end.Render(); r != "# Current\n\n---\n\n## Active\n\nlast\n" {
		t.Errorf("region at end:\n%q", r)
	}
}

// CRC: crc-Current.md | Seq: seq-current.md#1.2 | R385, R406
func TestExactlyOneActive(t *testing.T) {
	for src, want := range map[string]error{
		"# Current\n\n---\n\n## Standing\n\ntext\n":              ErrNoActive,
		"# Current\n\n---\n\n## Active\n\na\n\n## Active\n\nb\n": ErrManyActive,
		"# Current\n\n```\n## Active\n```\n":                     ErrNoActive,
	} {
		_, err := ParseCurrent(src)
		if err == nil {
			t.Errorf("%q: parsed", src)
			continue
		}
		if !errors.Is(err, want) { // R406
			t.Errorf("%q: %v, want %v", src, err, want)
		}
	}
}

// CRC: crc-Current.md | R405
func TestAGroupOpenAtEndOfInputIsUnreadInCurrent(t *testing.T) {
	src := "# C\n\n## Active\n\n_No active item._\n\n## Notes\n\n`x\n"
	c, err := ParseCurrent(src)
	if err != nil {
		t.Fatal(err)
	}
	if c.Occupied() {
		t.Errorf("active %q, want the region left empty", c.Active())
	}
	u := c.Unread()
	if len(u) != 1 || u[0].Line != 9 || !strings.Contains(u[0].Text, "never closed") {
		t.Errorf("unread %+v, want the span at line 9", u)
	}
}
