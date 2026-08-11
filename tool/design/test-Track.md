# Test Design: Track
**Source:** crc-Track.md

Every case calls `Verify` against a **fake `Git`** — a small double stating whether
the tree is a working tree and which paths it ignores. No repository is created, no
`git` is invoked, and no filesystem is consulted for the ignore facts. The consistency
check is pure decision logic over two inputs, so the expensive-to-reach paths are
avoided entirely and there is no excuse for leaving it hand-checked.

The fake is the reason `Git` is a separate card: the seam is what makes this table
possible.

## Test: each value parses, and nothing else does
**Purpose:** validates the closed set — an unrecognised value is malformed rather than
a fourth behavior
**Input:** `none`, `private-trajectory`, `all`, plus `None`, `private`, `""`, `yes`
**Expected:** the three parse; every other input is rejected as malformed
**Refs:** crc-Track.md — R131

## Test: `none` agrees only when git is absent
**Purpose:** validates the git-presence half of the check in the direction that catches
a project gaining git after `init`
**Input:** `track: none` against a fake reporting a working tree, and against one
reporting none
**Expected:** working tree present → a mismatch naming git presence; absent → no
mismatch
**Refs:** seq-bootstrap.md#1.7 — R132, R147

## Test: `private-trajectory` and `all` agree only when git is present
**Purpose:** the mirror direction — a declared git style in a tree with no repository
**Input:** each value against a fake reporting no working tree
**Expected:** both produce a mismatch naming git presence
**Refs:** seq-bootstrap.md#1.7 — R147

## Test: `private-trajectory` requires the trajectory files ignored
**Purpose:** validates the ignore-state half, which fails independently of git presence
**Input:** `private-trajectory`, fake reports a working tree, ignores nothing
**Expected:** a mismatch naming `PENDING.md`, `CURRENT.md` and `DONE.md`
**Refs:** seq-bootstrap.md#1.9 — R133, R148

## Test: `all` requires the trajectory files *not* ignored
**Purpose:** the opposite expectation from the same fact — the asymmetry that makes
repair symmetric
**Input:** `all`, fake reports a working tree and ignores all three trajectory files
**Expected:** a mismatch naming them as ignored when they should not be
**Refs:** seq-bootstrap.md#1.9 — R132, R148

## Test: `.carves/` must be ignored under `all` too
**Purpose:** the one path whose expectation does not follow the queue's — a directory
whose whole purpose is privacy does not become public because the queue did
**Input:** `all`, fake reports a working tree, trajectory files not ignored (correct),
`.carves/` present and not ignored
**Expected:** a mismatch naming `.carves/` only
**Refs:** crc-Track.md — R134

## Test: `.carves/` is not required when it does not exist
**Purpose:** validates that the directory is checked on demand rather than mandated
**Input:** `private-trajectory`, correct trajectory ignores, no `.carves/` on disk
**Expected:** no mismatch
**Refs:** crc-Track.md — R133

## Test: both facts can disagree at once, and both are reported
**Purpose:** validates that the check collects disagreements rather than returning the
first — a repair message naming one problem when there are two sends the user round
twice
**Input:** `none`, fake reports a working tree *and* ignores the trajectory files
**Expected:** the report names git presence; the run exits non-zero
**Refs:** seq-bootstrap.md#1.10 — R149

## Test: the mismatch report names the repair command
**Purpose:** the crank handle — a gripe that does not say what to run is a nag
**Input:** any mismatching combination
**Expected:** the message contains `init --track-` and `--repair`, and says to confirm
the value with the user
**Refs:** seq-bootstrap.md#1.11 — R149, R146

## Test: no git means silence, not a mismatch
**Purpose:** validates that a tree with no repository is a supported shape rather than
a misconfiguration
**Input:** `track: none`, fake reports no working tree, nothing ignored
**Expected:** no mismatch and no preference output
**Refs:** crc-Track.md — R151
