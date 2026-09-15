# Minispec Tool Design

## Intent

A Go CLI tool that handles mechanical operations on mini-spec design files, reducing AI token usage for structural queries and updates.

## Cross-cutting Concerns

### Error Handling
All operations return errors with context (file path, line number where applicable). CLI prints errors to stderr and exits with code 1.

### File Encoding
All files are UTF-8. Tool preserves existing line endings (LF/CRLF).

## Artifacts

### CRC Cards
- [ ] crc-Project.md → `cmd/minispec/main.go`, `internal/project/project.go`, `internal/project/config.go`
- [x] crc-RepoRoot.md → `internal/project/reporoot.go`
- [x] crc-Git.md → `internal/project/git.go`
- [x] crc-Track.md → `internal/project/track.go`
- [x] crc-Init.md → `internal/project/init.go`
- [ ] crc-Parser.md → `internal/parser/testdoc.go`, `internal/parser/types.go`, `internal/parser/requirements.go`, `internal/parser/crc.go`, `internal/parser/design.go`, `internal/parser/traceability.go`, `internal/parser/seqdoc.go`
- [x] crc-Query.md → `internal/query/query.go`, `internal/query/alarms.go`, `internal/query/gaps.go`
- [x] crc-Carve.md → `internal/parser/carve.go`
- [x] crc-Backup.md → `internal/backup/backup.go`
- [x] crc-Update.md → `internal/update/update.go`, `internal/update/alarmfields.go`
- [ ] crc-Validate.md → `internal/validate/validate.go`
- [x] crc-CLI.md → `internal/cli/cli.go`, `internal/cli/bootstrap.go`, `internal/cli/pending.go`
- [x] crc-Phase.md → `internal/phase/phase.go`
- [ ] crc-Alarm.md → `internal/alarm/alarm.go`, `internal/alarm/brief.go`
- [x] crc-Trajectory.md → `internal/parser/trajectory.go`
- [x] crc-Pending.md → `internal/pending/pending.go`
- [x] crc-CommitMessage.md → `internal/pending/message.go`
- [x] crc-TrajectoryValidate.md → `internal/validate/trajectory.go`
- [x] crc-Current.md → `internal/minispecsdom/current.go`
- [x] crc-Done.md → `internal/minispecsdom/done.go`, `internal/minispecsdom/unread.go`
- [x] crc-PendingSdom.md → `internal/minispecsdom/pending.go`
- [x] crc-CarveSdom.md → `internal/minispecsdom/carve.go`
- [x] crc-PartLine.md → `internal/minispecsdom/partline.go`
- [x] crc-MarkerSpan.md → `internal/minispecsdom/partline.go`
- [x] crc-TraceabilityComment.md → `internal/minispecsdom/comment.go`
- [x] crc-TestDoc.md → `internal/minispecsdom/testdoc.go`, `internal/minispecsdom/mdbase.go`
- [x] crc-Gaps.md → `internal/minispecsdom/gaps.go`
- [x] crc-Requirements.md → `internal/minispecsdom/requirements.go`
- [x] crc-Markdown.md → `internal/minispecsdom/mdbase.go`
- [x] crc-Links.md → `internal/query/links.go`
- [x] crc-LinkRepair.md → `internal/update/links.go`
- [x] crc-FinishedCarve.md → `internal/update/finish.go`

### Sequences
- [x] seq-queue-item.md → `internal/pending/pending.go`, `internal/parser/trajectory.go`, `internal/cli/pending.go`, `internal/pending/message.go`
- [x] seq-validate-trajectory.md → `internal/validate/trajectory.go`, `internal/cli/cli.go`
- [x] seq-init.md
- [x] seq-parse.md
- [x] seq-query.md
- [x] seq-update.md
- [x] seq-validate.md
- [x] seq-phase.md
- [x] seq-reporoot.md
- [ ] seq-config.md
- [x] seq-bootstrap.md
- [ ] seq-alarm-freshness.md
- [x] seq-carve-status.md → `internal/parser/carve.go`, `internal/cli/cli.go`
- [x] seq-backup.md → `internal/backup/backup.go`, `internal/project/git.go`
- [x] seq-links.md → `internal/minispecsdom/mdbase.go`, `internal/query/links.go`, `internal/update/links.go`, `internal/update/finish.go`
- [x] seq-current.md → `internal/minispecsdom/current.go`
- [x] seq-done.md → `internal/minispecsdom/done.go`
- [x] seq-pending.md → `internal/minispecsdom/pending.go`
- [x] seq-carve.md → `internal/minispecsdom/carve.go`
- [x] seq-partline.md → `internal/minispecsdom/partline.go`
- [x] seq-anchor.md → `internal/minispecsdom/comment.go`
- [x] seq-testdoc.md → `internal/minispecsdom/testdoc.go`
- [x] seq-gaps.md → `internal/minispecsdom/gaps.go`
- [x] seq-requirements.md → `internal/minispecsdom/requirements.go`

### Test Designs
- [ ] test-Parser.md → `internal/parser/parser_test.go`, `internal/parser/testdoc_test.go`
- [ ] test-Update.md → `internal/update/update_test.go`, `internal/update/alarmfields_test.go`, `internal/update/addreq_test.go`
- [ ] test-Query.md → `internal/query/gaps_test.go`
- [ ] test-CLI.md → `internal/cli/cli_minted_test.go`
- [ ] test-Validate.md → `internal/validate/validate_test.go`
- [x] test-RepoRoot.md → `internal/project/reporoot_test.go`
- [ ] test-Config.md → `internal/project/config_test.go`
- [x] test-Git.md → `internal/project/git_test.go`
- [x] test-Track.md → `internal/project/track_test.go`
- [x] test-Init.md → `internal/project/init_test.go`
- [x] test-Bootstrap.md → `internal/cli/bootstrap_test.go`
- [ ] test-Alarm.md → `internal/alarm/alarm_test.go`, `internal/alarm/brief_test.go`, `internal/query/alarms_test.go`, `internal/cli/cli_alarms_test.go`
- [x] test-Carve.md → `internal/parser/carve_test.go`, `internal/cli/cli_carves_test.go`
- [x] test-Backup.md → `internal/backup/backup_test.go`, `internal/project/git_test.go`, `internal/parser/trajectory_test.go`
- [x] test-Pending.md → `internal/pending/pending_test.go`, `internal/cli/cli_pending_test.go`, `internal/pending/create_test.go`
- [x] test-CommitMessage.md → `internal/pending/message_test.go`, `internal/cli/cli_message_test.go`
- [x] test-TrajectoryValidate.md → `internal/validate/trajectory_test.go`
- [x] test-Trajectory.md → `internal/parser/trajectory_test.go`, `internal/cli/cli_next_id_test.go`
- [x] test-Markdown.md → `internal/minispecsdom/mdbase_test.go`
- [x] test-Links.md → `internal/query/links_test.go`, `internal/cli/cli_links_test.go`
- [x] test-LinkRepair.md → `internal/update/links_test.go`, `internal/cli/cli_repair_test.go`
- [x] test-FinishedCarve.md → `internal/update/finish_test.go`, `internal/cli/cli_finish_test.go`
- [x] test-Current.md → `internal/minispecsdom/current_test.go`
- [x] test-Done.md → `internal/minispecsdom/done_test.go`
- [x] test-PendingSdom.md → `internal/minispecsdom/pending_test.go`
- [x] test-CarveSdom.md → `internal/minispecsdom/carve_test.go`
- [x] test-PartLine.md → `internal/minispecsdom/partline_test.go`
- [x] test-TraceabilityComment.md → `internal/minispecsdom/comment_test.go`
- [x] test-TestDoc.md → `internal/minispecsdom/testdoc_test.go`
- [x] test-Gaps.md → `internal/minispecsdom/gaps_test.go`
- [x] test-Requirements.md → `internal/minispecsdom/requirements_test.go`

## Documentation

- [x] docs/user-manual.md
- [x] docs/developer-guide.md

## Gaps

- A1: R37 (MCP server mode) deferred to future version
- A2: R1-R66 pre-existing code lacks inline requirement refs

- T1: R56 retired by R117 (2026-08-07 repository-root detection)
- [ ] O1: query uncovered lists retired requirements while validate skips them — R56 appears there now that its ref was correctly removed, reading as work to do. Both are defensible for their purpose, but the raw query invites a reader to re-cover a requirement that is deliberately dead
- [ ] O2: DetectFrom in project.go duplicates the new isDir helper from reporoot.go (os.Stat + IsDir inline). One-line reuse, same package, noticed during simplification of the repo-root work
- [ ] O3: Project.RootPath is the design root but its name says neither — the exact ambiguity R107 exists to remove. Renaming touches every call site, so it was left out of the repository-root work rather than folded in
- [ ] O4: No test covers loadProject end-to-end against a real design root: resolveConfigFrom is well covered, but the wiring from Detect through loadProject to a Project with Origins populated is only exercised by running the binary
- [ ] O5: gate() and runInit() have no automated test: both resolve the repository root from the process working directory, so testing them needs os.Chdir — process-global state, unsafe alongside parallel tests. The refusal messages, the exempt/known command sets and the flag table are covered directly (test-Bootstrap.md); the wiring between them is only exercised by running the binary, which was done by hand end-to-end this session. Repair: thread a start directory through gate as RepoRootFrom already does for RepoRoot. **Amended 2026-08-16: the stated blocker is softer than it reads.** internal/cli/cli_next_id_test.go now chdirs into a temp tree and drives runQuery, restoring via t.Cleanup — so chdir testing is workable today, provided the test is never marked parallel. That does not close this gap (gate and runInit are still untested and threading a start directory is still the better repair) but it removes "we cannot test it" as the reason, which is the part that would otherwise keep the gap parked forever. Note t.Chdir, the tidier form, needs go1.24 while the module declares go1.21
- [ ] O6: Git.Tracked reports a genuine git failure as "not tracked". ls-files --error-unmatch exits non-zero both when a path is untracked and when the command itself fails, and the error is discarded. Harm is bounded — a spurious preference note, never a refusal — but it is the project's own "report absence as error, never as silence" theme violated at a seam that exists to interrogate an external tool. Repair: distinguish exit 1 from other failures
- [ ] O7: printUsage's query-subcommand list is a hand-maintained mirror with no completeness check. knownCommands has TestKnownCommandsCoversEveryDispatchedCommand asserting it covers every dispatched top-level command; nothing does that for the query subcommands, which live in a prose block. next-id was added to it by hand and nothing would have reported the omission. Repair: a table the dispatch switch and the usage text both read, or a test that greps the switch cases against the block
- [ ] O8: strconv.Atoi(strings.TrimPrefix(id, "R")) is now a sixth copy of one idiom — validate.go:248, :431, :454, update.go:extractNum, and query.go:nextReqID. A parser.ReqNum(id) (int, bool) collapses all six. Noticed during simplification of the next-id work; left out because it is a cross-package edit of code that pass did not touch
- [ ] O9: A done entry whose header carries no identifier slot is skipped silently. `parseDoneIDs` reads the slot between the date's em dash and the title's colon; a header with no colon yields nothing and is not distinguished from one that legitimately discharged no ID. Measured 2026-08-16 in ark: 2 of 54 entries are in a pre-adoption shape with no colon, hiding queue IDs #84, #83 and #46, and the per-file count (R197) reports 46 identifiers without saying that 2 entries were unread. Latent rather than live — ark's maximum #117 comes from a conforming entry — but it is the R190 collision arriving through a third door. The fix is conformance reporting, which belongs to `validate trajectory` (carves/trajectory-tool.md#3), not to next-id

- [x] O10: A bare `**Inject:**` anchor may resolve to a *use* or a *doc comment* rather than its declaration. `sitePattern` (R206) hands git a bounded name, and git's `-L` takes the first line that matches — which in this repository's own `sdom/doc.go` on `old-sdom` was the `// Render returns …` comment two lines above `func (d *Doc) Render`. Measured 2026-09-04. The alarm then watches the comment's lines and reports a clean reading over a rewritten body. `old-sdom` fixed this with an extent computed from a parse (its R404–R406) instead of asking git for the range; that is the reclaim behind the stopgap, over `github.com/zot/simple-dom`'s declarations once the module dependency is wired.
- [x] O11: git's `-L` range for a declaration runs to the line before the next declaration, so it includes the trailing blank line — and appending code after a method makes that method report *changed*. Measured 2026-09-04 while testing R205: `func (a *A) Run()` with nothing after it was attributed to the commit that appended `B` below it, and the test had to place a function after each method to get a stable range. In a live repository this is a **spurious stale** on every alarm whose site is the last declaration in its file. The same repair as the gap above: an extent from a parse, which stops at the closing bracket.
- [x] O12: The per-item commit discipline is a rule an agent must remember, and nothing checks it. DECIDED (Bill, 2026-09-04): one squashed commit per queue item, checkpoints folded with `squash` (never `fixup`) so every message survives, the body rewritten as `Step N — <subject>` sections in commit order so the sequence is visible (successive checkpoints touch the same files and a later one can alter an earlier one), one sign-off, and no commit hash in a `**Pulled:**` line since the squash rewrites it away. The reason it is a rule and not a mechanism: the census asks git and git sees only committed code, so the honest shape was commit → pull → record → commit, three per item at the floor and five on 2026-09-04. The tool's half: `pending finish` is the natural place to perform or verify the squash — it already knows the item's first commit from the current file's context and leaves the carve flip as an uncommitted tail today — and a `**Pulled:**` record that named a commit the tool owns would close mini-spec-tool's `O20` (the same-day blind spot) at the same time. Until then the rule lives in SKILL.md and in memory.
- A3: R235 — the slot is cranked out in full by its verbs and documented in `/mini-spec`, revert especially — has no code site in this design root. The documentation half is discharged in `.claude/skills/mini-spec/SKILL.md` (*When a queue operation goes wrong*), which is **outside this design root** and so earns the requirement no implementation coverage; the crank-out half is the `pending revert`/`replay` verbs' and lands with them (carves/done/sdom-reclaim.md Item 3, landed as `#69`), at which point this entry can be revisited. Approved rather than open because nothing in `tool/` will ever satisfy the documentation half, and a check that reports a permanent absence is a nag rather than a gap.
- [x] O13: The dependency's Pending reader reads an entry title with a lazy `\*\*(.*?)\*\*` match, so a title carrying emphasis inside it would read back cut at the interior run — the flanking-rule read the August tree had (old R389's second half). Measured 2026-09-05: 0 of 11 live entries carry one. Raised with mini-spec-tool, which held the reader until the 2026-09-14 move; the write half (R255) refuses only a wrapped title
- [x] O14: `minispecsdom.Pending.Place` at the last position appends at the end of the document, after any `---` rule and commentary that follow the entries — a queue entry inside the commentary, which trajectory-format.md forbids (R260). Measured 2026-09-05 by probe on a fixture with trailing prose. The live PENDING.md ends with its entries, so `--last` places correctly there today. Raised with mini-spec-tool, which held the reader until the 2026-09-14 move; the adapter has no API to insert before the rule
- [x] O15: `add-item --last` then `finish` leaves the pending file one blank line longer than it started: `Pending.Place` at the end prepends a newline when the document does not end in one, and `Remove` does not take it back, so the two are not inverses on the file (R270). Measured 2026-09-05 with and without trailing commentary. `TestAddThenFinishLeavesTheQueueFileByteIdentical` skips naming this gap until the dependency lands the fix
- [x] O16: `minispecsdom.Carve.SetMarker` treats only `OPEN` as a transient, so `pending replay` after `pending revert` appends `**OPEN (#N.)**` beside the `**REVERTED (#N.)**` it should replace — one line asserting both states, the contradiction the marker rule exists to prevent (R230). `specs/backup.md` (R230) and the skill name REVERTED as the transient a revert writes; the format's `### Markers` section never listed it, which is the gap on our side, closed the same day. Measured 2026-09-05 by smoke test on a copy of this repository and by probe; Item 2's replay test asserted the new marker's presence and could not see the survivor. Raised with mini-spec-tool, which held the reader until the 2026-09-14 move
- [x] O17: `minispecsdom.Carve.SetMarker` with no transient to replace appends the marker at the end of the line, after any trailing prose — `… **REVERTED (#N.)** Needs Item 1. **OPEN (#N.)**` — where the format's grammar is Head Marker* Text?, so the written line is one the reader lists as non-conforming. Measured 2026-09-05 by probe. Raised with mini-spec-tool, which held the reader until the 2026-09-14 move
- [ ] O18: The independent flat-scan cross-check of every carve's `## Status` block (old-sdom R449, R450, R452, R453: a second opinion written from the documented format, never from the reader's own pattern, so a swallowed part line is found by something that does not read through the same parse) is not carried into `validate trajectory`; R298 reports the reader's own `Stateless()` list instead, which cannot see a line the parse never returned. Its home is queue item #55, which owns `UnreadEntries` and its callers
- [ ] O19: Two readers of the queue files coexist: `ScanTrajectory` still reads item IDs by regex for `next-id item` (its alarms in test-Trajectory.md pin `parseDoneIDs`), while `validate trajectory` reads entries through the dependency's Pending and Done readers (`ScanQueue`). They should converge on the dependency or be made an explicit second opinion that reports where they disagree; today a disagreement is silent. Measured 2026-09-05 on this repository: regex 58 IDs in DONE.md, the dependency 17 entries — see the next gap
- [x] O20: The dependency's markdown base absorbs from a double-backtick code span to the end of the file: DONE.md line 711 carries ``What vouches for an ` `` — valid CommonMark, a two-backtick span holding one backtick — which the bracket table, knowing only one- and three-backtick spans, reads as an empty span plus a stray opener, inverting the file's backtick parity from there; `minispecsdom.ParseDone` returns no further entries and lists nothing unread; 41 of 58 entries vanish silently, and every downstream check reads the remainder as the whole ledger. Found 2026-09-05 by porting `validate trajectory`, made visible by R300's second-opinion check, which lists the 42 IDs. `make validate` is red on this repository until the dependency reads the span correctly (or the sentence is rewritten, which would be repairing the subject for the instrument's sake). Sent as `requests/double-backtick-span-absorbs.md`, corrected and decided in `requests/backtick-run-groups.md` (groups for five backticks down to one)
- T2: R205 retired by R304 (2026-09-06 sdom-reclaim Item 6: the extent replaces the pattern)
- T3: R206 retired by R303 (2026-09-06 sdom-reclaim Item 6: the extent replaces the pattern)
- [ ] O21: Only Go sources have a site extent: `siteExtent` parses with `sdom.LangGo` and `schema.Go`, so an `**Inject:**` naming a markdown heading or a symbol in another bundled language reads unresolvable. Every site in this repository and mini-spec-tool is Go today (225 of 225, measured 2026-09-06); old-sdom's `headingExtent` (its R402) is the markdown half to reclaim when one appears
- [ ] O22: `validate` does not report an unclosed backtick run, fence or `<!--` in `design/` or `specs/` documents. `validate trajectory` reports it for the queue files, the current file and every carve (R302); the design-document readers here are the pre-sdom line readers and say nothing. On `old-sdom` R443 reported both, wired 2026-08-22 after one unclosed run hid 78 of 115 gap entries for a day. Returns with the readers `requests/three-readers.md` asks for
- [ ] O23: `update add-gap` and `update retire` take their prose as a shell argument; the trajectory verbs take `--…-file` forms because a backtick in a double-quoted argument is command substitution and the text vanishes silently. `old-sdom` gave every prose-taking verb a file form (`add-gap --body-file`, `retire --reason-file`); these two are the ones this tree still lacks
- [x] O24: `pending add-item` on a part marked `REVERTED` refuses ("part N already carries queue ID #M; a part records exactly one item") before the backup slot's release of the reverted attempt runs, so re-adding the very part that was reverted — the common case — is refused, while a mutation on a sibling part releases it and the re-add then succeeds. Measured by mini-spec-tool 2026-09-06 (`requests/reverted-part-readd-collision.md`). Repair: run the release before the collision check, or exempt a `REVERTED` marker naming the slot's own reverted item
- T4: R379 retired by R396 (2026-09-04 sdomification Item 4: Unread carries lines)
- T5: R373 retired by R407 (2026-09-05 sdomification Item 6: Place lands where the entries end)
- T6: R375 retired by R408 (2026-09-05 sdomification Item 6: Remove closes the tail; Place then Remove is the identity)
- T7: R367 retired by R409 (2026-09-05 sdomification Item 7: REVERTED is a transient; a marker inserts before trailing prose)
- [ ] O25: `minispecsdom.Pending.Place` accepts entry text that opens a bracket group never closed (a lone asterisk in a status), so the entry itself reads back and the check passes while every later entry falls inside the unclosed run and is not read. A write path should refuse text whose placement grows `Unread`, since the file it produces is one its own reader cannot fully read. Ported 2026-09-14 from mini-spec-tool O28, measured there 2026-09-06 on its queue
- [ ] O26: `minispecsdom.Gaps.Add` appends its line to a file with no final newline without supplying one, so the new gap glues onto the last gap's line and the read-back finds nothing; the write should end the previous line before appending, as `Requirements.Add` and the trajectory writers do. Ported 2026-09-14 from mini-spec-tool O30, measured there 2026-09-07 at `ceafd0c`, where every `add-gap` was refused until a newline was added by hand
- [x] O27: Every relative link in a carve breaks when the carve moves to `carves/done/`, and nothing rewrites them: `query links carves/done/*.md` reads 32 `missing` in `carves/done/trajectory-tool.md` alone, every one written relative to `carves/` (measured 2026-09-15, the verb's first run). The skill says to rewrite the links as part of the move; the move is a hand `git mv` today, so the rule has no forcing function. Repair is two-sided: rewrite the 32 in the document, and either give the tool a move verb that rewrites `../` and `done/` links, or wire `query links` into `validate` (reference-discipline Item 3) so a broken move is reported on the next run rather than found by a cloner. Filed 2026-09-15 as [reference-discipline.md](../../carves/done/reference-discipline.md) Item 4, which carries the predicate and the two open decisions
- [x] O28: `update finished-carve` and `update repair-links` read links only in the live carves, the done carves and the three trajectory files, so a spec or design document that links a carve — `specs/index.md` links `carves/done/minispecsdom-move.md` today — is outside both and breaks unreported when that carve moves. The population is repository-scoped and the design roots are below it (two here), so widening it means walking every `design/` and `specs/` under the root; `query links` over those files after the fact is the workaround (2026-09-15, #85).
- T8: R244 retired by R476 (2026-09-15 item-identifiers: the item number is the identifier, no hash)
- T9: R245 retired by R477 (2026-09-15 item-identifiers: the item number is the identifier, no hash)
- T10: R247 retired by R478 (2026-09-15 item-identifiers: the item number is the identifier, no hash)
- T11: R455 retired by R488 (2026-09-15 reference-discipline Item 2: the class is git's, the population is every tracked markdown file)
- T12: R463 retired by R489 (2026-09-15 reference-discipline Item 2: the class is git's, the population is every tracked markdown file)
- T13: R472 retired by R490 (2026-09-15 reference-discipline Item 2: the class is git's, the population is every tracked markdown file)
- T14: R485 retired by R491 (2026-09-15 reference-discipline Item 2: the class is git's, the population is every tracked markdown file)
