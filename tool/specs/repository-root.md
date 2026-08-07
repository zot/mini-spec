# Repository Root

## Two roots, and why the distinction matters

"Project root" names two different directories, and conflating them is what this
feature exists to fix.

- The **design root** is the directory containing `design/`. It owns `specs/`,
  `design/`, and `src/`, and it is what the tool detects today and reports as
  `root:`.
- The **repository root** is the top of the version-controlled working tree. It
  owns `.claude/`, `carves/`, and the trajectory files.

A repository holds **one** work queue and may hold **several** design roots. This
repository is such a case: `tool/` and `example/` each have their own `design/`,
`src/` and `specs/`, while the queue and `carves/` sit above both. Run from the top,
the tool resolves no design root at all.

So repo-scoped artifacts do not fit design-root detection, and the tool needs both
notions rather than one.

## Finding the repository root

Search upward from the current directory, collecting markers as you go, and never
consider the user's home directory or anything above it.

**Strong markers — the deepest directory carrying any of these is the repository
root:**

- `.git`
- `.minispec/`
- `carves/`
- a trajectory file (`PENDING.md`, `CURRENT.md`, or `DONE.md`)

All four are repository-level by mandate or by definition, so where they coexist
they agree and no ordering among them is needed.

**Weaker fallbacks, in order, used only when no strong marker is found:**

1. the deepest `.claude` directory
2. the deepest `.minispec.yaml`

**Otherwise the search fails**, reporting the directory it started from.

Only `.git` counts as a version-control marker. A tree managed by something else
falls through to the next marker down, which is the right answer there rather than a
gap.

## Why the search collects markers instead of stopping at the first

A walk that decided at each level could not know whether a stronger marker sat
above it, so it would have to accept the first weak one it met. Collecting lets the
strong marker win from any depth while the weak ones stay available if it never
appears.

This is not hypothetical. `.claude` exists at three levels of one path on the
maintainer's machine — the repository, its parent, and the home directory — so a
first-match walk accepting `.claude` would resolve a directory that is not a
repository at all as the repository root.

`.minispec.yaml` is likewise the design-root marker, and a repository may hold
several. The innermost wins any upward walk, so "the first `.minispec.yaml` going
up" would resolve this repository's root to `tool/`. Presence is not a declaration,
which is why it ranks last.

## The home-directory exclusion

The search stops before the user's home directory and never considers it or its
parents. `~/.claude` exists on a normal installation, so without the exclusion any
tree carrying no markers of its own would resolve its repository root to the user's
home directory — and anything the tool later created would be deposited there.

## Reporting

`query project` reports the repository root alongside the design root, so the two
are visible and distinguishable rather than merged under one label. When the two are
the same directory — the common case, and the reason the distinction stayed hidden
for so long — it says so rather than printing the path twice without comment.

## Version checking uses the repository root

`minispec check-version` compares the tool's version against the mini-spec skill's
`README.md`. It looks for that file under the **repository root** first, then under
the user's home directory.

This replaces looking in the current directory first, which is the wrong place to
ask because `.claude/` is repository-scoped: run from anywhere below the top and
there is no local `.claude/`, so the lookup falls silently through to the user-level
skill. The check still prints a verdict, so the mistake is invisible.

**The case that makes this load-bearing is a project that vendors its own copy of
the skill**, keeping the artifacts and the skill together in one repository so the
version is pinned. Under the old lookup, running from any subdirectory of such a
project consulted the *user-level* skill instead and reported success against it —
defeating the pin the vendored copy exists to enforce, and reporting green while
doing so. Measured 2026-08-07 with a project pinned at 2.10.0 against a user-level
skill at 2.11.0: the check printed `ok: tool and skill both at 2.11.0`.

## Out of scope

Detection failure is a plain error naming the search origin. Turning that into a
crank handle that asks the user to confirm the root — and recording the answer — is
part of `init`, which cannot run before the configuration feature exists.
