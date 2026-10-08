# Configuration

## Project Detection

The tool looks for design files in these locations (in order):
1. `./design/` (current directory)
2. Walk up to find a directory containing `design/`

This resolves the **design root** — the directory owning `specs/`, `design/` and
`src/`. It is not the same as the repository root, which owns `.claude/`, `carves/`
and the trajectory files, and which a repository may share between several design
roots. Repository-root detection is a separate search with its own markers; see
[repository-root.md](repository-root.md).

## File Paths

Default paths (can be overridden):
- Requirements: `design/requirements.md`
- Design manifest: `design/design.md`
- CRC cards: `design/crc-*.md`
- Sequences: `design/seq-*.md`
- Source code: `src/` (from Artifacts references)

## Artifacts Format

The Artifacts section in `design.md` uses inline checkbox syntax with optional grouped sections:

```markdown
## Artifacts

### CRC Cards
- [x] crc-Store.md → `src/store.ts`
- [x] crc-View.md → `src/view.ts`, `src/viewlist.ts`

### Sequences
- [x] seq-crud.md → `src/store.ts`, `src/view.ts`

### UI Layouts
- [ ] ui-dashboard.md → `web/html/dashboard.html`

### Test Designs
- [ ] test-Store.md → `src/store_test.ts`
```

Format rules:
- Section headers (`### CRC Cards`, etc.) are optional grouping
- Each line: `- [x] design.md → code-file(s)` or `- [ ] design.md`
- Multiple code files: comma-separated after `→`
- Backticks around code paths are optional (stripped during parsing)
- Checkbox state applies to all code files on that line

## Config Scopes

Configuration lives at two scopes, matching the two roots.

| file | scope | holds |
|---|---|---|
| `<repo root>/.minispec/config.toml` | the repository | settings shared by every design root in it |
| `<design root>/.minispec.toml` | one design root | only what differs from the repository config |
| `<repo root>/.minispec.toml` | — | **an error, with no exception** |

`.minispec/` is a tool-managed directory at the repository root. It holds the
repository configuration and the tool's machine-local working files, which is why
the repository config lives there rather than beside `.git`.

**One setting belongs to the repository scope alone: `track`.** It records whether the
project is version-controlled and whether its trajectory files are private, it is
written only by `minispec init`, and it is verified rather than merely stored. A
design root cannot set or override it, because it describes the repository. See
[initialization.md](initialization.md).

**A top-level `.minispec.toml` is always an error.** Not because the case never
arises, but because the inheritance would be incoherent: it would have to inherit
from `.minispec/config.toml`, a file inside its own directory. Where the repository
root *is* a design root, `.minispec/config.toml` is that design root's configuration
too — there is no second file. The check is one existence test, and it catches a
symlink as readily as a regular file, since a link to the new location is still the
forbidden shape rather than a supported alias.

**Inheritance.** Settings resolve in three layers, each applied over the one before:
built-in defaults, then the repository config, then the design root's own file.

- **Scalars replace.** `design_dir`, `src_dir` — a scalar cannot merge.
- **Lists merge as a union — between configuration layers.** `code_extensions`
  entries from the repository are kept and the design root's additions appended, with
  duplicates dropped and repository order preserved so the result is deterministic.
  **The first configuration layer to set a list replaces the built-in defaults**
  rather than adding to them.

The rule each layer follows is therefore the same one: **state only what you add or
change, never what you keep.** A design root whose settings match the repository
needs no file at all.

*The one asymmetry, and its remedy:* a design root can **add** to a list but cannot
**remove** from one. That is not a dead end — the fix is to drop the setting from the
repository configuration and state it in each design root that needs it, moving the
decision down a level rather than overriding it from below. So removal is always
available, and no removal syntax has to exist. The rule pushes the repository config
toward holding only what is genuinely shared, which is where it belongs anyway.

*And that remedy is exactly why the defaults are exempt.* It works by moving a
setting **down** a level, and nothing sits below the built-in defaults to move it to.
Union onto the shipped `code_extensions` could only ever grow it, so the default list
would become permanently un-narrowable and a project could never say "this one is
Go-only." Defaults are also not a layer anyone authored, so "state only what you add"
cannot sensibly apply to them — you cannot have added to a list you never wrote.

**Provenance.** Because the effective configuration is not what any single file
says, the tool can report which file each setting came from. Otherwise "why is this
value what it is" requires reading two files and knowing the precedence by heart.

**Two layouts, and a project should pick one.**

- **Full-repo** — the repository root is the design root. One project, one config,
  all of it in `.minispec/config.toml`.
- **Repo-projects** — no design root at the top level; each project has its own
  `design/` and its own `.minispec.toml` inheriting the shared settings.

Mixing them — repo-projects *plus* the top level as a project in its own right — is
possible but not recommended: the top level's settings have to live in
`.minispec/config.toml`, so every repo-project then has to override all of them. The
answer is to avoid the shape rather than design around it.

## Config File (Optional)

`.minispec.toml` in a design root:

```toml
design_dir = "design"
src_dir = "src"
```

## Languages

A project can define languages beyond the built-in tables (see
[traceability-comment.md](traceability-comment.md)), or replace a built-in table for the
extensions it names. Each definition is a `[[languages]]` table whose fields are sdom's own
`BracketLang` and `BracketGroup` fields in snake case, so what is written is what sdom
receives:

```toml
[[languages]]
name = "c-family"
extensions = [".c", ".h"]
comment = { prefix = "// ", suffix = "\n", kind = "comment" }

[[languages.brackets]]
open = ["//"]
close = "\n"
allowed_inner = []
kind = "comment"

[[languages.brackets]]
open = ["{"]
close = "}"
```

- **A group's fields** are `open`, `open_regex`, `close`, `close_regex`, `close_is_open`,
  `separators`, `escape`, `after_open`, `before_open`, `before_close`,
  `reject_longer_closes`, `demote_unclosed`, `blank_line_bound`, `line_head_unbound`,
  `allowed_inner`, `allowed_parent` and `kind`, each meaning what sdom's field means.
- **`allowed_inner` absent is code mode; `allowed_inner = []` is raw.** TOML keeps the two
  apart, and the difference is the whole difference between a bracket and a comment.
- **Groups are listed in matching order**, exactly as in sdom: the first opener that matches
  wins, so a longer marker precedes any marker that is its prefix.
- **`comment` is required** for a definition of its own. It is the form written in that
  language, and `query comment-patterns` reports it.
- **`tab`, `transparent` and `continuation`**, when any is set, make the definition an
  indent language, as sdom's `IndentLang` is a `BracketLang` with those three fields.
- **A definition names its extensions, its files, or both.** Extensions override the
  built-in table for each. `files` is a list of patterns matched against the file's path
  relative to the repository with Go's `path.Match` — `*` and `?` within one path segment,
  `[…]` a class, no `**` — so an exact path is a pattern with no special characters:
  `files = ["install/linkapp", "bin/*"]`. A file matching a pattern is read with that
  definition whatever its extension; the first matching definition wins, in the order the
  configuration lists them. A malformed pattern is an error when the configuration loads.
- **A definition may attach files to a built-in table instead of defining one.** One that
  gives a built-in's name (see the table in [traceability-comment.md](traceability-comment.md))
  and `files`, and no `comment`, `brackets` or `extensions`, reads those files with that
  built-in table:

  ```toml
  [[languages]]
  name = "shell"
  files = ["install/linkapp", "install/mcp"]
  ```

  A file whose first line names its interpreter needs no entry at all (see *Interpreter
  lines* there); `files` is for the scripts that do not.
- **Layering:** a design root's definition replaces a repository definition of the same
  name whole, and adds one of a new name; definitions never merge field by field.
- **Every definition is checked when the configuration loads**, with sdom's own check, and a
  definition sdom rejects is an error naming the file, the language and the group.

An example defining C, C++, Java, Go and Python ships in the skill directory as
`languages-example.toml`, to copy from.

## Format

Configuration files are **TOML**. Every file is decoded strictly:

- **An unknown key is an error** naming the file and the key, and the command stops. A key
  the tool does not read would leave someone editing a line with no effect and no way to
  find out, which is the reason a design root stating `track` is refused rather than
  ignored. `comment_patterns` and `comment_closers` get a message of their own: they are
  **retired**, since code files are now read through a language table chosen by extension
  (see [traceability-comment.md](traceability-comment.md)), and how to write a comment in
  each extension is what `query comment-patterns` reports (see [queries.md](queries.md)).
- **A YAML configuration is an error.** The format was YAML until 2026-09-25. A
  `.minispec/config.yaml` or a `.minispec.yaml` found where a configuration would be read is
  named, with the instruction to convert it to TOML by hand, and the command stops. Nothing
  reads YAML, so the tool never guesses at an old file's meaning, and a file converted by
  hand is read under the same strict rules as any other.

## Version

The tool reports its version:
- In the help output header (e.g., `minispec v2.1.0`)
- Via `--version` flag (displays version and exits)
- Via `minispec check-version`: finds the skill's `README.md` in `.claude/skills/mini-spec/` under the **repository root** first, then under the user's home directory; extracts the `Version:` line, and compares it against the tool's version. Exits 0 on match, 1 on mismatch or if not found. `.claude/` is repository-scoped, so the lookup is anchored at the repository root rather than the current directory — see [repository-root.md](repository-root.md).

## Command-Line Flags

All commands accept:
- `--design-dir PATH` - override design directory
- `--src-dir PATH` - override source directory
- `--quiet` - minimal output
- `--json` - output as JSON (for tooling integration)
- `--version` - display version and exit
