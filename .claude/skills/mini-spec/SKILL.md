---
name: mini-spec
description: use **whenever designing, updating, or implementing the design, code, or docs**
---

# Mini-spec

Build and maintain 3-level architectures. Drectory structure
```
PROJECT
├── specs/       # Human readable specs
├── design/      # SOURCE OF TRUTH (apps AND shared components) -- all design files
├── docs/        # Human readable documentation
└── CODE-DIRS/   # names depend on the app language(s)
```

## 🎯 Core Principles
- Use **SOLID principles** in all implementations
- Create comprehensive **unit tests** for all components
- Code and specs are as MINIMAL as POSSIBLE

## 🔗 Traceability: Design ↔ Code Sync

The **Artifacts** section in `design.md` is a memory bank for project state. Each design file lists its implementing code files as checkboxes.

**When code changes:**
1. Find the CRC comment on the changed method → identifies which design file
2. Uncheck `[x]` → `[ ]` for that code file in Artifacts
3. Later, grep for `[ ]` to find stale design that needs review

**When updating design:**
1. Find unchecked items in Artifacts
2. Read the code, update the design file
3. Re-check the box

This bidirectional link keeps design and code in sync without reading everything.

## 🔍 Gap Analysis

The **Gaps** section in `design.md` tracks discrepancies and potential issues:

- **Spec → Design**: Features in spec not yet designed
- **Design → Code**: Design elements not yet implemented
- **Code → Design**: Implementation details not reflected in design
- **Oversights**: Potential issues (missing validation, UX problems, edge cases)

Review gaps when planning work or before releases.

## Workflow

**ALWAYS READ SPECS FIRST** to understand what the user wants.
- The specs **MUST** indicate the desired language(s), environment(s), etc.

### Phase Separation

**"Design" = design only.** Do not implement.
**"Implement" = code only.** Do not redesign—just update `design.md` Artifacts checkboxes.
**"Code changes" = update Artifacts, then ask.** When code changes independently of design:
1. Uncheck affected code files in Artifacts: `[x]` → `[ ]`
2. Ask user: "Design/specs are now out of sync. Update design, specs, or defer?"

This keeps phases distinct and avoids scope creep.

### Design Phase

Create files in `design/` directory:
- `design.md`: main design file
   - **Intent**: What the system accomplishes
   - **Artifacts**: design files, each with sublist of code file checkboxes (unchecked `[ ]`)
- `crc-*`: CRC cards with references to relevant sequence diagram files
- `seq-*`: sequence diagrams
- `ui-*`: Terse, scannable, ASCII art for layouts, reference CRC cards for types/behavior, styling requirements
- `test-*`: test designs
- `manifest-ui.md` for cross-cutting UI concerns: Routes, Global components, UI patterns, Theme, View lifecycle

### Implementation Phase

Create code and tests in the language(s) specified in the specs:
- Add traceability comments on modules and methods:
  ```
  // ContactStore - Observable data store
  // CRC: crc-ContactStore.md | Seq: seq-crud.md, seq-search.md

  // CRC: crc-ContactStore.md | Seq: seq-crud.md
  add(data: Omit<Contact, 'id'>): Contact {
  ```
- Mark code files as implemented: `[ ]` → `[x]` in `design.md` Artifacts

### Documentation Phase

Create `docs/user-manual.md` and `docs/developer-guide.md` with traceability links.
  
