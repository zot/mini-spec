# Test Design: Config Scopes
**Source:** crc-Project.md

Temporary trees again, no fixtures. The three merge rules are pure functions over
parsed structures, so each is exercised directly; the layer resolution needs a tree
only because it reads files.

Every tree carries a strong root marker so repository-root detection is not itself
under test here — `test-RepoRoot.md` owns that.

## Test: the repository config applies to a design root beneath it
**Purpose:** validates the basic two-scope case — one repository, one project below
**Input:** `root/.git/`, `root/.minispec/config.toml` setting `src_dir = "lib"`,
`root/tool/design/` with no `.minispec.toml`
**Expected:** the design root at `root/tool` resolves `src_dir` to `lib`
**Refs:** seq-config.md#1.4 — R118, R121

## Test: a design root's file overrides the repository config
**Purpose:** validates layer precedence
**Input:** repository config setting `src_dir = "lib"`; `root/tool/.minispec.toml`
setting `src_dir = "source"`
**Expected:** `src_dir` resolves to `source`
**Refs:** seq-config.md#1.5, seq-config.md#2.1 — R124, R125

## Test: a design root with no file inherits everything
**Purpose:** validates that a matching project needs no file at all
**Input:** repository config setting `src_dir = "lib"` and `code_extensions`;
`root/tool/` with no `.minispec.toml`
**Expected:** both resolve from the repository layer
**Refs:** seq-config.md#1.5 — R121

## Test: lists union rather than replace
**Purpose:** the rule Bill chose so a project states additions, not the whole list
**Input:** repository `code_extensions = [".go", ".ts"]`; design root `[".lua"]`
**Expected:** `[.go, .ts, .lua]` — inherited order preserved, addition appended
**Refs:** seq-config.md#2.3 — R127

## Test: list duplicates are dropped
**Purpose:** validates that repeating an inherited entry is harmless
**Input:** repository `[".go", ".ts"]`; design root `[".ts", ".lua"]`
**Expected:** `[.go, .ts, .lua]` — `.ts` appears once, in its inherited position
**Refs:** seq-config.md#2.3 — R127

## Test: the first configuration layer replaces the built-in defaults
**Purpose:** validates that a project can still narrow a shipped list — the
capability union-everywhere would have silently removed
**Input:** repository config setting `code_extensions = [".go"]`, no design-root file
**Expected:** exactly `[.go]`, not the shipped defaults plus `.go`
**Refs:** seq-config.md#2.3 — R130

## Test: a design root alone also replaces the defaults
**Purpose:** validates the exception applies to whichever layer speaks first, not
specifically to the repository
**Input:** no repository layer; design root sets `code_extensions = [".lua"]`
**Expected:** exactly `[.lua]`
**Refs:** seq-config.md#2.3 — R130

## Test: a repository-root .minispec.toml is rejected
**Purpose:** the flat rule, on the shape it exists to forbid
**Input:** `root/.git/`, `root/.minispec/config.toml`, and `root/.minispec.toml`
**Expected:** an error naming the forbidden path
**Refs:** seq-config.md#1.2 — R122

## Test: rejection catches a symlink
**Purpose:** validates that a link to the new location is the forbidden shape rather
than a supported alias — the exact transition scaffolding left in the reference
project
**Input:** `root/.minispec/config.toml` with `root/.minispec.toml` symlinked to it
**Expected:** an error naming the forbidden path, identical to the regular-file case
**Refs:** seq-config.md#1.2 — R123

## Test: rejection precedes reading, so contents never matter
**Purpose:** validates that the error is about presence, not parseability
**Input:** `root/.minispec.toml` containing invalid TOML
**Expected:** the forbidden-path error, not a parse error
**Refs:** seq-config.md#1.2 — R122

## Test: rejection applies when the roots coincide
**Purpose:** the "no exception" half of the rule, on the layout where someone would
most expect an exception
**Input:** `root/.git/`, `root/design/`, `root/.minispec/config.toml`, and
`root/.minispec.toml` — repository root and design root are the same directory
**Expected:** an error; the coincident case is not privileged
**Refs:** seq-config.md#1.2 — R120, R122

## Test: no repository root means no repository layer
**Purpose:** validates that pre-existing single-project behavior survives
**Input:** a marker-free tree with `design/` and a `.minispec.toml` beside it
**Expected:** the design root's file applies directly over the built-in defaults; no
error
**Refs:** seq-config.md#1.1 — R124

## Test: provenance names the file each setting came from
**Purpose:** validates that a value's origin is answerable without reading two files
**Input:** repository sets `src_dir` and `code_extensions`; design root sets
`design_dir` and adds to `code_extensions`
**Expected:** each setting reports its originating layer — repository, design root,
or built-in default for anything neither set
**Refs:** seq-config.md#2.4 — R127, R129

## Test: an unknown key stops the command, naming the file and the key
**Purpose:** a key the tool does not read would be a line with no effect and no way to find out
**Input:** a repository config setting `src_dir = "lib"` and `srcdir = "x"`
**Expected:** loading fails with an error naming `.minispec/config.toml` and `srcdir`; nothing
resolves
**Code:** internal/project/config_test.go
**Alarm:** 1
**Fire alarm:** skip the undecoded-key loop in `decodeLayer`, so an unknown key decodes silently,
and confirm the test reports no error where one was expected
**Inject:** internal/project/config.go:decodeLayer
**Pulled:** 2026-09-29 — rang: with `decodeLayer`'s undecoded-key check removed, `expected an error for the unknown key srcdir`; restore clean (empty diff), delegated
**Refs:** seq-config.md#2.1.1 — R521

## Test: a retired key is named as retired
**Purpose:** a configuration carried over from the YAML era still sets `comment_patterns`; the
message says what replaced it rather than only that the key is unknown
**Input:** a design-root `.minispec.toml` holding a `[comment_patterns]` table
**Expected:** an error naming the file and `comment_patterns` as retired, pointing at
`query comment-patterns`
**Refs:** seq-config.md#2.1.1 — R521

## Test: a YAML repository configuration stops everything
**Purpose:** the format change has no reader for the old file, so it must be reported rather
than passed over as "no configuration"
**Input:** `root/.git/` and `root/.minispec/config.yaml`, no `config.toml`
**Expected:** an error naming `.minispec/config.yaml` and saying to convert it to TOML by hand;
not the no-configuration refusal
**Code:** internal/project/config_test.go
**Alarm:** 2
**Fire alarm:** make `LegacyConfigError` return nil, so the YAML file is passed over as though
there were no configuration, and confirm both YAML tests go red
**Inject:** internal/project/config.go:LegacyConfigError
**Pulled:** 2026-09-29 — rang: with `LegacyConfigError` returning nil, four tests red, `TestYAMLRepoConfigIsReported: expected an error naming the YAML configuration` first; restore clean (empty diff), delegated
**Refs:** seq-config.md#1.2.1 — R522

## Test: a YAML design-root configuration stops everything
**Purpose:** the same rule at the design-root scope, where the old file sits beside `design/`
**Input:** a repository `config.toml` and `root/tool/.minispec.yaml`
**Expected:** an error naming `tool/.minispec.yaml` and saying to convert it; the repository
layer is not applied on its own as though the design root had no file
**Refs:** seq-config.md#1.2.1 — R522

## Test: languages layer by name
**Purpose:** a design root redefines one language without restating the others
**Input:** repository defines languages `a` and `b`; design root defines `b` differently and `c`
**Expected:** `a` from the repository, `b` from the design root whole, `c` added
**Refs:** seq-config.md#2.5 — R527

## Test: a language sdom rejects stops the load, naming the file
**Purpose:** a malformed definition is reported where it was written
**Input:** a repository config whose `[[languages]]` group sets both `close` and `close_regex`
**Expected:** an error naming the file, the language and the group
**Refs:** seq-config.md#2.1.1 — R528
