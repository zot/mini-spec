# Test Design: Init
**Source:** crc-Init.md

Every case runs against a temporary directory and a **fake `Git`**, so no repository is
created and no `git` is invoked. What `init` writes is a function of two inputs — the
chosen value and whether the tree is git-managed — which makes the whole matrix cheap
to state.

The refusal cases matter as much as the creating ones: three of this card's
requirements are about what it declines to do.

## Test: the track flag is mandatory
**Purpose:** validates that the value is never inferred or defaulted — the decision the
tool must not make for the user
**Input:** `init` with no `--track-*` flag; and with two of them
**Expected:** both refuse; neither writes anything
**Refs:** seq-bootstrap.md#2.1 — R136

## Test: creation writes the configuration with the chosen value
**Purpose:** the happy path, and that the value round-trips
**Input:** `init --track-private-trajectory` in a tree with no `.minispec/`
**Expected:** `.minispec/config.yaml` exists and parses with `track:
private-trajectory`
**Refs:** seq-bootstrap.md#2.4 — R137, R138

## Test: `private-trajectory` writes both kinds of ignore line
**Purpose:** validates the value that ignores the most
**Input:** `init --track-private-trajectory`, fake reports a working tree
**Expected:** `.gitignore` gains `.minispec/backup` and all three trajectory filenames
**Refs:** seq-bootstrap.md#2.6, seq-bootstrap.md#2.7 — R139, R140

## Test: `all` writes only the backup ignore line
**Purpose:** the discriminating case — `all` still hides the tool's machine-local files
**Input:** `init --track-all`, fake reports a working tree
**Expected:** `.gitignore` gains `.minispec/backup` and no trajectory filename
**Refs:** seq-bootstrap.md#2.6 — R139

## Test: `none` writes no ignore lines at all
**Purpose:** validates that a project with no repository is left alone rather than told
about files it does not have
**Input:** `init --track-none`, fake reports no working tree
**Expected:** `config.yaml` written; no `.gitignore` created or edited
**Refs:** seq-bootstrap.md#2.5 — R139, R151

## Test: an existing `.gitignore` is appended to, not replaced
**Purpose:** the tool edits a file it does not own, so clobbering it is the failure that
would cost a user the most
**Input:** a `.gitignore` with unrelated lines; `init --track-private-trajectory`
**Expected:** the original lines survive verbatim and the new lines are added
**Refs:** seq-bootstrap.md#2.6 — R139, R140

## Test: ignore lines are not duplicated
**Purpose:** validates that an already-correct `.gitignore` is left as it is
**Input:** a `.gitignore` already containing every line the value requires; `init`
**Expected:** the file is unchanged, and the report says so rather than claiming an edit
**Refs:** seq-bootstrap.md#2.6 — R139, R141

## Test: an already-ignored path is not duplicated
**Purpose:** **the defect this caught the first time `init` ran on a real repository.**
mini-spec's own `.gitignore` carried the anchored form `/PENDING.md`; textual line
matching did not recognise it and appended the bare form, which — since the last
matching pattern wins — replaced the project's narrow rule with one that also ignores
`docs/PENDING.md`. Two failures in one line: a redundant entry and a widened rule
**Input:** a `.gitignore` holding `/PENDING.md`, `/CURRENT.md`, `/DONE.md`; a fake git
reporting all three ignored; `init --track-private-trajectory`
**Expected:** each filename appears exactly **once**. Counting rather than asserting
presence is the whole point — the broken version passed a `Contains` check
**Refs:** seq-bootstrap.md#2.6 — R170, R171

## Test: a path ignored by a wildcard is not duplicated
**Purpose:** the case no textual match could ever catch, and so the one that pins *why*
the question goes to git rather than to the file
**Input:** a `.gitignore` holding only `*.md`; a fake reporting the trajectory files
ignored; `init --track-private-trajectory`
**Expected:** no trajectory filename is added
**Refs:** seq-bootstrap.md#2.6 — R170

## Test: added ignore lines are anchored
**Purpose:** every governed path is mandated at the repository root, so a bare pattern
reaches further than anyone asked
**Input:** `init --track-private-trajectory` into an empty tree
**Expected:** every line written begins with `/`
**Refs:** seq-bootstrap.md#2.7 — R171

## Test: repair reports what it could not make public
**Purpose:** this command owns the top-level `.gitignore` and nothing else, so a rule in
a nested `.gitignore`, `.git/info/exclude`, or the user's global excludes is out of
reach. A repair that half-worked and announced success is the silent partial success
this layer exists to prevent
**Input:** a project at `all`; a fake insisting the trajectory files are ignored while
the top-level file holds no line naming them; `--repair --track-all`
**Expected:** all three are named in `Unresolved`, and the report says `STILL IGNORED`
**Refs:** seq-bootstrap.md#3.8 — R172

## Test: plain `init` refuses when `.minispec/` exists
**Purpose:** the precondition that keeps the two forms apart, and the refusal that must
not stomp
**Input:** a tree already holding `.minispec/config.yaml`; plain `init --track-all`
**Expected:** refuses, writes nothing, and the message names `--repair` and says to
confirm with the user
**Refs:** seq-bootstrap.md#2.3 — R142, R146

## Test: `--repair` refuses when `.minispec/` does not exist
**Purpose:** the inverted precondition, checked in the other direction
**Input:** an empty tree; `init --track-all --repair`
**Expected:** refuses, writes nothing, and points at plain `init`
**Refs:** seq-bootstrap.md#3.2 — R145

## Test: repair changes the value and adds the lines the new value requires
**Purpose:** the forward half of symmetry
**Input:** a project at `all`; repair to `private-trajectory`
**Expected:** `track` becomes `private-trajectory` and the trajectory filenames appear
in `.gitignore`
**Refs:** seq-bootstrap.md#3.6 — R144

## Test: repair removes the lines the new value forbids
**Purpose:** the reverse half — the direction a one-directional repair could never
reach, and the reason this command can resolve a mismatch either way
**Input:** a project at `private-trajectory` with the trajectory files ignored; repair
to `all`
**Expected:** `track` becomes `all` and the trajectory filenames are gone from
`.gitignore`, with unrelated lines untouched
**Refs:** seq-bootstrap.md#3.7 — R144

## Test: repair refuses a malformed configuration
**Purpose:** validates the precondition behind the one place the agent may edit the
config — a flag cannot undo arbitrary damage
**Input:** a `config.yaml` containing unparseable YAML; `init --track-all --repair`
**Expected:** refuses, writes nothing, states the parse problem, points at the skill's
configuration documentation, and says the agent may edit the file
**Refs:** seq-bootstrap.md#3.3 — R161, R162

## Test: the report names every file created or edited
**Purpose:** validates the crank handle — the agent did not write these files and must
not have to infer what changed
**Input:** `init --track-private-trajectory` in a tree with a working tree
**Expected:** the output names `.minispec/config.yaml` and `.gitignore`
**Refs:** seq-bootstrap.md#2.8 — R141

## Test: `--repair` accepts a configuration with no `track`
**Purpose:** validates R173 from the repair side — absence is what this verb is for, so
it must not be refused as damage
**Input:** a `config.yaml` holding only `design_dir`; `init --track-all --repair`
**Expected:** succeeds, and `LoadTrack` afterwards reads `all`
**Alarm:** 1
**Fire alarm:** make `validateWellFormed` reject an empty `Track` as well as an
unparseable one, and confirm this goes red
**Inject:** internal/project/init.go:validateWellFormed
**Pulled:** 2026-08-11 — covered by the LoadTrack injection above; not pulled alone
**Refs:** seq-bootstrap.md#3.3 — R173

## Test: repair preserves comments, key order, and unmodelled settings
**Purpose:** validates R177 — the writer edits one key and disturbs nothing else. The
suite asserted what the file *gained* and never what it kept, which is why the defect
shipped
**Input:** a `config.yaml` with a two-line comment block, a quoted `comment_patterns`
entry, and a `future_setting: 42` this binary does not model; `--repair`
**Expected:** every comment, the quoted form, the unknown key and the original key
order all survive, with `track` appended last
**Alarm:** 2
**Fire alarm:** restore the struct round-trip — unmarshal into `Config`, set `Track`,
`yaml.Marshal` back — and confirm every preserved string is reported dropped. This is
the literal defect, measured on ark: it deleted a ten-line comment block explaining a
non-obvious setting
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-08-11 — rang: five dropped strings reported, restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R177

## Test: changing an existing value keeps the comment on it
**Purpose:** the *replace* branch, which the append case does not reach — a repair in
either direction must edit the one scalar and leave its annotation standing
**Input:** a `config.yaml` with a comment above `track: all`; `--repair` to
`private-trajectory`
**Expected:** the comment survives, the value changes, the neighbouring setting is
untouched
**Alarm:** 3
**Fire alarm:** replace the value node wholesale instead of setting it in place, and
clear the key's head comment — confirm the comment is reported dropped
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-08-11 — rang, restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R177

## Test: a no-op repair rewrites nothing
**Purpose:** validates that an already-correct value leaves the file byte-for-byte
alone, so a repair cannot reformat what it had no reason to touch
**Input:** a `config.yaml` with irregular spacing (`track:   all`) and a comment;
`--repair` to the same value
**Expected:** the file is byte-identical afterwards and the report says `unchanged`
**Alarm:** 4
**Fire alarm:** delete the already-correct early return from `setTrack` so it always
re-encodes, and confirm the byte comparison goes red on the collapsed spacing
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-08-11 — rang, restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R177

## Test: the degenerate documents do not panic or emit `null`
**Purpose:** an absent, empty or comment-only file parses to a document with no
content, which the node walk must not dereference
**Input:** `""`, `"\n\n"`, and a comments-only body
**Expected:** each yields a document setting `track`, and none contains `null`
**Alarm:** 5
**Fire alarm:** disable the empty/non-mapping guard and confirm an `index out of
range [0] with length 0` panic
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-08-11 — rang (panicked as predicted), restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R177
