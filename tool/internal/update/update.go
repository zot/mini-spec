// CRC: crc-Update.md | Seq: seq-update.md
package update

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/minispecsdom"
)

// Update provides atomic modification operations on design files
type Update struct {
	Project *project.Project
}

// New creates a new Update instance
func New(p *project.Project) *Update {
	return &Update{Project: p}
}

// Check checks a checkbox in the specified file
func (u *Update) Check(file, item string) error {
	return u.setCheckbox(file, item, true)
}

// Uncheck unchecks a checkbox in the specified file
func (u *Update) Uncheck(file, item string) error {
	return u.setCheckbox(file, item, false)
}

func (u *Update) setCheckbox(file, item string, checked bool) error {
	path := u.Project.DesignPath(file)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	quotedItem := regexp.QuoteMeta(item)

	patterns := []*regexp.Regexp{
		regexp.MustCompile(`^(\s*-\s*)\[([ x])\](\s*` + quotedItem + `\s*[→:].*)$`),
		regexp.MustCompile(`^(\s*-\s*)\[([ x])\](\s*` + quotedItem + `)$`),
	}

	newMark := " "
	if checked {
		newMark = "x"
	}

	for i, line := range lines {
		for _, pattern := range patterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				lines[i] = matches[1] + "[" + newMark + "]" + matches[3]
				return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			}
		}
	}

	return fmt.Errorf("item %q not found in %s", item, file)
}

// AddRef adds a requirement reference to a CRC card's Requirements field
func (u *Update) AddRef(crcFile, reqID string) error {
	path := u.Project.DesignPath(crcFile)
	card, err := parser.ParseCRCCard(path)
	if err != nil {
		return err
	}

	if slices.Contains(card.Requirements, reqID) {
		return nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	reqPattern := regexp.MustCompile(`^(\*\*Requirements:\*\*\s*)(.*)$`)

	for i, line := range lines {
		if matches := reqPattern.FindStringSubmatch(line); matches != nil {
			existing := strings.TrimSpace(matches[2])
			newReqs := reqID
			if existing != "" {
				newReqs = existing + ", " + reqID
			}
			lines[i] = matches[1] + newReqs
			break
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// RemoveRef removes a requirement reference from a CRC card
func (u *Update) RemoveRef(crcFile, reqID string) error {
	path := u.Project.DesignPath(crcFile)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	reqPattern := regexp.MustCompile(`^(\*\*Requirements:\*\*\s*)(.*)$`)

	for i, line := range lines {
		if matches := reqPattern.FindStringSubmatch(line); matches != nil {
			existing := strings.TrimSpace(matches[2])
			if existing == "" {
				break
			}
			parts := strings.Split(existing, ",")
			var newParts []string
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != reqID {
					newParts = append(newParts, p)
				}
			}
			lines[i] = matches[1] + strings.Join(newParts, ", ")
			break
		}
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// permanentTypes are gap types that are permanent (never resolved); their
// lines are written without a checkbox marker. R74, R75
var permanentTypes = map[string]bool{"A": true, "T": true}

// nextGapID returns the next available <type>n ID by scanning existing gaps.
//
// The counting moved to parser.NextGapNum so a read-only caller can ask the same
// question without importing this package; minting the ID string stays here, with the
// only code that writes one.
func nextGapID(gaps []parser.Gap, gapType string) string {
	return fmt.Sprintf("%s%d", gapType, parser.NextGapNum(gaps, gapType))
}

// editGaps reads design.md through the dependency's gaps reader, applies one write, and
// renders back through the atomic file write; a refusal reaches the caller with no byte
// written. R326
func (u *Update) editGaps(write func(g *minispecsdom.Gaps) error) error {
	return parser.EditFile(u.Project.DesignMdPath(), func(src string) (string, error) {
		g := minispecsdom.ParseGaps(src)
		if err := write(g); err != nil {
			return "", err
		}
		return g.Render()
	})
}

// CRC: crc-Update.md | Seq: seq-update.md | R82, R83, R326
// AddGap mints the next free ID of the type — the maximum ever assigned plus one, retired
// and resolved ones counted — and appends the entry through the reader, which writes the
// checkbox for a tracked type and none for a permanent one (R74, R75).
func (u *Update) AddGap(gapType, description string) (string, error) {
	gaps, err := parser.ParseGaps(u.Project.DesignMdPath())
	if err != nil {
		return "", err
	}
	newID := nextGapID(gaps, gapType)
	return newID, u.editGaps(func(g *minispecsdom.Gaps) error { return g.Add(newID, description) })
}

// CRC: crc-Update.md | Seq: seq-update.md | R326
// ResolveGap checks a tracked gap's box through the reader, which refuses a permanent gap
// (nothing to close) and one already resolved (so a second resolution is visible).
func (u *Update) ResolveGap(gapID string) error {
	return u.editGaps(func(g *minispecsdom.Gaps) error { return g.Resolve(gapID) })
}

// CRC: crc-Update.md | Seq: seq-update.md | R83, R326
// ApproveGap converts a tracked gap to an approved one, minting the next `A` number here
// and rewriting the head line through the reader; a gap already approved is left as it is
// and its own ID is returned.
func (u *Update) ApproveGap(gapID string) (string, error) {
	gaps, err := parser.ParseGaps(u.Project.DesignMdPath())
	if err != nil {
		return "", err
	}
	for _, g := range gaps {
		if g.ID == gapID && g.Type == "A" && !g.HasCheckbox {
			return g.ID, nil
		}
	}
	newID := nextGapID(gaps, "A")
	return newID, u.editGaps(func(g *minispecsdom.Gaps) error { return g.Approve(gapID, newID) })
}

// reqIDRe matches a bare Rn requirement identifier.
var reqIDRe = regexp.MustCompile(`^R\d+$`)

// CRC: crc-Update.md | Seq: seq-update.md | R80, R103, R326
// Retire rewrites the requirement's head line to its retired form through the requirements
// reader and appends the `Tn` gap through the gaps reader — two documents, one verb — and
// returns the assigned Tn with the requirement's `**Source:**` specs, so the CLI can print
// the supersede-at-source reminder. `-` or "" as the replacement means no replacement.
func (u *Update) Retire(oldReq, replacement, reason string) (string, []string, error) {
	if !reqIDRe.MatchString(oldReq) {
		return "", nil, fmt.Errorf("invalid requirement ID: %q", oldReq)
	}
	noReplacement := replacement == "" || replacement == "-"
	if !noReplacement && !reqIDRe.MatchString(replacement) {
		return "", nil, fmt.Errorf("invalid replacement requirement ID: %q (use Rn or -)", replacement)
	}
	reqs, err := parser.ParseRequirements(u.Project.RequirementsPath())
	if err != nil {
		return "", nil, err
	}
	var target *parser.Requirement
	for i := range reqs {
		if reqs[i].ID == oldReq {
			target = &reqs[i]
			break
		}
	}
	if target == nil {
		return "", nil, fmt.Errorf("requirement %s not found", oldReq)
	}
	if target.Retired {
		return "", nil, fmt.Errorf("requirement %s is already retired", oldReq)
	}
	gaps, err := parser.ParseGaps(u.Project.DesignMdPath())
	if err != nil {
		return "", nil, err
	}
	newTn := nextGapID(gaps, "T")
	var clause, gapDesc string
	if noReplacement {
		clause = "no replacement"
		gapDesc = fmt.Sprintf("%s retired (%s)", oldReq, reason)
	} else {
		clause = fmt.Sprintf("see %s", replacement)
		gapDesc = fmt.Sprintf("%s retired by %s (%s)", oldReq, replacement, reason)
	}
	err = parser.EditFile(u.Project.RequirementsPath(), func(src string) (string, error) {
		r := minispecsdom.ParseRequirements(src)
		if err := r.Retire(oldReq, newTn, clause); err != nil {
			return "", err
		}
		return r.Render()
	})
	if err != nil {
		return "", nil, err
	}
	if err := u.editGaps(func(g *minispecsdom.Gaps) error { return g.Add(newTn, gapDesc) }); err != nil {
		return "", nil, err
	}
	return newTn, target.Sources, nil
}

// ownMarkerRe matches a requirement body that opens with its own `**Rn:**` label.
var ownMarkerRe = regexp.MustCompile(`^\*\*(R\d+):\*\*`)

// CRC: crc-Update.md | Seq: seq-update.md | R324, R325, R326
// AddReq mints the next free Rn for each text — the maximum ever assigned, retired ones
// counted — and appends them to the named section in one invocation: no command hands out
// a number without recording it. The section is addressed by heading text at whatever level
// it lives, with or without the `Feature: ` prefix; an unknown heading is refused, since a
// section title is a judgment about how the design decomposes and the tool owns IDs, not
// prose; a title several headings carry is refused too. Entries land at the end of the
// section's own content, before its first sub-heading — the reader's rule.
//
// A body opening with its own `**Rn:**` is refused, not stripped: the verb mints both the
// identifier and the label, so a caller writing one is duplicating rather than choosing —
// measured 2026-08-22, when `**R414:** …` produced a well-formed line with a doubled marker
// that every check accepted.
func (u *Update) AddReq(section string, texts []string) ([]string, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("nothing to add: give at least one --req or --req-file")
	}
	for i, text := range texts {
		if m := ownMarkerRe.FindStringSubmatch(strings.TrimSpace(text)); m != nil {
			return nil, fmt.Errorf("requirement %d opens with %s, and this verb writes that label itself; pass the text alone — the number is assigned here, so a written one is a duplicate or a guess", i+1, m[1])
		}
	}
	reqs, err := parser.ParseRequirements(u.Project.RequirementsPath())
	if err != nil {
		return nil, err
	}
	maxNum := 0
	for _, req := range reqs {
		if n, err := strconv.Atoi(strings.TrimPrefix(req.ID, "R")); err == nil {
			maxNum = max(maxNum, n)
		}
	}
	ids := make([]string, len(texts))
	err = parser.EditFile(u.Project.RequirementsPath(), func(src string) (string, error) {
		r := minispecsdom.ParseRequirements(src)
		title := section
		found := r.Section(title)
		if len(found) == 0 {
			title = "Feature: " + section
			found = r.Section(title)
		}
		switch {
		case len(found) == 0:
			return "", fmt.Errorf("no section of requirements.md is named %q; write the heading by hand first — a section title is a judgment about how the design decomposes, and it carries no ID, so writing it races nothing", section)
		case len(found) > 1:
			return "", fmt.Errorf("%d sections of requirements.md are named %q; the reader does not pick", len(found), title)
		}
		for i, text := range texts {
			ids[i] = fmt.Sprintf("R%d", maxNum+1+i)
			if err := r.Add(title, ids[i], strings.TrimSpace(text)); err != nil {
				return "", err
			}
		}
		return r.Render()
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// MigrationComplete moves specs/migrations/<name>.md to
// specs/migrations/complete/<NNN>-<name>.md with the next zero-padded
// three-digit prefix. Returns the new path (relative to project root). R81
func (u *Update) MigrationComplete(name string) (string, error) {
	name = strings.TrimSuffix(name, ".md")
	srcRel := filepath.Join("specs", "migrations", name+".md")
	src := filepath.Join(u.Project.RootPath, srcRel)
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("migration spec not found: %s", srcRel)
	}

	completeDir := u.Project.MigrationsCompleteDir()
	if err := os.MkdirAll(completeDir, 0755); err != nil {
		return "", err
	}

	maxN := 0
	prefixRe := regexp.MustCompile(`^(\d{3})-`)
	entries, err := os.ReadDir(completeDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if m := prefixRe.FindStringSubmatch(e.Name()); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > maxN {
				maxN = n
			}
		}
	}

	dstName := fmt.Sprintf("%03d-%s.md", maxN+1, name)
	dst := filepath.Join(completeDir, dstName)
	if err := os.Rename(src, dst); err != nil {
		return "", err
	}

	rel, err := filepath.Rel(u.Project.RootPath, dst)
	if err != nil {
		rel = filepath.Join("specs", "migrations", "complete", dstName)
	}
	return rel, nil
}

// SortRequirements sorts a comma-separated list of requirements numerically
func SortRequirements(reqs []string) []string {
	sorted := make([]string, len(reqs))
	copy(sorted, reqs)
	sort.Slice(sorted, func(i, j int) bool {
		return extractNum(sorted[i]) < extractNum(sorted[j])
	})
	return sorted
}

func extractNum(reqID string) int {
	numStr := strings.TrimPrefix(reqID, "R")
	num, _ := strconv.Atoi(numStr)
	return num
}
