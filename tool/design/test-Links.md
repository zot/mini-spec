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
**Purpose:** R488, R458, R461
**Input:** a repository with `carves/a.md` (one tracked link) and `carves/done/b.md` (a missing link) both tracked, an untracked `notes.md` with a missing link, a tracked `x/testdata/s.md` fixture with a missing link, a tracked `gone.md` deleted on disk with the deletion unstaged, and no argument; then `--all`; then `carves/done/b.md` named explicitly
**Expected:** the population is every tracked markdown file (R488): the done carve's missing link is listed and exit is 1, `notes.md` is never read, nor is the fixture (R492), nor the deleted file; `--all` lists the tracked link; naming the done carve lists its missing link with its class and the error mark, and exits 1 (`--json` was checked by hand on 2026-09-15, not by this test)
**Refs:** crc-Links.md, seq-links.md#2.1
**Code:** internal/cli/cli_links_test.go
**Alarm:** 3
**Fire alarm:** walk the tree for every `*.md` instead of asking the index. Red: the untracked `notes.md` is read and its missing link listed.
**Inject:** internal/query/links.go:defaultPopulation
**Pulled:** 2026-09-15 — re-pulled 2026-09-15 after `#90` made the population every tracked markdown file: with a tree walk in place of the index, the untracked `notes.md` and the fixture `x/testdata/s.md` were both read and their missing links listed, `4 links in 5 files … missing 3`; restore checksummed clean. *Earlier —* 2026-09-15 — rang on the first shape, walking `carves/` recursively: `1 links in 1 files` where the done carve had been read *Earlier —* 2026-09-15 — rang, by hand, with the population walked recursively: the done carve's missing link listed and `2 links in 2 files … missing 1` where the summary wants one file; restore checksummed clean
