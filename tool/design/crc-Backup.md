# Backup
**Requirements:** R221, R222, R223, R224, R225, R226, R227, R228, R229, R230, R231, R232, R233, R234, R239, R235

One level of undo and one of redo over the trajectory files. Owns `.minispec/backup/` — the
copies and the stamp together, which is what makes "one stamp, atomic across all the files"
structural rather than implied.

**Scoped to the trajectory files, and that is a decision rather than an increment.** The
`update` verbs stay unprotected and keep writing in place. The pending item is the smallest
unit of work, so it is the right thing to be able to undo; the rest of the working tree is a
different and much larger problem, and it belongs to [Git](crc-Git.md)'s anchor, which reports
rather than restores (R222).

*Reclaimed 2026-09-04 from `old-sdom` over the Item 1 adapters.* The slot's own logic is
unchanged; what moved is beneath it — the carve marker is written through
[Carve](crc-Carve.md)'s `SetMarker(path, key, verb, attribution)`, whose rules are the
dependency's, and the pending entries are read through [Trajectory](crc-Trajectory.md)'s
`PendingEntries`, which is `minispecsdom.Pending` behind a path.

## Knows
- root: the repository root, and nothing about a design root — the layer sits above every
  design root, the same reason `Carves` and `NextItemID` take a repository root
- State: `changed`, `reverted`, `replayed`. **Three, not two.** Two would count
  *configurations*, of which there are indeed two, and conclude `changed` and `replayed` are
  one. They are the same configuration and different **nodes**: `changed` is the entry,
  reachable only by a fresh mutation. Collapsing them makes a fresh mutation report
  `replayed`, which is the tool asserting something untrue (R228)
- covered: the three trajectory files, taken from Project's `TrajectoryFiles` rather than
  restated. A carve is **not** among them and is never restored, which is what lets a part's
  marker survive a revert (R222, R230)

## Does
- Record(perform): a mutation is about to happen — copy the live files aside, let the caller
  write, stamp `changed`, move the copies over the old backups. Resets the slot from **any**
  state and discards whatever was revertable, without ceremony (R224, R229)
- Revert() / Replay(): the same swap in the two directions. One backup set serves both
  because it always holds *the other state* (R225)
- legal(state): from every state **exactly one** of revert / replay is legal; anything else
  refuses with a crank handle naming the state and what it accepts (R228)
- drift(): the covered files modified since the stamp, by **mtime against the single stamp**
  — no hashing, no per-file snapshots; one stamp is what makes the check atomic across the
  files (R227)
- swap(next, perform): anchor the worktree first, then copy live → temp, perform, write the
  stamp, **move** temp over the old backups. Do not simplify the move into a direct write:
  the shorter form passes every test that does not interrupt it (R224, R236)
- mark(before, after, verb, attribution, onlyOpen): the departed entries' parts get their
  marker through Carve's adapter — `REVERTED (#N.)` on revert, `OPEN (#N.)` on replay,
  `OPEN (not queued.)` on release. The item is *derived* from the pending file's two sides,
  which is why the stamp needs no field for it (R226, R230, R231)
- releaseAttempt(): **only from `reverted`** (R232), and **never over a `[x]` part** (R233).
  Two independent reasons not to write, because a completion produces the same entry diff as
  a revert. No done entry is written (R231)
- Released(): the item IDs a new mutation would abandon, answerable *before* the mutation so
  the vend can announce a reuse (R234)

## The diff direction is not enough, and it read as though it were
`Released`'s reasoning once said the release set is *"non-empty only from `reverted`, and the
diff direction is what makes that true rather than a state check"*, so a state guard was
written and then deleted as dead. **It was not dead; it was unreachable** — the injection that
proved nothing changed ran before the completion verb existed. An injection proves only what
it could reach, and a **deletion** justified by a silent injection inherits every gap in that
reach.

## Collaborators
- Project: the repository root, `TrajectoryFiles`, and the `.minispec/backup` path `init`
  already ignores
- Carve: `SetMarker` and `PartIsLanded` over the carve that owns a departed part — the queue
  rolls backward while the carve moves forward, so this is a *write*, never a restore (R230)
- Trajectory: `PendingEntries(path)` on both sides of the swap (R240)
- Git: the worktree anchor at the top of every swap, before anything the transition changes,
  including the slot's own marker writes; a tree with no git skips it (R236, R239)

## Notes
**Why returning the number is not an exception to permanence.** The tool mints and writes in
one act so that "a number is in a document the moment it exists and `max()` over the documents
stays the whole truth." A permanent gap **breaks** that; returning the number **keeps** it. What
permanence forbids is a reused number resolving to different work *with nothing able to detect
it*, and the announcement removes exactly that (R234).

**`/mini-spec` documents this, and revert especially** (R235). The obligation is discharged by
editing `SKILL.md` and is part of this item. The verbs that expose the slot are Item 3's.

## Sequences
- seq-backup.md
