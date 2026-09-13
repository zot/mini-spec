// CRC: crc-Pending.md | R332, R333, R334
package pending

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/parser"
)

// R332, R333 — a queue verb in a project with no layer refuses by type, naming the files;
// creation writes the missing ones with their preambles and never touches an existing
// one; after it the verb proceeds.
func TestAMissingLayerIsRefusedByNameAndCreatedWhole(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"carves", ".minispec"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "carves", "x.md"), []byte(carveSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "DONE.md"), []byte("# Done\n\nkept as it was\n\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec")
	var m *MissingTrajectoryError
	if !errors.As(err, &m) || strings.Join(m.Files, ",") != "PENDING.md,CURRENT.md" {
		t.Fatalf("add-item with no layer = %v; want a refusal naming PENDING.md and CURRENT.md", err)
	}
	if _, err := Start(root, 1, "c"); !errors.As(err, &m) {
		t.Errorf("start with no layer = %v; want the typed refusal", err)
	}
	made, err := CreateTrajectory(root)
	if err != nil || strings.Join(made, ",") != "PENDING.md,CURRENT.md" {
		t.Fatalf("created %v (%v); want the two missing files and not DONE.md", made, err)
	}
	if d := read(t, root, "DONE.md"); !strings.Contains(d, "kept as it was") {
		t.Error("an existing file was overwritten by creation")
	}
	for _, f := range []string{"PENDING.md", "CURRENT.md"} {
		if body := read(t, root, f); !strings.HasPrefix(body, "# ") || !strings.Contains(body, "\n---\n") {
			t.Errorf("%s was not created with its preamble and rule:\n%s", f, body)
		}
	}
	if !strings.Contains(read(t, root, "CURRENT.md"), "## Active\n\n_No active item._") {
		t.Error("the current file has no Active region — the one region a tool may clear")
	}
	again, err := CreateTrajectory(root)
	if err != nil || len(again) != 0 {
		t.Errorf("a second creation wrote %v; want nothing", again)
	}
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Errorf("add-item after creation: %v", err)
	}
}

// R334 — the scaffold is a carve the reader reads: one open part, refusing an existing
// file and a name that is a path.
func TestInitCarveScaffoldsAReadableCarve(t *testing.T) {
	root := t.TempDir()
	rel, err := CreateCarve(root, "review-console")
	if err != nil || rel != "carves/review-console.md" {
		t.Fatalf("CreateCarve: %s %v", rel, err)
	}
	c, err := parser.ReadCarve(filepath.Join(root, rel), rel)
	if err != nil {
		t.Fatal(err)
	}
	if !c.HasStatus || c.Open() != 1 || len(c.Stateless) != 0 || c.NonConforming() != 0 {
		t.Errorf("the scaffold does not read as one open conforming part: status=%v open=%d stateless=%d nonconforming=%d", c.HasStatus, c.Open(), len(c.Stateless), c.NonConforming())
	}
	if _, err := CreateCarve(root, "review-console"); err == nil {
		t.Error("an existing carve was overwritten")
	}
	for _, bad := range []string{"", "a/b", "x.md", "two words"} {
		if _, err := CreateCarve(root, bad); err == nil {
			t.Errorf("%q was accepted as a carve name", bad)
		}
	}
}
