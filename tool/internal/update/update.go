// CRC: crc-Update.md | Seq: seq-update.md
package update

import (
	"fmt"
	"os"
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

	// Patterns: item with separator (→ or :) and trailing content, or bare item at end of line
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

	// Check for duplicate
	if slices.Contains(card.Requirements, reqID) {
		return nil // Already present
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
			var newReqs string
			if existing == "" {
				newReqs = reqID
			} else {
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

// AddGap adds a new gap item with auto-numbered ID
func (u *Update) AddGap(gapType, description string) (string, error) {
	path := u.Project.DesignMdPath()
	gaps, err := parser.ParseGaps(path)
	if err != nil {
		return "", err
	}

	// Find next ID for this type
	maxNum := 0
	for _, g := range gaps {
		if g.Type == gapType {
			numStr := strings.TrimPrefix(g.ID, gapType)
			if num, err := strconv.Atoi(numStr); err == nil && num > maxNum {
				maxNum = num
			}
		}
	}
	newID := fmt.Sprintf("%s%d", gapType, maxNum+1)

	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("Gaps section not found in design.md")
	}

	newLine := fmt.Sprintf("- [ ] %s: %s", newID, description)

	// Insert new gap line
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, newLine)
	newLines = append(newLines, lines[insertIdx:]...)

	return newID, os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

// ResolveGap marks a gap as resolved (checks its checkbox)
func (u *Update) ResolveGap(gapID string) error {
	return u.Check("design.md", gapID)
}

// SortRequirements sorts a comma-separated list of requirements numerically
func SortRequirements(reqs []string) []string {
	sorted := make([]string, len(reqs))
	copy(sorted, reqs)
	sort.Slice(sorted, func(i, j int) bool {
		numI := extractNum(sorted[i])
		numJ := extractNum(sorted[j])
		return numI < numJ
	})
	return sorted
}

func extractNum(reqID string) int {
	numStr := strings.TrimPrefix(reqID, "R")
	num, _ := strconv.Atoi(numStr)
	return num
}
