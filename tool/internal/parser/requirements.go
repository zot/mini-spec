// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	featureRe     = regexp.MustCompile(`^## Feature:\s*(.+)`)
	sourceRe      = regexp.MustCompile(`^\*\*Source:\*\*\s*(.+)`)
	requirementRe = regexp.MustCompile(`^- \*\*R(\d+):\*\*\s*(.+)`)
	inferredRe    = regexp.MustCompile(`^\(inferred\)\s*`)
)

// ParseRequirements parses a requirements.md file
func ParseRequirements(path string) ([]Requirement, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var requirements []Requirement
	var currentSource string
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for Feature header (resets source)
		if featureRe.MatchString(line) {
			currentSource = ""
			continue
		}

		// Check for Source line
		if matches := sourceRe.FindStringSubmatch(line); matches != nil {
			currentSource = strings.TrimSpace(matches[1])
			continue
		}

		// Check for requirement line
		if matches := requirementRe.FindStringSubmatch(line); matches != nil {
			text := strings.TrimSpace(matches[2])
			inferred := false

			if inferredRe.MatchString(text) {
				inferred = true
				text = inferredRe.ReplaceAllString(text, "")
			}

			requirements = append(requirements, Requirement{
				ID:       "R" + matches[1],
				Text:     text,
				Source:   currentSource,
				Inferred: inferred,
				Line:     lineNum,
			})
		}
	}

	return requirements, scanner.Err()
}
