# Test Design: CommitMessage
**Source:** crc-CommitMessage.md

## Test: the uncommitted items are the ones no message names, bounded
**Purpose:** R479, R480, R481
**Input:** the fixture with `#4` finished; then a commit whose message names `#40`; then an old-scheme entry with a hash in its slot; then a commit naming `#4`
**Expected:** first, a message with subject `#4: a part to queue`, body opening `Items #4.` and carrying the entry's body; after the `#40` commit the same, because `#40` does not name `#4`; the hash-bearing entry is never composed; after the `#4` commit a refusal naming `#4` and that commit, and an older entry with neither hash nor naming commit below `#4` stays history; the `#4` commit a refusal naming `#4` and that commit
**Refs:** crc-CommitMessage.md, seq-queue-item.md#4.2
**Code:** internal/pending/message_test.go
**Alarm:** 1
**Fire alarm:** match `#N` as a prefix rather than bounded. Red: after the `#40` commit the verb refuses, reading `#4` as named. A second injection at `uncommitted`: judge each entry alone instead of stopping at the first named one. Red: the older unnamed entry is composed.
**Inject:** internal/project/git.go:Git.NamedItems, internal/pending/message.go:uncommitted
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, twice: with `#N` matched on one digit, `#40 was read as naming #4` and the verb refused; with `uncommitted` judging each entry alone (`continue` for `break`), `an entry older than a named one was composed`; restore checksummed clean both times

## Test: amend appends after the existing message and refuses on a remote
**Purpose:** R482, R483
**Input:** the fixture with `#4` finished and `HEAD` an ordinary commit; then `HEAD` pushed to a bare remote
**Expected:** the amended text begins with `HEAD`'s message byte for byte, then a blank line, `Also lands #4.`, the entry; after the push, a refusal naming the remote branch
**Refs:** crc-CommitMessage.md, seq-queue-item.md#4.5
**Code:** internal/pending/message_test.go
**Alarm:** 2
**Fire alarm:** compose the amend as a fresh message with the old one appended after it. Red: the text no longer begins with `HEAD`'s message.
**Inject:** internal/pending/message.go:Compose
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the amend composed fresh and the old message appended after: the text began with `Also lands #4.` and the existing subject came last; restore checksummed clean

## Test: the CLI writes to stdout or the file, and never commits
**Purpose:** R484
**Input:** the fixture, `pending commit-message --out <file>`, run from a subdirectory
**Expected:** exit 0, the file holds the message, `git log` has the same single commit as before, and the backup slot holds no stamp afterwards — the stamp's presence, since the recorded state is the zero value and a state comparison alone cannot see a fresh record
**Refs:** crc-CommitMessage.md, seq-queue-item.md#4.6
**Code:** internal/cli/cli_message_test.go
**Alarm:** 3
**Fire alarm:** route the verb through the backup-slot crank so it records a transition. Red: the slot's state changes, and `pending changes` reports a transition that never happened.
**Inject:** internal/cli/pending.go:CLI.runPending
**Pulled:** 2026-09-15 — rang, by hand after the simplification pass, with the verb recording a no-op into the slot: `the verb entered the backup slot: changed/false → changed/true`. *First attempt did not distinguish:* the test compared states alone, and `Changed` is the zero value, so a fresh stamp read equal to no stamp — the alarm found the assertion blind and the test now compares the stamp's presence too; restore checksummed clean
