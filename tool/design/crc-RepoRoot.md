# RepoRoot
**Requirements:** R107, R108, R109, R110, R111, R112, R113, R114

Resolves the **repository root** — the top of the version-controlled working tree,
which owns `.claude/`, `carves/` and the trajectory files.

Deliberately separate from `Project`, which resolves the **design root** by walking
up for `design/`. The two searches share nothing but the direction of travel: they
look for different markers, answer different questions, and a repository may hold
several design roots beneath one repository root. Merging them is the ambiguity this
card exists to remove.

## Knows
- startDir: the absolute directory the search begins from
- homeBoundary: the user's home directory — the search never reaches it or above
- strongMarkers: `.git`, `.minispec/`, `carves/`, and the trajectory filenames
  (`PENDING.md`, `CURRENT.md`, `DONE.md`) — equal, unordered
- claudeCandidate: deepest `.claude` directory seen so far, if any
- yamlCandidate: deepest `.minispec.yaml` seen so far, if any

## Does
- RepoRoot(): walk upward from the current directory, bounded by the user's home
  directory, and return the repository root — or an error naming the search origin
  when no marker is found
- RepoRootFrom(startDir, homeBoundary): the same walk from an explicit origin and an
  explicit boundary. Both are parameters rather than environment lookups, which is
  what makes the home exclusion testable on a temporary tree
- hasStrongMarker(dir): report whether a directory carries any strong marker
- atBoundary(dir, home): report whether a directory is the home directory or an
  ancestor of it

**Why weak candidates are recorded rather than returned on sight.** The first strong
marker met going up *is* the deepest, so the search returns there immediately. A weak
marker cannot be trusted the same way: a `.git` may sit above the first `.claude`,
and `.claude` exists at three levels of an ordinary path — the repository, its
parent, and the home directory. So weak markers are noted and the walk continues;
only exhausting the walk promotes one.

## Collaborators
- os/filepath: for path operations and parent traversal
- os: for directory and file existence checks, and the user's home directory

## Sequences
- seq-reporoot.md
