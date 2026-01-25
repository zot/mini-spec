// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	sectionRe    = regexp.MustCompile(`^## (.+)`)
	designFileRe = regexp.MustCompile(`^- (.+\.md)`)
	codeFileRe   = regexp.MustCompile(`^  - \[([ x])\] (.+)`)
	gapRe        = regexp.MustCompile(`^- \[([ x])\] ([SRDCO])(\d+):\s*(.+)`)
)

// ParseArtifacts parses the Artifacts section of design.md
func ParseArtifacts(path string) ([]Artifact, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var artifacts []Artifact
	var current *Artifact
	inArtifacts := false
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for section headers
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			section := strings.TrimSpace(matches[1])
			if section == "Artifacts" {
				inArtifacts = true
				continue
			} else if inArtifacts {
				// End of Artifacts section
				break
			}
			continue
		}

		if !inArtifacts {
			continue
		}

		// Check for design file line
		if matches := designFileRe.FindStringSubmatch(line); matches != nil {
			if current != nil {
				artifacts = append(artifacts, *current)
			}
			current = &Artifact{DesignFile: matches[1]}
			continue
		}

		// Check for code file checkbox
		if matches := codeFileRe.FindStringSubmatch(line); matches != nil && current != nil {
			current.CodeFiles = append(current.CodeFiles, CodeFile{
				Path:    strings.TrimSpace(matches[2]),
				Checked: matches[1] == "x",
				Line:    lineNum,
			})
		}
	}

	// Don't forget the last artifact
	if current != nil {
		artifacts = append(artifacts, *current)
	}

	return artifacts, scanner.Err()
}

// ParseGaps parses the Gaps section of design.md
func ParseGaps(path string) ([]Gap, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var gaps []Gap
	inGaps := false
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for section headers
		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			section := strings.TrimSpace(matches[1])
			if section == "Gaps" {
				inGaps = true
				continue
			} else if inGaps {
				// End of Gaps section
				break
			}
			continue
		}

		if !inGaps {
			continue
		}

		// Check for gap line
		if matches := gapRe.FindStringSubmatch(line); matches != nil {
			gaps = append(gaps, Gap{
				ID:          matches[2] + matches[3],
				Type:        matches[2],
				Description: strings.TrimSpace(matches[4]),
				Resolved:    matches[1] == "x",
				Line:        lineNum,
			})
		}
	}

	return gaps, scanner.Err()
}
