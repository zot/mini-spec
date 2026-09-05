# Sequence: the backup slot

Numbered diagrams — the swap every operation shares, a mutation, revert, and the worktree
anchor — numbered so code can pin to a step. *The lead-in names them rather than counting
them, per `O26`: a count over an enumeration silently becomes wrong when the enumeration
grows, and this one did.*

# Sequence 1: the swap

Every operation is this. Change, revert and replay differ only in what step 1.2 does.

```
1. Backup -> Backup: swap(perform)

   1.1 copy the covered live files to a temporary location
       the three trajectory files; a carve is never among them        // R250

   1.2 perform the operation
       a mutation writes; revert and replay restore from the backups

   1.3 write the stamp
       state only — `changed`, `reverted` or `replayed`               // R254, R256

   1.4 move the temporary copies over the old backups                 // R252
       **Move, not copy, and last.** A move within a filesystem is atomic, so a
       crash leaves either the old backup intact or the new one complete — never a
       half-written backup. The same reason one writes a temp file and renames
       rather than truncating in place. Do not simplify this into a direct write.

   1.5 write the worktree anchor                                      // diagram 4, R266
       **RUNS FIRST — before 1.1.** The number is an identifier, not the order.
       Sequence IDs are append-only for the same reason `Rn` is, so a step that
       belongs at the front still takes the next free number; renumbering 1.1-1.4
       to make room would orphan every pin at them. First instance in this project
       of that rule forcing an out-of-order step, and it is cheaper than the
       alternative it prevents.

       Here rather than in diagrams 2 and 3 because the swap is the one operation
       every transition shares, so one call covers `changed`, `reverted` and
       `replayed` and no future path can forget it.                   // R266
```

# Sequence 2: a mutation

```
2. CLI -> Backup: Record()

   2.1 Backup -> Backup: swap(write the caller's change)              // diagram 1
       Resets the slot from **any** state. Whatever was revertable is gone, without
       ceremony — the whole memory model, and why no history is needed.  // R257

   2.2 Backup -> Backup: release the previous attempt, if the slot held one
       2.2.1 Carve: the part returns to `**OPEN (not queued.)**`
             aborting an attempt is not aborting the part              // R259
       2.2.2 the number returns to the pool
       2.2.3 no done entry — the done file is the completion ledger, and an
             unfinished attempt reconstructs nothing                   // R259

   2.3 CLI -> CLI: crank out what changed, in full                     // R263
       and, when a returned number is vended again, announce the reuse // R260
```

# Sequence 3: revert, and replay by the same path

```
3. CLI -> Backup: Revert()

   3.1 Backup -> Backup: legal(revert)?
       From every state exactly one of revert / replay is legal, so there is never
       a choice to disambiguate. Illegal -> refuse, naming the state the slot is in
       and what it will accept.                                        // R256

   3.2 Backup -> Backup: drift()
       mtime of each covered file against the **single** stamp — no content hashing,
       no per-file snapshots. One stamp is also what makes the check atomic across
       the files at once.                                              // R255

       3.2.1 any file changed since the stamp -> **refuse**
             name which files changed and where their backups are. The operator is
             better placed than the tool to reconcile a hand edit with a pending
             revert, and a revert that silently clobbers a later edit is worse than
             no revert.

   3.3 Backup -> Backup: swap(restore the covered files from the backups) // diagram 1
       One backup set serves both directions, because it always holds *the other
       state*.                                                         // R253

   3.4 Backup -> Carve: set the part's marker                          // R258
       revert -> `**REVERTED (#N.)**`; replay -> `**OPEN (#N.)**`
       The queue rolls **backward** while the carve moves **forward**, so this is a
       write and never a restore. The item number comes from the backed-up pending
       file, which is why the stamp needs no field for it.             // R254

   3.5 ~~Backup -> Git: refresh the working-tree snapshot~~           // was R261
       **MOVED into the swap (1.5), 2026-08-17.** Nothing happens at this step.
       Kept as a numbered step because sequence IDs are permanent — deleting it
       would renumber 3.6 and orphan the pin at it.

       Why it moved: as drawn it ran *after* 3.4, which writes a marker into a
       **tracked** carve — so the anchor swallowed the tool's own write. The
       anchor must precede everything a transition changes, and the only place
       that is true for all three transitions is the swap.            // R266

   3.6 CLI -> CLI: crank out every change, and where the anchor is        // R262, R272
       it holds no ignored file — the trajectory files are the slot's business —
       and it is deliberately outside the stash listing, so this is the only
       place it surfaces: name the ref *and* the commands to restore a file at
       either reference point, rather than a location the reader must decode
```

# Sequence 4: writing the worktree anchor

Reached from 1.5, so every transition takes this path and none can skip it.

```
4. Backup -> Git: write the anchor

   4.1 build a tree from a scratch index                              // R267, R268, R269
       `GIT_INDEX_FILE=<tmp>` then add everything and write the tree. Three
       properties, and each is structural rather than a rule anyone must keep:
       untracked file **contents** are in it, **ignored** paths are not, and the
       real index and working tree are never touched.

       4.1.1 the obvious command is the wrong one, and one word away
             the familiar stash verb *relocates* the changes, which is
             destructive and exactly what must not happen to work in progress
       4.1.2 the flag that looks like the answer is silently ignored
             `stash create -u` exits 0, returns a plausible object, and holds
             no untracked file. Measured 2026-08-17. A check asking only
             "did it error?" reports it working                       // R269

   4.2 commit that tree with the current commit as first parent       // R271
       so the commit it was taken from is recoverable from the anchor itself,
       and no field of ours can drift out of step with the tree it describes

   4.3 point the tool's own ref at that commit, replacing what it held  // R266, R270
       One act, so exactly one anchor exists and no find-then-drop is needed.

       4.3.1 outside the stash on purpose
             an entry in the stash can be applied by someone reaching for their
             own work, and clearing the stack destroys it at any position —
             there is no place *within* the stash that a routine cleanup spares

   4.4 nothing is restored, moved or staged                           // R273
       reference, not undo. Reporting a difference destroys nothing, which is
       what makes it safe to take over work the tool has no business owning;
       undoing one is a different act with a different risk. No path may slide
       from the first into the second.
```
