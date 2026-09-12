// CRC: crc-Validate.md | R327, R328
package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func designDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// R327, R328 — the two readings differ where one of them is blind: a fenced example is body
// to the reader and an entry to the naive line scan, in every one of the three documents,
// and the readers' unread lines — here an opener never closed, demoted to text by the
// dependency since 2026-09-12 — are counted beside the findings. Until that base landed the
// fixture was an unclosed span swallowing the tail, which the reader now recovers from.
func TestReadersDisagreeOverASwallowedTail(t *testing.T) {
	findings, unread := readerAgreement(designDir(t, map[string]string{
		"design.md":       "# D\n\n## Gaps\n\n- [ ] O1: fine\n- [ ] O2: quotes a `span that never closes\n\n```\n- [ ] O3: a fenced example\n```\n\n## Other\n\n- [ ] O9: not in the section\n",
		"requirements.md": "# R\n\n## Feature: A\n**Source:** specs/a.md\n\n- **R1:** one\n- **~~R2:~~** (Retired T1 — no replacement) two `\n\n```\n- **R3:** a fenced example\n```\n",
		"test-Fixture.md": "# T\n\n## Test: a\n**Purpose:** `never closed\n\n```\n## Test: b\n```\n",
	}))
	joined := strings.Join(findings, "\n")
	for _, want := range []string{
		"design.md: the line scan read gap O3 that the document reader returned no entry for",
		"requirements.md: the line scan read requirement R3 that the document reader returned no entry for",
		"test-Fixture.md: 2 test entries by line, 1 by the document reader",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing finding %q in:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "O9") {
		t.Errorf("a gap-shaped bullet outside the Gaps section was counted by the line scan:\n%s", joined)
	}
	for _, name := range []string{"design.md", "requirements.md", "test-Fixture.md"} {
		if unread[name] == 0 {
			t.Errorf("%s: the reader's unread line was not counted", name)
		}
	}
	r := &ValidationResult{ReaderDisagreement: findings, Unread: unread}
	out := r.FormatText()
	if !strings.HasPrefix(out, "issues:\n  the two readers disagree:\n") || !strings.Contains(out, "note: 3 line(s) were not read") {
		t.Errorf("the report does not lead with the disagreement and carry the note:\n%s", out)
	}
	if clean := (&ValidationResult{Unread: unread}).FormatText(); !strings.HasPrefix(clean, "note:") || !strings.HasSuffix(clean, "phase: validate OK\n") {
		t.Errorf("a clean result must still state its coverage:\n%s", clean)
	}
}

// R327 — agreement is silence: healthy documents produce no finding.
func TestReadersAgreeOverHealthyDocuments(t *testing.T) {
	findings, unread := readerAgreement(designDir(t, map[string]string{
		"design.md":       "# D\n\n## Gaps\n\n- [ ] O1: fine\n- A1: approved\n  - sub\n\n## Other\n",
		"requirements.md": "# R\n\n## Feature: A\n**Source:** specs/a.md\n\n- **R1:** one\n\n### Notes\n\n- **R2:** two\n",
		"test-Fixture.md": "# T\n\n## Test: a\n**Purpose:** p\n\n## Test: b\n**Fire alarm:** x\n",
	}))
	if len(findings) != 0 || len(unread) != 0 {
		t.Errorf("healthy documents: findings %v unread %v; want none", findings, unread)
	}
}
