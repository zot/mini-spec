# Test Design: TrajectoryValidate
**Source:** crc-TrajectoryValidate.md

Every alarm below targets a property that **passes by default**. A checker that resolved
nothing, or that swept prose for citations, satisfies any test written without these
fixtures.

**Reclaimed from `old-sdom` on 2026-09-05 (`#70`), over the dependency's readers.** Every
`**Pulled:**` line was dropped in the port and the injections re-run on this tree before being
written back. Two August cases — *a swallowed part line is reported* and *a suppressed status
heading is reported* — are not carried: they proved the independent flat-scan cross-check,
which is gap `O18`'s. Two cases are new: the deviating stateless line (R298) and the second
opinion of the two readers (R300), the latter written the day its fixture happened on this
repository's own ledger (`O20`).

**Pulled on this tree 2026-09-05, all fifteen, with two probes past the list.** Fourteen rang; alarm 3
cannot ring from this side and says so. Of the probes, dropping the `carves/` guard in `missingPart`
changed nothing because the carve loop already answers the same question — a redundant guard, not a
missing alarm — and dropping the landed guard in `checkOrphans` was silent until the fixture gained an
open part citing a live item, which is now part of alarm 6's case.

## Test: a dangling citation is found
**Purpose:** every `#N` a carve cites resolves to a pending or done entry (R287)
**Input:** a carve whose status block cites `**OPEN (#99.)**`, against queue files holding `#1` and `#2`
**Expected:** one dangling citation, naming the carve and the part key
**Refs:** crc-TrajectoryValidate.md, seq-validate-trajectory.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 1
**Fire alarm:** resolve citations against the pending file alone. Goes red when the target sits in the done file — which is the common case, since most cited items have completed. The same one-file failure R190 records for `next-id`, arriving through a third door
**Inject:** internal/validate/trajectory.go:idSet
**Pulled:** 2026-09-05 — rang on two tests, `got 2 dangling, want 1` — the done-file citation `#2` reported dangling — and the numbering check went with it

## Test: a queue entry missing from its carve is found
**Purpose:** the direction the carve→queue check structurally cannot see (R288)
**Input:** a pending entry whose `Source:` names `carves/x.md` part `#7`, against a carve whose status block has no `Item 7`
**Expected:** one finding, naming the queue entry and the key it claims
**Refs:** crc-TrajectoryValidate.md, seq-validate-trajectory.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 2
**Fire alarm:** delete the queue→carve pass entirely. Goes red here and **stays green on every other test**, which is the point: an item completing against a part its carve never recorded is invisible from the carve side, so only this direction reports it
**Inject:** internal/validate/trajectory.go:RunTrajectory
**Pulled:** 2026-09-06 — rang again after item 71's commit landed past midnight and staled it, `got 0 missing-part findings, want 1`; restore byte-clean by copy. Previously 2026-09-05, same signature, every other test green

## Test: a backquoted prose example is not a citation
**Purpose:** citations are ingested by position — a part line's marker — and never from a pattern swept over prose (R289)
**Input:** this repository's own line 75 shape, verbatim: a status block containing the prose sentence ``Ark's live carves already read this way — `**Item 8 — a test harness…** **OPEN (#121.)**`.`` alongside one real part citing a live ID
**Expected:** exactly one citation found, and no dangling report for `#121`
**Refs:** crc-TrajectoryValidate.md, crc-Carve.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 3
**Fire alarm:** collect citations by sweeping `**[A-Z][A-Z ]*\(…\)` across the whole status block instead of reading part lines. Goes red by reporting `#121` dangling. **This is not a strawman — it is the real defect, run and believed on 2026-08-16 while scoping this very item**, and the fifth ad-hoc instrument in this layer's history to return a confident wrong answer. Restoring it is what proves the test catches *that*, rather than proving the checker dislikes being shuffled
**Inject:** internal/parser/carve.go:QueueID
**Pulled:** 2026-09-05 — **did not ring, and cannot from this side.** The August injection swept the whole status block inside `ParseCarve`; here the reader is the dependency's and prose is never a `Part`, so a sweep over `rendered(p.part)` reaches only part lines and the backticked example is unreachable by construction. Recorded as a pin on the dependency's structure rather than a proof of this code; a first attempt with fresh imports did not build

## Test: a repository with no trajectory layer passes
**Purpose:** nothing that could be inconsistent is a clean result, not a failure and not "could not check" (R286)
**Input:** a repository root with neither queue file and no carve directory
**Expected:** no issues, exit 0, and a report that says the layer is absent
**Refs:** crc-TrajectoryValidate.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 4
**Fire alarm:** return an error, or an empty clean result with no statement. Both go red — the first on the exit assertion, the second on the report. The second half matters more: a silent clean result over an absent layer is indistinguishable from a clean result over a checked one, which is this project's own theme violated
**Inject:** internal/validate/trajectory.go:RunTrajectory
**Pulled:** 2026-09-06 — rang again after item 71's commit landed past midnight and staled it, `a repository with no trajectory layer was not reported absent` and `the report does not state that the layer is absent`; restore byte-clean by copy. Previously 2026-09-05

## Test: only a cross-file ID collision is reported
**Purpose:** an ID held by **both** queue files is a collision; repetition *within* the done file is a legitimate staged record (R290)
**Input:** two trees — one whose done file records `#41` three times as staged passes, and one where `#8` is live in the pending file and completed in the done file
**Expected:** no finding for the staged record; exactly one for the collision, naming both files
**Refs:** crc-TrajectoryValidate.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 5
**Fire alarm:** report within-file repetition as reuse. **This is the real defect, not a strawman** — the first draft of this check did exactly that, and running it against ark reported `#41` (five staged entries: `Pass 1`, `2a`, `2b`, `2c`, `CLOSED`), `#65` and `#98` as reused IDs. All three were correct work. A check that fires on correct work is the shape that gets muted, and a muted check reports nothing at all; the spec and R290 were both narrowed as a result, so this test guards a *corrected* rule rather than the one originally written
**Inject:** internal/validate/trajectory.go:checkDuplicateIDs
**Pulled:** 2026-09-05 — rang, `a staged multi-entry record was reported as reuse: #41 …` — the real defect, restored

## Test: both orphan forms are found
**Purpose:** a landed part with no done entry, and a done entry naming a part no carve records (R291)
**Input:** a carve part marked `**LANDED (`abc1234`, 2026-08-16 — `#5`.)**` with no `#5` in the done file; and a done entry whose `Part` pointer names `carves/x.md#42`, a key that carve does not hold
**Expected:** two findings, one of each form
**Refs:** crc-TrajectoryValidate.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 6
**Fire alarm:** implement only the first form. Goes red on the second, which is the one a reader is least likely to notice by eye: the carve looks complete and the ledger looks full, and only the join between them disagrees
**Inject:** internal/validate/trajectory.go:checkOrphans
**Pulled:** 2026-09-05 — rang, `got 1 orphans, want 2` — the second form, the one nobody catches by eye, gone. **A probe past the list the same day** dropped the *landed* guard instead and nothing objected; the fixture gained an open part citing a live item, and the guard now rings as `got 3 orphans, want 2`

## Test: a done entry with no identifier slot is named unmigrated
**Purpose:** a header with no colon yields nothing, and nothing is also correct for an entry that discharged no ID — so the shape itself must be checked (R292)
**Input:** three done entries — one conforming, one in a pre-adoption shape with no colon, and one that legitimately discharged no ID
**Expected:** exactly one unmigrated finding, and the legitimately-empty entry is not reported
**Refs:** crc-TrajectoryValidate.md, crc-Trajectory.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 7
**Fire alarm:** report every entry that contributed no ID. Goes red on the third fixture by flagging a correct entry — the false-positive direction, which is the one that gets a check muted. Measured 2026-08-16: 2 of ark's 54 entries are genuinely unmigrated, hiding three queue IDs, while `next-id item` reported 46 identifiers with no indication that two entries went unread
**Inject:** internal/validate/trajectory.go:checkLedgerConformance
**Pulled:** 2026-09-05 — rang, `got 2 unmigrated, want 1` — the false-positive direction, flagging the entry that legitimately discharged no queue ID

## Test: a line whose markings disagree is found
**Purpose:** checkbox, strikethrough and marker must agree, the checkbox being authoritative (R294)
**Input:** five status lines — one agreeing, one `- [x]` with no strikethrough, one `- [ ]` struck through, one `- [ ]` carrying a `**LANDED**` marker, and one `- [x]` struck but carrying an `**OPEN**` marker
**Expected:** four findings, one per branch, and the agreeing line is silent
**Refs:** crc-TrajectoryValidate.md, crc-Carve.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 8
**Fire alarm:** compare only checkbox against strikethrough. Goes red on the fourth line. **This check finds nothing in the live corpus** — measured 2026-08-16, no line among nine carves disagrees with itself — so it is a sentry rather than a repair, and the fixtures are the only place it will ever ring until the corpus drifts. That is exactly the case where an untested guard rots unnoticed
**Inject:** internal/validate/trajectory.go:checkLineAgreement
**Pulled:** 2026-09-05 — rang, `got 2 disagreements, want 4` — both marker branches gone, both counted by the fixture

## Test: the report says what it could not read
**Purpose:** conformance is named with its migration target, and unreachable references are stated rather than passed over (R293)
**Input:** a carve keyed on the superseded bare `#N` scheme, so its citations sit in key position where the position rule cannot see them
**Expected:** the report names the migration target and states that the citations were unreachable — it does **not** report a clean integrity result
**Refs:** crc-TrajectoryValidate.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 9
**Fire alarm:** drop the unreachable-count from the report. Goes red on the second assertion while the first still passes, which is the shape of the underlying danger: naming the migration target *looks* like a complete report, and a reader takes the silence about integrity as a pass. Measured 2026-08-16: ark yields 3 reachable citations, the rest sitting in key position
**Inject:** internal/validate/trajectory.go:FormatText
**Pulled:** 2026-09-05 — rang on both assertions — the migration target and the coverage statement both live in the dropped note

## Test: item numbers in no readable entry are found
**Purpose:** every number from 1 to the maximum assigned appears in a readable entry (R295)
**Input:** two ledgers — one contiguous, and one holding `#1`, `#3` and `#5` with `#2` and `#4` absent
**Expected:** no gaps in the first; exactly `[2 4]` in the second, and it counts as a finding rather than a note
**Refs:** crc-TrajectoryValidate.md, seq-validate-trajectory.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 10
**Fire alarm:** silence the numbering loop. Goes red on both the gap list and the `HasIssues` assertion. `trajectory-format.md` says a gap in the sequence is *expected*, which is true of a deliberately abandoned ID — this check exists because "expected" is a claim worth testing, and the measurement disagreed: ark has 17 gaps and 16 are mentioned in its own queue files, `#46` nineteen times
**Inject:** internal/validate/trajectory.go:checkNumbering
**Pulled:** 2026-09-05 — rang on three tests, `MissingIDs = [], want [2 4]` first

## Test: unrecognized entry-like lines are counted
**Purpose:** what the reader could not recognize is reported as coverage, and the gap finding names it (R297)
**Input:** a pending file with one numbered item and one unnumbered `## add to the list` — ark's real shape — and a done file with two recognized entries plus one in the older date-outside-the-bold form
**Expected:** one unread line per file, a coverage note in the report, and the gap finding pointing at it
**Refs:** crc-TrajectoryValidate.md, crc-Trajectory.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 11
**Fire alarm:** drop the per-file count. Goes red on the coverage note. **The fixture had to be built around a real limit of the sibling check**, which is worth knowing before trusting either: an unrecognized entry holding the *highest* ID does not raise the maximum, so it leaves no gap and `checkNumbering` cannot see it at all. Only this count can — which is why this is the honest half and R295 the symptom. Measured 2026-08-16 in ark: 139 entry-like lines in the done file, 54 recognized, and 6 unnumbered items in the pending file, two of them titled "add to the list"
**Inject:** internal/validate/trajectory.go:countUnread
**Pulled:** 2026-09-06 — rang again after item 71's commit landed past midnight and staled it, `Unread = map[], want one per file` (the `record` closure made to return always); restore byte-clean by copy. Previously 2026-09-05, same signature

## Test: the current file and each carve are counted as unread too
**Purpose:** the coverage note names every file a reader could leave partly unread, not the two queue files alone (R302)
**Input:** a current file whose Active section opens a fence it never closes, and a carve whose body opens a code span it never closes, beside a pending file and an empty done file
**Expected:** one unread line under `CURRENT.md` and one under `carves/x.md`, and the note naming both with their counts
**Refs:** crc-TrajectoryValidate.md, crc-Trajectory.md, crc-Carve.md
**Code:** internal/validate/trajectory_test.go
**Fire alarm:** drop the carve loop from `countUnread`, so only the queue files and the current file are counted. Goes red on `carves/x.md (1)` missing from the note
**Inject:** internal/validate/trajectory.go:countUnread
**Pulled:** 2026-09-06 — rang again after item 71's commit landed past midnight and staled it, `Unread = map[CURRENT.md:1], want one for the current file and one for the carve`; restore byte-clean by copy. Pulled twice on 2026-09-05, the second after the simplification pass rewrote `countUnread` around a `record` closure

## Test: the machine-readable form uses one key convention
**Purpose:** every JSON key is snake_case, and the fields a consumer needs are present (R299)
**Input:** a scan with findings in several buckets, marshalled
**Expected:** no key differs from its lowercase form; `missing_ids`, `unreachable` and `absent` are all present
**Refs:** crc-TrajectoryValidate.md, crc-CLI.md
**Code:** internal/validate/trajectory_test.go
**Alarm:** 12
**Fire alarm:** drop the json tag from one field. **Written because the sibling command shipped exactly this defect** — `query carves` emitted Go field names beside snake_case in one document until gap `O18` caught it by audit rather than by check, and nothing validates JSON shape anywhere in the tool. This guards the class at the point it would recur
**Inject:** internal/validate/trajectory.go:TrajectoryIssues
**Pulled:** 2026-09-05 — rang, `key "Unreachable" is not snake_case — a Go field name reached the JSON`. Site is the struct, so Inject names the type

## Test: `CURRENT.md` without exactly one `## Active` is reported
**Purpose:** validates R296 — the heading bounds the region `pending finish` clears, and `activeRange` refused both the missing one and a duplicated one *at write time only*, so a corrupted file sat clean until the next completion met it. Measured 2026-08-19: green twice in one session over a file the verb had just damaged, in files git cannot diff, where the backup slot holds one level of undo
**Input:** a `CURRENT.md` with no `## Active`; one with two; and one with exactly one
**Expected:** the first two each report one structure finding naming which shape it is; the third reports none
**Refs:** crc-TrajectoryValidate.md — R296
**Code:** internal/validate/trajectory_test.go
**Alarm:** 13
**Fire alarm:** report only the **missing** heading and let a duplicated one pass, and confirm this goes red on the two-`Active` case alone. **The duplicate is the shape that actually occurred** — `O62` corrupted the file by matching `## Active` as a substring inside a backticked mention — so a check that looks only for absence would have reported OK on the very file that prompted this. *Injecting by removing the call entirely* fails on both and proves the check runs, not that it distinguishes the two ways the region can be unusable
**Inject:** internal/parser/trajectory.go:parseCurrent
**Pulled:** 2026-09-05 — rang, `two Actives reported OK`. The injection lives in `parseCurrent`, which `CheckActive` shares with the write path — Inject re-sited

## Test: a deviating stateless line is reported, a SPLIT parent is not
**Purpose:** a status-block line the reader could not read as a part and lists as deviating is an issue; a checkbox-less parent carrying no deviation is the format's own shape (R298)
**Input:** a status block with a keyed part, a checkbox-less `SPLIT` parent, and a bullet with neither checkbox nor head
**Expected:** exactly one finding, naming the file and the bullet's line; `HasIssues` true
**Refs:** crc-TrajectoryValidate.md, seq-validate-trajectory.md#2.8
**Code:** internal/validate/trajectory_test.go
**Alarm:** 14
**Fire alarm:** drop the `len(s.Deviations) == 0` filter in `checkStateless`. Red as `got 2 stateless findings, want 1` — the SPLIT parent reported, which is the false-positive direction that gets a check muted, and the direction the first live run took on this repository (five SPLIT and MOVED parents flagged)
**Inject:** internal/validate/trajectory.go:checkStateless
**Pulled:** 2026-09-05 — rang, `got 2 stateless findings, want 1` — the SPLIT parent flagged, exactly the shape the first live run on this repository produced

## Test: the two readers of the queue files must agree
**Purpose:** the line scan is the document reader's second opinion, and a disagreement is a finding listed first (R300)
**Input:** a done file whose second entry's body carries an unclosed backtick, so the document reader returns one entry while the line scan reads three headers; and a pair of files the two read identically
**Expected:** a finding naming `#1` and `#2` as read by the line scan and not returned by the document reader; `HasIssues` true; nothing for the agreeing pair
**Refs:** crc-TrajectoryValidate.md, seq-validate-trajectory.md#2.10
**Code:** internal/validate/trajectory_test.go
**Alarm:** 15
**Fire alarm:** drop the `t.checkReaderAgreement(traj, q)` call from `RunTrajectory`. Red as `the two readers disagreed about the done file and nothing said so` — and on this repository's ledger the report goes back to four confident orphans over a file the document reader saw 17 of 58 entries of
**Inject:** internal/validate/trajectory.go:RunTrajectory
**Pulled:** 2026-09-06 — rang again after item 71's commit landed past midnight and staled it, `the two readers disagreed about the done file and nothing said so`; restore byte-clean by copy. Previously 2026-09-05
