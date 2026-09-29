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
**Expected:** `.minispec/config.toml` exists and decodes with `track =
"private-trajectory"`
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
**Expected:** `config.toml` written; no `.gitignore` created or edited
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
**Input:** a tree already holding `.minispec/config.toml`; plain `init --track-all`
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
**Input:** a `config.toml` containing unparseable TOML; `init --track-all --repair`
**Expected:** refuses, writes nothing, states the parse problem, points at the skill's
configuration documentation, and says the agent may edit the file
**Refs:** seq-bootstrap.md#3.3 — R161, R162

## Test: the report names every file created or edited
**Purpose:** validates the crank handle — the agent did not write these files and must
not have to infer what changed
**Input:** `init --track-private-trajectory` in a tree with a working tree
**Expected:** the output names `.minispec/config.toml` and `.gitignore`
**Refs:** seq-bootstrap.md#2.8 — R141

## Test: `--repair` accepts a configuration with no `track`
**Purpose:** validates R173 from the repair side — absence is what this verb is for, so
it must not be refused as damage
**Input:** a `config.toml` holding only `design_dir`; `init --track-all --repair`
**Expected:** succeeds, and `LoadTrack` afterwards reads `all`
**Alarm:** 1
**Fire alarm:** make `validateWellFormed` reject an empty `Track` as well as an
unparseable one, and confirm this goes red
**Inject:** internal/project/init.go:validateWellFormed
**Pulled:** 2026-08-11 — covered by the LoadTrack injection above; not pulled alone
**Refs:** seq-bootstrap.md#3.3 — R173

## Test: repair preserves every byte outside the `track` line
**Purpose:** validates R524 — the writer edits one line and disturbs nothing else. The
suite once asserted what the file *gained* and never what it kept, which is why the
YAML-era defect shipped
**Input:** a `config.toml` with a two-line comment block, a key this binary does not
model, blank lines, and no `track`; `--repair`
**Expected:** the file afterwards is the original with exactly one `track = "…"` line
added, every other byte unchanged
**Alarm:** 2
**Fire alarm:** restore the struct round-trip — decode into `Config`, set `Track`, encode
back with the TOML encoder — and confirm the comparison reports the comments and blank
lines dropped. This is the literal defect in its TOML form; in YAML it deleted a ten-line
comment block on ark
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-09-27 — rang: "repair changed more than one line"; restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R524

## Test: changing an existing value keeps its comments
**Purpose:** the *replace* branch, which the insert case does not reach — a repair in
either direction must change the value and leave the line's annotations standing
**Input:** a `config.toml` with a comment above `track = "all"  # why this value` and a
neighbouring setting; `--repair` to `private-trajectory`
**Expected:** the value changes; the comment above, the trailing `# why this value` and the
neighbouring setting are unchanged
**Alarm:** 3
**Fire alarm:** rewrite the whole line as `track = "<value>"` instead of replacing only the
quoted value, and confirm the trailing comment is reported dropped
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-09-27 — rang: "repair changed more than the value"; restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R524

## Test: a no-op repair rewrites nothing
**Purpose:** validates that an already-correct value leaves the file byte-for-byte
alone, so a repair cannot reformat what it had no reason to touch
**Input:** a `config.toml` with irregular spacing (`track   =   "all"`) and a comment;
`--repair` to the same value
**Expected:** the file is byte-identical afterwards and the report says `unchanged`
**Alarm:** 4
**Fire alarm:** delete the already-correct early return from `setTrack` so it always
rewrites the line, and confirm the test goes red on the missing `unchanged` report. The line
edit keeps the key's spacing, so rewriting a correct value leaves the bytes identical and only
the report can tell; under the YAML writer it was the byte comparison that rang
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-09-27 — rang: "no-op repair did not report the config as unchanged"; restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R524

## Test: an inserted `track` goes above the first table
**Purpose:** TOML assigns every key after a table header to that table, so a `track` line
appended at the end of a file with a `[[languages]]` entry would decode as that table's key
— the file would still *contain* `track`, which is all a text check asks
**Input:** a `config.toml` holding `design_dir = "design"` and a `[[languages]]` table, no
`track`; `--repair`
**Expected:** the `track` line lands before the `[[languages]]` header, and the file decodes
with a top-level `track` and a `languages` entry carrying no `track`
**Alarm:** 6
**Fire alarm:** append the new line at end of file instead of before the first table header,
and confirm the decode reads no top-level `track`
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-09-27 — rang: top-level track nil and the table gained `track = "all"`; restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R524

## Test: the degenerate documents gain `track` and nothing else
**Purpose:** an absent, empty or comment-only file has no line to replace and no table to
insert before
**Input:** `""`, `"\n\n"`, and a comments-only body
**Expected:** each gains one `track` line after its existing content and decodes with it
**Alarm:** 5
**Fire alarm:** with no table header to insert before, put the line at the start of the file
instead of after the existing content, and confirm the comments-only case goes red on order
**Inject:** internal/project/init.go:setTrack
**Pulled:** 2026-09-27 — rang: blank-lines, comments-only and no-final-newline cases all put the line first; restore byte-clean
**Refs:** seq-bootstrap.md#3.4 — R524

## Test: init reports a YAML configuration before its own refusals
**Purpose:** `init` is exempt from the gate, so it checks for itself; reported after the
existence tests, a YAML file would draw "nothing to repair" or "already exists" and send the
user the wrong way over a file they already have
**Input:** `.minispec/config.yaml` and no `config.toml`; plain `init`, then `init --repair`
**Expected:** both name the YAML file and say to convert it to TOML; no `config.toml` is written
**Code:** internal/project/init_test.go
**Alarm:** 7
**Fire alarm:** move the `LegacyConfigError` check in `RunInit` below the `.minispec/` existence
tests, and confirm the repair case reports "nothing to repair" instead
**Inject:** internal/project/init.go:RunInit
**Pulled:** 2026-09-27 — rang in both forms: plain init said "already exists", repair said "cannot read config.toml"; restore byte-clean
**Refs:** seq-bootstrap.md#2.1 — R522
