# Sequence: Bootstrap — the order of business, creation, and repair

**Requirements:** R136, R137, R138, R139, R140, R141, R142, R144, R145, R147, R148, R149, R152, R154, R155, R161, R162, R173, R174, R177

Three diagrams: the gate every non-trivial command passes through, creating a
configuration, and repairing one.

## 1. The order of business

Run before any command that does more than report its version.

```
1. Gate a command on the repository's bootstrap state
1.1. Exempt the commands needing no mini-spec files: --version, help, check-version — return immediately
1.2. Resolve the repository root; on failure, refuse with the no-root crank handle naming the absolute path searched from
1.3. Look for <repo root>/.minispec/config.yaml
1.4. If it is absent, refuse with the no-configuration crank handle and stop — only init may proceed from here
1.5. Load the configuration; if it does not parse, refuse with the malformed crank handle authorising the agent to edit it
1.6. Read the track value; an unrecognised value is malformed, not a fourth behavior
1.6.1. No track at all — refuse with the pre-track crank handle naming --repair, and stop
1.6.2. A value outside the closed set — refuse with the malformed crank handle authorising a hand edit, and stop
1.7. Ask Git whether the tree is a working tree
1.8. If it is not, and track is none, the state agrees — proceed with the git preferences silent
1.9. Ask Git, in one invocation, the ignore state of every path this track value governs
1.10. Compare both facts against the declared value and collect the disagreements
1.11. If any disagree, gripe with the mismatch crank handle naming the repair, and exit
1.12. Report the unmet git preferences, then proceed
```

**Step 1.1 is why `check-version` is exempt.** A skill's first instruction is to run
it, and a version check that refused in an uninitialized project would leave an agent
unable to establish whether its tool matches its skill *before* being told what to do
about that.

**Step 1.4 is the whole adoption path** for a project with no `.minispec/` at all.
Because nothing but `init` creates a configuration, such a project meets this refusal
once. That is the point rather than a side effect — the refusal is how it gets asked the
question, and the crank handle is what carries the ask.

**Steps 1.6.1 and 1.6.2 are the adoption path's other half, and the whole reason they
are two steps is that they resolve differently.** A configuration with no `track`
predates the setting; a flag supplies exactly what is absent, so 1.6.1 names `--repair`.
A configuration with an unrecognised value has been damaged by a hand that may have
changed other things too, so 1.6.2 authorises the hand edit. **Absence is a version
difference; a wrong value is damage.**

Step 3.3 has always drawn that line — it accepts a missing `track` and refuses an
unparseable one. Collapsing 1.6.1 into 1.6.2, which is what a single step did, made the
gate refuse a repair its own repair path would have accepted, and sent the agent to
hand-edit a file a flag could fix. The hand edit then sets `track` and leaves
`.gitignore` unreconciled — half a repair, offered as the only one available.

**1.6.1 asks for less than 1.4 does, deliberately.** A configuration exists, so whether
this is a mini-spec project and where its root sits are both already settled; the only
thing left is the intent `track` was made mandatory for. So 1.6.1 carries the question,
the stop, and the command, and omits 1.4's inspection step entirely.

**Steps 1.7 and 1.9 are separate questions, not one.** Git presence and ignore state
fail independently: a project can gain git after `init` (1.7 disagrees) or have its
ignore lines hand-edited (1.9 disagrees), and the repair message has to say which.

**Step 1.9 costs one invocation** because `git check-ignore` takes a list, so checking
four paths is no more expensive than checking one. That is what makes verifying on
*every* run affordable rather than something to sample.

**Step 1.11 exits rather than warning.** A standing advisory is what this project
already measured failing: the one category carrying only a reminder sat at 19% stale
while every category carrying a forcing function sat at zero.

**Step 1.12 reports but does not exit.** The two git preferences are preferences, not
invariants — an unmet one is worth saying every run and is never worth refusing over.

## 2. Creating a configuration

```
2. minispec init --track-<style>
2.1. Require exactly one --track-<style> flag; refuse when absent, because the value cannot be inferred
2.2. Resolve the repository root the configuration will be written under
2.3. Refuse if .minispec/ already exists, naming --repair and saying to confirm the value with the user first
2.4. Create .minispec/ and write config.yaml holding the track value
2.5. Ask Git whether the tree is a working tree; if not, skip every ignore line
2.6. Ask Git, in one invocation, which of the governed paths some rule already ignores
2.7. Add anchored lines for the paths this value requires ignored and step 2.6 found uncovered
2.8. Report every file created or edited, in full
```

**Step 2.1 refuses rather than defaulting** because the two answers mean opposite
things about a repository's contents, and a default would make that choice for someone
without their noticing.

**Step 2.3 is the precondition that keeps the two forms apart.** Plain `init` requires
absence; `--repair` requires presence. Because the precondition is *inverted* rather
than supplemented, neither form ever has to guess which the caller meant.

**Step 2.5 makes a non-git project quiet rather than broken.** Skipping the ignore
lines is correct there — an instruction naming a file the project does not have is
worse than no instruction.

**Step 2.6 exists because step 2.7 cannot be done by reading the file.** `/PENDING.md`,
`PENDING.md` and `*.md` are the same intent written three ways. The first version of
this matched lines textually and, the first time `init` ran on a real repository,
appended a bare duplicate beside an existing anchored line — which, because the last
matching pattern wins, replaced the project's narrow rule with a broader one. Asking
git is the same move the consistency check already makes for the same reason.

**Step 2.7 writes anchored patterns** for the same reason: every governed path is
mandated at the repository root, so a bare pattern reaches further than anyone asked.

**Step 2.8 is not a courtesy.** These are files the agent did not write, and an agent
that cannot see what changed will reason from a stale picture and eventually assert it.

## 3. Repairing a configuration

```
3. minispec init --track-<style> --repair
3.1. Require exactly one --track-<style> flag, as creation does
3.2. Refuse if .minispec/ does not exist — there is nothing to repair, and creation is the other verb
3.3. Parse the existing configuration; refuse with the malformed crank handle if it will not parse or carries an unrecognised track — a missing track is accepted, being what repair is for
3.4. Set track to the given value, preserving the file's comments, its key order, and every setting this binary does not model
3.5. Ask Git whether the tree is a working tree; if not, skip every ignore line
3.6. Add the anchored lines the new value requires and no rule already covers
3.7. Remove the lines the new value forbids, in either spelling, so reconciliation runs both ways
3.8. Name any forbidden path git still ignores for which no line was found to remove
3.9. Report every file edited, in full
```

**Steps 3.6 and 3.7 together are what "symmetric" means.** A one-directional repair
could only ever make files more private. Because either direction is reachable, a
mismatch under `private-trajectory` can be resolved by ignoring the files *or* by
declaring them public — opposite meanings, and the flag value is how the user chooses
between them. That choice is the user's, which is why the crank handle requires the
agent to confirm before running this at all.

**Step 3.8 is the honest half of step 3.7.** This command owns the top-level
`.gitignore` and nothing else, so a path ignored by a nested `.gitignore`, by
`.git/info/exclude`, or by the user's global excludes stays ignored no matter what is
deleted here. Git says it is ignored and no line names it, so something out of reach is
responsible — and saying so beats reporting a repair that half-worked. A partial
success that announces itself as a success is the failure this layer exists to prevent.

**Step 3.4 edits one key in place rather than rewriting the document.** Reading the file
into the settings this binary models and writing those settings back out is the obvious
implementation and silently discards three things: the comments, the order of the keys,
and any setting the binary does not model — the last of which is a setting from a *newer*
tool version, deleted by an older one without a word. The losses run opposite to their
importance, which is what makes the shortcut worth naming here rather than leaving to
whoever next touches the writer.

It does not promise byte-fidelity. A blank line separating a comment from what it
annotates is not preserved; the comment and its attachment are.

**Step 3.3 refuses rather than repairing**, and that refusal is the one place the agent
is authorised to edit the configuration by hand. `--repair` sets a value; it cannot
undo arbitrary damage, most likely a stray character typed while the file was open in
an editor. So the message states every problem found, points at the skill's
configuration documentation, and says the edit is permitted — at that point the agent
is the only actor left who can act.
