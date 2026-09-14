# Requirements

## Feature: node protocol
**Source:** specs/node-protocol.md

- **R1:** A `Node` renders its bytes and reports its location; equality is per kind and
  ignores provenance.
- **R2:** (inferred) A compound's children tile its span.
- **~~R3:~~** (Retired T1 — see R7) A node carries a parent pointer.
- **~~R4:~~** (Retired T2 — no replacement) Nodes are registered with the document.

### Notes

A sub-heading owns its own content. `## Feature:` above ends here.

- **R5:** A requirement under the notes belongs to the notes.

## Feature: fences
**Source:** specs/fences.md

An example of the form, quoted rather than stated:

```markdown
## Feature: quoted
- **R99:** quoted, not a requirement
```

- **R6:** The fence above is body.
- **R7:** The last entry of the section, one line.

## Feature: empty
**Source:** specs/empty.md
