# Init
**Requirements:** R136, R137, R138, R139, R140, R141, R142, R143, R144, R145, R146, R161, R162, R163, R170, R171, R172

The sole creator of `.minispec/config.yaml`, and the only repair path for a `track`
value or a `.gitignore` that disagrees with one.

Two forms, mutually exclusive on one precondition: plain `init` requires `.minispec/`
to be **absent**, `--repair` requires it to be **present**. Neither has to guess the
caller's intent, and there is no mode flag to reconcile — the precondition is
inverted, not added to.

## Knows
- repoRoot: where `.minispec/` and the top-level `.gitignore` live
- trackValue: the mandatory `--track-<style>` choice, never inferred and never
  defaulted
- repair: which of the two forms was invoked
- changed: every file created or edited, accumulated so the report can be complete

## Does
- Create(): write `<repoRoot>/.minispec/config.yaml` holding the chosen `track` and
  nothing it does not need — a basic file, not a commented template of every setting
- Repair(): set `track` on an existing configuration *and* bring `.gitignore` into
  agreement with it. Symmetric, which is what lets one command resolve a mismatch in
  either direction: unignored trajectory files under `private-trajectory` can be fixed
  by adding the ignore lines or by declaring the files public, and the flag value is
  how the user chooses between those opposite meanings
- reconcileIgnore(): add `.minispec/backup` under any git-managed value, and the
  trajectory filenames under `private-trajectory`; remove what the value forbids.
  Skipped entirely when git is absent
- ignoreLine(path): render a governed path as an **anchored** pattern, `/PENDING.md`.
  Every path `track` governs is mandated at the repository root, so a bare pattern
  reaches nested paths too — and since the last match wins, appending the bare form to
  a file that already had the anchored one *widens* the project's rule rather than
  duplicating it

- Report(): print every file created or edited, in full, and name any path it could not
  make public. An agent that mutates state through a tool and cannot see the result will
  reason from a stale picture and eventually assert it
- refuseExisting(): plain `init`'s refusal when `.minispec/` is already there, naming
  `--repair` as the way to change a value and saying to confirm with the user first
- validateWellFormed(): `--repair`'s precondition. A configuration that fails to parse
  cannot be repaired by a flag, because the damage is arbitrary

**Coverage is asked of git, never matched as text.** `/PENDING.md`, `PENDING.md` and
`*.md` are the same intent written three ways, and only git knows the third counts. The
textual version of this check was written first and shipped a duplicate the first time
`init` ran on a real repository — the same lesson the ignore check itself already
encodes by shelling out instead of parsing the file.

**What repair cannot reach, it names.** This card owns the top-level `.gitignore` and
nothing else, so a path ignored by a nested `.gitignore`, `.git/info/exclude`, or the
user's global excludes stays ignored no matter what is deleted here. It is reported by
name rather than swallowed: a repair that half-worked and announced success is the
silent partial success this layer exists to prevent.

**The malformed-configuration path is the one place the agent may edit the config.**
`--repair` sets a value; it cannot fix a stray character typed while the file was open
in an editor. So that refusal states every problem found, points at the skill's
configuration documentation, and **explicitly authorises the edit** — at that point
the agent is the only actor left who can act. The agent backs the file up first, and
skips the backup when it would be byte-identical to one already there.

**Why the mandatory flag.** Whether a repository *is* git-managed is checkable after
the fact; whether you want your queue to ship is not inferable at all, and a default
would make that choice for someone without their noticing.

**Why nothing runs this implicitly.** The tool never invokes `init` itself and no
skill instructs an agent to run it, so the only paths that reach it are a direct
request from the user or the no-configuration refusal — whose entire purpose is
obtaining exactly that assent. This is an absence rather than a prohibition: a
prohibition is text that can be misread or argued around, it needs maintaining, and it
would make the agent refuse a legitimate direct request, which is worse behavior than
running the command.

## Collaborators
- Git: asked whether the tree is git-managed, which decides whether ignore lines are
  written at all
- Track: supplies the closed value set and the paths a value requires to be ignored
- Project: supplies the repository root and the configuration path beneath it

## Sequences
- seq-bootstrap.md
