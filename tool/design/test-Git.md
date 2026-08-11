# Test Design: Git
**Source:** crc-Git.md

The one component here that cannot be faked, because faking it would test the fake.
Every case builds a **real repository** in a temporary directory with `git init` and
asks the real `git` binary — cheap enough to run always, and skipped with a clear
message when `git` is not on `PATH` rather than silently passing.

These are the tests that keep watch on a dependency this tool does not control. `git
check-ignore` answering wrongly would be invisible everywhere else, because every other
component is tested against a fake that assumes this one is right.

## Test: a fresh repository reports as a working tree
**Purpose:** the baseline the whole consistency check rests on
**Input:** `git init` in a temporary directory
**Expected:** `IsRepo()` is true
**Refs:** crc-Git.md, seq-bootstrap.md#1.7 — R167

## Test: a bare directory does not report as a working tree
**Purpose:** validates that absence is an answer rather than an error — a project with
no version control is a supported shape
**Input:** a temporary directory with no `.git`
**Expected:** `IsRepo()` is false and no error is raised
**Refs:** crc-Git.md — R151, R168

## Test: a subdirectory of a repository still reports as a working tree
**Purpose:** the tool runs from wherever the user is, not from the root
**Input:** `git init` at `root`, query from `root/a/b`
**Expected:** `IsRepo()` is true
**Refs:** crc-Git.md — R167

## Test: ignore state is reported per path in one invocation
**Purpose:** validates the batching that makes checking on every run affordable
**Input:** a repository whose `.gitignore` lists two of four queried paths
**Expected:** exactly those two report ignored, in correspondence with the paths asked
about — not merely the right *count*
**Refs:** seq-bootstrap.md#1.9 — R148

## Test: a path ignored by a nested `.gitignore` is reported ignored
**Purpose:** the reason to ask git rather than parse `.gitignore` ourselves — ignore
rules compose in ways a naive reader gets wrong
**Input:** `root/.gitignore` empty, `root/sub/.gitignore` ignoring `x.md`; query
`root/sub/x.md`
**Expected:** reported ignored
**Refs:** crc-Git.md — R167

## Test: a negated pattern is reported not ignored
**Purpose:** the sentry case — `!` reverses an earlier rule, and a hand-rolled matcher
is exactly where that is missed
**Input:** `.gitignore` containing `*.md` then `!KEEP.md`; query both `x.md` and
`KEEP.md`
**Expected:** `x.md` ignored, `KEEP.md` not
**Refs:** crc-Git.md — R167

## Test: a tracked file reports as tracked, an untracked one does not
**Purpose:** the fact behind the `.minispec/config.yaml` preference
**Input:** a repository with one added file and one untracked file
**Expected:** `IsTracked` distinguishes them
**Refs:** crc-Git.md — R164

## Test: a non-git tree reports ignore state as unavailable
**Purpose:** validates that the narrowing is stated rather than passed silently — the
failure mode this project is most alert to, where a check that could not look still
returns a clean answer
**Input:** a temporary directory with no repository; ask for ignore state
**Expected:** `Unavailable()` is true and no path is reported as "not ignored", which a
caller could mistake for a verified answer
**Refs:** crc-Git.md — R168

## Test: no git operation mutates the repository
**Purpose:** codifies the boundary — the tool reads git and never drives it
**Input:** a repository with a known `git status --porcelain` and `HEAD`; run every
`Git` method
**Expected:** status and `HEAD` are byte-identical afterwards
**Refs:** crc-Git.md — R166
