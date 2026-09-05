# Sequence: the carve census

```
1. CLI -> CLI: runQuery("carves", args)          // before any design root is resolved  R213
   1.1 CLI -> CLI: parseFlagsAnywhere(--open, --json)
   1.2 CLI -> Project: RepoRootFrom(cwd)
       Project --> CLI: repoRoot
   1.3 CLI -> Carve: ScanCarves(repoRoot)
       1.3.1 Carve -> Carve: locate carves/ and .carves/
             both absent -> ErrNoCarveDirs        // R214: no answer, not an empty census
             carves/done/ is not entered          // R208
       1.3.2 loop each *.md directly in a present directory
             1.3.2.1 Carve -> Carve: ReadCarve(path, rel)                 // diagram 2
             end
       Carve --> CLI: CarveScan{Dirs, Carves}
   1.4 CLI -> CLI: printCarves(scan, open)  or  carveReport(scan, open) under --json   // R215
       1.4.1 one line per carve: path, open, landed, then non-conforming and stateless
             counts when non-zero                                          // R207, R217
       1.4.2 a carve with no status block prints that instead of counts     // R211
       1.4.3 each part that is listed: key, queue label, title; a deviation per line beneath
             listed = any deviation, or --open and open                    // R212
       1.4.4 each stateless line: `(stateless)  L<n>  <reason>: <text>`, under --open, or
             always when it deviates                                       // R216
       1.4.5 the census, zeros included                                    // R218
```

# Sequence 2: ReadCarve

```
2. Carve -> os: ReadFile(path)
   2.1 Carve -> minispecsdom: ParseCarve(src)
       minispecsdom --> Carve: *Carve                // status region on heading nodes,
                                                     // fences invisible, depth from the bullet  R209, R210
   2.2 Carve -> minispecsdom: HasStatus(), Parts(), Stateless()
   2.3 loop each Part: wrap with State, QueueID, Title, Line()
   2.4 loop each stateless PartLine: line from its offset in Doc().Source(), head text,
       Deviations()                                                            // R216, R219
   Carve --> caller: Carve{Path, HasStatus, Parts, Stateless}
```

# Sequence 3: a marker write

```
3. caller -> Carve: SetMarker(path, key, verb, attribution)   // or SetPartLanded
   3.1 Carve -> os: ReadFile(path)
   3.2 Carve -> minispecsdom: ParseCarve(src).SetMarker(key, verb, attribution)   // or Land
       refused -> the reader's error, file untouched     // R220: DeviationError, ErrReopen,
                                                          //       ErrLanded, ErrNoPart
   3.3 Carve -> minispecsdom: Render()
   3.4 Carve -> os: write to <path>.tmp-*, then rename over path
```
