// CRC: crc-Pending.md | Seq: seq-queue-item.md | R332, R333, R334
package pending

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/parser"
)

// CRC: crc-Pending.md | R332
// MissingTrajectoryError is a queue verb's refusal when the layer it writes is not there.
// Named by type so the CLI can crank out the create instruction rather than relay a raw
// open error — which is what a fresh project got until 2026-09-13.
type MissingTrajectoryError struct {
	RepoRoot string
	Files    []string
}

func (e *MissingTrajectoryError) Error() string {
	return fmt.Sprintf("the trajectory layer is not here: %s missing beneath %s", strings.Join(e.Files, ", "), e.RepoRoot)
}

// trajectoryFiles are the three files the layer is, in the order the format names them.
var trajectoryFiles = []string{pendingFile, currentFile, doneFile}

// missingTrajectory names the trajectory files absent beneath repoRoot, or nil.
func missingTrajectory(repoRoot string) error {
	var missing []string
	for _, f := range trajectoryFiles {
		if _, err := os.Stat(filepath.Join(repoRoot, f)); errors.Is(err, os.ErrNotExist) {
			missing = append(missing, f)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return &MissingTrajectoryError{RepoRoot: repoRoot, Files: missing}
}

// The lifecycle preambles, from trajectory-format.md. Each file is created whole with the
// prose the format mandates above its rule, so a file the tool made reads like one a
// person made from the reference, and there is nothing to paraphrase.
var preambles = map[string]string{
	pendingFile: `# Pending

The work queue. Ordered by intent — **the top item is active** — and each entry is a
*pointer*, not the design: subject, the skill that runs it (or nothing), and one doc link.
Rationale, status, findings and open questions live in the linked doc.

IDs are assigned once at creation and never reused. The next free ID is the maximum
assigned **anywhere** — across this file *and* ` + "`DONE.md`" + `. Completing an item removes its
entry (after marking it done in its source doc, clearing ` + "`CURRENT.md`" + `, and adding it to
` + "`DONE.md`" + `). Entries keep their numbers, so a gap in the sequence is expected.

---
`,
	currentFile: `# Current

Working context for the active item only — never a log (that's the done file). To
pause: lift this into a sub-item under that item's ` + "`##`" + ` heading in the pending file,
then reset here, freeing it for what you pick up next.

---

## Active

_No active item._
`,
	doneFile: `# Done

Completed items, most-recent first.

---
`,
}

// CRC: crc-Pending.md | R333
// CreateTrajectory writes each missing trajectory file with its lifecycle preamble and
// returns their names in the order written. A file that exists is never touched: creation
// is the one write that must refuse rather than overwrite, since the file it would replace
// is the whole record.
func CreateTrajectory(repoRoot string) ([]string, error) {
	var created []string
	for _, f := range trajectoryFiles {
		path := filepath.Join(repoRoot, f)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(preambles[f]), 0o644); err != nil {
			return created, err
		}
		created = append(created, f)
	}
	return created, nil
}

// CRC: crc-Pending.md | R334
// CreateCarve scaffolds carves/<name>.md: the title, a status block holding one open part,
// the decisions section, and the part's elaboration stub — the shape the format mandates,
// so the three notations ark's carves invented before there was anything to copy from
// cannot recur. Refuses an existing file; creates the carves directory when absent.
func CreateCarve(repoRoot, name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "/\\ ") || strings.HasSuffix(name, ".md") {
		return "", fmt.Errorf("a carve is named by one word, without a path or an extension, not %q", name)
	}
	dir := filepath.Join(repoRoot, parser.PublicCarvesDir)
	rel := parser.PublicCarvesDir + "/" + name + ".md"
	path := filepath.Join(dir, name+".md")
	if _, err := os.Stat(path); err == nil {
		return rel, fmt.Errorf("%s already exists; a scaffold never overwrites a carve", rel)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	body := "# Carve: " + name + `

<one paragraph: the problem this carve decomposes, and what is settled about it.>

## Status

- [ ] **Item 1 — <the first part>.** **OPEN (not queued.)**

## Decisions

<dated, attributed: **DECIDED (name, YYYY-MM-DD): …** — append-only, supersede in place.>

## Item 1

**Item 1** — <what the part is, what it needs, and what it does not do.>
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return rel, nil
}
