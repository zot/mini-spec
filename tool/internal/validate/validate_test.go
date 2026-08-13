// CRC: crc-Validate.md | R92, R188
package validate

import (
	"strings"
	"testing"
)

// TestSourceFixInstructions pins which halves of the crank-handle block each
// Source diagnostic earns. The two halves answer different questions — how a
// Source line is shaped, and what to do when its spec is gone — so firing the
// wrong one is as bad as firing none.
func TestSourceFixInstructions(t *testing.T) {
	const (
		formatMark  = "must match this exact format"
		missingMark = "never by deleting the requirements"
	)
	cases := []struct {
		name            string
		result          ValidationResult
		format, missing bool
	}{
		{"nothing fired", ValidationResult{}, false, false},
		{"malformed only", ValidationResult{MalformedSpecSources: []string{"specs/a b.md"}}, true, false},
		{"suspicious only", ValidationResult{SuspiciousSourceLines: []string{"Source: specs/a.md"}}, true, false},
		{"missing only", ValidationResult{MissingSpecSources: []string{"specs/gone.md"}}, false, true},
		{
			"both",
			ValidationResult{
				MalformedSpecSources: []string{"specs/a b.md"},
				MissingSpecSources:   []string{"specs/gone.md"},
			},
			true, true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.result.sourceFixInstructions()
			if !tc.format && !tc.missing {
				if got != "" {
					t.Fatalf("no diagnostic fired, want empty block, got %q", got)
				}
				return
			}
			if n := strings.Count(got, "fix instructions:"); n != 1 {
				t.Errorf("want exactly one header, got %d", n)
			}
			if has := strings.Contains(got, formatMark); has != tc.format {
				t.Errorf("format half present = %v, want %v", has, tc.format)
			}
			if has := strings.Contains(got, missingMark); has != tc.missing {
				t.Errorf("missing half present = %v, want %v", has, tc.missing)
			}
		})
	}
}

// TestMissingSourceFixNamesEveryRepair guards the content rather than the
// trigger: the block is only worth emitting if it names all three repairs and
// forecloses the renumber, which is the failure it exists to prevent.
func TestMissingSourceFixNamesEveryRepair(t *testing.T) {
	r := ValidationResult{MissingSpecSources: []string{"specs/gone.md"}}
	got := r.sourceFixInstructions()
	for _, want := range []string{
		"renamed", "merged", "deleted",
		"minispec update retire",
		"specs/deleted.md",
		"never renumbered and never reused",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing-Source block does not mention %q:\n%s", want, got)
		}
	}
}
