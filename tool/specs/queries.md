# Query Commands

All queries read from design files and output results. No modifications.

## minispec query project

Show the project paths resolved by the tool: root, design directory, src
directory, and specs directory. Useful for verifying which project the tool
detected (especially when running from a subdirectory or when reported missing
paths look correct).

Output:
```
root:   /home/me/work/ark
design: /home/me/work/ark/design
src:    /home/me/work/ark/src
specs:  /home/me/work/ark/specs
```

JSON output uses keys `root`, `design`, `src`, `specs`.

## minispec query requirements

List all requirements from requirements.md.

Output: List of Rn with text and source spec.

## minispec query coverage

For each requirement, show which design files reference it.

Output:
```
R1: crc-Store.md, crc-View.md
R2: crc-Store.md
R3: (none)
```

## minispec query uncovered

List requirements that have no design file references.

Output: List of Rn identifiers.

## minispec query orphan-designs

List CRC cards that have no Requirements field or an empty one.

Output: List of file paths.

## minispec query artifacts

List all artifacts from design.md with their checkbox states.

Output:
```
crc-Store.md
  [x] src/store.ts
  [ ] src/store_test.ts
crc-View.md
  [x] src/view.ts
```

## minispec query gaps

List all items from the Gaps section of design.md.

Output:
```
[ ] S1: spec item not in requirements
[x] R1: resolved requirement gap
[ ] D1: design without code
```

## minispec query traceability [file]

Check if a code file has proper traceability comments.

Output: CRC and Seq references found, or "missing" indicator.

## minispec query traceability --all

Scan all code files listed in Artifacts and report traceability status.

## minispec query unindexed-specs

List per-feature specs (`specs/*.md`, non-recursive) that are not referenced
anywhere in the root spec index `specs/index.md`. The index file itself and
anything under `specs/migrations/` are excluded. Matching is by exact `.md`
token, so `search.md` is not treated as indexed merely because
`fuzzy-search.md` appears. When `specs/index.md` does not exist, every
per-feature spec is listed (nothing is indexed yet).

Output: relative spec paths, sorted; empty (exit 0) when every spec is indexed.
