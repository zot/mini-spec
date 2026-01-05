# UI Layout

## Main Layout

```
+------------------------------------------------------------------+
|  HEADER                                                          |
|  +---------------------------+  +-----+  +-----+  +------------+ |
|  | [Search...]               |  | 12  |  | Add |  | Dark [x]   | |
|  +---------------------------+  +-----+  +-----+  +------------+ |
+------------------------------------------------------------------+
|                                                                  |
|  +------------------------+  +--------------------------------+  |
|  | CONTACT LIST           |  | DETAIL PANEL                   |  |
|  |                        |  |                                |  |
|  | > John Doe        VIP  |  |  Name:   [________________]    |  |
|  |   jane@test.com        |  |  Email:  [________________]    |  |
|  |   Bob Smith            |  |  Status: ( ) Active            |  |
|  |                        |  |          ( ) Inactive          |  |
|  |                        |  |  VIP:    [ ] Yes               |  |
|  |                        |  |                                |  |
|  |                        |  |  [Save]  [Cancel]  [Delete]    |  |
|  |                        |  |                                |  |
|  +------------------------+  +--------------------------------+  |
|                                                                  |
+------------------------------------------------------------------+
```

## Components

**Header** (crc-App.md)
- Search input: text field, placeholder "Search..."
- Badge: shows filtered count
- Add button: primary action
- Dark mode toggle: checkbox with label

**Contact List** (crc-ContactListView.md)
- Scrollable list
- Each item: name (bold if VIP), email below
- Selected item highlighted
- Hover state

**Detail Panel** (crc-DetailPanel.md)
- Hidden when no selection
- Form fields: name, email, status radio, VIP checkbox
- Action buttons: Save (primary), Cancel, Delete (danger)

## Styling

- Font: system sans-serif
- Spacing: 8px base unit
- Border radius: 4px
- Colors: see manifest-ui.md
