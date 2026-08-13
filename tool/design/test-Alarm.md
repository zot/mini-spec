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
