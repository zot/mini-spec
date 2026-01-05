# Contacts App Developer Guide

## Architecture

The app follows a simple MVC-like pattern with observable stores.

```
src/
├── models/Contact.ts    # Data entity
├── stores/ContactStore.ts   # Observable data store
├── views/
│   ├── ContactListView.ts   # List rendering
│   └── DetailPanel.ts       # Form handling
├── App.ts               # Coordinator
└── main.ts              # Entry point
```

*Design: [design.md](../design/design.md)*

## Building

```bash
npm install
npm run build
```

This compiles TypeScript to `dist/`.

## Running Tests

```bash
npm test
```

*Design: [test-contacts.md](../design/test-contacts.md)*

## Key Components

### ContactStore

Observable store with localStorage persistence. All CRUD operations notify subscribers.

*Design: [crc-ContactStore.md](../design/crc-ContactStore.md)*

### ContactListView

Renders filtered contact list, handles selection.

*Design: [crc-ContactListView.md](../design/crc-ContactListView.md)*

### DetailPanel

Form for viewing/editing contacts. Manages its own visibility.

*Design: [crc-DetailPanel.md](../design/crc-DetailPanel.md)*

### App

Wires everything together: event handlers, dark mode, badge updates.

*Design: [crc-App.md](../design/crc-App.md)*

## Traceability

Code files contain CRC comments linking to design docs:

```typescript
// CRC: crc-ContactStore.md | Seq: seq-crud.md
add(data: Omit<Contact, 'id'>): Contact {
```

See [design.md](../design/design.md) Artifacts section for code↔design mapping.
