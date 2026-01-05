# CRC: App

**Type**: Coordinator

## Responsibilities
- Initialize all components
- Wire up event handlers
- Manage selection state
- Handle dark mode toggle
- Update badge count

## Collaborators
- ContactStore (owns)
- ContactListView (owns)
- DetailPanel (owns)

## Interface
```typescript
class App {
  constructor(root: HTMLElement);
  init(): void;
}
```

## Sequences
- seq-crud.md (coordinates flow)
- seq-search.md (wires search to list)
