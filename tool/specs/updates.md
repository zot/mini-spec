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
atomically. The assigned `Tn` is reported on **stdout** as a sentence — `Retired R12 as
T7 (see R40)` — like every other `update` verb's report, suppressed by `--quiet`, and
carried by `--json` for a caller that reads it back. *Until 2026-09-13 stdout carried the
bare `Tn` as a return value (old R103); a minted value is a report, not a return value,
and `--json` is the channel for a caller that needs it.* `migration-complete` and
`add-req` report the same way: `Completed migration <name>: <path>`, `Added R5-R6 to
"<section>"` (R331).

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
# stdout: Retired R12 as T7 (see R40)
# stderr: ⚠ supersede at the source: R12 is retired, but directives describing
#           its old behavior remain and can cause a future revert.
#           • originating spec prose: specs/storage.md  (R12's **Source:**)
#           • design prose: grep design/ for R12 and the old names
#           completion test: could an agent reading only specs + design undo this?
```

## minispec update add-req --section [heading] --req [text] ...

Mint the next free `Rn` for each requirement given, append them to the named section
of `design/requirements.md`, and crank out what was minted. Assignment and the write
are **one act**: the tool never hands out a bare number for the caller to write down
later, because a number the caller is holding is a second copy of the numbering state.

**A batch is the primary form, not an accommodation.** Requirements arrive in blocks —
the skill's own instruction is to *"merge all specs into numbered requirements"* — so
`--req` may be given repeatedly and the entries are appended in the order supplied. The
crank handle reports a run as a range, `R963-R966`, in the syntax the tool already
parses in traceability comments and `query gaps` arguments.

*Why a repeated flag rather than trailing positional arguments.* `add-gap` joins
everything after its type into **one** description, so a sibling verb that split the
same tail into **many** requirements would give two neighbouring commands opposite
readings of the same shape — and an unquoted five-word requirement would silently
become five requirements. The flag cannot be misread that way.

**The destination is addressed by heading text, at whatever level it exists.** Not by
`**Source:**`, which is emphatically non-unique — 195 lines carry 142 distinct values
in ark — and not by feature alone, which is too coarse: `Source Monitoring` there holds
68 entries across its sub-sections.

A heading matches when the argument is its **literal text**, or its literal text with a
leading `Feature: ` removed. So `--section "Updates"` reaches `## Feature: Updates` in a
flat file, `--section "Go API"` reaches a sub-heading, and one flag serves both without
the caller having to know which level a title lives at.

**A new requirement lands at the end of the addressed heading's *own* content, never
inside a child.** A heading's own content runs to the next heading of **any** level, so
a `## Feature:` carrying both direct requirements and sub-sections gets the new entry
after its own last one and before its first sub-heading. Four of ark's 193 features are
that shape, and the rule is what keeps `--section "Table Sort"` from appending into
`### InboxEntry statusDate field`.

**An unknown heading is refused.** The tool owns IDs, not prose: a section title is a
judgment about how the design decomposes, and inventing one would be authoring. Write
the heading by hand first — it carries no ID, so it races nothing.

**An ambiguous heading is refused, and the refusal hands back forms that are not.** Each
candidate is listed with its level and its parent, and with the exact argument that
selects it: a sub-heading is selected by `<parent>/<title>`, and a feature by its literal
`Feature: <title>`. Nothing is written.

*Measured in ark, where the ambiguity is larger than a count of sub-headings shows.*
Eight titles are ambiguous across sixteen sites: six duplicated among sub-headings
(`CLI`, `Endpoint Integration`, `Go API`, `Lua Integration`, `Package Structure`,
`Store API`), and **two that collide across levels** — `Chunk Retrieval` and
`Server Lifecycle` each exist as both a feature and a sub-heading. The level-crossing
pair is the one worth guarding: `## Feature: Server Lifecycle` and the
`### Server Lifecycle` under `## Feature: Embedded UI Engine` sit 187 lines and one
unrelated feature apart, so a wrong choice files a requirement about the embedded UI
engine under `ark serve`. Both qualified forms resolve to exactly one heading, because
no sub-heading in either corpus begins with `Feature: `.

Output conforms to the markdown-by-default rule rather than deviating from it: a prose
sentence on stdout, suppressed by `--quiet`.

Examples:
```
minispec update add-req --section "Fuzzy Search" \
  --req "search --fuzzy accepts a maximum edit distance" \
  --req "(inferred) the default maximum edit distance is 2"
# stdout: Added R963-R964 to "Fuzzy Search"

minispec update add-req --section "Server Lifecycle" --req "..."
# stderr: "Server Lifecycle" names 2 headings, so the destination is ambiguous:
#           ## Feature: Server Lifecycle          --section "Feature: Server Lifecycle"
#           ### Server Lifecycle                  --section "Embedded UI Engine/Server Lifecycle"
#              under ## Feature: Embedded UI Engine
#         Re-run with one of the forms on the right. Nothing was written.
```

**`--req-file` is the file twin of `--req`**, repeatable like it and never mixed with it: the
two are separate repeated flags and nothing preserves their interleaved order, while the order
is exactly what assigns the numbers. Measured 2026-09-07 on this repository: R317–R326 were the
first requirements minted through the verb, ten lines added and nothing else touched.

## The gap verbs and `retire` write through the readers

`add-gap`, `resolve-gap`, `approve-gap` and `retire` edit `design.md`'s Gaps section and
`requirements.md` through `minispecsdom`'s gaps and requirements readers (R326): the reader
decides every refusal — a permanent gap resolved, a gap resolved twice, a requirement retired
twice, an unknown or ambiguous section — before a byte moves, and reads its own write back.
What stays this tool's is minting: the next `On`, `An`, `Tn` or `Rn` is computed here from what
the reader returned, retired and resolved numbers counted, and passed in. `retire` is two
documents in one verb — the head line rewritten in one, the `Tn` gap added in the other.

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

## minispec update repair-links [file...]

Repairs the links a carve's move broke, in both directions. When `trajectory-tool.md` moved
to `carves/done/` on 2026-09-14, every relative link in it kept pointing where it used to
live, and every document that linked it broke the other way; `query links` read 32 `missing`
in the moved file alone (gap `O27`). This is the repair for what has already broken; the
move verb that prevents it is `carves/reference-discipline.md` Item 5.

**With no file, the population is every live carve and every carve under `carves/done/`**
(and `.carves/`), because both ends of a move can hold a broken link. With files, exactly
those.

**The predicate is what makes writing into a human's document safe: a link is rewritten only
when it is `missing` now and exactly one sibling relocation resolves it.** The candidates are
the four a move between `carves/` and `carves/done/` can produce:

| direction | the link was written when | candidate |
|---|---|---|
| outgoing | the citing file lived in `carves/` and now lives in `carves/done/` | resolve the path from `carves/` instead |
| outgoing | the citing file lived in `carves/done/` and now lives in `carves/` | resolve it from `carves/done/` instead |
| incoming | the target lived in `carves/` and now lives in `carves/done/` | insert `done/` before the target's name |
| incoming | the target lived in `carves/done/` and now lives in `carves/` | remove `done/` before the target's name |

A candidate resolves when a file or directory is at the relocated path inside the repository.
**Exactly one must resolve.** None, and the link is not the move's doing — it is reported and
left alone; more than one, and the tool cannot tell which the author meant — reported as
ambiguous and left alone. Every other class (`tracked`, `outside`, `external`, `local`,
`untracked`, `ignored`) is untouched: the verb repairs one defect, a move, and does not
retarget links for any other reason. The fragment is kept as written.

**The rewrite is a byte-range splice through the `Markdown` reader** ([links-schema.md](links-schema.md)),
replacing the destination bytes alone: the text, the fragment, the title and everything
around the link stay byte for byte, and the reader reads its own write back. The new
destination is the relocated target relative to the citing file's directory, slash-separated,
`<…>`-wrapped only if the old one was.

**It reports every link it considered, per file: rewritten (old → new), left because no
relocation resolves, left because two do, and the counts**, zeros included. A run over
documents with no broken links writes nothing and says so; a second run finds nothing to
repair, which is what makes it safe to run after any move rather than something to schedule.
Exit status is 1 when any link was left unrepaired, so the run itself says whether the
documents are clean.

**Resolves at the repository root and needs no design root**, like `query links` and
`query carves`: carves are repository-scoped. Git is not consulted — a relocation either
exists on disk or it does not, and the class after the repair is `query links`' to report.


`pulled` takes `--body-file`, as every trajectory verb does. A backtick inside a double-quoted
shell argument is **command substitution**, and when it fires the text is simply **gone** from
what the tool receives, with nothing reporting it — measured 2026-08-20, four instances in one
session, to a caller who knew about the hazard. `add-gap` and `retire` still take their prose as
arguments here (gap `O23`).
