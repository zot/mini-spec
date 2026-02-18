# A 3-Level Process for AI Development to Stop Your App From Shifting Under Your Feet When You Change It

You're iterating: you prompt your AI, it writes code, repeat. Things are going well... but then you make a change to one part of your app and something else quietly breaks. A feature you relied on isn't there. A UI element moved. A workflow you never touched now behaves differently.

You didn't ask for any of that.

If this sounds familiar, you're not alone. It's not a bug: it's because AIs work by telling stories. They make stories that make sense as much as possible, so they fill in missing details automatically. If you don't supply them, the AI generates them. That's what makes AI-assisted coding so productive. But it's also what makes apps unstable across multiple changes.

## The problem: unanchored code

However you work with your AI: vibe coding, specs and code, or whatever, it does its best to make minimal, targeted changes. And for simple, localized updates, that usually works fine. But because of how AI works, it can *add* features you didn't ask for. Sometimes they're good. You come to rely on them. A nice loading animation. A keyboard shortcut. A bit of polish that makes the app feel right.

The problem is that since you didn't ask for these features, they only exist in code. They're not in your specs. They're not documented anywhere. They're **unanchored**.

And anything unanchored is up for grabs the next time the AI touches that code. It can and will remove features that it added.

Simple changes (fixing a bug, tweaking a label) usually leave unanchored features alone. But cross-cutting changes are a different story. Move items from one page to another. Restructure a navigation flow. Refactor a shared component. These changes hit enough code that unanchored features get caught in the blast radius. They can be silently introduced, altered, or removed, and you won't know until you or your users stumble into it.

*This happened to me: there was a back button I hadn't asked for at the upper left of every page of my web app; it was a nice addition. Then at some point I added another page and it didn't get a back button. That page wasn't technically wrong because I never asked for back buttons but the new page wasn't like all the other ones. It's just that this time around, the AI just looked at the home page to add an entry for the new page and it didn't "see" the other pages with their back buttons.*

This makes it hard to maintain a shippable product. Users develop a "flow": the way they navigate your app, the features they rely on, the muscle memory they build up. When unanchored features shift after an upgrade to the app, that flow breaks. Users hate this. It erodes trust in the product. Just ask anyone who uses office software that changed the locations of its buttons and menus.

If you're building personal tools, the problem hits even harder. There's no QA team between a code change and you hitting the breakage in the middle of your own workflow. You *are* the user, and the surprise is immediate.

But here's the worst part: these changes can lie in wait. Your automated tests won't catch them because they were written against the features you *planned*, not the ones the AI added. The user might not even hit the dissonant change until weeks later or even after a second upgrade.

This isn't unique to spec-driven development, either.

If you're vibe coding, prompting your way through a project without formal specs, you have the most extreme version of this problem. Your code *is* the spec. Every prompt changes it.

*Nothing is anchored.*

## The fix: a design layer that anchors your code

The fix actually occurred thousands of years ago to solve a communication problem between the people who commissioned a device and the people who built it. The gap between "what did I ask for" and "what did I get" is old. It turns out the same gap exists between you and your AI.

A 3-level process adds a layer between specs and code: **design**.

This doesn't mean waterfall or UML ceremony. We chose one of the most lightweight agile processes for [mini-spec](https://github.com/zot/minispec). The process supports incredibly fast development turnaround so the people who ask for code can see it as soon as possible, correct miscommunications, iterate, and *just ship it already*. That's exactly the dynamic in AI-assisted coding.

In mini-spec, the design layer is a set of small artifacts: CRC cards that describe your components and their responsibilities, sequence diagrams that capture workflows, UI layouts that show what the user sees. It's not the only way to do a 3-level process, but it's simple and the artifacts are brief and easy for humans to read. And that means they're also brief and easy to read for AI, which lowers token costs. They're written in the same conversational style as your specs, and the AI generates them from your specs just like it generates code.

The key insight is that the design layer **dictates far more of the codebase than specs alone**. Your specs describe what you want. Your design describes how the AI interpreted what you want, including the unplanned features it will add. This lets you peek into how it will implement your specs and, if you invest the time to read through the design, you can save a lot of iteration time (i.e. tokens) by correcting misunderstandings there *before* it generates code.

Once a feature is in the design layer, it's anchored. Code updates respect what's in design. Even if a feature isn't in the spec, if it's in the design, it stays. The AI won't silently remove it or change it when you make a cross-cutting update, because the design is right there telling it what the code should do.

This changes the stability picture:

- **Cross-cutting changes propagate through design first.** When you update a spec, the AI updates the design before touching code. You can review the design change, which is much smaller and easier to read than a code diff, and verify the interpretation before it reaches your codebase.

- **Gap analysis catches what's missing.** Comparing specs, design, and code against each other reveals holes: things present in one layer but missing from another. You catch these before they become surprises.

- **Traceability is the anchor line.** Code ties back to design (CRC cards, sequences), design ties back to specs. This binding is what makes anchoring work. When refactoring, you can see *why* code exists, not just what it does. You won't accidentally remove behavior that traces to a real requirement.

- **Restructuring becomes safe.** If the intent is documented in the design, you can verify that refactored code still satisfies it. The design is the contract the code is built against.

Gap analysis and traceability are decades-old software engineering practices, not something we invented for AI. They just happen to solve the AI stability problem remarkably well.

## The bonus: design as a whiteboard

Once the design layer exists, something unexpected happens: it becomes the best place to think about your project.

Design artifacts are much smaller than code, easier to review, and fit in AI context windows where full codebases don't. This makes the design layer a natural "whiteboard," a collaborative surface where you and the AI think together about the project.

Instead of iterating at the code level (expensive, slow, hard to review), you iterate at the design level. Rearrange components. Discuss workflow changes. Spot problems. All before a single line of code changes. After a design change, it updates the code *and* your specs (when needed).

UI designs make this tangible for everyone. Even non-technical stakeholders can look at a screen layout and spot a missing feature ("where's the delete button?") or catch a misinterpretation of the spec before it reaches code.

Sequence diagrams do the same for code behavior. You don't need deep code knowledge to follow a sequence diagram and notice that a step is missing, or that the user shouldn't have to confirm twice. Two complementary review surfaces: UI designs for what the user sees, sequences for what the system does.

When design-level decisions solidify, when you've iterated enough that you're confident in an approach, they get promoted back to specs as the source of truth. Specs are the authority. Code is the output. Design is where the conversation happens.

---

A 3-level process isn't a mandate. If you're vibe coding and it's working for you, keep going. If you're doing spec-driven development and your project is small enough that instability isn't a problem yet, you might not need this.

But if you've ever looked at your app after an update and thought "I didn't change that," if features keep shifting under your feet, the design layer is how you make them stay.
