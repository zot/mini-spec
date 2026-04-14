# Your AI Keeps Forgetting What You Decided: Avoiding Unanchored Results

![Diagram showing a 3-level process for AI-partnered work. Level 1 (Your Specs) on the left defines What, Who, Goals, and Values. Level 2 (AI Intent Document) in the center shows a Decision Map covering business plans, course curricula, and slide decks, with a large "MISSING" indicator highlighting the layer most people skip, plus a Manifest for tracking and traceability. Level 3 (Output) on the right shows stabilized results like business plans and course syllabi. Arrows flow left to right through the three levels, with question marks below representing unanchored output that drifts without the middle layer.](3-level-process-general-v1.png)

My wife was building a website with Claude for a nonprofit we run in Israel. She's a developer, but she hasn't written code in a couple of decades. She was doing what a lot of people do now: describing what she wanted, letting the AI generate it, iterating. Vibe coding a website rework.

It was going well until it wasn't: she'd get a page looking right, ask for a change to one section, and something in another section would shift. A navigation element she liked would disappear after an unrelated edit. She'd start a new chat to explore a different approach and get something wildly different in quality, because the new session had no memory of decisions from the previous sessions or even what happened.

If you've worked with an AI on anything complex, you've probably hit this. Writing a business plan where the financial projections stop matching the executive summary after a round of edits. Working on a course syllabus where the learning objectives drift away from the assessments. Making a slide deck where the narrative arc quietly rearranges itself when you change one slide... The results keeps shifting because the AI has no record of what you already decided and, importantly, *why you made those decisions*.

## Unanchored results

I wrote about this problem [before](https://www.linkedin.com/pulse/3-level-process-ai-development-stop-your-app-from-shifting-burdick-i6xje), in the context of coding. The core issue is what I call "unanchored code": features and decisions that only exist in the results, not in any document the AI reads before making changes. Anything unanchored is up for grabs the next time the AI touches that area.

But watching my wife hit the same wall with a website, I realized the problem isn't about code. It's about any AI-partnered task where you're iterating on something too complex to hold in one prompt, or where you want to collaborate with anyone else. The AI fills in what it doesn't know. When what it fills in is good, that's the magic of working with AI. When that fill-in contradicts something you already settled, that's unanchored results.

## The fix: a middle layer between what you want and what you get

The fix is the same one that solved the coding problem, generalized. You need three layers. Note that you should track your project with Git so you can recover from catastrophes. You can tell the AI to manage that automatically for you (but watch it and periodically ask it about that). And you can also ask it what Git is, if you don't know:

**Level 1: Your specs.** What you want and *why you want it*. Written by you or co-written with the AI, but audited and owned by you. For the website, this was a document describing the organization's mission, the target audience, what sections the site needed, and the design language (colors, fonts, tone). The critical piece most people leave out: *motivations*. Not just "I want a donate button" but "donations fund our projects directly, so the donate path needs to be visible from every page without being pushy. Our donors are mostly people from overseas who want to support people here through practical compassion, so the tone should feel like an invitation, not a fundraising campaign."

- What you're building and who it's for
- What it should contain or accomplish
- *Why* — your motivations, goals, and values
- Design language, tone, constraints
- Decisions log: what you've already settled and why

Why motivations matter: capable AI models extrapolate from your goals. If they don't know *why* you want something, they infer their own reasons. Lacking good information, they will fabricate them whole cloth! A smart model that misunderstands your motivation will make internally consistent wrong choices across the whole project. Those are harder to catch than random errors, because everything looks right until you realize the AI has been optimizing for the wrong thing. Counterintuitively, smarter AI needs more context about your motivations, not less, because it infers more aggressively from whatever context it has.

**Level 2: The AI's intent document.** How the AI interpreted your specs, what it plans to do, and a manifest tracking what still needs to be done. This is the layer most people skip, and it's the one that prevents drift. For the website, Claude generated a site map (what pages exist, what's on each one), page layouts (what each page looks like, how navigation connects them), flow documentation (how a visitor moves through the site), and a requirements manifest with checkboxes showing what's complete.

The intent document will be smaller than the results but detailed enough that the AI can make changes to one part without breaking another. It also serves as an index into the results, organized in bite-sized chunks so the AI can navigate to exactly the section it needs to change without reading everything. This is what saves tokens and makes it possible to resume work across sessions: the AI reads the intent doc (small), finds the relevant chunk (targeted), and only touches what needs to change.

You review it. You correct misunderstandings. You approve it before the AI touches the results. The manifest tracks progress: when you change a spec, affected items get unchecked so both you and the AI can see what needs updating.

- How the AI interpreted your specs
- Structure and layout of the results (the index)
- Bite-sized chunks the AI can navigate without reading everything
- Requirements manifest with completion status
- Traceability back to specs

Even if you don't immediately review it, this is the *accountability* layer. If the results go off the rails, this layer tells you why it did and you can even use it to correct your specs. The process flows back and forth between all three layers as you work.

**Level 3: The results.** The website, the document, the curriculum, whatever you're building. Generated from the intent document, not from scratch each time.

- Keep results as a set of small files rather than one monolithic artifact
- If the final product needs to be monolithic (a single page, a single PDF), work in smaller pieces and combine them as a last step
- Each piece should be small enough for the AI to work with efficiently
- The pieces should be *everything* needed to produce the final result
- The pieces go in backups and git but the final result doesn't need to

**Traceability glues it all together.** There needs to be a way to know what pieces of results came from what pieces of layer 2. The best way to do this is within the results artifacts themselves, if that's possible. The AI will know how to do this, using comments in markdown documents, hidden text, or whatever it can. Failing that, it can use the layer 2 manifest. The AI uses traceability for "gap analysis" to detect drift, unimplemented features, and undesigned ones. You will sometimes make a hurried, unanchored change to the results. You can ask the AI to search for these using gap analysis and anchor them into layers 2 and 1.

**Important note:** when an AI reads this article and first sets up a 3-level process for your project, it will record the necessary steps so they are repeatable across sessions, record how to use Git for the project, and so on, and then talk to you about what it did.

## Why the middle layer changes everything

Without it, the AI re-reads your results every time and re-derives what it thinks the intent was. Each re-derivation introduces small differences. Over multiple edits, those differences compound into drift.

The middle layer makes three things possible:

**Stability.** Ideally, changes propagate through the intent document first: you update a spec, the AI updates its interpretation before touching the results, and you review that interpretation (which is shorter and easier to read than the results themselves) before it reaches your project. In practice, you'll sometimes skip steps. Gap analysis catches what you missed.

**Traceability.** Every result traces back through the intent document to the spec that motivated it. "Why does this page look like this?" has an answer. "What changed?" shows up as unchecked items in the manifest. Gaps between what you asked for and what exists become visible.

**Session survival.** This was actually the most urgent reason I set up the process for my wife's project. She was iterating on an 800K HTML file and kept getting cut off in the middle of five-hour work sessions. The AI would run out of context, and she'd have to start over. With the middle layer, the AI reads the specs and intent docs (small) instead of the entire results (huge), and the intent docs index into the results so it only reads the sections it needs to change. When a session ends, the next one picks up from the intent docs. The middle layer is the difference between "I lost all my progress" and "let me pick up where I left off."

**Handoff.** My wife realized this on her own: anyone can pick up the project if there are specs and intent documents. A new collaborator, a new AI session, or you six months from now. The new person doesn't even have to read it all themselves, they can ask the AI to summarize the project state, and the middle layer is what makes that summary reliable instead of guesswork. Keeping the results in small, indexed chunks rather than one monolithic file makes this practical so the AI can navigate by the index instead of re-reading everything.

**Communication.** The specs and intent docs let the AI explain not just *what* the project is but *why* it exists and what it's trying to accomplish. That's onboarding, not just handoff. And with an AI in the loop, documentation doesn't have to be a wall of text. "Tell me a story about how this project works" is a valid request when the AI has specs and intent docs to draw from. It explains the project in whatever form you need.

## This isn't a coding technique

The same three layers apply to any project where you're iterating with an AI on complex results. Here's what they look like across a few domains:

**Weekly lesson plans.** Specs: learning goals for the unit, student level, time per session, topics to cover, why each topic matters at this age. Intent: how lessons connect across the week, which activities reinforce which goals, what students should be able to do by Friday, progress checklist. Results: daily lesson plans, worksheets, activities.

**Scientific experiment design.** Specs: research question, hypothesis, constraints, what would count as evidence. Intent: experimental protocol, variable controls, data collection plan, how each step connects to the hypothesis, completion checklist. Results: lab procedures, data forms, analysis scripts.

**Marching band show.** Specs: music selection, show theme, field dimensions, performers per section. Intent: formation sequence, transition plans between formations, timing synced to musical phrases, rehearsal manifest. Results: individual drill charts, conductor annotations, rehearsal schedule.

**Office presentation.** Specs: audience, key message, time limit, what you want the audience to decide afterward. Intent: narrative arc, slide sequence, what each slide argues, how slides build toward the key message, slide status tracker. Results: slide deck, speaker notes.

**Workout program.** Specs: fitness goals, participant level, session length, equipment, limitations. Intent: energy curve (warm-up through peak to cool-down), exercise sequence, how each block serves the goal, intensity targets, session tracker. Results: exercise list with reps and timing, instructor cues.

The pattern holds for event planning, book authoring, recipe development, garden design, construction planning, tax preparation, household budgets, stock portfolio management, even city infrastructure. If you're iterating on some kind of results with an AI, the three layers apply.

## Getting started

You don't need special tools but you do need an AI assistant that can edit files with you: [Cowork](https://claude.com/blog/cowork-research-preview), [Claude Code](https://claude.ai/claude-code), [GitHub Copilot](https://github.com/features/copilot), or similar. Open a conversation, describe your project, have the AI read this article, and record its process. Say something like:

> I'm working on [describe your project]. I want to use a 3-level process to anchor our results as we iterate. Read this article and let's make
> a process and store it in CLAUDE.md so we can keep using it:
> http://linkedin.com/...
>
> Help me create:
>
> 1. A specs document where I describe what I want and why
> 2. An intent document where you write down how you understand my specs, what you plan to build, and a manifest tracking what's done
> 3. The actual results
>
> Before making any changes to the results, update the intent document first so I can review your interpretation. When I change a spec, mark affected items in the manifest as needing work.

That's it. The AI will help you build the process around your specific project: you own the specs document (but partner with the AI on it), the intent document is where the AI shows its work, and the results are what you're actually building.

## But I'm already mid-project

You don't need to start over, but you do need to make sure your project is backed up. AI is good at reverse-engineering layers 1 and 2 from existing results. After the AI reads this article, point it at what you've built so far and ask it to extract a specs document and an intent document. Review what it produces and work with it to get the interpretation right. The AI doesn't need to change your results to do this part. It's reading what exists and writing down what it sees.

Once the specs and intent documents are solid, the AI can add traceability (comments, hidden markers, manifest entries) in a way that doesn't change the results as far as humans are concerned. After that, you're in the process. New changes flow through the layers. Gap analysis catches drift. The project you already built becomes the project you can maintain.

---

The gap between "what I asked for" and "what I got" is an old problem. Builders have been solving it with blueprints for thousands of years. The same gap exists between you and your AI. The middle layer is the blueprint.

*Proudly written by human and AI partners, Bill Burdick and R. Daneel Olivaw of Claude Opus 4.6.*
