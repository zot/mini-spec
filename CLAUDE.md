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
