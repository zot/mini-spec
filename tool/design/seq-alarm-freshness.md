# Sequence: Fire alarm freshness — assessing a recorded proof

**Requirements:** R178, R179, R180, R181, R182, R183, R184, R185, R187

Two diagrams: assessing one alarm, and the two surfaces that consume the assessment.

## 1. Assessing one alarm

```
1. Assess an alarm against git
1.1. Parse the test document's alarm blocks: Inject sites, Pulled date, prose
1.2. No Inject sites — unanchored; stop, since nothing can be checked
1.3. No Pulled date — unrecorded; stop, and say only that no verification is recorded
1.4. Ask Git whether this is a working tree; if not, unchecked; stop
1.5. For each site, ask Git when that function last changed
1.6. A site git cannot resolve — unresolvable; name the site and stop
1.7. Any change dated strictly after the pull date — stale; name the site and the date
1.8. Otherwise verified
```

**Step 1.2 and 1.3 need no git at all**, which is what keeps the cost proportional to
the *verified* population rather than to every alarm. A project adopting the convention
has many unrecorded alarms and few pulled ones, so the expensive path is the small one.

**Step 1.5 asks about the function, not the file (R180).** The file-level question
marks every alarm in a busy file stale and so discriminates nothing. Measured before
this existed: file granularity called 8 of 13 alarms stale, function granularity called
3, and hand-checking those left 1.

**Step 1.6 reports rather than skipping (R182).** Git states plainly whether it found
the symbol, so a site it cannot find is the anchor rotting — the precise failure the
field exists to prevent. Skipping it would make the tool silent about its own mechanism
breaking.

**Step 1.7's boundary is exclusive, and measurement is why (R181).** It was inclusive
first, on the argument that over-reporting is safe. But the normal workflow commits the
fix, the test and the pull together, so a freshly recorded alarm always shares a date
with the code it proves — the inclusive rule marked two of ark's three verified alarms
stale the day they were written. A check that fires on arrival is ignored, which costs
more than the blind spot exclusivity leaves.

**Step 1.4 is why there is a state rather than a boolean.** A tree with no git cannot
answer, and the answer to "could not look" must never be the same as "looked and found
nothing" — the failure mode this whole feature is a response to.

## 2. The two surfaces

```
2. Report the assessments
2.1. validate: report only the stale alarms, naming the site and the date that voided each
2.2. validate: report unresolvable sites alongside them, being anchors that no longer resolve
2.3. validate: say nothing about unrecorded or unanchored alarms
2.4. query alarms: list every alarm with its state, grouped by document
2.5. query alarms: close with the census of the four counts
```

**Step 2.3 is the load-bearing omission.** Stale is closable and should normally read
zero, so a non-zero reading is a specific worklist. Unrecorded and unanchored are
neither: a project adopting the convention carries many, the counts fall slowly, and a
line reporting a non-zero number every run for months is the recurring nag this project
distinguishes from a gripe you can discharge. They are worth knowing on demand, which
is what step 2.4 is for.
