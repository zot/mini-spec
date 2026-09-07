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
**Alarm:** 1
**Fire alarm:** restore the `len(hash) != 40` gate and confirm the `sha256` case goes
red reporting no change for a function just committed
**Inject:** internal/project/git.go:LastChanged
**Pulled:** 2026-09-06 — rang again after `LastChanged` moved to the computed extent (delegated `alarm-puller` in a worktree at `3931bcc`, evidence read by hand): `sha256: LastChanged reported no change for a function that was just committed`; restore byte-clean. Previously 2026-09-04 and 2026-08-13, same signature
**Refs:** crc-Git.md — R180

## Test: a new function is told from a gone one
**Purpose:** validates R182 and R184 — `git log -L` searches the file as committed, so
a function added since the last commit is absent from history while present on disk.
Reporting that as a rotted anchor is false and fires on every newly written function
**Input:** a repository with one committed function and one added but uncommitted; and
a symbol in neither
**Expected:** `ErrNoHistory` for the new one, `ErrUnresolvedSite` for the absent one
**Alarm:** 2
**Fire alarm:** treat every `-L` failure as a rotted anchor — the pre-fix shape — and
confirm the new function reports `ErrUnresolvedSite`
**Inject:** internal/project/git.go:LastChanged
**Pulled:** 2026-09-06 — rang again after `LastChanged` moved to the computed extent (delegated `alarm-puller` in a worktree at `3931bcc`, evidence read by hand): the `-L` failure is now the parse finding nothing, so the injection dropped the on-disk `siteExtent` check: `a newly written function = git cannot resolve that symbol in that file, want ErrNoHistory`; restore byte-clean. Previously 2026-09-04 and 2026-08-13, same signature
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
**Alarm:** 3
**Fire alarm:** ignore the receiver in `siteExtent` — compare `receiverType` against a value no receiver can be, `typ+"!"` (dropping the condition or replacing it with `false` leaves a variable unused and does not compile, which is not a pull) —
and confirm `A.Run` goes red: it resolves to the first `Run`, B's, and reports day one.
*Until 2026-09-06 the site was `sitePattern`, the stopgap the extent replaced*
**Inject:** internal/project/extent.go:siteExtent
**Pulled:** 2026-09-06 — rang, and not where predicted: with the receiver ignored both `Run`s answer to `A.Run`, so it read `A.Run is declared 2 times in that file; name the receiver` — the ambiguity check caught it before the date comparison could; restore byte-clean by copy, and rang again the same day after the simplification pass restructured `siteExtent` and `groupEnd`. Previously 2026-09-04 against `sitePattern`, signature `A.Run: git holds no history for that path`
**Refs:** crc-Git.md — R304
**Code:** internal/project/git_test.go

## Test: a bare anchor is bounded, not a substring
**Purpose:** validates R206 — git takes the first line its pattern matches, so an unbounded
`Lookup` resolves to `LookupPath` and the alarm watches the wrong function with a clean
reading
**Input:** a repository whose one committed file declares `func LookupPath()` and nothing
named `Lookup`
**Expected:** `LastChanged("x.go", "Lookup")` returns `ErrUnresolvedSite`; `"LookupPath"`
returns a date
**Alarm:** 4
**Fire alarm:** match the name as a prefix in `siteExtent` — `strings.HasPrefix(text, want)`
in place of equality — and confirm `Lookup` goes green with `LookupPath`'s date. *Until
2026-09-06 the site was `sitePattern` and the injection dropped its `\b` boundaries*
**Inject:** internal/project/extent.go:siteExtent
**Pulled:** 2026-09-06 — rang: `Lookup resolved (<nil>); want ErrUnresolvedSite — an unbounded name matched inside LookupPath`; restore byte-clean by copy, and rang again the same day after the simplification pass restructured `siteExtent` and `groupEnd`. Previously 2026-09-04 against `sitePattern`, same signature
**Refs:** crc-Git.md — R303

## Test: a comment-only edit does not stale the declaration
**Purpose:** validates R303 and R306 — the range is the declaration's own lines and nobody's comment. Git's range gave a declaration its successor's doc block, and old-sdom's first repair gave it its own; the second turned three verified alarms stale over traceability lines rewritten inside doc blocks
**Input:** a repository committing `Foo` with a doc comment and `Bar` after it on day one; on day two only the two comments change
**Expected:** `LastChanged("x.go", "Foo")` reports day one
**Alarm:** 5
**Fire alarm:** start the extent one line above the keyword — `d.Line(from) - 1`, which takes the doc comment in — and confirm `Foo` goes red reporting day two
**Inject:** internal/project/extent.go:siteExtent
**Pulled:** 2026-09-06 — rang: `Foo changed 2026-01-02 after a comment-only edit; want 2026-01-01`; restore byte-clean by copy, and again the same day after the simplification pass restructured `siteExtent` and `groupEnd`, same signature
**Refs:** crc-Git.md — R303, R306
**Code:** internal/project/git_test.go

## Test: an append after the last declaration does not stale it
**Purpose:** validates R305 — the range ends where the groups close, not at the line before the next declaration; git's range took the trailing blank line and staled the last function in a file on every append (gap O11)
**Input:** a repository committing `A.Run` as the last declaration on day one, and appending `B` below it on day two
**Expected:** `LastChanged("x.go", "A.Run")` reports day one
**Alarm:** 6
**Fire alarm:** extend the extent by one line past the close in `groupEnd` and confirm `A.Run` goes red reporting day two
**Inject:** internal/project/extent.go:groupEnd
**Pulled:** 2026-09-06 — rang: `A.Run changed 2026-01-02 after an append below it; want 2026-01-01`; restore byte-clean by copy, and again the same day after the simplification pass restructured `siteExtent` and `groupEnd`, same signature
**Refs:** crc-Git.md — R305
**Code:** internal/project/git_test.go

## Test: an ambiguous bare name is refused, and the receiver form resolves
**Purpose:** validates R307 — two declarations answering to one name is reported with the count, never resolved to the first
**Input:** a repository committing `A.Run` and `B.Run`
**Expected:** `LastChanged("x.go", "Run")` returns an `AmbiguousSiteError` with `N` 2 matching `ErrAmbiguousSite`; `"A.Run"` returns a date
**Alarm:** 7
**Fire alarm:** stop counting past the first match in `siteExtent` — `break` after the first hit — and confirm `Run` goes green with a date
**Inject:** internal/project/extent.go:siteExtent
**Pulled:** 2026-09-06 — rang: `Run = <nil>; want AmbiguousSiteError{N: 2}`; restore byte-clean by copy, and again the same day after the simplification pass restructured `siteExtent` and `groupEnd`, same signature
**Refs:** crc-Git.md — R307
**Code:** internal/project/git_test.go

## Test: the extent over the shapes a repository test does not reach
**Purpose:** validates R303, R304, R305 on the parse alone — a signature spanning lines, a nested group, grouped `var` members at their own lines, a receiver carrying a type parameter, the bare form of a method, a receiver that does not match
**Input:** one twenty-line Go source
**Expected:** `Multi` is lines 9–16, `A` and `B` are 4 and 5, `Set.Add` and `Add` are 18, `One` is 20, `Other.Add` and `Nope` are absent
**Alarm:** 8
**Fire alarm:** stop the walk in `groupEnd` at the first closer instead of the last — confirm `Multi` reads 9–11, the parameter group's close
**Inject:** internal/project/extent.go:groupEnd
**Pulled:** 2026-09-06 — rang: `Multi: got 9-11 count 1, want 9-16 count 1`; restore byte-clean by copy, and again the same day after the simplification pass restructured `siteExtent` and `groupEnd`, same signature
**Refs:** crc-Git.md — R303, R304, R305
**Code:** internal/project/git_test.go
