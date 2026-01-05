# Contacts App: an in-browser TypeScript app

A contact manager with search and CRUD operations. Currently there is only one contact list but in the future there might be a way to select different named contact lists.

Vanilla HTML5, no external libraries.

## Layout

Header with search box, count badge, Add button, and dark mode toggle. Below that, a two-column view: contact list on left, detail form on right.

## Features

**Contact List**
- Search filters by name/email as you type
- Badge shows filtered count
- Click list item to edit

**Detail Panel**
- Edit name, email, status (active/inactive), VIP flag
- Save commits changes; Cancel discards
- Delete removes contact
- Hidden when nothing selected

**Add Contact**
- Creates empty contact in detail panel
- Not added to list until Save

**Dark Mode**
- Toggle switch in header
