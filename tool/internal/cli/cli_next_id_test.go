// CRC: crc-CLI.md | R191, R198
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// R198. The wiring property, and the one a parser test cannot reach: `next-id item`
// must answer in a tree with **no design root at all**.
//
// This is a regression test for a defect found by running the command rather than by
// reading it. runQuery resolved a design root before dispatching, so the subcommand
// whose whole point is repository scope failed in the layout it exists for.
//
// Fire alarm: delete the early-dispatch branch at the top of runQuery and this goes red
// with "no design/ directory found". Pulled 2026-08-14 — rang.
func TestNextIDItemNeedsNoDesignRoot(t *testing.T) {
	dir := t.TempDir()
	// A repository root marker, the queue, and deliberately no design/ anywhere.
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"PENDING.md": "# Pending\n\n## 4. **live**. Active.\n",
		"DONE.md":    "# Done\n\n- **2026-08-14 — #7: a thing.** (`abc1234`) Part `carves/x.md#2`.\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// os.Chdir rather than t.Chdir: the latter needs go1.24 and this module declares
	// go1.21, so it fails `go vet` and locks anyone on an older toolchain out of the
	// tests. The restore is what t.Chdir would do anyway, and this test must never be
	// parallel — which is the hazard O5 named when it called chdir testing unsafe.
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(prev) })

	c := &CLI{}
	if code := c.runQuery([]string{"next-id", "item"}); code != 0 {
		t.Fatalf("runQuery exited %d in a tree with no design root; want 0", code)
	}
}
