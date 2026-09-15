# Test Design: Links
**Source:** crc-Links.md

## Test: every class, in a temporary repository
**Purpose:** R456, R457, R459
**Input:** a `git init` tree with a tracked file, an untracked file, an ignored file (`.gitignore`), a tracked directory, a citing document linking to each plus a missing path, `../escape.md`, `/abs.md`, `https://x`, and `#frag`
**Expected:** one link per class — `tracked` for the file and the directory, `untracked`, `ignored`, `missing`, `outside` twice, `external`, `local`; `Errors()` true; the count names every class
**Refs:** crc-Links.md, seq-links.md#2.3
**Code:** internal/query/links_test.go
**Alarm:** 1
**Fire alarm:** classify before resolving so `../escape.md` is asked of the disk. Red: `outside` reads `missing` (or `tracked`, when the fixture's parent holds the file).
**Inject:** internal/query/links.go:classify
**Pulled:** 2026-09-15 — rang, by hand, with the root check removed: `[e](../../escape.md): got untracked, want outside` — and `[p](../private/x.md#frag): got untracked, want ignored` with it, because the outside path entered the `check-ignore` batch and git refused the whole batch, so the ordering constraint guards the ignored class too; restore checksummed clean

## Test: no git is a refusal
**Purpose:** R460
**Input:** a plain directory with one citing document
**Expected:** an error naming the reason; no report
**Refs:** crc-Links.md, seq-links.md#2.2
**Code:** internal/query/links_test.go
**Alarm:** 2
**Fire alarm:** return an empty report instead of the error. Red: the call succeeds with zero links.
**Inject:** internal/query/links.go:CheckLinks
**Pulled:** 2026-09-15 — rang, by hand, at `CheckLinks` — the site the 2026-09-15 pull was actually made at; the design had named it `Check` and the census could not resolve that: `want ErrNoGitTree and no report, got <nil>, &{Files:[] Links:[] Counts:map[]}`; restore checksummed clean
*Pulled at `internal/query/links.go:Check` — 2026-09-15 — rang, by hand: `want ErrNoGitTree and no report, got <nil>, &{Files:[] Links:[] Counts:map[]}`; restore checksummed clean — and the site has since moved, so this is history rather than a record.*
## Test: the default population is the live carves, and the CLI's exit status follows the errors
**Purpose:** R455, R458, R461
**Input:** a repository with `carves/a.md` (one tracked link), `carves/done/b.md` (a missing link) and no argument; then `--all`; then `carves/done/b.md` named explicitly
**Expected:** no lines and a count with `tracked 1`, exit 0; `--all` lists the tracked link; naming the done carve lists its missing link with its class and the error mark, and exits 1 (`--json` was checked by hand on 2026-09-15, not by this test)
**Refs:** crc-Links.md, seq-links.md#2.1
**Code:** internal/cli/cli_links_test.go
**Alarm:** 3
**Fire alarm:** walk `carves/` recursively for the default population. Red: the first run exits 1 on the done carve's missing link.
**Inject:** internal/query/links.go:defaultPopulation
**Pulled:** 2026-09-15 — rang, by hand, with the population walked recursively: the done carve's missing link listed and `2 links in 2 files … missing 1` where the summary wants one file; restore checksummed clean
