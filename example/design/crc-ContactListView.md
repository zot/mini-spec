# CRC: ContactListView

**Type**: View Component

## Responsibilities
- Render contact list items
- Handle item selection
- Update on store changes
- Apply search filter from header

## Collaborators
- ContactStore (observes)
- App (notifies selection)
- Contact (displays)

## Interface
```typescript
class ContactListView {
  constructor(container: HTMLElement, store: ContactStore);
  setSearchTerm(term: string): void;
  onSelect(callback: (id: string) => void): void;
  render(): void;
}
```

## Sequences
- seq-search.md (filter display)
- seq-crud.md (re-render on changes)
