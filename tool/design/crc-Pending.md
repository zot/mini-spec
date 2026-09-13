# Pending
**Requirements:** R241, R242, R243, R244, R246, R247, R248, R252, R253, R256, R257, R262, R263, R264, R265, R267, R268, R269, R271, R272, R273, R274, R275, R276, R277, R278, R279, R280, R281, R245, R249, R250, R258, R259, R282, R330

The three verbs over the trajectory files — `pending add-item` and `pending finish`, which
write **both sides of the item↔part link**, and `pending start`, which opens the item. Package
`internal/pending`, and an orchestrator rather than a reader — it owns the *order* of the
writes and delegates every shape to whoever already owns it: the queue entry, the `## Active`
region and the done entry to [Trajectory](crc-Trajectory.md)'s adapters over the dependency's
readers, the part line to [Carve](crc-Carve.md)'s.

**It is the backup slot's first production caller** (R248). Reclaimed from `old-sdom` on
2026-09-05 (`#69`); the requirement map is in `.scratch/REVISE.md`.

## Knows
- The part pointer `<doc>#<part>`: the document and the key inside it, resolved against the
  repository root. It is the **queue side** of the link, and the only side that is written
  as a pointer at all (R242)
- **The link's asymmetry, which is structural rather than a choice.** Trajectory files are
  private in every project, so the public→private direction can never be a markdown link: a
  carve cannot point at a file a cloner does not have. The carve carries a **bare key** and
  the queue side holds the pointer, so there is one shape to build rather than one per
  project — and it is the half hand-maintenance cannot supply (R242)
- Nothing about file shapes. The queue entry, the `## Active` region and the done entry belong
  to [Trajectory](crc-Trajectory.md) over the dependency's readers, the part line to
  [Carve](crc-Carve.md), the format itself to the skill. This card names the sequence, never
  the syntax

## Does
- releasable(repoRoot, part): whether a part's queue ID belongs to the attempt the slot holds
  as reverted — a `REVERTED` marker while the slot state is reverted — the one case the
  one-item-per-part check lets through, because the release runs inside the coming mutation
  ahead of the new marker (R330)
- AddItem(from, entry, place): **mint the ID and write both sides in one invocation**
  (R241). Assignment and the write that records it are one act, which is what keeps R190's
  `max()` the whole truth — a number is in a document the moment it exists. A command that
  reserved a number for the agent to write later would be a second copy of the numbering
  state by construction, which is the stale-counter failure with the hand taken out of it
- **Write the whole entry, not half a heading** (R252). The `##` heading with its number,
  title, skill and one-line status; the `Source:` line; and the `Next:` line when one is
  given. The verb minted a two-line stub until 2026-08-21, so the house shape could not be
  produced by the verb at all and all ten live entries were hand-finished
- **A gap is a source too, distinguished by shape** (R271, R272). `--from O136` beside
  `--from carves/x.md#3`, with no flag, because a gap ID is a letter and digits carrying no
  `/`, no `#` and no `.md`. The grammar is one capital letter and digits, and a gap-shaped token that is not one gap ID —
  `O5-O6`, `O5,O6` — is refused **by the rule that an entry carries one pointer** rather than
  as a parse error (R272). *`ExpandIDRefs`, which the August tree reused here, returns with
  the gap selectors — carve Item 7 — and this grammar is what it replaces until then*
- *Why it earns a case at all:* repairing a gap is the second most common thing an item does
  and the verb refused it, so every such entry was typed in by hand and bypassed the placing
  verb — the hand edit this carve exists to eliminate, arriving through the verb it had
  already fixed
- **A gap source needs a design root; a part pointer does not** (R273). Gap IDs are scoped to
  a `design.md` and a repository may hold several roots — this one holds two — so the root is
  resolved as every other gap verb resolves it and its absence is a refusal naming what was
  looked for and where. A part pointer names its own document, which is why `pending` resolves
  no project for it
- **Nothing is written on the gap side** (R274). The part case writes both halves because the
  carve is public and the queue is private; a gap has no marker and gains none, since a
  queued-marker shape would be a second thing a gap line can say while that class is
  mid-migration. *The cost is recorded rather than hidden:* there is no gap analogue of the
  carve→queue cross-check, so a pointer at a resolved gap goes unreported
- **One gap per source** (R276), held in the scalar a part pointer fills. Further gaps are
  named in the entry's prose, as `#60` did — the same ruling the multi-part question carries,
  since growing the entry shape and the scalar together through the gap door would build for a
  consumer that has not arrived
- **Require the status sentence and not the `Next:` line** (R253), and the cut is principled
  rather than a taste about entry length: the status lives *inside* the heading this verb
  mints, so omitting it leaves the agent completing a line the tool has just written. *A tool
  may decline to write a line; it may not write two thirds of one.*
- **Place the entry where the caller said** (R256): `--next`, `--nth N`, `--after N` or
  `--last`, mutually exclusive, defaulting to `--last`. Which position an item deserves stays
  the caller's judgment; moving the block there without disturbing its neighbours is
  mechanics, and the verb used to conflate the two by appending and cranking out *"order it
  by intent"*
- **Resolve `--next` against the current file** (R257), which is one boolean and not a
  reading: `## Active` holding something other than its placeholder means a step is in
  progress, and the pending file's own rule that the top item is active turns that into
  position 1 or position 2
- **Refuse rather than guess** (R243), naming what was looked for and where: an unresolvable
  document, an unresolvable key, a document with no status block, or a part that
  **already carries a queue ID** — a part records exactly one item, and re-queuing it
  silently would leave the older pointer resolving to work it never described
- Finish(id, commit): the completion, **in the mandated order — source first** (R244). Each
  discharged part is checked off in its carve; then the current file's `## Active` section is
  reset; then the entry
  moves from the pending file to the done file. Source first because the carve is the copy a
  future reader trusts, and the one nobody thinks to check
- **A gap-sourced completion requires a decision** (R279, R280). `--resolve` or
  `--no-resolve`, refused together, and **neither is a refusal raised before anything is
  written** — a caller who has not decided pays a retry rather than a half-completed item.
  Resolving is never inferred: an item may address a gap **partly** — `#60` named three and
  closed two on purpose — and a gap wrongly marked resolved is work that silently never
  happens, with the entry that would report the loss being the one just closed
- *This replaced a notice, and the notice's own argument is what replaced it* (Bill,
  2026-08-26; `R279` retired as `T7`). That rule said the caller is told the gap is still open
  and left to act, reasoning that a flag one must remember to type is not a reminder — true,
  and the wrong conclusion. What a caller must not be able to do is **pass silently through the
  decision**, and only a refusal prevents that; giving both answers a spelling is what makes
  the question unskippable
- **`--no-resolve` is a record, not a shrug** (R281). A gap left open by decision and one left
  open by oversight are identical in `design.md` — the checkbox is unticked either way — and
  the completion is the only place that difference is known, so the report states which
  happened. Same move `NOT VERIFIED` makes rather than leaving a part unmarked. Nothing
  anywhere records that `#60`'s third gap was deliberate
- *What earns the flags their place in this verb rather than in `update resolve-gap` alone:*
  `--discharged` already brings the gap IDs here at completion time and lets them die as free
  text in a header. This makes a conversation the verb is already having actionable
- **The done entry carries the gap pointer** (R278), as `` Gap `<doc>#<gap>` `` in the shape
  `` Part `<doc>#<key>` `` uses, so the ledger keeps the link the pending file held instead of
  leaving it to whatever text `--discharged` happened to carry
- **Check the parts the item recorded and nothing else** (R246). A parent part is never
  checked directly — it completes when its subparts do, derived rather than stored — so no
  parent box is set and nothing is inferred from what a part's siblings mean. An item may
  discharge parts in several documents, and each is checked: a part records one item, an
  entry records a list of parts
- **Compose the done entry's header, never its body — and place a body it is handed**
  (R247, R262, R263). Identifiers, date, title, commit and part pointer are facts the tool
  was handed; *enough to reconstruct the change without re-reading the code* is a judgment
  about a future reader. The tool owns IDs, not prose — the same line the requirement-minting
  verb holds, and a tool inventing a summary would do the one part that is not mechanical,
  worse than the agent standing beside it.
  *What moved on 2026-08-24 is who types it in, which was never the authoring half.* The
  verb wrote a header and then told the agent to hand-edit the file it had just written —
  **nine consecutive completions**, into the region immediately beneath a line it had just
  placed. Composing stays with the agent; placing comes here, which is the split every other
  part of this carve already uses
- **Open an item: compose its identity line, place the caller's context** (R265, R267). The
  ID and title come from the queue entry rather than from the caller, so the active block and
  a later done header cannot disagree about what was worked. *The third verb, and the one
  edit in an item's round trip no command emitted* — made into a gitignored file **before any
  record of the work exists anywhere else**, which is why it runs inside the slot like the
  other two and has the strongest claim of the three to be there
- **Fill the identifier slot rather than half of it** (R268, R269). The `#N` is this card's
  because IDs are; whatever else the item discharged is the caller's, and the two are joined
  with the ` / ` the format mandates. *Half a field is a hand edit inside the line the verb
  just wrote* — measured four for four on 2026-08-24. **A colon in the slot is refused**,
  because the reader takes the slot as the run between the em dash and the colon that opens
  the title, so a colon inside it silently ends the slot early and every identifier after it
  stops being read. *A verb that owns a format refuses the input that would make its own
  reader wrong.* And the range is never inferred from `requirements.md`: a generated
  identifier list reads exactly like an authored one, which is R247's objection one field
  along
- **With no body, crank the hand edit and name the flag that avoids it** (R264). The crank
  handle is the fallback and never the design — *an untrusted system costs more than none,
  because you pay for the system and still carry everything yourself*
- **Ask Trajectory to clear the current file's `## Active` section**, and hold no opinion
  about how that region is found. The region is the dependency's `Current` and the two
  refusals are surfaced by Trajectory (R249, R250) — this card owns only *when* it happens,
  which is after the source and before the queue
- **Run inside the slot** (R248). Both verbs are one `Slot.Record` each, so the copy, the
  worktree anchor and the state transition cover the **whole invocation** rather than any
  single file — which is what makes a mis-typed part pointer a `pending revert` instead of a
  repair. A verb that wrote three files outside the slot would leave two of them unprotected
  in exactly the way the slot exists to prevent
- Crank everything out (R282): the number minted, the files written, the parts checked, the
  entry moved. Satisfied here rather than restated

## The completion write is the dependency's `Land`, and that is worth stating
The marker rule — replace the first transient, delete the remaining transients, append when
there is none, never touch a record — is the dependency's (R219), and `Land` composes the box,
the strike and the `LANDED` record in one act behind [Carve](crc-Carve.md)'s `SetPartLanded`
(R220, R245). This card hands it the attribution `` `<hash>`, <date> — `#N`. `` and nothing
else. *The case that proves the rule:* a part carrying `**NOT VERIFIED.**` **and**
`**OPEN (#N.)**` completes to `**NOT VERIFIED.** **LANDED (…)**` — the assessment untouched,
the queue state replaced — and that is the dependency's test to keep, not this one's.

## Collaborators
- Trajectory: reads and writes the queue entries, the `## Active` region and the done entry,
  as adapters over the dependency's readers
- Carve: resolves the part by key and performs the marker and completion writes on its line
- Backup.Slot: `Record` wraps each verb, and takes the worktree anchor with it
- Project: resolves the repository root, which is where all four files live

## Sequences
- seq-queue-item.md
