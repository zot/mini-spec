// CRC: crc-Pending.md | R241, R242, R243, R244, R245, R246, R247, R248
package pending

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/backup"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/simple-dom/minispecsdom"
)

const carveSrc = `# Carve: x

## Status

- [ ] **Item 1 — an already queued part.** **OPEN (#3.)**
- [ ] **Item 7 — a part to queue.** **NOT VERIFIED.**
- **Item 8 — a split parent.** **SPLIT (Bill, 2026-08-14.)**
  - [ ] **8.1 — the first half.** **OPEN (not queued.)**
  - [ ] **8.2 — the second half.** **OPEN (not queued.)**

## Why

prose that must not move.
`

const currentSrc = `# Current context for this project's queue

Scratch space for the **active** item, and nothing else — completed work goes to the done
file, never here. This paragraph is the project's own prose and must survive a completion.

---

## Active

` + "`#4`" + ` — a part to queue.

lots of active context that must not survive a completion.

### a subheading the agent wrote

deeper than the section's own level, so it belongs to the active item and goes with it.

## Waiting on the user

standing context, which must survive a completion exactly as the preamble does — and it
sits **below** the active section, which is where the 320 lines were when they went.

## Left elsewhere

- a pointer to work outside this repository
- a second one, so the reset meets more than one trailed node
`

const pendingSrc = `# Pending

---

## 3. **an older item** (mini-spec).
   Source: [carves/x.md](carves/x.md), part ` + "`#1`" + `.

---

prose after the entries.
`

// fixture builds a repository with the trajectory files, a carve, and enough git for the
// slot's worktree anchor to have something to anchor to.
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"carves", ".minispec"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("PENDING.md", pendingSrc)
	write("DONE.md", "# Done\n\n---\n")
	// **A realistic current file, and the realism is the load-bearing part.** It was a bare
	// `# Current` line until 2026-08-18, and that thinness hid a defect: `finish` overwrote
	// the whole file with a hardcoded template, and no fixture with nothing to lose could
	// show it. The preamble was added that morning — and the file *still* had nothing to lose
	// where the loss actually happened, because everything below the rule was the active item
	// and the verb was entitled to it. So this fixture carries standing sections **after** the
	// active one, which is where a real project keeps them and where 320 lines went. A fixture
	// contains only what its author thought to include, twice over.
	write("CURRENT.md", currentSrc)
	write("carves/x.md", carveSrc)
	write(".minispec/config.yaml", "track: private-trajectory\n")
	write(".gitignore", "PENDING.md\nCURRENT.md\nDONE.md\n.minispec/backup\n")
	for _, args := range [][]string{
		{"init", "-q"}, {"add", "-A"},
		{"-c", "user.email=t@example.com", "-c", "user.name=t", "commit", "-qm", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v: %s", err, out)
		}
	}
	return root
}

func read(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// R241 — both sides of the link are written in one invocation, so they cannot be created out
// of step. A queue entry pointing at a part that does not know it is queued reads as correct
// from the queue side, which is the side anyone checks first.
func TestBothSidesOfTheLinkAreWritten(t *testing.T) {
	root := fixture(t)
	got, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 4 {
		t.Errorf("minted #%d, want #4 — max() across both queue files", got.ID)
	}
	pending := read(t, root, "PENDING.md")
	if !strings.Contains(pending, "## 4. **a part to queue** (mini-spec).") {
		t.Errorf("the queue entry is missing:\n%s", pending)
	}
	if !strings.Contains(pending, "part `#7`") {
		t.Error("the entry carries no part pointer — the queue side of the link")
	}
	carve := read(t, root, "carves/x.md")
	if !strings.Contains(carve, "**NOT VERIFIED.** **OPEN (#4.)**") {
		t.Errorf("the part line is missing its marker, or lost its record:\n%s", carve)
	}
}

// R241 — the mint reads **both** queue files. Either alone collides: the pending file's
// maximum is too low right after completions, the done file's while the highest IDs are live.
func TestTheMintReadsBothQueueFiles(t *testing.T) {
	root := fixture(t)
	// The highest ID lives in the *done* file, which is the case a pending-only mint gets
	// wrong — and gets wrong plausibly, which is worse.
	done := read(t, root, "DONE.md") + "\n- **2026-08-01 — #9: an older completion.** (`abc1234`) Part `carves/x.md#1`.\n"
	if err := os.WriteFile(filepath.Join(root, "DONE.md"), []byte(done), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := addItem(root, "carves/x.md#7", "a part to queue", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != 10 {
		t.Errorf("minted #%d, want #10 — the maximum is in the done file", got.ID)
	}
}

// R243 — refuse rather than guess, and write nothing on the way out.
func TestAnUnresolvablePartIsRefusedAndNothingIsWritten(t *testing.T) {
	for _, tc := range []struct{ name, from, want string }{
		{"no such document", "carves/nope.md#7", "no document"},
		{"no such key", "carves/x.md#99", "no part keyed"},
		{"not a pointer", "carves/x.md", "is not <doc>#<part>"},
		// A part records exactly **one** item. Re-queuing would leave the older pointer
		// resolving to work it never described, and both entries would be individually
		// well-formed — so nothing downstream could detect it.
		{"already queued", "carves/x.md#1", "already carries queue ID #3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := fixture(t)
			before := read(t, root, "PENDING.md") + read(t, root, "carves/x.md")
			_, err := addItem(root, tc.from, "a title", "")
			if err == nil {
				t.Fatal("expected a refusal")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("refusal = %q, want it to mention %q", err, tc.want)
			}
			if after := read(t, root, "PENDING.md") + read(t, root, "carves/x.md"); after != before {
				t.Error("a refusal wrote something — it must name what it looked for and stop")
			}
		})
	}
}

// R244, R245 — the completion, all four surfaces.
func TestCompletionWritesAllFourSurfaces(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	got, err := Finish(root, 4, "abc1234", FinishOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Parts) != 1 || got.Parts[0].Key != "7" {
		t.Fatalf("completed parts = %+v, want the one the entry recorded", got.Parts)
	}

	carve := read(t, root, "carves/x.md")
	// Three markings at once, and the record superseding the transient while the
	// **assessment** stands. A rule selecting by what it writes would have taken the
	// NOT VERIFIED — the marking the format calls the most expensive to lose.
	want := "- [x] ~~**Item 7 — a part to queue.**~~ **NOT VERIFIED.** **LANDED (`abc1234`,"
	if !strings.Contains(carve, want) {
		t.Errorf("the completed part line is wrong:\n%s", carve)
	}
	if !strings.Contains(carve, "— `#4`.)**") {
		t.Error("the record carries no queue ID — the carve→queue join lives there and nowhere else")
	}
	if !strings.Contains(carve, "prose that must not move.") {
		t.Error("the completion moved prose")
	}
	if d := read(t, root, "DONE.md"); !strings.Contains(d, "#4: a part to queue.") || !strings.Contains(d, "Part `carves/x.md#7`.") {
		t.Errorf("the done entry header is wrong:\n%s", d)
	}
	if p := read(t, root, "PENDING.md"); strings.Contains(p, "## 4.") {
		t.Error("the entry is still in the pending file — an ID in both files is what R241 reports")
	}
	c := read(t, root, "CURRENT.md")
	if strings.Contains(c, "lots of active context") {
		t.Error("the current file kept its context — it is a resume buffer, never a log")
	}
	// **The preamble is the project's prose and the tool has no business rewriting it.** The
	// rule is the boundary: everything above it survives, everything below is the tool's to
	// clear. Overwriting the file with a template is the tool authoring, which the whole
	// mechanism is bounded against — structural edits only.
	if !strings.Contains(c, "This paragraph is the project's own prose and must survive a completion.") {
		t.Errorf("the completion rewrote the preamble:\n%s", c)
	}
	if !strings.Contains(c, "_No active item._") {
		t.Errorf("the active section was not reset to the format's empty state:\n%s", c)
	}
	// **Standing context sits below the active section, and that is where the loss happened.**
	// The preamble check above passed on the day 320 lines went, because the destroyed
	// sections were on the far side of the rule from it.
	for _, keep := range []string{"## Waiting on the user", "standing context", "## Left elsewhere", "a pointer to work outside this repository"} {
		if !strings.Contains(c, keep) {
			t.Errorf("the completion destroyed standing context (%q is gone):\n%s", keep, c)
		}
	}
	if strings.Contains(c, "a subheading the agent wrote") {
		t.Error("a `###` inside the active section survived — the section runs to the next heading of the same level or higher")
	}
}

// R247 — the tool writes the header and never the body.
func TestTheDoneEntryCarriesAHeaderAndNoBody(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(read(t, root, "DONE.md"), "\n") {
		if !strings.HasPrefix(l, "- **2026") {
			continue
		}
		// Everything on the header line is a fact the tool was handed. A body would be a
		// judgment about what a future reader needs, which is authoring.
		for _, fact := range []string{"#4:", "a part to queue.", "`abc1234`", "Part `carves/x.md#7`."} {
			if !strings.Contains(l, fact) {
				t.Errorf("header is missing %q:\n%s", fact, l)
			}
		}
	}
	if n := strings.Count(read(t, root, "DONE.md"), "\n- **"); n != 1 {
		t.Errorf("got %d entries, want 1", n)
	}
	// **The header is one line and nothing follows it**, which counting entries cannot see:
	// a body is a *continuation* of the entry rather than a second one. Added after an
	// injection that appended a summary line went silent against the count alone.
	lines := strings.Split(read(t, root, "DONE.md"), "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "- **2026") {
			continue
		}
		if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
			t.Errorf("the entry carries a body the tool wrote: %q\n"+
				"the body is a judgment about what a future reader needs, which is authoring",
				lines[i+1])
		}
	}
}

// R246 — a parent completes when its subparts do, derived rather than stored. A parent box
// would be a second copy of a fact the subparts already carry, and two copies disagree.
func TestAParentPartIsNeverChecked(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#8.2", "the second half", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err != nil {
		t.Fatal(err)
	}
	carve := read(t, root, "carves/x.md")
	if !strings.Contains(carve, "- [x] ~~**8.2 — the second half.**~~") {
		t.Errorf("the recorded subpart was not completed:\n%s", carve)
	}
	if !strings.Contains(carve, "- [ ] **8.1 — the first half.** **OPEN (not queued.)**") {
		t.Error("a sibling was touched — completion checks the parts the item recorded and nothing else")
	}
	if !strings.Contains(carve, "- **Item 8 — a split parent.** **SPLIT (Bill, 2026-08-14.)**") {
		t.Error("the parent line changed — a parent keeps no checkbox, and inventing one asserts what the document declined to say")
	}
}

// R244 — the source first, and the failure case is what proves the order. If the carve write
// fails, the queue must be untouched: the carve is the copy a future reader trusts.
func TestASourceFailureLeavesTheQueueUntouched(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	before := read(t, root, "PENDING.md")
	if err := os.Remove(filepath.Join(root, "carves", "x.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err == nil {
		t.Fatal("expected the completion to fail when its source cannot be written")
	}
	if after := read(t, root, "PENDING.md"); after != before {
		t.Error("the queue moved even though the source write failed — a part left open " +
			"against an item already in the done file is exactly the orphan R242 reports")
	}
	if d := read(t, root, "DONE.md"); strings.Contains(d, "#4:") {
		t.Error("a done entry was written for a completion that did not happen")
	}
}

// Skipped for half a day naming gaps O14 and O15 — the dependency's Place-at-end landed after
// trailing commentary and left a newline Remove did not take back — and un-skipped the same day
// their fix landed (mini-spec-tool #29).
// The two verbs are **inverses on the queue file**, which nothing asked for and which falls
// out of the separator belonging to the entry rather than to its neighbour. Found by running
// them back to back, 2026-08-18.
func TestAddThenFinishLeavesTheQueueFileByteIdentical(t *testing.T) {
	root := fixture(t)
	before := read(t, root, "PENDING.md")
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err != nil {
		t.Fatal(err)
	}
	if after := read(t, root, "PENDING.md"); after != before {
		t.Errorf("the queue file did not come back:\n  before %q\n  after  %q", before, after)
	}
}

// R248 — the whole invocation is one slot transaction, so a mis-typed pointer is a revert
// rather than a repair. These verbs are the slot's first production callers.
func TestTheInvocationIsOneSlotTransaction(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	stamp := filepath.Join(root, ".minispec", "backup", "stamp")
	if _, err := os.Stat(stamp); err != nil {
		t.Fatalf("the slot was not engaged: %v — Record wraps the verb, or nothing does", err)
	}
	for _, f := range []string{"PENDING.md", "CURRENT.md", "DONE.md"} {
		if _, err := os.Stat(filepath.Join(root, ".minispec", "backup", f)); err != nil {
			t.Errorf("%s was not backed up: %v", f, err)
		}
	}
	// **One Record, not one per write** — and the only witness is a revert. With a Record per
	// write the second one's backup already holds the new entry, so a revert puts back a queue
	// the invocation had already changed. The August test stopped at the stamp and could not
	// tell the two apart; measured 2026-09-05 when its injection stayed silent.
	if err := backup.New(root).Revert(); err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, "PENDING.md"); strings.Contains(got, "a part to queue") {
		t.Errorf("revert left the new entry in place — the slot covered a fragment of the invocation, not the whole:\n%s", got)
	}
}

// A current file with no `## Active` heading is **refused**, not guessed at, and the refusal
// reaches the caller through the completion. R250
//
// Without the heading there is no region the tool can name, and clearing *the rest of the
// file* instead is exactly how 320 lines of standing context went. Reporting absence as an
// error rather than as silence is this project's standing rule, applied to its own file.
func TestACurrentFileWithNoActiveHeadingIsRefused(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "CURRENT.md"), []byte("# Current\n\n---\n\nno heading here.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Finish(root, 4, "abc1234", FinishOpts{})
	if err == nil {
		t.Fatal("expected a refusal — the tool cannot tell the active item from the standing context")
	}
	if !strings.Contains(err.Error(), "carries no `## Active` heading") {
		t.Errorf("refusal = %q, want it to name the missing heading and the repair", err)
	}
}

// R232, R233 — a completed part is not reopened by the next queue mutation.
//
// **This is the 2026-08-18 corruption, reproduced.** `releaseAttempt` ran on every mutation
// and marked every entry present in the backup and absent from the live pending file as an
// abandoned attempt — and a completion removes the entry exactly as a revert does, so the
// first mutation after any completion wrote `**OPEN (not queued.)**` over the landed part.
// It happened to `carves/trajectory-tool.md` Item 6 in a tracked public file.
func TestACompletedPartIsNotReopenedByTheNextMutation(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err != nil {
		t.Fatal(err)
	}
	landed := read(t, root, "carves/x.md")
	if !strings.Contains(landed, "LANDED") {
		t.Fatalf("the completion did not mark the part landed:\n%s", landed)
	}

	// Any new mutation. This is the step that used to reopen it.
	if _, err := addItem(root, "carves/x.md#8.1", "another part", ""); err != nil {
		t.Fatal(err)
	}
	after := read(t, root, "carves/x.md")
	for _, line := range strings.Split(after, "\n") {
		if !strings.Contains(line, "7") {
			continue
		}
		if strings.Contains(line, "OPEN (not queued.)") {
			t.Errorf("the next mutation reopened a completed part:\n  %s", strings.TrimSpace(line))
		}
		if !strings.Contains(line, "LANDED") || !strings.Contains(line, "[x]") {
			t.Errorf("the completed part lost its landed state:\n  %s", strings.TrimSpace(line))
		}
	}
}

// R232 in isolation — the state guard, where the landed guard cannot reach.
//
// **Written because the alarm above went silent under R232's injection alone.** Both guards
// stop the measured failure, so that test proves only the *pair*. This one departs an entry
// the tool did not remove — a hand edit — leaving the part **unchecked**, so R233 has nothing
// to refuse and only the state guard is left standing.
//
// *The first draft of this test was also silent, and the reason is worth keeping:* after a
// single `add-item` the backup **predates** that entry, so `departed(backup, live)` is empty
// however the entry left and the release path had nothing to consider either way. It takes a
// second mutation to move the entry into the backup before the hand edit can depart it. A
// test that cannot reach the code it names is the thing the injection exists to find.
//
// The behaviour it pins is that the tool does not infer an abandoned attempt from a diff it
// did not cause. The slot is in `changed`, not `reverted`, so there is no attempt to release.
func TestAHandRemovedEntryIsNotTreatedAsAnAbandonedAttempt(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#8.1", "a part to queue", ""); err != nil {
		t.Fatal(err)
	}
	// A second mutation, so the backup now holds the first entry.
	if _, err := addItem(root, "carves/x.md#7", "another part", ""); err != nil {
		t.Fatal(err)
	}

	// A hand edit removes the first entry. The part stays `[ ]`, so R233 does not apply, and
	// the backup holds it while the live file does not — the exact shape a revert produces.
	pending := read(t, root, "PENDING.md")
	i := strings.Index(pending, "## 4.")
	if i < 0 {
		t.Fatal("the entry is not in the pending file")
	}
	j := strings.Index(pending[i:], "\n\n")
	if err := os.WriteFile(filepath.Join(root, "PENDING.md"), []byte(pending[:i]+pending[i+j+2:]), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := addItem(root, "carves/x.md#8.2", "a third part", ""); err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(read(t, root, "carves/x.md"), "\n") {
		if strings.Contains(line, "8.1") && strings.Contains(line, "not queued") {
			t.Errorf("a hand-removed entry was read as an abandoned attempt and the part was released:\n  %s",
				strings.TrimSpace(line))
		}
	}
}

// addItem is AddItem with the prose slots and the placement a case does not care about, so the
// cases about the *link* stay about the link. The slots and the placement have their own.
func addItem(root, from, title, skill string) (Created, error) {
	return AddItem(root, "", from, parser.Entry{Title: title, Skill: skill, Status: "a one-line status."}, parser.Place{})
}

// idle replaces the fixture's active block with the format's placeholder, which is the one
// thing `--next` reads.
func idle(t *testing.T, root string) {
	t.Helper()
	if err := parser.ResetCurrent(filepath.Join(root, "CURRENT.md")); err != nil {
		t.Fatal(err)
	}
}

func placed(t *testing.T, root, from, title string, place parser.Place) (Created, error) {
	t.Helper()
	return AddItem(root, "", from, parser.Entry{Title: title, Status: "Where this stands."}, place)
}

// R257 — `--next` means **next to be worked**, not position 1.
//
// The alarm: resolve `--next` to position 1 unconditionally, which is the `--first` this flag
// replaced. Red is the new entry displacing the item being worked and making itself active —
// exactly the objection `--next` was designed to dissolve.
func TestNextMeansNextToBeWorked(t *testing.T) {
	busy := fixture(t)
	got, err := placed(t, busy, "carves/x.md#7", "a part to queue", parser.Place{Kind: parser.PlaceNext})
	if err != nil {
		t.Fatal(err)
	}
	if got.Position != 2 {
		t.Errorf("with a step in progress --next placed at %d, want 2 — it displaced the active item", got.Position)
	}

	free := fixture(t)
	idle(t, free)
	got, err = placed(t, free, "carves/x.md#7", "a part to queue", parser.Place{Kind: parser.PlaceNext})
	if err != nil {
		t.Fatal(err)
	}
	if got.Position != 1 {
		t.Errorf("with nothing in progress --next placed at %d, want 1", got.Position)
	}
}

// R258 — `--nth 1` is refused while a step is in progress, and the message names **both**
// repairs.
//
// The alarm: name only `--next`. Red is a refusal offering one repair, which makes preempting
// the active item read as forbidden — a real thing to want, refused by omission.
func TestNthOneRefusedWhileAStepIsInProgress(t *testing.T) {
	root := fixture(t)
	_, err := placed(t, root, "carves/x.md#7", "a part to queue", parser.Place{Kind: parser.PlaceNth, N: 1})
	if err == nil {
		t.Fatal("--nth 1 was accepted while a step was in progress")
	}
	for _, repair := range []string{"--next", "park"} {
		if !strings.Contains(err.Error(), repair) {
			t.Errorf("the refusal never mentions %q, so one repair reads as forbidden: %v", repair, err)
		}
	}
	if before := read(t, root, "PENDING.md"); strings.Contains(before, "#4") {
		t.Errorf("a refused placement wrote anyway:\n%s", before)
	}
	idle(t, root)
	if _, err := placed(t, root, "carves/x.md#7", "a part to queue", parser.Place{Kind: parser.PlaceNth, N: 1}); err != nil {
		t.Errorf("--nth 1 with nothing in progress: %v", err)
	}
}

// CRC: crc-Pending.md | R262, R263
// The body lands beneath the header the same invocation minted, in one write.
func TestFinishPlacesTheBodyBeneathTheHeader(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	body := "  A first line, mentioning `a code span`.\n  A second line.\n"
	if _, err := Finish(root, 4, "abc1234", FinishOpts{Body: body}); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(root, "DONE.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	at := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "- **") {
			at = i
			break
		}
	}
	if at < 0 {
		t.Fatalf("no done header in:\n%s", src)
	}
	want := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	for k, w := range want {
		if at+1+k >= len(lines) || lines[at+1+k] != w {
			t.Fatalf("line %d beneath the header = %q, want %q\nfile:\n%s",
				k+1, lines[min(at+1+k, len(lines)-1)], w, src)
		}
	}
}

// CRC: crc-Pending.md | R264
// With no body the verb behaves exactly as it did, and writes nothing extra.
func TestFinishWithNoBodyWritesNoBody(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{}); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile(filepath.Join(root, "DONE.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, "- **") {
			continue
		}
		if i+1 < len(lines) && lines[i+1] != "" {
			t.Errorf("line beneath the header = %q, want it empty", lines[i+1])
		}
	}
}

// CRC: crc-Pending.md | R265
// The identity line comes from the queue entry; the context comes from the caller.
func TestStartWritesTheActiveSection(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	// The fixture opens with an item already active, which is `finish`'s business, not this one's.
	if err := parser.ResetCurrent(filepath.Join(root, "CURRENT.md")); err != nil {
		t.Fatal(err)
	}
	got, err := Start(root, 4, "  context the agent composed, with `a code span`.")
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "a part to queue" {
		t.Errorf("Title = %q, want the queue entry's", got.Title)
	}
	src, err := os.ReadFile(filepath.Join(root, "CURRENT.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"`#4` — a part to queue",
		"  context the agent composed, with `a code span`.",
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("the active section is missing %q:\n%s", want, src)
		}
	}
	if strings.Contains(string(src), "_No active item._") {
		t.Error("the placeholder survived an opening")
	}
	// The standing context after the section is what 320 lines went out of.
	if !strings.Contains(string(src), "must survive a completion") {
		t.Errorf("standing context did not survive the opening:\n%s", src)
	}
}

// CRC: crc-Pending.md | R266
// Opening over an item discards it, in a file with no diff — so it is refused.
func TestStartRefusesAnOccupiedActiveSection(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(root, "CURRENT.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = Start(root, 4, "context that must not land")
	if err == nil {
		t.Fatal("an occupied `## Active` was overwritten without complaint")
	}
	if !strings.Contains(err.Error(), "park") {
		t.Errorf("the refusal does not name the repair: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(root, "CURRENT.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Error("the refusal still wrote to the file")
	}
}

// CRC: crc-Pending.md | R268
// The slot carries the queue ID the tool owns joined to what the caller says was discharged.
func TestTheIdentifierSlotCarriesWhatTheItemDischarged(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	if _, err := Finish(root, 4, "abc1234", FinishOpts{Discharged: "R268–R269"}); err != nil {
		t.Fatal(err)
	}
	done := read(t, root, "DONE.md")
	if !strings.Contains(done, "— #4 / R268–R269: a part to queue.") {
		t.Errorf("the identifier slot was written half-filled:\n%s", done)
	}
}

// CRC: crc-Pending.md | R269
// A colon ends the slot early, so one inside it is refused rather than written.
func TestAColonInTheIdentifierSlotIsRefused(t *testing.T) {
	root := fixture(t)
	if _, err := addItem(root, "carves/x.md#7", "a part to queue", "mini-spec"); err != nil {
		t.Fatal(err)
	}
	before := read(t, root, "DONE.md")
	_, err := Finish(root, 4, "abc1234", FinishOpts{Discharged: "R1: and more"})
	if err == nil {
		t.Fatal("a colon in the slot was written, and every identifier after it stops being read")
	}
	if !strings.Contains(err.Error(), "colon") {
		t.Errorf("the refusal does not name the problem: %v", err)
	}
	if read(t, root, "DONE.md") != before {
		t.Error("the refusal still wrote the entry")
	}
}

// withGaps adds a design root holding two gaps, and answers the path `add-item` is handed.
// Kept out of `fixture` deliberately: every part-pointer test must keep working in a tree with
// no design root at all, which is the layout `pending` exists for.
func withGaps(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, "design")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := "# Design\n\n## Gaps\n\n- [ ] O5: a gap an item may repair\n- [ ] O6: another\n"
	path := filepath.Join(dir, "design.md")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func addGapItem(root, gaps, from string) (Created, error) {
	return AddItem(root, gaps, from, parser.Entry{Title: "repair a gap", Skill: "mini-spec",
		Status: "a one-line status."}, parser.Place{})
}

// R271, R274. A gap is a source, and **nothing is written on its side** — the part case writes
// both halves because the carve is public and the queue private; a gap has no marker and gains
// none. The design document must come back byte for byte.
func TestAGapIsASourceAndNothingIsWrittenOnItsSide(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	before, err := os.ReadFile(gaps)
	if err != nil {
		t.Fatal(err)
	}

	got, err := addGapItem(root, gaps, "O5")
	if err != nil {
		t.Fatal(err)
	}
	if got.Part.Kind != minispecsdom.SourceGap || got.Part.Key != "O5" {
		t.Errorf("source read as %+v, want a gap keyed O5", got.Part)
	}
	pending := read(t, root, "PENDING.md")
	if !strings.Contains(pending, "gap `O5`") {
		t.Errorf("the entry carries no gap pointer:\n%s", pending)
	}
	if strings.Contains(pending, "part `#O5`") {
		t.Error("a gap was written as a part pointer — the `#` belongs to part keys only")
	}
	after, err := os.ReadFile(gaps)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Errorf("the gap side was written to:\n%s", after)
	}
	// The files it reports are the files it wrote, and design.md is not among them.
	for _, f := range got.Files {
		if strings.Contains(f, "design.md") {
			t.Errorf("design.md reported as written: %v", got.Files)
		}
	}
}

// R272. A range is well-formed in the ID grammar and is not a source. The refusal names the
// **rule** rather than the syntax, because the syntax is fine and the decision is what the
// caller needs told.
func TestAGapRangeIsRefusedByTheOnePointerRule(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	_, err := addGapItem(root, gaps, "O5-O6")
	if err == nil {
		t.Fatal("a range was accepted as a source")
	}
	if !strings.Contains(err.Error(), "one pointer") {
		t.Errorf("the refusal does not name the rule: %v", err)
	}
}

// R272. A gap ID needs a design root, and its absence is a refusal naming what was looked for
// — never a search, since a repository may hold several roots.
func TestAGapSourceWithNoDesignRootIsRefused(t *testing.T) {
	root := fixture(t)
	_, err := addGapItem(root, "", "O5")
	if err == nil {
		t.Fatal("a gap source was accepted with no design root")
	}
	if !strings.Contains(err.Error(), "design root") || !strings.Contains(err.Error(), "O5") {
		t.Errorf("the refusal names neither what it needed nor what it looked for: %v", err)
	}
}

// R271. An ID the document holds no gap for is a refusal, not a queue entry pointing nowhere.
func TestAnUnknownGapIsRefused(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	_, err := addGapItem(root, gaps, "O99")
	if err == nil {
		t.Fatal("an unknown gap was accepted")
	}
	if !strings.Contains(err.Error(), "O99") {
		t.Errorf("the refusal does not name the gap: %v", err)
	}
	if strings.Contains(read(t, root, "PENDING.md"), "O99") {
		t.Error("a refused source still placed an entry")
	}
}

// R271, R275. The pointer round-trips, and a gap is **not** a part: `Parts()` must stay empty
// or completion would open design.md as a carve and mark a part landed in it.
func TestAGapSourcedEntryRoundTripsAndIsNotAPart(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	got, err := addGapItem(root, gaps, "O5")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := parser.PendingEntries(filepath.Join(root, "PENDING.md"))
	if err != nil {
		t.Fatal(err)
	}
	var e *parser.QueueEntry
	for i := range entries {
		if entries[i].ID == got.ID {
			e = &entries[i]
		}
	}
	if e == nil {
		t.Fatalf("the entry for #%d did not read back", got.ID)
	}
	if len(e.Parts()) != 0 {
		t.Errorf("a gap read back as %d part(s) — completion would mark a part landed in design.md", len(e.Parts()))
	}
	ref, ok := e.Gap()
	if !ok || ref.Key != "O5" {
		t.Errorf("Gap() = %+v, %v; want O5", ref, ok)
	}
}

// R277, R278. Completion leaves the gap open unless asked, and the done entry keeps the
// pointer either way — without it the link survives only in whatever `--discharged` carried.
func TestFinishLeavesTheGapOpenUnlessAsked(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	got, err := addGapItem(root, gaps, "O5")
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(gaps)
	if err != nil {
		t.Fatal(err)
	}
	// R279 — neither flag is a refusal, **before anything is written**.
	if _, err := Finish(root, got.ID, "abc1234", FinishOpts{Body: "x"}); err == nil {
		t.Fatal("a gap-sourced item completed with no decision about its gap")
	} else if !strings.Contains(err.Error(), "--no-resolve") || !strings.Contains(err.Error(), "--resolve") {
		t.Errorf("the refusal does not name both spellings: %v", err)
	}
	if strings.Contains(read(t, root, "DONE.md"), "repair a gap") {
		t.Error("the refusal still moved the entry — it must land before anything is written")
	}

	done, err := Finish(root, got.ID, "abc1234", FinishOpts{Body: "x", DeclineResolve: true})
	if err != nil {
		t.Fatal(err)
	}
	if !done.HasGap || done.GapResolved {
		t.Errorf("HasGap=%v GapResolved=%v; want the gap reported and left open", done.HasGap, done.GapResolved)
	}
	after, err := os.ReadFile(gaps)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Error("the gap was resolved under --no-resolve")
	}
	if d := read(t, root, "DONE.md"); !strings.Contains(d, "Gap `design/design.md#O5`") {
		t.Errorf("the done entry lost the gap pointer:\n%s", d)
	}
}

// R277. With the act supplied, the gap is resolved — source first, exactly as a part is.
func TestFinishResolvesTheGapWhenAsked(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	got, err := addGapItem(root, gaps, "O5")
	if err != nil {
		t.Fatal(err)
	}
	var asked string
	done, err := Finish(root, got.ID, "abc1234", FinishOpts{
		Body:       "x",
		ResolveGap: func(g parser.PartRef) error { asked = g.Key; return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if asked != "O5" {
		t.Errorf("resolved %q, want O5", asked)
	}
	if !done.GapResolved {
		t.Error("GapResolved is false after a successful resolve")
	}
}

// R271. A near-miss gap ID reports **its own** parse error. Falling through to the part path
// hands the caller `part pointer "O22-R5" is not <doc>#<part>` — a syntax they did not use —
// while discarding the precise message the ID grammar already produced.
func TestANearMissGapIDReportsItsOwnRefusal(t *testing.T) {
	root := fixture(t)
	gaps := withGaps(t, root)
	_, err := addGapItem(root, gaps, "O22-R5")
	if err == nil {
		t.Fatal("a range crossing types was accepted")
	}
	if strings.Contains(err.Error(), "<doc>#<part>") {
		t.Errorf("a gap-shaped token was refused as a part pointer: %v", err)
	}
	if !strings.Contains(err.Error(), "one pointer") {
		t.Errorf("the gap grammar's own refusal was discarded: %v", err)
	}
	// A genuine part pointer keeps the part path's refusal, or the branch would have moved.
	if _, err := addGapItem(root, gaps, "carves/x.md"); err == nil ||
		!strings.Contains(err.Error(), "<doc>#<part>") {
		t.Errorf("a malformed part pointer no longer gets the part refusal: %v", err)
	}
}
