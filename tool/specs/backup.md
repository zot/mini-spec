# The Backup Slot

One level of undo and one of redo over the trajectory files, plus a non-destructive anchor
recording the rest of the working tree for reference. It exists because the tool edits
files that are mostly prose, where a bad write has more to destroy than a flipped
checkbox in `design.md` does.

**Two records, one event.** Every transition writes both: the trajectory files are *copied*
so they can be put back, and the working tree is *anchored* so what happened to it can be
seen. Only the first is undo. The split is what lets the second be taken safely over work
the tool has no business restoring.

**Explicitly not an undo stack** — the 80s `vi` `u` with goldfish memory. The most recent
change is immediately revertable and replayable and nothing older is recoverable. That is
enough for the real emergency, a command that did the wrong thing thirty seconds ago, and
it avoids owning a history the VCS already owns better.

## Scope

**The slot covers the trajectory files only** — the pending, current and done files at the
repository root. It does **not** retrofit onto the `update` verbs, which continue to write
in place with no backup.

The unit is right for the same reason the scope is: a pending item is the smallest unit of
work, so it is the right thing to be able to undo. Tracking the rest of the working tree is
a different and much larger problem, and **the tool never reaches for blanket git resets or
staging** — only safe, verifiable operations.

*A carve is not a trajectory file* — the format sites the three queue files as those and
lists `carves/` separately — so a carve is never *restored*. It is still written, which the
revert section below covers, and that asymmetry is deliberate rather than an oversight.

## The slot

Everything lives in `.minispec/backup/`: the copies and the stamp together, which is what
makes "one stamp, atomic across all the files" structural rather than implied. The
directory is already named by the tool, already written into `.gitignore` by `init`, and
already checked as ignored on every run.

**Every operation is the same swap.** Change, revert and replay each: copy the live files
to a temporary location, perform the operation, write the stamp, then move the temporary
copies over the old backups.

**The move ordering is load-bearing and must not be simplified away.** A move within a
filesystem is atomic, so a crash leaves either the old backup intact or the new one
complete, never a half-written backup — the same reason one writes a temporary file and
renames rather than truncating in place.

**One backup set suffices**, because the backup always holds *the other state*. Revert and
replay are one mechanism with two names, and a second copy for replay would be redundant.

### The stamp, and the drift check

The stamp records the state and nothing else. **It does not record the item ID**, because
it does not have to: the snapshot *is* the pending file, so after a revert the backed-up
copy holds the entry with its number in it.

On revert or replay, the tool verifies that none of the files has changed since the stamp
was written. If any has, it **refuses**, and cranks out which files changed and where their
backups are. The operator is better placed than the tool to reconcile a hand edit with a
pending revert, and a revert that silently clobbers a later edit is worse than no revert.

The check is an mtime comparison against a single stamp — no content hashing, and no
per-file snapshots. That is `make`'s dependency model, and it is right for the same reason
it was then: the cheap check is *sufficient*, and anything cleverer buys accuracy nobody
needs. Comparing against one stamp rather than per file is also what keeps the check atomic
across all the files at once.

### Three states

```
        (any state) --<a new mutation>--> changed        [backup replaced]

        changed  --revert-->  reverted
        reverted --replay-->  replayed
        replayed --revert-->  reverted
```

`changed` is what a fresh mutation leaves behind, where a backup was taken but neither a
revert nor a replay has happened.

**Three and not two.** Two states counts *configurations*, of which there are indeed two,
and concludes `changed` and `replayed` are the same thing. They are the same configuration
and different **nodes**: `changed` is the entry, reachable only by a fresh mutation and
never by toggling. Collapsing them would make a fresh mutation set the state to `replayed`,
which is the tool asserting something untrue — nothing was replayed. A stamp that lies is
worse than a stamp with one more value in it.

**From every state, exactly one of revert / replay is legal.** There is never a choice to
disambiguate and never a "reverting twice" case to define. Anything else is refused with a
crank handle naming the state the slot is in and what it will accept.

**Goldfish, stated exactly:** a new mutation from *any* state resets to `changed` and
replaces the backup. Whatever was previously revertable is gone, without ceremony. That is
the whole memory model, and the reason the slot never needs a history.

## Revert, replay, and the carve

**Revert restores the trajectory files and leaves an audit trace in the carve.** The queue
rolls **backward** while the carve moves **forward** — a revert marks the part
`**REVERTED (#N.)**`, and a replay returns it to `**OPEN (#N.)**`. It is not a file-level
toggle over one set.

The carve write is a **structural edit only** — a marker the tool itself wrote. Nothing
here rewrites prose, and that bound is the scope of the whole mechanism, not a courtesy.

**Amended 2026-08-18 (Bill): the tool may *append* a record it composes; it never removes or
overwrites one.** The rule first read "the tool never writes or removes a record", and the
danger it was written about is overwriting an **assessment** — a `**NOT VERIFIED**` that only
a person can make and no tool can reconstruct. Appending a `LANDED (commit, date — #N.)`
overwrites nothing: every field in it is a fact the tool was handed, and the format asks for
it, since *a part with no marker is indistinguishable from one nobody has considered*.

**And the join depends on it.** A landed part's queue ID lives in that record and nowhere
else — delete the transient without appending one and the carve→queue checks have nothing
left to read. So the narrower rule is what keeps the format, R274's own caution, and the
referential checks all true at once: **never remove, never overwrite; append only what the
tool composed from facts it was given.**

**Since 2026-09-04 the part line is read and written by `github.com/zot/simple-dom`'s
`minispecsdom`**, so the two sections below describe rules that reader owns — its
`part-line.md` and `carve-schema.md` are normative — and this tool reaches them through the
path-taking adapters in `parser` (`SetMarker(path, key, verb, attribution)`, `PartIsLanded`).
They stay here as the decisions they record; the mechanism is the dependency's.

### Flexible on input, rigid on output

*Added 2026-08-18 (Bill).* **The reader accepts case variation in what it is handed; every
marker the tool writes is canonical.** A verb reads the same whether a document says
`**OPEN (#20.)**` or `**open (#20.)**`, and a transient's `not queued.` reads the same in any
casing — while the tool itself writes `OPEN`, `REVERTED` and `not queued.` in one form only.

**The reason is that agents edit these files, and sometimes should.** A carve corrupted by a
bad edit is repaired by hand, and the hand is often an agent's — which is exactly the moment a
verb gets typed in the wrong case. Without the leniency that marker is *invisible to the
writer*: the next queue write appends beside it instead of replacing it, and the line
accumulates two contradictory queue states, which is the failure the write rule exists to
prevent. Rigidity on the way out is what keeps the corpus uniform anyway, so the leniency
costs nothing that the tool's own writes do not immediately repay.

**The boundary, and it is the whole of the rule.** Flexibility covers *sloppiness in the
current format* — case, spacing, an optional full stop. It does **not** cover a superseded
**scheme**: a `#N` key or an `**OPEN, not queued.**` marker is a different format, and those
are **named as unmigrated** rather than quietly absorbed. Accepting variation in how a thing
is written is not the same as accepting a different thing, and collapsing the two is how a
format acquires a second grammar nobody chose.

### The vocabulary is open, so the verb's shape must admit words nobody has coined

*Added 2026-08-21 (Bill's ruling, out of gap `O82`.)* **A marker's verb is one or more
all-capital words joined by spaces or hyphens, and the reader admits any such word rather than a
fixed list** — lowercase appears only inside the attribution's parentheses.
The format's own sentence is the reason — *"the vocabulary is open; the shape is what gets
checked"* — and the classification rule already leans on it: a transient is told from a record
by its **attribution's** shape and never by its verb, precisely so a verb nobody has coined yet
is classified correctly on the day it appears. Ten verbs are tabulated as *in live use*; that
table is a census, not a permitted set, and treating it as one would refuse the words next
month needs.

**Why this needed stating rather than leaving implied.** The reader admitted an interior
**space** — which is what makes `NOT VERIFIED` legal — and refused an interior **hyphen**, so
`RE-CUT`, written into this project's own `carves/carve-stencils.md` on 2026-08-19, was
non-conforming from the day it was written and nothing reported it. A hyphen joins two words
into one verb exactly as the space already does. **Note which direction that drift ran:** the
reader was stricter than the format it enforces, so no document was ever wrong and the rule had
simply never been written down — the inverse of the usual failure, and invisible for it, since
nothing goes red when a document obeys a rule stated nowhere.

**The boundary above is unchanged.** An open vocabulary is not an open *shape*. The verb is
still capitals, spaces and hyphens inside a `**` pair; an attribution is still exactly `#N.` or
`not queued.` for a transient; and a run that does not close on its own line is still refused.
Admitting a new **word** is not admitting a new grammar, which is the same line the paragraph
above draws between sloppiness and a superseded scheme.

### When replay is no longer possible

**The precondition is the reverted state, and stating it is the repair.** A release happens
when a new mutation arrives while the slot holds a **reverted** attempt — an item that was
started and rolled back, whose entry sits in the backup and not in the live file. Only then is
there an attempt to be over.

*This sentence used to be missing, and the omission cost a corrupted carve.* The rule read "a
new mutation ends the replayable window; at that point the attempt is over", which describes
the reverted case and never says so. **A completion produces the identical shape** — it removes
the entry from the live file too — so an implementation reading the rule literally releases
after every completion, and the **first queue mutation after any completion reopens the landed
part**. Measured 2026-08-18 by running `pending finish` on the item repairing an earlier
defect: `carves/trajectory-tool.md` Item 6, `[x]`, struck through, `LANDED`, was marked
`**OPEN (not queued.)**` in a tracked public file. The checkbox-agreement check caught it.

**And a released part is never one the tool has recorded as done.** A release refuses to touch
a part whose checkbox is `[x]`, whatever the entry diff says. This is a **second, independent**
reason not to write, not a restatement of the first: the precondition says *when* a release is
legitimate, and this says what a legitimate one can never target. Either alone leaves the
other's failure reachable, and the pair is what makes "finished work is never reopened" a
property rather than a consequence of state bookkeeping.

**The release runs before the mutation's own marker, and the collision check knows it.** The
common path is `add-item`, `revert`, then `add-item` on the *same* part; until 2026-09-13 the
one-item-per-part check read the `REVERTED (#N.)` marker as a live queue ID and refused, while
an `add-item` on a sibling released the part and the re-add then worked. The check now exempts a
`REVERTED` marker while the slot is in the reverted state — and only then, since outside it the
release has already run and a `REVERTED` marker is stale hand state (R330).

*It costs nothing legitimate, and that is checkable rather than asserted.* On the path a
release exists for — `add-item`, then `revert`, then a new mutation — the part is `[ ]`
throughout, because completion never ran. And `finish` followed by `revert` marks nothing at
all, since the entry departs from the other side of the pair.

With that precondition met:

- the part returns to `**OPEN (not queued.)**` — **aborting an attempt is not aborting the
  part.** The queue item was *an attempt* at it; the part is still open and still to be
  completed.
- the number returns to the pool, and the next vend of it **announces the reuse**.

**No done entry is written.** The done file is the completion ledger — enough to
reconstruct a change without re-reading the code — and an unfinished attempt reconstructs
nothing; putting non-events there costs the file its identity.

**"Unfinished", never "aborted."** The loaded word reads as failure when usually nothing
failed, and a mechanism that sounds like an admission of error is one people avoid.

### Why reuse rather than a permanent gap

This looks like an exception to the rule that a number is never reused. It is the opposite:
it is what **keeps** the older invariant that rule rests on.

The tool mints an ID and writes it in one act precisely so that *"a number is in a document
the moment it exists and `max()` over the documents stays the whole truth."* A permanent gap
breaks that — a number would have been vended that no document holds, so `max()` stops being
the whole truth and the next-free-ID query answers from an incomplete picture. A reverted
number is genuinely not vended any more, so returning it to the pool is what makes the
documents say everything there is to say again.

What the permanence rule actually forbids is a reused number resolving to different work
**with nothing able to detect it**. Undetectability is the whole basis of it, and the
announcement removes exactly that: it reaches the one person who could have carried the
number into a chat, a commit message or a note. It is bounded by the goldfish rule — at most
one number is ever ambiguous, and only until the next mutation.

## The worktree anchor

Separate from the slot, and it does a job the slot cannot: it records the working tree as it
stood **immediately before a transition**, so that afterwards anyone can ask what has moved
since. Exactly one exists at a time, at a fixed location, and **every** transition —
`changed`, `reverted`, `replayed` — overwrites it before making its edit.

**It is reference, not undo.** Reporting a difference destroys nothing, which is what makes
it safe to take on someone else's work in progress. **Undoing that difference is a different
act with a different risk** — the user may have done unrelated work along the way — and
nothing here may slide from the first into the second. The tool never alters the working
tree; it reports, and the agent proposes.

**Nothing is moved out of the working tree to build it, and nothing is staged.** The obvious
command is the wrong one and it is one word away: pushing a stash *relocates* the changes,
which is destructive and exactly what must not happen. The anchor is built by writing a tree
from a scratch index, so the real index and the working tree are untouched.

**It holds untracked file contents, not merely their names.** A record of *which* files were
new can report that something is missing; only the contents can give back a new carve or a
fresh source file, which is what someone actually wants after a catastrophe.

**It excludes ignored paths, and that boundary is structural rather than a rule to remember.**
Ignored files are absent because of how the anchor is built, not because anything checks. So
the trajectory files stay the **slot's** business and the anchor covers everything else:
the two are not alternatives, and neither can do the other's job.

**It lives outside the stash**, at a location the tool owns. Three properties follow, and
all three are the reason rather than a bonus:

- **It cannot be popped.** An entry in the stash can be applied by a user or agent reaching
  for their own work by habit — and ours would apply a whole-worktree snapshot over whatever
  they have now.
- **It cannot be cleared.** `git stash clear` empties the stack regardless of position, so
  there is no place *within* the stash safe from a routine cleanup.
- **Overwriting is one atomic act**, rather than finding the previous entry and dropping it.

The trade is that it does not appear in the stash listing, so the tool's own output is the
only place it surfaces. That is discharged by the output section below naming the recovery
commands outright — better than the listing was, since a reader needs the command and not
the idiom.

**The commit it was taken from needs no record of ours.** The anchor's first parent is the
commit that was checked out when it was written, so that is recoverable from the anchor
itself. A field the tool wrote could drift out of step with the tree it describes; a parent
cannot. Two distinct repairs follow — restore a file as it stood at the transition
(uncommitted work included), or as it stood at the last commit.

**One limit remains, and it is the one worth stating:** the anchor holds a tree, so an
enormous untracked file that no ignore rule covers is copied into the object store along
with everything else.

## Output

**Every change the tool makes is cranked out**, printed in full, so an agent is never in the
dark about what happened to files it did not write. An agent that mutates state through a
tool and then cannot see the result will reason from a stale picture and eventually assert
it.

**The output names where the anchor is and how to reach it.** Since the anchor is
deliberately outside the stash listing, this is the only place a reader learns it exists —
so it carries the commands for restoring a file as of the transition and as of the commit
beneath it, rather than a location the reader must know what to do with.

Markdown on stdout by default; `--json` for the machine-readable form.

## The skill documents it, and revert especially

A safety mechanism nobody knows about is not a safety mechanism. `/mini-spec` is where an
agent learns the escape hatch exists *before* it needs one — which is the only time
learning it is any use.
