# Trajectory
**Requirements:** R190, R194, R195, R197

Reads the trajectory files at the repository root. The format is **owned by the skill**
— `trajectory-format.md` is normative for every shape — so this card names which shapes
it consumes and never redefines them.

The first component in the tool to read trajectory content at all. Until now the files
were only ever a root marker and an ignore-state question, both answerable without
opening them.

## Knows
- TrajectoryScan: what each queue file contributed — its name, whether it was
  **present**, and the IDs found in it. Presence is kept apart from the ID count because
  they answer different questions, and collapsing them is how a broken regex comes to
  look like an empty queue

## Does
- ScanTrajectory(repoRoot): read both files and report them **per file** rather than
  merged, so the caller can say what each contributed (R197). A package function, not a
  method — it touches no design-root state, and a signature claiming otherwise is what
  made the command refuse to run in this repository
- MaxItemID(): the maximum across **both** files. Either alone collides — the pending
  file's maximum is too low right after items complete, the done file's while the
  highest IDs are still live (R190)
- Missing(), AnyPresent(): name the files that were absent, and distinguish *no files at
  all* — which has no answer, so ErrNoTrajectoryFiles — from *one file missing*, which
  has an answer worth qualifying (R194, R195)

## Shapes consumed
Named here so the coupling is visible; `trajectory-format.md` defines them.

- **pending file** — an item entry is a `##` heading opening with the number:
  `## 8. **<title>** …`
- **done file** — an entry's **header** leads with the identifiers it discharged, in the
  run between the date's em dash and the colon that opens the title:
  `` - **YYYY-MM-DD — #8 / R189–R198: <title>.** (`<commit>`) Part `<doc>#<key>`. ``
  The slot may hold a gap ID or a requirement range instead, or several separated by
  `/`, so every `#N` in it counts — and only what is in it does.

**Only an entry's header is scanned**, deliberately. Done entries are prose several
lines long and quote other items freely: measured 2026-08-16, **five** body lines in
ark's ledger would contribute a queue ID if the header were not required, most of them
older pending entries quoted verbatim inside a later entry's body. Letting one in is the
same class of error as reading one file instead of two — a plausible number that is
wrong.

## Collaborators
- Project: to resolve the repository root

## Sequences
- seq-query.md
