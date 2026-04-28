# Update
**Requirements:** R18, R19, R20, R21, R22, R23, R4, R62, R80, R81, R82, R83

Atomic modifications to structured parts of design files.

## Knows
- project: loaded Project instance

## Does
- Check(file, item): check a checkbox in file
- Uncheck(file, item): uncheck a checkbox in file
- AddRef(crcFile, reqID): add Rn to CRC card's Requirements field
- RemoveRef(crcFile, reqID): remove Rn from CRC card's Requirements field
- AddGap(gapType, desc): add new gap with auto-numbered ID; A-typed and T-typed entries are written without a leading checkbox
- ResolveGap(gapID): mark gap as resolved (check its checkbox); refuses A and T types
- ApproveGap(gapID): convert existing gap to A type with next A-number, preserve description; written without checkbox
- Retire(oldReq, replacement, reason): rewrite the oldReq line in requirements.md to the strikethrough/Retired form AND append a new T-typed gap to design.md; returns the assigned Tn
- MigrationComplete(name): move specs/migrations/<name>.md to specs/migrations/complete/<NNN>-<name>.md with the next zero-padded prefix; returns the new path

## Collaborators
- Project: to locate files
- Parser: to find line numbers and current state
- os: file writing and rename

## Sequences
- seq-update.md
