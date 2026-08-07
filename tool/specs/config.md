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
| `<repo root>/.minispec/config.yaml` | the repository | settings shared by every design root in it |
| `<design root>/.minispec.yaml` | one design root | only what differs from the repository config |
| `<repo root>/.minispec.yaml` | — | **an error, with no exception** |

`.minispec/` is a tool-managed directory at the repository root. It holds the
repository configuration and the tool's machine-local working files, which is why
the repository config lives there rather than beside `.git`.

**A top-level `.minispec.yaml` is always an error.** Not because the case never
arises, but because the inheritance would be incoherent: it would have to inherit
from `.minispec/config.yaml`, a file inside its own directory. Where the repository
root *is* a design root, `.minispec/config.yaml` is that design root's configuration
too — there is no second file. The check is one existence test, and it catches a
symlink as readily as a regular file, since a link to the new location is still the
forbidden shape rather than a supported alias.

**Inheritance.** Settings resolve in three layers, each applied over the one before:
built-in defaults, then the repository config, then the design root's own file.

- **Scalars replace.** `design_dir`, `src_dir` — a scalar cannot merge.
- **Maps merge per key.** A design root adding one `comment_patterns` entry keeps
  every other entry the repository set.
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
  all of it in `.minispec/config.yaml`.
- **Repo-projects** — no design root at the top level; each project has its own
  `design/` and its own `.minispec.yaml` inheriting the shared settings.

Mixing them — repo-projects *plus* the top level as a project in its own right — is
possible but not recommended: the top level's settings have to live in
`.minispec/config.yaml`, so every repo-project then has to override all of them. The
answer is to avoid the shape rather than design around it.

## Config File (Optional)

`.minispec.yaml` in a design root:

```yaml
design_dir: design
src_dir: src
comment_patterns:
  .go: "//\\s*"
  .js: "//\\s*"
  .ts: "//\\s*"
  .py: "#\\s*"
  .lua: "--\\s*"
  .c: "//\\s*"
  .h: "//\\s*"
  .sh: "#\\s*"
  .pas: "\\{\\s*"
  .dpr: "\\{\\s*"
comment_closers:
  .pas: " }"
  .dpr: " }"
```

## Comment Patterns

The `comment_patterns` map defines regex patterns for single-line comments by file extension. The pattern matches the comment prefix; the tool appends `CRC:` to find traceability comments. The pattern may be an **alternation** — e.g. `.html: "<!--\\s*|//\\s*"` for HTML files that embed JS traceability in `<script>` alongside HTML comments. The tool wraps the configured prefix in a non-capturing group, so the alternation composes correctly (R106).

Default patterns (built-in):
| Extension | Pattern | Languages |
|-----------|---------|-----------|
| `.go` | `//\s*` | Go |
| `.js`, `.ts` | `//\s*` | JavaScript, TypeScript |
| `.c`, `.h`, `.cpp` | `//\s*` | C, C++ |
| `.py` | `#\s*` | Python |
| `.lua` | `--\s*` | Lua |
| `.sh`, `.bash` | `#\s*` | Shell |
| `.md` | `<!--\s*` | Markdown (HTML comments) |
| `.html` | `<!--\s*` | HTML |
| `.css` | `/\*\s*` | CSS |

Custom patterns override defaults for matching extensions.

## Comment Closers

The `comment_closers` map defines closing delimiters for block-comment languages. Extensions not listed use line-terminating comments (no closer needed).

Default closers (built-in):
| Extension | Closer | Languages |
|-----------|--------|-----------|
| `.md` | ` -->` | Markdown (HTML comments) |
| `.html` | ` -->` | HTML |
| `.css` | ` */` | CSS |

Custom closers override defaults for matching extensions. Example for Pascal:

```yaml
comment_closers:
  .pas: " }"
  .dpr: " }"
```

The closer string is stripped from the end of parsed traceability refs. This is necessary because block-comment languages require a closing delimiter after the traceability comment content, and without stripping, the closer would become part of the ref (e.g., `seq-compile.md }` instead of `seq-compile.md`).

**Danger:** Extensions with closers use block comments. An unclosed comment will silently swallow all subsequent code. The `query comment-patterns` command displays a WARNING when closers are configured to alert users to this risk.

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
