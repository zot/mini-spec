# Sequence: CRUD Operations

## Add Contact

```
User -> App: click Add button
App -> DetailPanel: show(null)
DetailPanel: display empty form
User -> DetailPanel: fill form, click Save
DetailPanel -> ContactStore: add(data)
ContactStore: generate id, store, persist
ContactStore -> ContactListView: notify change
ContactListView: re-render list
App -> DetailPanel: hide()
```

## Edit Contact

```
User -> ContactListView: click item
ContactListView -> App: onSelect(id)
App -> ContactStore: getById(id)
App -> DetailPanel: show(contact)
User -> DetailPanel: modify, click Save
DetailPanel -> ContactStore: update(id, data)
ContactStore: update, persist
ContactStore -> ContactListView: notify change
ContactListView: re-render list
```

## Delete Contact

```
User -> DetailPanel: click Delete
DetailPanel -> App: onDelete(id)
App -> ContactStore: delete(id)
ContactStore: remove, persist
ContactStore -> ContactListView: notify change
ContactListView: re-render list
App -> DetailPanel: hide()
```

## Cancel Edit

```
User -> DetailPanel: click Cancel
DetailPanel -> App: onCancel()
App -> DetailPanel: hide()
```
