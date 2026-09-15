# CLI
**Requirements:** R1, R2, R35, R36, R49, R50, R54, R55, R60, R62, R79, R80, R81, R82, R83, R89, R103, R117, R115, R116, R152, R153, R154, R155, R156, R157, R158, R159, R160, R169, R174, R175, R176, R189, R196, R198, R199, R204, R207, R208, R212, R213, R215, R216, R217, R218, R241, R244, R247, R253, R254, R255, R256, R261, R262, R264, R265, R268, R277, R280, R281, R282, R283, R285, R299, R301, R313, R314, R315, R317, R323, R324, R329, R331, R332, R334, R335, R336, R455, R458, R461

Command-line interface handling.

## Knows
- args: command-line arguments
- flags: parsed flags (--design-dir, --src-dir, --quiet, --json, --version)
- outputFormat: text or json
- version: build-time version string

## Does
- `pending changes`: the cleanup report over the anchor — names the question it answers,
  lists each moved path with its category, hands over the two repair commands (R335, R336)
- the missing-layer crank handle: on a queue verb's typed refusal, print which files are
  missing, the `add-item --create` that creates them, and what `track` means for them (R332);
  `init carve <name>` reports the scaffold's write and the add-item that queues it (R334)
- the minting verbs — `retire`, `migration-complete`, `add-req` — report their value as a
  sentence on stdout, suppressed by `--quiet`, carried by `--json`; never a bare value on
  stdout as a return (R331)
- describeWrites(repoRoot, files): the queue verbs' `written` line with each file's kind from
  git — tracked and uncommitted, ignored, untracked — so the carve flip is marked as the write
  that still needs a commit; bare names with no repository (R329)
- Parse(args): parse command and flags
- Run(): dispatch to appropriate handler (or show version if --version)
- CheckVersion(): find skill README.md under the **repository root** first, then the user's home directory, in .claude/skills/mini-spec/; extract Version: line, compare against tool version. Exit 0 if match, 1 if mismatch or not found. (R117)
- Gate(cmd): before dispatching anything that does more than report a version, require a repository configuration and a verified `track`. `--version`, `help` and `check-version` are exempt; `init` is the only command that may proceed without a configuration (R152, R153, R154)
- refuseNoConfig(): the no-configuration crank handle (R155–R159)
- refusePreTrack(): the crank handle for a configuration that parses but sets no
  `track`. Names `minispec init --track-<style> --repair`, asks the agent for the
  user's intent alone, and stops there (R174–R176)
- Output(data): format and print result (text or JSON)
- Error(err): print error to stderr
- PrintVersion(): display version and exit

**The refusal is written for the agent, and the agent writes for the human.** The tool
supplies the gist; a verbatim script cannot adapt to what the agent found, and would
still say "if this is a code project" after the agent had established that it plainly
is or plainly is not. One register, one rule.

Three things that message must carry, each closing a hole the others do not:

- **The absolute path** it would make the repository root, so the user's yes lands on a
  stated location rather than on "here" (R155).
- **The instruction to look, and the stop.** Whether this directory holds code is a
  *fact*, so the agent goes and checks rather than asking the user — but the moment the
  message says "check," a weaker agent can conclude *yes* and run `init`, building the
  road deliberately left unbuilt. So it ends with **report and wait** (R156, R157).
- **The negative finding it is really hunting for.** This refusal exists because
  someone ran from *above* their project, and a directory full of project directories
  is exactly what an agent can recognise. "This looks like a folder of projects" is more
  useful than a yes, and it is the case that would otherwise plant a repository root in
  the wrong place (R158). Declined, the answer is to run from inside the project (R160).

**Why `check-version` is exempt (R153).** A skill's first instruction is to run it. A
version check that refused in an uninitialized project would leave an agent unable to
establish whether its tool matches its skill *before* being told what to do about that.

**Every project predating `track` meets a refusal once (R169), but not the same one.**
That is the adoption path rather than a side effect — a refusal is how such a project
gets asked the question it was never asked. Which refusal depends on what it already
has, and the two name opposite verbs:

- **No `.minispec/` at all** — the no-configuration refusal above, answered by plain
  `init`.
- **A configuration written before `track` existed** — `refusePreTrack`, answered by
  `--repair`, because plain `init` refuses where `.minispec/` is present.

**`refusePreTrack` asks for strictly less (R175).** The no-configuration refusal has to
send the agent looking, because whether this directory is a project at all is unknown. A
configuration existing settles that, and settles the repository root with it. What is
left is the one thing no inspection can supply — whether the queue stays private or ships
— so the message carries the question, the stop (R176), and the command, and nothing
else.

**It is a separate message because the alternative was measurably wrong (R174).** Routing
a missing `track` to the malformed refusal told the agent to hand-edit a file that
`--repair` accepts, and a hand edit sets the value while leaving `.gitignore` alone.
- query carves: answered **before any design root is resolved**, beside `next-id item`, because
  carves are repository-scoped (R213); `--open` and `--json` parsed wherever they sit
  (R215); one line per carve, a part row for every non-conforming part and — under `--open`
  — every open part and clean stateless line, then the census with zeros (R207, R212, R216,
  R217, R218)
- query alarms: parses `--unverified` and `--brief` as **subcommand** flags through the same
  mechanism, and for the same two reasons — a global flag written after the subcommand is
  silently dropped, and a selection applied inside one rendering is a filter the other
  rendering ignores without a word (R199, R204)
  - The two are orthogonal by construction: `--unverified` chooses *which* alarms, `--brief`
    chooses *what each one prints*. They compose, and neither implies the other

## Collaborators
- Project: to initialize project context and to resolve the repository root for CheckVersion
- Init: for the `init` command, and named by the refusals as the way forward
- Track: to verify the declared value before any non-trivial command runs
- Query: for query subcommands
- Update: for update subcommands
- Validate: for validate command
- Phase: for phase subcommands
- flag: for flag parsing
- encoding/json: for JSON output

## Subcommands
```
minispec check-version
minispec query <subcommand>          # ... project, migrations
minispec update <subcommand>         # ... retire, migration-complete
minispec validate
minispec phase <spec|requirements|design|implementation|gaps>
```

## Notes
`query comment-patterns` output includes a closers section when closers are configured, with a WARNING that unclosed block comments silently swallow code.
