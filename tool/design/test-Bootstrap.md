# Test Design: Bootstrap
**Source:** crc-CLI.md

The gate's two refusal messages and its command classification, tested as what they
are: pure functions over strings and maps. No project, no repository, no filesystem.

These earn a test for a reason distinct from the rest of this work. The refusal text is
**load-bearing prose** — each clause closes a hole the others do not — and it is the
kind of thing a later editor tightens for readability. Deleting a clause breaks nothing
that compiles, produces no failing behavior anyone would notice, and quietly removes a
safeguard. A test is the only thing that objects.

## Test: only version-like commands are exempt
**Purpose:** validates the gate's classification in both directions. The gated side is
the one that matters: a command added later must fail safe by being absent from the
exempt list rather than accidentally present in a gated one
**Input:** the exempt set and the four gated verbs
**Expected:** `init`, `check-version`, `help`, `--version` and their aliases are exempt;
`query`, `update`, `validate`, `phase` are not
**Refs:** seq-bootstrap.md#1.1 — R152, R153

## Test: knownCommands covers every dispatched command
**Purpose:** an unrecognised command must be reported as unknown, not as a missing
project — a typo should not send someone to `init`
**Input:** every command the dispatcher handles, plus a nonsense one
**Expected:** all dispatched commands are known; the nonsense one is not
**Refs:** crc-CLI.md — R154

## Test: the no-configuration refusal names the absolute path
**Purpose:** the user's assent is to *make this a mini-spec project*, not to *here*, so
it has to land on a stated location
**Input:** a repository root path
**Expected:** the message contains it verbatim
**Refs:** seq-bootstrap.md#1.4 — R155

## Test: the refusal tells the agent to stop
**Purpose:** **the single clause most worth guarding.** The moment the message says
"check whether this is a code project," an agent can conclude *yes* and run `init`,
building the road left deliberately unbuilt
**Input:** any repository root
**Expected:** the message contains the report-and-wait instruction and forbids acting
on the agent's own conclusion
**Refs:** seq-bootstrap.md#1.4 — R157

## Test: the refusal carries each remaining clause
**Purpose:** the fact-checking instruction, the folder-of-projects hunt, the declined
answer, and the command to run — one assertion each, so a deletion names itself
**Input:** any repository root
**Expected:** every clause present, and the failure message says which requirement the
missing one served
**Refs:** seq-bootstrap.md#1.4 — R156, R158, R160

## Test: the no-root refusal forbids creating a marker
**Purpose:** the sibling refusal, whose specific temptation is different — an agent that
cannot find a root may try to manufacture one, planting a repository root wherever the
command happened to run
**Input:** a root-detection error
**Expected:** the underlying error survives into the message, and the message forbids
creating anything
**Refs:** seq-bootstrap.md#1.2 — R155

## Test: the track flags cover exactly the closed set
**Purpose:** three flags rather than one taking an argument, so a misspelled value is an
unknown flag rather than a silently-accepted string
**Input:** the flag table
**Expected:** exactly three, each named for the value it records
**Refs:** seq-bootstrap.md#2.1 — R136

## Test: the pre-`track` refusal names repair and not hand-editing
**Purpose:** validates R174 — this is the case the whole change exists to correct. The
malformed refusal sent the agent to hand-edit a file `--repair` accepts, and a hand
edit sets `track` while leaving `.gitignore` unreconciled
**Input:** `preTrackMessage` with a configuration path
**Expected:** names `--repair`, names the path, and carries **no** hand-edit
authorisation
**Alarm:** 1
**Fire alarm:** put the malformed refusal's "you are authorised to edit this file by
hand" clause into this message and confirm the negative assertion goes red
**Inject:** internal/cli/bootstrap.go:preTrackMessage
**Pulled:** 2026-08-11 — rang, restore byte-clean
**Refs:** seq-bootstrap.md#1.6.1 — R174

## Test: the pre-`track` refusal asks the intent and stops
**Purpose:** validates R175 and R176 — a configuration existing already settles what
the no-configuration refusal has to send the agent looking for, so this one asks only
the question no inspection can answer, and then stops
**Input:** `preTrackMessage`
**Expected:** carries the stop and the private-or-ships question; carries **neither**
of the no-configuration refusal's inspection clauses
**Alarm:** 2
**Fire alarm:** covered by the routing injection below, which swaps whole messages
**Inject:** internal/cli/bootstrap.go:preTrackMessage
**Refs:** seq-bootstrap.md#1.6.1 — R175, R176

## Test: absence and damage are routed to different refusals
**Purpose:** validates R174 at the branch rather than the message. This is the
regression that hides: `ErrNoTrack`'s own text names `--repair`, so sending absence
down the damage branch still prints something plausible while the intent question and
the stop silently vanish
**Input:** `trackRefusal` with `ErrNoTrack`, with a **wrapped** `ErrNoTrack`, and with
an unrelated malformed error
**Expected:** the first two return exactly `preTrackMessage(cfgPath)`; the third
returns the loader's own message unchanged
**Guard the guard:** whole-message equality rather than clause checks — the two tests
above own what each message *says*, this owns only which one an error earns, and
comparing whole messages additionally pins `cfgPath` being threaded through
**Alarm:** 3
**Fire alarm:** branch on `==` instead of `errors.Is` and confirm the wrapped row goes
red — that is the real-world slip. Separately, pass a literal in place of `cfgPath` and
confirm this goes red while both clause-style tests stay green, which is what proves
this assertion strictly stronger
**Inject:** internal/cli/bootstrap.go:trackRefusal
**Pulled:** 2026-08-11 — rang: both, separately, each restore byte-clean. A third
attempt deleted the branch outright, which orphaned an import and broke the *build* —
the test never ran, so it proved nothing and was reshaped
**Refs:** seq-bootstrap.md#1.6 — R174
