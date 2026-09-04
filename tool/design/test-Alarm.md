# Test Design: Alarm
**Source:** crc-Alarm.md

Every case runs `Assess` against a **fake Git** stating what changed and when, so no
repository is created and no `git` runs. The assessment is pure decision logic over
two inputs, which is what makes the whole state table cheap to pin.

## Test: each state is reachable, and only by its own shape
**Purpose:** validates the closed set — an alarm resolves to exactly one state, and
the four shapes that produce them are distinct
**Input:** an alarm with sites+pull and no change; with sites+pull and a later change;
with sites and no pull; with no sites
**Expected:** `verified`, `stale`, `unrecorded`, `unanchored` respectively
**Refs:** crc-Alarm.md — R185

## Test: a same-day change is verified, and the next day is stale
**Purpose:** validates R181 — the boundary this feature got wrong first. Same-day must
be *verified*: the normal workflow commits the fix, the test and the pull together, so
a freshly recorded alarm always shares a date with the code it proves. The inclusive
version marked two of ark's three verified alarms stale the day they were written, and
a check that fires on arrival is ignored
**Input:** an alarm pulled `2026-08-05` and one pulled `2026-08-04`, against a fake
reporting the site changed `2026-08-05`
**Expected:** `verified` and `stale` respectively
**Fire alarm:** make the comparison inclusive (`!when.Before(a.Pulled)`) — the literal
first version — and confirm the same-day case goes red while the later one stays green.
The two differ only in the boundary, which is the whole property
**Inject:** internal/alarm/alarm.go:assessOne
**Pulled:** 2026-08-13 — rang: same-day reported `stale`, restore byte-clean
**Refs:** crc-Alarm.md — R181

## Test: a site git cannot resolve is reported, not skipped
**Purpose:** validates R182 — a missing symbol is the anchor rotting, and silence
there would be the tool declining to report its own mechanism failing
**Input:** an alarm whose site names a symbol the fake reports as unfindable
**Expected:** `unresolvable`, and the site is named
**Fire alarm:** treat the lookup error as "unchanged" and confirm the alarm reports
`verified` instead — the exact silent-pass this field exists to prevent
**Inject:** internal/alarm/alarm.go:assessOne
**Pulled:** 2026-08-13 — rang: reported `verified` where the anchor had rotted, which is the exact silent pass this field exists to prevent; restore byte-clean
**Refs:** crc-Alarm.md — R182

## Test: no git yields `unchecked`, never `verified`
**Purpose:** validates R187 and R184 — a check that could not look must not return a
clean result
**Input:** an alarm with sites and a pull date; fake Git reporting no repository
**Expected:** `unchecked`; `unrecorded` and `unanchored` still resolve, needing no git
**Fire alarm:** return `verified` when git is absent and confirm this goes red
**Inject:** internal/alarm/alarm.go:assessOne
**Pulled:** 2026-08-13 — rang: reported `verified` with no repository, restore byte-clean
**Refs:** crc-Alarm.md — R184, R187

## Test: the census counts every alarm exactly once
**Purpose:** the four counts must partition the input — a census that double-counts or
drops is worse than none, since it reads as a measurement
**Input:** a mixed set of all four states
**Expected:** the counts sum to the number of alarms
**Refs:** crc-Alarm.md — R185

## Test: `Voided` carries only the closable states
**Purpose:** validates R183 — `validate` reports stale and unresolvable and nothing
else. Unrecorded and unanchored stay non-zero for months while a project adopts the
convention, and a permanent non-zero line is the nag this project distinguishes from a
gripe you can discharge
**Input:** one alarm of each of the four states
**Expected:** two assessments, both closable
**Fire alarm:** widen the filter to everything that is not `verified` and confirm the
count assertion goes red at 4
**Inject:** internal/alarm/alarm.go:Voided
**Pulled:** 2026-08-13 — rang, restore byte-clean
**Refs:** crc-Alarm.md — R183

## Test: an uncheckable site does not mask a stale one
**Purpose:** validates R179 — a site git cannot answer for is remembered rather than
returned at once, because a finding you have is worth more than one you could not
reach. Only with nothing worse to report does the alarm read `unchecked`
**Input:** an alarm whose first site has no history and whose second is definitely
stale; and the same first site alone
**Expected:** `stale` for the pair, `unchecked` for the lone one
**Fire alarm:** return `Unchecked` on the first erroring site — the pre-fix
short-circuit — and confirm the pair reports `unchecked` instead of `stale`
**Inject:** internal/alarm/alarm.go:assessOne
**Pulled:** 2026-08-13 — rang, restore byte-clean
**Refs:** crc-Alarm.md — R179

## Test: `--unverified` narrows the list and leaves the census whole
**Purpose:** validates R199 — the filtered form must be the full census minus the
repetitions of *nothing to do here* and minus nothing else. A census filtered along with
its list answers *how many are wrong* and silently drops *out of how many*, which is the
`| tail -2` workaround this flag exists to replace, one level in
**Input:** assessments covering `verified`, `stale`, `unrecorded` and `unanchored`
**Expected:** the selection holds every state but `verified`; the counts computed for the
closing line still total the whole input
**Fire alarm:** pass `selected` to `alarmCensusLine` instead of `assessments`, and confirm
the total drops to the number of unverified alarms while the listing stays byte-identical —
the failure is invisible in the list, which is the point
**Inject:** internal/cli/cli.go:queryAlarms
**Pulled:** 2026-09-04 — rang (pulled by a delegated `alarm-puller` at `2c90e41`, evidence read by hand, restore byte-clean): `TestUnverifiedNarrowsTheListNotTheCensus` reported `1 alarms: 1 unrecorded` over a population of 2 — the listing byte-identical, the total wrong. The puller had to move the census computation below the selection for the injection to compile, and said so
**Refs:** crc-Query.md — R199

## Test: the selection is applied before the output form is chosen
**Purpose:** validates R204 — filtering the text branch only renders everything under
`--json` while every markdown test stays green. `query gaps` met this exact silent ignore
one level in on `old-sdom` (its R326, R327 — not yet reclaimed here), so this is the same
defect guarded in a second place rather than a new one
**Input:** `query alarms --unverified --json` over a design holding verified and stale
alarms
**Expected:** the JSON array holds only the unverified assessments
**Fire alarm:** move the `--unverified` filter inside the text branch and confirm the JSON
case goes red while the text case stays green — one green branch beside one red is the
signature of this defect
**Inject:** internal/cli/cli.go:queryAlarms
**Pulled:** 2026-09-04 — rang (pulled by a delegated `alarm-puller` at `2c90e41`, evidence read by hand, restore byte-clean): `TestAlarmsJSONHonoursSelection` rendered 2 assessments where 1 was selected, and `TestBriefRidesOnTheAssessmentUnderJSON` likewise; the text branch stayed green — one green beside one red, as predicted. The puller left the brief branch unfiltered along with the JSON one and noted the choice
**Refs:** crc-CLI.md, crc-Query.md — R204

## Test: a brief quotes the fire-alarm prose whole
**Purpose:** validates R200 — the prose *is* the injection, so a paraphrase or a first
line is a different injection. The parser already folds continuation lines into one body
(R178); this pins that the brief carries all of what it folded
**Input:** an assessment whose alarm prose spans three wrapped lines
**Expected:** every one of the three lines appears in the brief
**Fire alarm:** emit only the first line of `Prose` and confirm this goes red. The
truncated brief still reads as a complete, well-formed brief, which is why the assertion
is on the last line rather than on the shape
**Inject:** internal/alarm/brief.go:Brief
**Pulled:** 2026-09-04 — rang (pulled by a delegated `alarm-puller` at `2c90e41`, evidence read by hand, restore byte-clean): `TestBriefQuotesProseWhole` named the two dropped lines; every other test stayed green, since the truncated brief is well-formed
**Refs:** crc-Alarm.md — R200

## Test: a brief states an absent part rather than dropping its line
**Purpose:** validates R203 — an `unanchored` alarm has no sites and an unmapped document
has no test files. A missing line reads as an oversight in the generator; a stated absence
reads as what it is, and keeps the number of briefs equal to the number of selected alarms
**Input:** an `unanchored` assessment with no sites, and a `stale` one whose document the
Artifacts manifest maps to nothing
**Expected:** both briefs carry a *Sites:* and a *Tests:* line; each names its absence
**Fire alarm:** return the empty string when the list is empty, so the line collapses to a
bare label, and confirm this goes red. The brief that results is shorter and entirely
well-formed, so nothing but this assertion notices
**Inject:** internal/alarm/brief.go:sitesLine, internal/alarm/brief.go:filesLine
**Pulled:** 2026-09-04 — rang (pulled by a delegated `alarm-puller` at `2c90e41`, evidence read by hand, restore byte-clean): `TestBriefStatesAnAbsentPart` failed both assertions and `TestAlarmBriefsCoverEverySelectedAlarm` failed on the unmapped document — two suites red from one edit, both on the absence rather than the shape
**Refs:** crc-Alarm.md — R203

## Test: a brief names the design root relative to the repository root
**Purpose:** validates R200 — a delegated puller runs in its own git worktree, where this
checkout's absolute path resolves to nothing. An absolute path is not a cosmetic flaw
here: it sends the agent to the *original* tree, where its injection would corrupt the
working copy the worktree exists to protect
**Input:** a brief minted with a design root of `tool`
**Expected:** the brief names `tool`, and contains no path beginning with the filesystem
root
**Fire alarm:** return `p.RootPath` from `designRootRel` instead of the relative form, and
confirm this goes red
**Inject:** internal/cli/cli.go:designRootRel
**Pulled:** 2026-09-04 — rang (pulled by a delegated `alarm-puller` at `2c90e41`, evidence read by hand, restore byte-clean): `TestBriefResolvesTestFilesThroughTheManifest` failed on the absolute `Design root:` line
**Refs:** crc-Query.md — R200

## Test: a brief carries the contract and no protocol
**Purpose:** validates R202 and R201 — *evidence, never a verdict* is the sentence that
decides whether delegating was worth doing, while the protocol is invariant and belongs to
the agent definition. And a brief that names a command invites the failure `SKILL.md`
names, an injection that breaks the build teaching nothing because the test never ran
**Input:** any assessment
**Expected:** the brief names the five things to return and says no verdict; it does not
contain a runnable test command
**Refs:** crc-Alarm.md — R201, R202
