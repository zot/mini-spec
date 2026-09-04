# Query Commands

All queries read from design files and output results. No modifications.

## minispec query project

Show the paths resolved by the tool: repository root, design root, design
directory, src directory, and specs directory. Useful for verifying which project
the tool detected (especially when running from a subdirectory or when reported
missing paths look correct). When the repository root and the design root are the
same directory — the common case — it says so rather than printing the path twice
unlabeled. See [repository-root.md](repository-root.md).

Output:
```
repo root:   /home/me/work/ark (same as design root)
design root: /home/me/work/ark
design:      /home/me/work/ark/design
src:         /home/me/work/ark/src
specs:       /home/me/work/ark/specs
```

JSON output uses keys `repo_root`, `root`, `design`, `src`, `specs`. `repo_root` is
omitted when no repository root could be resolved.

## minispec query config

Show every effective setting with its value and **the file that supplied it**,
sorted by name. Settings no configuration file set report the built-in defaults.

This exists because configuration resolves across three layers, so the effective
value is not what any single file says. Without it, "why is this value what it is"
means reading two files and knowing the precedence by heart. Map settings are listed
per key (`comment_patterns[.go]`), matching the granularity at which they merge. See
the Config Scopes section of [config.md](config.md).

Output:
```
design_dir               design                        (built-in defaults)
src_dir                  lib                           /home/me/work/ark/.minispec/config.yaml
comment_patterns[.html]  <!--\s*|//\s*                  /home/me/work/ark/tool/.minispec.yaml
```

JSON output is an array of objects with keys `setting`, `value`, `origin`.

## minispec query requirements

List all requirements from requirements.md.

Output: List of Rn with text and source spec.

## minispec query coverage

For each requirement, show which design files reference it.

Output:
```
R1: crc-Store.md, crc-View.md
R2: crc-Store.md
R3: (none)
```

## minispec query uncovered

List requirements that have no design file references.

Output: List of Rn identifiers.

## minispec query orphan-designs

List CRC cards that have no Requirements field or an empty one.

Output: List of file paths.

## minispec query artifacts

List all artifacts from design.md with their checkbox states.

Output:
```
crc-Store.md
  [x] src/store.ts
  [ ] src/store_test.ts
crc-View.md
  [x] src/view.ts
```

## minispec query gaps

List all items from the Gaps section of design.md.

Output:
```
[ ] S1: spec item not in requirements
[x] R1: resolved requirement gap
[ ] D1: design without code
```

## minispec query traceability [file]

Check if a code file has proper traceability comments.

Output: CRC and Seq references found, or "missing" indicator.

## minispec query traceability --all

Scan all code files listed in Artifacts and report traceability status.

## minispec query unindexed-specs

List per-feature specs (`specs/*.md`, non-recursive) that are not referenced
anywhere in the root spec index `specs/index.md`. The index file itself and
anything under `specs/migrations/` are excluded. Matching is by exact `.md`
token, so `search.md` is not treated as indexed merely because
`fuzzy-search.md` appears. When `specs/index.md` does not exist, every
per-feature spec is listed (nothing is indexed yet).

Output: relative spec paths, sorted; empty (exit 0) when every spec is indexed.

## minispec query alarms

The census of fault injections recorded in `design/test-*.md`. One line per alarm:
the test document, the test title, the `**Inject:**` sites, and its state.

Four states, and the distinction between the middle two is the point:

| state | meaning |
|---|---|
| `verified` | has `**Pulled:**`, and no injection site has changed since |
| `stale` | has `**Pulled:**`, but a site has changed since — the proof is void |
| `unrecorded` | has `**Inject:**` and no `**Pulled:**` — a prescription, not a record |
| `unanchored` | has no `**Inject:**` — describes its site in prose only, so nothing can check it |

`unrecorded` is **not** a claim that the injection was never run. It is a claim that
the repository does not say it was, which is the only thing readable from the
documents — the same reason a carve distinguishes `NOT VERIFIED` from unstarted.

`unanchored` is what a backfill cannot fix. The person who pulled the alarm knew the
file; six weeks later nobody does, and an anchor guessed wrong is worse than none
because every future check follows it. The count is a measure of how much of the
convention was written after the fact.

Without git, `verified` and `stale` cannot be told apart; both report as `unchecked`
rather than as a clean result.

Output: one line per alarm, grouped by document, plus a closing census —
`27 alarms: 3 verified, 0 stale, 20 unrecorded, 4 unanchored`.

### `--unverified` — the census minus the 96% that carries no decision

A census is *asked*, not emitted, and nearly all of it is the answer *nothing to do
here*. Measured 2026-08-20 in this repository: **188 lines / 11,309 bytes**, roughly
3,000 tokens, of which **13 lines / 842 bytes** are every entry that is not `verified`.
`--unverified` selects the states that carry a decision — `stale`, `unrecorded`,
`unanchored`, `unresolvable`, `unchecked` — and leaves the bare form as the full census
for the runs that want it. The command already builds the list and already labels each
entry, so this is a filter over data in hand rather than new analysis. `query gaps` has
taken `--open` and `--closed` for exactly this reason since it landed.

**The list narrows; the count does not.** The closing census still reports the whole
population under `--unverified`, so the filtered form is the full form minus 165
repetitions of *nothing to do here* and minus nothing else. A filtered count would answer
*how many are wrong* and silently drop *out of how many* — the shape of the standing
workaround this flag replaces, `| tail -2`, which reports **whether** anything is wrong
and never **which**.

When nothing is unverified the output is the census line alone.

### `--brief` — what a delegated re-pull is spawned with

Re-pulling an alarm is a read-edit-test-restore-diff cycle, and the cycle is where the
cost is. Measured 2026-08-20 by `/context` at the end of a long session: tool results
**190.9k tokens, 19% of the window**, the largest single category and roughly four times
what the exchanges themselves cost — and the bulk of it was nine pulls and re-pulls
inside a single item, not the census. So the cycle is delegated, one alarm to one agent,
and `--brief` is what that agent is spawned with.

A brief is **complete about its alarm and silent about the protocol.** It names the
design root as a path relative to the repository root, the test document and test title,
the assessed state, the `**Inject:**` sites, the test files the `design.md` Artifacts
manifest maps that document to together with the directories holding them, and the
`**Fire alarm:**` prose verbatim. The protocol — baseline first, inject, diff, re-run,
restore, prove the restore clean — belongs to the agent definition that runs it, which is
where an invariant belongs. A brief carrying it would repeat some 350 words per alarm
against a census whose entire purpose is to be small.

**The one thing a brief does restate is the contract, because it is the half that decides
whether delegating was worth anything: evidence, never a verdict.** What comes back is the
command, its output before the injection, the diff applied, the output after, and the diff
after restoring. Whether the alarm rang is read off that by someone who can weigh it. The
reason is measured rather than principled — three times in one item the naive verdict would
have been wrong: an injection at a correctly named site that did not ring because the rule
had two guards, an injection that rang across three packages while being *incapable* of
reaching the property it named, and an `**Inject:**` field naming a symbol the injection
only consulted.

**A brief names files and directories, never a command.** `minispec` knows document
structure; it does not know build systems, and a guessed command that fails to build
produces output a hurried reader scores as *rang* — which is precisely the failure
`SKILL.md` names, *an injection that breaks the build teaches nothing, because the test
never ran*. The protocol closes it from the other side: the baseline run comes first and
must be green, so a wrong command is caught before anything is injected.

**The filter selects; the output form does not second-guess it.** A brief is emitted for
every alarm the filters selected, including ones that cannot be pulled as written, and
where the material for a part is missing the brief **states the absence in place of that
part** rather than dropping the line. An `unanchored` alarm has no sites to edit and a
document with no Artifacts row has no test files to name; a missing *Tests:* line reads as
an oversight, while a stated absence reads as what it is. Suppressing those briefs would
also make their count disagree with the census, which is the one number a reader trusts.

`--brief` is an output form and `--unverified` is a filter, so they compose;
`query alarms --unverified --brief` is the ordinary call. Under `--json` the brief is a
field on each assessment rather than a separate output.

## minispec query next-id \<item|gap|req\>

The next free identifier for a class of permanent, never-reused number. Read-only and
**informational**: the verbs that mint an ID assign and write it in one act, so no
answer printed here is ever copied into a document. It exists for a human reading the
queue and for an agent orienting itself.

The class is required. It selects both what is counted and **which root the tool looks
under**, and those differ:

| class | counted across | scope |
|---|---|---|
| `item` | the pending file **and** the done file | repository root |
| `gap` | the Gaps section of `design/design.md` | design root |
| `req` | `design/requirements.md`, retired requirements included | design root |

**`item` reads both files, and reading either alone is wrong in a way that looks
right.** The pending file's maximum is too low right after several items complete; the
done file's is too low whenever the highest IDs are still live. Each in isolation
returns a plausible number that collides with an existing ID. The shapes it reads are
specified in the skill's `trajectory-format.md`, which owns them; nothing here restates
them, and the sentence that used to name one is why — it went on describing a done-entry
shape the format had already replaced.

In the done file only an entry's **header** is read, and within it only the leading
identifier slot. Entry bodies are prose that quotes other items freely, so a body scan
turns a citation into the maximum.

**`gap` reports every gap type, because gap numbering is per-type.** `S`, `R`, `D`,
`C`, `I`, `O`, `A` and `T` each run their own sequence, so a single number would have
to pick one arbitrarily or take an extra argument for a question the caller usually
wants answered in full. Types with no gaps report `1`.

**`req` counts retired requirements.** A retired `Rn` keeps its number forever so old
references still resolve, so skipping them would hand out a number that is already
taken — the precise failure the permanence rule exists to prevent.

**A class whose files are missing is reported as unanswerable, not as `1`.** A project
running no trajectory layer has no pending or done file, and "the next free item ID is
1" is a confident wrong answer rather than an absent one. When only one of the two
files exists, the answer is given and names the file it could not read, so the reader
can judge whether the number is trustworthy.

**The answer names its evidence.** Every response reports which files were read and how
many identifiers each contributed, because the number alone cannot be checked. A regex
that matches nothing and a genuinely empty queue both produce "next free ID is 1", and
those are a broken parser and a correct answer wearing the same face. Printing the
counts makes the difference visible without the reader having to know the format:

```
next free item ID: #11
read:
  PENDING.md       1
  DONE.md          9
```

The counted noun is deliberately absent: the same column carries items for the queue
files, gaps for `design.md` and requirements for `requirements.md`, so naming it would
be wrong for two classes out of three.

Output is markdown on stdout; `--json` gives the machine-readable form.
