# Project
**Requirements:** R32-35, R93, R118-125, R127-130, R135, R520-522, R527, R555, R556

Finds and loads a mini-spec project's configuration and design files.

## Knows
- rootPath: the **design root** — the directory containing `design/`. Named before
  the two roots were distinguished; the repository root is `RepoRoot`'s business
- designDir: path to design/ directory
- srcDir: path to src/ directory
- config: the **effective** configuration, resolved in three layers — built-in
  defaults, then `<repo root>/.minispec/config.toml`, then this design root's own
  `.minispec.toml`
- origins: for each setting, which layer supplied it. Recorded during merging, since
  after merging there is no way to tell where an inherited value came from
- config.languages: the `[[languages]]` definitions (`minispecsdom.LanguageDef`), layered by
  name — a later layer's definition replaces an earlier one of the same name whole, and a new
  name is added (R555, R527)
- config.track: the one **repository-scoped** setting. It resolves like any other
  scalar, but only the repository layer may supply one — see `applyLayerFile` below.
  Its meaning and verification belong to `Track`, not here

## Does
- Detect(): walk up from cwd to find design/ directory
- LoadConfig(): resolve the effective configuration across the three layers, rejecting
  a legacy YAML configuration — `.minispec/config.yaml` or a design root's `.minispec.yaml`
  — and a `.minispec.toml` at the repository root before reading anything (R122, R522). A missing
  repository root simply means no repository layer
- applyLayer(cfg, layer): apply one layer over what is resolved — scalars replace,
  languages replace by name (R527), lists union with duplicates dropped and inherited order kept —
  and record the layer as the origin of every setting it supplied
- applyLayerFile(cfg, path, isRepoLayer): read one layer and apply it, **refusing a
  design root that states `track`**. Refused rather than ignored: silently dropping it
  would leave someone editing a line that has no effect, with no way to discover that
- DesignPath(filename): resolve path within design dir
- SrcPath(filename): resolve path within src dir
- decodeLayer(path): build each `[[languages]]` definition the file holds and fail naming
  the file when one does not check (R528); then decode one file strictly — every key the file sets that the tool does not
  read is an error naming the file and the key, and `comment_patterns` / `comment_closers` are
  named as retired, pointing at `query comment-patterns` (R521)
- Languages(): the `minispecsdom.Configured` the definitions make — each extension to its
  table, and each `files` pattern, in configuration order, to its table; an attaching
  definition's table is the built-in it names (R555, R556)
- ResolveSpecSource(src): map a Source value to its on-disk path; for `specs/migrations/X.md` falls back to `specs/migrations/complete/<NNN>-X.md` (NNN digits) so requirements pointing at migrated-completed specs still resolve

## Collaborators
- Parser: to load and parse design files
- RepoRoot: to locate the repository configuration, and the forbidden `.minispec.toml`
  beside it. A failure to resolve is not an error here — it means no repository layer
- os/filepath: for path operations
- github.com/BurntSushi/toml: to decode each configuration layer; `MetaData.Undecoded()` is
  what names an unknown key

## Sequences
- seq-init.md
- seq-config.md
