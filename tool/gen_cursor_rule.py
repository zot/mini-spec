"""Generate .cursor/rules/{mini-spec,minimap}.mdc from the Claude skills.

Each rule is the skill's SKILL.md with a chain of anchored edits applied, so
Cursor gets the same methodology without the Claude-only machinery (the Skill
tool, TaskCreate/TaskUpdate, the code-simplifier agent).

Every edit asserts that its anchor matched. That is not defensive habit: the
anchors are text in a document maintained elsewhere, so an edit whose anchor
drifts becomes a silent no-op and the script still prints "wrote". Two of them
did exactly that for months -- one since the commit that split minimap out --
and the shipped rule lost its whole Cursor-integration section and kept telling
Cursor agents to call TaskCreate. A generator over someone else's prose has to
fail loudly or it does not fail at all.
"""
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[1]
skills = root / ".claude" / "skills"
rules = root / ".cursor" / "rules"


def sub(body, old, new, *, expect=1):
    """Replace `old` with `new`, requiring exactly `expect` matches.

    Pass expect=None for "one or more" where the count is incidental.
    """
    found = body.count(old)
    if found == 0 or (expect is not None and found != expect):
        want = "1 or more" if expect is None else expect
        sys.exit(
            f"gen_cursor_rule: anchor matched {found}x, expected {want}x:\n"
            f"    {old[:140]!r}\n"
            "The source SKILL.md changed under this script. Update the anchor "
            "rather than dropping the edit -- a missing edit ships silently."
        )
    return body.replace(old, new)


def load(skill):
    """Read a skill body with its YAML front matter stripped."""
    text = (skills / skill / "SKILL.md").read_text(encoding="utf-8").replace("\r\n", "\n")
    return re.sub(r"(?ms)^---\n.*?^---\n+", "", text, count=1)


check_only = "--check" in sys.argv[1:]
outputs = []


def emit(name, front, body):
    """Queue a rule. Nothing is written until every rule has been built, so a
    failure partway through leaves the whole .cursor/rules/ directory alone
    instead of updating one rule and abandoning the other."""
    outputs.append((rules / name, front + body))


# --- minimap: the structural model -----------------------------------------
#
# Cursor has no skill-invocation mechanism, so minimap cannot be a pointer the
# way `/minimap` is under Claude. It ships as its own always-applied rule, and
# the cross-references between the two become references to the other rule.

minimap = load("minimap")
minimap = sub(
    minimap,
    "**model**, not the process. Load `/minimap` to *navigate*; load `/mini-spec`\n"
    "to *maintain* (it loads this skill underneath).",
    "**model**, not the process. The `mini-spec` rule carries the *process* and\n"
    "builds on this one; both apply at all times.",
)
minimap = sub(minimap, "`/mini-spec` owns it", "the `mini-spec` rule owns it", expect=2)

emit(
    "minimap.mdc",
    """---
description: How a mini-spec project is laid out — the 3-level spec/design/code model, the root spec index, and the traceability links. Read before navigating or changing a mini-spec project. Companion to the mini-spec rule.
alwaysApply: true
---
""",
    minimap,
)


# --- mini-spec: the process -------------------------------------------------

body = load("mini-spec")
body = sub(body, "~/.claude/bin/minispec", "minispec", expect=None)
body = sub(body, "(Serena, Grep, etc.)", "(Grep, codebase search, etc.)")

cursor_block = """## Cursor integration

**`minispec` executable:** Use `minispec` on `PATH`, or the path in environment variable `MINISPEC` when running shell commands (PowerShell: `& $env:MINISPEC check-version`). Match tool version to `.claude/skills/mini-spec/README.md` via `minispec check-version`.

**Progress visibility (Claude `TaskCreate` / `TaskUpdate` parity):** Before reading files or starting phase work, create **todos** for each applicable phase (Spec, Requirements, Design, Implementation, Simplification, Gaps). During implementation, add per-file todos. Mark todos **completed** as phases finish—same intent as Claude Code tasks.

**Simplification (Claude `code-simplifier` parity):** No separate Cursor plugin. After Implementation phase, do one **in-session simplification pass** on changed files: clarity, consistency, naming—**no behavior change**—then Gaps.

---
"""

body = sub(body, "# Mini-spec\n\n## Load the model first", f"# Mini-spec\n\n{cursor_block}## Load the model first")

body = sub(
    body,
    "**IMMEDIATELY invoke `/minimap` using the Skill tool before doing anything else.** It carries",
    "**Read the `minimap` rule (`.cursor/rules/minimap.mdc`) before doing anything else.** It applies alongside this rule and carries",
)
body = sub(body, "is in `/minimap`", "is in the `minimap` rule", expect=1)
body = sub(body, "are in `/minimap`", "are in the `minimap` rule", expect=1)

body = sub(
    body,
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
    "Do NOT proceed until tasks exist. This is required for user visibility into progress.\n"
    "\n"
    "**If this harness has no task tool at all** \u2014 no `TaskCreate`, no `TodoWrite`, nothing\n"
    "under any other name \u2014 the requirement does not lapse, it relocates: list the phases in\n"
    "your response before starting, and name each one as you enter and finish it. What is\n"
    "mandatory here is that the user can see which phase you are in. The task list is *how*,\n"
    "not *what*, and a mandate with no defined outcome in the world the reader is standing in\n"
    "gets ignored whole \u2014 along with the version check and the migration check either side of\n"
    "it, which are real.\n",
    "## MANDATORY: todos before work\n\n"
    "**BEFORE reading any files or doing any work**, create todos for applicable phases, for example:\n\n"
    "- Spec Phase: [feature name]\n"
    "- Requirements Phase: [feature name]\n"
    "- Design Phase: [feature name]\n"
    "- Implementation Phase: [feature name]\n"
    "- Simplification Phase: [feature name]\n"
    "- Gaps Phase: [feature name]\n\n"
    "Do NOT proceed until these todos exist (user visibility).\n"
    "\n"
    "**If this mode has no todo list**, the requirement does not lapse, it relocates: list the phases in your response before starting, and name each one as you enter and finish it. What is mandatory here is that the user can see which phase you are in. The todo list is *how*, not *what*, and a mandate with no defined outcome in the world the reader is standing in gets ignored whole \u2014 along with the version check and the migration check either side of it, which are real.\n",
)

body = sub(
    body,
    "## Task Tracking\n\n"
    "**During implementation**, break down into per-file tasks:\n"
    "```\n"
    'TaskCreate: "Implement view.ts changes"\n'
    'TaskCreate: "Implement viewlist.ts changes"\n'
    'TaskCreate: "Update design docs"\n'
    "```\n\n"
    "**Mark phases complete** with TaskUpdate as you finish them.\n"
    "**Use Quality Checklist items** as tasks before finalizing.\n"
    "\n"
    "**With no task tool**, the same breakdown and the same completions go in your responses.\n",
    "## Task tracking\n\n"
    "**During implementation**, break down into per-file todos (e.g. implement `view.ts`, update design docs).\n\n"
    "**Mark phases complete** in the todo list as you finish them.\n"
    "**Use Quality Checklist items** as todos before finalizing.\n"
    "\n"
    "**With no todo list**, the same breakdown and the same completions go in your responses.\n",
)

body = sub(
    body,
    "Invoke the `code-simplifier` agent on the recently modified code. "
    "This refines code for clarity, consistency, and maintainability while preserving functionality.",
    "Perform the **in-session simplification pass** (see Cursor integration): "
    "refine recently modified code for clarity, consistency, and maintainability without changing behavior.",
)

# The in-session pass needs no plugin, so the note on a missing one does not apply.
body = sub(
    body,
    "\n\n`code-simplifier` is a standard Anthropic plugin. If it is not available, skip the "
    "simplification pass and tell the user they can install it with "
    "`claude plugin install code-simplifier`. The alarm pulls below still run.",
    "",
)

body = sub(
    body,
    "See `config-reference.md` (in this skill directory)",
    "See `.claude/skills/mini-spec/config-reference.md`",
)

# Cursor has no skill directory, so every on-demand reference has to become a repo
# path or it points at nothing. The parenthetical goes first, because the general
# rewrite below would otherwise leave "(in this skill directory)" hanging off a path
# that is no longer in one.
body = sub(
    body,
    "**Every shape is normative in `trajectory-format.md`** (in this skill directory),",
    "**Every shape is normative in `.claude/skills/mini-spec/trajectory-format.md`**,",
)

# Then every remaining bare mention, however many there are. Deliberately a class
# rewrite rather than one anchor per sentence: SKILL.md gains references to this file
# as the layer grows, and an anchor list extended by hand is a list that will be
# forgotten -- the failure this whole script was hardened against. expect=None because
# the count is incidental, and asserting it would turn every new mention into a
# spurious build break.
body = sub(
    body,
    "`trajectory-format.md`",
    "`.claude/skills/mini-spec/trajectory-format.md`",
    expect=None,
)

body = sub(
    body,
    "The `minispec` CLI tool (at `minispec`) performs structural operations",
    "The `minispec` CLI tool (install on PATH or set `MINISPEC`) performs structural operations",
)

# Prose *about* the harness task tool, not an instruction to call it — but the
# argument for trajectory tracking is the same one either way, so it survives
# the rename rather than being dropped.
body = sub(
    body,
    "what just landed. The harness task tool (`TaskCreate`/`TaskUpdate`) is\n"
    "session-local and dies with the session.",
    "what just landed. The editor's todo list is session-local and dies with\n"
    "the session.",
)

# Nothing Claude-only may survive into the rule. These are the tool names an
# agent could actually try to call, which is the failure worth catching. The
# injected Cursor block is exempt because it names them on purpose, to state
# what each maps to -- so the check runs over everything except that block.
inherited = body.replace(cursor_block, "")
leftovers = [name for name in ("TaskCreate", "TaskUpdate", "`/minimap`", "Skill tool") if name in inherited]
if leftovers:
    sys.exit(f"gen_cursor_rule: Claude-only references survived into the rule: {', '.join(leftovers)}")

emit(
    "mini-spec.mdc",
    """---
description: MANDATORY before writing or modifying code. Read design/design.md first; use minispec CLI; specs to design to code. Same methodology as Claude mini-spec skill.
alwaysApply: true
---
""",
    body,
)


# --- write or verify ---------------------------------------------------------
#
# `--check` is the forcing function. Making the generator fail loudly on a
# drifted anchor only helps if someone runs it, and "regenerate when SKILL.md
# changes" is a discipline with nothing behind it -- which is how the shipped
# rule went months without a Cursor-integration section. `make validate` runs
# this, so a stale rule is a computed property rather than something to
# remember.

if check_only:
    stale = [path for path, text in outputs
             if not path.exists() or path.read_text(encoding="utf-8") != text]
    if stale:
        sys.exit(
            "gen_cursor_rule: these Cursor rules no longer match the skills they mirror:\n"
            + "".join(f"    {p}\n" for p in stale)
            + "Run `python3 tool/gen_cursor_rule.py` to regenerate."
        )
    print("cursor rules up to date")
else:
    rules.mkdir(parents=True, exist_ok=True)
    for path, text in outputs:
        path.write_text(text, encoding="utf-8", newline="\n")
        print("wrote", path)
