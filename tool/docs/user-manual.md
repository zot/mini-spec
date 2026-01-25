# Minispec User Manual

A CLI tool for structural operations on mini-spec design files.

## Installation

```bash
go install github.com/zot/minispec/cmd/minispec@latest
```

Or build from source:
```bash
cd tool
go build -o minispec ./cmd/minispec
```

## Quick Start

```bash
# Run from any directory within a mini-spec project
minispec validate

# List requirements
minispec query requirements

# Check which requirements lack design coverage
minispec query uncovered
```

## Project Detection

Minispec automatically finds your project by walking up from the current directory until it finds a `design/` folder. You can override paths with flags:

```bash
minispec --design-dir /path/to/design validate
```

## Commands

### validate

Runs all structural validations and shows what was found.

```bash
minispec validate
```

Output shows:
- Requirements found (R1, R2, ...)
- Which CRC cards reference which requirements
- Coverage (covered vs uncovered requirements)
- Artifacts with checkbox states
- Gaps
- Issues (problems found)

Exit code: 0 if no issues, 1 if issues found.

### query requirements

Lists all requirements from `requirements.md`.

```bash
minispec query requirements
```

### query coverage

Shows which design files reference each requirement.

```bash
minispec query coverage
```

### query uncovered

Lists requirements with no design file references.

```bash
minispec query uncovered
```

### query orphan-designs

Lists CRC cards missing a `**Requirements:**` field.

```bash
minispec query orphan-designs
```

### query artifacts

Lists all artifacts with their checkbox states.

```bash
minispec query artifacts
```

### query gaps

Lists gap items from `design.md`.

```bash
minispec query gaps
```

### query traceability

Checks code files for traceability comments.

```bash
# Single file
minispec query traceability src/store.go

# All code files in artifacts
minispec query traceability --all
```

### update check / uncheck

Toggle checkboxes in design files.

```bash
# Check a gap item
minispec update check design.md D1

# Check an artifact code file
minispec update check design.md src/store.go

# Uncheck
minispec update uncheck design.md src/store.go
```

### update add-ref / remove-ref

Manage requirement references in CRC cards.

```bash
# Add R5 to crc-Store.md
minispec update add-ref crc-Store.md R5

# Remove R5
minispec update remove-ref crc-Store.md R5
```

### update add-gap

Add a new gap item with auto-numbered ID.

```bash
# Types: S (spec), R (requirement), D (design), C (code), O (oversight)
minispec update add-gap R "Requirement R5 has no design coverage"
# Output: Added R1: Requirement R5 has no design coverage
```

### update resolve-gap

Mark a gap as resolved (checks its checkbox).

```bash
minispec update resolve-gap D1
```

### phase

Run phase-specific validation after completing each workflow phase. Each phase command validates only the artifacts relevant to that phase.

```bash
# After completing spec phase
minispec phase spec

# After completing requirements phase
minispec phase requirements

# After completing design phase
minispec phase design

# After completing implementation phase
minispec phase implementation

# After completing gaps phase
minispec phase gaps
```

| Phase | Validates |
|-------|-----------|
| `spec` | Spec files exist in `specs/` and are non-empty |
| `requirements` | requirements.md format, sequential Rn numbering, spec sources exist |
| `design` | Design files, CRC cards have Requirements field, requirement coverage |
| `implementation` | Code files exist, have traceability comments, refs point to existing files |
| `gaps` | Gaps section structure, ID format, no duplicates |

Exit code: 0 if phase passes, 1 if issues found.

## Global Flags

| Flag | Description |
|------|-------------|
| `--design-dir PATH` | Override design directory |
| `--src-dir PATH` | Override source directory |
| `--quiet` | Minimal output |
| `--json` | Output as JSON |

## JSON Output

Use `--json` for machine-readable output:

```bash
minispec --json query gaps
```

## Configuration

Optional `.minispec.yaml` in project root:

```yaml
design_dir: design
src_dir: src
code_extensions:
  - .go
  - .ts
  - .lua
comment_patterns:
  .go: "//\\s*"
  .ts: "//\\s*"
  .lua: "--\\s*"
```

### Comment Patterns

The `comment_patterns` map defines regex patterns for single-line comments by file extension. The tool appends `CRC:` to find traceability comments.

Default patterns (built-in):
| Extension | Pattern | Languages |
|-----------|---------|-----------|
| `.go` | `//\s*` | Go |
| `.js`, `.ts` | `//\s*` | JavaScript, TypeScript |
| `.c`, `.h`, `.cpp` | `//\s*` | C, C++ |
| `.py` | `#\s*` | Python |
| `.lua` | `--\s*` | Lua |
| `.sh`, `.bash` | `#\s*` | Shell |

Custom patterns in `.minispec.yaml` override defaults for matching extensions.

## File Formats

### requirements.md

```markdown
## Feature: feature-name
**Source:** specs/feature.md

- **R1:** requirement text
- **R2:** (inferred) requirement text
```

### CRC Cards

```markdown
# ClassName
**Requirements:** R1, R3, R7
## Knows
...
```

### Traceability Comments

The comment pattern is determined by file extension:

Go, JavaScript, TypeScript, C:
```go
// CRC: crc-Store.md | Seq: seq-crud.md
func Add() {}
```

Python, Shell:
```python
# CRC: crc-Store.md | Seq: seq-crud.md
def add():
```

Lua:
```lua
-- CRC: crc-Store.md | Seq: seq-crud.md
function add()
```
