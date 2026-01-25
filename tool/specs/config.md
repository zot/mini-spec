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

Custom patterns override defaults for matching extensions.

## Command-Line Flags

All commands accept:
- `--design-dir PATH` - override design directory
- `--src-dir PATH` - override source directory
- `--quiet` - minimal output
- `--json` - output as JSON (for tooling integration)

## MCP Server Mode

`minispec serve` runs as an MCP server for direct AI integration.

Exposes all query and update operations as MCP tools.
