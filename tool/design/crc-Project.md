# Project
**Requirements:** R32, R33, R34, R35, R38, R39, R57, R58, R93

Finds and loads a mini-spec project's configuration and design files.

## Knows
- rootPath: project root directory
- designDir: path to design/ directory
- srcDir: path to src/ directory
- config: loaded configuration (from .minispec.yaml or defaults)
- commentPatterns: map of file extension to comment prefix regex (e.g., ".go" -> `//\s*`)
- commentClosers: map of file extension to closing delimiter (e.g., ".md" -> ` -->`)

## Does
- Detect(): walk up from cwd to find design/ directory
- LoadConfig(): read .minispec.yaml or use defaults (merges user closers over defaults)
- DesignPath(filename): resolve path within design dir
- SrcPath(filename): resolve path within src dir
- CommentPattern(ext): return regex pattern for the given extension (with defaults)
- CommentCloser(ext): return closing delimiter for the given extension (empty if line-terminating)
- ResolveSpecSource(src): map a Source value to its on-disk path; for `specs/migrations/X.md` falls back to `specs/migrations/complete/<NNN>-X.md` (NNN digits) so requirements pointing at migrated-completed specs still resolve

## Collaborators
- Parser: to load and parse design files
- os/filepath: for path operations

## Sequences
- seq-init.md
