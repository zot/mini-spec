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

## Test: LastChanged reads any object format
**Purpose:** validates R180 against a real repository in both `sha1` and `sha256`. The
first version gated on a 40-character hash, so in a SHA-256 repository no format line
matched, the scan fell through to "no changes", and **every alarm reported verified** —
a check that could not look returning a clean result, which is the precise failure the
freshness feature exists to prevent, inside its own implementation
**Input:** a repository per object format, one file, one function, one commit
**Expected:** a real date and no error from both
**Fire alarm:** restore the `len(hash) != 40` gate and confirm the `sha256` case goes
red reporting no change for a function just committed
**Inject:** internal/project/git.go:LastChanged
**Pulled:** 2026-09-04 — rang again after the site changed: `sha256: LastChanged
reported no change for a function that was just committed`; restore byte-clean by copy.
First pulled 2026-08-13, same signature
**Refs:** crc-Git.md — R180

## Test: a new function is told from a gone one
**Purpose:** validates R182 and R184 — `git log -L` searches the file as committed, so
a function added since the last commit is absent from history while present on disk.
Reporting that as a rotted anchor is false and fires on every newly written function
**Input:** a repository with one committed function and one added but uncommitted; and
a symbol in neither
**Expected:** `ErrNoHistory` for the new one, `ErrUnresolvedSite` for the absent one
**Fire alarm:** treat every `-L` failure as a rotted anchor — the pre-fix shape — and
confirm the new function reports `ErrUnresolvedSite`
**Inject:** internal/project/git.go:LastChanged
**Pulled:** 2026-09-04 — rang again after the site changed: `a newly written
function = git cannot resolve that symbol in that file, want ErrNoHistory`; restore
byte-clean by copy. First pulled 2026-08-13, same signature
**Refs:** crc-Git.md — R182, R184

## Test: a method anchor resolves to its declaration
**Purpose:** validates R205 — `Type.Method` handed to git literally matches no line of Go,
so every method-form anchor read *unresolvable* on 2026-09-04 (69 in a sibling project);
handed as a declaration-shaped pattern it resolves, to the method and not to a
same-named method on another type
**Input:** a repository whose file declares `func (b B) Run()` first, followed by another
function so its `-L` range never grows, committed on day one; `func (a *A) Run()` appended
and committed on day two. Chosen so the wrong resolution has a different date: git's `-L`
range includes the blank line after a declaration, so a method with nothing after it is
reported changed whenever something is appended — measured 2026-09-04 while writing this
**Expected:** `LastChanged("x.go", "A.Run")` reports day two and `("x.go", "B.Run")` day one;
a pattern resolving to the first `Run` in the file reports day one for both
**Fire alarm:** hand git the symbol as written — return it unchanged from `sitePattern` —
and confirm both method cases go red. The error is `ErrNoHistory`, not `ErrUnresolvedSite`:
the on-disk check still recognises the method, so the fallback reads "new, not gone"
**Inject:** internal/project/git.go:sitePattern
**Pulled:** 2026-09-04 — rang: `A.Run: git holds no history for that path`, and the bounded
test went red beside it since the same return dropped its boundaries; restore byte-clean by
copy. First written predicting `ErrUnresolvedSite`; corrected to what was observed
**Refs:** crc-Git.md — R205
**Code:** internal/project/git_test.go

## Test: a bare anchor is bounded, not a substring
**Purpose:** validates R206 — git takes the first line its pattern matches, so an unbounded
`Lookup` resolves to `LookupPath` and the alarm watches the wrong function with a clean
reading
**Input:** a repository whose one committed file declares `func LookupPath()` and nothing
named `Lookup`
**Expected:** `LastChanged("x.go", "Lookup")` returns `ErrUnresolvedSite`; `"LookupPath"`
returns a date
**Fire alarm:** drop the `\b` boundaries from the bare-symbol pattern and confirm `Lookup`
goes green with `LookupPath`'s date
**Inject:** internal/project/git.go:sitePattern
**Pulled:** 2026-09-04 — rang: `Lookup resolved (<nil>); want ErrUnresolvedSite`; the method
test stayed green, so the two alarms discriminate; restore byte-clean by copy
**Refs:** crc-Git.md — R206
**Code:** internal/project/git_test.go
