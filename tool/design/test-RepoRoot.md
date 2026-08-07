# Test Design: RepoRoot
**Source:** crc-RepoRoot.md

Every case builds a temporary directory tree and calls `RepoRootFrom(startDir,
homeBoundary)`. No
fixtures, no repository, no network — the resolver is pure decision logic over a
filesystem shape, so the expensive-to-reach paths are avoided entirely and there is
no excuse for leaving it hand-checked.

The home boundary is injected rather than read from the environment, so a test can
place it anywhere in the temporary tree and exercise the exclusion without depending
on the machine it runs on.

## Test: each strong marker resolves on its own
**Purpose:** validates that `.git`, `.minispec/`, `carves/`, and each trajectory
filename are all sufficient, and that none is privileged over the others
**Input:** four trees, each `root/a/b/` with exactly one strong marker at `root`;
start from `root/a/b`
**Expected:** every case returns `root`
**Refs:** crc-RepoRoot.md, seq-reporoot.md#1.3.2 — R109

## Test: coexisting strong markers agree
**Purpose:** validates that no ordering among strong markers is needed
**Input:** `root` holding `.git`, `.minispec/` and `carves/` together; start from
`root/a/b`
**Expected:** returns `root`
**Refs:** seq-reporoot.md#1.3.2 — R109

## Test: a strong marker above beats a weak marker below
**Purpose:** the property that makes collecting necessary — the regression a
first-match walk would introduce
**Input:** `.git` at `root`, `.claude` at `root/a`; start from `root/a/b`
**Expected:** returns `root`, not `root/a`
**Refs:** seq-reporoot.md#1.3.3 — R108

## Test: deepest strong marker wins over a shallower one
**Purpose:** validates that the walk returns on first strong marker met, which is the
deepest
**Input:** `.git` at `root`, `carves/` at `root/a`; start from `root/a/b`
**Expected:** returns `root/a`
**Refs:** seq-reporoot.md#1.3.2 — R109

## Test: .claude is used only when no strong marker exists
**Purpose:** validates the first weak fallback
**Input:** `.claude` at `root`, nothing stronger anywhere; start from `root/a/b`
**Expected:** returns `root`
**Refs:** seq-reporoot.md#1.4 — R110

## Test: deepest .claude wins among several
**Purpose:** validates that the recorded candidate is the deepest, not the last seen
**Input:** `.claude` at `root` and at `root/a`; start from `root/a/b`
**Expected:** returns `root/a`
**Refs:** seq-reporoot.md#1.3.3 — R110

## Test: .minispec.yaml is the last resort
**Purpose:** validates the second weak fallback and its rank below `.claude`
**Input:** (a) `.minispec.yaml` at `root` alone; (b) `.minispec.yaml` at `root/a` and
`.claude` at `root`; start from `root/a/b`
**Expected:** (a) returns `root`; (b) returns `root` — `.claude` outranks
`.minispec.yaml` regardless of depth
**Refs:** seq-reporoot.md#1.5 — R111

## Test: the home boundary is never crossed
**Purpose:** validates the exclusion that keeps created files out of the user's home
directory
**Input:** home boundary set to `root`; `.claude` and `.git` placed at `root`; start
from `root/a/b` with no markers below
**Expected:** the search fails; it does not return `root`
**Refs:** seq-reporoot.md#1.3.1 — R112

## Test: failure names the starting directory
**Purpose:** validates that absence is reported as an error carrying the search
origin, rather than as a silent or bare failure
**Input:** a marker-free tree; start from `root/a/b`
**Expected:** an error whose message contains `root/a/b`
**Refs:** seq-reporoot.md#1.6 — R113

## Test: a non-git VCS marker does not resolve
**Purpose:** validates that only `.git` counts, and that another VCS falls through
rather than being detected
**Input:** `.fslckout` at `root/a`, `.claude` at `root`; start from `root/a/b`
**Expected:** returns `root` via the `.claude` fallback, not `root/a`
**Refs:** crc-RepoRoot.md — R114

## Test: the two roots are resolved independently
**Purpose:** validates the distinction on the shape that motivated it — one
repository, several design roots
**Input:** `.git` and `carves/` at `root`; `design/` at `root/tool` and at
`root/example`; start from `root/tool`
**Expected:** repository root is `root`; design root is `root/tool`
**Refs:** crc-RepoRoot.md, crc-Project.md — R107
