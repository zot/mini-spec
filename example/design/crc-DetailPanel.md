# CRC: DetailPanel

**Type**: View Component

## Responsibilities
- Display/edit contact form
- Handle Save/Cancel/Delete actions
- Validate input
- Show/hide based on selection state

## Collaborators
- ContactStore (updates)
- App (receives selected contact)
- Contact (displays/edits)

## Interface
```typescript
class DetailPanel {
  constructor(container: HTMLElement, store: ContactStore);
  show(contact: Contact | null): void;  // null = new contact
  hide(): void;
  onSave(callback: (contact: Contact) => void): void;
  onDelete(callback: (id: string) => void): void;
  onCancel(callback: () => void): void;
}
```

## Sequences
- seq-crud.md (save, delete operations)
