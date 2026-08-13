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
- [ ] crc-Project.md → `cmd/minispec/main.go`, `internal/project/project.go`, `internal/project/config.go`
- [x] crc-RepoRoot.md → `internal/project/reporoot.go`
- [x] crc-Git.md → `internal/project/git.go`
- [x] crc-Track.md → `internal/project/track.go`
- [x] crc-Init.md → `internal/project/init.go`
- [ ] crc-Parser.md → `internal/parser/testdoc.go`, `internal/parser/types.go`, `internal/parser/requirements.go`, `internal/parser/crc.go`, `internal/parser/design.go`, `internal/parser/traceability.go`, `internal/parser/seqdoc.go`
- [x] crc-Query.md → `internal/query/query.go`
- [x] crc-Update.md → `internal/update/update.go`
- [ ] crc-Validate.md → `internal/validate/validate.go`
- [x] crc-CLI.md → `internal/cli/cli.go`, `internal/cli/bootstrap.go`
- [x] crc-Phase.md → `internal/phase/phase.go`
- [ ] crc-Alarm.md → `internal/alarm/alarm.go`

### Sequences
- [x] seq-init.md
- [x] seq-parse.md
- [x] seq-query.md
- [x] seq-update.md
- [x] seq-validate.md
- [x] seq-phase.md
- [x] seq-reporoot.md
- [ ] seq-config.md
- [x] seq-bootstrap.md
- [ ] seq-alarm-freshness.md

### Test Designs
- [ ] test-Parser.md → `internal/parser/parser_test.go`
- [ ] test-Update.md → `internal/update/update_test.go`
- [ ] test-Validate.md → `internal/validate/validate_test.go`
- [x] test-RepoRoot.md → `internal/project/reporoot_test.go`
- [ ] test-Config.md → `internal/project/config_test.go`
- [x] test-Git.md → `internal/project/git_test.go`
- [x] test-Track.md → `internal/project/track_test.go`
- [x] test-Init.md → `internal/project/init_test.go`
- [x] test-Bootstrap.md → `internal/cli/bootstrap_test.go`
- [ ] test-Alarm.md → `internal/alarm/alarm_test.go`

## Documentation

- [x] docs/user-manual.md
- [x] docs/developer-guide.md

## Gaps

- A1: R37 (MCP server mode) deferred to future version
- A2: R1-R66 pre-existing code lacks inline requirement refs

- T1: R56 retired by R117 (2026-08-07 repository-root detection)
- [ ] O1: query uncovered lists retired requirements while validate skips them — R56 appears there now that its ref was correctly removed, reading as work to do. Both are defensible for their purpose, but the raw query invites a reader to re-cover a requirement that is deliberately dead
- [ ] O2: DetectFrom in project.go duplicates the new isDir helper from reporoot.go (os.Stat + IsDir inline). One-line reuse, same package, noticed during simplification of the repo-root work
- [ ] O3: Project.RootPath is the design root but its name says neither — the exact ambiguity R107 exists to remove. Renaming touches every call site, so it was left out of the repository-root work rather than folded in
- [ ] O4: No test covers loadProject end-to-end against a real design root: resolveConfigFrom is well covered, but the wiring from Detect through loadProject to a Project with Origins populated is only exercised by running the binary
- [ ] O5: gate() and runInit() have no automated test: both resolve the repository root from the process working directory, so testing them needs os.Chdir — process-global state, unsafe alongside parallel tests. The refusal messages, the exempt/known command sets and the flag table are covered directly (test-Bootstrap.md); the wiring between them is only exercised by running the binary, which was done by hand end-to-end this session. Repair: thread a start directory through gate as RepoRootFrom already does for RepoRoot
- [ ] O6: Git.Tracked reports a genuine git failure as "not tracked". ls-files --error-unmatch exits non-zero both when a path is untracked and when the command itself fails, and the error is discarded. Harm is bounded — a spurious preference note, never a refusal — but it is the project's own "report absence as error, never as silence" theme violated at a seam that exists to interrogate an external tool. Repair: distinguish exit 1 from other failures