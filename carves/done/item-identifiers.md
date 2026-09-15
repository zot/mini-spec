# Carve: item identifiers — the item number names the change, not the commit hash

Three items landed on 2026-09-15 and took six commits: one per item, and one tail per item
carrying the carve flip and the post-commit alarm records, because the `LANDED` line names
the item's commit hash and a hash exists only after the commit. Bill's reading, the same
day: *this is an excessive number of commits; carve items can share commits; we can use a
different identifier than a hash.* This carve is that change.

**Provenance.** Bill, 2026-09-15, in conversation after `#83`–`#85`. Supersedes the 2026-09-04
decisions *one squashed commit per item* and *the `LANDED` line names the commit*, both
recorded in `carves/done/sdom-reclaim.md` and the skill; they stay there as history.

## Status

- [x] ~~**Item 1 — the format and the reader: `LANDED (<date> — #N.)`, no hash.**~~ **LANDED (2026-09-15 — `#86`.)**
- [x] ~~**Item 2 — `pending finish` without `--commit`, and the commit message the tool composes.**~~ **LANDED (2026-09-15 — `#87`.)** Needs Item 1.
- [x] ~~**Item 3 — the skill and the Cursor rule: one commit per batch, amend for the census.**~~ **LANDED (2026-09-15 — `#88`.)** Needs Item 2.

## Decisions

**DECIDED (Bill, 2026-09-15): the item number is the identifier.** A part's `LANDED` line
carries the date and the item number and no commit hash — the shape the `Pulled` line took on
2026-09-04 for the same reason: a hash names a commit that an amend or a squash rewrites away,
and it exists only after the commit, which is what forced the flip into a second commit. The
path from a part to its change is `git log --grep '#N'`, which works for a cloner as long as
the commit names its items. The done ledger, being private and never squashed away, may still
record the hash after the fact; nothing public does.

**DECIDED (Bill, 2026-09-15): carve items share commits.** One commit per batch, not per item.
`pending finish` no longer needs a commit to exist, so it runs *before* the commit, the flip
lands in its own item's commit, and any number of finished items close in one.

**DECIDED (Bill, 2026-09-15): the commit message names every item it lands, and the message
is the tool's to compose.** A commit that forgets its item numbers breaks the pointer silently,
so the subject is not an agent's to remember: `pending finish` prints the line, or a verb
composes the message from the items finished since the last commit — Item 2 decides which.

**DECIDED (Bill, 2026-09-15): the post-commit census re-pulls are amended into the commit they
belong to, and an amend *appends* — the existing message is preserved and the new information
follows it.** The census reads code as committed at `HEAD`, so alarms on new or changed code
can only be checked after the commit; the records go back into that commit rather than into a
tail. Two rules ride with it: amend only while the commit is unpushed (`git branch -r
--contains HEAD` is empty), a follow-up commit otherwise; and the amended message is the old
message with the new lines after it, never a rewrite — the previous message is part of the
record.

## The parts

**Item 1** — the format and the reader. `trajectory-format.md` and `part-line.md` state the
`LANDED` shape as `**LANDED (<date> — `#N`.)**`; the carve reader writes that and accepts the
old hash-bearing form on read, since three done carves are written with it and history is not
rewritten. The done entry's identifier slot follows: the hash becomes optional there too.

**Item 2** — the verbs. `pending finish <N>` drops `--commit`; the done entry is written
without a hash and the ledger may gain one later by hand. The message: decide between
`finish` printing the subject to include and a `pending commit-message` verb that composes
the whole message — subject naming the items, body from the done entries — for
`git commit -F`. The verb also serves the amend: `--amend` appends the new items' lines after
the existing message, honouring the append rule, and refuses when `HEAD` is on a remote.

**Item 3** — the skill. *One commit per item, squash, never fixup* becomes *one commit per
batch, naming its items; amend for the census while unpushed, appending*. Gap `O12`, which
banked the tool's half of the squash rule, closes or is rewritten. The Cursor rule regenerates.

## Open questions

1. Does `git log --grep '#N'` need the `#` anchored — `#85` also matches `#850`? A subject
   convention such as `items: #83, #84` with word boundaries, or the tool's own `--grep`
   composition, answers it; decide in Item 2.
