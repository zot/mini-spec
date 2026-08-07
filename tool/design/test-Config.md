# Test Design: Config Scopes
**Source:** crc-Project.md

Temporary trees again, no fixtures. The three merge rules are pure functions over
parsed structures, so each is exercised directly; the layer resolution needs a tree
only because it reads files.

Every tree carries a strong root marker so repository-root detection is not itself
under test here — `test-RepoRoot.md` owns that.

## Test: the repository config applies to a design root beneath it
**Purpose:** validates the basic two-scope case — one repository, one project below
**Input:** `root/.git/`, `root/.minispec/config.yaml` setting `src_dir: lib`,
`root/tool/design/` with no `.minispec.yaml`
**Expected:** the design root at `root/tool` resolves `src_dir` to `lib`
**Refs:** seq-config.md#1.4 — R118, R121

## Test: a design root's file overrides the repository config
**Purpose:** validates layer precedence
**Input:** repository config setting `src_dir: lib`; `root/tool/.minispec.yaml`
setting `src_dir: source`
**Expected:** `src_dir` resolves to `source`
**Refs:** seq-config.md#1.5, seq-config.md#2.1 — R124, R125

## Test: a design root with no file inherits everything
**Purpose:** validates that a matching project needs no file at all
**Input:** repository config setting `src_dir: lib` and a `comment_patterns` entry;
`root/tool/` with no `.minispec.yaml`
**Expected:** both resolve from the repository layer
**Refs:** seq-config.md#1.5 — R121

## Test: maps merge per key
**Purpose:** the rule that lets a project add one pattern without restating the rest
**Input:** repository `comment_patterns` sets `.go` and `.ts`; design root sets
`.html` only
**Expected:** all three present, `.go` and `.ts` unchanged
**Refs:** seq-config.md#2.2 — R126

## Test: a map key set by both layers takes the design root's value
**Purpose:** validates that merging per key still overrides at the key
**Input:** repository sets `.go` to one pattern; design root sets `.go` to another
**Expected:** the design root's value wins, other keys untouched
**Refs:** seq-config.md#2.2 — R126

## Test: lists union rather than replace
**Purpose:** the rule Bill chose so a project states additions, not the whole list
**Input:** repository `code_extensions: [.go, .ts]`; design root `[.lua]`
**Expected:** `[.go, .ts, .lua]` — inherited order preserved, addition appended
**Refs:** seq-config.md#2.3 — R127

## Test: list duplicates are dropped
**Purpose:** validates that repeating an inherited entry is harmless
**Input:** repository `[.go, .ts]`; design root `[.ts, .lua]`
**Expected:** `[.go, .ts, .lua]` — `.ts` appears once, in its inherited position
**Refs:** seq-config.md#2.3 — R127

## Test: the first configuration layer replaces the built-in defaults
**Purpose:** validates that a project can still narrow a shipped list — the
capability union-everywhere would have silently removed
**Input:** repository config setting `code_extensions: [.go]`, no design-root file
**Expected:** exactly `[.go]`, not the shipped defaults plus `.go`
**Refs:** seq-config.md#2.3 — R130

## Test: a design root alone also replaces the defaults
**Purpose:** validates the exception applies to whichever layer speaks first, not
specifically to the repository
**Input:** no repository layer; design root sets `code_extensions: [.lua]`
**Expected:** exactly `[.lua]`
**Refs:** seq-config.md#2.3 — R130

## Test: a repository-root .minispec.yaml is rejected
**Purpose:** the flat rule, on the shape it exists to forbid
**Input:** `root/.git/`, `root/.minispec/config.yaml`, and `root/.minispec.yaml`
**Expected:** an error naming the forbidden path
**Refs:** seq-config.md#1.2 — R122

## Test: rejection catches a symlink
**Purpose:** validates that a link to the new location is the forbidden shape rather
than a supported alias — the exact transition scaffolding left in the reference
project
**Input:** `root/.minispec/config.yaml` with `root/.minispec.yaml` symlinked to it
**Expected:** an error naming the forbidden path, identical to the regular-file case
**Refs:** seq-config.md#1.2 — R123

## Test: rejection precedes reading, so contents never matter
**Purpose:** validates that the error is about presence, not parseability
**Input:** `root/.minispec.yaml` containing invalid YAML
**Expected:** the forbidden-path error, not a parse error
**Refs:** seq-config.md#1.2 — R122

## Test: rejection applies when the roots coincide
**Purpose:** the "no exception" half of the rule, on the layout where someone would
most expect an exception
**Input:** `root/.git/`, `root/design/`, `root/.minispec/config.yaml`, and
`root/.minispec.yaml` — repository root and design root are the same directory
**Expected:** an error; the coincident case is not privileged
**Refs:** seq-config.md#1.2 — R120, R122

## Test: no repository root means no repository layer
**Purpose:** validates that pre-existing single-project behavior survives
**Input:** a marker-free tree with `design/` and a `.minispec.yaml` beside it
**Expected:** the design root's file applies directly over the built-in defaults; no
error
**Refs:** seq-config.md#1.1 — R124

## Test: provenance names the file each setting came from
**Purpose:** validates that a value's origin is answerable without reading two files
**Input:** repository sets `src_dir` and a `.go` comment pattern; design root sets
`design_dir` and a `.html` comment pattern
**Expected:** each setting reports its originating layer — repository, design root,
or built-in default for anything neither set
**Refs:** seq-config.md#2.4 — R127, R129
