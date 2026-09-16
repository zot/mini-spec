// CRC: crc-Markdown.md | R448, R449, R450, R451, R452, R453
package minispecsdom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/simple-dom/sdom"
)

const linksFixture = "# Doc\n\n" +
	"Prose [one](a.md) here and ![pic](img.png).\n" +
	"A span ` [text](path) ` is an example.\n\n" +
	"```\n## Not a heading\n[fenced](b.md)\n```\n"

// R448, R449 — links read outside code groups only, with their positions.
func TestLinksAreReadOutsideCodeGroupsOnly(t *testing.T) {
	m := ParseMarkdown(linksFixture)
	links := m.Links()
	if len(links) != 2 {
		t.Fatalf("want 2 links, got %d: %+v", len(links), links)
	}
	one, pic := links[0], links[1]
	if one.Text != "one" || one.Dest != "a.md" || one.Image || one.Line() != 3 || one.Raw != "[one](a.md)" {
		t.Errorf("first link: %+v", one)
	}
	if pic.Text != "pic" || pic.Dest != "img.png" || !pic.Image || pic.Raw != "![pic](img.png)" {
		t.Errorf("image link: %+v", pic)
	}
	if linksFixture[one.Offset()] != '[' || linksFixture[pic.Offset()] != '!' {
		t.Errorf("offsets do not land on the openers: %d %d", one.Offset(), pic.Offset())
	}
	for _, l := range links {
		if l.Dest == "path" || l.Dest == "b.md" {
			t.Errorf("a link inside a code group was read: %+v", l)
		}
	}
	if out, _ := m.Render(); out != linksFixture {
		t.Error("render is not byte-exact")
	}
}

// R450 — only the inline form; malformed brackets are text; a checkbox beside a link is
// passed over.
func TestOnlyTheInlineFormIsALink(t *testing.T) {
	src := "[text][ref] and <https://x> and https://y\n[unclosed and [text](unclosed\n- [ ] [real](a.md)\n"
	links := ParseMarkdown(src).Links()
	if len(links) != 1 || links[0].Text != "real" || links[0].Dest != "a.md" {
		t.Fatalf("want exactly the real link, got %+v", links)
	}
}

// R451 — the destination forms.
func TestDestinationForms(t *testing.T) {
	cases := []struct{ raw, dest, path, frag string }{
		{"<a b.md>", "a b.md", "a b.md", ""},
		{`a.md "title"`, "a.md", "a.md", ""},
		{"a.md#frag", "a.md#frag", "a.md", "frag"},
		{"#frag", "#frag", "", "frag"},
		{"a.md#x#y", "a.md#x#y", "a.md", "x#y"},
	}
	for _, c := range cases {
		l := splitDest(c.raw)
		if l.Dest != c.dest || l.Path != c.path || l.Fragment != c.frag {
			t.Errorf("%q: got dest %q path %q frag %q", c.raw, l.Dest, l.Path, l.Fragment)
		}
	}
}

// corpus is every markdown document this tool is trusted with: the repository root, carves
// recursively, specs and design, and a sibling ark checkout when one is present. Not
// fixtures — a fixture holds only what its author thought to include.
func corpus(t *testing.T) map[string]string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	add := func(pat string) {
		m, err := filepath.Glob(filepath.Join(root, pat))
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, m...)
	}
	walk := func(dir string) {
		filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && (d.Name() == ".git" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			if !d.IsDir() && strings.HasSuffix(p, ".md") {
				paths = append(paths, p)
			}
			return nil
		})
	}
	add("*.md")
	add("tool/design/*.md")
	walk(filepath.Join(root, "carves"))
	walk(filepath.Join(root, "tool", "specs"))
	if ark := filepath.Join(root, "..", "ark"); isDir(ark) {
		walk(ark)
	} else {
		t.Logf("no sibling ark checkout at %s; the corpus is this repository alone", ark)
	}
	out := make(map[string]string, len(paths))
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		out[p] = string(b)
	}
	return out
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// R452, R453 — byte-exact over the real corpus, with the count reported. The round trip
// alone cannot tell a render that rebuilds from the nodes from one that hands back its
// source (the R216 lesson), so each document also checks that the rendered length is the
// sum of the node lengths, and that a Split at a link's offset renders the same bytes.
func TestTheRenderIsByteExactOverTheCorpus(t *testing.T) {
	docs := corpus(t)
	t.Logf("corpus: %d documents", len(docs))
	if len(docs) < 100 {
		t.Fatalf("corpus is suspiciously small (%d documents) — the population is wrong", len(docs))
	}
	links, unread := 0, 0
	for path, src := range docs {
		m := ParseMarkdown(src)
		out, err := m.Render()
		if err != nil || out != src {
			t.Errorf("%s: render is not byte-exact (%v)", path, err)
			continue
		}
		if n := docLength(m.doc.Nodes()); n != len(src) {
			t.Errorf("%s: nodes sum to %d bytes, source is %d", path, n, len(src))
		}
		links += len(m.Links())
		for _, u := range m.Unread() {
			unread++
			t.Logf("%s:%d unread: %s", path, u.Line, u.Text)
		}
		if ls := m.Links(); len(ls) > 0 {
			l := ls[len(ls)/2]
			if n := nodeAt(m.doc.Nodes(), l.Offset()); n != nil && n.Location().Offset() < l.Offset() {
				var right sdom.Node
				if err := m.doc.Mutate(func() error {
					var err error
					_, right, err = m.doc.Split(n, l.Offset()-n.Location().Offset())
					if err != nil {
						return err
					}
					return m.doc.Remove(right)
				}); err != nil {
					t.Errorf("%s: split and remove at a link: %v", path, err)
					continue
				}
				if out2, _ := m.Render(); len(out2) != len(src)-right.Location().Length() {
					t.Errorf("%s: render does not follow the nodes after a removal at offset %d (%d bytes, want %d)", path, l.Offset(), len(out2), len(src)-right.Location().Length())
				}
			}
		}
	}
	t.Logf("corpus: %d links, %d unread lines", links, unread)
}

// R462 — SetDest rewrites the destination bytes alone, reads back, and refuses a bad index.
func TestSetDestRewritesTheDestinationBytesAlone(t *testing.T) {
	m := ParseMarkdown(linksFixture)
	if err := m.SetDest(0, "../x/a.md"); err != nil {
		t.Fatal(err)
	}
	if err := m.SetDest(1, "<i m.png>"); err != nil {
		t.Fatal(err)
	}
	links := m.Links()
	if links[0].Raw != "[one](../x/a.md)" || links[1].Raw != "![pic](<i m.png>)" || links[1].Dest != "i m.png" {
		t.Errorf("links after the writes: %+v", links)
	}
	want := strings.Replace(strings.Replace(linksFixture, "[one](a.md)", "[one](../x/a.md)", 1), "![pic](img.png)", "![pic](<i m.png>)", 1)
	if out, _ := m.Render(); out != want {
		t.Errorf("render is not the fixture with two substitutions:\n%s", out)
	}
	if err := m.SetDest(9, "x"); err != ErrNoLink {
		t.Errorf("want ErrNoLink, got %v", err)
	}
}

// R493, R494 — pointers are `doc.md#key` code spans outside fences; SetPointerDoc rewrites
// the document bytes alone.
func TestPointersAreReadAndRewritten(t *testing.T) {
	src := "- **x.** Part `carves/x.md#3`. See `specs/a.md` and `#7` and `R5` and `a b.md`.\n\n```\nPart `carves/fenced.md#1`\n```\n"
	m := ParseMarkdown(src)
	ps := m.Pointers()
	if len(ps) != 2 || ps[0].Doc != "carves/x.md" || ps[0].Key != "3" || ps[0].Raw != "`carves/x.md#3`" || ps[1].Doc != "specs/a.md" || ps[1].Key != "" || ps[0].Line() != 1 {
		t.Fatalf("pointers: %+v", ps)
	}
	if err := m.SetPointerDoc(0, "carves/done/x.md"); err != nil {
		t.Fatal(err)
	}
	want := strings.Replace(src, "`carves/x.md#3`", "`carves/done/x.md#3`", 1)
	if out, _ := m.Render(); out != want {
		t.Errorf("render:\n%s", out)
	}
	if m.Pointers()[0].Key != "3" || m.Pointers()[0].Doc != "carves/done/x.md" {
		t.Errorf("after the write: %+v", m.Pointers()[0])
	}
	if err := m.SetPointerDoc(9, "x"); err != ErrNoPointer {
		t.Errorf("want ErrNoPointer, got %v", err)
	}
}
