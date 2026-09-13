# Test Design: Pending
**Source:** crc-Pending.md, crc-Trajectory.md, crc-Carve.md

**Reclaimed from `old-sdom` on 2026-09-05 (`#69`), over the dependency's readers.** Every
case below that the August tree carried is carried here with its requirement numbers remapped
(the map is in `.scratch/REVISE.md`); every `**Pulled:**` line was dropped in the port, because
a pull recorded against the August code is a claim about this one, and the injections were
re-run on this tree before being written back.

**Two of the August cases are gone rather than carried**, because their subjects moved into
the dependency: *checking off a part is three markings at once* and *a record is superseded
only when it is a transient* are `minispecsdom.Carve.Land`'s and the marker rule's (R219,
R220), tested where they live. The adapter's own guarantee — atomic write, byte-identical on
refusal — is in `test-Carve.md`.

**One case skipped for half a day, and says so.** The add-then-finish round trip was not
byte-identical on the dependency's `Place`/`Remove` pair (gaps `O14`, `O15`); the test skipped
naming them, and was un-skipped the same afternoon when their fix landed (mini-spec-tool #29).

**Pulled on this tree 2026-09-05, all thirty, with a probe past the list.** Twenty-eight rang
first time; one (the slot transaction) needed its test strengthened with a revert witness and
then rang; one (the parent part) cannot be reached from this side and says so. The probe past
the list — write the done entry *before* removing the pending entry in `CompleteItem` — rang,
though on a weak witness (`got 2 entries, want 1`, because the ordinary write then ran too);
nothing asserts the pending-first order directly, and that is left as a known thin spot rather
than a gap, since the dependency's `Remove` refusal is the only failure the order guards.

**Most of what these verbs guard is invisible when violated**, which is where the injections
belong. A wrong number is loud and any test catches it. A wrong **order** is silent: the work
still happens, the files still change, and a completion that wrote the queue before the carve
looks exactly like one that did not — until the process dies in between. Same for a write that
escapes the slot, a parent box that gets checked, and a record that gets overwritten. Those
are what the cases below spend their injections on.

## Test: both sides of the link are written, or neither is
**Purpose:** creation writes the queue entry and the part marker in one invocation, so the two cannot be created out of step (R241)
**Input:** a temp repository with a carve holding an unqueued part; run `add-item --from carves/x.md#3 "a title"`
**Expected:** the pending file gains an entry whose `Source:` names `carves/x.md#3`; the part line gains `**OPEN (#N.)**`; and the number cranked out is the one in both
**Refs:** crc-Pending.md, seq-queue-item.md#1.6
**Code:** internal/pending/pending_test.go
**Alarm:** 1
**Fire alarm:** write the queue entry and skip the marker. Goes red on the part line, and it is the failure the verb exists to prevent: a queue entry pointing at a part that does not know it is queued reads as correct from the queue side, which is the side anyone checks first
**Inject:** internal/pending/pending.go:mintAndPlace
**Pulled:** 2026-09-05 — rang, `the part line is missing its marker, or lost its record`; the hand-removed-entry test rang with it. Injection is in `mintAndPlace`, where the marker write lives now — Inject re-sited

## Test: a mint and its write are one act
**Purpose:** the number is in a document the moment it exists, so `max()` over the queue files stays the whole truth (R241)
**Input:** a repository whose highest ID is in the **done** file rather than the pending one; mint twice in succession
**Expected:** the second mint is one higher than the first, and every minted number is present in a file before the command returns
**Refs:** crc-Pending.md, seq-queue-item.md#1.5
**Code:** internal/pending/pending_test.go
**Alarm:** 2
**Fire alarm:** mint from the pending file alone. Goes red on the fixture, whose maximum lives in the done file — the R190 collision, and the shape that made `next-id item` answer `DONE.md 0` correctly-by-luck in ark on 2026-08-16, right up until a high-numbered item completed
**Inject:** internal/pending/pending.go:mintAndPlace
**Pulled:** 2026-09-05 — rang, `minted #4, want #10 — the maximum is in the done file`. Site is `mintAndPlace` — Inject re-sited

## Test: an unresolvable part is refused, and the refusal names what it looked for
**Purpose:** refuse rather than guess (R243)
**Input:** four pointers — a document that does not exist, a key no part carries, a key on the superseded bare-`#N` scheme, and a well-formed pointer to a part that **already carries a queue ID**
**Expected:** four refusals, each naming the document and the key; and **nothing written** in any of the four — no entry, no marker, no ID consumed
**Refs:** crc-Pending.md, seq-queue-item.md#1.4
**Code:** internal/pending/pending_test.go
**Alarm:** 3
**Fire alarm:** drop the already-queued check. Goes red with a second entry pointing at the same part — the state that leaves the *older* pointer resolving to work it never described, which nothing downstream can detect because both entries are individually well-formed
**Inject:** internal/pending/pending.go:AddItem
**Pulled:** 2026-09-05 — rang, `expected a refusal` on the already-queued subtest

## Test: completion writes the source first
**Purpose:** the carve is the copy a future reader trusts, and the one nobody thinks to check (R244)
**Input:** a completion whose queue write is made to fail after the carve write succeeds
**Expected:** the carve carries the completed part line, and the failure is reported — the *recoverable* order, since the queue files are the ones the slot can restore
**Refs:** crc-Pending.md, seq-queue-item.md#2.3
**Code:** internal/pending/pending_test.go
**Alarm:** 4
**Fire alarm:** reorder so the queue moves first. **This one passes by default and is invisible when violated** — both orders produce identical files on the happy path, and only an interruption between them tells the difference. Injecting the swap and failing the second write leaves a part still open against an item already in the done file, which is exactly the orphan R242 reports and which no green suite would have shown
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang on four tests — `TestASourceFailureLeavesTheQueueUntouched` and three whose second `CompleteItem` then found no entry (`no entry for #4`). Injection: a `CompleteItem` hoisted above the parts loop

## Test: a parent part is never checked
**Purpose:** a parent completes when its subparts do — derived, never stored (R246)
**Input:** a carve holding a `SPLIT` parent with no checkbox and two subparts `8.1` and `8.2`; an item recording only `8.2`
**Expected:** `8.2` is checked; `8.1` is untouched; the parent line is byte-identical
**Refs:** crc-Pending.md, seq-queue-item.md#2.3.2
**Code:** internal/pending/pending_test.go
**Alarm:** 7
**Fire alarm:** check the parent when all its subparts are done. Goes red on the parent line — and the argument is that a parent box would be a **second copy** of a fact the subparts already carry, so the two can disagree and nothing says which is right. A parent with no checkbox *means* something, and inventing a state for it is the tool asserting what the document declined to say
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — **did not ring, and that is the finding again.** The injection — land the parent derived from a subpart key — errors with the dependency's `ErrNoPart`: a checkbox-less `SPLIT` parent is not a part to `minispecsdom.Carve` at all (the probe lists `Item 1`, `Item 7`, `8.1`, `8.2`), so the guard is the reader's by construction and nothing on this side can reach the property

## Test: an item discharging parts in several documents checks each
**Purpose:** the cardinality is asymmetric on purpose — a part records one item, an entry records a list (R246)
**Input:** an entry recording parts in two different carves
**Expected:** both parts completed, both files written, one entry moved
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
*No fire alarm, because there is nothing to inject into: this case is **deferred** and unwritten — the reader records at most one part per entry (the dependency's `pending-schema.md` keeps the door open as its Item 7).* The reader records at most one part, so the loop cannot be exercised
past one — and the census said so, reporting this as `unanchored` the moment the package
became visible to git. That is the distinction between a prescription and a *claim* doing its
job: an alarm with no site is not a weaker alarm, it is not one. When the reader records several, the
injection is *stop after the first part*, and this test is what will fail first.

## Test: the whole invocation is one slot transaction
**Purpose:** one level of undo covers the change rather than a fragment of it (R248)
**Input:** a completion touching three files, with the last write made to fail
**Expected:** `pending revert` restores all three trajectory files to their pre-invocation state, and the slot reports `changed` exactly once
**Refs:** crc-Pending.md, seq-queue-item.md#2.3
**Code:** internal/pending/pending_test.go
**Alarm:** 8
**Fire alarm:** wrap each file write in its own `Slot.Record`. **Silent on the happy path** — the files end up identical — and red here, where a revert undoes only the last of three writes and leaves the queue in a state no hand edit produced. This is the slot's first production caller, so nothing before this test has ever exercised the distinction
**Inject:** internal/pending/pending.go:mintAndPlace
**Pulled:** 2026-09-05 — **silent first, then rang.** The August test stopped at the stamp and could not tell one Record from three; a revert is the only witness — with a Record per write the second backup already holds the entry. Test strengthened with the revert, then rang: `revert left the new entry in place — the slot covered a fragment of the invocation`. Inject re-sited to `mintAndPlace`

## Test: the done entry carries a header and no body
**Purpose:** the tool owns IDs, not prose (R247)
**Input:** a completion with a commit hash and a title
**Expected:** an entry header carrying the date, the identifiers, the title, the commit and the part pointer — and **nothing after it** but the crank handle naming where the body goes
**Refs:** crc-Pending.md, crc-Trajectory.md, seq-queue-item.md#2.4
**Code:** internal/pending/pending_test.go
**Alarm:** 9
**Fire alarm:** compose a body from the title and the part list. Goes red on the entry's shape, and the reason is not fastidiousness: a generated summary reads exactly like an authored one, so the file's stated purpose — *enough to reconstruct the change without re-reading the code* — quietly becomes a claim nobody made
**Inject:** internal/pending/pending.go:doneHeader
**Pulled:** 2026-09-05 — rang on three tests, `the entry carries a body the tool wrote`

## Test: flags are parsed wherever they sit among the positional arguments
**Purpose:** an argument order the usage line advertises must not corrupt what it parses (R241, R244)
**Input:** `--skill` before, between, after and absent from a two-word positional, for both verbs
**Expected:** the positional comes back whole and the flag is seen, in every order
**Refs:** crc-CLI.md, seq-queue-item.md#1
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 10
**Fire alarm:** stop lifting positionals out and re-parsing the remainder — return at the first one, which is Go's own behaviour
**Inject:** internal/cli/cli.go:parseFlagsAnywhere
**Pulled:** 2026-09-05 — rang, `positional = "a --skill x title", want "a title" — a flag was folded into it`

*The sharp part is that this trap was **already found, fixed and commented** in `finish` an
hour earlier, and the fix was not carried across.* That is this repository's own recorded
lesson arriving inside the session that recorded it — knowing a failure mode is not a
substitute for a mechanism that holds it. The repair is therefore **one shared function**
rather than a second guard: the boilerplate was not worth factoring, and the *behaviour* was.

## Test: this project's own completion, end to end
**Purpose:** the verb reproduces by command what was done by hand, which is the only test that measures the item's actual claim (R244, R245)
**Input:** a fixture repository reconstructed from this repository's state before `#20` completed — the carve part open and queued, the entry in the pending file, the current file populated
**Expected:** the four files afterwards match what the hand edits produced, modulo the done entry's body
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go

*No fire alarm on the last case, and deliberately.* It guards the composition rather than any
one property, so every injection that would ring here rings more precisely somewhere above.
What it adds is that the pieces compose into the thing that was actually done by hand on
2026-08-18 — five edits across four files, three wrong on the first attempt.

## Test: a completion clears the active section and leaves the rest of the current file alone
**Purpose:** the completion **calls** the reset, and what survives it is the preamble *and* the standing context below the active section (R244, R249)
**Input:** a current file shaped as the format has it — a preamble, a `---` rule, an active section carrying a `###` of its own, and two standing `##` sections after it
**Expected:** the active context and its `###` gone, the section reading `_No active item._`, and the preamble and both standing sections byte-identical
**Refs:** crc-Pending.md, crc-Trajectory.md, seq-queue-item.md#2.3.3
**Code:** internal/pending/pending_test.go
**Alarm:** 11
**Fire alarm:** **delete the `parser.ResetCurrent` call from `Finish`** — a *wiring* injection, and deliberately not one inside the reset itself. Every alarm on the reset's own behaviour lives in `test-Trajectory.md` and every one of them keeps ringing with the call gone, which is the `O23` shape: no number of injections inside a unit proves it is reached
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang, `the current file kept its context — it is a resume buffer, never a log`, and the `###` survived

*What this case used to be, kept because the sequence is the argument.* It read "the current file's preamble survives a completion", and its alarm was the real first-implementation defect: `finish` wrote a hardcoded template over the **whole file**, which no test could catch because the fixture's current file was a bare `# Current` line with nothing to lose. The fixture gained a preamble that morning — and the file *still* had nothing to lose in the place the loss actually happened, because everything below the rule was the active item by definition and the verb was entitled to it. **320 lines of standing context went that afternoon.** A fixture contains only what its author thought to include, twice over; it now carries standing sections after the active one, which is where a real project keeps them

## Test: the reset's refusal reaches the caller
**Purpose:** a refusal from the reset fails the whole completion rather than being swallowed inside it (R250)
**Input:** a current file with no `## Active` heading, and a completion run against it
**Expected:** `Finish` returns the refusal, naming the missing heading and the repair
**Refs:** crc-Pending.md, crc-Trajectory.md
**Code:** internal/pending/pending_test.go
**Alarm:** 12
**Fire alarm:** discard the reset's error — `_ = parser.ResetCurrent(…)`. Also a wiring injection: the refusal itself is proven in `test-Trajectory.md`, and this proves it is not dropped on the floor between there and here. A swallowed refusal is the worst of both, since the completion reports success over a file it could not read
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang, `expected a refusal — the tool cannot tell the active item from the standing context`

## Test: `--next` means next to be worked, not position 1
**Purpose:** with nothing in progress the entry goes to the top; with a step in progress it goes immediately after the active item, which is position 2 by the pending file's own ordering rule (R257)
**Input:** the same queue twice — once with `## Active` holding the placeholder, once with it holding an active block
**Expected:** position 1 in the first case and position 2 in the second, from one unchanged command
**Refs:** crc-Pending.md, seq-queue-item.md#1.10.2
**Code:** internal/pending/pending_test.go
**Alarm:** 13
**Fire alarm:** resolve `--next` to position 1 unconditionally — the `--first` this flag replaced. Red reads *"the new entry displaced the item being worked, making itself active"*, which is exactly the objection `--next` was designed to dissolve
**Inject:** internal/pending/pending.go:resolveNext
**Pulled:** 2026-09-05 — rang, `with a step in progress --next placed at 1, want 2 — it displaced the active item`

## Test: `--nth 1` is refused while a step is in progress
**Purpose:** position 1 is the active item's slot, so `--nth 1` is the same refusal `--next` embodies from the other side, and the message names both repairs (R258)
**Input:** a current file with an active block, and `add-item --nth 1`
**Expected:** a refusal naming `--next` *and* parking the active item first; nothing written
**Refs:** crc-Pending.md, seq-queue-item.md#1.10.4
**Code:** internal/pending/pending_test.go
**Alarm:** 14
**Fire alarm:** name only `--next` in the message. Red reads *"the refusal offered one repair, so preempting the active item reads as forbidden"* — a real thing to want, refused by omission rather than by decision
**Inject:** internal/pending/pending.go:resolvePlacement
**Pulled:** 2026-09-05 — rang, `the refusal never mentions "park", so one repair reads as forbidden`

## Test: a prose slot given twice is refused
**Purpose:** each slot is supplied inline **or** from a file and never both, refused rather than resolved by precedence (R254)
**Input:** `--status` and `--status-file` together
**Expected:** a refusal naming the slot; nothing written, and no file read
**Refs:** crc-CLI.md, seq-queue-item.md#1.9
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 15
**Fire alarm:** let the file form win silently. Red reads *"both were given and one was used with nothing reporting the other"* — a precedence rule is a silent choice between two things the caller believed were one
**Inject:** internal/cli/pending.go:resolveSlot
**Pulled:** 2026-09-05 — rang, `both forms were accepted and "from the file" was used with nothing reporting the other`

## Test: the status sentence is required
**Purpose:** the status lives inside the heading the verb mints, so a missing one would leave the agent finishing a line the tool just wrote (R253)
**Input:** `add-item` with no `--status` and no `--status-file`
**Expected:** a refusal naming the slot and its file form; nothing written
**Refs:** crc-CLI.md, crc-Pending.md, seq-queue-item.md#1.9
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 16
**Fire alarm:** default the status to the empty string. Red reads *"the heading stopped after the skill and the verb reported success"* — the original defect restored as a default, which is how a required field quietly becomes optional
**Inject:** internal/cli/pending.go:runAddItem
**Pulled:** 2026-09-12 — rang again after `#31` threaded the repository root through `reportFinished` and labelled the written line; same injection and signature; restore byte-clean by copy. Previously 2026-09-05 — rang, and the red text is the point: the refusal never names `--status`, because the verb got as far as the carve and failed there instead

## Test: a body-file slot is read byte for byte
**Purpose:** the file forms exist so prose carrying backticks reaches the document intact (R254)
**Input:** a status file whose text contains a backticked identifier and an apostrophe
**Expected:** the heading carries those bytes unchanged
**Refs:** crc-CLI.md, seq-queue-item.md#1.9
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 17
**Fire alarm:** keep the heredoc's trailing newline — trim nothing at all. Red reads *"the status came back with a line break on the end, which lands mid-heading in the line the verb mints"*. The trailing newline is the shell's, not the prose's, and it is the **only** byte that comes off: everything the file says reaches the document unaltered
**Inject:** internal/cli/pending.go:readSlot
**Pulled:** 2026-09-05 — rang, `the file's bytes came back altered`. The trim lives in `readSlot` — Inject re-sited

## Test: the placement flags are mutually exclusive
**Purpose:** giving two placements is refused rather than resolved by precedence, and every single flag resolves to the intent it names (R256)
**Input:** `--next --nth 3`; `--last --after 4`; and each flag alone, plus no flag at all
**Expected:** both pairs refused; no flag resolves to `--last`, and each single flag to its own intent
**Refs:** crc-CLI.md, seq-queue-item.md#1.11
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 18
**Fire alarm:** let the last flag given win. Red reads *"`--next --nth 3` was accepted"* — a precedence between placements silently puts the entry somewhere the caller did not ask for, in a file with no diff
**Inject:** internal/cli/pending.go:placeFrom
**Pulled:** 2026-09-05 — rang, `--next and --nth were both accepted`

## Test: start writes the identity line and the caller's context
**Purpose:** the ID and title come from the queue entry, the context goes beneath byte for byte, and the standing context survives (R265, R267)
**Input:** a queue holding `#4`; `Start(root, 4, "context with a `code span`")` on a current file whose `## Active` holds the placeholder
**Expected:** the region reads `` `#4` — <title> `` then a blank line then the context; the preamble and both standing sections byte-identical
**Refs:** crc-Pending.md, crc-Trajectory.md, seq-queue-item.md#3
**Code:** internal/pending/pending_test.go
**Alarm:** 29
**Fire alarm:** take the title from the caller's context instead of the entry — compose the line as `` `#N` — <context> ``. Red as `the identity line does not carry the entry's title`, which is the active block and a later done header disagreeing about what was worked
**Inject:** internal/pending/pending.go:Start
**Pulled:** 2026-09-05 — rang, `the active section is missing "`#4` — a part to queue"`

## Test: start refuses an occupied active section
**Purpose:** overwriting the held item is the 320 lines one level up, in the file nothing reports a loss in (R266)
**Input:** `Start` on the fixture's current file, whose `## Active` already holds an item
**Expected:** a refusal naming the parking repair; the current file byte-identical
**Refs:** crc-Pending.md, crc-Trajectory.md, seq-queue-item.md#3.3.1
**Code:** internal/pending/pending_test.go
**Alarm:** 30
**Fire alarm:** reset the region before setting it — call `Reset` then `SetActive` in the adapter, so the occupied guard sees an empty region. Red as `an occupied active section was overwritten`
**Inject:** internal/parser/trajectory.go:SetActive
**Pulled:** 2026-09-05 — rang, `an occupied `## Active` was overwritten without complaint` — a `Reset` before `SetActive` in the adapter walks past the reader's guard

## Test: the identifier slot carries what the item discharged
**Purpose:** the tool writes the `#N` and joins the caller's identifiers with the ` / ` the format mandates (R268)
**Input:** `Finish` with `Discharged: "R476–R477"`
**Expected:** the done header reads `— #N / R476–R477:`; with no `Discharged` it reads `— #N:`
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 31
**Fire alarm:** drop the ` / ` join and write only `#N` — the half-filled field this requirement exists to end. Red as `the identifier slot lost what the item discharged`
**Inject:** internal/pending/pending.go:doneHeader
**Pulled:** 2026-09-05 — rang, `the identifier slot was written half-filled`

## Test: a colon in the identifier slot is refused
**Purpose:** a colon inside the slot ends it early for the reader, silently (R269)
**Input:** `Finish` with `Discharged: "R1: a note"`
**Expected:** a refusal naming the colon, before anything is written — `DONE.md` untouched
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 32
**Fire alarm:** delete the colon guard in `Finish`. The header writes, the reader's slot ends at the injected colon, and every identifier after it stops being read while the header still parses. Red as `a colon in the discharged slot was accepted`
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang, `a colon in the slot was written, and every identifier after it stops being read`

## Test: a gap is a source and nothing is written on its side
**Purpose:** the part case writes both halves because the carve is public; a gap has no marker
and gains none (R271, R274)
**Input:** a repository with a design root holding `O5`, and `add-item --from O5`
**Expected:** the entry carries `` gap `O5` `` and never `` part `#O5` ``; `design.md` comes back
**byte for byte**; `design.md` is not among the files reported written
**Refs:** crc-Pending.md, crc-Trajectory.md
**Code:** internal/pending/pending_test.go
**Alarm:** 19
**Fire alarm:** drop the `SourceKind` guard in `mintAndPlace` so the marker write runs for a gap
too. `SetMarker` then edits `design.md` looking for a part line, and the byte-comparison is the
only assertion that objects — a stray marker in a 145-gap document is not something a reader
finds. Red as `the gap side was written to`
**Inject:** internal/pending/pending.go:mintAndPlace
**Pulled:** 2026-09-05 — rang, and as in August **not on the predicted assertion**: `SetMarker` against `design.md` errors `no part with that key`, so four gap tests fail on the error rather than on the byte comparison

## Test: a gap range is refused by the one-pointer rule
**Purpose:** a range is well-formed in the ID grammar and is not a source, so the refusal names
the **rule** rather than the syntax (R272, R276)
**Input:** `add-item --from O5-O6`
**Expected:** an error naming *one pointer*; no entry placed
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 20
**Fire alarm:** delete the `len(ids) > 1` branch in `gapSource`, so a range takes `ids[0]` and
the rest are dropped. **Nothing errors and an entry appears**, pointing at the first gap while
the caller believes it named several — a silent narrowing of what the item claims to repair.
Red as `a range was accepted as a source`
**Inject:** internal/pending/pending.go:gapSource
**Pulled:** 2026-09-05 — rang, `the refusal does not name the rule` — the range fell through to the not-a-gap-ID message

## Test: a gap source with no design root is refused
**Purpose:** gap IDs are scoped to a `design.md` and a repository may hold several roots, so the
root is resolved rather than searched for (R273)
**Input:** `add-item --from O5` in a tree with no design root
**Expected:** an error naming both the design root and `O5`
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go

## Test: an unknown gap is refused
**Purpose:** an ID the document holds no gap for is a refusal, not an entry pointing nowhere
(R271)
**Input:** `add-item --from O99` against a design root holding `O5` and `O6`
**Expected:** an error naming `O99`; the pending file free of it
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go

## Test: a gap-sourced entry round-trips and is not a part
**Purpose:** the pointer reads back, and `Parts()` stays empty or completion would mark a part
landed inside `design.md` (R271, R275)
**Input:** the entry `add-item --from O5` wrote, read back through `PendingEntries`
**Expected:** `Parts()` empty, `Gap()` answering `O5`
**Refs:** crc-Trajectory.md
**Code:** internal/pending/pending_test.go
**Alarm:** 21
**Fire alarm:** drop the `SourceKind == SourceGap` test from `Parts`, so a gap reads back as a
part. `finish` then opens `design.md` as a carve and looks for a part keyed `O5`. Red as
`a gap read back as 1 part(s)`
**Inject:** internal/parser/trajectory.go:QueueEntry.Parts
**Pulled:** 2026-09-05 — rang, `a gap read back as 1 part(s) — completion would mark a part landed in design.md`

## Test: finish requires a decision, and the ledger keeps the pointer
**Purpose:** neither flag is a refusal before anything is written, and the done entry carries the
link the pending file held (R277, R278, R279)
**Input:** a gap-sourced item completed first with no decision, then with `DeclineResolve`
**Expected:** the first refused, naming both spellings, with `DONE.md` untouched; the second
leaving the gap unresolved, `design.md` byte-identical, and the done entry carrying
`` Gap `design/design.md#O5` ``
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 22
**Fire alarm:** two, and the second is the one that matters. Drop the `hasGap` clause in
`doneHeader` so the pointer is never written — the ledger then records a completion whose source
survives only in whatever free text `--discharged` carried, which is a record by luck and reads
exactly like a record. Red as `the done entry lost the gap pointer`
**Inject:** internal/pending/pending.go:doneHeader
**Pulled:** 2026-09-05 — rang, `the done entry lost the gap pointer`. *First attempt landed on the wrong anchor* — `&& false` on the file's first `gap.Kind == SourceGap`, which is `AddItem`'s dispatch, and the four gap tests failed as part-pointer refusals; that rings and proves nothing about the header. Re-anchored on `doneHeader`'s own clause and pulled again.

## Test: finish resolves the gap when asked
**Purpose:** with the act supplied the gap is resolved, source first exactly as a part is (R277)
**Input:** a gap-sourced item completed with a `ResolveGap` callback
**Expected:** the callback called with `O5`, and `GapResolved` true
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 23
**Fire alarm:** call `opt.ResolveGap` unconditionally in `Finish` — dropping the `!= nil` test is
a compile-safe way to write *always resolve*, so use `if hasGap` alone and have the caller pass
nothing. The gap closes on every completion, including the partial ones `#60` proves exist, and
**a wrongly closed gap is work that silently never happens.** Red on the sibling case as
`the gap was resolved without --resolve`
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang, `HasGap=true GapResolved=true; want the gap reported and left open`

## Test: a gap left open is recorded as a decision
**Purpose:** a gap left open deliberately and one left open by oversight are identical in
`design.md`, and the completion is the only place that difference is known (R281)
**Input:** a completion report carrying an unresolved gap
**Expected:** the output names `left open`, `O136` and `--no-resolve`, and never `still open` —
which would report a decision as an oversight
**Refs:** crc-Pending.md, crc-CLI.md
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 24
**Fire alarm:** delete the left-open block in `reportFinished`. Completion succeeds and the
record of *why* the gap is open is gone — `design.md` cannot tell a decision from an oversight,
so this line is the only place it was ever written down. Red as `the record is missing "left open"`
**Inject:** internal/cli/pending.go:reportFinished
**Pulled:** 2026-09-12 — rang again after `#31` threaded the repository root through `reportFinished` and labelled the written line; same injection and signature; restore byte-clean by copy. Previously 2026-09-05 — rang on all three strings, `the record is missing "left open"`

## Test: a resolved gap is reported and the notice stays quiet
**Purpose:** a report that both resolves a gap and warns it is open contradicts itself (R277)
**Input:** a completion report with the gap resolved
**Expected:** `resolved gap O136` present, `still open` absent
**Refs:** crc-CLI.md
**Code:** internal/cli/cli_pending_test.go

## Test: --resolve on a part-sourced item says it did nothing
**Purpose:** a flag that silently does nothing reads exactly like one that worked (R280)
**Input:** a completion report for an item naming a part, with `--resolve` given
**Expected:** the output says `--resolve did nothing`; the unresolved-gap notice does **not** fire
**Refs:** crc-Pending.md, crc-CLI.md
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 25
**Fire alarm:** delete the `askedResolve && !done.HasGap` block in `reportFinished`. The
completion succeeds, the flag resolved nothing, and the report is indistinguishable from one
where it worked — the same silence R279's refusal prevents, arriving on the side that has no
refusal to lean on. Red as `a flag that did nothing said nothing`
**Inject:** internal/cli/pending.go:reportFinished
**Pulled:** 2026-09-12 — rang again after `#31` threaded the repository root through `reportFinished` and labelled the written line; same injection and signature; restore byte-clean by copy. Previously 2026-09-05 — rang, `a flag that did nothing said nothing`

## Test: --resolve refuses a gap in another design root
**Purpose:** the gap resolved must be the one the **entry** named, in the document it named it
in (R277)
**Input:** a resolver bound to `tool/design/design.md`, handed a ref naming
`example/design/design.md`
**Expected:** a refusal naming the gap and both roots; the resolve act never runs; the matching
root still resolves
**Refs:** crc-Pending.md, crc-CLI.md
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 26
**Fire alarm:** drop the path comparison in `gapResolver` and resolve by ID alone. In this
repository — which has **two** design roots — a gap is then closed in whichever one the caller
is standing in: the wrong gap closes, the intended one stays open, and **both documents still
validate**, because nothing cross-checks a queue pointer against a gap. Red as `a gap in another
design root was resolved`
**Inject:** internal/cli/pending.go:gapResolver
**Pulled:** 2026-09-05 — rang, `a gap in another design root was resolved`

## Test: a near-miss gap ID reports its own refusal
**Purpose:** a gap-shaped token that fails the ID grammar gets that grammar's message, not the
part pointer's (R271)
**Input:** `--from O22-R5`, and `--from carves/x.md` for the other side
**Expected:** the first names the namespace rule and never `<doc>#<part>`; the second still gets
the part refusal
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go

## Test: --resolve and --no-resolve are refused together
**Purpose:** contradictory intents get no ordering, so honouring one would be the tool deciding
what the caller meant (R280)
**Input:** `finish 1 --commit abc1234 --resolve --no-resolve`
**Expected:** a non-zero exit and a refusal naming both flags, raised before the item is looked up
**Refs:** crc-Pending.md, crc-CLI.md
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 27
**Fire alarm:** delete the both-given test in `runFinish`. `--resolve` then wins by falling
through to the `if *resolve` branch, so a caller who wrote both — which can only mean they were
unsure — gets the **destructive** half silently. `update`'s own verbs refuse every paired slot
for this reason. Red as `contradictory flags were accepted`
**Inject:** internal/cli/pending.go:runFinish
**Pulled:** 2026-09-12 — rang again after `#31` threaded the repository root through `reportFinished` and labelled the written line; same injection and signature; restore byte-clean by copy. Previously 2026-09-05 — rang, `the refusal does not name both flags` — the exit code is not the witness, as before

## Test: a gap-sourced completion with no decision is refused before anything is written
**Purpose:** what a caller must not be able to do is pass silently through the decision, and only
a refusal prevents that (R279)
**Input:** `Finish` on a gap-sourced item with neither `ResolveGap` nor `DeclineResolve`
**Expected:** an error naming both spellings, and `DONE.md` untouched — the refusal lands before
the slot opens
**Refs:** crc-Pending.md
**Code:** internal/pending/pending_test.go
**Alarm:** 28
**Fire alarm:** delete the undecided guard in `Finish`. The completion then succeeds with the gap
silently left open, which is **exactly the state this whole change removed** — and it is the
state a notice permitted, so the injection restores the design that was superseded rather than
inventing a broken one. Red as `a gap-sourced item completed with no decision about its gap`
**Inject:** internal/pending/pending.go:Finish
**Pulled:** 2026-09-05 — rang, `a gap-sourced item completed with no decision about its gap`

## Test: the written line says which write is tracked
**Purpose:** validates R329 — a completion's four writes are named with their kind, so the carve flip reads as the one tracked, uncommitted file beside three ignored ones
**Input:** a real repository tracking `carves/x.md` and ignoring the three trajectory files; `finish 4 --commit abc1234`
**Expected:** `carves/x.md (tracked, uncommitted)`, `CURRENT.md (ignored)`, `PENDING.md (ignored)`, `DONE.md (ignored)` on the written line
**Refs:** crc-CLI.md — R329
**Code:** internal/cli/cli_pending_test.go
**Alarm:** 33
**Fire alarm:** print the names bare — return `strings.Join(files, ", ")` before asking git — and confirm all four words go missing while the completion still reports success
**Inject:** internal/cli/pending.go:describeWrites
**Pulled:** 2026-09-12 — rang: `the written line does not say "carves/x.md (tracked, uncommitted)"` and the three ignored ones; restore byte-clean by copy, and again after the simplification pass hoisted the git calls, same signature

## Test: a reverted part can be queued again
**Purpose:** validates R330 — the very part a revert released is the common re-add and must not be refused; a part carrying a live item still is
**Input:** `add-item` on part 7, `Revert`, `add-item` on part 7 again; then `add-item` on part 1, which carries `OPEN (#3.)`
**Expected:** the second add succeeds and the part reads `OPEN (#N.)` with no `REVERTED` left; the add on part 1 is refused
**Refs:** crc-Pending.md, crc-Backup.md — R330, R243
**Code:** internal/pending/pending_test.go
**Alarm:** 34
**Fire alarm:** drop the `releasable` conjunct from the collision check — the pre-fix shape — and confirm the re-add is refused with `already carries queue ID`
**Inject:** internal/pending/pending.go:AddItem
**Pulled:** 2026-09-13 — rang: `re-adding the part the revert released was refused: carves/x.md part 7 already carries queue ID #4; a part records exactly one item`; restore byte-clean by copy
