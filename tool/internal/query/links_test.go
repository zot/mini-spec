// CRC: crc-Links.md | R455, R456, R457, R459, R460
package query

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// linkRepo builds a repository holding one link of every class, cited from carves/doc.md.
func linkRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}
	parent := t.TempDir()
	root := filepath.Join(parent, "repo")
	files := map[string]string{
		"tracked.md":    "t\n",
		"dir/inside.md": "d\n",
		"untracked.md":  "u\n",
		"private/x.md":  "p\n",
		".gitignore":    "private/\n",
		"../escape.md":  "e\n",
		"carves/doc.md": "[t](../tracked.md) [d](../dir) [u](../untracked.md) [p](../private/x.md#frag)\n" +
			"[m](../missing.md) [e](../../escape.md) [a](/abs.md) [x](https://x.example/y) [l](#frag)\n" +
			"and ` [fenced](../nope.md) ` is an example\n",
	}
	for name, body := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"init", "-q"}, {"add", "tracked.md", "dir", ".gitignore", "carves"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return root
}

// R456, R457, R459 — every class, once each; the fenced link never read.
func TestEveryClassInATemporaryRepository(t *testing.T) {
	root := linkRepo(t)
	r, err := CheckLinks(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]LinkClass{
		"[t](../tracked.md)": ClassTracked, "[d](../dir)": ClassTracked, "[u](../untracked.md)": ClassUntracked,
		"[p](../private/x.md#frag)": ClassIgnored, "[m](../missing.md)": ClassMissing,
		"[e](../../escape.md)": ClassOutside, "[a](/abs.md)": ClassOutside,
		"[x](https://x.example/y)": ClassExternal, "[l](#frag)": ClassLocal,
	}
	if len(r.Links) != len(want) {
		t.Fatalf("want %d links, got %d: %+v", len(want), len(r.Links), r.Links)
	}
	for _, l := range r.Links {
		if want[l.Link] != l.Class {
			t.Errorf("%s: got %s, want %s (resolved %q)", l.Link, l.Class, want[l.Link], l.Resolved)
		}
		if l.File != "carves/doc.md" || l.Line < 1 || l.Line > 2 {
			t.Errorf("%s: cited at %s:%d", l.Link, l.File, l.Line)
		}
	}
	if !r.Errors() {
		t.Error("a report with ignored, missing and outside links claims no error")
	}
	if len(r.Counts) != len(LinkClasses) || r.Counts[ClassTracked] != 2 || r.Counts[ClassOutside] != 2 {
		t.Errorf("counts: %v", r.Counts)
	}
	wantError := map[LinkClass]bool{ClassIgnored: true, ClassMissing: true, ClassOutside: true}
	for _, c := range LinkClasses {
		if c.IsError() != wantError[c] {
			t.Errorf("%s: IsError is %v, want %v", c, c.IsError(), wantError[c])
		}
		if c.IsWarning() != (c == ClassUntracked) {
			t.Errorf("%s: IsWarning is %v, want %v", c, c.IsWarning(), c == ClassUntracked)
		}
	}
}

// R460 — no git is a refusal, never an empty report.
func TestNoGitIsARefusal(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "carves"), 0o755)
	os.WriteFile(filepath.Join(dir, "carves", "a.md"), []byte("[x](a.md)\n"), 0o644)
	r, err := CheckLinks(dir, nil)
	if !errors.Is(err, ErrNoGitTree) || r != nil {
		t.Fatalf("want ErrNoGitTree and no report, got %v, %+v", err, r)
	}
}
