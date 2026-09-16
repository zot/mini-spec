// CRC: crc-CLI.md | R241, R476
package cli

import (
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/pending"
)

// discard swallows a flagset's usage output, which is noise in a unit test.
type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

// Flags are parsed wherever they sit among the positional arguments.
//
// Go's flag package stops at the first non-flag argument, so
// `add-item --from x "a title" --skill mini-spec` folded the flag into the title — silently,
// and in the exact order the command's own usage line advertises. The corrupted title then
// rode into the done entry on completion.
//
// **The same trap had already been found and fixed in `finish` an hour earlier** and the fix
// was not carried across, which is why this is one shared function rather than two guards:
// the duplication was not worth removing until a defect proved the *behaviour* should be
// shared, not the boilerplate.
func TestFlagsAreParsedWhereverTheySit(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		wantPos   string
		wantSkill string
	}{
		{"flag after the positional", []string{"--skill=x", "a", "title"}, "a title", "x"},
		{"flag before the positional", []string{"--skill", "x", "a", "title"}, "a title", "x"},
		{"flag between positionals", []string{"a", "--skill", "x", "title"}, "a title", "x"},
		{"flag last", []string{"a", "title", "--skill", "x"}, "a title", "x"},
		{"no flag at all", []string{"a", "title"}, "a title", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := flag.NewFlagSet("t", flag.ContinueOnError)
			fs.SetOutput(discard{})
			skill := fs.String("skill", "", "")
			pos, err := parseFlagsAnywhere(fs, tc.args)
			if err != nil {
				t.Fatal(err)
			}
			if got := strings.Join(pos, " "); got != tc.wantPos {
				t.Errorf("positional = %q, want %q — a flag was folded into it", got, tc.wantPos)
			}
			if *skill != tc.wantSkill {
				t.Errorf("skill = %q, want %q — the flag was not seen", *skill, tc.wantSkill)
			}
		})
	}
}

// R253 — the status slot is **required**, because it sits inside the heading the verb mints.
//
// Fire alarm: default the status to the empty string. Red is the heading stopping after the
// skill while the verb reports success — the original defect restored as a default, which is
// how a required field quietly becomes optional.
func TestStatusSlotIsRequired(t *testing.T) {
	c := &CLI{}
	var code int
	// **The exit code alone proves nothing here**, and the injection is what said so: with the
	// guard removed the verb still exits 1, because it goes on to fail on a repository that
	// does not exist. The refusal has to be read, not counted.
	said := capturingStderr(t, func() {
		code = c.runAddItem(t.TempDir(), []string{"--from", "carves/x.md#7", "a title"})
	})
	if code == 0 {
		t.Error("add-item accepted a missing --status and reported success")
	}
	for _, want := range []string{"--status", "--status-file"} {
		if !strings.Contains(said, want) {
			t.Errorf("the refusal never names %s, so the verb failed for some other reason: %s", want, said)
		}
	}
}

// capturingStderr runs fn with os.Stderr redirected, and returns what it wrote.
func capturingStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	was := os.Stderr
	os.Stderr = w
	fn()
	os.Stderr = was
	w.Close()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// R254 — a prose slot is supplied inline **or** from a file, and never both.
//
// Fire alarm: let the file form win silently. A precedence rule is a silent choice between two
// things the caller believed were one.
func TestProseSlotGivenTwiceIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "status.txt")
	if err := os.WriteFile(path, []byte("from the file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveSlot("status", true, true, "inline", path)
	if err == nil {
		t.Fatalf("both forms were accepted and %q was used with nothing reporting the other", got)
	}
	if !strings.Contains(err.Error(), "--status") || !strings.Contains(err.Error(), "--status-file") {
		t.Errorf("the refusal does not name both forms: %v", err)
	}
}

// R254 — a file slot is read **byte for byte**, which is why the slot exists at all.
//
// Fire alarm: trim or re-quote the file's contents. Red is the backticked identifier coming
// back altered — the loss this slot makes unreachable, arriving from inside the tool instead of
// from the shell.
func TestBodyFileSlotIsReadByteForByte(t *testing.T) {
	const body = "the `**OPEN (#N.)**` marker isn't rewritten — *and* `$HOME` survives"
	path := filepath.Join(t.TempDir(), "status.txt")
	// The trailing newline is a heredoc's, not the prose's, and it is the only byte that comes
	// off. Nothing else is trimmed, re-quoted, or normalised.
	if err := os.WriteFile(path, []byte(body+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := resolveSlot("status", false, true, "", path)
	if err != nil {
		t.Fatal(err)
	}
	if got != body {
		t.Errorf("the file's bytes came back altered:\n  got  %q\n  want %q", got, body)
	}
}

// R256 — the four placement flags are mutually exclusive, and giving two is refused.
//
// Fire alarm: let the last one win. A precedence between placements silently puts the entry
// somewhere the caller did not ask for, in a file with no diff.
func TestPlacementFlagsAreMutuallyExclusive(t *testing.T) {
	if _, err := placeFrom(map[string]bool{"next": true, "nth": true}, 3, 0); err == nil {
		t.Error("--next and --nth were both accepted")
	}
	if _, err := placeFrom(map[string]bool{"last": true, "after": true}, 0, 4); err == nil {
		t.Error("--last and --after were both accepted")
	}
	for _, tc := range []struct {
		name  string
		gave  map[string]bool
		nth   int
		after int
		want  parser.Place
	}{
		{name: "nothing at all", gave: map[string]bool{}, want: parser.Place{Kind: parser.PlaceLast}},
		{name: "--last spelled out", gave: map[string]bool{"last": true}, want: parser.Place{Kind: parser.PlaceLast}},
		{name: "--next", gave: map[string]bool{"next": true}, want: parser.Place{Kind: parser.PlaceNext}},
		{name: "--nth", gave: map[string]bool{"nth": true}, nth: 2, want: parser.Place{Kind: parser.PlaceNth, N: 2}},
		{name: "--after", gave: map[string]bool{"after": true}, after: 9, want: parser.Place{Kind: parser.PlaceAfter, N: 9}},
		// **A zero is a value, not an absence.** `--nth 0` must reach `resolvePlace`, whose
		// refusal exists precisely so a position the caller was specific about is never
		// reinterpreted — and could never be reached while `nth != 0` meant "given".
		{name: "--nth 0", gave: map[string]bool{"nth": true}, nth: 0, want: parser.Place{Kind: parser.PlaceNth, N: 0}},
		{name: "--after 0", gave: map[string]bool{"after": true}, after: 0, want: parser.Place{Kind: parser.PlaceAfter, N: 0}},
	} {
		got, err := placeFrom(tc.gave, tc.nth, tc.after)
		if err != nil || got != tc.want {
			t.Errorf("%s: got %+v, %v; want %+v", tc.name, got, err, tc.want)
		}
	}
}

// CRC: crc-CLI.md | R262, R263
// The resolved body reaches the file, which is the seam a flag can be parsed and dropped at.
//
// Fire alarm: pass "" to pending.Finish instead of the resolved text in runFinish. Red is the
// body's line missing from DONE.md while the command still exits 0 and reports a completion —
// a silent drop that reads as success, into a file git shows no diff for.
func TestFinishWiresTheResolvedBodyThrough(t *testing.T) {
	dir := t.TempDir()
	for _, d := range []string{"carves", ".minispec", ".git"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("PENDING.md", "# Pending\n\n---\n\n## 4. **a part to queue** (mini-spec). Status.\n"+
		"Source: [carves/x.md](carves/x.md), part `#7`.\n")
	write("DONE.md", "# Done\n\n---\n")
	write("CURRENT.md", "# Current\n\n## Active\n\n`#4` — a part to queue.\n")
	write("carves/x.md", "# Carve: x\n\n## Status\n\n- [ ] **Item 7 — a part to queue.** **OPEN (#4.)**\n")

	const body = "  A body carrying `**OPEN (#N.)**` and a $HOME that must survive."
	bodyPath := filepath.Join(dir, "body.md")
	write("body.md", body+"\n")

	c := &CLI{}
	if code := c.runFinish(dir, []string{"4",
		"--body-file", bodyPath, "--discharged", "R268–R269"}); code != 0 {
		t.Fatalf("runFinish exited %d, want 0", code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "DONE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), body) {
		t.Errorf("the body never reached DONE.md:\n%s", got)
	}
	// R268 — the same seam one field along, and it fails the same silent way.
	if !strings.Contains(string(got), "— #4 / R268–R269:") {
		t.Errorf("the identifier slot never reached DONE.md:\n%s", got)
	}
}

// R281. A gap left open **by decision** and one left open by oversight are identical in
// `design.md` — the checkbox is unticked either way — and the completion is the only place that
// difference is known. So the report says which happened, the same move `NOT VERIFIED` makes
// rather than leaving a part unmarked.
func TestAGapLeftOpenIsRecordedAsADecision(t *testing.T) {
	done := pending.Finished{
		ID:     7,
		Files:  []string{"PENDING.md", "DONE.md"},
		Gap:    parser.PartRef{Doc: "tool/design/design.md", Key: "O136", Kind: minispecsdom.SourceGap},
		HasGap: true,
	}
	out := captureStdout(t, func() { reportFinished("", done, "a body", false) })

	for _, want := range []string{"left open", "O136", "--no-resolve"} {
		if !strings.Contains(out, want) {
			t.Errorf("the record is missing %q:\n%s", want, out)
		}
	}
	// It must read as a decision, not as an omission the reader is being nagged about.
	if strings.Contains(out, "still open") {
		t.Errorf("a deliberate choice is reported as an oversight:\n%s", out)
	}
}

// R277. A resolved gap is reported as resolved, and the notice does not also fire — a
// contradiction in one report is worse than either half.
func TestAResolvedGapIsReportedAndTheNoticeStaysQuiet(t *testing.T) {
	done := pending.Finished{
		ID:          7,
		Files:       []string{"tool/design/design.md", "PENDING.md", "DONE.md"},
		Gap:         parser.PartRef{Doc: "tool/design/design.md", Key: "O136", Kind: minispecsdom.SourceGap},
		HasGap:      true,
		GapResolved: true,
	}
	out := captureStdout(t, func() { reportFinished("", done, "a body", false) })

	if !strings.Contains(out, "resolved gap O136") {
		t.Errorf("a resolved gap is not reported:\n%s", out)
	}
	if strings.Contains(out, "still open") {
		t.Errorf("the notice fired over a resolved gap:\n%s", out)
	}
}

// R280. The report from the other side: `--resolve` on an item that names no gap resolved
// nothing and, until this, said nothing — the caller typed a flag and got a report that reads
// exactly like one where it worked.
func TestResolveOnAPartSourcedItemSaysItDidNothing(t *testing.T) {
	done := pending.Finished{ID: 9, Files: []string{"PENDING.md"},
		Parts: []parser.PartRef{{Doc: "carves/x.md", Key: "3"}}}
	out := captureStdout(t, func() { reportFinished("", done, "a body", true) })

	if !strings.Contains(out, "resolve flag did nothing") {
		t.Errorf("a flag that did nothing said nothing:\n%s", out)
	}
	if strings.Contains(out, "left open") {
		t.Errorf("the left-open record fired over an item with no gap:\n%s", out)
	}
}

// R277. The gap resolved must be the gap the **entry** named. The project comes from the
// working directory and the ref from wherever the item was created; in a repository with two
// design roots those differ, and resolving the same ID in the wrong document would close a gap
// nobody asked about while both files still validated.
func TestResolveRefusesAGapInAnotherDesignRoot(t *testing.T) {
	called := false
	resolve := gapResolver(func(string) error { called = true; return nil },
		filepath.Join("/repo", "tool", "design", "design.md"), "/repo", 9)

	err := resolve(parser.PartRef{Doc: "example/design/design.md", Key: "O5", Kind: minispecsdom.SourceGap})
	if err == nil {
		t.Fatal("a gap in another design root was resolved")
	}
	if called {
		t.Error("the resolve act ran despite the mismatch")
	}
	for _, want := range []string{"O5", "example/design/design.md", "tool/design/design.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	// And the matching case still resolves, or the guard would be refusing everything.
	if err := resolve(parser.PartRef{Doc: "tool/design/design.md", Key: "O5", Kind: minispecsdom.SourceGap}); err != nil {
		t.Fatalf("the entry's own design root was refused: %v", err)
	}
	if !called {
		t.Error("a matching root did not reach the resolve act")
	}
}

// R280. Contradictory intents are refused together, the way every paired slot on these verbs
// is: nothing preserves an ordering between them, and honouring one would be the tool deciding
// what the caller meant. Refused **before** the item is even looked up.
func TestResolveAndNoResolveAreRefusedTogether(t *testing.T) {
	var code int
	out := capture(t, &os.Stderr, func() {
		code = (&CLI{}).runFinish(t.TempDir(), []string{"1", "--resolve", "--no-resolve"})
	})
	if code == 0 {
		t.Error("contradictory flags were accepted")
	}
	// **The exit code is not the witness**, and that is why this case asserts on the message:
	// accepting the flags leaves the verb to fail on something else in this fixture, so the
	// code is 1 either way and only the refusal text tells the two apart.
	if !strings.Contains(out, "--resolve") || !strings.Contains(out, "--no-resolve") {
		t.Errorf("the refusal does not name both flags: %s", out)
	}
}

// R329 — the written line says which file is the tracked public document and which are
// ignored, so the carve flip a completion leaves behind is marked as the write that still
// needs a commit. A real repository, because the words come from git.
func TestWrittenLineSaysWhichWriteIsTracked(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Skipf("git unavailable: %v: %s", err, out)
		}
	}
	for _, d := range []string{"carves", ".minispec"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(rel, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(".gitignore", "PENDING.md\nCURRENT.md\nDONE.md\n.minispec/\n")
	write("PENDING.md", "# Pending\n\n---\n\n## 4. **a part to queue** (mini-spec). Status.\nSource: [carves/x.md](carves/x.md), part `#7`.\n")
	write("DONE.md", "# Done\n\n---\n")
	write("CURRENT.md", "# Current\n\n## Active\n\n`#4` — a part to queue.\n")
	write("carves/x.md", "# Carve: x\n\n## Status\n\n- [ ] **Item 7 — a part to queue.** **OPEN (#4.)**\n")
	run("init", "-q")
	run("add", ".gitignore", "carves/x.md")
	run("commit", "-qm", "init")

	out := captureStdout(t, func() {
		if code := (&CLI{}).runFinish(dir, []string{"4"}); code != 0 {
			t.Fatalf("runFinish exited %d", code)
		}
	})
	for _, want := range []string{"carves/x.md (tracked, uncommitted)", "CURRENT.md (ignored)", "PENDING.md (ignored)", "DONE.md (ignored)"} {
		if !strings.Contains(out, want) {
			t.Errorf("the written line does not say %q:\n%s", want, out)
		}
	}
}
