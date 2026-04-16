# Agent instructions (Cursor)

## Mini-spec

This repo uses the **mini-spec** methodology. Cursor loads it via [`.cursor/rules/mini-spec.mdc`](.cursor/rules/mini-spec.mdc) (`alwaysApply: true`) — same workflow as the Claude Code skill in `.claude/skills/mini-spec/SKILL.md`.

**Install `minispec`:** add the [release](https://github.com/zot/mini-spec/releases) binary to your `PATH`, or set environment variable `MINISPEC` to its full path. Run `minispec check-version` and align with the version in `.claude/skills/mini-spec/README.md`.

**Claude parity:**

| Claude | Cursor |
|--------|--------|
| Skill auto-invoked | Rule `mini-spec.mdc` always applies |
| `TaskCreate` / `TaskUpdate` | Cursor **todos** (see rule) |
| `~/.claude/bin/minispec` | `minispec` on `PATH` or `MINISPEC` |
| `code-simplifier` plugin | In-session **simplification pass** (rule) |
| `spec-agent` subagent | Use **Agent** mode + optional rule below |

For isolated design-heavy work (equivalent to invoking **spec-agent**), enable the rule **mini-spec — spec agent** from the rule picker or mention it in chat.

## Other

Release and contribution notes for maintainers: [`CLAUDE.md`](CLAUDE.md).
