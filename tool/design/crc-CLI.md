# CLI
**Requirements:** R1, R2, R35, R36, R49, R50

Command-line interface handling.

## Knows
- args: command-line arguments
- flags: parsed flags (--design-dir, --src-dir, --quiet, --json)
- outputFormat: text or json

## Does
- Parse(args): parse command and flags
- Run(): dispatch to appropriate handler
- Output(data): format and print result (text or JSON)
- Error(err): print error to stderr

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
