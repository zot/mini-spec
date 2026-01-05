# CRC: Contact

**Type**: Entity

## Responsibilities
- Hold contact data (id, name, email, status, vip)
- Provide immutable data structure

## Collaborators
- ContactStore (stored by)

## Interface
```typescript
interface Contact {
  id: string;
  name: string;
  email: string;
  status: 'active' | 'inactive';
  vip: boolean;
}
```

## Sequences
- seq-crud.md (all operations)
