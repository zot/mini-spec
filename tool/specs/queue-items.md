# Queue items: creation and completion

Three verbs over the trajectory files: `minispec pending add-item`, which creates a queue entry
and the part line it points at; `minispec pending start`, which opens the item in the current
file; and `minispec pending finish`, which moves the entry to the done file, clears the current
file, and checks off the parts the item discharged. `add-item` and `finish` write **both sides of
the item↔part link** in one act.

*The readers are `internal/minispecsdom`'s* — `Pending`, `Current` and `Done`, ported from
`github.com/zot/simple-dom` on 2026-09-14 — and the tool holds thin path-taking adapters over
them in its `parser` package, as the carve reader already does ([queries.md](queries.md), carve
status view).
Every guarantee below that is about *how a region is found* — an entry as a heading node, the
`## Active` section as a heading region, the ledger's rule found in a text run — is the reader's
and is consumed here rather than restated.

The link is the point. A queue entry and the carve part it discharges are two records of one
piece of work, and every hand edit is a chance for them to disagree.

## Why it exists, measured on this project's own last completion

**Completing `#20` by hand on 2026-08-18 took five edits across four files**, and
`validate trajectory` caught **three** mistakes in them — one per file that has a checker:

| what was edited | what went wrong |
|---|---|
| the carve's part line: checkbox, strikethrough, marker | the box was checked and the title left unstruck — `validate trajectory`'s agreement sentry, which had never fired on the live corpus, firing on the hand that wrote it |
| the done file: a new entry | written only after the checker reported the part landed against a queue ID with no done entry |
| the pending file: the entry removed | removed only after the checker reported the ID live in one file and completed in the other |
| the current file: reset | — |

Three of five edits were wrong on the first attempt, by someone who had read the format that
morning. **The checks caught all three, which is the argument for the verb rather than
against it:** a check that reports a mistake after the fact is worth having and is not the
same as not making it. The edits are mechanical, the order is mandated, and the tool is where
mechanical-and-mandated belongs.

## `minispec pending add-item --from <doc>#<part> "<title>" --status <text> [options]`

**It mints the item ID and writes both sides in one invocation.** The tool mints IDs and never
hands out a bare one: assignment and the write that records it are one act, so a number is in
a document the moment it exists and `max()` over the queue files stays the whole truth. A
command that reserved a number for the agent to write later would be a second copy of the
numbering state by construction.

**What it writes:**

- the **queue entry** in the pending file — the *whole* entry in the shape
  `trajectory-format.md` mandates: the `##` heading carrying the number, the title, the skill
  and the one-line status; the `Source:` line with the part pointer `<doc>#<part>`; and the
  `Next:` line when one is given;
- the **part line's marker** in the carve — `**OPEN (#N.)**`, written by the marker rule:
  replace the first transient, delete the rest, append when there is none, never touching a
  record.

**The link is asymmetric, and that is structural rather than a choice.** Trajectory files are
private in every project, so the public→private direction can never be a markdown link: the
carve carries a bare key and the queue side holds the recorded pointer. One shape to build,
not two — and it is the half hand-maintenance cannot supply, since nothing in a carve can
point at a file a cloner does not have.

**It refuses rather than guessing.** An unresolvable `<doc>#<part>` — no such document, no
such key, a key that the document does not key under the one mandated form — is a refusal
naming what it looked for and where. A part that already carries a queue ID is a refusal too
— unless that ID is the reverted attempt's: while the slot holds a reverted attempt the part
still reads `REVERTED (#N.)`, and the release that returns it to open and unqueued runs inside
the next mutation, ahead of the new marker, so re-queueing the very part just rolled back is
the common case and proceeds (R330; measured by mini-spec-tool 2026-09-06, when it was refused
while a sibling's mutation released it). Otherwise it is the refusal it was:
a part records exactly **one** item, and re-queuing it silently would make the older pointer
resolve to work it never described.

**A reused number announces itself.** When an unfinished attempt returns a number to the pool,
the next vend of it says so — reaching the one person who could have carried that number into
a chat or a commit message. Bounded: at most one number is ever ambiguous, and only until the
next mutation.

### A gap is a source too, distinguished by shape

**`--from` takes a gap ID as well as a part pointer** — `--from O136` beside
`--from carves/x.md#3` — because repairing a gap is the second most common thing an item does
and the verb refused it. Every gap-repairing item was typed into the pending file by hand and
bypassed the placing verb entirely, which is the hand edit this project's carve on hand edits
exists to eliminate, arriving through the one verb that carve had already fixed.

**No flag distinguishes them, because the shapes cannot collide.** A gap ID is a letter and
digits — no `/`, no `#`, no `.md` — and the grammar is the one `ExpandIDRefs` already parses for
inline ID selection, reused rather than restated. **A range or a list is refused naming the
rule**: `--from O136-O140` is well-formed in that grammar and is not a source, because an entry
carries **one** pointer. That refusal is worth more than a parse error — it names the decision
instead of the syntax.

**A gap ID needs a design root, and its absence is a refusal rather than a search.** Gap IDs are
scoped to a `design.md`, and a repository may hold several design roots — this one holds two. So
the root is resolved the way every other gap verb resolves it, and when there is none the verb
says what it needed and where it looked. `--from carves/x.md#3` is unaffected: a part pointer
names its own document and is repository-scoped, which is why `pending` resolves no project for it.

**Nothing is written on the gap side.** For a part the verb writes both halves — the entry and
the part's `**OPEN (#N.)**` marker — because the carve is the public document a cloner reads and
the queue is private. A gap has no such marker and does not gain one: inventing a queued-marker
shape would add a second thing a gap line can say while that document class is mid-migration, and
the `Gaps` reader would have to admit it. **The cost is stated rather than hidden:** nothing in
`design.md` says a gap is queued, so there is no gap analogue of the carve→queue cross-check and a
pointer at a resolved gap goes unreported. That is a gap to file when it bites, not a shape to
guess at now.

**One gap per source.** The entry carries a scalar pointer, the same scalar a part pointer fills,
and an item naming further gaps names them in its prose the way `#60` did. This is the same
ruling the multi-part question already has: the reader records one, the model permits several,
and growing both at once through the gap door would be building for a consumer that has not
arrived.

### Completing a gap-sourced item: `--resolve`, and a notice when it is absent

**Completing a gap-sourced item requires an explicit decision: `--resolve` or `--no-resolve`.**
Neither given is a refusal, named before anything is written, and it is a refusal rather than a
notice because the two failure modes are not equally recoverable. A notice mitigates forgetting;
a required choice makes it impossible. `--resolve` resolves the gap, which is the gap's answer to
the `LANDED` marker a part gets; `--no-resolve` leaves it open.

*This replaced a notice, and the notice's own argument is what replaced it.* The earlier rule
said the caller is told the gap is still open and left to act — reasoning that a flag the caller
must remember to type is not a reminder, so the reminder had to be the output. That is true and
it is the wrong conclusion: what a caller must not be able to do is **pass silently through the
decision**, and only a refusal prevents that. Giving both answers a spelling is what makes the
question unskippable.

**`--no-resolve` is a record, not a shrug.** A gap left open by decision and a gap left open by
oversight look identical in `design.md` — the checkbox is unticked either way — and the
completion is the only place that difference is known. So the report states which happened, the
same reason `NOT VERIFIED` earns its own words in a carve's status block rather than being left
unmarked. `#60` named three gaps and closed two on purpose, and nothing anywhere records that the
third was deliberate.

**The two are refused together**, the way every other paired slot on these verbs is: nothing
preserves an ordering between contradictory intents, and guessing one would be the tool deciding
what the caller meant.

**Resolving is never inferred, and the evidence is in the record.** `#60` named three gaps and
closed two, leaving one open on purpose. A gap wrongly marked resolved is work that silently
never happens, and the entry that would have reported the loss is the one just closed — this
system's own core failure mode. So the caller states the intent and the tool performs the act;
there is no default, because a default is a guess about which of those two cases this is.

*Why the flag lives here rather than being left to `update resolve-gap` alone:* `finish` already
takes `--discharged`, in which the caller names the gaps and requirement ranges the item
discharged. That information already flows into this verb at completion time and dies there as
free text in a header. The flag makes a conversation the verb is already having actionable,
rather than opening a second door onto an act it knew nothing about.

**The done entry carries the pointer.** A part-sourced completion records `` Part `<doc>#<key>` ``;
a gap-sourced one records `` Gap `<doc>#<gap>` `` in the same shape, so the ledger keeps the link
the pending file held. Without it the pointer survives only in whatever free text `--discharged`
was given, which is a record by luck.

### The written line says what kind of file each write is

Every queue verb ends by naming what it wrote, and since 2026-09-12 each name carries a word:
`carves/x.md (tracked, uncommitted)`, `CURRENT.md (ignored)`. A completion writes four files and
exactly one of them is a tracked public document — the carve flip — so it is the write that
still needs a commit, and until now nothing in the report told it apart from the three gitignored
files beside it. *Until 2026-09-15 this was structural:* a `LANDED` record carried the commit hash, so it could
not be written until that commit existed, which put the source edit one commit behind the work
it recorded, always — measured 2026-08-18 on `#18`, and again on `#83`–`#85`, three items in six
commits. The record now carries the item number instead, so the flip lands in its own item's
commit and the word `uncommitted` means only that the commit has not happened yet. The words come from
git — `tracked`, `ignored`, otherwise `untracked` — and print bare where there is no repository.
**The tool never stages or commits it** (Bill, 2026-08-04): this reports, the agent relays, the
human commits.

### The tool never leaves a line half-written

**The status sentence is required and the `Next:` line is not**, and the line between them is
not a preference about how much an entry should say. The status sentence lives **inside the
heading the tool mints**, so omitting it leaves the agent completing a line the tool has just
written — the entry's first line, in a gitignored file, with the tool's own text on both sides
of the hand edit. The `Next:` line is a whole line of its own, so leaving it out writes a
well-formed entry that is simply shorter. *A tool may decline to write a line; it may not write
two thirds of one.*

*Measured over the live queue rather than assumed:* **10 of 10** entries in this project's
pending file carry both. That is a census and not a majority — and until 2026-08-21 the verb
could produce neither, so every one of those ten was hand-finished into the file the verb had
just written.

*The first count of this was **12 of 12**, and it was wrong in the way this project keeps
banking:* the population was counted **after** the two entries filing this work had been added
to it. A `grep -c` run one command too late, against a file the same session had mutated.

**Three prose slots, each with a file form.** The title positionally or `--title-file`; the
status as `--status` or `--status-file`; the next action as `--next-action` or
`--next-action-file`. Giving one slot both ways is refused rather than resolved by precedence.
The file forms are read byte for byte and exist for the reason `update pulled --body-file`
does: **a backtick inside a shell argument is command substitution**, and when it fires the
text is simply *gone* from what the tool receives, with nothing reporting it. Prose destined
for a document this tool parses is exactly the payload that must not travel on a command line.

**The title is plain text, and the verb supplies the emphasis.** A title wrapped end to end in
`**…**` is refused, naming the repair. The wrap was unconditional, so an already-emphasised
title came out as `****…**` — malformed markdown, from valid input, with no complaint; measured
2026-08-18 filing `#32`–`#34`, where all three headings came out that way and were rewritten by
hand. **The verb is the format authority here**, which is the whole point of a stencil, so a
caller supplying emphasis is supplying part of the shape the tool owns. *Emphasis **inside** a
title is untouched* — a title mentioning `**OPEN**` is legitimate and is not a wrapped one, so
the refusal fires only when a single `**…**` run spans the entire argument.

**And the reader must read it back whole.** An entry's title is the heading's first emphasis
run, closed at that run's own end rather than at the first `**` following it, with code spans
suppressed and openers told from closers by **CommonMark's flanking rule** — so every title this
verb accepts survives a round trip through `PendingEntries`. *On this branch the reader is `minispecsdom`'s
, and its title read is the lazy form the history below found wanting; measured
2026-09-05, none of the 11 live entries carries emphasis inside its title, so nothing misreads
today. Recorded as a gap and raised with `mini-spec-tool`, which held the reader then, rather than patched around here.*
*The two halves are one rule and stood apart for six days.* The write half blesses `**OPEN**`
inside a title, above, while the read half was `^##\s+\d+\.\s*\*\*(.+?)\*\*` — a **lazy**
group over a raw line — which closes on that very interior run. Measured 2026-08-24 against the
real function: `A title mentioning **OPEN** here` reads back as `A title mentioning `, and
a title carrying `**` inside a code span reads back cut at the backtick, which is the form that
reached `DONE.md`. *Flanking rather than whitespace, because the whitespace half alone drops the same defect in
different dress:* a run preceded by punctuation reads as a closer to a test that only asks
about spaces, so `**Say (**loudly**) now**` came back as `Say (` — measured the same day,
while writing the reader that fixed the first two. **A verb that owns a format must be able
to read what it wrote**, and none of these failures reported anything — the regex matches, so the count of unrecognised lines
stays zero.

**`--next-action` and `--next` are neighbours in spelling and unrelated in kind**, which is
said here rather than left to be discovered: `--next-action` is prose written *into* the entry,
`--next` is *where the entry goes*.

### Placement: the caller states the intent, the tool performs the move

The verb appended and then cranked out *"Order it by intent: it was appended last, because
priority is a judgment the tool does not have."* **That message conflated the judgment with the
mechanics.** Which position an item deserves is a judgment and stays the caller's; moving the
`##` block to that position without disturbing its neighbours is mechanics, and it is the same
split every other verb here already draws.

- **`--next`** — *next to be worked*, not *position 1*. With nothing in progress the entry goes
  to the top; with a step in progress it goes **immediately after** the item being worked.
- **`--nth N`** — position `N`, 1-based, where `N` may be one past the last entry.
- **`--after N`** — immediately after the entry whose **item ID** is `N`.
- **`--last`** — the end of the queue. The default, spelled out.

*Why `--next` rather than a `--first`.* `--first` forced the caller to know something the tool
already knows, and position 1 is not neutral: the pending file's rule is that **the top item is
active**, so `--first` would not merely order the entry, it would make it the active item.
Naming the **intent** rather than the coordinate dissolves that rather than answering it.

**"A step is in progress" is `CURRENT.md`'s `## Active` section holding something other than
its placeholder**, and the top item is the active one by the pending file's own ordering rule —
so `--next` resolves to position 1 or position 2 and needs nothing more. `finish` already
addresses that section as a heading node and already writes that exact placeholder, so this
costs the node lookup the tool performs anyway plus one comparison, and no new reading of the
trajectory files.

**`--nth 1` is refused while a step is in progress** — the same refusal from the other side,
since position 1 is the active item's slot. The message names both repairs: `--next` if *next*
was what was meant, or park the active item first if the caller genuinely means to preempt it,
which is a real thing to want and must not read as forbidden.

**The flags are mutually exclusive** and giving two is refused. **An `--after N` naming no live
entry is refused**, the way `--from` refuses an unresolvable part and by the same argument. **An
`--nth N` outside `1 … entries+1` is refused rather than clamped**, because a clamp is a silent
reinterpretation of an instruction the caller was specific about.

***`--after` addresses by item ID; `--nth` addresses by position***, and this project's standing
rule is that IDs are addresses and positions are not. `--nth` survives it only because the
position is **consumed in the same invocation that states it** and is never stored. So
**`--after` is the safer form and `--nth` is the convenience** — said here so the wrong one does
not become the habit.

**The entry is placed as a document node, never as a line range.** An entry is a `##` heading as
the *reader* sees it, so a `## 5.` quoted inside a fenced example is inert: it cannot be
counted, targeted, or split by an insert — which is the defect the line-scanning queue writers
of August 2026 carried until the readers replaced them. The move is expressed as **one insert
beside the node the entry lands next to** — the heading it goes in front of, or the end of the
entries at the last position — which is the reader's own guarantee that an edit can only affect
the node addressed.

*The last position stops where the entries stop*, which is a `---` rule, the next heading of
level 2 or higher, or the end of the document — an insert that ran past the rule would put a
queue entry inside the commentary. **A rule is not a node**: the reader carries it inside an
opaque text run, so it is found by scanning that **one run's own content** — and that is
fence-safe *by construction* rather than by care, which is the whole difference from the
line-scanning writers this replaces. A fence is a node of its own, so its bytes are never part
of a text run and a `---` quoted inside an example cannot be reached from there.

## `minispec pending finish <N>`

**One command across all four surfaces**, in the order the format mandates:

1. **The source first** — each part the item discharged is checked off in its carve:
   `- [ ]` becomes `- [x]`, the title is struck through, and a
   `` **LANDED (<date> — `#N`.)** `` record is appended — **no commit hash** (Bill, 2026-09-15):
   the item number is the identifier and every commit names the items it lands, so `finish`
   runs *before* the commit and the flip lands in it. Source first because that
   is the copy a future reader trusts, and the one nobody thinks to check.
2. **The current file's `## Active` section** is reset to its empty shape — that section
   and nothing else, see below.
3. **The queue entry moves** from the pending file to the done file, as an entry header
   carrying the date, the identifiers, the title and the part pointer, and no commit; the
   ledger, being private, may gain one after the fact by hand.
4. **The body is placed with the header**, when `--body` or `--body-file` is given. *One write,
   not two* — the body lands beneath a header the same splice produced, so there is no second
   target and nothing to anchor against. The carve's guarantee is satisfied by construction
   rather than by a check: an edit that never happens cannot change the wrong region.

**The identifier slot is filled, not half-filled.** `--discharged <text>` supplies whatever the
item discharged besides its queue ID, and the tool joins the two with the ` / ` the format
mandates: `- **2026-08-24 — #58 / R476–R477: <title>.**`. **The tool writes the `#N` because the
tool owns IDs**, and it cannot know a requirement range, so without this it writes half a field and
the rest is a hand edit *inside the line it just wrote*. Measured across four consecutive
completions on 2026-08-24 — `#54`, `#56`, `#26`, `#57` — every one repaired by hand.

*The slot's openness is deliberate and is not what this fixes.* It takes a queue ID, a gap ID, a
requirement range, several separated by `/`, or nothing at all, and that is right. What was missing
was a way to hand the tool the rest of what it should say.

**A colon in the slot is refused, naming why.** The reader takes the slot as *the run between the
date's em dash and the colon that opens the title*, so a colon inside it ends the slot early and
every identifier after that colon stops being read — silently, since what remains is still a
well-formed header. **A verb that owns a format must refuse the input that would make its own reader
wrong**, rather than write it and let a later count come back short.

*Composing stays the caller's here too.* The tool must not infer a range by diffing
`requirements.md`: a generated identifier list reads exactly like an authored one, which is the same
objection the header rule makes about bodies, one field along.

**Both writes are node-addressed, which finishes what the placement rule started.** The entry is
removed as the **heading node** that opens it plus the nodes beneath it, and the done entry is
placed beneath the rule that ends the ledger's preamble — *a rule found in a text run's own
content, never in the file's lines*. A fence is a node of its own and a suppressor, so a `## 5.`
or a `---` quoted inside one cannot bound a removal or receive a record. **Fence-safe by
construction rather than by care**, which is the same guarantee `add-item`'s placement has.

*Both failures were live and both were silent.* Removing an entry whose body held a column-0 fence
stopped at the fenced heading, so the fence's opening marker went and its body and closing marker
were orphaned at column 0 against the rule — a structurally broken document, written without
complaint, into a gitignored file. And a done file whose preamble quotes the format in a
fence containing a `---` received the new entry beneath **that** rule: the completion record landed
inside a code block, where no reader sees it and the queue ID resolves to nothing. *That is
the sharper of the two, because the record it loses is the one a future reader is told to trust.*

**The boundary between an entry and what follows it can fall inside a node, and usually does.** The
reader merges an entry's `Source:` line, the rule that ends the entries, and the commentary below it
into one text run, so a removal that only drops whole nodes leaves the entry's own body standing
above the rule. The run is split at the rule instead. *`add-item` and `finish` are inverses or they
are not*, and the check that says which is a round trip asserting the queue file comes back byte for
byte — which is what caught this.

**The composing/placing boundary does not move, and that is the point.** The agent authors the
prose — a judgment about what a future reader will need, which no tool holds — and the tool owns
the format and the position. What changes is only *who types it into the file*, which was never
the authoring half. `--body-file` is read **byte for byte** and is refused alongside its inline
twin, the same pair every prose slot on `add-item` already uses.

*A file rather than stdin, and the carve named the test rather than the answer:* **which the
harness actually makes cheap, measured on a real completion.** Measured across **nine** consecutive
hand-written bodies — every one composed by writing prose into a file with a quoted heredoc, because
prose in this project carries backticks and a backtick in a shell argument is command substitution
that vanishes silently. Six existing flags already say the same thing.

**With no body the verb still cranks the hand edit, and now names the flag.** Nine completions ran
that way before 2026-08-24, each writing prose into a file the verb had just written. *The crank
handle is the fallback, never the design* — the primary answer is that the tool acts, because then
there is no anchor to get wrong, no delta to assert, and no temptation to append where a supersede
was meant.

**It appends a record and never removes or overwrites one.** The marker rule's caution is
about overwriting an **assessment** — a `NOT VERIFIED` that only a person can make and no
tool can reconstruct. A `LANDED` composed entirely from facts the tool was handed overwrites
nothing, and two things depend on it being written: the format states that *a part with no
marker is indistinguishable from one nobody has considered*, and a landed part's queue ID
lives in that record and nowhere else, so the carve→queue checks would have nothing left to
read.

**It refuses a part line that does not conform, and this now covers the schema and not only
the lexicon.** A write path stops where a read path lists — the damage is real here, and
refusing costs a message. Until 2026-08-20 "refused" meant a line whose *markup* the lexicon
could not close; it now also means a line whose key, separator, checkbox or `OPEN` marker is
not the one shape the carve format admits.

**Annotations come first, then prose, and nothing after it.** A part line is its head, then its
markers, then at most one run of trailing prose — `Head Marker* Text?`. A marker appearing
*after* prose is refused the way any other non-conforming shape is: listed by read paths, named
with its target, refused by write paths.

*Why the grammar was looser, and why that reason expired.* It alternated freely —
`Head (Marker | Text)*` — and the justification was measured rather than stylistic: four real
ark parts carried a marker separated from the head by prose, and a head-adjacent rule would have
left those markers unaddressable while every test stayed green. **All four were artifacts of the
superseded key scheme**, where the queue number sat in key position and the title fell outside
the bold — `**#68 `/mini-spec`** — RC doc-drift audit.` Migrating ark's keys on 2026-08-21
eliminated every one of them, so the rule outlived its evidence by a matter of hours and nothing
would have reported it.

*What the tightening buys is not tidiness.* The free alternation makes a bold run in prose
indistinguishable **by shape** from a marker on the part, which is tolerable while a part line is
one line and fatal as soon as a reader considers continuation lines: ark's own
`forge-review-console.md` carries `**NOT VERIFIED**` mid-sentence in a part's body, a legal
marker by shape *and* by vocabulary, describing three Lua bindings rather than the part. Under
`Head Marker* Text?` it is unreachable, because prose opened two lines above it. **Prose is the
terminator, so once it starts the annotations are known to be complete.**

*Trailing prose is kept because it carries the part's detail*, and dropping it would be a
migration rather than a clarification: measured 2026-08-21, **34 of 102** live part lines across
both projects end in prose after their last marker, while **0** interpose it. The tightening
therefore migrates nothing.

*The widening was earned by this verb's own defect.* Completing `#35`, the part carried
`**OPEN (#35, on re-source.)**`. That attribution is not a transient by the lexicon's test,
so the marker read as a **record**, and the append rule — correct in itself — put `LANDED`
beside it, leaving one line asserting both states. The verb wrote a line that contradicted
itself and nothing in the write path objected. Under the narrowing that line is refused
before anything is written, and the message names the shape it must take.

**A parent part is never touched.** A parent completes when its subparts do — derived, never
stored — so completion checks the parts the item recorded and nothing else. No parent boxes,
no inference about what a part's siblings mean.

**A part pointer needs both halves, and an entry that names only a document discharges
nothing.** The pointer is `<doc>#<key>`; a document alone is not one. A queue item whose
`Source:` names a spec, a pattern file or a plain document — most of them — records **no
part**, and completion simply has no source to mark. *Measured 2026-08-18 by running
`finish` on such an item:* the reader set the document from the `Source:` link
unconditionally and only required *both* halves to be empty before reporting no part, so
every entry yielded a part with an **empty key** and completion refused with
`has no status block, so it holds no part` — naming a real condition of the wrong document,
which is the sort of true-and-useless message that costs an hour. The rule was already
written correctly one consumer over, in the slot's release path, which requires both.

**An item may discharge parts in several documents**, and each is checked. The cardinality is
asymmetric on purpose: a part records one item, an entry records a list of parts.

### The current file: the region is addressed, never inferred

**`## Active` is the only region `finish` clears.** It is found as a *heading node* in the
parsed document and it ends at the next heading of the same level or higher, so the agent
may use `###` and below inside it as freely as it likes. Everything else in the file —
standing context in its own `##` sections, the preamble above the rule — is outside the
verb's reach **by construction**, because the verb has no way to name it.

**The edit is a removal of the nodes inside that region and one insert**, through `minispecsdom`'s
 `Current` reader — never line arithmetic and never an offset splice. That is what
makes the previous sentence structural rather than a promise: a raw byte range straddling node
boundaries corrupts structure silently, which is the failure the reader exists to prevent, and a
verb that can only remove nodes it enumerated cannot reach a node it did not.

**Two refusals, each naming the repair.** A file with **no** `## Active` is legacy or
damaged, so the tool says to add the heading beneath the rule rather than guessing where
the active item ends. A file with **more than one** is ambiguous: the guarantee is *correct
targeting*, and a tool that cannot tell which region it was asked to clear refuses rather
than picking. Absence is reported as an error, never as silence.

**The guard that would have caught the 320 lines is structural, not a check.** The region
ends at the next heading of level 2 or higher *by the reader's enumeration*, so a standing
section's heading can never be inside it — there is no computed bound for a check to disagree
with. The August tree also re-parsed the render and compared every node outside the region;
that check derived *outside* from the same computation it was checking, so an over-wide region
sat inside its own claim, and it is not carried here.

**Measured 2026-08-18 on this repository, which is why the shape changed.** `finish` reset
the current file by keeping everything above the `---` rule and replacing everything below
it — a definition of "the active item" that was correctly implemented and wrong about the
document. Four standing sections lived below that rule and **320 lines were deleted**:
the dogfooding priority, the answers already obtained from the user, the pointers to work
outside this repository, and the state of things. `validate` stayed green, `validate
trajectory` stayed green, the file is gitignored so there was no diff, and **nothing
reported it**. It surfaced when a person asked what was worth recording for the next
session and the honest answer required reading the file.

*The document was parsed and re-rendered byte for byte at the time, by the reader this verb
was not using:* 23,144 bytes, **six heading nodes**. Six is the shape of the failure — the
edit had to replace one section and took five.

### The tool writes the header; the agent writes the body

**The done entry's body is authoring, and the tool does not do authoring.** The done file's
own rule is that an entry records *enough to reconstruct the change without re-reading the
code* — which is a judgment about what a future reader will need, not a fact the tool holds.
So `finish` writes the entry **header**, which is identifiers and dates and pointers, and
cranks out where the body goes.

This is the same line the requirement-minting verb holds: the tool owns IDs, not prose. A
tool inventing a summary would be doing the one part that is not mechanical, and doing it
worse than the agent standing right there.

## `minispec pending start <N> [--context <text>|--context-file <path>]`

**The third verb, and the one edit in an item's round trip that no command emitted.** `finish`
*clears* `## Active`; nothing wrote it. So the edit made at the start of every single item, into
the one file whose whole purpose is to say what is in progress, was by hand — **into a gitignored
file, before any record of the work exists anywhere else**, which makes a loss there the loss of
the sole copy. This carve's `## Why` records that loss: 320 lines of standing context, with no
history finer than the two fossil check-ins either side.

**The tool writes the identity line; the caller composes the context.** `` `#N` — <title> `` is
facts the tool already holds from the queue entry — the same facts it puts in a done header, read
from the same place, so the two cannot disagree. Whatever the caller supplies goes beneath it,
byte for byte, through the `--context`/`--context-file` pair every prose slot here already uses.

**It refuses a `## Active` that already holds an item, and names the repair.** Overwriting it is
the 320 lines again one level up, and the format already says what to do instead: park the active
item's context as a sub-item in the pending file — *a stack you can push onto* — which frees the
section without discarding anything. A verb that silently replaced it would be doing the one thing
this whole carve exists to prevent, in the one file where nothing would report it.

**The region is addressed exactly as `finish` addresses it**, and the guard comes with it: the
`## Active` heading node, ending at the next heading of level 2 or higher, with a level-2-or-higher
heading *inside* that range refused as standing context the verb was not asked for. That check is
made against the nodes rather than derived from the computed bound, because a check that consults
the same computation it is checking proves a function equals itself. **Sharing the write path is
the point** — a second implementation of *find the active section* is a second chance to get it
wrong, in the file that has already been destroyed once.

**It is a `Slot` operation, which settles open question 2.** The carve asked whether `start` is one
verb with `finish` or its own, and named the test: whether it needs the worktree anchor. It does —
`Slot.Record` takes the anchor at the top of the swap because `releaseAttempt` writes a marker into
a **tracked** carve, and any verb going through `Record` inherits that. And `start` must go through
`Record`, because it writes the section that was destroyed. *So: the slot discipline of `finish`,
and its own command* — `#17`'s bound holds unchanged, since `start` touches neither side of the
item↔part link.

## `minispec pending commit-message [--amend] [--out <file>]`

Composes the commit message for the items finished since the last commit, so that every
commit names the items it lands — the obligation the item-number identifier rests on
(`carves/item-identifiers.md`, Bill, 2026-09-15). The tool writes the message and never
commits; the human stages and commits.

**Which items are uncommitted is git's to say.** The done file is private and never in a
commit, so the tool cannot ask git which entries it holds; it asks the other way round —
which `#N` the messages on `HEAD`'s history name — and the uncommitted entries are the
newest ones down to, and excluding, the first that a commit names, by a hash in its slot or
by every identifier appearing in a message. `#N` is bounded by a non-digit, so `#40` does not
name `#4` (open question 1, answered). **Everything older than a named entry is history,
whatever its slot says**: measured on this repository's first run, two entries from August
carried neither a hash the reader recognises nor a number any message names, and a rule that
looked at each entry alone composed them into the message. A repository with no commits
names nothing.

**The message.** The subject names the items and their titles: `#85, #86: <title>; <title>`.
The body opens with `Items #85, #86.` on its own line — the line `git log --grep` finds — and
then, in the order the items finished, each entry's `#N — <title>` followed by the entry's
body from the done file, which is *enough to reconstruct the change without re-reading the
code* and so is exactly what a commit body should carry. The tool adds no sign-off; that is
the committer's. Nothing to compose — every entry named — is a refusal, not an empty message.

**`--amend` appends, never rewrites.** It reads `HEAD`'s message and returns it unchanged
with the new items after it — `Also lands #87.` then the entries — because the previous
message is part of the record (Bill, 2026-09-15). It refuses when `HEAD` is on any remote
branch, since amending a shared commit rewrites history someone else holds; the follow-up
commit is the answer there. The items an amend names are the ones no message names, exactly
as without the flag, so an item already named by `HEAD` is never repeated.

**To stdout, or to `--out <file>` byte for byte**, for `git commit -F` and
`git commit --amend -F`. Resolves at the repository root; needs no design root.

## All three verbs go through the backup slot

Every write here is a mutation of the trajectory files, so all three run inside the slot: the
files are copied, the worktree anchor is taken, the operation performs, and the state becomes
`changed`. *`start` joined the other two on 2026-08-24, and has the strongest claim of the
three to be here*: it is the only verb that writes before any other record of the work
exists anywhere, so a loss under it is the loss of the sole copy. One level of undo covers the whole invocation rather than any single file, which is
what makes a mis-typed part pointer a `pending revert` instead of a repair.

**These are the slot's first production callers.** Until now it was reachable only from its own
tests, which is *specified, implemented, tested, uncalled* — a state indistinguishable from
finished in every view the tool has.

## Creation: refused, then `--create`

A fresh project has no trajectory layer, and the tool's verbs never create a file — so the
first `add-item` there used to fail with a raw `open …/PENDING.md: no such file or directory`
(measured 2026-09-13) and no way forward that the output named. Now every queue verb run where
any of the three files is missing **refuses before anything is minted and names the missing
files**, and the CLI cranks out the create instruction: the same `add-item` again with
`--create`, and what git will do with the files under the project's `track` — the `.gitignore`
question folded in, since `init --track-*` already answered it (R332).

**`--create` is the scaffold, reached by being refused rather than by knowing a verb exists.**
It writes each *missing* file with the lifecycle preamble `trajectory-format.md` mandates —
the pending file's ordering and ID rule above its rule, the current file with its `## Active`
placeholder, the done file — reports them like any other write, and proceeds with the item. A
file that exists is never touched: creation is the one write that must refuse rather than
overwrite, since the file it would replace is the whole record; a second `--create` writes
nothing (R333). A carve that does not exist yet is `minispec init carve <name>`
([initialization.md](initialization.md)), since no write path leads to one.

## Everything is cranked out

Every change is printed in full: the number minted, the files written, the parts checked, the
entry moved. An agent that mutates state through a tool and cannot see the result will reason
from a stale picture and eventually assert it. Markdown to stdout by default, `--json` for the
machine-readable form.

**And a crank handle never instructs an edit the tool could have made.** `add-item` printed
*"Order it by intent: it was appended last, because priority is a judgment the tool does not
have"* until 2026-08-21 — a verb telling its caller to move a block in the file it had just
written, which is *never instruct an edit whose target the tool cannot verify* broken by the
verb's own output. What replaced it says where the entry went and names the flags that would
have put it elsewhere. The judgment stayed the caller's; only the typing moved.

## What these verbs do not do

- **They do not author prose.** Not the done entry's body, not a part's elaboration, not a
  title, a status sentence or a next action beyond the ones they were given. *They place all of
  them*, which is the distinction this document turns on rather than a softening of it.
- **They do not decide what a part means.** An unkeyed part line is named, never absorbed; a
  document on a superseded key scheme is reported as unmigrated with its target named.
- **They do not touch the working tree**, beyond the trajectory files and the carve markers
  they were asked to write. The anchor is taken so the agent can *see* what moved; nothing
  here reverses it.
