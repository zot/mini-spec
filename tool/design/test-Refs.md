# Test Design: Refs
**Source:** crc-Refs.md

## Test: the owned documents are tracked plus sited, scratch excluded
**Purpose:** R497
**Input:** a repository tracking `carves/x.md`, `carves/other.md` and `.gitignore`, with an untracked `.carves/private.md`, ignored `DONE.md` and `PENDING.md`, and an ignored `.scratch/notes.md`
**Expected:** the population is exactly the two carves, the private carve and the two ledgers; scratch is absent
**Refs:** crc-Refs.md, seq-links.md#5.3
**Code:** internal/query/refs_test.go
**Alarm:** 1
**Fire alarm:** take the tracked documents alone. Red: the private carve and the ledgers vanish from the population.
**Inject:** internal/query/refs.go:OwnedDocuments
**Pulled:** 2026-09-16 — rang, by hand after the simplification pass, with the tracked documents alone: `population: [carves/other.md carves/x.md]`, the private carve and the ledgers gone; restore checksummed clean

## Test: every kind resolves by its own rule, and --to inverts
**Purpose:** R495, R496
**Input:** the same root; `carves/x.md` links `other.md` and points `other.md#1`; the private carve links and points at `../carves/x.md`; `DONE.md` carries a Part pointer to carves/x.md#1
**Expected:** `--to carves/x.md` lists exactly the private carve's link and pointer and the ledger's pointer, kinds named; the refs in `carves/x.md` are one link and one pointer both resolving to `carves/other.md`
**Refs:** crc-Refs.md, seq-links.md#5.1
**Code:** internal/query/refs_test.go
**Alarm:** 2
**Fire alarm:** resolve every pointer from the repository root, ignoring the citing directory. Red: the private carve's `../carves/x.md#1` misses and drops out of `--to`.
**Inject:** internal/query/refs.go:resolvePointer
**Pulled:** 2026-09-16 — rang, by hand after the simplification pass, at `resolvePointer` with every pointer resolved from the root: the private carve's `../carves/x.md#1` dropped out of `--to`, two references where three are expected; restore checksummed clean

## Test: the CLI lists the inventory and inverts it, needing no design root
**Purpose:** R495, R496
**Input:** a repository with a carve, a linked sibling and a ledger pointer, run from the root with `--to`
**Expected:** exit 0; the text names `DONE.md:5`, the pointer, its kind and the resolved path, and closes with the count
**Refs:** crc-Refs.md, seq-links.md#5.4
**Code:** internal/cli/cli_refs_test.go
**Alarm:** 3
**Fire alarm:** drop the pointer kind from `RefsIn` so only links are listed. Red: the ledger's pointer is missing and the count reads 0.
**Inject:** internal/query/refs.go:RefsIn
**Pulled:** 2026-09-16 — rang, by hand after the simplification pass, with pointers labelled as links: the inventory read `` `carves/x.md#1`  link ``; restore checksummed clean
