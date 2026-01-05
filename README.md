# Mini-Spec

**Create and maintain designs at minimal token cost.**

## The Problem

Generative AI hallucinates features you didn't ask for. Sometimes they're good—so good you come to rely on them, maybe even prefer them to what you specified.

But they're not in your spec. They exist only in the AI's interpretation.

When you update the code or regenerate the design, those features can *spontaneously disappear*. You might not notice until after you deliver. Imagine explaining to users why their favorite feature vanished—one you never actually designed.

## The Solution

A 3-level architecture that makes the AI's interpretation explicit:

```
specs/     What you asked for (human-written)
   ↓
design/    What the AI understood (reviewable)
   ↓
src/       What got built (traceable)
```

**The design layer is the checkpoint.** Review it before code generation. If the AI added features you didn't spec, decide now: add them to the spec, or remove them from the design. Either way, nothing disappears without you knowing.

Design docs are *much* smaller than code—easier for humans to review, and they fit in AI context windows where full codebases don't. Verify correctness at the design level while it's still cheap to fix.

## Why Mini-Spec?

| Feature | Mini-Spec | Heavy Frameworks |
|---------|-----------|------------------|
| Skill file | ~93 lines | 500+ lines |
| Agents | 1 optional | Multiple required |
| Scripts | None | Build tooling |
| Learning curve | Minutes | Hours |

Mini-spec gives you just enough structure to catch misalignment early, without the overhead of a full design methodology. See [methodology.md](.claude/skills/mini-spec/methodology.md) for background on CRC cards and how the design layer works.

## Quick Start

1. Copy `.claude/skills/mini-spec/` to your project's `.claude/skills/`
2. Write a spec in `specs/your-feature.md`
3. Tell Claude: **"design this"** → generates design docs, no code
4. Review the design, iterate if needed
5. Tell Claude: **"implement this"** → writes code, updates traceability

## What's Included

```
.claude/skills/mini-spec/
├── SKILL.md          # The skill definition (~93 lines)
└── methodology.md    # CRC background and theory

.claude/agents/
└── spec-agent.md     # Optional agent for isolated context

example/              # Full working Contacts app
├── specs/            # Human spec
├── design/           # Generated design (CRC, sequences, UI)
├── src/              # Implementation
└── docs/             # User & developer docs
```

## Workflow

The skill enforces phase separation:

- **"Design"** = design only. No code generated.
- **"Implement"** = code only. Updates Artifacts checkboxes.
- **Code changes independently** = flags drift, asks if you want to update design or specs.

This prevents scope creep and keeps documents in sync.

## Using the Agent

The `spec-agent` runs design/implementation tasks in isolated context. Use it when:

- **Context is getting large** — delegate design work to keep your main conversation lean
- **Validating designs** — agent loads methodology, checks consistency, reports findings
- **Batch operations** — create multiple CRC cards or implement several components

Invoke via Task tool:
```
Task(subagent_type="spec-agent", prompt="validate the design in example/design")
```

The agent automatically loads the mini-spec skill and follows the methodology. Results return to your main context as a summary.

## When to Use

- Token budgets matter
- Small-medium projects
- Rapid iteration
- Solo or small team
- You want verification before code generation

## When NOT to Use

- Large enterprise projects requiring detailed process documentation
- Teams needing comprehensive modeling (use full [CRC methodology](https://github.com/dustinwloring1988/claude-crc))
- Projects where token cost isn't a concern

## License

MIT
