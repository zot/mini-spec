# Sequence: creating and completing a queue item

Two diagrams over [Pending](crc-Pending.md), and both are the same shape: one `Slot.Record`
wrapping every write, with the order inside it mandated rather than convenient.

## 1. `pending add-item --from <doc>#<part> "<title>" --status <text>`

Both sides of the link are written inside one invocation, so they cannot be created out of
step — which is the whole reason the verb exists.

```
1. CLI → Pending: AddItem(from, entry, place) — `entry` carries title, skill, status and next action; `place` carries the placement intent
1.1. Project resolves the repository root; the queue files and carves/ both live there
1.9. CLI resolves the three prose slots — title, status, next action — each supplied inline or from a file and never both, refusing a title wrapped end to end in a single `**…**` run
1.11. CLI resolves the placement intent: at most one of `--next`, `--nth N`, `--after N`, `--last`, and none of them means `--last`
1.12. Decide which kind of source `--from` names, by shape and with no flag: a letter and digits is a **gap ID**, anything else is a part pointer. A range or list is well-formed in that grammar and is refused, naming the one-pointer rule rather than the syntax
1.13. For a gap: resolve the design root as every gap verb does — refusing, when none does, by naming what was looked for and where — read `design.md`, and refuse an ID it holds no gap for. **Steps 1.3, 1.4 and 1.7 are skipped: there is no carve, no part to strand, and nothing written on the gap side**
1.2. Split the pointer at `#` into a document path and a part key, and resolve the document beneath the root
1.3. Carve reads that document and finds the part keyed by that key — refusing if the document, the key, or the key's form does not resolve
1.4. Refuse a part that already carries a queue ID: a part records exactly one item, so re-queuing would strand the older pointer
1.5. Trajectory reads both queue files and mints the next free ID as max() across them — the pending file alone is too low right after completions, the done file alone while the highest IDs are live
1.10. Trajectory resolves that intent to a position among the entries it can see
1.10.1. The dependency's `Pending` reader parses the pending file; an entry is a `## N.` heading **node**, so a `## 5.` quoted inside a fenced example is not one and cannot be counted or targeted
1.10.2. `--next` asks the current file whether a step is in progress — `## Active` holding something other than its placeholder — and resolves to position 1 or position 2
1.10.3. `--after N` resolves to the position after the entry whose item ID is N, refusing when no live entry carries it
1.10.4. `--nth N` is refused outside `1 … entries+1`, and refused at 1 while a step is in progress, naming both repairs
1.6. Slot.Record wraps everything below: the trajectory files are copied, the worktree anchor is taken, and the state becomes `changed`
1.6.1. Trajectory renders the **whole** entry — heading, status sentence, `Source: <doc>#<part>` and the `Next:` line when given — and places it at the resolved position by replacing the node it lands in front of, so no other node's bytes can move
1.6.2. Carve writes `**OPEN (#N.)**` onto the part line through the marker rule, which replaces the first transient, deletes the rest, and never touches a record
1.7. Announce a reused number if this one returned to the pool from an unfinished attempt — reaching the one person who could have carried it into a commit message
1.8. Crank out the minted number, the position the entry took, and every file written — naming the placement flags rather than instructing a move by hand
```

**The numbers run out of sequence, and that is the discipline rather than a lapse.** Steps
are appended and never renumbered, because renumbering orphans every code comment pinned to
an ID — so position is the order of execution and the number is only the anchor. `1.9`, `1.10`
and `1.11` arrived in August 2026; `1.12` and `1.13` with the gap source.

## 2. `pending finish <N>`

**The source first**, because the carve is the copy a future reader trusts and the one nobody
thinks to check.

```
2. CLI → Pending: Finish(id, commit)
2.1. Trajectory reads the pending file and finds the entry for N, refusing if it is absent or already completed
2.2.1. For a **gap**-sourced entry, refuse unless the caller decided — `--resolve` or `--no-resolve`, never both, and neither is a refusal named **before the slot opens**, so an undecided caller pays a retry rather than a half-completed item   // R279, R280
2.2. Collect the parts the entry recorded — a list, since one item may discharge parts in several documents
2.3. Slot.Record wraps everything below, exactly as in 1.6
2.3.1. For each recorded part, Carve performs the completion write on its line: checkbox to `[x]`, title struck through, and a `LANDED (date — `#N`.)` record — no hash, the item number is the identifier (R477) — replacing the transient
2.3.2. Nothing else is checked — a parent part completes when its subparts do, derived rather than stored, so no parent box is set and no sibling is consulted
2.3.3. Reset the current file's `## Active` section — that section and nothing else
2.3.3.1. The dependency's `Current` reader parses the file and locates the `## Active` heading node, refusing when there is none or more than one; Trajectory names the repair
2.3.3.2. The region is enumerated by the reader: every node after that heading, up to the next heading of the same level or higher
2.3.3.3. `Current.Reset` removes the region's nodes and inserts the format's placeholder as one text — a removal and one insert, so nothing outside the enumeration can move
2.3.3.4. *Retired with the August reader:* the render is no longer re-parsed and compared, because the region cannot include a standing heading by construction — see the spec
2.3.3.5. Trajectory writes the file by temp-file-and-rename
2.3.4. Trajectory removes the entry from the pending file and appends the done entry's **header** — date, identifiers, title, commit, part pointer The body, when the caller supplied one, is spliced in the **same write** rather than found again afterwards, so the completion has no second target to anchor against. // R263 The slot carries the queue ID the tool owns joined to whatever else the caller says the item discharged, and a colon in that text is refused before anything is written. // R268, R269 Both writes are node-addressed: the entry is the heading node that opens it plus the nodes beneath it, and the ledger's rule is found in a text run's own content, so a `## N.` or a `---` quoted inside a fence can neither bound the removal nor receive the record. // R270, R270, R270
2.3.5. For a **gap**-sourced entry: with `--resolve`, Update resolves the gap in the document the **entry** named, refusing a different design root; with `--no-resolve` it is left open. Never inferred — an item may address a gap only partly, and a gap wrongly resolved is work that silently never happens   // R277
2.5. Report which of the two happened: a gap resolved, or one **left open by decision**. The two are identical in design.md, and this completion is the only place the difference is known   // R281
2.4. Crank out every part checked and every file written, and name where the done entry's **body** goes — that is authoring, and the tool does not author
3. CLI → Pending: Start(id, context) — the third verb, and the only one that writes before any record of the work exists elsewhere    // R265
3.1. CLI resolves the context slot — `--context` or `--context-file`, refused together, read byte for byte                              // R265
3.2. Trajectory reads the pending file and takes the entry's ID and title, so the identity line and a later done header cannot disagree   // R265
3.3. Backup opens the slot: the same one call per invocation as steps 1.6 and 2.3, since this write is a mutation like any other        // R267
3.3.1. Trajectory finds `## Active` as a heading node and refuses a section that already holds an item, naming the parking repair       // R266
3.3.2. Trajectory writes the identity line and the caller's context through the **same path** that clears the section, so the standing-context guard and the outside-unmoved assertion are inherited rather than restated   // R267
3.4. Crank out the file written and the item now active
```

**Step 2.3.1 is three markings in one write, and all three are required.** The checkbox is
what a machine reads, the strikethrough is what a skimmer reads, and the record is what
provenance reads — the format requires them to agree, and a landed part's queue ID lives in
that record and nowhere else, so a completion that cleared the transient without appending
would leave the carve→queue checks reading a line that cites nothing.

**Step 2.3.3 is where the region has to be *addressed* rather than inferred, and 2.3.3.4 is
what makes that checkable.** The shape it replaces defined the active item as *everything
below the `---` rule*, which is a definition rather than an address: correctly implemented,
and wrong about a document holding standing context down there. Locating a heading node
cannot run past the next heading, and the reader's write removes only what it enumerated. The
August tree also re-parsed the render and compared the nodes outside the region; that check
derived *outside* from the computation it was checking, and is not carried.

**Step 1.6 and step 2.3 are the same call, and that is the point.** One `Slot.Record` per
invocation rather than one per file: a verb writing three files outside the slot would leave
two of them unprotected in exactly the way the slot exists to prevent, and the undo would
cover a fragment of the change rather than the change.

## 4. `pending commit-message [--amend]`

4. `Compose(root, amend)`
   4.1. Read the done file; every entry with its line
   4.2. Ask git for every `#N` named by a message on `HEAD`'s history, bounded by a non-digit;
        no commits names nothing
   4.3. Keep the newest entries down to, and excluding, the first a commit names — a hash in
        its slot, or every identifier in a message — oldest first; none is a refusal naming
        the newest entry and that it is named
   4.4. Compose: subject `#N, #M: title; title`; body `Items #N, #M.` then each entry's
        `#N — title` and the done file's lines beneath its header
   4.5. With `amend`: refuse when `HEAD` is in a remote branch; else `HEAD`'s message
        unchanged, a blank line, `Also lands #N.`, the entries
   4.6. Print, or write `--out` byte for byte
