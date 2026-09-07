# Update Commands

Atomic modifications to structured parts of design files.

## minispec update check [file] [item]

Check a checkbox in the specified file.

Examples:
```
minispec update check design.md D1      # Check gap item D1
minispec update check design.md src/store.ts  # Check artifact checkbox
```

## minispec update uncheck [file] [item]

Uncheck a checkbox in the specified file.

## minispec update add-ref [crc-file] [Rn]

Add a requirement reference to a CRC card's Requirements field.

Example:
```
minispec update add-ref crc-Store.md R5
# Changes: **Requirements:** R1, R3
# To:      **Requirements:** R1, R3, R5
```

## minispec update remove-ref [crc-file] [Rn]

Remove a requirement reference from a CRC card.

## minispec update add-gap [type] [description]

Add a new gap item to design.md Gaps section with auto-numbered ID.

Types: S (spec), R (requirement), D (design), C (code), I (implementation), O (oversight), A (approved)

Example:
```
minispec update add-gap R "Requirement R5 has no design coverage"
# Adds: - [ ] R2: Requirement R5 has no design coverage
# (assuming R1 already exists)
```

## minispec update resolve-gap [id]

Mark a gap as resolved (check its checkbox).

Alias for `minispec update check design.md [id]`

## minispec update approve-gap [id]

Convert an existing gap to approved (A) type. The gap's type changes to A with a new auto-numbered A-ID. The description is preserved. Approved gaps are always unchecked (`[ ]`).

Example:
```
minispec update approve-gap D3
# Changes: - [ ] D3: Some design gap
# To:      - [ ] A1: Some design gap
# (assuming no A gaps exist yet)
```

## minispec update retire [Rold] [Rnew|-] [reason]

Retire a requirement: rewrite the `Rold` line in `requirements.md` with the
strikethrough/Retired marker and append a new `Tn` gap to `design.md`,
atomically. The assigned `Tn` is printed to **stdout** — the machine-readable
result a caller captures.

After retiring, the command prints a **supersede-at-source reminder** to
**stderr** (suppressed by `--quiet`). Striking out `Rold` does not remove the
prose that *described* its old behavior: a stale spec sentence or design
bullet is a trap a future agent can read as current intent and "fix" the code
back toward, reverting the change the retirement was part of. The reminder
names `Rold`'s originating spec (its feature's `**Source:**` in
`requirements.md`) so the prose to reconcile is one pointer away, prompts a
grep of `design/` for the old directives, and states the completion test:
*could an agent reading only specs + design be led to undo this change?* When
`Rold` has no recorded `**Source:**`, the reminder says so instead of naming a
file. The reminder is advisory — it never blocks the retirement.

Example:
```
minispec update retire R12 R40 "ec-rekey: keys moved to chunkID"
# stdout: T7
# stderr: ⚠ supersede at the source: R12 is retired, but directives describing
#           its old behavior remain and can cause a future revert.
#           • originating spec prose: specs/storage.md  (R12's **Source:**)
#           • design prose: grep design/ for R12 and the old names
#           completion test: could an agent reading only specs + design undo this?
```

## minispec update number-alarms [file...]

Assigns `**Alarm:**` numbers to every alarm that has none, writing the field into the
document directly above its `**Fire alarm:**` line. With no argument it covers every
`design/test-*.md`; named files narrow it.

**It is append-only and it never renumbers.** An alarm that already carries a number keeps
it, whatever order it now sits in. A new number is the next free one **in that file** — the
maximum ever assigned there plus one, so a number freed by deleting an alarm is not handed
out again. That is the requirement rule applied to a smaller scope, and for the identical
reason: a reused number makes every recorded reference to the old one silently name the new.

**It reports what it wrote, per file, and it writes nothing else.** The check that matters
after a run is that each document differs from its previous self by added `**Alarm:**` lines
and nothing more — measured 2026-09-07 on this repository's fifteen test designs, 22 numbers
assigned across eight of them and no other line touched. *Idempotent by construction:* a
second run finds no alarm lacking a number and writes nothing, which is what makes it safe to
re-run after adding a test rather than something to schedule. A document whose unnumbered
entry carries a deviation — a doubled field — is refused whole, before any byte moves, with
the deviation named; the first run here stopped on exactly one, a `**Code:**` line written
twice the day before.

## minispec update pulled \<doc\>#\<n\> --body-file \<file\>

Records a fire alarm as pulled. Today's date comes from the **system clock**; the body is read
from a file, byte for byte.

**The leading date is what the census reads**, so a re-pull must move it. Appending *"re-pulled
today"* further along the line records the history for a human and leaves the alarm stale to
the tool — measured 2026-08-18, when three alarms were re-pulled, all three rang, and the census
stayed red until the leading dates moved. The previous line is folded after the new one as
` *Earlier —* …` rather than overwritten, because a pull's history is the only record of what an
injection used to do. A first pull is inserted directly after the `**Inject:**` line the record
vouches for.

**The body comes from a file and not from an argument**, and that is fidelity by construction:
a backtick inside a shell argument is command substitution, and when it fires the text is
simply gone from what the tool receives, with nothing reporting it. *The date is the system's,
not the session's:* a session that read its date from its opening greeting once wrote two days
of stamps and staled three alarms on the next commit. **No commit follows the date** (Bill,
2026-09-04): items are squashed to one commit before they finish, so a hash written during the
work names a commit the squash rewrites away.

**What the tool does not supply is the verdict.** Whether the alarm rang, on which assertion,
with what caveat — that is the reader's judgment, and the one thing a delegate is forbidden to
send back. The tool supplies the placement, the stamp and the fold.

## minispec update inject \<doc\>#\<n\> \<file:symbol\>[, ...]

Re-sites a fire alarm, and **voids its `**Pulled:**` when the sites resolve to different code.**

*The comparison is of resolved extents, not of the field's text.* Text is right about a move
and wrong about the three other ways a field changes — a rename, a re-formatting, and a
disambiguation from `Parse` to `Head.Parse` — all of which read as a move to a string
comparison and would void a proof still good. Extents are self-sorting: rewriting an anchor to
name the declaration it already resolved to keeps the record, and naming a different one
clears it. **The comparison is asymmetric on purpose**: the old sites are resolved in `HEAD`,
where a renamed symbol still exists, and the new sites on disk, where the rename is. `old-sdom`
resolved both on disk and so could not tell a rename from a move — its `O120`, closed here by
the extent Item 6 gave `LastChanged`.

That voiding is the whole verb. A pull date was earned at the *old* symbol, so once the site
moves it vouches for nothing — and nothing can detect that: the census asks git about whatever
the field now names and answers confidently about a function the injection was never run
against. **The claim is voided and the record is kept**: the line is demoted out of field
shape, naming the sites the pull was earned at, rather than deleted. The alarm reads
`unrecorded`, which is the truthful state.

**A site that resolves to nothing counts as different**, since an anchor naming no code cannot
be shown to name the *same* code, and the conservative direction is to clear: understating a
proof costs one re-pull, overstating it leaves a date vouching for a function nobody checked.
**Rewriting the sites to what they already hold changes nothing and clears nothing**, so the
verb stays usable for tidying. An empty site list is refused rather than written — an alarm
with no site is `unanchored`, a state to record rather than a value to write.

## Prose reaches the tool through a file, not a shell argument

`pulled` takes `--body-file`, as every trajectory verb does. A backtick inside a double-quoted
shell argument is **command substitution**, and when it fires the text is simply **gone** from
what the tool receives, with nothing reporting it — measured 2026-08-20, four instances in one
session, to a caller who knew about the hazard. `add-gap` and `retire` still take their prose as
arguments here (gap `O23`).
