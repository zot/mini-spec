# Design: Sample

## Intent

A document with a gaps section, for the reader.

## Artifacts

- [x] crc-Sample.md → `sample.go`

## Gaps

- A1: R37 (MCP server mode) deferred to future version
- T1: R56 retired by R117 (2026-08-07 repository-root detection)
  - reason: the root moved from the design directory to the
    repository
- [x] I1: R12 has design coverage but no inline ref in any code file, because no parse
  context exists yet. Closes when Item 2 lands a schema with a parse context.
- [ ] O1: R28 understates the rule the code implements. It says a compound is altered if any of
  its children is; `Compound.Location` also requires the children's spans to run **contiguously
  from the compound's own offset**.
- [ ] O2: Test coverage gaps
  - [ ] Feature A (5 scenarios)
  - [ ] Feature B (3 scenarios)

- [ ] O3: A document about the format quotes an entry in a fence:

  ```markdown
  - [ ] O99: quoted, not a gap
  ```

  and the quoted entry is body.
- [ ] O4: last entry, one line

## Notes

- [ ] O5: a bullet after the section is not a gap
