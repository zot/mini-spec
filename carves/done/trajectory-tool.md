# Carve: offload the trajectory layer to the tool

The skill describes the trajectory layer — the pending/current/done queue, carves,
items — as prose conventions and asks people to maintain them by hand. Most of
that prose is *mechanics*: file shapes, the lifecycle, the next-free-ID rule, the
status block format. Mechanics can be delegated the way the phase sections already
delegate. Judgment cannot, and the split between the two is what this carve is
about.

**Provenance.** Promoted 2026-08-04 from a private working note,
`TRAJECTORY-TOOL-20260804.md`, which holds the raw diagnosis and the drafts that
were settled against. Everything from it that a stranger needs is restated here;
the note is not linked because `.scratch/` is gitignored in this repo, which is
the case the [reference rule](reference-discipline.md) exists to prevent.

**Annotations.** `@undecided:` marks a proposal we are not yet sure is a decision —
the state between unmarked thinking and `DECIDED`, and named for the state rather
than for an action so that nothing reads it as an instruction to go and decide. It
converts to `DECIDED` when settled, or it is deleted. It shares a stem with
`DECIDED`, so one grep for `decid` spans both states.

**DECIDED (Bill, 2026-08-04): bare is a use, backquoted is a mention.** Written
bare, it is a real tag and ark indexes it, so what is unsettled is findable from the
corpus rather than only from this file. Written in backquotes — as everywhere in
this paragraph — it is a mention: ark skips it, and so can a reader's grep, by
dropping any hit whose next character is a backtick. One character does the work a
markdown parser would otherwise have to do. It does not retire that parser: a *link*
inside a code span carries no such marker, which is still Item 1 of
[reference-discipline.md](reference-discipline.md).

## Status

Ordered by intent, so the numbers run out of sequence. Position is the priority; the
number is only the identifier.

**DECIDED (Bill, 2026-08-04): parts are numbered.** A part's `Item N` is the join
between its one-line status here and its elaboration below — without it the only way
to connect the two is by matching title text, which breaks the moment a title is
edited, the same fragility as anchoring on line numbers.

**DECIDED (Bill, 2026-08-13): subparts are numbered too, and get their own header
entries.** *This supersedes the second half of the 2026-08-04 decision, which said a
sub-item "has no separate elaboration to join to, so it needs no key," and the rule
that a sub-item is numbered only when it enters a pending item by itself.* Both rested
on sub-items being mere breakdown. Under the split rule below they are scheduling
units, so they need the same join their parents need, for the same reason.

A subpart is **namespaced by its parent and numbered from 1** — `3.1`, `3.2` — with a
header entry keyed by bare number exactly as a part is. Numbering is **append-only**:
once a queue entry points at `3.2`, renumbering `3.2` leaves the pointer resolving to
different work with nothing able to detect it, which is the same silent failure that
makes `Rn` and sequence-step IDs permanent. Gaps in the sequence are expected.

**DECIDED (Bill, 2026-08-13): split until no part is completed by more than one
item.** When a part would take several items, issue subparts so each is discharged by
exactly one. The parent is then completed **when its subparts are, derived rather than
stored** — it keeps no checkbox of its own, since a parent box would be a second copy
of a fact the subparts already carry, and two copies are what disagree. Item 1 is the
worked example: `SPLIT`, no parent checkbox, three sub-items landing in three commits.

So **the number appears exactly at the scheduling boundary**, and never at two levels
for the same work — the 2026-08-04 formulation, now reached by splitting rather than
by withholding keys. That remains checkable: for any part, either the part carries a
`#N` or its subparts do, not both.

**The reverse direction stays open on purpose** (Bill, 2026-08-13): one *item* may
still discharge parts in several documents — a harness item closing gaps in three
carves. So the link is asymmetric. A part records exactly one item; a queue entry
records a list of parts. Splitting fixes many-items-per-part; it was never meant to
forbid many-parts-per-item.

`Item N` and `#N` are two keys doing two jobs, not one thing said twice: `Item N`
identifies within the document, `#N` identifies in the queue and appears in the
marker. Ark's live carves already read this way — `**Item 8 — a test harness…**
**OPEN (#121.)**`.

**DECIDED (Bill, 2026-08-13): a part is identified by `<doc-path>#<key>.<subkey>`.**
The fragment resolves *inside* the named document, so the key scheme is a per-document
property and no project-wide convention is needed: `carves/x.md#7` is whatever `x.md`
keys as `7`, and cannot mean a `#7` from somewhere else. The parser accepts every live
form and records the literal key. Measured across the eight carves in ark and this repo
(2026-08-13): three schemes are in use — `Item N` (four documents), the bare queue
number (three), and `Part X` letters (one) — and **no document mixes two**, which is
what makes the fragment unambiguous. Subkeys follow the same shape, so the three read
`3.2`, `72.1` and `B.1`.

That uniformity is convention rather than a checked property, and one document holding
both an `Item 7` and a `#7` is the single state where resolution breaks. `validate
trajectory` asserts it — see Item 3.

The shape is deliberately the one already in use for sequence anchors (`Seq:
seq-foo.md#1.4`): same `file#dotted-fragment` syntax, a different resolver, and context
always says which, since a `Seq:` reference and a part pointer never appear in the same
position.

**DECIDED (Bill, 2026-08-16): the done-entry shape is ark's, plus a part pointer — and
mini-spec migrates, not ark.** `- **YYYY-MM-DD — #8 / R189–R198: title.** (`commit`) Part
`<doc>#<key>`.` Identifiers lead, so a reader scanning three thousand lines sees the joins
before the prose; the slot takes a queue ID, a gap ID, a requirement range, several
separated by `/`, or nothing for an incident. The part pointer is the one genuine addition
— ark's shape has nowhere to record which part a change discharged, and Item 3's
referential integrity and Item 6's bidirectional link both need it.

*Recorded because the error is instructive and it is mine.* `trajectory-format.md`'s
original done-entry shape described **mini-spec's `DONE.md`** — a ten-entry file I had
been authoring that same week — and I wrote it down as the target that older projects
migrate *to*. Measured 2026-08-16: it matched **0 of ark's 54** done entries, while ark's
own shape had survived a 3,753-line ledger. The pending half was fine (62 of ark's 66
entries match), so the divergence was one section, invented rather than observed.

This is the same failure as the key-form decision below, one level worse. There a real
measurement across eight carves was read as normative; here there was no measurement at
all — the normative claim was generalized from the instance I happened to be writing, in
the days I was writing it. **A format document is a target, which makes the question of
where its shapes came from load-bearing rather than academic.** The rule that falls out:
when writing down a shape, name the corpus it was read from, or do not claim it.

*How it surfaced is worth keeping too.* `query next-id item` run in ark reported
`DONE.md 0` — zero identifiers from a 3,753-line ledger — and the answer it gave was
nonetheless **correct**, because ark's highest live ID (`#123`) happens to exceed its
highest done one (`#118`). It becomes wrong the moment a high-numbered item completes,
which is exactly the collision R190 exists to prevent, arriving through a door R190 did
not cover. Without R197's per-file counts it would have read as an ordinary right answer.
The evidence line earned its keep on its first use outside this repository.

**DECIDED (Bill, 2026-08-16): one key form — `Item N` — and the other two are migrated
away.** *This supersedes the half of question 2's answer that said the parser accepts
every live form and records the literal key.* Three schemes were measured live on
2026-08-13 and that measurement was **descriptive**; it was then read as **normative**,
which is the error. A format document exists so an agent can bring its carves up to code,
not so it can enshrine every shape it happened to find.

*One argument settles which, and it is structural rather than a preference: the queue
number cannot be a key.* A part exists long before it is scheduled — nine of twelve open
parts here are "not queued" right now — so a key drawn from `#N` is undefined for most
parts most of the time. `Item N` is the only form that always resolves, it was already the
plurality, and its subparts extend naturally into the decided fragment shape (`Item 9` →
`#9.1`). `Part B` letters carry no ordering and do not subdivide cleanly.

**What this deletes is the point.** Item 3 was handed a check — *"a document uses exactly
one key form"* — that existed **only** because three were allowed. It required scanning
free text for key-shaped strings outside the status block, which is the one place the
tool's read could not be positional. With one mandated form that check becomes a
*conformance* question answerable from the status block itself: is the key `Item N`?
Positional, cheap, and it tells a migrating agent exactly what to fix.

*So the ingest/lint split raised on 2026-08-16 is no longer needed here.* It was proposed
to accommodate that one free-text scan; removing the scan removes the exception. **Ingest
by position** stands alone as the rule, which is the better outcome — the simplification
dissolved the complication rather than managing it.

*Migration is eight documents and mechanical.* Backward compatibility is explicitly not a
concern (Bill): `trajectory-format.md` states the target, and bringing a carve up to code
is an agent's job.

**DECIDED (Bill, 2026-08-16): one shared markdown-aware extractor, not one per parser.**
[reference-discipline.md](reference-discipline.md) Item 1 already owns *"it must parse
markdown, not grep it"*, and it earned that from the identical failure — a throwaway
checker run on **this carve** on 2026-08-04 flagged a `[text](path)` inside a code span as
a missing file. Trajectory Item 3 needs the same skill for status entries: a fenced
`## Status` example sits at exactly the positions the tool reads.

*The consequence is a scheduling one, so it is recorded rather than left implicit:*
**reference-discipline Item 1 becomes a prerequisite of trajectory Item 3**, and rises in
the order below. The two carves stay separate documents; only the parser is shared.

**DECIDED (Bill, 2026-08-16): the tool/agent contract is written into
`trajectory-format.md`, folded into Item 3** rather than minted as a part of its own. It
states which positions the tool ingests and that everything else is the agent's to write
freely — currently learnable only by reading the parser. It becomes load-bearing exactly
when the tool starts parsing carve content, which is Item 3.

**DECIDED (Bill, 2026-08-16): dogfooding sets the order from here.** Build what we need
as we need it, then *use* it instead of a UNIX tool. Both halves are the rule — the
failure he actually caught was the second one: `query next-id` was built in the middle of
`#10` and the requirements added later that same item were still hand-numbered by grep.

*So the ranking axis changes.* It was drift-reduction; it is now **how much hand-labor
this part removes from the next session**, measured from what we actually did by hand:

| part | what it replaces | measured today |
|---|---|---|
| **8.2** `query carves` | `grep '^- \[ \]' carves/*.md` | run ~6 times, and it is the instrument this carve opened by documenting — a tree-wide version counts 32 open items where 13 exist |
| **3** `validate trajectory` | reading ledgers by eye | found three `#N` in the carve resolving to nothing in `DONE.md`, by hand |
| **5** → **6** the link and completion | four hand edits per item | three items landed this session, twelve hand edits, each a chance to forget one |
| **12** `add-req` | `grep -oE '\*\*~?~?R[0-9]+' \| sort -n \| tail -1` | R189–R198 numbered that way |

**5 before 6 is a real dependency**, not a preference — the completion verb writes prose,
and the backup slot is what makes a bad write recoverable.

*The reorder below is mine, from that table; the intent behind it is Bill's.* Override the
positions freely — the numbers are stable and only the order carries meaning.

*Landed parts sit below the open ones as of 2026-08-16.* The block exists so "what is
still open?" is answerable from the first screen, and a completed part carries no
priority — so leaving them interleaved spent the first screen on finished work. A split
parent moves with its sub-items, since they are one entry.

*Brought up to this tree on 2026-09-06.* This branch restarted from `abd78fd` (Item 13's commit)
on 2026-09-04, so six parts that landed on `old-sdom` between 08-16 and 08-18 have their done
entries in the shared ledger but their code only on that branch. Four of them were re-landed here
over simple-dom by [sdom-reclaim.md](sdom-reclaim.md) and are marked with both commits; the
worktree anchor came across with the backup port; `add-req` did not, and is open again. The
`carve-stencils` carve that Item 10 and half of Item 7 moved to on `old-sdom` is not in this
tree either, so both read as they did on 2026-08-16.

**Open, in dogfood order:**

- [x] ~~**Item 12 — `update add-req`, the requirement-minting verb.**~~ **LANDED (`a543897`, 2026-09-07 — `#75`.)**
  Landed on `old-sdom` as `d367eee` (2026-08-18 — `#32`) and not in this tree; returned as
  [sdom-reclaim.md](sdom-reclaim.md) Item 7, over the dependency's requirements reader.
- [x] ~~**Item 10 — status is a checkbox, not a prose stamp.**~~ **MIGRATED (2026-09-13 — `trajectory-format.md`, *The interior is `[ ]` or `[x]`* and the stateless rule; R216, R217.)**
  This carve's own Open questions took checkboxes on 2026-08-13; retrofitting ark's carves is ark's.
- **Item 9 — `trajectory-format.md`: the normative format reference.** **SPLIT (Bill, 2026-08-14.)** No
  checkbox: the sub-items carry the state.
  - [x] ~~**9.1 — write `trajectory-format.md`.**~~ **LANDED (`64b9fd6`, 2026-08-14 — `#8`.)**
  - [x] ~~**9.2 — the generated section, and the check that keeps it honest.**~~ **SUPERSEDED (Bill, 2026-09-14.)**
    Decided 2026-08-14 when the tool's own reader implemented the shapes, so a copy of the
    format document in the tool's spec, checked by `make validate`, was what made a cite able
    to fail. Since the 2026-09-04 restart the shapes are implemented by mini-spec-tool's
    readers, whose specs state them, and this tool's specs cite the format document the way
    `file-formats.md` already does; what keeps the shapes honest is `validate` reading real
    files through those readers and the second opinion disagreeing when they diverge (R300,
    R327). A generated copy would guard two texts against each other, and neither is what runs.
- [x] ~~**Item 7 — creation: the refusal path, and `init carve`.**~~ **LANDED (`d3cfe3e`, 2026-09-13 — `#78`.)**
- [x] ~~**Item 11 — conform minted-value output to markdown-by-default.**~~ **LANDED (`d8f4a4f`, 2026-09-13 — `#77`.)**
- [x] ~~**Item 16 — the completion verb says which of its writes is tracked.**~~ **LANDED (`1da22f9`, 2026-09-12 — `#31`.)**
  Small, and it closes a hole the format opens by construction.
- [x] ~~**Item 14 — the worktree cleanup report.**~~ **LANDED (`d5e25ef`, 2026-09-14 — `#79`.)** Last on purpose:
  belt and suspenders for an emergency Bill does not expect to reach for (2026-08-17). Its
  one real dependency ships early regardless — the worktree anchor is written at each
  transition, and a report built before any anchor exists has no baseline to diff against.

**Closed:**

- [x] ~~**Item 6 — the bidirectional item↔part link.**~~ **LANDED (`d34cbdf`, 2026-08-18 — `#17`.)** **NARROWED (Bill, 2026-08-17.)**
  Re-landed here by sdom-reclaim Item 3 (`eaf9604`, 2026-09-05): `add-item` writes both sides.
- [x] ~~**Item 15 — the worktree anchor.**~~ **LANDED (`4b9c311`, 2026-08-17 — `#19`.)** Split
  out of Item 6 the day it was built. `Snapshot` and `refs/minispec/snapshot` are in this tree
  (R236–R238); closed `O23`, retired R261/R262, deleted `SnapshotLimits` on `old-sdom`.
- [x] ~~**Item 5 — the backup slot: revert and replay.**~~ **LANDED (`42b119d`, 2026-08-17 — `#16`.)**
  Re-landed here by sdom-reclaim Item 2 (`7dd50a0`, 2026-09-04).
- [x] ~~**Item 3 — `validate trajectory`.**~~ **LANDED (`7953a87`, 2026-08-16 — `#15`.)**
  Re-landed here by sdom-reclaim Item 4 (`7e6e293`, 2026-09-05).
- **Item 8 — shrink the skill.** **SPLIT (Bill, 2026-08-14.)** No checkbox: the sub-items
  carry the state, and both now carry it.
  - [x] ~~**8.1 — move the mechanics out of `SKILL.md`.**~~ **LANDED (`9dfe5f8`, 2026-08-14 — `#9`.)**
  - [x] ~~**8.2 — `minispec query carves`.**~~ **LANDED (`42d771e`, 2026-08-16 — `#13`.)**
    Re-landed here by sdom-reclaim Item 1 (`2a050a2`, 2026-09-04).
- [x] ~~**Item 13 — read the adopted done-entry shape, and migrate this repo's ledger.**~~ **LANDED (`abd78fd`, 2026-08-16 — `#12`.)** This branch's starting commit.
- [x] ~~**Item 2 — `query next-id`.**~~ **LANDED (`c86c4b8`, 2026-08-16 — `#10`.)**
- **Item 1 — the bootstrap: root, config, and `init`.** **SPLIT (Bill, 2026-08-07.)** No
  checkbox: the sub-items carry the state, and a parent box would be a second copy of it.
  - [x] ~~**1.1 — repository-root detection.**~~ **LANDED (`8197c6c`, 2026-08-07 — `#1`.)**
  - [x] ~~**1.2 — `.minispec/` and the config move.**~~ **LANDED (`fe0cd11`, 2026-08-07 — `#3`.)**
  - [x] ~~**1.3 — `init`, `track`, and `--repair`.**~~ **LANDED (`0711319`, 2026-08-11 — `#4`.)**
- **Item 4 — reference-discipline checking.** **MOVED (Bill, 2026-08-04 — [reference-discipline.md](reference-discipline.md).)** No checkbox: nothing here left to close.

## Why

The skill already names this defect about carves, and the prediction was about
itself:

> A migration gets a forcing function for free (the retire reminder, the prose
> grep, the completion ritual). A carve gets none, so its decisions rot silently
> while still reading as current.

### The drift arrived in one afternoon

Four failures in ark on 2026-08-04, all mechanically checkable, none needing
judgment to *detect*:

1. **Stale next-free-ID counters.** The current file said next item `#120`, next
   gap `O209`; the real maxima were `#121` and `O211`. The next person to file by
   that line would have reused `#120`. The same file documents the next-free-ID
   rule three hundred lines below the stale copy, and warns that a stale copy
   caused an ID reuse six days earlier. It went stale again anyway.
2. **A checkbox that outlived its work** — `- [ ] the tiers`, after the tiers had
   landed and the carve said so.
3. **A gap closed in prose and open in its checkbox**, for a day. The checkbox was
   right: the prose closed it when the mechanism landed, while the sentence it was
   written about was still true.
4. **Six carves, three notations, three with no status at all.** Nothing could
   answer "what is open across all carves" until they were normalised by hand.

### The sharper argument: ad-hoc checks are silently wrong

Diagnosing the above meant writing throwaway shell. Four instruments were written
that afternoon and **four returned confident wrong answers**:

| instrument | failure |
|---|---|
| `grep 'R338[0-9]'` over a coverage report | the report prints **ranges** (`R3386-3388`), so the tail of every range was invisible; reported a requirement covered when it was not |
| `grep -oE '~~[^~]{0,55}~~'` | the bound was 55, the target string 57 chars; concluded a carve had no status block when it had a table |
| `awk "/func \(db \*DB\) Fetch/,/^}/"` | `\*` from a shell variable became a regex quantifier, so the pattern wanted `db` + spaces + `DB`; reported that the doors did **not** call `resolvePath`, which would have falsely accused a carve of lying |
| `grep -oE "next free item ID is\s*#[0-9]+"` | the text wraps mid-phrase inside bold markers; reported the counter absent |

Every one failed the same way: **reported absence rather than error.** That is the
argument for tested tool code over per-session improvisation — not convenience,
but that a wrong check is indistinguishable from a passing one.

## Decisions

**DECIDED (Bill, 2026-08-04): this is a carve, not a migration and not a single
spec.** A migration is a concise A→B document that expires by design, and B is not
determined here — three of the open questions below are data-model questions whose
answers change what the commands are. The backup slot settles it independently: it
is a tool-wide safety mechanism that would retrofit onto the eight existing
`update` verbs, so it is schedulable and valuable whether or not a single
trajectory check is ever written. A migration containing a part that is not about
the migration is the wrong container.

### Reference discipline — moved

The rule that governs which documents may cite which, and its four decisions, now
live in [reference-discipline.md](reference-discipline.md). Stated once, there, so
there is no second copy to drift.

It still binds *these* documents: this carve is public, so it may cite only
VCS-managed files, which is why the provenance note above names its working note in
prose instead of linking it.

### Crank handle and revert

**Every change the tool makes is cranked out** — printed in full, so the agent is
never in the dark about what happened to files it did not write. An agent that
mutates state through a tool and then cannot see the result will reason from a
stale picture and, eventually, assert it.

**DECIDED (Bill, 2026-08-07): the tool mints IDs, and never hands out a bare one.**
Assignment and the write that records it are one invocation, so a number is in a
document the moment it exists and `max()` over the documents stays the whole truth.
The alternative — a command that reserves an ID for the agent to write later — is a
second copy of the numbering state by construction, which is the stale-counter
failure recorded above with the hand taken out of it. Keeping such a copy honest
would need reservations to expire, and expiry is the lease the lock was already
decided not to need.

*The idiom already exists in the tool.* `update add-gap` auto-numbers and writes in
one act, and there is no `query next-gap-id` because none was ever needed. What is
missing is the same shape for the other two classes: items, which Item 6 covers, and
requirements, which have no verb at all today.

**DECIDED (Bill, 2026-08-07): every minted ID is cranked out.** The agent has to know
the number in order to refer to it afterwards, and it cannot read one out of a file it
did not write. This is the paragraph above applied to the one piece of state the tool
*creates* rather than modifies, which is the case where being in the dark is total
rather than partial.

**DECIDED (Bill, 2026-08-07): the requirement verb is `minispec update add-req`, an
`add-gap` twin addressed by section heading.** It mints the next `Rn`, appends the
entry to the named section, and cranks out the number — or the range, for a batch.

*Scheduled 2026-08-16 as **Item 12**, which is where its status lives.* Everything below
is the design; nothing below is a plan. **The decision spent a week with no part**, which
is the failure this carve is about happening to this carve — see Item 12 for what that
cost and how it surfaced.

*Measured across both projects before deciding.* Requirements are defined in
`design/requirements.md` and nowhere else: **zero of ark's 147 spec files** carry a
`- **Rn:**` definition, and the 63 that mention an `Rn` are citing one. There is a
single mint site, so the earlier worry that a spec edit might be the minting act was
unfounded.

| | ark | mini-spec |
|---|---|---|
| `## Feature:` headings | 193, all unique | 17, all unique |
| `### ` sub-headings | 366, 6 titles duplicated | none — the file is flat |
| `**Source:**` lines | 195, of **142 distinct values** | non-unique |
| spec files defining an `Rn` | 0 of 147 | 0 of 9 |

**Address the destination by heading text, at whatever level it exists.** Not by
`**Source:**`, which is emphatically non-unique — `specs/main.md` heads nine separate
groups in ark. And not by Feature, which is too coarse: `Source Monitoring` alone
holds 68 entries spread across its sub-sections. One `--section "<heading>"` serves a
flat file and a two-level one alike, and it resolves uniquely for 360 of ark's 366
sub-headings. The six collisions — `CLI`, `Go API`, `Store API` and their kin — refuse
and crank out the qualified forms for the caller to re-run, which is the same refusal
shape as everywhere else here rather than a new mechanism.

**It appends to an existing section and refuses on an unknown one, and that is the
line worth holding: the tool owns IDs, not prose.** A sub-section title is authoring —
"Phase A: Config-Triggered Reconcile" is a judgment about how the design decomposes,
and a tool inventing one would be doing the part that is not mechanical. A new heading
is hand-written first, which races nothing because it carries no ID.

**Batch is the primary form, not an accommodation.** The skill's own instruction is to
"merge all specs into numbered requirements", and ark's file shows blocks throughout:
R338-R341 landed together, R961-R962 much later in the same sub-section. The crank
handle reports a minted block as `R963-R966`, in the range syntax the tool already
parses.

*One correction, recorded because this carve is about instruments that lie.* The first
measurement of append practice split on `## ` alone and reported that 20% of groups do
not end at their highest number, which read as evidence that requirements are
sometimes inserted rather than appended. Wrong: ark's grouping is two levels deep, so
a Feature was never the unit, and those out-of-order tails are ordinary appends to
sub-sections. The instrument answered exactly the question it was asked.

**DECIDED (Bill, 2026-08-04): the slot covers the trajectory files only, and the
surface is `minispec pending [add-item | start | finish | revert | replay]
[ITEM-NUM [FLAG...]]`.** The pending item is the smallest unit of work, so it is the
right thing to be able to undo. Tracking the rest of the working tree is a different
and much larger problem, and *the tool must never reach for blanket git resets or
staging* — only safe, verifiable operations.

*Consequence, stated so nobody is surprised later:* the eight existing `update` verbs
stay unprotected. They still write via bare `os.WriteFile` with no backup. Scoping
the slot to trajectory files first is a fine increment; it just is not the tool-wide
mechanism an earlier draft of this carve assumed.

**DECIDED (Bill, 2026-08-04): one trajectory stash at a time, named `mini-spec
trajectory`.** Starting a new pending item drops the previous one and takes a fresh
snapshot; revert and replay do the same. The stash is *reference* — it tells you what
changed in the working tree since the item began. Reporting a difference destroys
nothing, which is what makes it safe to build. **Undoing that difference is a
different act with a different risk**, since the user may have done unrelated work
along the way, and nothing here should slide from the first into the second.

**The command matters, and the obvious one is wrong.** `git stash push` moves the
changes out of the working tree — destructive, and exactly what must not happen. The
safe idiom is `git stash create` to build the commit object, then `git stash store -m
"mini-spec trajectory"` to record it. Verified 2026-08-04 in a scratch repository:
the working tree is untouched and the entry appears in `git stash list` under that
name. Find it by message rather than index, since a user's own stashes shift the
indices.

**Two limits, both measured, both changing what the snapshot can claim.**

*It cannot hold the trajectory files.* They are gitignored, and `git stash create`
captures tracked modifications only — ignored paths are absent from the commit. So
the stash and the backup slot are not alternatives: copy-then-stamp keeps the
trajectory files, the stash covers the tracked working tree, and neither can do the
other's job.

**DECIDED (Bill, 2026-08-04): the trajectory backups live in a directory, not a
stash.** A plain directory holds ignored files, which is exactly what git refuses to
do, and it is where the stamp file belongs too — keeping the copies and the stamp
together is what makes "one stamp, atomic across all the files" obvious rather than
implied.

**DECIDED (Bill, 2026-08-04): the directory is `.minispec/`, and the repo-level
config moves into it as `.minispec/config.yaml`.** Not `.carve/saved`, which sits
one letter from `.carves/` — two directories differing by a single character at the
same level is a permanent misread — and which would put tool-managed scratch inside a
directory of hand-written human documents.

**The rename is the load-bearing half.** Until now `.minispec.yaml` meant two things
at once: the design-root config, and the repo-root config that the lock and the root
search wanted. That ambiguity produced the same hazard twice in one evening — touch
it at a design root and you teach the search that a design root is a repository.
Moving the repo-level file into `.minispec/` makes the two scopes *structurally*
distinct rather than same-name-different-place, so the hazard cannot be written.

*It also upgrades the root search.* `.minispec/` is tool-created at the repository
root by construction, so it is an unambiguous strong marker and belongs beside
`.git`, `carves/` and the trajectory files. `.minispec.yaml` stays where it was in
the chain — the weak final fallback, meaningful only because a design root is
sometimes also the repository root.

*No migration falls out of this.* Ark's `.minispec.yaml` sets `comment_patterns`,
which is design-root configuration; it stays exactly where it is. A repo-level file
appears only where something repo-level needs saying, and today nothing does — the
trajectory config surface is empty, so its first job is being the lock.

**MIGRATED 2026-08-07 (`fe0cd11`).** The config structure decided here — both scopes, the
inheritance, the flat top-level error, and the merge rules that fell out of
implementing it — is now specified in the **Config Scopes** section of
[tool/specs/config.md](../../tool/specs/config.md) and numbered R118–R130. **That spec is
the authority; what follows is the reasoning that produced it.** Where the two ever
disagree, the spec wins and this section is stale.

*One rule the carve never stated, discovered by building it:* lists merge as a union
**between configuration layers**, but the first configuration layer *replaces* the
built-in defaults. Without that exception the shipped `code_extensions` would be
permanently un-narrowable — union can only grow a list, and the remedy for an unwanted
entry (drop it from the layer above, state it here) has nowhere to reach, since
nothing sits below the defaults. Found by a failing test, not by review.

**DECIDED (Bill, 2026-08-04): the config structure is `/.minispec/config.yaml` plus
per-project `.minispec.yaml` files that inherit from it, and there is never a
`.minispec.yaml` at the top level of the repository.**

- `/.minispec/config.yaml` — the repository's configuration. Where the repository
  root is itself a design root, this is that design root's configuration too.
- `<repo-project>/.minispec.yaml` — one per design root beneath the top, stating only
  what differs from the repository config. A project whose settings match it needs no
  file at all.
- `/.minispec.yaml` — **an error, with no exception.** The tool flags it.

No exception is needed because the top level never has its own `.minispec.yaml` in
either style: a full-repo project keeps everything in `/.minispec/config.yaml`, and a
repo-project layout has no design root up there to configure. The rule stays flat, so
the check is one existence test rather than a condition about what else sits beside
it.

**DECIDED (Bill, 2026-08-07): the real reason is that the inheritance would be
incoherent, not that the case never arises.** `/.minispec.yaml` would have to inherit
from `/.minispec/config.yaml` — a file inheriting from a file inside its own
directory. The argument above is from absence, which invites a reader to go looking
for the case it says does not exist; this one holds whatever layout you are in, and
it is why the rule needs no exception rather than merely happening not to need one.

*Consequence, and it lands on the reference project.* Ark had `design/` at its
repository root **and** a `.minispec.yaml` beside it setting `comment_patterns`. Bill
moved that file to `.minispec/config.yaml` on 2026-08-04 and left a symlink behind at
the old path. Verified the same day: the released binary follows the link and still
resolves ark's custom `.html` alternation, so the move costs nothing while the tool
still reads the old location.

**DECIDED (Bill, 2026-08-04): the symlink is transition scaffolding, removed once
this lands.** It is not a supported shape, so the rule stays flat and the check stays
a single existence test — no special case for links, and no indefinite support for
binaries reading the old location. Until then ark works normally through it.

**DISCHARGED (`5ea35f1`, 2026-08-11 — queue item #5.)** Ark is repaired: `track:
private-trajectory` recorded, `/.minispec/backup` ignored, and the symlink gone.

*The order this went in matters more than the removal, and it is the one thing to
carry forward.* The paragraph above assumed the upgraded tool lands first and its gripe
is what prompts the removal. It went the other way, because the binary was unreleased —
and removing the link took the **old** binary's only route to ark's config with it, so
`~/.claude/bin/minispec` silently stopped resolving ark's `.html` comment alternation
and would have under-reported coverage with no error. Closed by rebuilding
`minispec-linux-amd64` from current source. **Transition scaffolding comes out after
the thing replacing it is installed, never before** — removed early, the breakage is
silent rather than noisy, which inverts the whole point of the gripe.

*And the sequencing needs no coordination* (Bill, 2026-08-04). The upgraded tool
gripes on its first run, and that gripe *is* the reminder to remove the link — the
same shape as every other refusal here: the tool detects and instructs rather than
anything being arranged in advance. A draft of this section treated "the reference
project fails the check" as a problem to schedule around; it is the notification
working.

**DECIDED (Bill, 2026-08-04): it cranks out the instruction and exits — no attempt to
continue, and no backward compatibility for the old location.** The message names the
file, says the configuration now lives in `.minispec/config.yaml`, and stops. A draft
here proposed reporting without blocking so work could proceed; that is precisely the
backward compatibility not wanted, and a warning nobody has to act on is one that
survives for years. A hard stop costs one `rm` once.

### Git: preferences, not manipulation

**DECIDED (Bill, 2026-08-04): mini-spec does not manipulate the git repository**, its
one trajectory stash aside. No staging, no commits, no resets, no branch state.
*Editing `.gitignore` is not manipulation in that sense* — it is editing a file, not
changing git's state — which is what keeps this consistent with the `--create` path
that already writes a trajectory filename there.

**DECIDED (Bill, 2026-08-04): two preferences, griped about while unmet.**

- `.minispec/config.yaml` **should be tracked** by git.
- `.minispec/backup` **should be ignored** by git.

Both are computed properties — `git ls-files` and `check-ignore` answer them — so
there is nothing to assert and nothing to go stale. That is the test applied to
everything else here, and these pass it.

**DECIDED (Bill, 2026-08-04): the stamp lives in `.minispec/backup`.** One ignore
line then covers everything machine-local, and `config.yaml` is the only thing left
in `.minispec/` — so "the tracked one is the only one you can see" holds by structure
rather than by remembering two paths.

**DECIDED (Bill, 2026-08-04): `minispec init` creates a basic `.minispec/config.yaml`
and adds `.minispec/backup` to the top-level `.gitignore`.**

### `track`, and the bootstrap rule

**MIGRATED 2026-08-11 (`0711319`, extended by `5ea35f1`).** Everything decided in this
section is now specified in
[tool/specs/initialization.md](../../tool/specs/initialization.md) and numbered R131–R177.
**That spec is the authority; what follows is the reasoning that produced it**, kept
because a spec states behavior rather than why it was chosen, and the three-tier
argument in particular is worth more than the rule it produced. Where the two disagree,
the spec wins.

*R173–R177 have no ancestor here either, and they came from the same source as R170–R172:
running the thing on a real repository.* The gate reported a configuration with no
`track` as malformed and sent the agent to hand-edit, while `--repair` had always
accepted that exact case — so the tool refused a repair its own repair path would have
performed, and the hand edit it recommended sets `track` without reconciling
`.gitignore`. **Absence is a version difference; a wrong value is damage.** And
`--repair` itself round-tripped the configuration through the settings struct, deleting
ark's ten-line comment block along with any key a newer tool version might have written.
Both were invisible to every fixture, because a fixture has no comments to lose.

*Three requirements have no ancestor here, because building it produced them.* R170:
whether a path is already ignored is asked of git, never decided by matching lines —
`/PENDING.md`, `PENDING.md` and `*.md` are the same intent written three ways. R171:
lines the tool writes are anchored, since every governed path is mandated at the
repository root. R172: a path the chosen value wants public but some rule outside the
top-level `.gitignore` still ignores is *named*, not silently left.

That trio came from one defect, found the first time `init` ran on a real repository —
this one. Text matching did not recognise the anchored `/PENDING.md` already present,
appended the bare form beside it, and because the last matching pattern wins, replaced
a narrow rule with a broader one. Recorded here rather than only in the spec because it
is the sharpest instance of this carve's own argument: **every fake agreed with the
broken code, and the tests all still passed after the bug was found.** They asserted
presence; the property was multiplicity. Ad-hoc reasoning about a format is silently
wrong in exactly the way ad-hoc shell was.

**DECIDED (Bill, 2026-08-04): `minispec init --track-STYLE`, and the flag is
mandatory.** It records a `track` setting in `.minispec/config.yaml`:

| flag | setting | meaning |
|---|---|---|
| `--track-none` | `none` | no version control |
| `--track-private-trajectory` | `private-trajectory` | git; `init` puts the trajectory files in `.gitignore` |
| `--track-all` | `all` | git; `init` does not ignore them |

Mandatory because the tool cannot infer it. Whether a repository *is* git-managed is
checkable after the fact; whether you want your queue to ship is not inferable at
all, and a default would make that choice for someone without their noticing.

**DECIDED (Bill, 2026-08-04): consistency is checked at startup, and a mismatch
gripes and exits.** `track: none` verifies the repository is *not* git-managed and
says to change the setting to `private-trajectory` or `all` if it is; any other value
verifies that it *is*, and says to change it to `none` if not. This catches the real
drift — a project that gains git after `init`.

*Amended 2026-08-07: the check covers ignore state too*, not only whether git is
present. Under `private-trajectory` the trajectory files must be ignored; under `all`
they must not be; `.carves/` must be ignored whenever it exists, and it need not
exist, since it is created on demand. `git check-ignore` takes a list, so the whole
question costs one invocation beside the one that answers git presence. What a
mismatch does is settled below.

**DECIDED (Bill, 2026-08-04): `init` is the sole creator of `.minispec/config.yaml`,
and with no config the only legal commands are `init` and those needing no
minispec-specific files at all** — `minispec version` and its kin. This supersedes an
earlier decision that the tool would touch the config into existence so it had
something to lock: that would have produced a config with no `track`, which is
exactly the undefined state the mandatory flag abolishes.

**Two things this simplifies, and one it complicates.**

*The `--create` privacy question disappears.* With `track` already recorded,
`--create` is uniform — it creates, honouring the setting — so `--no-ignore` goes
away and the refusal table below loses its case 3. Exactly one case then needs a
human, and it is the root confirmation.

*Root detection can no longer fail after `init`.* A config implies `.minispec/`
exists, and `.minispec/` is a strong root marker, so the upward search always
succeeds. **The ask-the-user fallback is therefore an init-time mechanism**, not
something an arbitrary command might hit.

*But the "run `init`" gripe can point somebody at the wrong action.* Run a command
from **above** your project and no config is found, so the tool says to initialise —
and doing that plants a repository root in the wrong directory.

**DECIDED (Bill, 2026-08-04): the refusal is a crank handle carrying the gist plus an
instruction to the agent to check whether the directory looks like a code project.**
The gist:

```
It doesn't look like this is a minispec project.
If DIR is a code project, would you like me to make it work with minispec?
```

**Why the agent checks rather than the user being asked outright: never ask a human a
question the agent can answer.** The user's irreducible contribution is *intent* —
"yes, I want this to be a mini-spec project." Whether there is code in the directory
is a *fact*, and the agent can go and look. Asking a non-technical user "if DIR is a
code project…" hands them the one part they may be least equipped for while the agent
stands right there able to settle it.

That gives three tiers, each doing only what the others cannot: **the tool detects
the missing config, the agent inspects the directory, the human supplies the intent.**

Three things this needs to get right.

*The most valuable finding is the negative one.* This message exists because somebody
ran from **above** their project, and a directory containing many project directories
is exactly what an agent can recognise. "This looks like a folder of projects rather
than a project — did you mean one of these?" is far more useful than a yes, and it is
the case that would otherwise plant a repository root in the wrong place.

*The crank handle must include the stop.* The moment it says "check whether this looks
like a code project," a weaker agent can conclude *yes* and run `init` — building
precisely the road left deliberately unbuilt. The instruction ends with **report and
wait**. This is the one place in the design where an agent is told to go and look at
something ahead of a user decision, so it is the one place the boundary has to be
spelled out rather than left as an absence.

*The tool gives the gist; the agent writes the sentence.* Verbatim copy cannot adapt
to what the agent found — it would still say "if DIR is a code project" after the
agent has established that it plainly is, or plainly is not. So the tool supplies what
to check and what to ask, and the agent composes for the human.

*That also dissolves the two-register problem* rather than managing it. An earlier
draft here split crank-handle output into agent-facing instructions and user-facing
copy, and worried about which the agent should relay. Under this shape **the tool is
always agent-facing** and the human-facing sentence is always composed. One register,
one rule, and the message fits the situation instead of being a script read aloud.

This is also the case where the agent may run `init` — because the user asked it to,
which is the standing rule.

**DECIDED (Bill, 2026-08-07): `track` is verified on every run, not merely stored.**
This settles the objection that it duplicates what `.gitignore` already says. It does
duplicate it — but a copy re-checked on every non-trivial invocation cannot drift
undetected, and undetected drift was the whole of the objection. A stored claim rots;
a checked one is a checksum. The copy also earns its keep independently, because
`--create` needs the policy months after `init`, possibly before any ignore line
exists to read it from.

*The weaker alternative was proposed and rejected.* Daneel argued for treating
`track` as creation policy and letting divergence raise a standing advisory. Rejected
on this project's own measurement: in ark, every category carrying a forcing function
sits at zero, and the one category carrying only a reminder sits at 19%.

**The order of business, for any command doing more than reporting its version:**

1. **No `.minispec/config.yaml`.** `.gitignore` does not matter yet. This is the
   no-config refusal already decided above, now also carrying the `track` choice. The
   agent establishes whether the directory is git-managed — that is a fact, and the
   three-tier rule says the agent goes and looks — then asks the user, in as few plain
   words as possible, whether they want this to be a mini-spec project and whether the
   queue should stay private or ship with the repository. On yes, the agent runs
   `minispec init --track-VALUE`.
2. **Config present.** Lock it, load it, verify it is well-formed, and verify `track`
   agrees with both facts: git presence, and the ignore state of the trajectory files
   and of `.carves/` when it exists.
3. **Disagreement.** Gripe and exit with a crank handle. The agent repairs by running
   `minispec init --track-VALUE --repair`; it does not hand-edit the config.

**DECIDED (Bill, 2026-08-07): `--repair` is symmetric, and the agent confirms with the
user before running it.** It sets `track` to the given value *and* brings `.gitignore`
into agreement with it. Symmetry is what lets one command resolve a mismatch in either
direction: `private-trajectory` with unignored files can be fixed by adding the ignore
lines or by declaring the files public, those mean opposite things, and the flag value
is how the user chooses between them. That choice is the user's, which is why the
confirmation is required rather than polite — the tool detects, the agent relays, the
human decides. Not a new mechanism.

**DECIDED (Bill, 2026-08-07): plain `init` refuses when `.minispec/` exists**, because
it would be stomping on something. Its crank handle names `--repair` as the way to
change a `track` value, and says to confirm with the user first. The two forms are
mutually exclusive on that one precondition, which is what keeps either from having to
guess the caller's intent.

**DECIDED (Bill, 2026-08-07): a malformed config is the one place the agent may edit
it.** `--repair` validates well-formedness before doing anything, and a config that
fails validation cannot be repaired by a flag — the damage is arbitrary, most likely a
stray character typed while the file was open in an editor. So the crank handle states
every problem found, points at the configuration documentation in the skill directory,
and **explicitly authorises the edit**, because at that point the agent is the only
actor left who can act. Back up the config first, and skip the backup when it is
byte-identical to the one already there.

*That documentation is an obligation this decision creates.*
[`.claude/skills/mini-spec/config-reference.md`](../../.claude/skills/mini-spec/config-reference.md)
documents `.minispec.yaml` — the design-root config — and says nothing about
`.minispec/config.yaml`, `track`, or inheritance. The crank handle would point at a
section that does not exist, so writing it is part of the work rather than a follow-up.

**On the standing gripe.** It fires on every run until satisfied, which is right on
tonight's evidence: ark's one-shot stderr reminder on retirement left 19% of retired
requirements still referenced from live code. A reminder seen once decays to nothing.
This one cannot become wallpaper either, because it is *closable* — `init` once and
it is gone. A recurring nag you can never discharge is a different animal. In a tree
with no git neither preference is meaningful, so it stays quiet.

**DECIDED (Bill, 2026-08-04): `init` is for the user to run; `--create` is
agent-runnable with the user's permission. Neither is enforced.** If the user asks
the agent to run `minispec init`, it runs it — no prohibition goes in the skill.

*The mechanism is absence, not enforcement, and that is the point.* In both commands
the authority is the user; what differs is whether a path exists that could route
around them. `--create` has a defined path *through* the user — the tool refuses, the
agent relays the question, the user answers, the agent re-runs — so the agent is a
courier rather than a decider. `init` has no path at all: **the tool never runs it
implicitly and the skill never tells an agent to run it**, so the only way an agent
reaches it is a direct request, which is the user deciding in person.

A prohibition would be worse on both counts. It is text, so it can be misread,
rationalised around, or argued out of, and it needs maintaining; an unbuilt road needs
none of that. And it would be actively wrong — it would make the agent refuse a
legitimate direct request, which is worse behaviour than running the command.

Note the gripe is a crank handle aimed *through* the agent at the user, the same shape
as the root-confirmation refusal. Not a new mechanism.

*It cannot see new untracked files.* A source file created during an item is absent
from the snapshot, so a diff against it **under-reports** — the failure mode this
whole carve is most alert to, since a check that reports absence rather than error is
indistinguishable from a passing one. The output has to say what it did not look at,
rather than leaving the reader to discover it.

**One level of undo, and one of redo.** Explicitly *not* an undo stack: the 80s
`vi` `u` with goldfish memory. The most recent change is immediately revertable and
replayable, and nothing older is recoverable. That is enough for the real emergency
— a command that did the wrong thing thirty seconds ago — and it avoids owning a
history the VCS already owns better.

**DECIDED (Bill, 2026-08-04): revert by copy-then-stamp.** Copy the files to a
scratch area, make all the changes, then touch a stamp file. On revert, verify none
of those files has changed since the stamp; if any has, **refuse** and crank out
which files changed and where their backups are. The operator is better placed than
the tool to reconcile a hand edit with a pending revert, and a revert that silently
clobbers a later edit is worse than no revert.

The stamp is what makes this cheap: no content hashing, only an mtime comparison,
and comparing against a single stamp rather than per-file snapshots is what keeps
the check atomic across all four files at once. This is `make`'s dependency model —
marker file, mtime, no history — and it is right for the same reason it was then:
the cheap check is *sufficient*, and anything cleverer buys accuracy nobody needs.

**DECIDED (Bill, 2026-08-04): every operation is the same swap, and the stamp file
carries the state.** Each of change, revert and replay does: tmp-copy the live
files, perform the operation, write the stamp, then `mv` the tmp copies over the old
backups. The stamp records `changed`, `reverted` or `replayed`, and revert and
replay run the identical drift check.

*Why one backup set suffices:* the backup always holds **the other state**, so
revert and replay are one mechanism with two names. A second copy for replay is
redundant, and the swap is what makes it so.

*Why the tmp-then-`mv` ordering is load-bearing, so nobody simplifies it away:*
`mv` within a filesystem is atomic, so a crash leaves either the old backup intact
or the new one complete, never a half-written backup. Same reason one writes a temp
file and renames rather than truncating in place.

**DECIDED (Bill, 2026-08-04): three states, and these are the only transitions.**

```
        (any state) --<a new mutation>--> changed        [backup replaced]

        changed  --revert-->  reverted
        reverted --replay-->  replayed
        replayed --revert-->  reverted
```

`changed` is what a fresh mutation leaves behind — an item retirement, say — where
a backup was taken but neither a revert nor a replay has happened.

*Why three and not two.* Two states counts *configurations*, of which there are
indeed two (live-holds-the-change, live-holds-the-original), and concludes that
`changed` and `replayed` are the same thing. They are the same configuration and
different **nodes**: `changed` is the entry, reachable only by a fresh mutation and
never by toggling. Collapsing them would mean a fresh retirement setting the state
to `replayed`, which is the tool asserting something untrue — nothing was replayed.
A stamp that lies is worse than a stamp with one more value in it.

**The property to build on:** from every state, **exactly one** of `revert` /
`replay` is legal. There is never a choice to disambiguate and never a
"reverting twice" case to define — the state fully determines the single legal next
operation, and anything else is refused with a crank-handle message saying which
state the slot is in and what it will accept.

**Goldfish, stated exactly:** a new mutation from *any* state resets to `changed`
and replaces the backup. Whatever was previously revertable is gone, without
ceremony. That is the whole memory model, and the reason the slot never needs a
history.

## The split

**Item 1** — the bootstrap every other part stands on. *Renamed 2026-08-07: it was
"project parameterization", after the question it was opened to answer rather than
after the work.* That question is answered and the answer was **no parameters** —
siting mandated, prefix dropped, carve directory fixed, privacy universal, part keys a
per-document property. The trajectory config surface came out empty. What remains
under the number is everything the decisions built instead, and it splits three ways:

- ~~**repository-root detection (#1)**~~ — **LANDED (`8197c6c`, 2026-08-07.)** The
  upward marker search, the two roots the word "project" was hiding, and the `$HOME`
  exclusion. Decisions migrated to
  [tool/specs/repository-root.md](../../tool/specs/repository-root.md) as R107–R117,
  with `crc-RepoRoot.md`, `seq-reporoot.md` and `test-RepoRoot.md`; the sections
  below keep the reasoning and the measurements, not the specification.
- **`.minispec/` and the config move (#3)** — the directory, `config.yaml`, the
  inheritance model, and the flat top-level-`.minispec.yaml` error. Needs #1, since all
  of it is sited at a root the tool must first be able to find. Decisions migrated to
  the **Config Scopes** section of [tool/specs/config.md](../../tool/specs/config.md) as
  R118–R130. *The lock target moved out of this sub-item:* a requirement saying "one
  lock per repository" is untestable while nothing locks, so it belongs with whatever
  implements locking rather than banking an immediate implementation gap here.
- **`init`, `track`, and `--repair` (#4)** — the creation verbs, the startup
  consistency check, and the refusal paths. Needs #3, since it writes the config.

*Ordering is a real dependency chain, not a preference*, which is why this splits into
three scheduled items rather than one. Each lands something usable on its own.

**Item 9** — `trajectory-format.md` in the skill directory: the normative reference for every
trajectory file shape — the three files, the item entry, the done entry, the carve
status block and its bare-key detail, the marker vocabulary, the ID rule. Loaded on
demand rather than always, the way `config-reference.md` already is.

Three jobs, and the last two are what make it necessary rather than tidy:

1. **The tool implements it.** Items 2, 3 and 7 each encode format knowledge and
   would otherwise each invent their own reading of it.
2. **An agent repairs from it.** When a file is damaged, or predates the tool, or
   was hand-written by someone with only prose to copy from, the agent needs the
   shapes stated normatively — and needs them precisely when the tool is what cannot
   help.
3. **Old formats migrate to it.** A project already running a hand-maintained layer
   has files in some earlier shape. Without a written target there is nothing to
   migrate *to*.

*The coupling to Item 1 is discharged, and discharged in the direction that makes this
easier.* This read "write them together, or the config will parameterize fields the
format does not have" while Item 1 was open. Item 1 landed with an **empty** config
surface — siting mandated, prefix dropped, carve directory fixed, privacy universal — so
there are no varying slots and `trajectory-format.md` states mandates rather than defaults. Nothing
here waits on anything.

### The four decisions Item 9 was holding

**DECIDED (Bill, 2026-08-14): `trajectory-format.md` is the authority for the shapes; the tool spec
carries a generated section and never restates them by hand.**

*The mechanical fact that reshapes this, found before deciding.* A `**Source:**` value
must be a relative path resolving inside `specs/` — no leading slash, no `..`
([tool/specs/validate.md](../../tool/specs/validate.md), Spec Source Validation) — so a skill
file **can never be a requirement's Source**. The moment the tool parses trajectory
content, a `tool/specs/trajectory.md` must exist regardless. So the question was never
which file wins; it was whether that spec restates the shapes or cites them.

The authority half is already recorded and only needs honouring:
[tool/specs/file-formats.md](../../tool/specs/file-formats.md) says the trajectory files'
"content is owned by the skill, not the tool," and that file already models
cite-don't-restate for `design.md` — `config.md` owns the format, `validate.md` owns the
structural rules. The format is tool-agnostic besides: the skill says trajectory tracking
"ships with mini-spec but is not about mini-spec," and ark has run the layer for months
with no tool support at all.

**What "cite" alone could not do is fail.** That is this project's own recurring finding,
and `#7` is ten days old: a generator over prose maintained elsewhere must fail loudly or
it does not fail, and an honest generator only helps when something runs it. So the shapes
live in `trajectory-format.md`, `tool/specs/trajectory.md` is hand-written for **behavior** with one
**generated section** holding the shapes, and `make validate` asserts it. Section-scoped
generate-and-compare — generate the part that is owned elsewhere, hand-write the prose
around it. That is Item **9.2**, and it is why this part splits.

**DECIDED (Bill, 2026-08-14): the marker vocabulary is open, and the *shape* is what gets
checked.** The dichotomy this carve recorded — "a closed set is checkable; an open one is
honest, not both" — does not survive Item 10. The checkbox is the queryable signal and the
stamp is provenance, so the machine never needs to read the verb. `trajectory-format.md` **reserves**
the live verbs with fixed meanings and permits more; Item 3 asserts `**VERB
(attribution)**` and never the word.

*Measured 2026-08-14 across 15 carve files in both projects, and the measurement is what
decides it.*

| | verbs |
|---|---|
| pre-existing | `LANDED` `OPEN` `NOT VERIFIED` `DEFERRED` `SENT` `ASSESSED` |
| coined by **this carve**, in ten days | `SPLIT` `MOVED` `DISCHARGED` `MIGRATED` — none of the four appears in any of ark's 13 carve files |

Ten live verbs, four of them added in ten days by one document. A set closed today would
refuse four of the words currently in use, and it would have refused them *while this
carve was arguing that closing it might be right*. `SKILL.md` already says "let the verb
vary where it carries information"; this ratifies that rather than overturning it.

**DECIDED (Bill, 2026-08-14): no version number — absence is the version.** A file the
tool created carries the marker; a file without one is legacy, and that is the whole
signal a migration needs. Open question 9 asked whether the format carries a version
marker and made it conditional on migration being real rather than hypothetical: it is
real but small and one-time — 15 carve files, three key notations, already retrofitted by
hand once. A number in every file, forever, buys nothing until there is a second version
to name, and that is when it earns its place.

**DECIDED (Bill, 2026-08-14): `@undecided:` and *blocked* are one mechanism at two
severities** — one greppable marker carrying a required *what would settle it* clause,
ordinary in a carve, a counted defect in a spec. **Kept minimal for now**, and the reason
is adoption rather than doubt: at the moment of deciding, the marker had **zero live
uses** — all four occurrences in this carve were backquoted mentions, and ark has none. So
`trajectory-format.md` documents the marker and its clause, and does not yet build the
blocked state on top of it.

*It acquired its first live use an hour later, in the naming question at the end of this
section* — which is the better argument for keeping it than the count was. A marker exists
for the state where something is genuinely proposed rather than settled, and writing 9.1
produced exactly one.

*One thing that measurement did establish.* Those occurrences being mentions is the
use/mention rule working exactly as designed — which makes this document's own claim about
what `@undecided:` remains a computed property rather than an assertion. One grep
separates the two on a single trailing character, with no markdown parser involved.

*And one shape it misreads, found by writing the grep down.* A regular expression quoting
the tag puts a metacharacter where the backtick would go, so the heuristic scores it a
use. Harmless here — it was reworded out of this paragraph rather than left as a false
positive in the reference instance — but it is a second witness for the same limitation
the annotation decision already names: the marker handles the common case, and a code span
still needs the markdown-aware extraction that is Item 1 of
[reference-discipline.md](reference-discipline.md).

### The split

**9.1** — the normative reference itself: the three files and their mandated siting, the
item entry, the done entry, the carve status block and its bare-key detail, subpart
numbering, the reserved marker vocabulary and its shape, the checkbox-versus-stamp rule
from Item 10, the `@undecided:` marker, and the ID rule.

**9.2** — `tool/specs/trajectory.md` gains a section generated from
`trajectory-format.md`, with `make validate` asserting it — the `#7` shape, scoped to a
section rather than a file. Depends on 9.1, since there is nothing to generate from until
the authority exists.

**DECIDED (Bill, 2026-08-14): the file is named `trajectory-format.md`, not the
`format.md` this carve called it for ten days.** Daneel's call while writing 9.1, put up
as this document's first live `@undecided:` and confirmed the same day. The reasoning: the
skill directory already documents CRC cards, sequences, test cases and requirements
formats inside `SKILL.md`, so a bare `format.md` would be the one unqualified name among
them and will attract the wrong content later; the sibling it is modelled on,
`config-reference.md`, is qualified by its subject for the same reason.

*Recorded because the marker's whole lifecycle ran in one session.* It was documented in
the morning with zero live uses, acquired its first use an hour later on this question,
carried the required *what would settle it* clause — Bill saying yes or no, since no
measurement can answer a naming question — and converted to `DECIDED` on his answer. That
is the state the marker exists to hold: not unmarked thinking, not a settled call, and
visible to a grep the whole time it was neither.

**The annotation marker belongs here too** (Bill, 2026-08-04): the tag this carve
invented for its own use should become part of the documented format, for **specs as
well as carves**. Two consequences worth holding onto.

*It is greppable, so it can be checked.* "How many unsettled points does this spec
carry?" becomes a computed property rather than a reading exercise — which is this
repository's own test for whether a convention earns its keep.

*In a spec it is a different animal than in a carve.* A carve is allowed to be
unsettled; a spec describes how the system **is**, so an unresolved point inside one
is a hole in the specification. That makes the marker the lightweight inline form of
the *blocked* state already identified as missing — "we cannot specify this yet, and
here is precisely what we would need to know" — where no amount of effort closes the
item, only an answer does. ~~Whether the two should be one mechanism or two is
genuinely open.~~ **Answered 2026-08-14: one mechanism, two severities** — see the
decisions below.

**One drift hazard to settle up front — settled 2026-08-14, see the decisions below.**
The tool's own specs describe these same formats, and
[tool/specs/file-formats.md](../../tool/specs/file-formats.md) exists as the summary spec for
exactly that. Two hand-maintained normative copies of one format is the failure this whole
carve is about. The fork recorded here — `trajectory-format.md` authoritative with the tool spec
linking, or the tool spec authoritative with `trajectory-format.md` generated from it — was answered
by taking the first and mechanizing it, which is 9.2.

**Item 2** — `minispec query next-id <item|gap|req>`. Prints the next free number
as `max()` across every file that can hold one — for items that is **both** the
pending and the done file, which is the part people get wrong in both directions
(pending's max is too low right after completions; done's is too low while the
highest IDs are still live). Retires the hand-copied counter entirely. Smallest
real thing on the list, read-only, and it removes failure 1 above by construction.

*Amended 2026-08-07: informational only.* With the tool minting IDs, this no longer
feeds a write — it is for a human reading the queue, or an agent orienting itself. Its
correctness still matters, since it is the same `max()` the minting verbs use, but a
stale answer can no longer end up written into a document.

**Item 3** — `minispec validate trajectory`, folded into `validate` or standing
alone. Referential integrity (every `#N` in a carve's status block resolves to a
live queue item or a done entry; every queued item whose linked doc is a carve
appears in that carve's status block), checkbox agreement, counter freshness,
orphans, duplicate IDs across pending and done, and status-block presence. All four
ark failures fall here. Read-only.

~~**Plus one key-scheme check** (Bill, 2026-08-13): a document uses exactly one key
form.~~ **Superseded 2026-08-16 by the single-key-form decision, and this is the shape of
the win.** That check existed only because three forms were allowed, and enforcing it
meant scanning free text for key-shaped strings *outside* the status block — the one read
in the whole design that could not be positional. With `Item N` mandated it becomes a
**conformance** check answerable from the block itself: is the key `Item N`? A document
still using the bare queue number or `Part X` is not ambiguous, it is **unmigrated**, and
the check says so and names the target. Cheaper, positional, and it gives a migrating
agent an instruction instead of a complaint.

**A third conformance check, banked as `O9` while landing Item 13 (2026-08-16): a done
entry whose header carries no identifier slot.** `parseDoneIDs` reads the run between the
date's em dash and the title's colon; a header with no colon yields nothing, and *nothing*
is also the right answer for an entry that legitimately discharged no ID — so the two are
indistinguishable and the skip is silent. Measured in ark the same day: **2 of 54** entries
sit in a pre-adoption shape with no colon, hiding `#84`, `#83` and `#46`, while
`next-id item` reported `DONE.md 46` with no indication that two entries went unread.

Latent rather than live — ark's maximum `#117` comes from a conforming entry — but that is
the third time this exact collision has arrived through a door nobody was watching, after
the one-file read (R190) and the shape mismatch that opened Item 13. It belongs here rather
than in `next-id`: the number is right, the *ledger* is unmigrated, and saying so is a
conformance question. Same shape as the key-form check above — name the target, don't
complain.

**Two things fold in here** (Bill, 2026-08-16):

- **The shared markdown-aware extractor**, from
  [reference-discipline.md](reference-discipline.md) Item 1.1, which is therefore a
  prerequisite. Without it a fenced `## Status` example is indistinguishable from a status
  block: measured 2026-08-16, a grep widened past `carves/` counts 32 open items where 12
  exist, four of the phantoms coming from `trajectory-format.md`'s own example. The same
  failure is recorded in reference-discipline from 2026-08-04, on this document, with a
  link inside a code span.
- **The tool/agent contract section** for `trajectory-format.md`: which positions the tool
  ingests, and that everything else is the agent's to write freely. Today that is
  learnable only by reading the parser. It becomes load-bearing precisely when the tool
  starts reading carve content, which is this part.

*The rule the contract states is one line:* **ingest by position.** What the tool treats
as data comes from fixed positions — a `##` heading at column 0, a `- **` entry header,
lines inside the status block — never from a pattern swept over free text. That is what
makes an agent's prose safe by construction rather than by inspection: the `#8` entry in
this project's own ledger cites `#7` in its body, and the parser cannot see it.

**Item 4** — moved to [reference-discipline.md](reference-discipline.md). The number
stays retired here so nothing renumbers and no later part reuses it.

**Item 10** (Bill, 2026-08-13) — every list this format tracks progress on uses
`- [ ]` / `- [x]`, not a prose stamp. This carve's own Open questions section was the
demonstration: five of ten were settled, marked with strikethrough and an "answered
2026-08-04" phrase, and entirely invisible to the one query the format is built around.
`grep '^- \[ \]'` is *the* machine view across every live carve — a section it cannot
see does not exist to it, however carefully a human marked it up.

The general rule: **a stamp records what happened; a checkbox records what is left.**
Those are different jobs and the format needs both, but only the second is queryable, so
the second is the one that cannot be optional. Strikethrough and the dated marker stay —
they are the provenance a reader wants — they just stop being the only signal.

Three consequences, and the third is the reason this is a part rather than a footnote:

- The Open questions section takes checkboxes (done here, 2026-08-13).
- Anything with no checkbox now *means* something — a standing constraint or accepted
  residue, nothing to close — which is the distinction an approved gap already draws.
  Question 8 is the worked example.
- **Existing carves need retrofitting**, and that is the work. Ark has seven; three had
  independently invented a status notation and three had none, which is the same
  evidence that motivated Item 7's scaffold. A convention that only new documents
  follow leaves the old ones silently unqueryable.

Feeds Item 9, which is where the rule gets written down normatively, and hands Item 3 a
check: a section that tracks progress and carries no checkboxes is a section the machine
view is blind to.

**Item 11** (Bill, 2026-08-13) — bring the `update` verbs that mint an identifier into
line with question 7's answer: markdown to stdout by default, machine-readable behind
`--json`.

*Measured 2026-08-13, and the split is not arbitrary.* Six `update` verbs already
conform — `add-gap` prints `Added O3: <desc>`, `check` prints `Checked X in Y`,
`add-ref`/`remove-ref` likewise — a prose sentence on stdout, suppressed by `--quiet`.
The two that deviate are exactly the two that **mint** something: `retire` prints a bare
`Tn`, and `migration-complete` prints a bare new path. Neither honors `--quiet`. That is
not carelessness; whoever wrote them treated a minted value as a *return value*, which
was the reasonable reading before a rule existed to say otherwise.

`retire` is the sharp case, because the deviation is *specified*: R103 states that
stdout carries only the `Tn` while the supersede reminder goes to stderr. So the
crank-handle output — the part that actually tells an agent what to do next — is on the
channel the rule now assigns to diagnostics, and conforming means **editing a
requirement, not just an implementation.** That is why this is a part rather than a
cleanup, and why it needs its own gap in `design/` when it is scheduled.

Two things to settle when it is: whether `--json` carries the minted value for callers
that want it, and whether `--quiet` should suppress a minted identifier at all — it is
the one output a caller may genuinely need to read back.

*Placement note: parked last because it is the least urgent of these, not because it is
the smallest. Move it if that reads wrong.*

**Item 5** — the backup slot. Tool-wide, not trajectory-specific: every `update`
subcommand currently writes in place via bare `os.WriteFile`
([tool/internal/update/update.go](../../tool/internal/update/update.go)) with no backup
anywhere, so this retrofits onto eight existing verbs and is worth landing on its
own. Prerequisite for Item 6 — and **question 5's answer (2026-08-13: the tool does
edit trajectory files) is why**: the files it will be writing are mostly prose, so a bad
write has more to destroy here than a flipped checkbox has in `design.md`.
**`/mini-spec` must document it, and revert especially** — a safety mechanism nobody knows about is not a safety mechanism, and
the skill is where an agent learns the escape hatch exists before it needs one.

**Item 6** — the bidirectional item↔part link and the commands that use it:
`minispec add-item --from <carve-or-migration>#<part>`, which creates the queue
entry *and* the corresponding part line so the two cannot be created out of step;
and completion, which moves the item to done, clears the current file, **and checks
off the carve or migration parts it discharges** — one command across all four
surfaces rather than four hand edits with three chances to forget one. For
completion to check the right box the link must be recorded, and that is the core
data model here.

**Unblocked 2026-08-13.** Questions 2 and 3 are answered, and between them the data
model is settled: a part is `<doc-path>#<key>.<subkey>`; a part records exactly one
item, a queue entry records a list of parts; and a parent part is never checked
directly, since it completes when its subparts do. So completion iterates the item's
recorded parts, checks each, and checks nothing else — no parent boxes, no inference
about what a part's siblings mean.

Still constrained by the asymmetry in the survey below: with trajectory files private
in every project, the public half of the link is *never* a markdown link — a carve
carries a bare key and the queue side holds the recorded pointer. The riskiest part,
and the one hand-maintenance structurally cannot supply.

**Item 15** (Bill, 2026-08-17) — the worktree anchor. **Split out of Item 6**; its decisions are
sited here and the specification is [backup.md](../../tool/specs/backup.md)
with R266–R273. Only the shape of the work belongs here.

**DECIDED (Bill, 2026-08-17): the anchor holds untracked *contents*, and lives at
`refs/minispec/snapshot` rather than in the stash.** *This supersedes the "capture the
untracked list" decision taken earlier the same day* — it replaces that one rather than
extending it. A decision written and superseded inside an hour is exactly what
supersede-in-place is for, and exempting it because it is fresh is how the rule stops
being a rule.

**Contents rather than names, because a list cannot give a file back.** A captured list can
report that something is new; only the contents can return an untracked new carve or a fresh
source file, which is precisely what someone wants after a catastrophe. It also *removes*
machinery rather than adding it: with the anchor holding untracked files, all three
categories fall out of one tree-vs-tree diff and the side-car list disappears.

*Measured 2026-08-17, because two of these are traps rather than preferences:*

- **`git stash create -u` accepts `-u` and silently ignores it.** Exit 0, a plausible sha,
  two parents, untracked file absent from the tree. A check asking only "did it error?"
  reports it working. First instance recorded here of a *flag* leading the witness, where
  every earlier one was a filter or a query.
- **A temp index gets the boundary right for free.** `GIT_INDEX_FILE=<tmp> git add -A .`
  then `write-tree` includes untracked and excludes **ignored** — so the trajectory files
  stay out *by construction*, and the slot/anchor division of labour is preserved by the
  mechanism rather than by care. The real index and working tree are untouched throughout.
- **`git stash store` refuses a plain commit** ("not a stash-like commit"), so living in the
  stash would have meant faking the merge shape with a synthetic index parent.

**And the stash is the wrong home whatever its shape — which reverses my own argument of an
hour earlier.** I had argued for `git stash list` because a human types that in an emergency.
Two measurements overturned it: **`git stash pop` reached for by habit applies our
whole-worktree snapshot over the user's tree**, and **`git stash clear` destroys the anchor
at any position in the stack.** So Bill's question — could it be stored at the *bottom*? —
turns out to have no safe answer even where it has a technical one, and it has neither:
`stash store` carries no position flag, the only lever is hand-writing
`.git/logs/refs/stash`, and an anchor refreshed at every transition is never old enough to
sink.

`refs/minispec/snapshot` is unpoppable, survives `stash clear` and an aggressive `gc` with
the reflogs expired — anything under `refs/` is a gc root, verified rather than assumed —
and makes the overwrite a single atomic `update-ref` instead of find-by-message-then-drop.
It loses visibility in `git stash list`, and the crank handle covers that better than the
stash list did: it prints the recovery command instead of requiring the reader to know the
idiom.

**The base commit needs no field of our own.** The anchor commit's **first parent is HEAD at
capture time**, so `<anchor>^1` recovers it — and a parent cannot drift out of step with the
tree it describes, where a hash we wrote down could. Two distinct repairs fall out of one
anchor: `git checkout <anchor> -- <path>` restores the file as at the transition,
uncommitted work included, and `<anchor>^1` restores it as at the last commit.

1. **`Git.Snapshot` is rebuilt rather than merely called.** Temp-index tree
   (`GIT_INDEX_FILE=<tmp> git add -A .`, then `write-tree`), committed with HEAD as first
   parent, written to `refs/minispec/snapshot` by `update-ref`.
2. **It gains its callers — plural.** The call goes in `Slot.swap`, the one operation every
   path shares, ahead of `perform` — and ahead of `releaseAttempt`'s marker write on the
   `Record` path, which is the one place a marker is written before the swap.
3. **`seq-backup.md#3.5` moves ahead of 3.3**, since 3.4 writes into a tracked carve and as
   drawn the anchor swallows the tool's own write.
4. **R261 and R262 retire only after the code lands** (→ R266 and R273), with the inline
   comments repointed in the same pass. Retiring first would leave a retired `Rn` as the
   only thing the code fulfils.

*And one obligation the item creates for itself:* `test-Git.md`'s alarm on
`TestSnapshotDoesNotMoveAnythingOutOfTheWorkingTree` was pulled 2026-08-17 against the
**stash** implementation. Rewriting `Snapshot` voids that proof on the day it was taken, and
the injection changes with it — the destructive form is no longer "use the push verb" but
"stage into the real index instead of a temp one". Re-pull, do not re-date.

**Item 14** (Bill, 2026-08-17) — the worktree cleanup report. After a transition, crank out
what has moved in the tracked tree since the anchor, so the agent can help the user clean
up. **The tool reports; the agent proposes; the user decides** — nothing here alters the
working tree, which is the *one transition, recorded two ways* decision in `## Decisions`
stated as a scope.

**Three categories, from a single tree-vs-tree diff** — which is the shape the anchor
decision above bought. Build a tree of the working tree as it stands now, by the same temp-
index method the anchor was built with, and diff the two:

| status | meaning |
|---|---|
| `M` | changed since the transition |
| `A` | new since the transition — tracked or not |
| `D` | deleted since the transition, **and recoverable**, because the anchor holds the contents |

*Verified 2026-08-17 on a tree carrying all three at once*, the `D` being a deleted
**untracked** file — the case the earlier name-list sketch could report and could not return.
No side-car list, and no second source to keep in step with the first.

**Why it schedules after `#17` rather than with it, and it is a dependency rather than a
preference.** `#17` writes the anchor at every transition. A report built first would have
nothing to diff against until the next transition, and **would read as working** for exactly
one run — the failure mode this carve keeps meeting, where the absence of evidence renders
as a clean result.

**Two repairs from one anchor, and they are different answers rather than one answer twice.**
The anchor commit's first parent is HEAD at capture time, so both are addressable without
any bookkeeping of ours:

```
git checkout refs/minispec/snapshot -- <path>     # as at the transition, uncommitted work included
git checkout refs/minispec/snapshot^1 -- <path>   # as at the last commit
```

*One caveat the report must state rather than imply.* The diff shows everything that moved
since the transition, **including work the user did that has nothing to do with the item**.
That is correct and is the point — the question is *what has moved since the queue last
transitioned*, never *what did this item do* — but an agent reading the list as the latter
will offer to revert unrelated work. Naming the question the report answers is its own
instance of the companion rule below: say what you could not ingest.

*And the crank handle carries the discoverability the private ref gives up.* Nothing about
`refs/minispec/snapshot` shows up in `git stash list`, so this report is the only place it
surfaces at all — which is the better trade rather than a cost absorbed, since it hands over
the command instead of requiring the reader to know the idiom. A report naming the files and
not the ref would leave a reader knowing exactly what broke and not how to reach it.

**Item 16** (2026-08-18) — `pending finish` already cranks out every file it wrote. What it does
not say is that **one of them is tracked and the other three are not**, so the agent is left
holding an uncommitted edit to a public document with nothing marking it as different from the
three gitignored writes beside it.

*Why it is structural rather than an oversight.* A `LANDED` record carries the commit hash, so
it **cannot be written until that commit exists** — which puts the source edit one commit behind
the work it records, always, in every project. The format calls the source *"the copy a future
reader trusts, and the one nobody thinks to check"* and puts it first for exactly that reason;
this is the moment it is most likely to be skipped, because the work is done, the commit is
made, and all that remains is a checkbox in a file nobody is looking at.

*Measured 2026-08-18, and the reason it went unnoticed for four items.* `#18` was the **first**
carve part this project has discharged through the verb rather than by hand. `#27`, `#29` and
`#30` all had spec or skill sources with no part to check, so the source-first step had never
actually run here — the hole was reachable the whole time and nothing walked into it.

**The repair is narrow: the crank handle already names the file, so it only has to say what
kind of file it is.** `checked  carves/carve-stencils.md#1` and
`written  carves/carve-stencils.md, CURRENT.md, PENDING.md, DONE.md` are both already printed;
what is missing is the one fact that distinguishes them, and `track` plus `check-ignore` already
answer it on every run.

*Bounded by a standing decision, stated here so the wrong repair is foreclosed:* **the tool must
not stage or commit it.** *"mini-spec does not manipulate the git repository"* (Bill,
2026-08-04) — no staging, no commits, no resets. This reports; the agent relays; the human
commits. Same shape as every other refusal in this carve rather than a new mechanism.

*Placed beside Item 14 because they are the same shape* — both tell the operator what the tool
left behind in a tree it will not touch — and ahead of it because this one is a line of output
rather than a diff engine.

**Item 13** (Bill, 2026-08-16) — make the parser read the done-entry shape the format now
mandates, then migrate this repository's ledger onto it. Small, and ordered first because
the two halves must land together: the shape is decided and the code still reads the old
one, so today the format document and the tool disagree.

*Measured before scheduling, on identical content in both shapes:*

```
- **2026-08-16 — #10 / R189–R198: query next-id.** (`c86c4b8`)   →  next free item ID: #1
- **2026-08-16 — query next-id (`#10`).** `c86c4b8`.             →  next free item ID: #11
```

`doneIDRe` wants a **backticked** `` `#10` ``; the adopted shape has a bare one. So
migrating the ledger first would make `next-id item` compute from the pending file alone
and return a plausible wrong number rather than an error — the R190 collision, arriving
through the door R190 did not cover.

**The order is the work:**

1. **Change `parseDoneIDs`** ([tool/internal/parser/trajectory.go](../../tool/internal/parser/trajectory.go))
   to read `- **DATE — #N / …:` — identifiers leading, bare, `/`-separated, with the slot
   optionally holding a gap ID or a requirement range instead. **Edit R190's text**, which
   describes the shape it reads; this is a requirement change, not just an implementation.
2. **Re-pull the two alarms on that file.** Their `**Pulled:**` proofs are void the moment
   the function is rewritten — that is the whole point of the field — and
   `test-Trajectory.md`'s fixtures are written in the old shape, so they change too.
3. **Migrate this repo's ten `DONE.md` entries.** The new shape is a strict superset:
   date, title, queue ID, commit and part pointer all survive, and the requirement range
   is *promoted* out of body prose into the header. Nothing is lost from the ledger.
4. **Confirm** `next-id item` still answers from both files — `DONE.md 10`, not `0`.

*Ark needs no migration, which was the point.* Its 54 entries are already in this shape;
what changes is that the tool can finally read them. Verify there too once it lands.

**Item 12** (Bill, 2026-08-16) — `minispec update add-req`, the third minting verb. Its
whole design is already settled in the crank-handle section above and is **not restated
here**: mint the next `Rn`, address the destination by `--section "<heading>"` at whatever
level it exists, refuse on an unknown heading because inventing one is authoring, take a
batch as the primary form, and crank out the number or the range.

*It exists as a part because it did not, which is the point.* That decision was dated
2026-08-07 and carried a full design, a measurement across both projects, and a named
consequence — and until today it appeared **exactly once in the repository**, in the
paragraph that decided it. No status entry, no part, no spec, no code. The carve's own
prediction about itself, executed: a decision with no schedulable home reads as current
forever while nothing moves toward it. It surfaced only because `#10` went looking at what
`query next-id` neighbours.

*And `#10` is the evidence for building it.* R189 through R198 were written by hand that
day — the next free number computed by grep, the feature block placed by eye, the range
never checked against anything — which is precisely the work this verb removes. Two of
those requirements (R197, R198) were also appended *after* their CRC cards existed, and
`validate` caught the second as an uncovered requirement. A minting verb does not fix
that, but it is the same class of bookkeeping the tool should be carrying.

*Placement is a guess: parked beside Item 6 because they are the same shape* — the two
remaining minting verbs, one for items and one for requirements — rather than because the
priority was weighed. Move it if that reads wrong.

**Item 7** — creation. Now reached through the refusal path rather than a verb you
have to know about (see the Item 1 section): a write to a missing trajectory file
fails and cranks out the create instruction, with the `.gitignore` question folded
into it. What follows is the original framing, which still describes what creation
*produces*; only the way it is reached has changed. `minispec init carve <name>` is
untouched by that, since no write path leads to a carve that does not exist yet.

This is the half that prevents the problem rather than detecting it. Creating a
trajectory file writes its lifecycle preamble with it, and refuses rather than
overwriting if one already exists. `init carve` creates a carve with its `## Status`
block already at the top and the bare-key detail section beneath. Note this
introduces a **new verb class** — the tool's five existing verbs never create a
file, and `update` only rewrites files that already exist.

`init` itself lands here too, with its two forms and their opposite preconditions:
plain `init --track-VALUE` refuses when `.minispec/` exists, and
`init --track-VALUE --repair` requires it. Both write `.gitignore` to agree with the
chosen `track`, and both reach the user through the agent rather than acting on their
own. Decisions in the `track` section above.

*The evidence is this document.* Ark's six carves were retrofitted by hand in one
afternoon; three had independently invented a status notation and three had none,
which is exactly what a scaffold prevents. The three notations existed because
there was nothing to copy from except prose, and prose gets paraphrased. This carve
was hand-written too.

`init` also front-loads Item 1, because it has to *choose* siting, prefix and key
format rather than merely tolerate them. If `init` can produce a project's
conventions correctly, the validators can read them.

**Item 8** — shrink the skill. The trajectory and carve sections are the largest
prose block in [SKILL.md](../../.claude/skills/mini-spec/SKILL.md), and most of it is
mechanics. ~~Lands last by construction: the mechanics cannot leave the prose until the
tool carries them.~~

**Superseded 2026-08-14, by the paragraph directly beneath it.** The mechanics do not
move to *the tool*; they move to `trajectory-format.md`, which the next sentence has said
since the day this was written and which 9.1 landed as `64b9fd6`. The blocker was a
premise that outlived its own paragraph — and it did so unnoticed for ten days precisely
because a status marker records what is *left*, while the reason a part is parked lives in
prose that nothing re-reads. Recorded rather than deleted: it is this carve's own argument
happening to this carve, for the third time.

*This is a move, not a deletion*, and Item 9 is where they move to. Deleting the
mechanics outright would strip the format from the one place an agent can read when
the tool is unavailable — which is exactly the situation in which a damaged file
needs repairing.

**DECIDED (Bill, 2026-08-14): split, because only half of this still waits on the tool.**

**8.1** — the move itself, unblocked the moment 9.1 landed. It also closes a duplicate
9.1 *created*: `SKILL.md`'s "File shapes" section still carries all three fenced blocks
that `trajectory-format.md` now owns, so the format is currently stated normatively in two
hand-maintained places, which is the failure this carve was opened about. And it deletes
the stale siting prose measured on 2026-08-14 — `specs/migrations/` as the default, plus
a parameterization list naming siting, prefix, case convention and part-keying, all four
decided away 2026-08-04 and all four now contradicted by
[tool/internal/project/track.go](../../tool/internal/project/track.go), which hardcodes the
three filenames at the repository root.

**8.2** — `minispec query carves`, the cross-document status view: one line per carve,
open/landed counts, optionally the open items. Today that is
`grep -n '^- \[ \]' carves/*.md`, and it works only because the notation was standardised
by hand. This is the half that genuinely waits on the tool.

*A note on what the split reveals about 9.2* (Daneel, 2026-08-14). 9.2 was recommended as
the next part this morning and should not have been: it generates a format section into a
tool spec, and the tool has **no requirements about trajectory shapes yet** — measured,
three of eleven `tool/specs/` files carry zero requirements and all three are summary
specs, so a new `tool/specs/trajectory.md` would be a per-feature spec with no feature.
`file-formats.md` points rather than restates, so it is not a second copy either. 9.2
becomes real when Items 2, 3 or 7 write requirements that parse these shapes. Until then
it is a mechanism in search of the drift it prevents.

## What stays in the skill

No tool can carry these:

- Why a carve exists at all, and the coherence-versus-schedulability argument.
- **Promotion is the maintainer's judgment**, explicitly not the agent's. A tool
  must not decide it, and should not even nag.
- Priority ordering — position, not number, and by intent.
- **`NOT VERIFIED` as a distinct state.** A repaired *symptom* reads exactly like a
  satisfied *requirement*; only a human who read the code can tell them apart. The
  tool can see the marker, never validate it.
- Migrate-on-landing, which is about where decisions go, not about file shape.

## Surveyed against two real projects, 2026-08-04

Ark runs the full layer and mini-spec has just adopted it, and they disagree on
almost every parameter — which is the evidence that Item 1 is real work rather than
bookkeeping.

| | ark | mini-spec |
|---|---|---|
| queue siting | top level | top level |
| queue visibility (as surveyed) | **private** — fossil-only, untracked in git | **not ignored** |
| how privacy was declared | filename case (top-level uppercase = private) | nothing |
| project prefix | in the title (`# PENDING: Ark State`), not the filename | none |
| item entry | `## 119. **Title** (`/skill`). status. [link](path)` | same shape |
| part keys in carves | `#N`, `Item N`, and `Part B` — all three, live | `Item N` |
| `.minispec.yaml` | overrides `comment_patterns` only | absent |

*The visibility row is what the survey found, not what was decided.* Both rows now
read "private" under the Item 1 decision below, which is what makes the next
paragraph hold everywhere instead of per project.

**The finding that changes a design, not just a config field:** because ark's queue
is private and its carves are public, a carve there cites an item as a bare `#N`
that *deliberately cannot be a link* — its carve README says so. **The item↔part
reference is therefore asymmetric**: private→public may be a markdown link,
public→private may not.

That asymmetry was the survey's most awkward finding while visibility varied by
project, because the linkable direction flipped with the setting and Item 6 could
not assume either. **Deciding that trajectory files are private everywhere makes it
a constant instead of a variable**: the public→private half is *never* a link, in
any project, so Item 6 records the link on the private side and the public side
carries a bare key. One shape to build, not two.

Ark also uses all three part-key notations at once, in live carves. That is what made
question 2 a real question rather than a matter of picking a favourite — and what
answered it: since no single document mixes two notations, the literal key is
unambiguous within the document that wrote it.

## Item 1 — the bootstrap

Settled — mostly on 2026-08-04, with the last open points closed on 2026-08-07. Split
into three scheduled sub-items the same day; the split and its dependency chain are in the
section above.

*This paragraph said "no `@undecided:` remains anywhere in this carve" until 2026-08-14,
when the naming question in the Item 9 section made that false.* Corrected rather than
deleted, because it is a small worked instance of what the whole carve argues: a true
statement about a document, written inside that document, expires the moment the document
changes — and nothing but a reader noticing stands between the expiry and a confident
wrong answer. The claim belongs to a grep, not to a sentence.

*The section title said "parameterization" until 2026-08-07*, naming the question
rather than the work. The answer turned out to be that there are no parameters, so the
heading had come to advertise an empty config surface.

### The measurement that decides the method

Git classifies a *directory* cleanly — ark's `.scratch/` and `.carves/` are
gitignored, so anything under them is certainly private. It returns the
**ambiguous** answer for the queue: `PENDING.md`, `CURRENT.md` and `DONE.md` are
untracked *and not ignored*, because fossil owns them and git never saw them. That
is indistinguishable from "written five minutes ago and not committed", which is
the case the warn rule exists for.

**So visibility must be declared.** It can still be *verified* once declared, and
the verification survives whatever form the declaration takes: a file that is
ignored *and* tracked has leaked — git permits that state, since ignoring never
untracks — and catching it costs one more question to a `git` the tool is already
running.

### Privacy is a convention, not a parameter

**DECIDED (Bill, 2026-08-04): trajectory files are private by default.** This
supersedes a proposed `trajectory_public` config key and removes visibility from the
per-command surface.

*Amended 2026-08-04 by the `track` setting.* This was first written as "gitignored in
every project — not a per-project setting," with `--no-ignore` as the escape hatch.
The substance is unchanged, but the wording is now wrong: `track` **is** a
per-project setting, chosen once at `init`, with `private-trajectory` as the
recommended style and `all` as the named alternative. The hatch did not disappear; it
moved from a per-command flag to a one-time declaration.

**DECIDED (Bill, 2026-08-04): `.gitignore` is the declaration, and the creation path
is where it is made.** Normal trajectory writing fails when the file does not exist
and cranks out the instruction:

- "PENDING.md does not exist. Re-run with `--create`."

*Amended 2026-08-04.* Three variants were written here, one of which asked the user
whether the file should be ignored. `track` answers that at `init`, so `--create` is
uniform and the branch is gone.

**DECIDED (Bill, 2026-08-04): no git is fine — just create the file, and have the
tool check.** A project without a repository gets no ignore prompt rather than an
instruction naming a file it does not have. Consistent with the reference rule,
which already says non-git projects go unchecked and are told so rather than passed
silently.

**DECIDED (Bill, 2026-08-04): `--create` on the failing command**, not a separate
creation verb. The message can then say *re-run what you just ran, with `--create`*,
which is the shortest thing to read at the moment it is read, and it keeps the
tool's verb count where it is.

**Why `.gitignore` beats a config key.** It is already the file whose job is
declaring what is private, everything already reads it, and using it creates no
second copy to disagree with the first. The consequence is larger than the saving:
once privacy lives there, a carve linking the queue stops being a trajectory concept
and becomes an ordinary instance of a rule already decided in
[reference-discipline.md](reference-discipline.md) — **error on ignored**. The
visibility question is not answered, it is dissolved.

**The human makes the call, once, at creation, and the agent never does.** The tool
detects the state and dictates the response by reference; the agent relays. That
also relocates Item 7: `--create` *is* the scaffold, but reached by doing normal
work and being refused rather than by knowing an `init` verb exists. Discoverable by
construction.

**Adoption, not only creation.** The rule above fixes files the tool makes; existing
ones need a one-time repair. Ark's queue is **not** in `.gitignore` — untracked only
because fossil owns it — so a tool reading it today concludes "public" and would
pass a carve→queue link that dangles for a git cloner. Mini-spec's own `PENDING.md`
was in the same state. Adding the filenames to `.gitignore` costs nothing (it has no
bearing on what fossil tracks) and makes the state unambiguous.

### Two roots, and the word that names them both

Before the decisions, the vocabulary, because its ambiguity cost a design detour on
2026-08-04. **"Project root" means two different things in this repository:**

- the **repository root** — the top of the git working tree, where `.claude/`,
  `carves/` and the queue live. `/home/deck/work/mini-spec`.
- the **design root** — the directory containing `design/`, which is what the tool
  detects and prints as `root:`. `/home/deck/work/mini-spec/tool`, and separately
  `/home/deck/work/mini-spec/example`.

They coincide in ark, which is why the collision stayed invisible. Item 9 and the
skill should use the two explicit terms and never the bare word.

### Siting

**DECIDED (Bill, 2026-08-04): trajectory files live at the repository root,
mandated — no config key.** The reasoning is the one that decides it: *the
trajectory goes across techniques*, so its scope is the repository, not any one
design root. A repo holds one work queue and may hold several design roots — this
one holds two, `tool/` and `example/` — so siting the queue under either would be
wrong, and making it configurable would only offer a way to get it wrong.

This supersedes the earlier pair: the default-plus-key shape, and the
`sibling_trajectory` option that measurement had already shown to be a no-op. The
config surface for trajectory siting is now empty.

**DECIDED (Bill, 2026-08-04): carves are repo-level too**, and for the same reason
stated in one word — a carve is *cross-approach*. That is not an analogy to the
trajectory argument, it is the definition: a carve is "one coherent problem
decomposed into items that may each need a different skill." Work spanning several
skills cannot be scoped to any one design root.

### Finding the repository root

**MIGRATED 2026-08-07 (`8197c6c`).** Everything decided in this section is now
specified in [tool/specs/repository-root.md](../../tool/specs/repository-root.md) and
numbered R107–R117. **That spec is the authority; what follows is the reasoning that
produced it**, kept because the measurements are the argument and a spec states
behavior rather than why it was chosen. Where the two ever disagree, the spec wins
and this section is stale.

**DECIDED (Bill, 2026-08-04): search upward, tracking markers as you go, and never
consider `$HOME` or above.** No explicit key.

- **A `.git`, a `.minispec/`, a `carves/`, or a trajectory file ends the search** —
  the deepest directory carrying any of them is the repository root. All four are
  repo-level by mandate or by definition, so where they coexist they agree, and the
  ordering among them never has to be decided.
- Otherwise, the **deepest `.claude`**.
- Otherwise, the **deepest `.minispec.yaml`** — the weak final fallback, meaningful
  only because a design root is sometimes also the repository root.
- **Otherwise ask.** See the fallback below.

Excluding `$HOME` closes a concrete failure rather than a stylistic one:
`/home/deck/.claude` exists, so without the exclusion a tree carrying none of the
markers resolves its repository root to the user's home directory, and `--create`
then deposits `PENDING.md` there.

**DECIDED (Bill, 2026-08-04): `specs/migrations/` is not a home for trajectory
files — remove it.** It was the siting the skill documented until today, and it is
what made the ranking above a question at all: a project following the published
guidance would have had trajectory files several levels down, so a search-ending
marker there would have resolved `specs/migrations/` as its repository root — after
which every path the tool computed would be wrong *together*, the hardest kind of
wrong to notice. Retiring the siting removes the shape rather than accommodating it,
which is why the three strong markers can be equals.

*Consequence for Item 8, beyond the prose edit:* a project already in that state has
files to move. The tool can see this cheaply — trajectory files found under a
`specs/` directory are in the retired location — so the upgrade can be a warning
naming the files and where they belong, rather than something a reader has to notice
in a changelog.

**DECIDED (Bill, 2026-08-04): when no marker is found, ask — and record the answer
as a marker.** The question is *is the current working directory the repository
root?*, since people normally run from their project directory. On yes, `carves/` is
created there, and the user is never asked again because that directory is itself
the marker every later run finds.

**The tool has no user to ask, so it cranks out an error saying to ask** (Bill,
2026-08-04). It fails, names the working directory, and instructs the agent to put
the question to the user and re-run once answered. That keeps the assent where it
belongs: **the tool detects, the agent relays, the human answers, and the tool
records.** An agent that answered on the user's behalf would be manufacturing the
one thing it must never manufacture — and it would do so invisibly, since a created
`carves/` is indistinguishable afterwards from one the user asked for.

This is the same refusal shape as `--create`: fail, state the situation, name the
re-run. Not a second mechanism to learn.

**This is the third time the same shape has won tonight, and it is worth naming:
declare by artifact, not by config key.** `.gitignore` records privacy, `carves/`
records the root. Both are files that had to exist anyway, both are already read by
something, and neither creates a second copy that can disagree with the first. A
config key would have recorded the same answer in a place nothing else looks.

It also answers the question it replaces — silent default or refusal? — with
neither. A refusal stops work; a silent default writes files somewhere nobody chose;
asking does neither, and it is the one option that ends with the ambiguity gone
rather than deferred.

*One wrinkle, unresolved and small:* git does not track empty directories, so a
`carves/` created purely as a marker would not survive a clone. It barely matters,
since a clone has a `.git` that wins anyway — but ark's `carves/` carries a
`README.md` explaining what a carve is, which persists and earns its place with a
reader. Creating that alongside the directory costs nothing and composes with
`init carve` in Item 7.

Tracking is what makes the fallbacks correct. A walk that decided at each level
could not know whether a `.git` sat above it, so it would have to settle for the
first weaker marker it met; collecting lets the strong marker win from any depth
while the weak ones stay available if it never appears.

*Only `.git`, not `.fslckout`* — consistent with the git-only decision in
[reference-discipline.md](reference-discipline.md), and it costs nothing: a
fossil-only tree simply falls through to the next marker down the list, which is the
right answer there rather than a gap. Ark is such a tree in principle and carries a
`.git` anyway, so the fallback goes unexercised there.

**Why collecting beats a first-match walk**, in two measurements.

`.claude` is a weak marker: it exists at three levels of this very path.

```
/home/deck/work/mini-spec    .claude  .git
/home/deck/work              .claude
/home/deck                   .claude
```

A first-match walk that accepted `.claude` would resolve `/home/deck/work` as the
repository root for any design root lacking one of its own — a directory that is not
a repository at all. The decided algorithm is immune here, because `.git` at the
repository root wins from any depth; that immunity is exactly what collecting buys,
and it is why `.claude` ranks below the VCS marker. It remains a per-directory
convention rather than a repository boundary, which is why the search stops short of
`$HOME`. The shallowest `.claude` in that listing is the home directory, reached
only when nothing deeper carries a marker — which is precisely the case the
exclusion exists for.

`.minispec.yaml` is already the design-root marker, read at
`<design root>/.minispec.yaml`
([tool/internal/project/project.go](../../tool/internal/project/project.go)). In a repo
with two design roots each may carry its own, and the innermost wins any upward
walk — so "first `.minispec.yaml` going up" would resolve this repository's root to
`tool/`, precisely the wrong answer. **Presence is not a declaration**, which is why
it is the fallback rather than the first heuristic.

*Caveat on all of it:* ark's repository root, design root, config and `.claude` are
one directory, so none of this is observable there. Mini-spec is the only project
where the distinction exists, which makes it a sample of one for the interesting
behaviour.

**DECIDED (Bill, 2026-08-04): drop the filename-prefix parameter.** The skill lists
it; zero of two projects use it. Ark carries its prefix in the *title*
(`# PENDING: Ark State`), which nothing parses and so needs no configuration.

**DECIDED (Bill, 2026-08-04): the lock is `.minispec/config.yaml`, and the tool
ensures it exists — touching it if it does not.**

*Amended 2026-08-04: the tool does not touch it into existence.* `init` is the sole
creator, and with no config the only legal commands are `init` and those needing no
minispec files at all. Auto-creation would have produced a config with no `track`,
which is the undefined state the mandatory flag exists to abolish — and the lock
target is guaranteed anyway, because nothing that needs a lock can run before `init`.

**Always the repository file — one lock per repository, never per repo-project.** A
design-root lock would let `tool/` and `example/` take different locks on the same
queue, which is no lock at all. The move of the repository config into `.minispec/`
is what makes this unambiguous: there is now exactly one file it could mean.

**It is a short-term, low-contention lock**, and knowing that bounds the design. Its
only job is keeping two agents from stomping on each other mid-edit; it is not
protecting a hot resource, so it needs no fairness, no queueing, and no lease. Held
across a single file edit and released.

**DECIDED (Bill, 2026-08-07): every command touching a trajectory file takes the
lock, reads included.** One flat rule beats a per-command judgment about which reads
are really read-only, and it buys a consistent read rather than one taken mid-write by
another agent.

*This is a consistency detail, not a correctness one, and the distinction is worth
keeping.* A per-invocation lock cannot make an ID safe on its own — `query next-id`
followed by a separate write leaves a gap no lock spans, since the lock is released at
exit. That hazard is closed by the tool minting IDs inside the writing invocation, not
by locking the read. Lock everything because it is cheap and uniform; do not mistake
it for the thing that prevents collisions.

**Both files at once — superseded 2026-08-07, kept as a record.** This paragraph read:
"When the repository root *is* a design root (ark's shape), `.minispec.yaml` and
`.minispec/config.yaml` legitimately sit side by side in one directory. When the roots
differ, a top-level `.minispec.yaml` is the error above."

**That contradicted the config-structure decision above and was wrong**, on exactly
the case that matters most — the reference project's own layout. A top-level
`.minispec.yaml` is a flat error in *every* shape, ark's included, because it would
have to inherit from a file inside its own directory. Where the repository root is a
design root, `/.minispec/config.yaml` is that design root's configuration too; there
is no second file.

*Found 2026-08-07 while scoping sub-item #3, before any code was written.* It is the
rot this carve was opened about, committed inside the carve: a dated `DECIDED` and an
unmarked paragraph disagreeing, with the unmarked one reading as current because
nothing about it looked provisional. The only reason it surfaced is that implementing
the section forced the two to be read together.

**DECIDED (Bill, 2026-08-04): a repo-project's `.minispec.yaml` inherits from
`/.minispec/config.yaml`.** Settings shared across projects go in the repository
config once; a project's own file states only what differs. Where nothing differs, a
project needs no file at all.

**Two styles, and a project should pick one.**

- **A full-repo project** — the repository root is the design root. One project, one
  config.
- **Repo-projects** — no specs directory at the top level; each project has its own
  design root and its own `.minispec.yaml` inheriting the shared settings.

Mixing them — repo-projects *plus* the top level as a project in its own right — is
possible but not recommended: the top level's settings have to live in
`/.minispec/config.yaml`, which means every repo-project then has to override all of
them. The awkwardness is real, and the answer is to avoid the shape rather than
design around it.

*The tool should still be able to say which file a setting came from.* Inheritance
means the effective configuration is not what any single file says, so "why is this
value what it is" needs an answer that does not require reading two files and knowing
the precedence by heart.

Note this is why a top-level `.minispec.yaml` can be a flat error: in the mixed style
the top level's settings live in `/.minispec/config.yaml` like everything else, so
there is no legitimate use for the file the rule forbids.

**DECIDED (Bill, 2026-08-04): no `carves_dir` setting. `carves/` sits at the top of
the repository.** Its location was already mandated; removing the name override
makes it a fixed point, which is what lets it serve as a root marker at all — a
directory whose name a project could change could not identify anything.

Derived, not configured: `carves/done/` beneath it, the way `migrations/complete/`
derives today; and the **private carve directory as the dotted sibling** —
`.carves/` beside `carves/` — which is what ark already does and which
`check-ignore` can confirm rather than take on faith.

### What the siting decision rests on — measured 2026-08-04

`SpecsDir()` returns `<design root>/specs` unconditionally
([tool/internal/project/project.go](../../tool/internal/project/project.go)), so **"beside
the specs directory" and "the design root" name the same directory in every
project, by construction.** Confirmed on both surveyed projects.

**The axis that actually varies is repository root versus design root, and this
repository is the case that needs it.** The tool detects a design root by walking up
for `design/`, which finds **two** here — `tool/` and `example/`, each with its own
`design/`, `src/` and `specs/` — while `PENDING.md` sits at the repository root above
both. Run from that root, the tool resolves no project at all: *"no design/
directory found."*

So the structural fact is **trajectory scope ≠ project scope**: a repository holds
one work queue and may hold several mini-spec projects. Ark hides this because its
two roots coincide.

That carries an implementation consequence for Item 1 — trajectory commands need
their own root detection, walking up to `.git` rather than to `design/`, instead of
sharing the existing one.

Ark is unchanged by the decision, this repository becomes correct exactly as it
already stands, and the config surface stays empty. The only case it forbids — a
design root keeping its own separate queue — has no instance in either project.

### The same split already bites `check-version`, measured 2026-08-04

The trajectory files are not the only repo-scoped artifact the tool cannot see.
`.claude/` is repo-scoped too, and the version check looks for it in the wrong
place.

[tool/internal/cli/cli.go](../../tool/internal/cli/cli.go) builds two candidates for the
skill README — `<cwd>/.claude/…`, then `<home>/.claude/…`. **It checks the current
directory and then the home directory, and never walks up.** Observed directly:

| run from | README consulted |
|---|---|
| repository root | `…/mini-spec/.claude/skills/mini-spec/README.md` — the repo's own |
| `tool/` | `/home/deck/.claude/skills/mini-spec/README.md` — the *installed* copy |

So in the repository where the skill is authored, running the check from the
directory where the Go work happens validates the binary against the **installed**
skill rather than the source being edited.

**Amended 2026-08-07: this paragraph understated the finding twice, and the fix
measured both.** It said the two files are byte-identical "so nothing is currently
wrong," and predicted the failure would appear "the moment the repo's README is
bumped without reinstalling."

*Wrong about this repository, in the harmless direction.* `~/.claude/skills/mini-spec`
is a **symlink** to this repo's copy — same inode. They cannot diverge, so there is
nothing to reinstall and the predicted failure can never fire here.

*Wrong about the world, in the direction that matters.* Some projects keep their own
copy of the skill in the repository so the artifacts and the skill travel together
and the version is **pinned** — `~/work/microfts2` is one. There the old lookup fires
today, from any subdirectory, and it is the pin it defeats. Measured 2026-08-07 with
microfts2 vendored at 2.10.0 and the user-level skill at 2.11.0: run from
`microfts2/cmd/`, the released binary printed

```
ok: tool and skill both at 2.11.0
```

— green, against the wrong file, ignoring the pin entirely. The vendored copy exists
precisely to state a version, and a cwd-anchored lookup reads past it. That is a
confident wrong answer from the tool's own check, which is the failure class this
carve was opened over.

**LANDED (`8197c6c`, 2026-08-07)** as part of sub-item #1. The behavior now lives in
[tool/specs/repository-root.md](../../tool/specs/repository-root.md) with R107–R117 and
`crc-RepoRoot.md`; this section keeps only the diagnosis and the measurement.

*This is not a trajectory bug and does not belong to any item here.* It is recorded
because it is the same underlying fact from a second direction: **the tool has
exactly one notion of root, and repo-scoped artifacts do not fit it.** `.claude/`,
the queue, and `carves/` are all repo-scoped; `design/`, `specs/` and `src/` are
project-scoped. That is an argument for git-root detection as a capability the tool
lacks generally, not a knob invented for the trajectory files — which is the
strongest thing that can be said for option 3.

**Both defaults change `SKILL.md` prose, not just tool behavior**, so they land in
Item 8 with a version bump behind them.

### One listed parameter is not a project parameter at all

Part-key notation, measured across ark's seven carves: `Item N` in three, `#N` in
three, `Part B`/`Part C` in one — same project, same period. The skill already
implies this ("key parts however the document already does"), so the tool must
**parse all three forms** rather than be told one.

That is what answered question 2: the question was never which notation to
configure, but what the link *records* — and the answer is `<doc-path>#<literal
key>`, whatever the document wrote. It also feeds Item 9, which has to decide
whether the marker vocabulary (`LANDED`, `SENT`, `DEFERRED`, `NOT VERIFIED`, and
the `MOVED` coined in this carve's own status block) is a closed set or an open
one. A closed set is checkable; an open one is honest. Not both.

### Out of scope, but adjacent

`specs/` is hardcoded in the tool while `design_dir` and `src_dir` are configurable.
The trajectory layer neither widens that nor fixes it, since it ends up adding no
keys at all. Worth a separate decision, not this one.

### The refusal cases, enumerated

The tool can be missing two different things: the repository root, and the file the
command wants. They are not independent, and the dependence collapses most of the
combinations.

**Two entailments, both forced by the marker rules.** A `.git` ends the search, so
*root unknown implies no git anywhere on the upward path* — the `.gitignore` branch
cannot arise. A trajectory file also ends the search, so *root unknown implies the
trajectory files are missing too* — they are markers, and a found marker is a found
root.

*Amended 2026-08-04 by `track`.* A third case sat here — root known, file missing,
git present but the path not yet ignored — and it was the only one besides the root
question that needed a person, because someone had to answer whether the file should
be private. `track` answers that once at `init`, so the case is gone along with
`--no-ignore`.

| | root | file | git | re-run needs | human answer |
|---|---|---|---|---|---|
| 1 | known | exists | — | nothing | — |
| 2 | known | missing | any | `--create` | no |
| 3 | unknown | *necessarily* missing | *necessarily* absent | root confirmation | **yes** — is this the root? |

**Exactly one case needs a person**, and it is the root confirmation. Case 2 the
agent may simply re-run: the root is known, privacy is already settled by `track`,
and the creation was the point of the original invocation.

**Case 3 does both jobs from one answer.** The confirmation authorizes `carves/`,
which records the root; the original command authorizes the trajectory file. Nothing
else is asked, because nothing else is in question.

*Renumbered 2026-08-07.* This and the paragraph below said **Case 5** until today,
left behind when the `track` amendment deleted the cases between. Both always meant
the root confirmation, which is now case 3 — the stale-pointer failure this carve
argues about, committed in the carve itself.

**A sixth situation was recorded here and is now unreachable.** *Superseded
2026-08-07, kept as a record so the reasoning is not rediscovered.* It read: the file
exists but is *not* ignored — the state both surveyed projects were in before tonight
— and concluded that work should proceed under an advisory, since refusing would be
wrong. Two later decisions abolish the state it describes. Nothing but `init` runs
without a config, so a project can no longer carry trajectory files with no
declaration at all; and with `track` verified on every run, unignored files are a
mismatch that gripes and exits under `private-trajectory` and are simply correct under
`all`. The adoption gap it named is closed by `init` writing the ignore lines, not by
an advisory nobody has to act on.

**Case 3, answered "no".** The user is running from outside their project — most
likely above it, since the search only walks up and would have found a marker
below. The refusal should say so: run from inside the project. Creating anything
here would put a repository root in the wrong place, and `carves/` is durable enough
that the mistake would persist.

### The root confirmation needs no flag

**DECIDED (Bill, 2026-08-07): nothing is spelled, because the question dissolves.**
An `@undecided:` sat here asking whether the root confirmation should reuse
`--create` or take a flag of its own. Its premise was wrong on re-reading: root
detection cannot fail after `init`, and nothing but `init` runs without a config, so
**case 3 is reachable only by `init` itself.** A non-init command in that state hits
the no-config refusal instead. `--create` therefore lives on trajectory writes and the
root question lives on `init` — two flags on two verbs, never competing for one slot,
and never able to appear on the same command line.

*And `init` needs no confirmation flag either, because the assent already happened.*
The tool never runs `init` implicitly, the skill never tells an agent to run it, and
the only paths that reach it are a direct user request or the no-config crank handle —
whose entire purpose is obtaining exactly this yes. A flag would re-ask a question
answered one step earlier, and an unbuilt road needs no maintenance.

**The one residual risk is closed by naming the path, not by adding a flag.** The
user's yes is to *make this a mini-spec project*, not to *here specifically*. The
negative finding already sought by the crank handle catches standing **above** the
project — "this looks like a folder of projects" — but not standing **below** it, in a
tree with no VCS anywhere to walk up to. Rare, recoverable, and silent. So the refusal
**names the absolute path it is about to make the repository root**, which makes the
assent land on a stated location. The decision above already has it naming the working
directory; this makes that load-bearing rather than incidental, and it is the same
instinct as `carves/` declaring the root by artifact.

*Recorded because being wrong twice about one marker is itself evidence.* The first
reason for wanting a distinct flag was that the two situations could co-occur — false,
and the entailments in this section disprove it. The second was that only a naming
judgment remained — also false. Neither error was visible from the paragraph itself;
both needed the surrounding decisions read together, which is exactly the reading a
carve exists to make possible.

## Open questions

Checkboxes match the Status block, so one `grep '^- \[ \]'` spans both. The number lives
in the label rather than in markdown list position: prose elsewhere cites "question 2",
and a positional number silently moves when a question is inserted above it — the same
reason `Rn` and part keys are never renumbered. **Question 8 carries no checkbox**: it is
a standing constraint rather than a fork, so there is nothing to close, the distinction an
approved gap draws.

- [x] ~~**1. Project parameterization.**~~ — **answered 2026-08-04**, see the Item 1
  section. Convention over configuration won on the evidence: privacy became a
  universal convention rather than a setting, the prefix parameter was dropped, and
  siting became a mandate. The trajectory config surface is empty. What loose ends
  remain carry an `@undecided:` there rather than sitting here.
- [x] ~~**2. What identifies a "part"?**~~ — **answered 2026-08-13:
  `<doc-path>#<key>.<subkey>`.** The lean recorded here was right, and the reason it
  works is the one this question already half-stated: the fragment resolves inside the
  named document, so the key scheme is a per-document property and a bare `7` cannot
  be confused with a `7` from elsewhere. Measured across eight carves — three schemes
  live, none mixed within a document. See the Status section for the decision.

  **Half of this was superseded 2026-08-16: there is one key form, not three.** The
  fragment shape `<doc-path>#<key>.<subkey>` stands. What fell is *"the parser accepts
  every live form and records the literal key"*, and with it the per-document uniformity
  check this answer handed to Item 3. See the Status section.
- [x] ~~**3. Cardinality of the item↔part link.**~~ — **answered 2026-08-13, by
  normalization rather than by modelling.** Split parts into subparts until no part is
  completed by more than one item, so the awkward direction never arises; the parent
  completes when its subparts do, derived rather than stored. The other direction is
  deliberately left open — one item may still discharge parts in several documents —
  so the link is asymmetric: scalar on the part side, a list on the queue side. See
  the Status section.
- [x] ~~**4. Does the counter assertion survive?**~~ — **answered 2026-08-16: delete it.**
  Ratified by Bill; the rule was already written. `trajectory-format.md` states it under
  Item IDs: *"Do not write the next-free ID down. A copied counter is a second copy of the
  numbering state, and it goes stale silently."* With `next-id` landed (`#10`) the answer
  is also cheap to act on — the number is a query rather than a line to maintain.

  *Recorded because of how it was answered.* The rule was written during 9.1, hours before
  `next-id` existed, by someone (Daneel) who did not notice he was closing a listed open
  question. So the answer sat in a normative document while the question still read `- [ ]`
  open here, and one `grep '^- \[ \]'` would have reported work that was already decided.
  That is this document's own argument arriving from a third direction in one day: a claim
  expired when its document changed, a blocker outlived its own paragraph, and now an
  answer landed in one file while the question stayed open in another. All three are a
  fact recorded in prose, in a place nothing re-reads.
- [x] ~~**5. Should the tool ever edit trajectory files?**~~ — **answered 2026-08-13:
  yes** (Bill). `update check` already edits `design.md` checkboxes, and Items 6 and 7
  were designed on this assumption throughout — completion checks off the parts an item
  discharges and clears the current file, creation writes the lifecycle preamble. The
  decision ratifies what the document already assumed rather than changing anything.

  **The risk the question named does not go away, and now has somewhere to go.** These
  files are mostly prose, so a bad write has more to destroy here than a flipped
  checkbox in `design.md`. That is what makes **Item 5 a prerequisite rather than a
  convenience** — it already says so for Item 6, and this is the reason.

  The yes is bounded by what the items actually describe: **structural edits only** —
  checkboxes, markers, and entries the tool itself wrote. Nothing in this carve asks the
  tool to rewrite prose, and the scope of the answer is the scope of those designs.
- [x] ~~**6. Does `add-item` renumber or reuse?**~~ — **answered 2026-08-07: neither.**
  The question existed only because the agent was carrying an ID between two
  invocations. With the tool minting and writing in one act, there is no interval in
  which a second agent can take the same number, nothing to hold stable, and nothing
  to reuse. See the crank-handle section, where the verb that mints a requirement is
  decided too.
- [x] ~~**7. Where does crank-handle output go?**~~ — **answered 2026-08-13: stdout,
  as markdown** (Bill). The question's worry rested on a wrong premise: an agent
  invoking a CLI *does* receive stdout as the tool result, so the destination was never
  the risk. What remains of the worry is sequencing — a side-effect print during a batch
  may go unread — and that is a property of when a command is called, not of where it
  writes.

  **`minispec` is a command for agents, so markdown is the default output, not a
  presentation layer over some other form.** Baby Food: the model reads the output
  directly and never parses a format. Program-parseable output is the *exception* and
  already has its mechanism — `--json` is a global flag
  ([tool/internal/cli/cli.go](../../tool/internal/cli/cli.go), documented in
  [cli-commands.md](../../tool/specs/cli-commands.md)) — so nothing new is needed; the
  decision just names what it is an exception to.

  *Existing deviations are **Item 11**, added 2026-08-13.* Two `update` verbs put a bare
  minted value in the agent channel — `retire`'s `Tn` and `migration-complete`'s new
  path — and `retire`'s split is stated in R103, so conforming edits a requirement
  rather than an implementation.
- **8. Private carves must not leak.** *No checkbox: a standing constraint on every
  part that reads or reports, not a question with an answer.* Some projects keep a
  gitignored carve directory alongside the public one. The tool must read both to
  validate, and must never write a private carve's contents into anything that ships —
  a report, an error message, a summary file.
- [x] ~~**9. Does the format carry a version marker?**~~ — **answered 2026-08-14: no
  number, and absence is the version.** The question made itself conditional on migration
  being real rather than hypothetical, and the measurement says real but small and
  one-time: 15 carve files across both projects, three key notations, already retrofitted
  by hand once. A file the tool created carries the marker; a file without one is legacy,
  which is the whole signal a migration needs. A number in every file forever buys nothing
  until there is a second version to name — and that is when it earns its place. See the
  Item 9 section.
- [x] ~~**10. Does mini-spec run its own trajectory layer?**~~ — **answered 2026-08-11:
  yes.** The pending file and this carve directory arrived 2026-08-04; the done and
  current files followed on 2026-08-07, when the first item completed and there was
  nowhere to record it. All three exist and are in use, and no project prefix was ever
  wanted — the bare names are the mandated ones. The smell is closed: this project runs
  the layer it is building tooling for, and `#4` is the proof, since `init` refused to
  run here until this repository declared its own `track`.
