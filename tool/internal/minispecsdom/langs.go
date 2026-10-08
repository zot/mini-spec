// CRC: crc-Languages.md | Seq: seq-harvest.md#2.1 | R552
package minispecsdom

import (
	"fmt"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/zot/simple-dom/sdom"
)

// The tables sdom does not ship, built the way its lang.go says a consumer should. Within
// a table the groups are in matching order: the first opener that matches wins, so a
// longer marker precedes any marker that is its prefix, and a group that must win over
// another (a raw string over a plain one) is listed first.

// comment is the kind every table here labels its comment groups with, and the kind its
// Comment style declares, so Comments tries exactly those openers.
const comment = "comment"

// raw is a parse-restricted group's AllowedInner with nothing listed: a pure literal
// interior. Non-nil, which is the whole difference from code mode.
var raw = []string{}

// cGroups is the C family's shared body: two comment forms, two quoted forms, three
// brackets. Each language that extends it puts its own string forms ahead of these.
func cGroups() []sdom.BracketGroup {
	return []sdom.BracketGroup{
		{Open: []string{"//"}, Close: "\n", AllowedInner: raw, Kind: comment},
		{Open: []string{"/*"}, Close: "*/", AllowedInner: raw, Kind: comment},
		{Open: []string{`"`}, Close: `"`, Escape: `\`, AllowedInner: raw},
		{Open: []string{"'"}, Close: "'", Escape: `\`, AllowedInner: raw},
		{Open: []string{"{"}, Close: "}"},
		{Open: []string{"("}, Close: ")"},
		{Open: []string{"["}, Close: "]"},
	}
}

// ahead puts a language's own string forms before the shared body, where they must be to
// win over the plain quote they begin with.
func ahead(first []sdom.BracketGroup, rest []sdom.BracketGroup) []sdom.BracketGroup {
	return append(first, rest...)
}

var slashComment = sdom.CommentStyle{Prefix: "// ", Suffix: "\n", Kind: comment}

// CRC: crc-Languages.md | R510
// LangC is LangGo without the backtick raw string, which C does not have.
var LangC = sdom.BracketLang{Comment: slashComment, Brackets: cGroups()}

// CRC: crc-Languages.md | R510
// LangCPP adds raw strings, R"delim( … )delim": the opener and closer patterns name the
// same capture group, so sdom closes the string only where the delimiter matches the one it
// opened with (mini-spec-tool #43). A delimiter is up to sixteen characters, none of them a
// parenthesis, backslash or space.
var LangCPP = sdom.BracketLang{Comment: slashComment, Brackets: ahead([]sdom.BracketGroup{
	{OpenRegex: `(?:u8|u|U|L)?R"(?P<delim>[^()\\ ]{0,16})\(`, CloseRegex: `\)(?P<delim>[^()\\ ]{0,16})"`, AllowedInner: raw},
}, cGroups())}

// CRC: crc-Languages.md | R510
// LangJava adds text blocks, listed ahead of the plain string so `"""` is not read as an
// empty string followed by another.
var LangJava = sdom.BracketLang{Comment: slashComment, Brackets: ahead([]sdom.BracketGroup{
	{Open: []string{`"""`}, Close: `"""`, Escape: `\`, AllowedInner: raw},
}, cGroups())}

// CRC: crc-Languages.md | R525
// LangElisp reads Emacs Lisp. Its character literals are the reason it needs more than a
// comment and a string: `?(` is the character `(`, and a table without them reads an
// unbalanced bracket, or — for `?"` and `?;` — opens a string or comment that swallows
// what follows. A `?` opens one only at the start of a token (sdom's BeforeOpen,
// mini-spec-tool #45), because `?` also ends predicate names like `f-exists?`; measured
// 2026-09-27 over installed Elisp, 16,571 of those against 76,091 at a token start. The
// closer is the one character after it, escaped or behind `C-`, `M-`, `S-`, `H-`, `s-`,
// `A-` or `^` modifiers: `?a`, `?\(`, `?\C-x`, `?\^M`.
var LangElisp = sdom.BracketLang{
	Comment: sdom.CommentStyle{Prefix: "; ", Suffix: "\n", Kind: comment},
	Brackets: []sdom.BracketGroup{
		{Open: []string{";"}, Close: "\n", AllowedInner: raw, Kind: comment},
		{Open: []string{`"`}, Close: `"`, Escape: `\`, AllowedInner: raw},
		{Open: []string{"?"}, BeforeOpen: "[\\s()\\[\\]'`,#]", CloseRegex: `(?:\\(?:[CMSHsA]-|\^))*\\?.`, AllowedInner: raw},
		{Open: []string{"("}, Close: ")"},
		{Open: []string{"["}, Close: "]"},
	},
}

var htmlComment = sdom.CommentStyle{Prefix: "<!-- ", Suffix: " -->", Kind: comment}

// CRC: crc-Languages.md | R510
// LangMarkdown reads one thing: an HTML comment. Nothing else in a markdown file is code.
var LangMarkdown = sdom.BracketLang{Comment: htmlComment, Brackets: []sdom.BracketGroup{
	{Open: []string{"<!--"}, Close: "-->", AllowedInner: raw, Kind: comment},
}}

// CRC: crc-Languages.md | R510
// LangCSS reads block comments, and both quoted strings so a `/*` inside `content: "…"`
// stays text.
var LangCSS = sdom.BracketLang{
	Comment: sdom.CommentStyle{Prefix: "/* ", Suffix: " */", Kind: comment},
	Brackets: []sdom.BracketGroup{
		{Open: []string{"/*"}, Close: "*/", AllowedInner: raw, Kind: comment},
		{Open: []string{`"`}, Close: `"`, Escape: `\`, AllowedInner: raw},
		{Open: []string{"'"}, Close: "'", Escape: `\`, AllowedInner: raw},
	},
}

// jsParents are the groups JavaScript is live inside, in a page: the script itself and the
// brackets it nests. sdom checks a group's allowed parent against the group immediately
// enclosing it, so a comment in a function body sits under `{`, not `<script`, and every
// JavaScript container has to be listed.
var jsParents = []string{"<script", "{", "(", "[", "${"}

// jsOrStyleParents adds the stylesheet, for the block comment and quoted strings CSS
// shares with JavaScript.
var jsOrStyleParents = append(slices.Clone(jsParents), "<style")

// CRC: crc-Languages.md | R511
// LangHTML reads a page, and the script and style inside it. Outside a script nothing
// JavaScript is live, so the apostrophe in "don't" opens no string and the `//` in a URL
// opens no comment. A stylesheet is restricted to its comments and strings: CSS braces
// would share a name with JavaScript's `{`, and JavaScript's `//` would then come alive in
// `url(//cdn/…)`.
var LangHTML = sdom.BracketLang{Comment: htmlComment, Brackets: []sdom.BracketGroup{
	{Open: []string{"<!--"}, Close: "-->", AllowedInner: raw, Kind: comment},
	{Open: []string{"<script"}, Close: "</script>"},
	{Open: []string{"<style"}, Close: "</style>", AllowedInner: []string{"/*", `"`, "'"}},
	{Open: []string{"//"}, Close: "\n", AllowedInner: raw, Kind: comment, AllowedParent: jsParents},
	{Open: []string{"/*"}, Close: "*/", AllowedInner: raw, Kind: comment, AllowedParent: jsOrStyleParents},
	{Open: []string{"${"}, Close: "}", AllowedParent: []string{"`"}},
	{Open: []string{"`"}, Close: "`", Escape: `\`, AllowedInner: []string{"${"}, AllowedParent: jsParents},
	{Open: []string{`"`}, Close: `"`, Escape: `\`, AllowedInner: raw, AllowedParent: jsOrStyleParents},
	{Open: []string{"'"}, Close: "'", Escape: `\`, AllowedInner: raw, AllowedParent: jsOrStyleParents},
	{Open: []string{"{"}, Close: "}", AllowedParent: jsParents},
	{Open: []string{"("}, Close: ")", AllowedParent: jsParents},
	{Open: []string{"["}, Close: "]", AllowedParent: jsParents},
}}

// CRC: crc-Languages.md | R552
// builtIn is the extension map, sdom's tables and these. Python is an indent language;
// comments are read through its bracket table.
var builtIn = map[string]*sdom.BracketLang{
	".go":   &sdom.LangGo,
	".js":   &sdom.LangJavaScript,
	".ts":   &sdom.LangTypeScript,
	".lua":  &sdom.LangLua,
	".sh":   &sdom.LangShell,
	".bash": &sdom.LangShell,
	".py":   &sdom.LangPython.BracketLang,
	".pas":  &sdom.LangPascal,
	".dpr":  &sdom.LangPascal,
	".c":    &LangC,
	".h":    &LangC,
	".cpp":  &LangCPP,
	".hpp":  &LangCPP,
	".cc":   &LangCPP,
	".java": &LangJava,
	".el":   &LangElisp,
	".html": &LangHTML,
	".md":   &LangMarkdown,
	".css":  &LangCSS,
}

// CRC: crc-Languages.md | R552, R556
// builtInNames is the name each built-in table answers to, the name a configuration attaches
// files to (`name = "shell"`).
var builtInNames = map[string]*sdom.BracketLang{
	"go": &sdom.LangGo, "javascript": &sdom.LangJavaScript, "typescript": &sdom.LangTypeScript,
	"lua": &sdom.LangLua, "shell": &sdom.LangShell, "python": &sdom.LangPython.BracketLang,
	"pascal": &sdom.LangPascal, "c": &LangC, "cpp": &LangCPP, "java": &LangJava,
	"elisp": &LangElisp, "html": &LangHTML, "markdown": &LangMarkdown, "css": &LangCSS,
}

// CRC: crc-Languages.md | R553
// interpreters maps an interpreter line's base name to the built-in it is read with. An
// interpreter not listed leaves its file unread: nothing is guessed.
var interpreters = map[string]string{
	"sh": "shell", "bash": "shell", "zsh": "shell", "dash": "shell", "ksh": "shell",
	"python": "python", "lua": "lua", "luajit": "lua", "node": "javascript", "nodejs": "javascript",
}

// Configured is what a project's configuration adds: a table for each extension a definition
// names, and its `files` rules in configuration order.
type Configured struct {
	Ext   map[string]*sdom.BracketLang
	Files []FileRule
}

// FileRule is one `files` pattern and the table it reads its matches with.
type FileRule struct {
	Pattern string
	Lang    *sdom.BracketLang
}

// CRC: crc-Languages.md | Seq: seq-harvest.md#2.1 | R552, R527
// LanguageFor returns the table an extension is read with: the configured one when a
// definition names it, else the built-in one, and false when there is neither.
func LanguageFor(ext string, configured Configured) (*sdom.BracketLang, bool) {
	if l, ok := configured.Ext[ext]; ok {
		return l, true
	}
	l, ok := builtIn[ext]
	return l, ok
}

// CRC: crc-Languages.md | Seq: seq-harvest.md#2.1 | R552, R553
// LanguageForFile chooses the table a code file is read with: the first configured `files`
// pattern its repository-relative, slash-separated path matches; else its extension's
// table; else the built-in its interpreter line names. False when none answers.
func LanguageForFile(file, firstLine string, configured Configured) (*sdom.BracketLang, bool) {
	// Seq: seq-harvest.md#2.1.1
	for _, r := range configured.Files {
		if ok, _ := path.Match(r.Pattern, file); ok {
			return r.Lang, true
		}
	}
	// Seq: seq-harvest.md#2.1.2
	if l, ok := LanguageFor(path.Ext(file), configured); ok {
		return l, true
	}
	// Seq: seq-harvest.md#2.1.3
	l, ok := builtInNames[interpreters[Interpreter(firstLine)]]
	return l, ok
}

// versionRe is a trailing interpreter version: `3`, `3.11`, `-5.4`.
var versionRe = regexp.MustCompile(`[-.]?\d+(\.\d+)*$`)

// CRC: crc-Languages.md | R553
// Interpreter is the interpreter a `#!` line names: the first word's base name, or under
// `env` the first word that is not an option, a trailing version stripped. Any other line is
// "". Strict on purpose: a guess that picks a wrong comment syntax loses refs silently.
func Interpreter(firstLine string) string {
	rest, ok := strings.CutPrefix(firstLine, "#!")
	if !ok {
		return ""
	}
	words := strings.Fields(rest)
	if len(words) == 0 {
		return ""
	}
	name := path.Base(words[0])
	if name == "env" {
		name = ""
		for _, w := range words[1:] {
			if !strings.HasPrefix(w, "-") {
				name = path.Base(w)
				break
			}
		}
	}
	return versionRe.ReplaceAllString(name, "")
}

// CRC: crc-Languages.md | R518
// Extensions lists every extension with a table, built in or configured, sorted.
func Extensions(configured Configured) []string {
	seen := map[string]bool{}
	for ext := range builtIn {
		seen[ext] = true
	}
	for ext := range configured.Ext {
		seen[ext] = true
	}
	out := make([]string, 0, len(seen))
	for ext := range seen {
		out = append(out, ext)
	}
	sort.Strings(out)
	return out
}

// CommentForm is one way a table writes or reads a comment: an opener, and the closer it
// needs, or "" when a newline ends it.
type CommentForm struct {
	Open, Close string
}

// CRC: crc-Languages.md | R518
// CommentForms reports a table's comment forms: the one it writes, from its Comment style,
// and every comment-kind group it reads.
func CommentForms(lang *sdom.BracketLang) (write CommentForm, read []CommentForm) {
	write = CommentForm{Open: lang.Comment.Prefix, Close: closerOf(lang.Comment.Suffix)}
	for _, g := range lang.Brackets {
		if g.Kind != lang.Comment.Kind {
			continue
		}
		for _, o := range g.Open {
			read = append(read, CommentForm{Open: o, Close: closerOf(g.Close)})
		}
		// A pattern group — Lua's long-bracket comments — is shown by its patterns, since
		// no single text stands for every opener it accepts.
		if g.OpenRegex != "" {
			closer := closerOf(g.Close)
			if g.CloseRegex != "" {
				closer = "/" + g.CloseRegex + "/"
			}
			read = append(read, CommentForm{Open: "/" + g.OpenRegex + "/", Close: closer})
		}
	}
	return write, read
}

// closerOf is what a writer must append: nothing for a comment a newline ends.
func closerOf(groupClose string) string {
	if groupClose == "\n" {
		return ""
	}
	return groupClose
}

// CRC: crc-Languages.md | R555
// GroupDef is one bracket group of a configured language: sdom's BracketGroup field for
// field, in snake case, and in sdom's order — Build converts one to the other directly, so
// a field sdom adds or reorders stops this compiling rather than being silently dropped.
// AllowedInner stays a nil slice when the key is absent, which is code mode, and an empty
// one when it is written `[]`, which is raw.
type GroupDef struct {
	Open               []string `toml:"open"`
	OpenRegex          string   `toml:"open_regex"`
	Separators         []string `toml:"separators"`
	Close              string   `toml:"close"`
	CloseIsOpen        bool     `toml:"close_is_open"`
	CloseRegex         string   `toml:"close_regex"`
	BeforeOpen         string   `toml:"before_open"`
	AfterOpen          string   `toml:"after_open"`
	BeforeClose        string   `toml:"before_close"`
	Escape             string   `toml:"escape"`
	RejectLongerCloses bool     `toml:"reject_longer_closes"`
	DemoteUnclosed     bool     `toml:"demote_unclosed"`
	BlankLineBound     bool     `toml:"blank_line_bound"`
	LineHeadUnbound    bool     `toml:"line_head_unbound"`
	AllowedInner       []string `toml:"allowed_inner"`
	AllowedParent      []string `toml:"allowed_parent"`
	Kind               string   `toml:"kind"`
}

// CommentDef is a configured language's comment style.
type CommentDef struct {
	Prefix string `toml:"prefix"`
	Suffix string `toml:"suffix"`
	Kind   string `toml:"kind"`
}

// CRC: crc-Languages.md | R555
// LanguageDef is one `[[languages]]` configuration entry: a name, the extensions it reads,
// its comment style, its groups in matching order, and the three fields that make sdom's
// IndentLang when any is set.
type LanguageDef struct {
	Name         string      `toml:"name"`
	Extensions   []string    `toml:"extensions"`
	Files        []string    `toml:"files"`
	Comment      *CommentDef `toml:"comment"`
	Brackets     []GroupDef  `toml:"brackets"`
	Tab          int         `toml:"tab"`
	Transparent  string      `toml:"transparent"`
	Continuation string      `toml:"continuation"`
}

// CRC: crc-Languages.md | R555, R528
// Build turns a definition into the table sdom reads, groups in the order written, and
// checks it with sdom's own check, so a malformed definition is an error at load rather than
// a panic at parse. An indent definition's three indent fields are carried as written; its
// bracket table is what is built and checked, and what the harvest reads comments through.
func (d LanguageDef) Build() (*sdom.BracketLang, error) {
	if d.Name == "" {
		return nil, fmt.Errorf("a language definition has no name")
	}
	for _, p := range d.Files { // R555: a malformed pattern is an error at load
		if _, err := path.Match(p, ""); err != nil {
			return nil, fmt.Errorf("language %s: files pattern %q: %w", d.Name, p, err)
		}
	}
	// R556 — a built-in's name with files and nothing of its own attaches to that table.
	if b, ok := builtInNames[d.Name]; ok && len(d.Files) > 0 && d.Comment == nil && len(d.Brackets) == 0 && len(d.Extensions) == 0 {
		return b, nil
	}
	if d.Comment == nil || d.Comment.Prefix == "" {
		return nil, fmt.Errorf("language %s: `comment` is required — it is the form written in this language", d.Name)
	}
	if len(d.Extensions) == 0 && len(d.Files) == 0 {
		return nil, fmt.Errorf("language %s names no extensions and no files", d.Name)
	}
	lang := &sdom.BracketLang{Comment: sdom.CommentStyle{Prefix: d.Comment.Prefix, Suffix: d.Comment.Suffix, Kind: d.Comment.Kind}}
	for _, g := range d.Brackets {
		lang.Brackets = append(lang.Brackets, sdom.BracketGroup(g))
	}
	// The bracket table is what sdom checks: an IndentLang adds no checks of its own
	// (mini-spec-tool's answer to sdom-check-and-matching-close-groups), so the indent
	// fields are carried on the definition as written and not validated here. The harvest
	// reads comments, which the bracket table alone decides.
	if err := lang.Check(); err != nil {
		return nil, fmt.Errorf("language %s: %w", d.Name, err)
	}
	return lang, nil
}
