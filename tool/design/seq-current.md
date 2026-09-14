# Sequence: the resume buffer
**Requirements:** R385, R386, R388, R389

## 1. Reading

1. `ParseCurrent(src)`
   1.1. Parse with the markdown base
   1.2. Find the level-2 `Heading` whose text is `Active`; none or two is an error
   1.3. The region runs to the next heading of level 2 or higher

## 2. Writing

2. `SetActive(body)` or `Reset`
   2.1. `SetActive` refuses when the region holds anything but the placeholder
   2.2. Inside a window: split the heading's following text after its title line; remove
        every node from the right half to the region's end; insert one synthetic text —
        a blank line, the body, a blank line — before the next heading or at the end
   2.3. Re-read the document
