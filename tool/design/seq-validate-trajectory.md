# Sequence: validating the trajectory layer

Two numbered diagrams: the scan and dispatch, and the checks it runs. Numbered so code can
pin to a step.

# Sequence 1: the run

```
1. CLI -> CLI: minispec validate trajectory

   1.1 CLI -> Project: RepoRoot()
       Dispatched before any design root is resolved. The layer is repository-scoped
       and a repository may hold several design roots.        // R285

   1.2 CLI -> TrajectoryValidate: RunTrajectory(repoRoot)

       1.2.1 TrajectoryValidate -> Trajectory: ScanTrajectory(repoRoot), then ScanQueue(repoRoot)
             Trajectory --> TrajectoryValidate: the line-scanned IDs per file; then the
             entries through the dependency's readers — pending entries with Source
             pointers, done entries with slot IDs and part pointers — and what each reader
             left unread                                             // R287, R297
             Both files are read through the document model. An entry is a `##` heading
             node, so a queue heading quoted inside a fence is not an entry and cannot
             claim an identifier.                                  // R260
             An entry's title is its heading's first emphasis run, closed at that run's
             own end with code spans suppressed; its `Source:` is read from that entry's
             own region, never from the most recent matching line. // R255, R260

       1.2.2 TrajectoryValidate -> Carve: ScanCarves(repoRoot)
             Carve --> TrajectoryValidate: every carve, its parts, and each part's
             citation — already ingested by position                // R289

       1.2.3 TrajectoryValidate -> TrajectoryValidate: absent?
             neither queue file and no carve directory -> clean, and say so
             Nothing that could be inconsistent is not a failure, and it is not
             "could not check" either.                              // R286

       1.2.4 TrajectoryValidate -> TrajectoryValidate: run the checks   // diagram 2

   1.3 TrajectoryValidate --> CLI: TrajectoryIssues

   1.4 CLI -> CLI: FormatText or --json
       1.4.1 issues only, each naming its migration target rather than a complaint
       1.4.2 a closing note on what could not be read              // R293
       1.4.3 exit 0 when consistent, 1 when issues were found      // R299
```

# Sequence 2: the checks

```
2. TrajectoryValidate -> TrajectoryValidate: check(entries, carves)

   2.1 checkCarveToQueue
       every `#N` cited by a carve part resolves to a pending or done entry   // R287
       unresolved -> dangling citation, named with the carve and the part key

   2.2 checkQueueToCarve
       every queue entry whose Source: names a carve appears in that carve's
       status block under the key it claims                                   // R288
       This is the direction 2.1 cannot see.

   2.3 checkDuplicateIDs
       one number held by two entries across pending and done                 // R290

   2.4 checkOrphans
       2.4.1 a part marked landed against a queue ID with no done entry
       2.4.2 a done entry naming a part no carve records                      // R291

   2.5 checkLedgerConformance
       a done header with no identifier slot is unmigrated, not skipped       // R292
       nothing is also the right answer for an entry that discharged no ID,
       so the two are indistinguishable unless the shape is checked

   2.6 checkLineAgreement
       checkbox, strikethrough and marker must agree; the checkbox wins       // R294
       finds nothing today — a sentry over a corpus normalised by hand

   2.7 checkNumbering
       every number from 1 to the maximum assigned appears in a readable entry   // R295
       an abandoned ID is the expected exception; measured, ark's 17 gaps are not

   2.8 checkStateless
       a status-block line the reader could not read as a part and lists as
       deviating; a checkbox-less SPLIT parent carries no deviation            // R298

   2.9 coverage: what the readers could not read, per file — the queue files,      // R297, R302
       the current file, and each carve
       the honest half, and 2.7 is its symptom — a shape-based check cannot
       answer for a line it never recognized as an entry

   2.10 checkReaderAgreement — run before 2.1, numbered last because steps are
        appended, never renumbered
        every ID the line scan read and the document reader returned no entry
        for, and the reverse, per file; listed first in the report            // R300

   2.12 checkLinks — query.CheckLinks over the live and done carves; errors bucketed as
        findings, untracked as a note, no git tree as a note saying the links went unclassified
   2.11 TrajectoryValidate --> TrajectoryValidate: TrajectoryIssues, bucketed by class
```
