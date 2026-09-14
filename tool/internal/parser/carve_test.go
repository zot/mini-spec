// CRC: crc-Carve.md | R209, R210, R211, R214, R216, R219, R220
package parser

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/minispecsdom"
)

// Fixtures carry real shapes wherever one exists; the failures these guard were measured
// in the corpus before any of this was written.

// R209 — a document about the format quotes a status block in a fence; a line scan counted
// it (ark's carves/README.md, 14 open reported where 13 existed).
func TestFencedStatusExampleIsNotData(t *testing.T) {
	src := "# Carves\n\n**Status: a `## Status` block near the top.**\n\n" +
		"```markdown\n## Status\n\n" +
		"- [x] ~~**Item 1 — record and resolve.**~~ **LANDED (4c6e974, 2026-08-04.)**\n" +
		"- [ ] **Item 4 — fail fast when onboarding does not take.** **OPEN (#122.)**\n```\n"
	c := parseCarve("carves/README.md", src)
	if c.HasStatus || len(c.Parts) != 0 {
		t.Errorf("HasStatus=%v parts=%d; want false, 0: the only status block is fenced", c.HasStatus, len(c.Parts))
	}
}

const fixture = "# Carve: x\n\n## Status\n\n" +
	"- [ ] **Item 3 — a top-level part.** **OPEN (not queued.)**\n" +
	"- **Item 8 — shrink the skill.** **SPLIT (Bill, 2026-08-14.)** No checkbox.\n" +
	"  - [x] ~~**8.1 — move the mechanics out.**~~ **LANDED (`9dfe5f8`, 2026-08-14 — `#9`.)**\n" +
	"  - [ ] **8.2 — `minispec query carves`, the census.** **OPEN (#13.)**\n" +
	"\n## Open questions\n\n- [ ] one\n- [x] two\n"

// R209, R210, R216 — subparts count, the block bounds the count, a SPLIT parent is stateless.
func TestSubpartsCountAndTheStatusBlockBoundsTheCount(t *testing.T) {
	c := parseCarve("carves/x.md", fixture)
	if c.Open() != 2 || c.Landed() != 1 || len(c.Stateless) != 1 {
		t.Errorf("got %d open, %d landed, %d stateless; want 2, 1, 1", c.Open(), c.Landed(), len(c.Stateless))
	}
	keys := map[string]bool{}
	for _, p := range c.Parts {
		keys[p.Key()] = true
	}
	for _, want := range []string{"3", "8.1", "8.2"} {
		if !keys[want] {
			t.Errorf("part %q missing from %v", want, keys)
		}
	}
}

// R216 — the stateless line's number is derived from its offset; a wrong derivation points
// a reader at the wrong line.
func TestAStatelessLineCarriesItsLineAndReason(t *testing.T) {
	c := parseCarve("carves/x.md", fixture)
	if len(c.Stateless) != 1 {
		t.Fatalf("want 1 stateless line, got %d", len(c.Stateless))
	}
	s := c.Stateless[0]
	if s.Line != 6 || s.Reason != "no checkbox" || !strings.Contains(s.Text, "8") {
		t.Errorf("stateless = L%d %q %q; want L6, no checkbox, the Item 8 line", s.Line, s.Reason, s.Text)
	}
}

// The census wants the whole title, code spans included, and the queue ID from the OPEN
// marker alone.
func TestTitleAndQueueIDAreDerivedFromTheLine(t *testing.T) {
	c := parseCarve("carves/x.md", fixture)
	var sub Part
	for _, p := range c.Parts {
		if p.Key() == "8.2" {
			sub = p
		}
	}
	if got := sub.Title(); got != "`minispec query carves`, the census" {
		t.Errorf("Title = %q; want the whole head including its code span", got)
	}
	if sub.QueueID() != 13 || c.Parts[0].QueueID() != 0 {
		t.Errorf("QueueID: 8.2=%d (want 13), Item 3=%d (want 0)", sub.QueueID(), c.Parts[0].QueueID())
	}
}

func writeCarve(t *testing.T, root, rel, body string) string {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// R211, R208 — a document with no status block stays in the scan; carves/done/ is not entered.
func TestADocumentWithNoStatusBlockIsReportedNotDropped(t *testing.T) {
	root := t.TempDir()
	writeCarve(t, root, "carves/a.md", fixture)
	writeCarve(t, root, "carves/notes.md", "# Notes\n\nprose only\n")
	writeCarve(t, root, "carves/done/old.md", fixture)
	scan, err := ScanCarves(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Carves) != 2 || scan.NoStatus() != 1 || scan.WithStatus() != 1 {
		t.Errorf("carves=%d noStatus=%d withStatus=%d; want 2, 1, 1", len(scan.Carves), scan.NoStatus(), scan.WithStatus())
	}
}

// R214 — no carve directory at all has no answer.
func TestNoCarveDirectoryHasNoAnswer(t *testing.T) {
	_, err := ScanCarves(t.TempDir())
	if !errors.Is(err, ErrNoCarveDirs) {
		t.Errorf("got %v; want ErrNoCarveDirs", err)
	}
}

// R220 — the write is by rename, and none of the reader's refusals reaches the file.
func TestAMarkerWriteIsAtomicAndARefusalLeavesTheFileByteIdentical(t *testing.T) {
	root := t.TempDir()
	path := writeCarve(t, root, "carves/x.md", fixture)
	if err := SetMarker(path, "3", "OPEN", "#40."); err != nil {
		t.Fatalf("SetMarker: %v", err)
	}
	after, _ := os.ReadFile(path)
	if !strings.Contains(string(after), "**Item 3 — a top-level part.** **OPEN (#40.)**") {
		t.Errorf("marker not written:\n%s", after)
	}
	if strings.Count(string(after), "\n") != strings.Count(fixture, "\n") {
		t.Errorf("the write changed more than the marker")
	}
	landed, err := PartIsLanded(path, "8.1")
	if err != nil || !landed {
		t.Errorf("PartIsLanded(8.1) = %v, %v; want true", landed, err)
	}
	before, _ := os.ReadFile(path)
	for _, tc := range []struct {
		name string
		err  error
		do   func() error
	}{
		{"OPEN over a landed part", minispecsdom.ErrReopen, func() error { return SetMarker(path, "8.1", "OPEN", "#41.") }},
		{"Land over a landed part", minispecsdom.ErrLanded, func() error { return SetPartLanded(path, "8.1", "`abc`, 2026-09-04 — `#41`.") }},
		{"a key no part carries", minispecsdom.ErrNoPart, func() error { return SetMarker(path, "99", "OPEN", "#41.") }},
	} {
		err := tc.do()
		if !errors.Is(err, tc.err) {
			t.Errorf("%s: got %v; want %v", tc.name, err, tc.err)
		}
		now, _ := os.ReadFile(path)
		if string(now) != string(before) {
			t.Errorf("%s: the file changed on a refusal", tc.name)
		}
	}
	if entries, _ := os.ReadDir(filepath.Dir(path)); len(entries) != 1 {
		t.Errorf("temp files left behind: %d entries in the directory", len(entries))
	}
}

// R219 — this repository's own carves, through the reader whose rules they must obey.
func TestThisRepositorysCarvesStayConformant(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	scan, err := ScanCarves(root)
	if err != nil {
		t.Skipf("no carves at %s: %v", root, err)
	}
	for _, c := range scan.Carves {
		for _, p := range c.Parts {
			if !p.Conforms() {
				t.Errorf("%s L%d %s: %v", c.Path, p.Line(), p.Key(), p.Deviations())
			}
		}
		for _, s := range c.Stateless {
			if len(s.Deviations) > 0 {
				t.Errorf("%s L%d (stateless): %v", c.Path, s.Line, s.Deviations)
			}
		}
	}
}

// R220. A write the dependency cannot read back panics with a ReadBackError; editFile turns
// that into a refusal naming the file and leaves the file untouched. Any other panic is
// still a panic — the recovery is for the one invariant the dependency chose to panic on.
func TestAReadBackPanicBecomesARefusalAndTheFileIsUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.md")
	if err := os.WriteFile(path, []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := editFile(path, func(string) (string, error) {
		panic(&minispecsdom.ReadBackError{Reader: "Carve", Write: "SetMarker", Key: "3", Want: "OPEN", Got: "nothing"})
	})
	if err == nil {
		t.Fatal("a read-back panic was not turned into a refusal")
	}
	if !strings.Contains(err.Error(), "x.md") || !strings.Contains(err.Error(), "did not read back") {
		t.Errorf("the refusal does not name the file and the cause: %v", err)
	}
	if got, _ := os.ReadFile(path); string(got) != "before\n" {
		t.Errorf("the file was written despite the refusal: %q", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %d entries", len(entries))
	}
}
