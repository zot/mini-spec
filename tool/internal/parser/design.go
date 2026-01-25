// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	sectionRe       = regexp.MustCompile(`^## (.+)`)
	subsectionRe    = regexp.MustCompile(`^### .+`)
	designFileRe    = regexp.MustCompile(`^- (.+\.md)`)
	codeFileRe      = regexp.MustCompile(`^  - \[([ x])\] (.+)`)
	gapRe           = regexp.MustCompile(`^- \[([ x])\] ([SRDCO])(\d+):\s*(.+)`)
	inlineArtifactRe = regexp.MustCompile(`^- \[([ x])\] ([^\s→]+\.md)(?:\s*→\s*(.+))?$`)
)

// ParseArtifacts parses the Artifacts section of design.md
// Supports both legacy nested format and new inline format:
// Legacy: - design.md\n  - [x] code.ts
// Inline: - [x] design.md → code.ts, code2.ts
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

		// Check for section headers (## ...)
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

		// Skip subsection headers (### CRC Cards, etc.)
		if subsectionRe.MatchString(line) {
			continue
		}

		// Try new inline format first: - [x] design.md → code.ts, code2.ts
		if matches := inlineArtifactRe.FindStringSubmatch(line); matches != nil {
			checked := matches[1] == "x"
			designFile := matches[2]
			codeFilesStr := matches[3]

			// Save any pending artifact from legacy format
			if current != nil {
				artifacts = append(artifacts, *current)
				current = nil
			}

			artifact := Artifact{DesignFile: designFile}

			if codeFilesStr != "" {
				// Split on comma, strip backticks and whitespace
				for _, cf := range strings.Split(codeFilesStr, ",") {
					cf = strings.TrimSpace(cf)
					cf = strings.Trim(cf, "`")
					if cf != "" {
						artifact.CodeFiles = append(artifact.CodeFiles, CodeFile{
							Path:    cf,
							Checked: checked,
							Line:    lineNum,
						})
					}
				}
			}

			artifacts = append(artifacts, artifact)
			continue
		}

		// Legacy format: design file line without checkbox
		if matches := designFileRe.FindStringSubmatch(line); matches != nil {
			if current != nil {
				artifacts = append(artifacts, *current)
			}
			current = &Artifact{DesignFile: matches[1]}
			continue
		}

		// Legacy format: code file checkbox (indented)
		if matches := codeFileRe.FindStringSubmatch(line); matches != nil && current != nil {
			current.CodeFiles = append(current.CodeFiles, CodeFile{
				Path:    strings.TrimSpace(matches[2]),
				Checked: matches[1] == "x",
				Line:    lineNum,
			})
		}
	}

	// Don't forget the last artifact (legacy format)
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
