# Sequence: Resolving the effective configuration

**Requirements:** R122, R123, R124, R125, R126, R127, R129

Two diagrams: resolving the layers, and applying one of them.

## 1. Resolving the layers

```
1. Resolve the effective configuration for a design root
1.1. Resolve the repository root; if none is found there is no repository layer
1.2. Reject a .minispec.yaml at the repository root, by existence test, before reading anything
1.3. Start from the built-in defaults
1.4. Apply the repository configuration at <repo root>/.minispec/config.yaml if it exists
1.5. Apply the design root's own .minispec.yaml if it exists
1.6. Return the effective configuration and, for each setting, the layer that supplied it
```

**Step 1.2 is what makes step 1.5 unconditional**, and that is the whole payoff of the
flat rule. Where the repository root and the design root are the same directory, the
design root's `.minispec.yaml` *is* the repository root's, so 1.2 has already rejected
it — 1.5 needs no "unless the roots coincide" clause, and there is no second file to
reason about. A conditional rule here would have produced a check about what else sits
beside the file rather than one existence test.

**Step 1.2 precedes every read**, so the error is about the file's presence and never
about its contents. A malformed file at the forbidden path reports the forbidden path,
not a parse failure.

**Step 1.1 may legitimately find nothing.** A tree with no markers has no repository
layer, and the design root's own file applies directly over the defaults — which is
the behavior that existed before repository configuration did.

## 2. Applying one layer

```
2. Apply one configuration layer over what is already resolved
2.1. Replace each scalar setting the layer sets
2.2. Merge each map setting the layer sets, per key, keeping keys the layer does not mention
2.3. Union each list setting the layer sets against what another configuration layer already supplied, preserving inherited order and appending entries not already present; replace the built-in defaults outright when no configuration layer has set that list yet
2.4. Record this layer as the origin of every setting it supplied
```

**The three rules are one rule seen through three types:** a layer states only what it
adds or changes, never what it keeps. A scalar cannot merge, so replacement is the
only form that rule can take for it.

**Step 2.3 preserves inherited order rather than sorting**, so the result is
deterministic and a reader can see which entries came from where. Dropping duplicates
means a design root repeating an inherited entry is harmless rather than doubling it.

**The defaults exception in 2.3 is load-bearing, not a special case for tidiness.**
Removal works by moving a setting *down* a level, and nothing sits below the built-in
defaults to move it to — so unioning onto them would make the shipped
`code_extensions` permanently un-narrowable, and no project could ever say "this one
is Go-only." Defaults are also not a layer anyone authored, so "state only what you
add" has nothing to attach to. Whether a configuration layer has spoken yet is
already recorded by step 2.4, so the exception needs no extra state.

**Step 2.4 is what makes provenance possible at all.** Recorded during application
rather than reconstructed afterward, because after merging there is no way to tell
which layer an inherited value came from.
