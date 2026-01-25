# Mini-spec 2.0: Now with teeth

Three weeks ago I released mini-spec: a ~93-line skill for spec-driven development. The idea was simple—add a design layer between specs and code so you can catch AI hallucinations before they become phantom features.

It worked. But there was a problem: **nothing enforced the workflow**.

Claude could skip the requirements phase. It could forget to link requirements to design artifacts. It could mark things implemented that weren't. The traceability was only as good as the AI's attention span.

**GitHub:** https://github.com/zot/mini-spec

**It's still the same 3-level process but now the enforcement is way better:**

```
Level 1: Specs      What you want (human-written)
Level 2: Design     What the AI understood (requirements, CRC cards, sequences)
Level 3: Code       What got built
```

The design level makes the AI's interpretation **explicit and reviewable** before expensive code generation. Reviewing design documents is far cheaper than reviewing code.

## What's new in 2.0

### The `minispec` CLI tool -- automatically used by the mini-spec skill

A Go binary that validates your design structure:

```bash
minispec phase spec           # Verify spec files exist
minispec phase requirements   # Verify requirements.md format
minispec phase design         # Verify all Rn are covered by design artifacts
minispec phase implementation # Verify code has traceability comments
minispec validate             # Run everything
```

The tool catches:

- Requirements not covered by any design artifact
- CRC cards missing their Requirements field
- Code files without traceability comments
- Gaps section malformed
- Design files not listed in Artifacts

This isn't AI validation—it's structural validation. The tool parses markdown and checks references. Fast, deterministic, no tokens burned.

### requirements.md: The "nothing gets lost" layer

Before, specs linked directly to CRC cards. Now there's an intermediate step:

```
specs/feature.md
    ↓
design/requirements.md    ← NEW: numbered requirements (R1, R2...)
    ↓
design/crc-*.md           ← References: R1, R3, R7
    ↓
src/*.ts                  ← Traceability comments
```

Every spec item becomes exactly one numbered requirement. Every CRC card declares which requirements it implements. The tool validates complete coverage.

**Why this matters:** When you have 50 spec items across 6 files, things get lost. A requirement might be clear in your head but never make it to design. Or it's in design but no code implements it. Numbered requirements with tool validation close these gaps.

### Phase gates

The skill now requires running `minispec phase <name>` after each workflow phase. Claude can't jump from specs to implementation—it has to pass through requirements and design first, with validation at each step.

### Mandatory task tracking

Before reading any files or doing any work, Claude must create tasks:

```
TaskCreate: "Spec Phase: user authentication"
TaskCreate: "Requirements Phase: user authentication"
TaskCreate: "Design Phase: user authentication"
TaskCreate: "Implementation Phase: user authentication"
...
```

This gives you real-time visibility into where Claude is in the workflow. You see tasks transition from pending → in_progress → completed. No more wondering "is it still designing or did it start coding?"

During implementation, Claude breaks down further into per-file tasks. You can watch progress file by file instead of waiting for a wall of code at the end.

## Why a tool? Why not a database?

I considered several approaches:

**Structured data (JSON/YAML/SQLite):** Easier to parse, but harder to review. The whole point of mini-spec is human-readable design docs. If I can't `cat design/requirements.md` and see what's there, I've lost the plot.

**AI-based validation:** Burns tokens, non-deterministic, slow. The tool runs in milliseconds and gives the same answer every time.

**No validation:** What I had before. It worked until it didn't.

**A CLI tool parsing markdown** threads the needle: design docs stay human-readable, but structure is enforced. The tool understands the format well enough to validate references and coverage without understanding intent.

## The numbers

| Metric              | Jan 7 (v0.9.3) | Now (v2.1.0)            |
|---------------------|----------------|-------------------------|
| Skill lines         | ~93            | ~267                    |
| Tool                | None           | ~2000 lines Go          |
| Validation          | Manual         | Automated per-phase     |
| Requirements        | Implicit       | Explicit, numbered      |
| Progress visibility | None           | Mandatory task tracking |

The skill grew 3x because it now documents the requirements.md format, phase validation commands, and tool usage. Worth it for the guarantees.

## What stayed the same

- Three-level architecture (specs → design → code)
- CRC cards, sequences, UI layouts
- Traceability comments in code
- Drift detection via Artifacts checkboxes
- Gaps section for tracking discrepancies
- Optional spec-agent for isolated context

## Getting started

1. Install the tool:
   ```bash
   # Download from https://github.com/zot/mini-spec/releases
   chmod +x minispec
   mv minispec ~/.claude/bin/
   ```

2. Install the code-simplifier plugin:
   ```bash
   claude plugin install code-simplifier
   ```

3. Copy `mini-spec/` to your project's `.claude/skills/`

4. Write specs, then tell Claude: **"/mini-spec design this"**

**Or if you already have mini-spec just ask claude, `update the mini-spec skill`!**

The tool will keep Claude honest at every phase.

I've been using this heavily across four projects. It works. And the tool saves a lot of tokens by handling validation that would otherwise require AI parsing.
