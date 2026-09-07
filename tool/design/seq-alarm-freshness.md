# Sequence: Fire alarm freshness — assessing a recorded proof

**Requirements:** R178, R179, R180, R181, R182, R183, R184, R185, R187, R199, R200, R202, R203, R204, R303, R305, R307, R308, R309

Three diagrams: assessing one alarm, the two surfaces that consume the assessment, and
minting the brief a delegated re-pull is spawned with.

## 1. Assessing one alarm

```
1. Assess an alarm against git
1.1. Parse the test document's alarm blocks: Inject sites, Pulled date, prose
1.2. No Inject sites — unanchored; stop, since nothing can be checked
1.3. No Pulled date — resolve each site first, at the cheap half of the cost; a site naming nothing or two things is unresolvable; otherwise unrecorded, and say only that no verification is recorded
1.4. Ask Git whether this is a working tree; if not, unchecked; stop
1.5. For each site, ask Git when that function last changed
1.5.1. Read the file as committed at HEAD, once per file
1.5.2. Compute the declaration's extent from the Go parse: declaring line through the line its groups close on
1.5.3. Two declarations answer to the name — ambiguous; report it as unresolvable with the count
1.5.4. git log -L start,end over that range; the first format line is the answer
1.6. A site git cannot resolve — unresolvable; name the site and stop
1.7. Any change dated strictly after the pull date — stale; name the site and the date
1.8. Otherwise verified
```

**Step 1.2 needs no git at all, and step 1.3 needs only the cheap half (R309).** The
history walk stays proportional to the *verified* population: a project adopting the
convention has many unrecorded alarms and few pulled ones. But a prescription's anchor is
resolved — one parse of HEAD's file, cached per file — because the site nobody has run is
the one most likely to have rotted, and until 2026-09-06 it was the only shape never
looked at. Without git, or with no history yet, the reading stays `unrecorded`.

**Step 1.5 asks about the function, not the file (R180).** The file-level question
marks every alarm in a busy file stale and so discriminates nothing. Measured before
this existed: file granularity called 8 of 13 alarms stale, function granularity called
3, and hand-checking those left 1.

**Steps 1.5.1–1.5.4 compute the range before asking git (R303, R305, R308).** Handing git a
pattern asked it to find the symbol as well as bound it, and git bounds a declaration at the
line before the next one: the successor's doc comment and the trailing blank line were in
every range, so a comment edit or an append staled the neighbour. The range is computed over
HEAD's bytes because that is the coordinate system `-L <start>,<end>` walks. Step 1.5.3 is
counted before the extent is trusted, because the extent cannot report it (R307).

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
2.6. query alarms --unverified: drop the verified entries from the list, and from the list only
2.7. query alarms --brief: replace each surviving entry's line with its brief — diagram 3
```

**Step 2.3 is the load-bearing omission.** Stale is closable and should normally read
zero, so a non-zero reading is a specific worklist. Unrecorded and unanchored are
neither: a project adopting the convention carries many, the counts fall slowly, and a
line reporting a non-zero number every run for months is the recurring nag this project
distinguishes from a gripe you can discharge. They are worth knowing on demand, which
is what step 2.4 is for.

**Steps 2.6 and 2.7 are applied before the output form is chosen**, not after. Choosing
text or JSON first and filtering the text branch renders everything under `--json` while
every markdown test stays green — the silent ignore `query gaps` already met one level in,
and the reason its selection sits where this one does (R199, R204).

**Step 2.6 narrows the list and never the census.** The closing count is computed over the
whole population either way. A filtered count answers *how many are wrong* and silently
drops *out of how many*, which is the standing `| tail -2` workaround wearing the whole
question's clothes; keeping it whole means `--unverified` loses nothing at all.

## 3. Minting one brief

```
3. Mint the brief a delegated re-pull is spawned with
3.1. Express the design root as a path relative to the repository root
3.2. Look the alarm's document up in design.md's Artifacts manifest; take its test files
3.3. Reduce those files to the distinct directories holding them, in first-seen order
3.4. Head the brief with the document, the test title and the assessed state
3.5. Name the sites to edit, or say there are none and that nothing can be injected
3.6. Name the test files and their directories, or say the manifest maps the document to none
3.7. Quote the **Fire alarm:** prose verbatim
3.8. Close with the evidence contract, and with no protocol
```

**Step 3.1 is not cosmetic.** A delegated puller runs in its own git worktree, where the
absolute path of this checkout resolves to nothing. Naming the design root relative to the
repository root is the one form that is true in every checkout of this repository (R200).

**Steps 3.5 and 3.6 state an absence rather than dropping a line (R203).** A missing
*Sites:* line reads as an oversight in the generator; a line saying the alarm is
`unanchored` reads as what it is, and tells the reader the next move is to write the
anchor rather than to spawn anyone. It also keeps the number of briefs equal to the number
of selected alarms, so the two can be compared.

**Step 3.7 quotes rather than summarises.** The prose is the injection, and a paraphrase of
an injection is a different injection. The parser already folds an alarm's continuation
lines into one body for exactly this reason (R178), so what is quoted here is the whole of
what the document said.

**Step 3.8 is the boundary, and it is where the value is.** The protocol belongs to the
agent definition that runs it — invariant, so minting it per alarm would be repetition
rather than information. The contract does not: *evidence, never a verdict* is the single
sentence that decides whether the delegation is worth doing, since a puller that returns a
verdict has pre-empted the judgment the delegation exists to keep (R202).
