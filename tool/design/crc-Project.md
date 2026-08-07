# Project
**Requirements:** R32, R33, R34, R35, R38, R39, R57, R58, R93, R118, R119, R120, R121, R122, R123, R124, R125, R126, R127, R128, R129, R130

Finds and loads a mini-spec project's configuration and design files.

## Knows
- rootPath: the **design root** — the directory containing `design/`. Named before
  the two roots were distinguished; the repository root is `RepoRoot`'s business
- designDir: path to design/ directory
- srcDir: path to src/ directory
- config: the **effective** configuration, resolved in three layers — built-in
  defaults, then `<repo root>/.minispec/config.yaml`, then this design root's own
  `.minispec.yaml`
- origins: for each setting, which layer supplied it. Recorded during merging, since
  after merging there is no way to tell where an inherited value came from
- commentPatterns: map of file extension to comment prefix regex (e.g., ".go" -> `//\s*`)
- commentClosers: map of file extension to closing delimiter (e.g., ".md" -> ` -->`)

## Does
- Detect(): walk up from cwd to find design/ directory
- LoadConfig(): resolve the effective configuration across the three layers, rejecting
  a `.minispec.yaml` at the repository root before reading anything. A missing
  repository root simply means no repository layer
- applyLayer(cfg, layer): apply one layer over what is resolved — scalars replace,
  maps merge per key, lists union with duplicates dropped and inherited order kept —
  and record the layer as the origin of every setting it supplied
- DesignPath(filename): resolve path within design dir
- SrcPath(filename): resolve path within src dir
- CommentPattern(ext): return regex pattern for the given extension (with defaults)
- CommentCloser(ext): return closing delimiter for the given extension (empty if line-terminating)
- ResolveSpecSource(src): map a Source value to its on-disk path; for `specs/migrations/X.md` falls back to `specs/migrations/complete/<NNN>-X.md` (NNN digits) so requirements pointing at migrated-completed specs still resolve

## Collaborators
- Parser: to load and parse design files
- RepoRoot: to locate the repository configuration, and the forbidden `.minispec.yaml`
  beside it. A failure to resolve is not an error here — it means no repository layer
- os/filepath: for path operations
- gopkg.in/yaml.v3: to parse each configuration layer

## Sequences
- seq-init.md
- seq-config.md
