// CRC: crc-CLI.md | R495, R496
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/query"
)

// R495, R496 — the inventory, and --to inverted, needing no design root.
func TestQueryRefsListsAndInverts(t *testing.T) {
	dir := t.TempDir()
	for name, body := range map[string]string{
		"carves/x.md":     "# x\n\nsee [o](other.md) and `other.md#1`\n",
		"carves/other.md": "# o\n",
		"DONE.md":         "# Done\n\n---\n\n- **2026-09-01 — #5: x.** Part `carves/x.md#1`.\n",
		".gitignore":      "DONE.md\n",
	} {
		os.MkdirAll(filepath.Join(dir, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitInit(t, dir, "carves", ".gitignore")
	prev, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	c := &CLI{}
	if code := c.runQuery([]string{"refs", "--to", "carves/x.md"}); code != 0 {
		t.Errorf("exited %d", code)
	}
	refs, err := query.Refs(dir, nil, "carves/x.md")
	if err != nil {
		t.Fatal(err)
	}
	out := query.FormatRefs(refs)
	if !strings.Contains(out, "DONE.md:5  `carves/x.md#1`  pointer  carves/x.md") || !strings.HasSuffix(out, "1 references\n") {
		t.Errorf("inventory:\n%s", out)
	}
	all, _ := query.Refs(dir, []string{"carves/x.md"}, "")
	if len(all) != 2 {
		t.Errorf("refs in carves/x.md: %+v", all)
	}
}
