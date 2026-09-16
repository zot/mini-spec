// CRC: crc-FinishedCarve.md | R469, R470, R471, R490, R473, R474
package update

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const landedStatus = "# Carve: x\n\n## Status\n\n- [x] ~~**Item 1 — done.**~~ **LANDED (`abc`, 2026-09-01 — `#5`.)**\n\n"

// finishRoot lays out a finished carve with links in every direction, and a twin at the
// destination directory that must never be touched.
func finishRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"tool/a.md":              "a\n",
		"carves/x.md":            landedStatus + "see [a](../tool/a.md#s), [o](done/old.md), [t](other.md), [n](nowhere.md), [e](https://x) and [l](#top)\n",
		"carves/other.md":        landedStatus + "back to [x](x.md#4) and ` [x](x.md) `\n",
		"carves/done/old.md":     landedStatus + "up to [x](../x.md)\n",
		"carves/done/x-twin.md":  "twin\n",
		"carves/done/nowhere.md": "a file the carve never pointed at\n",
		"PENDING.md":             "# Pending\n\n## 5. **x**. Source: [carves/x.md](carves/x.md), part `#1`.\n",
		"specs/index.md":         "see the [carve](../carves/x.md#4).\n",
		".carves/private.md":     "note [x](../carves/x.md) and `../carves/x.md#2`\n",
		"DONE.md":                "# Done\n\n---\n\n- **2026-09-01 — #5: x.** Part `carves/x.md#1`.\n",
		".gitignore":             "PENDING.md\nDONE.md\n.carves/\n",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, root, "carves", "tool", "specs", ".gitignore")
	return root
}

func treeHash(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			b, _ := os.ReadFile(p)
			rel, _ := filepath.Rel(root, p)
			out[rel] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		return nil
	})
	return out
}

// R469, R471, R490, R474 — the move, and every link follows in both directions.
func TestAFinishedCarveMovesAndEveryLinkFollows(t *testing.T) {
	root := finishRoot(t)
	before := treeHash(t, root)
	r, err := FinishCarve(root, "carves/x.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "carves", "x.md")); !errors.Is(err, fs.ErrNotExist) {
		t.Error("the old file is still there")
	}
	want := map[string]string{
		"carves/done/x.md":   landedStatus + "see [a](../../tool/a.md#s), [o](old.md), [t](../other.md), [n](nowhere.md), [e](https://x) and [l](#top)\n",
		"carves/other.md":    landedStatus + "back to [x](done/x.md#4) and ` [x](x.md) `\n",
		"carves/done/old.md": landedStatus + "up to [x](x.md)\n",
		"PENDING.md":         "# Pending\n\n## 5. **x**. Source: [carves/x.md](carves/done/x.md), part `#1`.\n", // sited: rewritten
		".carves/private.md": "note [x](../carves/done/x.md) and `../carves/done/x.md#2`\n",
		"DONE.md":            "# Done\n\n---\n\n- **2026-09-01 — #5: x.** Part `carves/done/x.md#1`.\n",
		"specs/index.md":     "see the [carve](../carves/done/x.md#4).\n",
	}
	for rel, w := range want {
		got, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("%s: %v", rel, err)
		}
		if string(got) != w {
			t.Errorf("%s:\n got %q\nwant %q", rel, got, w)
		}
	}
	after := treeHash(t, root)
	if after["carves/done/x-twin.md"] != before["carves/done/x-twin.md"] || after["tool/a.md"] != before["tool/a.md"] {
		t.Error("a file the move should never touch changed")
	}
	if r.Counts[Rewritten] != 10 || r.Counts[Left] != 1 || r.From != "carves/x.md" || r.To != "carves/done/x.md" {
		t.Errorf("report: %+v", r)
	}
}

// R469, R470, R473 — every refusal leaves the whole tree byte for byte as it was.
func TestTheRefusalsLeaveEveryFileAsItWas(t *testing.T) {
	root := finishRoot(t)
	os.WriteFile(filepath.Join(root, "carves", "open.md"), []byte("# c\n\n## Status\n\n- [ ] **Item 1 — still open.** **OPEN (not queued.)**\n"), 0o644)
	os.WriteFile(filepath.Join(root, "carves", "nostatus.md"), []byte("# c\n\nprose only\n"), 0o644)
	os.WriteFile(filepath.Join(root, "carves", "x-twin.md"), []byte(landedStatus+"would collide\n"), 0o644) // finished, so only the destination refuses it
	os.WriteFile(filepath.Join(root, "elsewhere.md"), []byte(landedStatus), 0o644)
	gitInit(t, root, "carves", "elsewhere.md")
	before := treeHash(t, root)
	cases := map[string]string{
		"carves/open.md":     "open part",
		"carves/nostatus.md": "no status block",
		"carves/done/old.md": "not directly in a carve directory",
		"elsewhere.md":       "not directly in a carve directory",
		"carves/x-twin.md":   "already at the destination",
	}
	for carve, reason := range cases {
		r, err := FinishCarve(root, carve)
		if err == nil || r != nil || !strings.Contains(err.Error(), reason) {
			t.Errorf("%s: want a refusal naming %q, got %v, %+v", carve, reason, err, r)
		}
	}
	var ope *OpenPartsError
	if _, err := FinishCarve(root, "carves/open.md"); !errors.As(err, &ope) || len(ope.Parts) != 1 || ope.Parts[0] != "1" {
		t.Errorf("the open-part refusal does not name the part: %v", err)
	}
	after := treeHash(t, root)
	if len(after) != len(before) {
		t.Fatalf("a refusal created or removed a file: %d files before, %d after", len(before), len(after))
	}
	for rel, h := range before {
		if after[rel] != h {
			t.Errorf("%s changed under a refusal", rel)
		}
	}
}
