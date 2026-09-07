# Git
**Requirements:** R151, R164, R165, R166, R167, R168, R180, R184, R182, R236, R237, R238, R303, R304, R305, R306, R307, R308, R309, R315

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
  busy file stale and so discriminates nothing. Five outcomes stay apart — no work tree,
  no history for the path, a tracked file whose symbol the parse cannot find, a symbol
  declared more than once, and a function that exists and has never changed. Only the
  last is a clean result (R180, R182, R184, R307)
- SiteResolves(file, symbol): whether a site names exactly one declaration in HEAD's copy of
  the file, with LastChanged's failures kept apart and none of its history walk — what the
  census asks of a prescription (R309)
- siteExtent(src, symbol): the site's 1-based line range and how many declarations
  answered to the name, from the dependency's `sdom.LangGo` parse and `schema.Go` pass —
  the declaring line through the line its groups close on, no comment; `Type.Method`
  matched on the receiver group's type (R303, R304, R305, R306)
- headFile(file): the file as committed at HEAD, `git show HEAD:./file`, cached per file
  for one invocation; the range is computed over these bytes and never the working
  tree's, because `-L <start>,<end>` resolves against HEAD (R308)
- Snapshot(): the **worktree anchor** — a tree written from a scratch index (`GIT_INDEX_FILE`),
  committed with HEAD as first parent when there is one, and pointed at by `refs/minispec/snapshot`
  in one `update-ref`. Untracked contents in, ignored paths out, the real index and working tree
  untouched — three properties that are structural rather than rules to keep. Reference, never
  undo. `ErrNoGit` outside a repository, which the slot treats as "skip", not "fail" (R236,
  R237, R238)

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
- simple-dom `sdom` and `sdom/schema`: the Go parse and declaration pass the extent is computed over
- Track: the caller that turns these facts into a verdict
- Init: asks whether git is present before writing ignore lines

## Sequences
- seq-bootstrap.md
