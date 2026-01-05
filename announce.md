# Announcing mini-spec: spec-driven development for Claude in ~93 lines

AI hallucinates features you didn't ask for. Sometimes they're great, so good you rely on them. Then you regenerate and they vanish. Mini-spec fixes this.

**GitHub:** https://github.com/zot/mini-spec

## The three-level approach

```
specs/     What you asked for (human-written)
   ↓
design/    What the AI understood (reviewable)
   ↓
src/       What got built (traceable)
```

The middle layer is **key**. It shows how Claude interprets your specs: what you left out, where the ambiguities are, what you over-specified. Design docs are small and readable compared to code, so you can verify correctness before a single line is written.

## How it tracks drift

`design.md` serves as the project's status memory. Each design file lists implementation files with checkboxes. When Claude changes code, it unchecks affected artifacts so it knows:

1. A design review is needed
2. Exactly what to look at

There's also a "Gaps" section tracking discrepancies between specs, design, and implementation.

## Why mini-spec?

| Feature | Mini-Spec | Heavy Frameworks |
|---------|-----------|------------------|
| Skill file | ~93 lines | 500+ lines |
| Agents | 1 optional | Multiple required |
| Scripts | None | Build tooling |
| Learning curve | Minutes | Hours |

This supersedes my earlier [claude-crc](https://github.com/dustinwloring1988/claude-crc) methodology. Smaller, cleaner, and what I'm using now.

## What's included

- **Skill** (~93 lines) for design and implementation workflow
- **Optional agent** for running design tasks in isolated context (keeps main conversation lean)
- **Full example app** (Contacts manager with specs, design, and implementation)

Mini-spec lets you make designs, implement them with tests and docs, update code, reverse-engineer existing code, and keep everything in sync.
