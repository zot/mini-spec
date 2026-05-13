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
// CRC: crc-Store.md | Seq: seq-crud.md#1.4
// CRC: crc-Store.md | Seq: seq-crud.md#1.4, seq-other.md#2.1
```

The `Seq:` value is one or more references separated by commas. Each
reference is a sequence-diagram filename, optionally followed by `#`
and a dotted-number anchor pointing at a specific step inside that
diagram. The bare filename remains valid for diagrams that aren't
numbered.

### Numbered Sequence Diagrams

Sequence diagrams may number their steps so code comments can pin
to a specific step. Numbers are dotted (1, 1.1, 1.1.1, ...), local
to the file, and placed wherever the diagram style allows:

- Tree/outline style: `1.4. action` on the line with the step
- UML actor-lane style: number on the line directly above the arrow
- Mermaid/pseudo-Mermaid: number at the start of the step

A file may contain more than one numbered diagram. Items in the
first numbered diagram begin with `1.`, the second with `2.`, and
so on. The first dotted segment is the diagram index *K*; items
within diagram *K* all share that prefix.

The full anchor `file.md#K.x.y` is what's globally unique. A bare
number is local to its file — `1.4` in seq-foo.md is unrelated to
`1.4` in seq-bar.md. Within a single file, every dotted ID must
appear at most once.
