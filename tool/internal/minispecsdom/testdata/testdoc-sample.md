# Test Design: Sample
**Source:** crc-Sample.md

## Test: a same-day change is verified, and the next day is stale
**Purpose:** validates R181 — the boundary this feature got wrong first. Same-day must
be *verified*: the normal workflow commits the fix, the test and the pull together
**Input:** an alarm pulled `2026-08-05` and one pulled `2026-08-04`, against a fake
reporting the site changed `2026-08-05`
**Expected:** `verified` and `stale` respectively
**Fire alarm:** make the comparison inclusive (`!when.Before(a.Pulled)`) — the literal
first version — and confirm the same-day case goes red while the later one stays green.
The two differ only in the boundary, which is the whole property
**Inject:** internal/alarm/alarm.go:assessOne
**Pulled:** 2026-09-06 — rang again after `assessOne` gained the ambiguous-site case: `same-day = "stale", want verified`; restore byte-clean. First pulled 2026-08-13, same signature
**Refs:** crc-Alarm.md — R181
**Code:** internal/alarm/alarm_test.go
**Alarm:** 3

## Test: pending entries read through **the dependency**
**Purpose:** validates R353 — the adapter reads ID, source document, source key and kind
**Input:** a pending file with a part-sourced and a gap-sourced entry
**Expected:** two entries with the expected IDs, keys and kinds
**Refs:** crc-Trajectory.md — R353
**Code:** internal/parser/trajectory_test.go

## Test: the pre-`track` refusal asks the intent and stops
**Purpose:** validates R175 and R176
**Input:** `preTrackMessage`
**Expected:** carries the stop and the private-or-ships question
**Fire alarm:** swap the message for the no-configuration refusal's. Red: the
private-or-ships question is missing
**Inject:** internal/cli/bootstrap.go:preTrackMessage, internal/cli/bootstrap.go:routeRefusal
**Refs:** seq-bootstrap.md#1.6.1 — R175, R176

## Test: a document about the format quotes its own fields
**Purpose:** validates that a fenced example is body
**Input:** the block below, which is an example and not a record:

```markdown
## Test: quoted
**Fire alarm:** quoted
**Inject:** quoted.go:Quoted
**Alarm:** 99
```

**Expected:** no entry named `quoted`, and no alarm numbered 99
**Fire alarm:** read fields with a line scan that ignores fences. Red: alarm 99 appears
**Inject:** internal/minispecsdom/testdoc.go:TestDoc.scan
**Alarm:** 1

## Notes

A level-2 heading that is not a test.
