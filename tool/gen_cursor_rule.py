"""Generate .cursor/rules/mini-spec.mdc from .claude/skills/mini-spec/SKILL.md."""
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
src = root / ".claude" / "skills" / "mini-spec" / "SKILL.md"
dst = root / ".cursor" / "rules" / "mini-spec.mdc"
body = src.read_text(encoding="utf-8").replace("\r\n", "\n")
body = re.sub(r"(?ms)^---\n.*?^---\n+", "", body, count=1)
body = body.replace("~/.claude/bin/minispec", "minispec")
body = body.replace("(Serena, Grep, etc.)", "(Grep, codebase search, etc.)")

cursor_block = """## Cursor integration

**`minispec` executable:** Use `minispec` on `PATH`, or the path in environment variable `MINISPEC` when running shell commands (PowerShell: `& $env:MINISPEC check-version`). Match tool version to `.claude/skills/mini-spec/README.md` via `minispec check-version`.

**Progress visibility (Claude `TaskCreate` / `TaskUpdate` parity):** Before reading files or starting phase work, create **todos** for each applicable phase (Spec, Requirements, Design, Implementation, Simplification, Gaps). During implementation, add per-file todos. Mark todos **completed** as phases finish—same intent as Claude Code tasks.

**Simplification (Claude `code-simplifier` parity):** No separate Cursor plugin. After Implementation phase, do one **in-session simplification pass** on changed files: clarity, consistency, naming—**no behavior change**—then Gaps.

---
"""

body = body.replace(
    "# Mini-spec\n\n## Design Docs First",
    f"# Mini-spec\n\n{cursor_block}## Design Docs First",
    1,
)

body = body.replace(
    "## MANDATORY: Create Tasks First\n\n"
    "**BEFORE reading any files or doing any work**, create tasks for applicable phases:\n\n"
    "```\n"
    'TaskCreate: "Spec Phase: [feature name]"\n'
    'TaskCreate: "Requirements Phase: [feature name]"\n'
    'TaskCreate: "Design Phase: [feature name]"\n'
    'TaskCreate: "Implementation Phase: [feature name]"\n'
    'TaskCreate: "Simplification Phase: [feature name]"\n'
    'TaskCreate: "Gaps Phase: [feature name]"\n'
    "```\n\n"
    "Do NOT proceed until tasks exist. This is required for user visibility into progress.\n\n"
    "---\n\n",
    "## MANDATORY: todos before work\n\n"
    "**BEFORE reading any files or doing any work**, create todos for applicable phases, for example:\n\n"
    "- Spec Phase: [feature name]\n"
    "- Requirements Phase: [feature name]\n"
    "- Design Phase: [feature name]\n"
    "- Implementation Phase: [feature name]\n"
    "- Simplification Phase: [feature name]\n"
    "- Gaps Phase: [feature name]\n\n"
    "Do NOT proceed until these todos exist (user visibility).\n\n"
    "---\n\n",
    1,
)

body = body.replace(
    "## Task Tracking\n\n"
    "**During implementation**, break down into per-file tasks:\n"
    "```\n"
    'TaskCreate: "Implement view.ts changes"\n'
    'TaskCreate: "Implement viewlist.ts changes"\n'
    'TaskCreate: "Update design docs"\n'
    "```\n\n"
    "**Mark phases complete** with TaskUpdate as you finish them.\n"
    "**Use Quality Checklist items** as tasks before finalizing.\n",
    "## Task tracking\n\n"
    "**During implementation**, break down into per-file todos (e.g. implement `view.ts`, update design docs).\n\n"
    "**Mark phases complete** in the todo list as you finish them.\n"
    "**Use Quality Checklist items** as todos before finalizing.\n",
    1,
)

body = body.replace(
    "Invoke the `code-simplifier` agent on the recently modified code. "
    "This refines code for clarity, consistency, and maintainability while preserving functionality.",
    "Perform the **in-session simplification pass** (see Cursor integration): "
    "refine recently modified code for clarity, consistency, and maintainability without changing behavior.",
    1,
)

body = body.replace(
    "See `config-reference.md` (in this skill directory)",
    "See `.claude/skills/mini-spec/config-reference.md`",
    1,
)

body = body.replace(
    "The `minispec` CLI tool (at `minispec`) performs structural operations",
    "The `minispec` CLI tool (install on PATH or set `MINISPEC`) performs structural operations",
    1,
)

front = """---
description: MANDATORY before writing or modifying code. Read design/design.md first; use minispec CLI; specs to design to code. Same methodology as Claude mini-spec skill.
alwaysApply: true
---
"""

dst.parent.mkdir(parents=True, exist_ok=True)
dst.write_text(front + body, encoding="utf-8", newline="\n")
print("wrote", dst)
