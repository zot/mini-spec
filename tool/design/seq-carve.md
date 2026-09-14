# Sequence: reading a carve and landing a part
**Requirements:** R362, R363, R364, R365, R368, R391, R392, R393, R394, R409

## 1. Reading

1. `ParseCarve(src)`
   1.1. Parse with the markdown base
   1.2. Find the status heading: level 2, text `Status`; none means no status block
   1.3. The region ends at the next heading of level 2 or higher, or the end
   1.4. For each `ListItem` whose offset is inside the region, `PartLine.Parse` and
        splice, each in its own window; others are left alone
   1.5. A line with a checkbox is a part, depth from the bullet's leading whitespace,
        parent the nearest preceding part with a smaller depth; one without a checkbox
        is stateless

## 2. Landing

2. `Land(key, attribution)`
   2.1. Find the part by key — unknown is an error
        2.1.1. Refuse with a `DeviationError` when the line carries deviations
        2.1.2. Refuse with `ErrLanded` when the checkbox is already checked
   2.2. `SetChecked(true)` on its checkbox; `Strike(true)` on its head
   2.3. `SetMarker("LANDED", attribution)`: replace the first `OPEN` marker, remove any
        other, append when none
        2.3.1. `SetMarker` itself refuses first: a `DeviationError` over deviations, and
               `ErrReopen` when the verb is `OPEN` over a checked box
