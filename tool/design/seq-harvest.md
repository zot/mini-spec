# Sequence: the harvest

**Requirements:** R513, R514, R515

```
1. Harvest.HarvestArtifacts(root, artifacts)
   1.1. each code file the manifest lists, once, in manifest order
   1.2. skip a listed file that does not exist (validate reports it as a missing artifact)
   1.3. HarvestFile(root, path) for the rest; collect its comments and its unread entry
2. Harvest.HarvestFile(root, path)
   2.1. Languages.LanguageFor(ext); none → unread, "no language for <ext>", line 0; stop
   2.2. sdom.Parse(src) through a BracketParser over that table
   2.3. minispecsdom.Comments(doc, ctx): every comment-kind opener tried, each success spliced
   2.4. each comment → HarvestComment{Line, Text on one line, CRC, Seq as written, Refs expanded to Rn}
   2.5. the first unclosed restricted group (a string or comment run to end of input) → unread,
        at its line, keeping the comments already read; stray closers and open code brackets
        hide nothing and are not reported
```

Nothing in 2.3 or 2.4 depends on the language: `Comments` compares each opener's group kind
with the table's `Comment.Kind`, so a table with three comment forms (HTML's `<!--`, and the
`//` and `/*` inside its scripts) yields comments from all three with no code per form.
