# Git
**Requirements:** R151, R164, R165, R166, R167, R168, R180, R184, R182, R205, R206

Answers questions about the version-controlled working tree by invoking the `git`
command line. Every question it answers is a **computed property** — there is
nothing here to store and therefore nothing to go stale.

Deliberately the only place in the tool that knows a version-control system exists.
Isolating it is what makes `Track`'s consistency check testable without a real
repository: a fake `Git` states the world the check is being run against.

## Knows
- workDir: the repository root every query is asked relative to
- available: whether a usable `git` and a work tree were found — resolved once, since
  a tree either is git-managed or is not for the whole of one invocation

## Does
- IsRepo(): report whether workDir is inside a git working tree. False is a legitimate
  answer, not an error — a project with no version control is a supported shape
- Ignored(paths): report, per path, whether git would ignore it. Takes a list because
  the whole question costs one invocation, and `Track` always asks about several paths
  at once. **Returns an error where no ignore state can be determined** — a tree
  managed by another version-control system reaches this, and is told so
- Tracked(path): report whether a path is known to the index — the question behind the
  `.minispec/config.yaml` preference
- LastChanged(file, symbol): report when a named **function** last changed. Asked of
  the function rather than the file because a file-level answer marks every alarm in a
  busy file stale and so discriminates nothing. Four outcomes stay apart — no work tree,
  no history for the path, a tracked file whose symbol cannot be found, and a function
  that exists and has never changed. Only the last is a clean result (R180, R182, R184)
- sitePattern(symbol): the pattern git is handed for `-L`. `Type.Method` becomes a
  declaration-shaped pattern over the receiver, name and star optional; a bare symbol is
  bounded on both sides. Git's `-L` regex is POSIX basic, so the escapes are git's, not
  Go's, and the tests run against a real repository for that reason (R205, R206)

**Unavailability is an error, not a boolean.** A "cannot determine" flag beside a map
of falses invites a caller to read the map first and the flag never; an error cannot be
mistaken for "none of these are ignored". This is the project's standing rule — report
absence as error, never as silence — applied at the one seam where the tool asks a
question it may be unable to ask.

**What it deliberately cannot do.** No staging, no commits, no resets, no branch
state. The tool reads git and never drives it. Editing `.gitignore` is not excluded
by that rule and does not live here anyway — it is an ordinary file edit, owned by
`Init`.

**Why the CLI rather than a library.** Supporting a second version-control system, or
linking one in, means knowing how to *operate* it — a large surface for checks this
small. Shelling out gets tracked and ignored status for free and stays correct as git
changes. The cost is stated rather than hidden: a non-git project gets no ignore
checking and is told so.

## Collaborators
- os/exec: to invoke `git`
- Track: the caller that turns these facts into a verdict
- Init: asks whether git is present before writing ignore lines

## Sequences
- seq-bootstrap.md
