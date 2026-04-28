// CRC: crc-Validate.md | R85
package validate

import "testing"

func TestFormatRanges(t *testing.T) {
	cases := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{"R1"}, "R1"},
		{[]string{"R1", "R2", "R3"}, "R1-3"},
		{[]string{"R44", "R45", "R46", "R72"}, "R44-46, R72"},
		{[]string{"R177", "R178", "R179", "R180", "R181"}, "R177-181"},
		{[]string{"R3", "R7", "R8"}, "R3, R7-8"},
	}
	for _, tc := range cases {
		got := FormatRanges(tc.in)
		if got != tc.want {
			t.Errorf("FormatRanges(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
