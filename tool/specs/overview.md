# Minispec Tool

Go CLI tool for querying and updating structured parts of mini-spec design files.

**Language:** Go
**Environment:** CLI, single binary

## Purpose

Reduce AI token usage by handling mechanical operations on design files:
- Traceability queries (which requirements are covered?)
- Structural gap detection (missing references, unchecked items)
- Atomic updates (check/uncheck, add references)

The tool does NOT interpret intent—it only works with the formal structure.

## File Formats It Understands

### requirements.md
```markdown
## Feature: feature-name
**Source:** specs/feature.md

- **R1:** requirement text
- **R2:** (inferred) requirement text
```

### CRC Cards (crc-*.md)
```markdown
# ClassName
**Requirements:** R1, R3, R7
...
```

### design.md Artifacts Section
```markdown
## Artifacts
- crc-Store.md
  - [x] src/store.ts
  - [ ] src/store_test.ts
```

### design.md Gaps Section
```markdown
## Gaps
- [ ] S1: description
- [ ] R1: description
- [x] D1: description (resolved)
```

### Code Traceability Comments
```
// CRC: crc-Store.md | Seq: seq-crud.md
```
