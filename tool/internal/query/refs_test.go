// CRC: crc-Refs.md | R495, R496, R497
package query

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func refsRoot(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	root := t.TempDir()
	files := map[string]string{
		"carves/x.md":        "# x\n\n## Status\n\n- [x] ~~**Item 1 — a.**~~ **LANDED (2026-09-01 — `#5`.)**\n\nsee [o](other.md) and `other.md#1`\n",
		"carves/other.md":    "# o\n",
		".carves/private.md": "private [x](../carves/x.md) and `../carves/x.md#1`\n",
		"DONE.md":            "# Done\n\n---\n\n- **2026-09-01 — #5: a.** Part `carves/x.md#1`.\n",
		"PENDING.md":         "# Pending\n\n---\n",
		".scratch/notes.md":  "[x](../carves/x.md)\n",
		".gitignore":         "DONE.md\nPENDING.md\n.carves/\n.scratch/\n",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "carves", ".gitignore"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

// R497 — the owned documents: tracked plus sited, scratch excluded.
func TestOwnedDocumentsAreTrackedPlusSited(t *testing.T) {
	root := refsRoot(t)
	files, err := OwnedDocuments(root)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{".carves/private.md": true, "DONE.md": true, "PENDING.md": true, "carves/other.md": true, "carves/x.md": true}
	if len(files) != len(want) {
		t.Fatalf("population: %v", files)
	}
	for _, f := range files {
		if !want[f] {
			t.Errorf("unexpected %s in the population", f)
		}
	}
}

// R495, R496 — every kind, resolved by its own rule; --to inverts.
func TestRefsResolveAndInvert(t *testing.T) {
	root := refsRoot(t)
	refs, err := Refs(root, nil, "carves/x.md")
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range refs {
		got[r.File+" "+r.Text] = r.Kind
	}
	want := map[string]string{
		".carves/private.md [x](../carves/x.md)": "link",
		".carves/private.md `../carves/x.md#1`":  "pointer",
		"DONE.md `carves/x.md#1`":                "pointer",
	}
	if len(got) != len(want) {
		t.Fatalf("refs to carves/x.md: %+v", refs)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %q, want %q", k, got[k], v)
		}
	}
	all, _ := Refs(root, []string{"carves/x.md"}, "")
	if len(all) != 2 || all[0].Kind != "link" || all[0].Resolved != "carves/other.md" || all[1].Kind != "pointer" || all[1].Resolved != "carves/other.md" {
		t.Errorf("refs in carves/x.md: %+v", all)
	}
}
