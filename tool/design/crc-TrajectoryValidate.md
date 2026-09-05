# TrajectoryValidate
**Requirements:** R284, R285, R286, R287, R288, R290, R291, R292, R293, R294, R295, R296, R297, R298, R299, R289, R300

Checks the trajectory layer for consistency: the queue files at the repository root against
the carves that point at them. Read-only, like [Validate](crc-Validate.md), and separate
from it for a reason that is structural rather than stylistic.

**Repository-scoped, and that is why it is not folded into `Validate`.** `Validate` resolves
a *design root* and runs once per one; a repository may hold several, as this project holds
`tool/` and `example/`. The trajectory layer sits above all of them, so folding this in
would run it once per design root and report the same drift twice — the two-roots theme in
[specs/index.md](../specs/index.md) violated at the seam that theme was written about. It
therefore takes a repository root and never a Project, the same shape as `NextItemID` and
`Carves`. R285

## Knows
- TrajectoryIssues: the findings, bucketed by class so each is countable and each names its
  own repair. Buckets rather than a flat list because the classes have different repairs:
  a dangling citation is edited in the carve, an unmigrated ledger entry in the done file
- readable: what the scan could *see*, kept beside what it found, in **three** dimensions. A
  check that reports clean over what it never read is the failure this layer exists to
  prevent. Citations: unkeyed part lines sit where the position rule cannot read them (R293).
  Entries: lines the dependency's readers did not recognize (R297). And what the document
  reader returned against what the line scan read (R300) — the dimension the other two cannot
  supply, since a swallowed tail is neither unrecognized nor unkeyed, it is simply absent

## Does
- RunTrajectory(repoRoot): scan the queue files and every carve, run the checks, return
  TrajectoryIssues. Nothing is written (R284)
- checkReaderAgreement(): **the second opinion, run first** (R300). `ScanTrajectory`'s line scan
  and the dependency's document readers read the same two files; every item ID one saw and the
  other did not is a finding, per file. Measured 2026-09-05 on this repository the day it was
  written: 58 IDs by line, 17 entries by document, nothing unread — one double-backtick span
  absorbed 41 entries (`O20`) and four landed parts read as orphans downstream
- checkStateless(): every status-block line the reader could not read as a part **and** lists
  as deviating (R298), from the dependency's `Stateless()` — a checkbox-less `SPLIT` parent
  carries no deviation and is not one. *The August tree's independent flat scan of the status
  block is not carried; gap `O18` names its home*
- absent(): a repository with neither queue file **and** no carve directory has nothing
  that could be inconsistent, so the result is clean and says so. Distinct from *could not
  check*, and the two must never collapse — the same distinction `ScanTrajectory` draws
  with ErrNoTrajectoryFiles, in the other direction (R286)
- checkCarveToQueue(): every `#N` a carve's status block cites resolves to an entry in the
  pending or done file (R287)
- checkQueueToCarve(): every queue entry whose `Source:` names a carve appears in that
  carve's status block under the key it claims. The direction R287 cannot see: an item can
  otherwise complete against a part its carve never recorded (R288)
- checkDuplicateIDs(): one number held by two entries across pending and done makes every
  pointer to it ambiguous, silently (R290)
- checkOrphans(): a part marked landed against a queue ID with no done entry, and a done
  entry naming a part no carve records (R291)
- checkLedgerConformance(): a done entry header with no identifier slot is named
  **unmigrated** rather than skipped (R292)
- checkLineAgreement(): a status line's checkbox, strikethrough and marker must agree, the
  checkbox being authoritative. Currently finds nothing, which is the point — it is a
  sentry over a corpus normalised by hand (R294)
- checkNumbering(): every number from 1 to the maximum assigned should appear in a readable
  entry, since an ID is assigned at creation. A deliberately abandoned ID is the expected
  exception, and the format says gaps are expected — but *expected* is a claim worth testing:
  measured 2026-08-16, ark has 17 gaps and **16 are mentioned in its own queue files**, so
  they are losses. When entries went unread the report says so beside them, because that
  very likely explains them (R295)
- FormatText(): issues only, each class naming its migration target rather than stating a
  complaint, and a closing note on what could not be read — **both** the citations the
  position rule could not reach (R293) and the entry-like lines the shape did not recognize
  (R297)


- **`CURRENT.md` carries exactly one `## Active`, and this is where that is noticed** (R296). The
  heading bounds the region `pending finish` clears; `activeRange` has always refused both the
  missing one and a duplicated one, *at write time only*, so a corrupted file sat clean until the
  next completion met it — measured 2026-08-19, green **twice in one session** over a file the
  verb had just damaged
  - *Why a check rather than care:* the trajectory files are gitignored under
    `track: private-trajectory`, so **git cannot diff them**, and the backup slot holds **one**
    level of undo which the next operation spends. A check running between operations is the only
    thing standing where a diff would normally stand
  - **Reported before the reference-level findings**, because a file whose own shape is wrong makes
    every statement about the references inside it a claim about a document nobody can trust. A
    missing trajectory layer is not this finding — absence is already answered, and two checks
    giving one fact different words is worse than either alone
## Collaborators
- Carve: the carve status blocks, their parts and their stateless lines, already ingested by
  position through the dependency's reader
- Trajectory: the queue files twice — `ScanTrajectory`'s line-scanned IDs and `ScanQueue`'s
  entries through the dependency's readers — and `CheckActive` for the current file's shape
- Project: to resolve the repository root, and nothing else

## Notes
**Ingest by position is inherited, not re-implemented.** Citations reach this card already
extracted by [Carve](crc-Carve.md), which reads a `#N` from a part line's marker and from
nowhere else. That matters because the alternative failed in measurement: an extractor
sweeping `**VERB (…)**` over a whole status block read a backquoted prose example as a live
citation and reported a dangling `#121` in a repository that never had one. R289 states the
rule; this card is a consumer of it rather than a second implementation.

**Conformance is named, never complained about.** A carve on a superseded key scheme is
unmigrated rather than wrong, and an agent handed a target can act where one handed a
complaint cannot — the same crank-handle shape the Source diagnostics already use.

## Sequences
- seq-validate-trajectory.md
