# Track
**Requirements:** R131, R132, R133, R134, R135, R147, R148, R149, R150, R151, R173

Owns the `track` setting — the one thing a tool cannot infer about a repository —
and verifies it against what git actually reports, on every non-trivial run.

**`track` is a checked copy, not a stored claim.** It duplicates what `.gitignore`
already says, deliberately: a copy re-verified every run cannot drift undetected,
which was the whole of the objection to keeping it. A stored claim rots; a checked one
is a checksum.

## Knows
- Value: one of `none`, `private-trajectory`, `all` — a closed set, so an unrecognised
  value is a malformed configuration rather than a fourth behavior
- GovernedPaths: `PENDING.md`, `CURRENT.md`, `DONE.md` at the repository root, plus
  `.carves/` when it exists. Mandated locations, not configuration — the trajectory
  is repository-scoped because it crosses design roots

## Does
- Parse(string): resolve a configuration value to a Value, rejecting anything outside
  the closed set
- Load(cfgPath): read the declared value, distinguishing a configuration that sets no
  `track` from one that sets an unrecognised one. Both refuse, and they refuse
  *differently* — see below (R173)
- ExpectedIgnored(repoRoot): the paths this value requires git to ignore. Under
  `private-trajectory` that is the trajectory files and `.carves/`; under `all` it is
  `.carves/` alone, since a directory whose whole purpose is privacy does not become
  public because the queue did; under `none` there is nothing to expect
- Verify(git, repoRoot): compare the declared value against two facts — whether the
  tree is git-managed, and the actual ignore state of the governed paths — and return
  the disagreements. `none` requires that git be absent; the other two require that it
  be present
- Mismatch.Report(): the crank handle. Names which fact disagrees and gives
  `minispec init --track-<style> --repair` as the repair, with the instruction to
  confirm the value with the user first

**Why the gripe repeats.** It fires on every run until repaired rather than once per
session, because a one-shot reminder decays to nothing — measured on this project's
own sibling repository, where the single category with no standing check sat at 19%
stale while every checked category sat at zero. It cannot become wallpaper either,
because it is closable: repair once and it is gone.

**Absence is a version difference; a wrong value is damage (R173).** A configuration
with no `track` is one written before the setting existed, and the flag supplies exactly
what is missing — so it is repairable, and `Load` says so. A configuration with a value
outside the closed set was edited by a hand that may have changed other things, so no
flag can be trusted to fix it. Init's well-formedness check has always drawn the line
here; this card's job is to report the two cases apart so the gate can draw it in the
same place, instead of refusing a repair the repair path would have accepted.

**Why a design root cannot override it.** `track` describes the repository, and a
repository may hold several design roots. Letting one of them answer for the whole
would be the two-roots confusion this tool has already paid for once.

**Silence with no git.** In a tree with no repository at all, neither the check nor the
preferences mean anything, so nothing is said. An instruction naming a file the
project does not have is worse than no instruction.

## Collaborators
- Git: supplies the two facts the declared value is checked against
- Project: holds the resolved configuration the value is read from
- Init: the only writer of the value, and the repair path this card's message names

## Sequences
- seq-bootstrap.md
