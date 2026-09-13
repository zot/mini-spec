# Initialization

[repository-root.md](repository-root.md) says how the tool *finds* a repository root
and [config.md](config.md) says how configuration *resolves* once found. This spec
says how a repository comes to have either in the first place, and how the tool keeps
that bootstrap honest afterwards.

One command creates it — `minispec init` — and one setting, `track`, records the
choice a tool cannot infer: whether this project's work queue is private or ships
with the repository.

## The `track` setting

`track` lives in the repository configuration, `<repo root>/.minispec/config.yaml`.
It has three values:

| value | means |
|---|---|
| `none` | the project is not under version control |
| `private-trajectory` | git; the trajectory files are ignored |
| `all` | git; the trajectory files are tracked with everything else |

The trajectory files are `PENDING.md`, `CURRENT.md` and `DONE.md` at the repository
root, and `.carves/` — the private carve directory — whenever it exists. `.carves/`
is ignored under every git-managed value, since a directory whose whole purpose is
privacy does not become public because the queue did. Public `carves/` is never
ignored.

**`track` is not inferable, which is why the flag is mandatory.** Whether a
repository *is* git-managed the tool can check; whether you *want* your queue to ship
it cannot, and a default would make that choice for someone without their noticing.

**`track` is a checked copy, not a stored claim.** It duplicates what `.gitignore`
already says, and that duplication is deliberate: a copy re-verified on every
non-trivial run cannot drift undetected, which was the whole of the objection to
keeping it. A stored claim rots; a checked one is a checksum. It also earns its keep
independently, because the creation path needs the policy months after `init`,
possibly before any ignore line exists to read it from.

## `minispec init --track-STYLE`

The flag is mandatory and takes one of `--track-none`,
`--track-private-trajectory`, `--track-all`.

`init` is the **sole creator** of `.minispec/config.yaml`. No other command brings it
into existence, so there is no path to a configuration with no `track` value in it.

It creates:

- `<repo root>/.minispec/config.yaml`, holding `track` and nothing else it does not
  need. A basic file, not a commented template of every available setting.
- an ignore line for `<repo root>/.minispec/backup` in the top-level `.gitignore`,
  under any git-managed `track` value.
- under `private-trajectory`, ignore lines for the trajectory files.

**It adds no line for a path git already ignores.** Whether a path is covered is a
question for git, not for text matching: `/PENDING.md`, `PENDING.md` and `*.md` are the
same intent written three ways, and only git can tell you the third one counts. This is
the same reason the consistency check shells out rather than parsing `.gitignore` — the
rules compose in ways a naive reader gets wrong.

**The lines it writes are anchored** — `/PENDING.md`, not `PENDING.md`. Every path
`track` governs is mandated at the repository root, and an unanchored pattern also
matches `docs/PENDING.md` and every nested one. Since the last matching pattern wins,
appending a bare form to a file that already had the anchored one does not merely
duplicate it; it silently *widens* a rule the project had written narrowly.

It reports every file it created or edited, in full. An agent that mutates state
through a tool and then cannot see the result will reason from a stale picture and,
eventually, assert it.

**`init` refuses when `.minispec/` already exists**, because running would be
stomping on something. The refusal names `--repair` as the way to change a `track`
value, and says to confirm with the user first.

**`init` is for the user to run.** The tool never runs it implicitly and no skill
tells an agent to run it, so the only ways an agent reaches it are a direct request
from the user or the no-configuration refusal below — whose entire purpose is
obtaining exactly that assent. This is an absence, not a prohibition: there is no rule
forbidding an agent from running `init`, because a rule is text that can be misread or
argued around, and because refusing a user's direct request would be worse behavior
than running the command.

## `minispec init --track-STYLE --repair`

`--repair` is the way an existing project changes its `track` value or resolves a
mismatch. It is **symmetric**: it sets `track` to the given value *and* brings
`.gitignore` into agreement with it.

Symmetry is what lets one command resolve a mismatch in either direction.
`private-trajectory` with unignored trajectory files can be fixed by adding the ignore
lines or by declaring the files public — those mean opposite things, and the flag value
is how the user chooses between them.

**That choice is the user's, so the agent confirms before running it.** The tool
detects, the agent relays, the human decides.

**`--repair` sets one key and disturbs nothing else.** It preserves the configuration's
comments, the order of its keys, and every setting the tool does not model. The obvious
implementation — read the file into the settings this binary knows about, then write
those settings back out — silently deletes all three, and the losses run opposite to
their importance: the comments are the reasoning a human left for the next reader, and
an unmodelled key is a setting from a *newer* tool version, discarded by an older binary
without a word. A repair that quietly narrows a configuration to what the running binary
happens to understand is the same silent partial success this layer refuses everywhere
else.

The one thing it does not promise is byte-fidelity: a blank line separating a comment
from what it annotates is not preserved. Comments survive, their attachment survives,
the spacing between them does not.

**What repair cannot reach, it names.** Making a path public means deleting the line
that ignores it, and the tool only owns the top-level `.gitignore`. A path ignored by a
nested `.gitignore`, by `.git/info/exclude`, or by the user's global excludes stays
ignored however many lines are removed here. The tool says so and names the path,
rather than reporting a repair that half-worked — a partial success that announces
itself as a success is the failure this whole layer exists to prevent.

Plain `init` and `init --repair` are mutually exclusive on one precondition: plain
`init` requires that `.minispec/` be absent, `--repair` requires that it be present.
Neither has to guess the caller's intent.

## The startup consistency check

Every command doing more than reporting its version verifies `track` against two
facts before doing its work:

1. **Is the repository git-managed?** `none` requires that it is not; the other two
   require that it is.
2. **Does the ignore state agree?** Under `private-trajectory` the trajectory files
   must be ignored; under `all` they must not be; `.carves/` must be ignored whenever
   it exists.

This catches the real drift — a project that gains git after `init`, or one whose
ignore lines were edited by hand.

**A mismatch gripes and exits.** The message states which fact disagrees and names
`minispec init --track-VALUE --repair` as the repair, with the instruction to confirm
the value with the user first. The agent does not hand-edit the configuration.

**The gripe fires on every run until satisfied**, and that is right rather than
excessive: a one-shot reminder decays to nothing, which this project has measured on
itself. It also cannot become wallpaper, because it is *closable* — repair once and it
is gone. A recurring nag you can never discharge is a different animal.

In a tree with no git at all, neither this check nor the preferences below mean
anything, so the tool stays quiet.

## Running without a configuration

With no `.minispec/config.yaml`, the only commands that run are `init` and those
needing no mini-spec files at all — `--version`, `help`, and `check-version`.

`check-version` is deliberately among them. A skill's first instruction is to run it,
and a version check that refuses in an uninitialized project would leave an agent
unable to establish even whether its tool matches its skill before being told what to
do about it.

Everything else refuses with a crank handle. What the refusal says is shaped by one
rule: **never ask a human a question the agent can answer.** The user's irreducible
contribution is *intent* — "yes, I want this to be a mini-spec project." Whether there
is code in this directory is a *fact*, and the agent can go and look.

That gives three tiers, each doing only what the others cannot: **the tool detects the
missing configuration, the agent inspects the directory, the human supplies the
intent.**

The gist the tool prints:

```
It doesn't look like this is a minispec project.
If <absolute path> is a code project, would you like me to make it work with minispec?
```

Four things this message must get right.

**It names the absolute path** it would make the repository root. The user's yes is to
*make this a mini-spec project*, not to *here specifically*, and the assent has to land
on a stated location. Standing one directory below your project, in a tree with no
version control to walk up to, is rare, recoverable, and otherwise silent.

**The most valuable finding is the negative one.** This message exists because someone
ran the tool from *above* their project, and a directory containing many project
directories is exactly what an agent can recognise. "This looks like a folder of
projects rather than a project — did you mean one of these?" is far more useful than a
yes, and it is the case that would otherwise plant a repository root in the wrong
place.

**The instruction must include the stop.** The moment it says "check whether this looks
like a code project," an agent can conclude *yes* and run `init` — building precisely
the road left deliberately unbuilt. The instruction ends with **report and wait**. This
is the one place an agent is told to go and look at something ahead of a user decision,
so it is the one place the boundary is spelled out rather than left as an absence.

**The tool gives the gist; the agent writes the sentence.** A verbatim script cannot
adapt to what the agent found — it would still say "if this is a code project" after the
agent has established that it plainly is, or plainly is not. So the tool is always
agent-facing and the human-facing sentence is always composed.

**Answered "no"**, the refusal says to run from inside the project. The search only
walks up, so a marker below would already have been found; the user is outside their
project, most likely above it. Creating anything here would put a repository root in
the wrong place, and it is durable enough that the mistake would persist.

## A configuration that predates `track`

A configuration that parses but sets no `track` is **not malformed**. It is an older
schema, and `--repair` is precisely the verb for it.

The line between the two cases is one sentence: **absence is a version difference; a
wrong value is damage.** A key that was never written is what every configuration
predating this setting has, and a flag supplies exactly the thing that is missing. A
value outside the closed set is a character somebody typed, and a flag that sets one
value cannot know what else the same hand changed.

That line is not new. `--repair` has always drawn it — it accepts a configuration with
no `track` and refuses one whose value it cannot parse. What was missing is the startup
check drawing it in the *same place*, so that the two agree about which configurations a
flag can fix. A tool that refuses a repair its own repair path would have accepted sends
the agent to hand-edit the file, which sets `track` and leaves `.gitignore` untouched —
half a repair, reported as the only one available.

**The refusal asks for the intent and nothing else.** Unlike the no-configuration case,
there is nothing here for the agent to go and establish: a configuration exists, so this
is settled as a mini-spec project and the repository root is already known. The single
irreducible question is the one `track` was made mandatory for — does this project's
work queue stay private or ship with the repository — and only the user can answer it.
So the message carries the question, the stop, and the command, and skips the inspection
step that the no-configuration refusal needs.

**The stop is the same and for the same reason.** An agent that picks a `track` value on
the user's behalf is manufacturing assent, and the two answers mean opposite things
about what the repository publishes. Report and wait.

## A malformed configuration

`--repair` validates well-formedness before doing anything, and a configuration that
fails validation cannot be repaired by a flag — the damage is arbitrary, most likely a
stray character typed while the file was open in an editor.

So this is the **one place the agent is explicitly authorised to edit the
configuration**. The refusal states every problem found, points at the configuration
documentation in the skill directory, and says so outright, because at that point the
agent is the only actor left who can act.

The agent backs the file up before editing, and skips the backup when it would be
byte-identical to one already there.

## Git preferences

Two properties the tool prefers and gripes about while unmet:

- `.minispec/config.yaml` **should be tracked** by git.
- `.minispec/backup` **should be ignored** by git.

Both are computed — `git ls-files` and `git check-ignore` answer them — so there is
nothing to assert and nothing to go stale.

The stamp and working files live in `.minispec/backup`, so one ignore line covers
everything machine-local and `config.yaml` is the only thing left in `.minispec/`.
"The tracked one is the only one you can see" then holds by structure rather than by
remembering two paths.

## What the tool does not do to git

**The tool does not manipulate the git repository** — no staging, no commits, no
resets, no branch state. Editing `.gitignore` is not manipulation in that sense; it is
editing a file, not changing git's state.

**Git only, through the `git` command line.** No second version-control system and no
linked-in library: supporting another VCS means knowing how to *operate* it, which is a
large surface for checks this small, while shelling out to `git` gets tracked and
ignored status for free and stays correct as git changes.

The consequence is a real narrowing and the tool states it rather than passing
silently: a project whose files are managed by something other than git gets no ignore
checking, and is told so. `track: none` is the honest declaration for such a project —
it says "not git-managed," which is true, rather than claiming an ignore state nothing
can verify.

## Consequence: existing projects must run `init` once

Because nothing but `init` creates a configuration and nothing but version-like
commands run without one, **every project already using mini-spec must run `init` once**
before other commands work. That is the adoption cost of making `track` mandatory, and
it is the point rather than a side effect: the one-time refusal is how a project that
predates `track` is asked the question, and the crank handle is what carries the ask.

**There are two such populations, and they take opposite forms of the command.** A
project with no `.minispec/` at all meets the no-configuration refusal and runs plain
`init`. A project that already has a `.minispec/config.yaml` written before `track`
existed meets the pre-`track` refusal and runs `--repair`, because plain `init` refuses
where `.minispec/` is present. Both are asked the same question and neither can be
inferred; only the verb differs, and the refusal each meets is the one that names the
right verb.

The second population exists because `init` is the sole creator *now* but was not
always: a configuration predating the command was necessarily written by hand. That
makes the population finite and knowable rather than open-ended — but it includes this
tool's own reference projects, which is where the case was found.

## `minispec init carve <name>`

Scaffolds `carves/<name>.md` in the shape the format mandates: the title line, a paragraph
placeholder for the problem, a `## Status` block holding one open unqueued part, a
`## Decisions` section, and the part's elaboration stub keyed `**Item 1**`. It creates `carves/`
when absent, refuses an existing file — a scaffold never overwrites a carve — and refuses a name
that is a path or carries an extension. The write is reported with its git kind, and the report
names the `add-item` that queues the first part (R334).

*Why a scaffold rather than a paragraph in the skill.* Ark's six carves were retrofitted by
hand in one afternoon; three had independently invented a status notation and three had none,
because there was nothing to copy from except prose, and prose gets paraphrased. This is the
only creation verb: the trajectory files are created by being refused
([queue-items.md](queue-items.md), *Creation*), since a write path leads to them and none
leads to a carve that does not exist yet.
