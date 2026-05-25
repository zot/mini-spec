# Project Guidelines

## Releases

Version is tracked in `.claude/skills/mini-spec/README.md`. Increment the patch number unless the user says otherwise.

Release contents (zip file):
- `.claude/skills/mini-spec/` (skill files)
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

If `.claude/skills/mini-spec/SKILL.md` changes, regenerate `.cursor/rules/mini-spec.mdc` so Cursor stays aligned:

`python tool/gen_cursor_rule.py`

## Summary specs

The Go tool's specs live in `tool/specs/`. Summary specs there index per-feature specs along a cross-cutting axis (CLI surface, storage layout, API set, capabilities). Mini-spec's normal anchoring doesn't catch drift in them, so when you add, rename, or retire something along one of these axes, update the matching summary spec by hand.

Summary specs in `tool/specs/`:

- [`cli-commands.md`](tool/specs/cli-commands.md) — full inventory of `minispec` subcommands and global flags. Update whenever a subcommand is added, renamed, or retired.
- [`file-formats.md`](tool/specs/file-formats.md) — every file the tool reads or writes, with links to the per-feature spec that defines each format. Update when a file format is added, changed, or retired.

When you create another, list it here so future maintainers know which summary specs to keep in sync.
