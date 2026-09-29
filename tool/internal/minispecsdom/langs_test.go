// CRC: crc-Languages.md | Seq: seq-harvest.md | R509, R510, R511, R512, R518, R525, R526, R528, R529
package minispecsdom

import (
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/zot/simple-dom/sdom"
)

// readRefs parses src with lang, runs the pass, and returns every requirement ref the
// comments carry, with whether the brackets balanced — which is all the harvest sees.
func readRefs(t *testing.T, src string, lang *sdom.BracketLang) (refs []int, balanced bool) {
	t.Helper()
	d, ctx := parse(src, lang)
	cs, err := Comments(d, ctx)
	if err != nil {
		t.Fatalf("Comments: %v", err)
	}
	for _, c := range cs {
		if c.Refs() != nil {
			refs = append(refs, c.Refs().Items()...)
		}
	}
	return refs, len(ctx.Unclosed()) == 0 && len(ctx.Unpaired()) == 0
}

// wrap writes a traceability comment in lang's own style, so fixtures are built from the
// table rather than written as literals a line harvester could mistake for real ones.
func wrap(lang *sdom.BracketLang, interior string) string {
	return lang.Comment.Prefix + interior + lang.Comment.Suffix
}

// R509 — the extension map and the spec's table stay one list.
func TestEveryExtensionMapsToATable(t *testing.T) {
	for _, ext := range []string{".go", ".js", ".ts", ".lua", ".sh", ".bash", ".py", ".pas", ".dpr",
		".c", ".h", ".cpp", ".hpp", ".cc", ".java", ".el", ".html", ".md", ".css"} {
		if _, ok := LanguageFor(ext, nil); !ok {
			t.Errorf("no table for %s", ext)
		}
	}
	if _, ok := LanguageFor(".zig", nil); ok {
		t.Error(".zig has a table")
	}
}

// R510, R512 — every built-in table checks, and writes a comment it reads back.
func TestEveryTableChecksAndReadsItsOwnComment(t *testing.T) {
	for _, ext := range Extensions(nil) {
		lang, _ := LanguageFor(ext, nil)
		if err := lang.Check(); err != nil {
			t.Errorf("%s: %v", ext, err)
			continue
		}
		c := New(lang, Fields{CRC: []string{"crc-X.md"}, Refs: []int{7}})
		src, err := c.Render()
		if err != nil {
			t.Errorf("%s: render: %v", ext, err)
			continue
		}
		if refs, _ := readRefs(t, src, lang); !slices.Equal(refs, []int{7}) {
			t.Errorf("%s: wrote %q, read back refs %v", ext, src, refs)
		}
	}
}

// R510 — a raw string closes only on its own delimiter, so a `//` inside it is text.
func TestCPPRawStringsCloseOnTheirDelimiter(t *testing.T) {
	// The string spans lines, so closing it early at `)"` would expose the R1 comment on
	// the next line as a real one.
	src := "auto s = R\"x( a )\"\n" + wrap(&LangCPP, "R1") + ")x\";\n" + wrap(&LangCPP, "R2")
	refs, balanced := readRefs(t, src, &LangCPP)
	if !slices.Equal(refs, []int{2}) || !balanced {
		t.Errorf("refs %v balanced %v, want [2] true", refs, balanced)
	}
}

// R510 — `"""` is one text block, not an empty string followed by another.
func TestJavaTextBlocksAreOneString(t *testing.T) {
	src := "String s = \"\"\"\n  a \" quote and // R1 not a comment\n  \"\"\";\n" + wrap(&LangJava, "R2")
	refs, balanced := readRefs(t, src, &LangJava)
	if !slices.Equal(refs, []int{2}) || !balanced {
		t.Errorf("refs %v balanced %v, want [2] true", refs, balanced)
	}
}

// R525 — character literals balance, and a `?` ending a name is not one.
func TestElispCharacterLiteralsBalance(t *testing.T) {
	src := "(defun f (c)\n  (list (eq c ?\\)) (eq c ?() (f-exists? c) (f-exists?) ?; ?\" [?\\C-x ?\\^M]))\n" + wrap(&LangElisp, "R3")
	refs, balanced := readRefs(t, src, &LangElisp)
	if !slices.Equal(refs, []int{3}) || !balanced {
		t.Errorf("refs %v balanced %v, want [3] true", refs, balanced)
	}
}

// R511 — a page's comment, its script's and its stylesheet's are read, and nothing else is:
// not the `//` in a URL, not an apostrophe, not a commented-out script.
func TestHTMLReadsPageScriptAndStyleComments(t *testing.T) {
	// R1 sits after a URL on the same line: a `//` live in page text would swallow it.
	src := "<p>don't miss https://example.com (see note)</p>" + wrap(&LangHTML, "R1") + "\n" +
		"<!-- <script>var x = 1; // R9</script> -->\n" +
		"<script type=\"module\">\nfunction f() {\n  // R2\n  return `a ${ {b: 1}.b }`;\n}\n</script>\n" +
		"<style>\n.a { background: url(//cdn/x.png); content: \"/* not */\"; }\n/* R3 */\n</style>\n"
	refs, _ := readRefs(t, src, &LangHTML)
	if !slices.Equal(refs, []int{1, 2, 3}) {
		t.Errorf("refs %v, want [1 2 3]", refs)
	}
}

// R510 — CSS reads only its comments; markdown reads only HTML comments.
func TestCSSAndMarkdownReadOnlyTheirComments(t *testing.T) {
	if refs, _ := readRefs(t, ".a { content: \"/*\"; }\n/* R7 */\n", &LangCSS); !slices.Equal(refs, []int{7}) {
		t.Errorf("css refs %v, want [7]", refs)
	}
	if refs, _ := readRefs(t, "Prose with `code` and (parens) and R9.\n<!-- R8 -->\n", &LangMarkdown); !slices.Equal(refs, []int{8}) {
		t.Errorf("markdown refs %v, want [8]", refs)
	}
}

// R518 — what `query comment-patterns` prints comes from the tables.
func TestCommentFormsReportTheWrittenFormAndItsCloser(t *testing.T) {
	for _, tc := range []struct {
		ext   string
		write CommentForm
		reads string
	}{
		{".go", CommentForm{Open: "// "}, "/*"},
		{".html", CommentForm{Open: "<!-- ", Close: " -->"}, "<!--"},
		{".css", CommentForm{Open: "/* ", Close: " */"}, "/*"},
	} {
		lang, _ := LanguageFor(tc.ext, nil)
		write, read := CommentForms(lang)
		if write != tc.write {
			t.Errorf("%s writes %+v, want %+v", tc.ext, write, tc.write)
		}
		if !slices.ContainsFunc(read, func(f CommentForm) bool { return f.Open == tc.reads }) {
			t.Errorf("%s reads %+v, want one opening %q", tc.ext, read, tc.reads)
		}
	}
}

// R526 — the configuration's code/raw distinction and order reach sdom unchanged.
func TestLanguageDefBuildsTheTableItDescribes(t *testing.T) {
	var cfg struct {
		Languages []LanguageDef `toml:"languages"`
	}
	src := `
[[languages]]
name = "toy"
extensions = [".toy"]
comment = { prefix = "## ", suffix = "\n", kind = "comment" }
  [[languages.brackets]]
  open = ["##"]
  close = "\n"
  allowed_inner = []
  kind = "comment"
  [[languages.brackets]]
  open = ["<"]
  close = ">"
  [[languages.brackets]]
  open = ["{"]
  close = "}"

[[languages]]
name = "indented"
extensions = [".ind"]
comment = { prefix = "# ", suffix = "\n", kind = "comment" }
tab = 4
  [[languages.brackets]]
  open = ["#"]
  close = "\n"
  allowed_inner = []
  kind = "comment"
`
	if _, err := toml.Decode(src, &cfg); err != nil {
		t.Fatal(err)
	}
	lang, err := cfg.Languages[0].Build()
	if err != nil {
		t.Fatal(err)
	}
	if lang.Brackets[0].AllowedInner == nil || len(lang.Brackets[0].AllowedInner) != 0 {
		t.Errorf("`allowed_inner = []` did not build raw: %#v", lang.Brackets[0].AllowedInner)
	}
	if lang.Brackets[1].AllowedInner != nil {
		t.Errorf("an absent allowed_inner did not build code mode: %#v", lang.Brackets[1].AllowedInner)
	}
	if lang.Brackets[1].Open[0] != "<" || lang.Brackets[2].Open[0] != "{" {
		t.Errorf("groups out of the written order")
	}
	if _, err := cfg.Languages[1].Build(); err != nil {
		t.Errorf("indent definition: %v", err)
	}
}

// R528 — a definition sdom rejects is an error naming the language and the group.
func TestLanguageDefRejectedNamesLanguageAndGroup(t *testing.T) {
	d := LanguageDef{Name: "broken", Extensions: []string{".b"},
		Comment:  &CommentDef{Prefix: "// ", Suffix: "\n", Kind: "comment"},
		Brackets: []GroupDef{{Open: []string{"<"}, Close: ">", CloseRegex: ">"}}}
	_, err := d.Build()
	if err == nil || !strings.Contains(err.Error(), "broken") || !strings.Contains(err.Error(), "group 0") {
		t.Errorf("error %v does not name the language and group 0", err)
	}
}

// R529 — the example shipped with the skill loads and checks as a configuration would.
func TestExampleLanguagesConfigLoads(t *testing.T) {
	data, err := os.ReadFile("../../../.claude/skills/mini-spec/languages-example.toml")
	if err != nil {
		t.Fatal(err)
	}
	var cfg struct {
		Languages []LanguageDef `toml:"languages"`
	}
	md, err := toml.Decode(string(data), &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if u := md.Undecoded(); len(u) > 0 {
		t.Errorf("the example sets keys a configuration would refuse: %v", u)
	}
	var names []string
	for _, d := range cfg.Languages {
		lang, err := d.Build()
		if err != nil {
			t.Errorf("%v", err)
			continue
		}
		names = append(names, d.Name)
		// The example claims to copy the built-in tables; hold it to that, so it cannot
		// drift into teaching a table the tool does not use.
		for _, ext := range d.Extensions {
			built, _ := LanguageFor(ext, nil)
			if !reflect.DeepEqual(lang.Brackets, built.Brackets) || lang.Comment != built.Comment {
				t.Errorf("the example's %s differs from the built-in table for %s", d.Name, ext)
			}
		}
	}
	for _, want := range []string{"c", "cpp", "java", "go", "python"} {
		if !slices.Contains(names, want) {
			t.Errorf("the example does not define %s (it defines %v)", want, names)
		}
	}
}
