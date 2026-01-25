# CLI
**Requirements:** R1, R2, R35, R36, R49, R50, R54, R55

Command-line interface handling.

## Knows
- args: command-line arguments
- flags: parsed flags (--design-dir, --src-dir, --quiet, --json, --version)
- outputFormat: text or json
- version: build-time version string

## Does
- Parse(args): parse command and flags
- Run(): dispatch to appropriate handler (or show version if --version)
- Output(data): format and print result (text or JSON)
- Error(err): print error to stderr
- PrintVersion(): display version and exit

## Collaborators
- Project: to initialize project context
- Query: for query subcommands
- Update: for update subcommands
- Validate: for validate command
- Phase: for phase subcommands
- flag: for flag parsing
- encoding/json: for JSON output

## Subcommands
```
minispec query <subcommand>
minispec update <subcommand>
minispec validate
minispec phase <spec|requirements|design|implementation|gaps>
```
