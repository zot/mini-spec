# Test Design: Backup
**Source:** crc-Backup.md

Every alarm below targets a property that **passes by default**: a slot that never checked
drift, copied instead of moving, or collapsed two states into one satisfies any test written
without these fixtures. Re-derived 2026-09-04 from `old-sdom`'s design over the reclaimed
adapters; every alarm is a prescription until pulled here.

## Test: drift refuses rather than clobbering
**Purpose:** validates R227 — revert and replay refuse when a covered file changed since the stamp, and leave the live files alone
**Input:** a slot in `changed`; one trajectory file touched after the stamp
**Expected:** a `DriftError` naming the file and the backup directory; the hand edit intact
**Fire alarm:** skip the drift check and revert unconditionally; goes red on the untouched-file assertion, which is the one that matters
**Inject:** internal/backup/backup.go:toggle
**Refs:** crc-Backup.md — R227
**Pulled:** 2026-09-04 — rang: `err = <nil>, want a DriftError`; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: the backup is moved into place, not written into place
**Purpose:** validates R224 — a failed operation leaves the previous backup intact
**Input:** a mutation whose `perform` fails after the copies are staged
**Expected:** the backup directory still holds the previous complete backup
**Fire alarm:** replace the temp-then-move with a direct write to the backup path before `perform`; goes red on the backup-contents assertion. This is the alarm most likely to be "simplified" away
**Inject:** internal/backup/backup.go:swap
**Refs:** crc-Backup.md — R224
**Pulled:** 2026-09-04 — rang: `a failed operation replaced the backup`, and three restore tests beside it since the backup already held the post-mutation bytes; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: exactly one of revert and replay is legal from each state
**Purpose:** validates R228
**Input:** each of the three states offered both operations
**Expected:** three legal, three refused, each refusal naming the state and what it accepts
**Fire alarm:** allow revert from `reverted`; goes red on the refusal
**Inject:** internal/backup/backup.go:legal
**Refs:** crc-Backup.md — R228
**Pulled:** 2026-09-04 — rang: `revert from reverted: err = <nil>, want IllegalError`, and four tests refusing replay; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: `changed` and `replayed` are distinct states
**Purpose:** validates R228 — a fresh mutation from `replayed` stamps `changed`, never `replayed`
**Input:** mutation, revert, replay, then a fresh mutation
**Expected:** the stamp reads `changed`
**Fire alarm:** stamp `Replayed` when the prior state was `Replayed`; goes red — a stamp that lies
**Inject:** internal/backup/backup.go:Record
**Pulled:** 2026-09-04 — rang: `state after a fresh mutation from replayed = replayed, want changed`; restore byte-clean by copy
**Refs:** crc-Backup.md — R228
**Code:** internal/backup/backup_test.go

## Test: a new mutation discards what was revertable
**Purpose:** validates R229, the goldfish rule
**Input:** mutation, revert, second mutation, then an attempted replay
**Expected:** the replay is refused; the backup holds the state before the *second* mutation
**Fire alarm:** skip the swap when the slot is `reverted`, preserving the prior backup; goes red on the replay refusal — the beginning of an undo stack
**Inject:** internal/backup/backup.go:Record
**Pulled:** 2026-09-04 — rang: `backup = "…second", want the state before the second mutation ("…third")`, and the release test beside it; restore byte-clean by copy
**Refs:** crc-Backup.md — R229
**Code:** internal/backup/backup_test.go

## Test: a carve is written, never restored
**Purpose:** validates R222, R230 — the queue rolls backward while the carve moves forward, so the part's number survives a revert
**Input:** a mutation writing the pending file and marking a carve part `OPEN (#16.)`; then a revert; then a replay
**Expected:** the pending file restored; the carve not restored and reading `**REVERTED (#16.)**`; after replay `**OPEN (#16.)**`
**Fire alarm:** skip the marker write at the end of `toggle`; goes red on the revert-trace assertion. *The injection first written here — add the carve to the covered set — cannot reach the property in this code:* the covered names are flat files and the staging directory has no `carves/` in it, so it crashes on the copy (and, with the directory created, on the rename) before any assertion runs. A crash reads like a red test and proves nothing
**Inject:** internal/backup/backup.go:toggle
**Refs:** crc-Backup.md, crc-Carve.md — R222, R230
**Pulled:** 2026-09-04 — rang: `the carve does not carry the revert trace`; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: an ended attempt releases the number and reopens the part
**Purpose:** validates R231, R234 — aborting an attempt is not aborting the part; no done entry
**Input:** mutation, revert, then a second mutation ending the replayable window
**Expected:** `Released()` names the ID before the mutation; afterwards the part reads `**OPEN (not queued.)**`, the done file is unchanged, and nothing is left to release
**Fire alarm:** write a done entry for the ended attempt in `releaseAttempt`; goes red on the unchanged-ledger assertion — proposed and rejected during design
**Inject:** internal/backup/backup.go:releaseAttempt
**Refs:** crc-Backup.md — R231, R234
**Pulled:** 2026-09-04 — rang: `an unfinished attempt was written to the completion ledger`; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: a live item is not released
**Purpose:** validates R231 — from `changed` or `replayed` the entry is live, so a mutation abandons nothing
**Input:** a live entry; `Released()` from `changed`, and again after revert and replay
**Expected:** nothing released either time
**Fire alarm:** reverse the diff direction in `Released`; goes red by naming a live number
**Inject:** internal/backup/backup.go:Released
**Refs:** crc-Backup.md — R231, R234
**Pulled:** 2026-09-04 — rang: `Released() from changed = [16], want none` and the mirror from replayed; the ended-attempt test lost its `[16]`; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: a restored file is not read as a hand edit
**Purpose:** validates R227 — the drift check must not fire on the tool's own restore
**Input:** mutation, revert, replay
**Expected:** the replay proceeds with no drift reported
**Fire alarm:** normalise a copied file's modification time with `Chtimes(time.Now())` in `copyFile`; goes red with a `DriftError` on the replay. The bug was written once: the kernel stamps ordinary writes from a coarse clock and `Chtimes` from a fine one, so the restore lands milliseconds ahead of the stamp
**Inject:** internal/backup/backup.go:copyFile
**Refs:** crc-Backup.md — R227
**Pulled:** 2026-09-04 — rang: `replay read the restore as a hand edit: refusing: PENDING.md changed since the backup was taken`, and four more replays refused — the two-clocks bug, reproduced; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: a completed part is not reopened by the next mutation
**Purpose:** validates R232, R233 — a completion and an abandoned attempt produce the same entry diff
**Input:** a slot in `changed` whose backup holds an entry the live file lacks and whose part is `[x]` and `LANDED`; then any mutation
**Expected:** the part keeps `[x]` and its `LANDED` record and never reads `OPEN (not queued.)`
**Fire alarm:** neuter both guards — release from any state and drop the landed check; goes red. Either alone is insufficient by design. *Since 2026-09-04 a third guard stands beneath both:* the dependency's `SetMarker` refuses `OPEN` over a checked part (`ErrReopen`), so with both of this package's guards gone the write surfaces as that error rather than as a reopened part
**Inject:** internal/backup/backup.go:Record, internal/backup/backup.go:mark
**Refs:** crc-Backup.md — R232, R233
**Pulled:** 2026-09-04 — rang: `minispecsdom: the part is landed; a write may not reopen it` — the dependency's third guard caught what this package's two no longer did; the hand-removed-entry test went red on its own assertion; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: a hand-removed entry is not an abandoned attempt
**Purpose:** validates R232 alone — the tool does not infer an abandoned attempt from a diff it did not cause
**Input:** two mutations so the backup holds the first entry; a hand edit removing that entry while the part stays `[ ]`; a third mutation
**Expected:** the part keeps `OPEN (#N.)` rather than becoming `OPEN (not queued.)`
**Fire alarm:** make the state test stop discriminating (`present || cur == Reverted`); goes red — the part is `[ ]`, so the landed guard has nothing to refuse
**Inject:** internal/backup/backup.go:Record
**Pulled:** 2026-09-04 — rang: `a diff the tool did not cause was read as an abandoned attempt`; restore byte-clean by copy
**Refs:** crc-Backup.md — R232
**Code:** internal/backup/backup_test.go

## Test: the swap anchors the worktree before the transition
**Purpose:** validates R236 — the wiring, which fails silently: every anchor test keeps passing if the call in `swap` is deleted
**Input:** a real repository holding a tracked non-trajectory file; one `Record` rewriting it
**Expected:** the anchor holds the **pre**-mutation content
**Fire alarm:** remove the `Snapshot()` call from `swap`; goes red with `invalid object name 'refs/minispec/snapshot'`
**Inject:** internal/backup/backup.go:swap
**Refs:** crc-Backup.md, crc-Git.md — R236
**Pulled:** 2026-09-04 — rang: `no anchor was written by the swap: exit status 128`; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: a tree with no git still restores
**Purpose:** validates R239
**Input:** a plain directory, no repository; mutation then revert
**Expected:** the revert succeeds
**Fire alarm:** treat `ErrNoGit` from the anchor as fatal in `swap`; goes red
**Inject:** internal/backup/backup.go:swap
**Refs:** crc-Backup.md — R239
**Pulled:** 2026-09-04 — rang: `not a git working tree` from every slot test, twelve of them; restore byte-clean by copy
**Code:** internal/backup/backup_test.go

## Test: the anchor touches neither the working tree nor the real index
**Purpose:** validates R237 — nothing is moved, staged or restored
**Input:** an uncommitted modification; `git status --porcelain` before and after
**Expected:** the modification present and the status byte-identical
**Fire alarm:** drop `GIT_INDEX_FILE` from `scratchTree`; goes red on the status comparison
**Inject:** internal/project/git.go:scratchTree
**Pulled:** 2026-09-04 — rang: `the anchor changed index or worktree state` — `seed.txt` came back staged; restore byte-clean by copy
**Refs:** crc-Git.md — R237
**Code:** internal/project/git_test.go

## Test: the anchor holds untracked contents and no ignored paths
**Purpose:** validates R237 — the division of labour with the slot
**Input:** an ignored file and an untracked `new-carve.md`
**Expected:** `ls-tree` names the untracked file and not the ignored one; `show` returns its contents
**Fire alarm:** `add -A` → `add -u` in `scratchTree`; goes red on the untracked file
**Inject:** internal/project/git.go:scratchTree
**Pulled:** 2026-09-04 — rang: `untracked contents are absent from the anchor: []`, and with nothing tracked-and-changed the tree was empty, so three anchor tests and the swap test failed on `exit status 128`; restore byte-clean by copy
**Refs:** crc-Git.md — R237
**Code:** internal/project/git_test.go

## Test: exactly one anchor exists
**Purpose:** validates R236 — writing the anchor replaces what it held
**Input:** two snapshots over a file whose content changed between them
**Expected:** the ref resolves to the second tree
**Fire alarm:** write the second anchor to a different ref name; goes red
**Inject:** internal/project/git.go:Snapshot
**Pulled:** 2026-09-04 — rang: the ref never existed under its own name, so every reader of it failed with `exit status 128`, five tests; restore byte-clean by copy
**Refs:** crc-Git.md — R236
**Code:** internal/project/git_test.go

## Test: the anchor is written in a repository with no commits
**Purpose:** validates R238 — the `err == nil` guard on `rev-parse HEAD` is the half that looks droppable: in an empty repository it exits 128 *and prints `HEAD`*
**Input:** `git init` with no commit and one untracked file
**Expected:** the anchor exists, holds the file, and `^1` does not resolve
**Fire alarm:** drop the `err == nil` guard, keeping only `h != ""`; goes red with `git commit-tree: exit status 128`
**Inject:** internal/project/git.go:Snapshot
**Pulled:** 2026-09-04 — rang: `no anchor in a repository with no commits: git failed: git commit-tree: exit status 128` — `-p HEAD` passed on the printed word; restore byte-clean by copy
**Refs:** crc-Git.md — R238
**Code:** internal/project/git_test.go

## Test: the anchor carries its base commit as first parent
**Purpose:** validates R238
**Input:** a repository with a commit; HEAD captured, then a snapshot
**Expected:** `<ref>^1` is the commit that was checked out
**Fire alarm:** build the commit with no parent; goes red because `^1` does not resolve
**Inject:** internal/project/git.go:Snapshot
**Pulled:** 2026-09-04 — rang: `the anchor has no first parent, so what it was taken from is unrecoverable: exit status 128`; restore byte-clean by copy
**Refs:** crc-Git.md — R238
**Code:** internal/project/git_test.go

## Test: pending entries read through the dependency
**Purpose:** validates R240 — the adapter reads ID, source document, source key and kind from `minispecsdom.Pending`, and a gap-sourced entry reads as such
**Input:** a pending file with a part-sourced and a gap-sourced entry
**Expected:** two entries with the expected IDs, keys and kinds
**Fire alarm:** none — the reading is the dependency's; this pins the adapter's field mapping
**Refs:** crc-Trajectory.md — R240
**Code:** internal/parser/trajectory_test.go
