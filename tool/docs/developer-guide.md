# Minispec Developer Guide

Architecture and extension guide for the minispec tool.

## Architecture

```
cmd/minispec/main.go     Entry point
internal/
  cli/cli.go             Command-line interface
  project/project.go     Project detection and config
  parser/                 File format parsers
    types.go             Data structures
    requirements.go      Parse requirements.md
    crc.go               Parse CRC cards
    design.go            Parse artifacts & gaps
    traceability.go      Parse code comments
  query/query.go         Read-only operations
  update/update.go       Modification operations
  validate/validate.go   Structural validation
```

## Design Traceability

| Component | CRC Card | Requirements |
|-----------|----------|--------------|
| CLI | crc-CLI.md | R1, R2, R35, R36 |
| Project | crc-Project.md | R32, R33, R34, R35 |
| Parser | crc-Parser.md | R5, R6, R7, R8, R9 |
| Query | crc-Query.md | R10-R17 |
| Update | crc-Update.md | R4, R18-R23 |
| Validate | crc-Validate.md | R3, R24-R31 |

## Key Data Structures

From `internal/parser/types.go`:

```go
type Requirement struct {
    ID       string  // "R1"
    Text     string
    Source   string  // spec file path
    Inferred bool
    Line     int
}

type CRCCard struct {
    Name         string
    Requirements []string  // ["R1", "R3"]
    Path         string
    ReqLine      int
}

type Artifact struct {
    DesignFile string
    CodeFiles  []CodeFile
}

type Gap struct {
    ID          string  // "D1", "R2"
    Type        string  // S, R, D, C, O
    Description string
    Resolved    bool
    Line        int
}
```

## Adding a New Query

1. Add method to `Query` struct in `internal/query/query.go`:

```go
func (q *Query) NewQuery() (ResultType, error) {
    // Use q.Project to locate files
    // Use parser functions to parse
    return result, nil
}
```

2. Add CLI handler in `internal/cli/cli.go` under `runQuery()`:

```go
case "new-query":
    result, err := q.NewQuery()
    // Handle output
```

3. Update help text in `printUsage()`

## Adding a New Update Operation

1. Add method to `Update` struct in `internal/update/update.go`:

```go
func (u *Update) NewOp(args...) error {
    path := u.Project.DesignPath(file)
    // Read, modify, write
    return os.WriteFile(path, content, 0644)
}
```

2. Add CLI handler in `runUpdate()`

## Adding a New Parser

1. Create `internal/parser/newformat.go`:

```go
// CRC: crc-Parser.md
package parser

func ParseNewFormat(path string) (NewType, error) {
    file, err := os.Open(path)
    if err != nil {
        return NewType{}, err
    }
    defer file.Close()

    // Parse with bufio.Scanner and regexp
    return result, nil
}
```

2. Add types to `types.go` if needed

## Adding Validation Checks

1. Add method to `Validate` struct in `internal/validate/validate.go`:

```go
func (v *Validate) validateNewThing(result *ValidationResult) error {
    // Parse relevant files
    // Check structural properties
    // Append to result.Issues if problems found
    return nil
}
```

2. Call from `Run()` method

## Testing

Run from the tool directory:

```bash
go test ./...
```

The tool can validate its own design:

```bash
./minispec validate
```

## Error Handling

- All errors include context (file path, line number where applicable)
- CLI prints errors to stderr and exits with code 1
- Parse errors are collected rather than failing fast

## File Modification Safety

- Updates read the entire file, modify in memory, write back
- Line endings are preserved (split/join on \n)
- File permissions preserved (0644 for new writes)
