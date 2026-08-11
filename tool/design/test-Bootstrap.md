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
