---
name: mini-spec
description: user whenever designing or updating application design, code, or docs
---

# Mini-spec

Build and maintain 3-level architectures. Drectory structure
```
project
├── specs/       # Human readable specs
├── design/      # SOURCE OF TRUTH (apps AND shared components)
├── docs/
├── CODE-DIRS/
```

## Workflow

1. **ALWAYS READ SPECS** to understand what to the user wants.

2. **Design**: design the app based on the requirements.
   - Write the design in `design/design.md`:
      - **Intent**: What the UI accomplishes
      - **Layout**: ASCII wireframe showing structure
      - **Data Model**: Tables of types, fields, and descriptions
      - **Methods**: Actions each type performs
      - **ViewDefs**: Template files needed
      - **Events**: JSON examples of user interactions
      - **Styling**: Visual guidelines (optional)

## Complete Example: Contact Manager with Chat

Demonstrates: design spec, lists, selection, nested views, forms, selects, switches, conditional display, computed values, notifications, **agent chat**.

### 1. Design Spec (`.claude/ui/apps/contacts/design.md`)

See [design.md](builder-examples/return.md)

### 2. Lua Code

See [code.lua](builder-examples/code.lua)

### 3. App Viewdef (`ContactApp.DEFAULT.html`)

See [ContactApp.DEFAULT.html](builder-examples/ContactApp.DEFAULT.html)

The ViewList wraps each item with `lua.ViewListItem`. The item's `type` field determines which viewdef renders it.

### 4. Contact Viewdef (`Contact.list-item.html`)

See [Contact.list-item.html](builder-examples/Contact.list-item.html)

### 5. Chat Message Viewdef (`ChatMessage.list-item.html`)

**Important**: ViewList uses `list-item` namespace by default. Items rendered in a ViewList need viewdefs with the `list-item` namespace (e.g., `Contact.list-item.html`, `ChatMessage.list-item.html`).

See [ChatMessage.list-item.html](builder-examples/ChatMessage.list-item.html)
