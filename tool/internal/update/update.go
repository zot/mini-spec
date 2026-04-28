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
func nextGapID(gaps []parser.Gap, gapType string) string {
	maxNum := 0
	for _, g := range gaps {
		if g.Type != gapType {
			continue
		}
		if num, err := strconv.Atoi(strings.TrimPrefix(g.ID, gapType)); err == nil && num > maxNum {
			maxNum = num
		}
	}
	return fmt.Sprintf("%s%d", gapType, maxNum+1)
}

// AddGap adds a new gap item with auto-numbered ID. R82, R83
func (u *Update) AddGap(gapType, description string) (string, error) {
	path := u.Project.DesignMdPath()
	gaps, err := parser.ParseGaps(path)
	if err != nil {
		return "", err
	}

	newID := nextGapID(gaps, gapType)
	return newID, u.appendGapLine(path, formatGapLine(gapType, newID, description))
}

// formatGapLine returns the canonical Gaps-section line for the given gap. R74, R75
func formatGapLine(gapType, id, description string) string {
	if permanentTypes[gapType] {
		return fmt.Sprintf("- %s: %s", id, description)
	}
	return fmt.Sprintf("- [ ] %s: %s", id, description)
}

func (u *Update) appendGapLine(path, newLine string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	gapsSectionRe := regexp.MustCompile(`^## Gaps`)
	nextSectionRe := regexp.MustCompile(`^## `)

	inGaps := false
	insertIdx := -1

	for i, line := range lines {
		if gapsSectionRe.MatchString(line) {
			inGaps = true
			continue
		}
		if inGaps && nextSectionRe.MatchString(line) {
			insertIdx = i
			break
		}
		if inGaps {
			insertIdx = i + 1
		}
	}

	if insertIdx == -1 {
		return fmt.Errorf("Gaps section not found in design.md")
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, newLine)
	newLines = append(newLines, lines[insertIdx:]...)

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

// ResolveGap marks a gap as resolved. Refuses A and T (permanent) types.
func (u *Update) ResolveGap(gapID string) error {
	if len(gapID) > 0 && permanentTypes[string(gapID[0])] {
		return fmt.Errorf("%s is a permanent gap type and cannot be resolved", gapID)
	}
	return u.Check("design.md", gapID)
}

// ApproveGap converts an existing gap to approved (A) type, written without a
// checkbox. R83
func (u *Update) ApproveGap(gapID string) (string, error) {
	path := u.Project.DesignMdPath()
	gaps, err := parser.ParseGaps(path)
	if err != nil {
		return "", err
	}

	var target *parser.Gap
	for i := range gaps {
		if gaps[i].ID == gapID {
			target = &gaps[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("gap %s not found", gapID)
	}

	if target.Type == "A" && !target.HasCheckbox {
		return target.ID, nil
	}

	newID := target.ID
	if target.Type != "A" {
		newID = nextGapID(gaps, "A")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	if target.Line > 0 && target.Line <= len(lines) {
		lines[target.Line-1] = formatGapLine("A", newID, target.Description)
	}

	return newID, os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

// reqIDRe matches a bare Rn requirement identifier.
var reqIDRe = regexp.MustCompile(`^R\d+$`)

// Retire rewrites the oldReq line in requirements.md with the strikethrough/
// Retired marker AND appends a new T-typed gap to design.md. Returns the
// assigned Tn. If replacement is "-" or empty, the marker says "no replacement".
// R80
func (u *Update) Retire(oldReq, replacement, reason string) (string, error) {
	if !reqIDRe.MatchString(oldReq) {
		return "", fmt.Errorf("invalid requirement ID: %q", oldReq)
	}
	noReplacement := replacement == "" || replacement == "-"
	if !noReplacement && !reqIDRe.MatchString(replacement) {
		return "", fmt.Errorf("invalid replacement requirement ID: %q (use Rn or -)", replacement)
	}

	reqsPath := u.Project.RequirementsPath()
	reqs, err := parser.ParseRequirements(reqsPath)
	if err != nil {
		return "", err
	}

	var target *parser.Requirement
	for i := range reqs {
		if reqs[i].ID == oldReq {
			target = &reqs[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("requirement %s not found", oldReq)
	}
	if target.Retired {
		return "", fmt.Errorf("requirement %s is already retired", oldReq)
	}

	gapsPath := u.Project.DesignMdPath()
	gaps, err := parser.ParseGaps(gapsPath)
	if err != nil {
		return "", err
	}
	newTn := nextGapID(gaps, "T")

	replacementClause := fmt.Sprintf("see %s", replacement)
	gapDesc := fmt.Sprintf("%s retired by %s (%s)", oldReq, replacement, reason)
	if noReplacement {
		replacementClause = "no replacement"
		gapDesc = fmt.Sprintf("%s retired (%s)", oldReq, reason)
	}

	if err := u.rewriteRetiredLine(reqsPath, target, newTn, replacementClause); err != nil {
		return "", err
	}
	if err := u.appendGapLine(gapsPath, formatGapLine("T", newTn, gapDesc)); err != nil {
		return "", err
	}
	return newTn, nil
}

// rewriteRetiredLine rewrites a requirement line in-place to its retired form.
func (u *Update) rewriteRetiredLine(path string, req *parser.Requirement, tn, replacementClause string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	if req.Line <= 0 || req.Line > len(lines) {
		return fmt.Errorf("requirement %s line %d out of range", req.ID, req.Line)
	}

	original := lines[req.Line-1]
	prefixRe := regexp.MustCompile(`^(- \*\*)R\d+:(\*\*\s*)(.*)$`)
	matches := prefixRe.FindStringSubmatch(original)
	if matches == nil {
		return fmt.Errorf("could not rewrite line %d (unexpected format): %q", req.Line, original)
	}
	prefix, sep, text := matches[1], matches[2], matches[3]
	lines[req.Line-1] = fmt.Sprintf("%s~~%s:~~%s(Retired %s — %s) %s",
		prefix, req.ID, sep, tn, replacementClause, text)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
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
