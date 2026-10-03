// CRC: crc-Update.md | Test: test-Update.md | R20, R21, R533, R534, R535, R536, R537
package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zot/minispec/internal/project"
)

// card writes crc-Store.md with body and returns an Update over its design directory.
func card(t *testing.T, body string) (*Update, string) {
	t.Helper()
	design := t.TempDir()
	path := filepath.Join(design, "crc-Store.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Update{Project: &project.Project{RootPath: design, DesignDir: design}}, path
}

// reqLine is the card's Requirements line, or "" when it has none.
func reqLine(t *testing.T, path string) string {
	t.Helper()
	for _, line := range strings.Split(read(t, path), "\n") {
		if strings.HasPrefix(line, "**Requirements:**") {
			return line
		}
	}
	return ""
}

// R20, R533 — the designed cases, all already canonical.
func TestAddRefAndRemoveRefDesignedCases(t *testing.T) {
	for _, c := range []struct{ body, op, ref, want string }{
		{"# Store\n**Requirements:** R1, R3\n", "add", "R5", "**Requirements:** R1, R3, R5"},
		{"# Store\n**Requirements:**\n", "add", "R5", "**Requirements:** R5"},
		{"# Store\n**Requirements:** R1, R5\n", "add", "R5", "**Requirements:** R1, R5"},
		{"# Store\n**Requirements:** R1, R3, R5\n", "remove", "R3", "**Requirements:** R1, R5"},
	} {
		u, path := card(t, c.body)
		op := u.AddRef
		if c.op == "remove" {
			op = u.RemoveRef
		}
		if err := op("crc-Store.md", c.ref); err != nil {
			t.Fatalf("%s %s on %q: %v", c.op, c.ref, c.body, err)
		}
		if got := reqLine(t, path); got != c.want {
			t.Errorf("%s %s on %q: got %q, want %q", c.op, c.ref, c.body, got, c.want)
		}
	}
}

// R533 — the whole field comes back sorted, each once, a run of three or more as a range.
func TestAddRefWritesTheCanonicalForm(t *testing.T) {
	u, path := card(t, "# Store\n**Requirements:** R8, R1, R5, R7\n\nA store.\n")
	if err := u.AddRef("crc-Store.md", "R6"); err != nil {
		t.Fatal(err)
	}
	if got, want := read(t, path), "# Store\n**Requirements:** R1, R5-8\n\nA store.\n"; got != want {
		t.Errorf("got %q, want %q — the field must be rewritten sorted and minimal", got, want)
	}
}

// R534 — no Requirements line gets one beneath the heading; a ref inside a range is present.
func TestAddRefWritesAMissingLineAndKnowsARangesMembers(t *testing.T) {
	u, path := card(t, "# Store\n\nA store.\n")
	if err := u.AddRef("crc-Store.md", "R5"); err != nil {
		t.Fatal(err)
	}
	if got, want := read(t, path), "# Store\n**Requirements:** R5\n\nA store.\n"; got != want {
		t.Errorf("got %q, want %q — an add with nowhere to write must make the line (O36)", got, want)
	}
	before := "# Store\n**Requirements:** R5-8\n"
	u, path = card(t, before)
	if err := u.AddRef("crc-Store.md", "R6"); err != nil {
		t.Fatal(err)
	}
	if got := read(t, path); got != before {
		t.Errorf("adding R6 to R5-8 changed the card: %q", got)
	}
}

// R535 — a range splits; the last ref takes the line with it.
func TestRemoveRefSplitsARangeAndRemovesTheLastLine(t *testing.T) {
	u, path := card(t, "# Store\n**Requirements:** R5-8\n")
	if err := u.RemoveRef("crc-Store.md", "R6"); err != nil {
		t.Fatal(err)
	}
	if got, want := reqLine(t, path), "**Requirements:** R5, R7, R8"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	u, path = card(t, "# Store\n**Requirements:** R5\n\nA store.\n")
	if err := u.RemoveRef("crc-Store.md", "R5"); err != nil {
		t.Fatal(err)
	}
	if got, want := read(t, path), "# Store\n\nA store.\n"; got != want {
		t.Errorf("got %q, want %q — the last ref must remove the line, not leave it empty", got, want)
	}
}

// R536 — a token that is not a ref survives, after the refs, as written.
func TestTheRewriteKeepsTokensThatAreNotRefs(t *testing.T) {
	u, path := card(t, "# Store\n**Requirements:** R5, TBD, R9\n")
	if err := u.AddRef("crc-Store.md", "R6"); err != nil {
		t.Fatal(err)
	}
	if got, want := reqLine(t, path), "**Requirements:** R5, R6, R9, TBD"; got != want {
		t.Errorf("got %q, want %q — a token validate would report was lost", got, want)
	}
}

// R537 — removing a ref the card lacks is refused, and the card is untouched.
func TestRemoveRefOfAnAbsentRefIsAnError(t *testing.T) {
	before := "# Store\n**Requirements:** R1, R5\n"
	u, path := card(t, before)
	err := u.RemoveRef("crc-Store.md", "R3")
	if err == nil || !strings.Contains(err.Error(), "R3") || !strings.Contains(err.Error(), "crc-Store.md") {
		t.Errorf("want an error naming R3 and the card, got %v", err)
	}
	if got := read(t, path); got != before {
		t.Errorf("a refused remove changed the card: %q", got)
	}
}

// R533 — an argument that is not a requirement ref is refused, and the card is untouched.
func TestAddRefRefusesAnArgumentThatIsNotARef(t *testing.T) {
	before := "# Store\n**Requirements:** R1\n"
	for _, arg := range []string{"5", "R0", "Rx", "O5"} {
		u, path := card(t, before)
		if err := u.AddRef("crc-Store.md", arg); err == nil {
			t.Errorf("%q: want a refusal", arg)
		}
		if got := read(t, path); got != before {
			t.Errorf("%q changed the card: %q", arg, got)
		}
	}
}

