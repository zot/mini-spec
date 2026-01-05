# CRC: ContactStore

**Type**: Observable Data Store

## Responsibilities
- Manage contact collection (CRUD)
- Persist to localStorage
- Notify observers on changes
- Filter contacts by search term

## Collaborators
- Contact (stores)
- ContactListView (notifies)
- DetailPanel (receives updates from)

## Interface
```typescript
class ContactStore {
  subscribe(callback: () => void): () => void;
  getAll(): Contact[];
  getFiltered(term: string): Contact[];
  getById(id: string): Contact | undefined;
  add(data: Omit<Contact, 'id'>): Contact;
  update(id: string, data: Partial<Contact>): void;
  delete(id: string): void;
}
```

## Sequences
- seq-crud.md (add, update, delete)
- seq-search.md (getFiltered)
