# Project Guidelines

## Releases

Version is tracked in `.claude/skills/mini-spec/README.md`. Increment the patch number unless the user says otherwise.

Release contents (zip file):
- `.claude/skills/mini-spec/` (skill files)
- `.claude/agents/spec-agent.md` (agent)

Release process:
1. Update version in `.claude/skills/mini-spec/README.md`
2. Commit the skill README (ask user about other new or changed files)
3. Tag it (e.g., `v0.9.3`)
4. Push commit + tag
5. Create GitHub release with `gh release create`
6. Upload zip to the release
