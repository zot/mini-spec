# Alarm
**Requirements:** R179, R180, R181, R182, R183, R185, R186, R187, R200, R201, R202, R203

Assesses recorded fault injections. Given the alarms a test design carries and a Git
that can answer "has this function changed", it says which proofs are still good.

A fire alarm is *evidence* — a claim that a specific test detects a specific defect,
obtained on a specific day. Like every claim about code, it expires when its subject
changes, and nothing about a passing suite says so. This card is what makes the
expiry a computed property instead of something a person has to remember.

## Knows
- Alarm: one recorded injection — its document, its test, the `file:symbol` sites it
  edits, the date it was pulled if any, and the prose describing it
- State: `verified`, `stale`, `unrecorded`, `unanchored`, `unresolvable`, `unchecked`

## Does
- Assess(alarms, git): resolve each alarm to a State, and nothing else. Pure over its
  two inputs, so the whole table is testable against a fake Git with no repository
- staleness(alarm, git): for an alarm with sites and a pull date, ask Git whether any
  named symbol changed **on or after** that date (R179, R181)
- census(assessments): the four counts `query alarms` closes with (R185)
- Brief(assessment, designRoot, testFiles): the spawn prompt for one delegated re-pull.
  Pure over its three inputs, so the whole shape is testable with no repository and no
  agent (R200, R201, R202, R203)

**Why the brief is complete about its alarm and silent about its protocol (R202).** The
variant half — this document, this test, these sites, this prose — changes per alarm and
so has to be minted per alarm. The invariant half — baseline first, inject, diff, re-run,
restore, prove the restore clean — is the same every time, so it lives in the agent
definition that runs it, which is what an agent definition is for. A brief carrying the
protocol would repeat some 350 words per alarm against a census whose entire purpose is
to be small.

**The one invariant a brief does restate is the contract, and it is the load-bearing
half: evidence, never a verdict (R202).** What comes back is the command, the output
before the injection, the diff applied, the output after, and the diff after restoring.
Whether the alarm rang is read off that by a reader who can weigh it. Measured rather
than principled: three times inside a single item the naive verdict would have been
wrong — an injection at a correctly named site that did not ring because the rule had two
guards, an injection that rang across three packages while being *incapable* of reaching
the property it named, and an `**Inject:**` field naming a symbol the injection only
consulted.

**Why files and directories, never a command (R201).** This tool knows document
structure; it does not know build systems. A guessed command that fails to build produces
output a hurried reader scores as *rang* — the failure `SKILL.md` names as *an injection
that breaks the build teaches nothing, because the test never ran*. The protocol closes it
from the other side, by demanding a green baseline before anything is injected, so a wrong
command is caught while the tree is still clean.

**Why an absent part is stated rather than dropped (R203).** An `unanchored` alarm has no
sites to edit and a document with no Artifacts row has no test files to name. A missing
*Tests:* line reads as an oversight in the generator; a line saying the manifest maps none
reads as what it is. It also keeps the brief count equal to the census count, which is the
one number a reader trusts — the same reason `Assess` reports an unresolvable site instead
of skipping it.

**Why the function and not the file (R180).** A file-level answer marks every alarm in
a busy file stale, so it discriminates nothing and would be discarded as noise within a
week. Git can answer the narrower question directly, and the narrower question is the
one the claim was ever about. Measured before this existed: file granularity called 8
of 13 alarms stale where function granularity called 3, and hand-checking those 3 left
1 — the coarse instrument was wrong in both directions at once.

**Why the comparison is strictly after (R181).** The first version counted same-day
changes as stale, reasoning that over-reporting is the safe direction. Measurement
refuted it: the normal workflow commits the fix, the test and the pull together, so the
code's last change and the pull date coincide for *every freshly recorded alarm* — the
inclusive rule marked two of ark's three verified alarms stale the day they were
written. A check that fires on arrival is ignored, and an ignored check is worth less
than the blind spot strictness leaves: a change made later the same day, caught by the
next change on any later day. Recording a commit hash beside the date would close that
gap exactly, and is the obvious refinement if the blind spot ever bites.

**Why an unresolvable site is reported, never skipped (R182).** Git says exactly
whether it could find the symbol, and a site it cannot find is the anchor rotting —
the failure `**Inject:**` exists to prevent, so silence there would be the tool
declining to report its own mechanism failing. Same reason a check with no git returns
`unchecked` rather than a clean result (R187).

**Why `unrecorded` is worded as it is (R186).** It says the repository does not record
a verification. It does *not* say the injection was never run, because the documents
cannot answer that, and a tool asserting the stronger claim would be manufacturing a
finding. The same distinction a carve draws when it gives `NOT VERIFIED` its own words.

## Collaborators
- Git: answers whether a named function has changed since a date
- Parser: supplies the parsed alarms; this card does no file reading
- Query: resolves the design root and the document's test files, and hands both in;
  this card reads no manifest

## Sequences
- seq-alarm-freshness.md
