# Update Commands

Atomic modifications to structured parts of design files.

## minispec update check [file] [item]

Check a checkbox in the specified file.

Examples:
```
minispec update check design.md D1      # Check gap item D1
minispec update check design.md src/store.ts  # Check artifact checkbox
```

## minispec update uncheck [file] [item]

Uncheck a checkbox in the specified file.

## minispec update add-ref [crc-file] [Rn]

Add a requirement reference to a CRC card's Requirements field.

Example:
```
minispec update add-ref crc-Store.md R5
# Changes: **Requirements:** R1, R3
# To:      **Requirements:** R1, R3, R5
```

## minispec update remove-ref [crc-file] [Rn]

Remove a requirement reference from a CRC card.

## minispec update add-gap [type] [description]

Add a new gap item to design.md Gaps section with auto-numbered ID.

Types: S (spec), R (requirement), D (design), C (code), O (oversight)

Example:
```
minispec update add-gap R "Requirement R5 has no design coverage"
# Adds: - [ ] R2: Requirement R5 has no design coverage
# (assuming R1 already exists)
```

## minispec update resolve-gap [id]

Mark a gap as resolved (check its checkbox).

Alias for `minispec update check design.md [id]`
