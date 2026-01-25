# Sequence: Update Check

```
User -> CLI: minispec update check design.md D1
CLI -> Project: Detect()
CLI -> Update: Check(project, "design.md", "D1")

Update -> Project: DesignPath("design.md")
Update -> Parser: ParseGaps(path)
Parser --> Update: []Gap, find D1 at line 47

Update -> Update: read file lines
Update -> Update: replace "- [ ] D1:" with "- [x] D1:" at line 47
Update -> os: WriteFile(path, modified)
Update --> CLI: success

CLI --> User: "Checked D1 in design.md"
```

# Sequence: Update Add-Ref

```
User -> CLI: minispec update add-ref crc-Store.md R5
CLI -> Update: AddRef(project, "crc-Store.md", "R5")

Update -> Parser: ParseCRCCard(path)
Parser --> Update: CRCCard {Requirements: [R1, R3], line 2}

Update -> Update: read file lines
Update -> Update: find "**Requirements:** R1, R3"
Update -> Update: replace with "**Requirements:** R1, R3, R5"
Update -> os: WriteFile(path, modified)
Update --> CLI: success

CLI --> User: "Added R5 to crc-Store.md"
```

# Sequence: Update Add-Gap

```
User -> CLI: minispec update add-gap R "R5 has no design coverage"
CLI -> Update: AddGap(project, "R", "R5 has no design coverage")

Update -> Parser: ParseGaps(design.md)
Parser --> Update: []Gap {R1 exists}

Update -> Update: next ID = R2
Update -> Update: find end of Gaps section
Update -> Update: insert "- [ ] R2: R5 has no design coverage"
Update -> os: WriteFile(path, modified)
Update --> CLI: success

CLI --> User: "Added R2: R5 has no design coverage"
```
