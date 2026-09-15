# CommitMessage
**Requirements:** R479, R480, R481, R482, R483, R484

Composes the commit message for the items finished since the last commit, so every commit
names the items it lands. Writes the message; never stages or commits.

## Knows
- the repository root and the done file's entries with their lines
- which `#N` the commit messages on `HEAD`'s history name, from git
- the uncommitted entries, in the order they finished

## Does
- `Compose(root, amend)`: reads the done file, asks git which items its messages name, keeps
  every entry with an identifier no message names, and refuses when none is left
- writes the subject `#N, #M: title; title` and the body `Items #N, #M.` followed by each
  entry's `#N — title` and its done-file body
- with `amend`: refuses when `HEAD` is on a remote branch; otherwise returns `HEAD`'s message
  unchanged with `Also lands #N.` and the entries appended after it

## Constraints
- **Git says what is committed**, not the done file: it is private and never in a commit
- **`#N` is bounded**: `#40` does not name `#4`
- **An amend appends and never rewrites** a byte of the existing message
- **No sign-off, no commit**: the message is the tool's, the trailer and the act are the committer's

## Collaborators
- minispecsdom.Done: the entries and their lines; the body is the file's lines between them
- Git: `NamedItems`, `HeadMessage`, `HeadOnRemote`
- CLI: `pending commit-message [--amend] [--out <file>]`

## Sequences
- seq-queue-item.md
