# Configuration

## Project Detection

The tool looks for design files in these locations (in order):
1. `./design/` (current directory)
2. Walk up to find a directory containing `design/`

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

## Config File (Optional)

`.minispec.yaml` in project root:

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

The `comment_patterns` map defines regex patterns for single-line comments by file extension. The pattern matches the comment prefix; the tool appends `CRC:` to find traceability comments.

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
- Via `minispec check-version`: finds the skill's `README.md` in `.claude/skills/mini-spec/` (project-level first, then user-level `~/`), extracts the `Version:` line, and compares it against the tool's version. Exits 0 on match, 1 on mismatch or if not found.

## Command-Line Flags

All commands accept:
- `--design-dir PATH` - override design directory
- `--src-dir PATH` - override source directory
- `--quiet` - minimal output
- `--json` - output as JSON (for tooling integration)
- `--version` - display version and exit

## MCP Server Mode

`minispec serve` runs as an MCP server for direct AI integration.

Exposes all query and update operations as MCP tools.
