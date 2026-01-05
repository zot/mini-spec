# Test Design: Contacts

## ContactStore Tests

**CRUD Operations**
- add() creates contact with unique id
- add() persists to localStorage
- getById() returns correct contact
- getById() returns undefined for missing id
- update() modifies contact fields
- update() persists changes
- delete() removes contact
- delete() persists removal

**Filtering**
- getFiltered("") returns all contacts
- getFiltered() matches name (case-insensitive)
- getFiltered() matches email (case-insensitive)
- getFiltered() returns empty for no matches

**Observers**
- subscribe() callback called on add
- subscribe() callback called on update
- subscribe() callback called on delete
- unsubscribe prevents further callbacks

## Integration Tests (Manual)

**Add Flow**
1. Click Add -> panel shows empty form
2. Fill name/email -> Save -> appears in list
3. Cancel without save -> list unchanged

**Edit Flow**
1. Click contact -> form populated
2. Modify fields -> Save -> list updated
3. Cancel -> original values retained

**Search Flow**
1. Type in search -> list filters
2. Badge updates with count
3. Clear search -> all contacts shown

**Dark Mode**
1. Toggle switch -> colors invert
2. Refresh page -> preference persisted
