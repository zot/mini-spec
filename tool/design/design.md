# Minispec Tool Design

## Intent

A Go CLI tool that handles mechanical operations on mini-spec design files, reducing AI token usage for structural queries and updates.

## Cross-cutting Concerns

### Error Handling
All operations return errors with context (file path, line number where applicable). CLI prints errors to stderr and exits with code 1.

### File Encoding
All files are UTF-8. Tool preserves existing line endings (LF/CRLF).

## Artifacts

- crc-Project.md
  - [x] cmd/minispec/main.go
  - [x] internal/project/project.go
- crc-Parser.md
  - [x] internal/parser/types.go
  - [x] internal/parser/requirements.go
  - [x] internal/parser/crc.go
  - [x] internal/parser/design.go
  - [x] internal/parser/traceability.go
- crc-Query.md
  - [x] internal/query/query.go
- crc-Update.md
  - [x] internal/update/update.go
- crc-Validate.md
  - [x] internal/validate/validate.go
- crc-CLI.md
  - [x] internal/cli/cli.go
- crc-Phase.md
  - [x] internal/phase/phase.go
- test-Parser.md
  - [ ] internal/parser/parser_test.go
- test-Update.md
  - [ ] internal/update/update_test.go
- test-Validate.md
  - [ ] internal/validate/validate_test.go
- seq-init.md
- seq-parse.md
- seq-query.md
- seq-update.md
- seq-validate.md
- seq-phase.md

## Documentation

- [x] docs/user-manual.md
- [x] docs/developer-guide.md

## Gaps

- [ ] O1: R37 (MCP server mode) deferred to future version