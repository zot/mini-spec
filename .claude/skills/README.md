# Mini-Spec Skills

## mini-spec/SKILL.md

The core skill for design and implementation workflow. Invoke with `/mini-spec`.

**What it does:**
- Enforces phase separation (design vs implementation)
- Defines CRC card, sequence diagram, and test case formats
- Maintains traceability between specs → design → code
- Detects drift when code changes without design updates

**Usage:**
- "Design X" — creates design artifacts, no code
- "Implement X" — writes code with traceability, updates Artifacts checkboxes
- Code changes independently — flags affected designs for review

See [methodology.md](mini-spec/methodology.md) for CRC background.

## Using with spec-agent

For large context situations, delegate to the agent instead of invoking the skill directly:

```
Task(subagent_type="spec-agent", prompt="design the authentication module")
```

The agent runs in isolated context, loads the skill automatically, and returns a summary. This keeps your main conversation lean while the agent handles the detailed design work.

**Agent location:** `.claude/agents/spec-agent.md`
