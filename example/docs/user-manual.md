# Contacts App User Manual

A simple contact manager with search and CRUD operations.

## Getting Started

Open `index.html` in a browser after building (see Developer Guide).

## Features

### Adding a Contact

1. Click the **Add** button in the header
2. Fill in name and email (required)
3. Set status (Active/Inactive) and VIP flag
4. Click **Save**

*Design: [seq-crud.md](../design/seq-crud.md) - Add Contact*

### Editing a Contact

1. Click a contact in the list
2. Modify fields in the detail panel
3. Click **Save** to commit changes, or **Cancel** to discard

*Design: [seq-crud.md](../design/seq-crud.md) - Edit Contact*

### Deleting a Contact

1. Select a contact from the list
2. Click **Delete** in the detail panel

*Design: [seq-crud.md](../design/seq-crud.md) - Delete Contact*

### Searching

Type in the search box to filter contacts by name or email. The badge shows the count of matching contacts.

*Design: [seq-search.md](../design/seq-search.md)*

### Dark Mode

Toggle the **Dark** checkbox in the header. Your preference is saved.

## Data Storage

Contacts are stored in your browser's localStorage. They persist across sessions but are local to your browser.
