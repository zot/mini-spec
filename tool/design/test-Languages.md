# Test Design: Languages
**Source:** crc-Languages.md

Each table is exercised through `Comments` over a small source, so what is pinned is what the
harvest will see: which comments are read, and whether the brackets balance.

## Test: every extension in the spec maps to a table
**Purpose:** the extension map and the spec's table stay one list
**Input:** each extension R552 names
**Expected:** `LanguageFor` returns a table for each, and none for `.zig`
**Refs:** crc-Languages.md — R552

## Test: every built-in table checks, and writes a comment it reads back
**Purpose:** a table sdom would reject, or whose `Comment` style it cannot read back, fails here
rather than on a user's file
**Input:** every table in the map
**Expected:** `Check()` passes; `New(lang, fields)` renders a comment that `Comments` reads back
with the same fields
**Refs:** crc-Languages.md — R510, R512

## Test: C++ raw strings close only on their own delimiter
**Purpose:** a comment marker inside a raw string is text
**Input:** a raw string `R"x( a )"` continued over a line holding a traceability comment R1,
closed by `)x"`, then a traceability comment R2
**Expected:** only R2 is read — R1 sits inside the string, which closes only at `)x"`; brackets
balance
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 3
**Fire alarm:** remove the named `delim` groups from both raw-string patterns, so the string closes
at the first `)` followed by a quote, and confirm the `// R1` inside it is read
**Inject:** internal/minispecsdom/langs.go:LangCPP
**Pulled:** 2026-09-29 — rang: refs [1], unbalanced; the first pull, whose exposed text was no valid comment, stayed green and the fixture was rewritten; restore byte-clean
**Refs:** crc-Languages.md — R510

## Test: a Java text block is one string
**Purpose:** `"""` is not an empty string followed by another
**Input:** a text block holding `"` and `// not a comment`, then a traceability comment
**Expected:** the text block holds both; only the traceability comment is read
**Refs:** crc-Languages.md — R510

## Test: Emacs Lisp reads its comments, and its character literals balance
**Purpose:** char literals are the 17% case; `?` ending a name is not one
**Input:** `(eq c ?\))`, `(f-exists? x)`, `(f-exists?)` — a `?` ending a name right before a
closer, the measured `_update?)` shape — `?;` and a `;` traceability comment
**Expected:** brackets balance, no char literal inside `f-exists?`, `?;` opens no comment, the
traceability comment is read
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 1
**Fire alarm:** remove the char-literal group's `BeforeOpen`, so `?` opens one inside `f-exists?`,
and confirm the brackets no longer balance
**Inject:** internal/minispecsdom/langs.go:LangElisp
**Pulled:** 2026-09-29 — rang: "balanced false"; the first pull, on a fixture with no `?)`, stayed green and the fixture was rewritten; restore byte-clean
**Refs:** crc-Languages.md — R525

## Test: HTML reads its page, script and style comments, and nothing else
**Purpose:** ark's case, structurally
**Input:** an `<!--` traceability comment on the same line as page text holding `https://x`
(where a live `//` would swallow it), a `// R5` inside `<script>`, a `/* R6 */` inside
`<style>`, page text with `https://x` and `don't`, a stylesheet `url(//cdn/x.png)`, and a
commented-out `<script>`
**Expected:** exactly the three traceability comments are read; nothing unbalances
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 2
**Fire alarm:** drop `AllowedParent` from HTML's `//` group, so a `//` in page text opens a
comment, and confirm the refs no longer match
**Inject:** internal/minispecsdom/langs.go:LangHTML
**Pulled:** 2026-09-29 — rang: refs [2 3], R1 swallowed; the first pull, with R1 on its own line, stayed green and the fixture was rewritten; restore byte-clean
**Refs:** crc-Languages.md — R511

## Test: CSS and Markdown read only their comments
**Purpose:** a `/*` inside a CSS string is text; markdown prose is never code
**Input:** CSS `content: "/*"` then `/* R7 */`; markdown prose with backticks and `<!-- R8 -->`
**Expected:** R7 and R8 read; nothing else
**Refs:** crc-Languages.md — R510

## Test: comment forms are reported with the written one and its closer
**Purpose:** what `query comment-patterns` prints comes from the tables
**Input:** `CommentForms` for `.go`, `.html` and `.css`
**Expected:** `.go` writes `// ` with no closer and also reads `/* */`; `.html` writes `<!-- `
` -->` with a closer; `.css` writes `/* ` ` */` with a closer
**Refs:** crc-Languages.md — R518

## Test: a definition builds the table it describes
**Purpose:** the configuration's code/raw distinction and order reach sdom unchanged
**Input:** a definition with a group lacking `allowed_inner`, one with `allowed_inner = []`, and
three groups in a stated order; a second with `tab` set
**Expected:** nil and empty `AllowedInner` respectively, groups in the written order; the second
builds an indent language
**Refs:** crc-Languages.md — R555

## Test: a definition sdom rejects names the language and the group
**Purpose:** a malformed table is an error at load, never a panic at parse
**Input:** a definition whose group has both `close` and `close_regex`
**Expected:** an error naming the language and the group
**Refs:** crc-Languages.md — R528

## Test: the example configuration loads
**Purpose:** the file shipped to copy from must work as shipped
**Input:** `.claude/skills/mini-spec/languages-example.toml`
**Expected:** it decodes; every definition builds and checks; C, C++, Java, Go and Python present
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 4
**Fire alarm:** change one C group in the example (drop its `escape`), and confirm the test
reports the example's c differing from the built-in table
**Inject:** .claude/skills/mini-spec/languages-example.toml
**Pulled:** 2026-09-29 — rang: "the example's c differs from the built-in table" for .c and .h; restore byte-clean
**Refs:** crc-Languages.md — R529

## Test: a file's table is chosen by pattern, then extension, then interpreter line
**Purpose:** validates R552 — the order, and that each step answers only when the one before it did not
**Input:** a configured rule `bin/*` → a toy table, and a configured `.go`; files `bin/tool.go`, `x.go`, `x.sh`, `install/linkapp` opening `#!/bin/bash`, `notes` opening `hello`
**Expected:** `bin/tool.go` the toy table (the pattern beats its extension); `x.go` the configured `.go`; `x.sh` shell; `install/linkapp` shell by its interpreter; `notes` none
**Refs:** crc-Languages.md, seq-harvest.md#2.1 — R552
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 5
**Fire alarm:** consult the extension before the patterns. Red: `bin/tool.go` reads through Go, not the pattern's table
**Inject:** internal/minispecsdom/langs.go:LanguageForFile
**Pulled:** 2026-10-08 — rang, by hand after simplification: consulting the extension before the patterns reads `bin/tool.go` through the configured `.go` table, not the pattern's, in `TestAFilesTableIsChosenByPatternThenExtensionThenInterpreter`; minispecsdom, parser, project run unfiltered with `-count=1`; restore byte-clean

## Test: an interpreter line is read strictly
**Purpose:** validates R553 — the base name, `env` and its options skipped, a version stripped, and nothing else read as one
**Input:** `#!/bin/bash`; `#! /usr/bin/env -S python3.11 -u`; `#!/usr/bin/env node`; `#!/usr/bin/perl`; `# a comment`; `class Foo:`
**Expected:** `bash`; `python`; `node`; `perl` (which maps to no table); "" for the last two
**Refs:** crc-Languages.md — R553
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 6
**Fire alarm:** take the first word after `env` without skipping options. Red: the second line reads as `-S`
**Inject:** internal/minispecsdom/langs.go:Interpreter
**Pulled:** 2026-10-08 — rang, by hand after simplification: not skipping `env`'s options reads `#! /usr/bin/env -S python3.11 -u` as `-S` in `TestAnInterpreterLineIsReadStrictly`; restore byte-clean

## Test: a definition attaches files to a built-in by name
**Purpose:** validates R555, R556 — `name` and `files` alone attach to the built-in; a pattern is checked at load; a definition of its own still needs `comment`
**Input:** `{name: shell, files: [install/linkapp]}`; `{name: shell, files: ["["]}`; `{name: zig, files: [a]}`
**Expected:** the first builds to `LangShell`; the second is an error naming the language and the pattern; the third is an error that `comment` is required
**Refs:** crc-Languages.md — R555, R556
**Code:** internal/minispecsdom/langs_test.go
**Alarm:** 7
**Fire alarm:** drop the attach branch in `Build`, so an attaching definition is held to a definition's own rules. Red: `language shell: comment is required`
**Inject:** internal/minispecsdom/langs.go:LanguageDef.Build
**Pulled:** 2026-10-08 — rang, by hand after simplification: with the attach branch unreachable, `TestADefinitionAttachesFilesToABuiltIn` gets `language shell: comment is required`; restore byte-clean

## Test: configured files reach the harvest in order
**Purpose:** validates R555, R556 — `files` decodes from config.toml, an attaching definition's rules read with the built-in, and the rules keep configuration order
**Input:** a repository config: `shell` attaching `install/linkapp` and `bin/*`, then a `toy` definition of its own with `files = ["tools/*"]`
**Expected:** `Languages().Files` patterns `install/linkapp bin/* tools/*`; the first reads with `LangShell`, the last with the toy's own table
**Refs:** crc-Project.md — R555, R556
**Code:** internal/project/config_test.go
**Alarm:** 8
**Fire alarm:** drop the `files` loop in `Project.Languages`. Red: `patterns [], want configuration order`
**Inject:** internal/project/config.go:Project.Languages
**Pulled:** 2026-10-08 — rang, by hand after simplification: dropping the `files` loop in `Project.Languages` gives `patterns [], want configuration order` in `TestLanguagesCarryFilesInOrder`; restore byte-clean
