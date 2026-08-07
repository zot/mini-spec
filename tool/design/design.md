# Minispec Tool Design

## Intent

A Go CLI tool that handles mechanical operations on mini-spec design files, reducing AI token usage for structural queries and updates.

## Cross-cutting Concerns

### Error Handling
All operations return errors with context (file path, line number where applicable). CLI prints errors to stderr and exits with code 1.

### File Encoding
All files are UTF-8. Tool preserves existing line endings (LF/CRLF).

## Artifacts

### CRC Cards
- [x] crc-Project.md → `cmd/minispec/main.go`, `internal/project/project.go`
- [x] crc-RepoRoot.md → `internal/project/reporoot.go`
- [ ] crc-Parser.md → `internal/parser/types.go`, `internal/parser/requirements.go`, `internal/parser/crc.go`, `internal/parser/design.go`, `internal/parser/traceability.go`, `internal/parser/seqdoc.go`
- [x] crc-Query.md → `internal/query/query.go`
- [x] crc-Update.md → `internal/update/update.go`
- [ ] crc-Validate.md → `internal/validate/validate.go`
- [x] crc-CLI.md → `internal/cli/cli.go`
- [x] crc-Phase.md → `internal/phase/phase.go`

### Sequences
- [x] seq-init.md
- [x] seq-parse.md
- [x] seq-query.md
- [x] seq-update.md
- [x] seq-validate.md
- [x] seq-phase.md
- [x] seq-reporoot.md

### Test Designs
- [ ] test-Parser.md → `internal/parser/parser_test.go`
- [ ] test-Update.md → `internal/update/update_test.go`
- [ ] test-Validate.md → `internal/validate/validate_test.go`
- [x] test-RepoRoot.md → `internal/project/reporoot_test.go`

## Documentation

- [x] docs/user-manual.md
- [x] docs/developer-guide.md

## Gaps

- A1: R37 (MCP server mode) deferred to future version
- A2: R1-R66 pre-existing code lacks inline requirement refs

- T1: R56 retired by R117 (2026-08-07 repository-root detection)
- [ ] O1: query uncovered lists retired requirements while validate skips them — R56 appears there now that its ref was correctly removed, reading as work to do. Both are defensible for their purpose, but the raw query invites a reader to re-cover a requirement that is deliberately dead
- [ ] O2: DetectFrom in project.go duplicates the new isDir helper from reporoot.go (os.Stat + IsDir inline). One-line reuse, same package, noticed during simplification of the repo-root work