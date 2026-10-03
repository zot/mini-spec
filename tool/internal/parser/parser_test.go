// CRC: crc-Parser.md | Test: test-Parser.md | R532
package parser

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// crcWith writes a CRC card whose Requirements field is field and parses it.
func crcWith(t *testing.T, field string) CRCCard {
	t.Helper()
	path := filepath.Join(t.TempDir(), "crc-X.md")
	if err := os.WriteFile(path, []byte("# X\n**Requirements:** "+field+"\n\nA card.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	card, err := ParseCRCCard(path)
	if err != nil {
		t.Fatal(err)
	}
	return card
}

// R532 — a range in either spelling names every member; a plain list is unchanged.
func TestCRCRequirementsReadRangesAsMembers(t *testing.T) {
	for field, want := range map[string][]string{
		"R5-8, R10":  {"R5", "R6", "R7", "R8", "R10"},
		"R5-R7":      {"R5", "R6", "R7"},
		"R1, R3, R7": {"R1", "R3", "R7"},
	} {
		if got := crcWith(t, field).Requirements; !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q — a range must name every member", field, got, want)
		}
	}
}

// R532 — text the grammar stops on is kept as tokens, so validate can report it.
func TestCRCRequirementsKeepLeftoverTokens(t *testing.T) {
	want := []string{"R5", "R9", "TBD"}
	if got := crcWith(t, "R5, TBD, R9").Requirements; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q — text after the list was dropped, not reported", got, want)
	}
	// A range written after a stray word still names its members.
	want = []string{"R5", "R10", "R11", "R12", "TBD"}
	if got := crcWith(t, "R5, TBD, R10-12").Requirements; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q — a range after a stray word was read as one token", got, want)
	}
}

// R532 — a field the grammar cannot start on keeps every token, so none goes unreported.
func TestCRCRequirementsKeepAFieldThatOpensWithJunk(t *testing.T) {
	want := []string{"R5", "TBD"}
	if got := crcWith(t, "TBD, R5").Requirements; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q — a field opening with a non-ref was dropped", got, want)
	}
}
