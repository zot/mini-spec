# Project Guidelines

## Releases

Version is tracked in `.claude/skills/mini-spec/README.md`. Increment the patch number unless the user says otherwise.

Release contents (zip file):
- `.claude/skills/mini-spec/` (skill files)
- `.claude/skills/minimap/` (skill files) — **required**, not optional: `SKILL.md`'s
  first instruction is to invoke `/minimap`, so a zip without it ships a skill whose
  opening directive points at nothing. v2.9.0 ships exactly that broken pair: minimap
  was split out in `520a783` (2026-06-07), the tag landed 2026-06-09, and the zip's
  SKILL.md names `/minimap` three times with no minimap alongside it.
- `.claude/agents/spec-agent.md` (agent)
- Exclude Emacs backup files: `zip -x "*.~undo-tree~"`

Release process:
1. Check existing tags (`git tag -l`) to confirm next version number
2. Update version in `.claude/skills/mini-spec/README.md`
3. Commit the skill README (ask user about other new or changed files)
4. Tag it (e.g., `v0.9.3`)
5. Push commit + tag
6. Create GitHub release with `gh release create`
7. Upload zip to the release
8. Build and upload binaries: `cd tool && make release` then `gh release upload <tag> bin/*`
9. Include in release notes: `claude plugin install code-simplifier`

Note: Binaries must be rebuilt for each release because they embed the version via `--version`.

## Cursor

`.cursor/rules/{mini-spec,minimap}.mdc` are generated from the two skills. If either
`SKILL.md` changes, regenerate both:

`python3 tool/gen_cursor_rule.py`

Don't rely on remembering — `make validate` runs `gen_cursor_rule.py --check`, which
regenerates in memory and fails if either rule on disk differs. Cursor has no
skill-invocation mechanism, so minimap ships as its own always-applied rule rather than
as a pointer, and the generator rewrites the cross-references between them.

The generator is a chain of anchored text edits over prose maintained elsewhere, so
every edit asserts that its anchor matched. Two of them silently became no-ops when
SKILL.md's headings drifted — one in `520a783`, the commit that split minimap out — and
the shipped rule spent months with no Cursor-integration section while still telling
Cursor agents to call `TaskCreate`. If an anchor fails, update the anchor; never drop
the edit.

`.cursor/rules/mini-spec-spec-agent.mdc` is hand-written, not generated.

## Summary specs

The Go tool's specs live in `tool/specs/`. Summary specs there index per-feature specs along a cross-cutting axis (CLI surface, storage layout, API set, capabilities). Mini-spec's normal anchoring doesn't catch drift in them, so when you add, rename, or retire something along one of these axes, update the matching summary spec by hand.

Summary specs in `tool/specs/`:

- [`cli-commands.md`](tool/specs/cli-commands.md) — full inventory of `minispec` subcommands and global flags. Update whenever a subcommand is added, renamed, or retired.
- [`file-formats.md`](tool/specs/file-formats.md) — every file the tool reads or writes, with links to the per-feature spec that defines each format. Update when a file format is added, changed, or retired.

When you create another, list it here so future maintainers know which summary specs to keep in sync.
