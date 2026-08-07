# Sequence: Resolving the repository root

**Requirements:** R108, R109, R110, R111, R112, R113

The upward marker walk. The ordering is the whole design: a strong marker returns
immediately because the first one met going up is by definition the deepest, while a
weak marker is only recorded, because something stronger may still sit above it.

```
1. Resolve the repository root
1.1. Resolve the starting directory to an absolute path
1.2. Determine the home boundary — the user's home directory
1.3. Walk upward from the starting directory
1.3.1. Stop walking if this directory is the home boundary or above it
1.3.2. Return this directory if it holds .git, .minispec/, carves/, or a trajectory file
1.3.3. Record this directory as the .claude candidate if it holds .claude and none is recorded
1.3.4. Record this directory as the .minispec.yaml candidate if it holds .minispec.yaml and none is recorded
1.3.5. Move to the parent, stopping if the parent is this directory
1.4. Return the .claude candidate if one was recorded
1.5. Return the .minispec.yaml candidate if one was recorded
1.6. Fail, naming the starting directory the search began from
```

**Step 1.3.2 is a return, not a record** — that is what makes the four strong
markers equal and unordered. Where several coexist in one directory they agree, so
nothing has to rank them; where they sit at different depths, the deepest is reached
first and the ones above are never examined.

**Steps 1.3.3 and 1.3.4 are independent, not alternatives.** One directory may hold
both a `.claude` and a `.minispec.yaml`, and each is recorded into its own slot.

**Step 1.3.1 precedes every check**, so no marker inside or above the home directory
is ever consulted. Without it, a tree carrying no markers of its own resolves to the
user's home directory, because `~/.claude` exists on a normal installation.
