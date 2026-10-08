# Update
**Requirements:** R4, R18-23, R62, R80-83, R103, R310, R311, R313-316, R324-326, R533-537, R547-549

Atomic modifications to structured parts of design files.

## Knows
- project: loaded Project instance

## Does
- Check(file, item): check a checkbox in file
- NumberAlarms(paths): number every unnumbered alarm in the named test designs, all by
  default, through the reader's append-only `NumberAlarms`; reports what each document was
  assigned and refuses a document whose entry deviates, before any byte moves (R311, R313)
- SetPulled(key, body, now): write an alarm's `**Pulled:**` line through the reader —
  `now` from the system clock, the body from the caller's file — the old line folded after
  it (R314)
- SetInject(key, sites, ranger): rewrite the sites and decide `void` by comparing the old
  sites' extents in HEAD with the new sites' on disk through a `Ranger`, so a rename or a
  disambiguation keeps the record and a move demotes it (R315)
- AddReq(section, texts): mint the next free `Rn` for each text, retired numbers counted, and
  append them through the requirements reader to the section addressed by heading text at any
  level, before its first sub-heading; unknown or ambiguous heading refused, a body writing its
  own label refused, nothing written on a refusal (R324, R325)
- AddGap, ResolveGap, ApproveGap, Retire: through the gaps and requirements readers' writes;
  minting stays here (`On`, `An`, `Tn` from what the reader returned), the refusals are the
  readers' (R82, R83, R80, R326)
- resolveAlarm(key): `<doc>#<n>` to a design-directory path and a number; anything else is
  refused by name (R310)
- Uncheck(file, item): uncheck a checkbox in file
- AddRef(crcFile, reqID): add Rn to CRC card's Requirements field, rewriting the whole field
  in canonical form; a card with no Requirements line gets one beneath its `#` heading; a ref
  already present, inside a range or not, changes nothing (R20, R533, R534)
- RemoveRef(crcFile, reqID): remove Rn, rewriting the field canonically so a range splits; the
  last ref removes the line; a ref the card lacks is an error and the card is untouched (R21,
  R533, R535, R537)
- rewriteRequirements(path, edit): the one field rewrite both verbs share. Reads the field
  through `sdom.ParseRequirementList` — refs as numbers, the unconsumed rest as tokens — lets
  the verb edit the number set, and writes `sdom.RequirementText(set)` followed by the tokens
  as written (R533, R536)
- AddGap(gapType, desc): add new gap with auto-numbered ID; A-typed and T-typed entries are written without a leading checkbox
- ResolveGap(gapID): mark gap as resolved (check its checkbox); refuses A and T types
- ApproveGap(gapID): convert existing gap to A type with next A-number, preserve description; written without checkbox
- Retire(oldReq, replacement, reason): rewrite the oldReq line in requirements.md to the strikethrough/Retired form AND append a new T-typed gap to design.md — both rendered and read back in memory before either file is written, the gap's file written first, and the Tn one past the highest T in either document (R547–R549); returns the assigned Tn AND the retired requirement's **Source:** spec(s), so the CLI can emit the supersede-at-source reminder
- MigrationComplete(name): move specs/migrations/<name>.md to specs/migrations/complete/<NNN>-<name>.md with the next zero-padded prefix; returns the new path

## Collaborators
- Project: to locate files
- Parser: to find line numbers and current state
- os: file writing and rename

## Sequences
- seq-update.md
