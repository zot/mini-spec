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

Types: S (spec), R (requirement), D (design), C (code), I (implementation), O (oversight), A (approved)

Example:
```
minispec update add-gap R "Requirement R5 has no design coverage"
# Adds: - [ ] R2: Requirement R5 has no design coverage
# (assuming R1 already exists)
```

## minispec update resolve-gap [id]

Mark a gap as resolved (check its checkbox).

Alias for `minispec update check design.md [id]`

## minispec update approve-gap [id]

Convert an existing gap to approved (A) type. The gap's type changes to A with a new auto-numbered A-ID. The description is preserved. Approved gaps are always unchecked (`[ ]`).

Example:
```
minispec update approve-gap D3
# Changes: - [ ] D3: Some design gap
# To:      - [ ] A1: Some design gap
# (assuming no A gaps exist yet)
```
