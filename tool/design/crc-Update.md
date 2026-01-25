# Update
**Requirements:** R18, R19, R20, R21, R22, R23, R4

Atomic modifications to structured parts of design files.

## Knows
- project: loaded Project instance

## Does
- Check(file, item): check a checkbox in file
- Uncheck(file, item): uncheck a checkbox in file
- AddRef(crcFile, reqID): add Rn to CRC card's Requirements field
- RemoveRef(crcFile, reqID): remove Rn from CRC card's Requirements field
- AddGap(gapType, desc): add new gap with auto-numbered ID
- ResolveGap(gapID): mark gap as resolved (check its checkbox)

## Collaborators
- Project: to locate files
- Parser: to find line numbers and current state
- os: file writing

## Sequences
- seq-update.md
