// CRC: crc-Parser.md | Seq: seq-parse.md | R77
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	featureRe = regexp.MustCompile(`^## Feature:\s*(.+)`)
	sourceRe  = regexp.MustCompile(`^\*\*Source:\*\*\s*(.+)`)
	// requirementRe matches `- **R1:** text` and `- **~~R1:~~** text` (retired form).
	requirementRe = regexp.MustCompile(`^- \*\*(~~)?R(\d+):(?:~~)?\*\*\s*(.+)`)
	inferredRe    = regexp.MustCompile(`^\(inferred\)\s*`)
	retiredPrefix = regexp.MustCompile(`^\(Retired\s+T\d+[^)]*\)\s*`)
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

		if featureRe.MatchString(line) {
			currentSource = ""
			continue
		}

		if matches := sourceRe.FindStringSubmatch(line); matches != nil {
			currentSource = strings.TrimSpace(matches[1])
			continue
		}

		matches := requirementRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		retired := matches[1] != ""
		text := strings.TrimSpace(matches[3])
		// R77: strip a leading "(Retired Tn — see Rxxx)" or "(Retired Tn — no replacement)"
		// marker if present, leaving the original text.
		if retired {
			text = retiredPrefix.ReplaceAllString(text, "")
		}

		inferred := false
		if inferredRe.MatchString(text) {
			inferred = true
			text = inferredRe.ReplaceAllString(text, "")
		}

		requirements = append(requirements, Requirement{
			ID:       "R" + matches[2],
			Text:     text,
			Source:   currentSource,
			Inferred: inferred,
			Retired:  retired,
			Line:     lineNum,
		})
	}

	return requirements, scanner.Err()
}
