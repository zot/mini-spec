# Sequence: Search

## Filter Contacts

```
User -> Header: type in search box
Header -> App: input event (term)
App -> ContactListView: setSearchTerm(term)
ContactListView -> ContactStore: getFiltered(term)
ContactStore: filter by name/email (case-insensitive)
ContactStore -> ContactListView: return filtered[]
ContactListView: render filtered list
App -> Badge: update count
```

## Clear Search

```
User -> Header: clear search box
Header -> App: input event ("")
App -> ContactListView: setSearchTerm("")
ContactListView -> ContactStore: getFiltered("")
ContactStore -> ContactListView: return all[]
ContactListView: render all contacts
App -> Badge: update count
```
