# Carve: reference discipline — links that resolve for a cloner

Publishing a document publishes its pointers. A public document citing a private
working note hands a cloner a dangling link: it resolves to nothing, and nothing
warned that it was ever going to. This carve is the rule for which documents may
cite which, and the check that enforces it.

**Provenance.** Split out of [trajectory-tool.md](trajectory-tool.md) on
2026-08-04 (Bill's call). It arrived there because trajectory documents are where
the problem was noticed, but the rule is a validator over markdown links in any
project document and is useful on `design/` and `specs/` today. Nothing about it
is trajectory-specific.

## Status

- [ ] **Item 1 — the checker: markdown-aware extraction, resolution, git status.** **OPEN (not queued.)**
- [ ] **Item 2 — the document-class model.** **OPEN (not queued.)**
- [ ] **Item 3 — wire into `validate` and report.** **OPEN (not queued.)**

## Decisions

**DECIDED (Bill, 2026-08-04): a reference is a real markdown link**, `[text](path)`
— not a prose mention. That is what makes it checkable at all, and it is the
precondition for everything else here.

**What each document class may point at:**

| document | may reference |
|---|---|
| private carve directory | anything, including planning scratch |
| public `carves/` | **only VCS-managed files** |
| `specs/migrations/` | **only VCS-managed files** |
| the pending file | in between — it points at both scratch notes and carves |

**DECIDED (Bill, 2026-08-04): git only, through the `git` command line.** No fossil
support and no linked-in git library. Supporting a second VCS means the tool has to
know how to *operate* it, a large surface for a check this small; shelling out gets
tracked/ignored status for free and stays correct as git changes.

*A real narrowing, stated rather than hidden:* a project whose public documents are
managed by something else gets no reference checking. In a git+fossil project the
rule still works wherever the public documents are the git-managed ones — the
arrangement that motivated this — but a fossil-only document's references go
unchecked, and the tool should say so plainly rather than pass silently.

**DECIDED (Bill, 2026-08-04): error on ignored, warn on untracked-but-not-ignored.**
The first is certainly wrong — a link into an ignored path can never resolve for a
cloner. The second is usually just early: writing a carve and its referenced doc in
one session and filing the item before committing is the ordinary workflow, not a
mistake.

## The split

**Item 1** — the checker. Extract links, resolve them relative to the containing
file, classify each as tracked / untracked-but-not-ignored / ignored / missing.

*Cheap, but not trivial — it must parse markdown, not grep it.* A throwaway version
run over the trajectory carve on 2026-08-04 flagged a missing file that was not a
link at all: the literal `[text](path)` inside the code span in the decision above.
Links in code spans and fenced blocks are examples, not references, and a regex
cannot tell the difference. That failure is the safe direction — a false alarm
rather than a silent pass — but it is the same lesson as the instrument table in
the trajectory carve, found on the document that argues it.

**Item 2** — the document-class model: which classes exist in this project, which
are public, and what each may cite. Coupled to Item 1 of
[trajectory-tool.md](trajectory-tool.md), which decides how a project declares its
siting; the class of a document is a fact about where it lives.

**Ark made the case that this cannot be inferred.** Its queue files are fossil-only
and untracked in git — private by a filename-case convention (top-level uppercase =
private) that git cannot see — while its `carves/` are public and pushed. Untracked
and not ignored is exactly the state of a file written five minutes ago, so the one
mechanism that could have answered returns the ambiguous answer for the real case.

**DECIDED (Bill, 2026-08-04): trajectory files are gitignored in every project**, so
the queue's class stops being inferred and starts being declared where the checker
already looks. That settles one class universally and makes "a public carve may not
link the queue" a constant rather than a per-project answer — but it settles only
that class. Everything else here still needs the model.

**Item 3** — wire into `validate` and report. Errors and warnings distinguished per
the decision above; the report names the citing file, the link, and why it failed.

## The hole the tool cannot close

The rule checks markdown links, so a *prose* mention of a private file is invisible
to it. Ark's carve README already governs that case: a working note named in prose
without a link must have its contribution *stated in the sentence that names it*, so
nothing is lost by not having the file. The two conventions compose, but only the
linked half is machine-checkable.

**The prose half stays a human obligation**, and it is load-bearing rather than
decorative: ark's own convention exists because its carves must reference private
queue items, which can never be links. So the unlinkable reference is not an
edge case to be stamped out — it is the normal way a public document points at a
private one, and the check must not push people toward deleting the mention
instead of writing the sentence.

## Validated against a real project, 2026-08-04

The rule was prototyped by hand over ark's six public carves before being written
down. Every markdown link resolved to a tracked file; nothing was flagged. Ark's
planning scratch is gitignored, so a link into it from a public carve *would* error
— the case the rule exists for is reachable, and currently unviolated.

The check also found a live violation of the prose half, written that same
afternoon: a carve line reading "verified live by <two private rig files>", which
tells a reader with only the repository nothing at all. Rewritten to state what the
rigs showed.

## Open questions

1. **How does a document declare its class?** By location (a configured list of
   public directories), by filename convention (ark's uppercase rule), or by a
   marker in the file? Location is the most mechanical; ark shows convention is
   real in the wild.
2. **What about links out of the repository** — a URL, or a path above the project
   root? Ignore them, or check only that they are well-formed?
3. **Does this run on every `validate`, or on demand?** It shells out to `git` once
   per referenced path unless batched, and `validate` is run constantly.
