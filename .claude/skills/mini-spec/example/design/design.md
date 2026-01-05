# Contact Manager

## Intent
Manage contacts with list/detail view. Search and filter.

## Layout
```
┌─────────────────────────────────────────────────────┐
│ [Search contacts...          ] [3] [+ Add] [□ Dark] │
├───────────────────────┬─────────────────────────────┤
│   Alice Smith         │ Name: [Alice Smith      ]   │
│ ▎Bob Jones      ←     │ Email: [bob@example.com ]   │
│   Carol White         │ Status: [Active ▼]          │
│                       │ VIP: [✓]                    │
│                       │ ─────────────────────────── │
│                       │ [Delete] [Cancel]    [Save] │
└───────────────────────┴─────────────────────────────┘
```

## Components

| Element       | ID/Class       | Handler          |
|---------------|----------------|------------------|
| Search input  | #search        | input → render() |
| Count badge   | #count         | —                |
| Add btn       | #addBtn        | click → add()    |
| Dark toggle   | #darkMode      | change → toggle  |
| Contact list  | #list          | —                |
| Row           | .contact-row   | click → select() |
| Detail panel  | #detail        | .hidden class    |
| Name input    | #name          | —                |
| Email input   | #email         | —                |
| Status select | #status        | —                |
| VIP checkbox  | #vip           | —                |
| Delete btn    | #deleteBtn     | click → delete() |
| Cancel btn    | #cancelBtn     | click → hide()   |
| Save btn      | #saveBtn       | click → save()   |

## State

```typescript
contacts: Contact[]      // All contacts
editing: Contact | null  // Original being edited
current: Contact | null  // Working copy in form
```

## Behavior
- Search filters list in real-time by name/email
- Add creates temp contact, shows detail panel
- Click row clones contact into form for editing
- Save commits changes, hides panel
- Cancel discards changes, hides panel
- Delete removes contact, hides panel
