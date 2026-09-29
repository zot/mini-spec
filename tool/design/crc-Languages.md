# Languages
**Requirements:** R509, R510, R511, R512, R525, R526, R528, R529

The bracket tables mini-spec reads code with, the map from a file's extension to one, and the
tables a project defines in its configuration. sdom ships Go, JavaScript, Lua, Shell, Python
and Pascal; the ones here are languages it does not, built the way sdom's own `lang.go` says
a consumer should. Package `minispecsdom`.

## Knows
- LangC: `LangGo`'s groups without the backtick raw string
- LangCPP: `LangC` plus a raw-string group — `OpenRegex` `(?:u8|u|U|L)?R"(?P<delim>…)\(`,
  `CloseRegex` `\)(?P<delim>…)"`, raw inside — ahead of the plain string
- LangJava: `LangC` plus a `"""` text-block group, raw with `\` escape, ahead of `"`
- LangElisp: `;` to newline, raw, kind `comment`; `"` strings with `\` escape; `( )` and
  `[ ]`; a character-literal group — `Open` `?`, `BeforeOpen` a token-start class, raw,
  `CloseRegex` one character, escaped or with `C-`/`M-`/`^` modifiers
- LangMarkdown: one group, `<!--` to `-->`, `AllowedInner: []`, kind `comment`
- LangCSS: `/*` to `*/` raw, kind `comment`, and both quoted strings
- LangHTML: `<!--` to `-->` raw and live everywhere; `<script` to `</script>` in code mode;
  `<style` to `</style>` restricted to `/*` and both quotes; every JavaScript group with
  `AllowedParent` naming `<script` and the JavaScript brackets `{`, `(`, `[`, `${`. The
  JavaScript line comment never lists `<style`, so it is live in a script and nowhere else
- builtIn: extension → table for every row of the spec's table, sdom's and these
- every table's `Comment` style: the one form written in that language
- LanguageDef: a `[[languages]]` configuration entry — name, extensions, `comment`, `brackets`
  with sdom's `BracketGroup` fields in snake case, and the three indent fields — decoded by
  `Project` with TOML tags declared here, since this package owns what a table is

## Does
- LanguageFor(ext, configured): the configured table for an extension when a definition names
  it, else the built-in one, and false when there is neither
- Extensions(configured): every extension with a table, built in or configured, sorted
- CommentForms(lang): every comment form the table recognizes — each `comment`-kind group's
  openers and closer — for reporting what the reader accepts beside what is written
- (LanguageDef) Build(): the `BracketLang` a definition describes, groups in the order written,
  `allowed_inner` absent → nil (code) and `[]` → empty (raw). An indent definition's `tab`,
  `transparent` and `continuation` are carried as written: sdom's `IndentLang` adds no checks
  of its own, and the harvest reads comments through the bracket table. Checked with sdom's
  `Check()`; its error, prefixed with the language's name,
  is returned for the loader to name the file (R526, R528)

## Collaborators
- sdom.BracketLang, sdom.IndentLang: the table types, their `Check()`, and the tables sdom ships
- TraceabilityComment: reads comments through whichever table is chosen; it needs no change
  per language, since `Comments` tries every opener whose group kind is the table's
  `Comment.Kind`
- Project: decodes `LanguageDef` from each configuration layer and layers them by name

## Constraints
- **A table's bracket order is matching order and nothing else.** The first matching opener
  wins, so a longer marker precedes its prefix (`--[[` before `--`). The form to write is
  `Comment`, never "the first comment group"
- **`AllowedParent` is checked against the immediately enclosing group**, which is why
  HTML's JavaScript groups list the JavaScript brackets as well as `<script`: a comment in a
  function body sits under `{`
- **CSS braces are never groups inside HTML.** A group is named by its opener text, so a CSS
  `{` and a JavaScript `{` would be one name, and JavaScript's `//` would become live in a
  stylesheet's `url(//…)`. Restricting `<style` to comments and strings avoids it; the
  stylesheet's structure is not needed to read its comments
- **A restricted group lists the char-literal group by its literal opener `?`** (or, for a
  pattern group, by its exact `OpenRegex` text) — sdom names groups that way (its R295)
- **The example configuration is kept honest by a test**: `languages-example.toml` in the
  skill directory is decoded and every definition built, as a configuration would be (R529)

## Sequences
- seq-harvest.md
