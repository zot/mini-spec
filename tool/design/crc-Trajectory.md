# Trajectory
**Requirements:** R190, R194, R195, R197, R240, R242, R244, R249, R250, R251, R252, R256, R257, R259, R260, R263, R265, R266, R267, R270, R287, R292, R296, R297

Reads and writes the trajectory files at the repository root, as thin path-taking adapters
over the dependency's `minispecsdom.Pending`, `Current` and `Done` readers. The format is **owned by the skill**
— `trajectory-format.md` is normative for every shape — so this card names which shapes
it consumes and never redefines them.

The first component in the tool to read trajectory content at all. Until now the files
were only ever a root marker and an ignore-state question, both answerable without
opening them.

## Knows
- TrajectoryScan: what each queue file contributed — its name, whether it was
  **present**, and the IDs found in it. Presence is kept apart from the ID count because
  they answer different questions, and collapsing them is how a broken regex comes to
  look like an empty queue

## Does
- PendingEntries(path): the pending file's entries through `minispecsdom.Pending` — ID, title,
  source document, source key and kind (part or gap), line — as `QueueEntry`. A missing file is
  no entries and no error, since the slot legitimately reads a side that has none (R240)
- Parts(), Gap() on a `QueueEntry`: what the entry discharges — one part when `Kind` is a part
  and both halves are present, one gap when it is a gap; an entry naming a document alone
  records nothing (R251, R276)
- ResolvePlace(pendingPath, place): an intent — `--last`, `--nth N`, `--after N` — to a
  1-based position among the entries the reader sees, refusing rather than clamping and
  naming the live IDs on a bad `--after`; `--next` is the orchestrator's to resolve first
  (R256, R259, R260)
- PlaceItem(pendingPath, entry, pos): the whole entry in the format's shape through
  `EntryText`, part or gap form by kind, inserted by `Pending.Place` beside the entry it
  precedes (R252, R260, R275)
- CompleteItem(pendingPath, donePath, id, header, body): `Pending.Remove` then
  `Done.Prepend`, the header and body in one write; pending side first so a failure leaves an
  ID in neither file rather than both (R244, R263, R270)
- ResetCurrent(path), SetActive(path, line, context): the `## Active` region through
  `Current.Reset` and `Current.SetActive`; the reader refuses a held region (`ErrOccupied`)
  and this card names the parking repair (R249, R265, R266, R267)
- ActiveInProgress(path): `Current.Occupied` — one boolean, for `--next` (R257)
- parseCurrent: the reader's two refusals — no `## Active`, more than one — with the repair
  named, since the shape is the skill's format and the tool's to name (R250)
- Every write goes through `editFile`: read, render, temp-file-and-rename, so a refusal
  leaves the file byte-identical (R220)
- ScanTrajectory(repoRoot): read both files and report them **per file** rather than
  merged, so the caller can say what each contributed (R197). A package function, not a
  method — it touches no design-root state, and a signature claiming otherwise is what
  made the command refuse to run in this repository
- MaxItemID(): the maximum across **both** files. Either alone collides — the pending
  file's maximum is too low right after items complete, the done file's while the
  highest IDs are still live (R190)
- Missing(), AnyPresent(): name the files that were absent, and distinguish *no files at
  all* — which has no answer, so ErrNoTrajectoryFiles — from *one file missing*, which
  has an answer worth qualifying (R194, R195)

## Shapes consumed
Named here so the coupling is visible; `trajectory-format.md` defines them.

- **pending file** — an item entry is a `##` heading opening with the number:
  `## 8. **<title>** …`
- **done file** — an entry's **header** leads with the identifiers it discharged, in the
  run between the date's em dash and the colon that opens the title:
  `` - **YYYY-MM-DD — #8 / R189–R198: <title>.** (`<commit>`) Part `<doc>#<key>`. ``
  The slot may hold a gap ID or a requirement range instead, or several separated by
  `/`, so every `#N` in it counts — and only what is in it does.

**Only an entry's header is scanned**, deliberately. Done entries are prose several
lines long and quote other items freely: measured 2026-08-16, **five** body lines in
ark's ledger would contribute a queue ID if the header were not required, most of them
older pending entries quoted verbatim inside a later entry's body. Letting one in is the
same class of error as reading one file instead of two — a plausible number that is
wrong.

## Collaborators
- minispecsdom (the dependency): owns every shape and every region; this card only
  hands it paths and takes back bytes
- Project: to resolve the repository root

## Sequences
- seq-query.md
- seq-queue-item.md
