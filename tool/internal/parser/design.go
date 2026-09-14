// CRC: crc-Parser.md | Seq: seq-parse.md | R71, R73, R74, R75
package parser

import (
	"bufio"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
)

var (
	sectionRe        = regexp.MustCompile(`^## (.+)`)
	subsectionRe     = regexp.MustCompile(`^### .+`)
	designFileRe     = regexp.MustCompile(`^- (.+\.md)`)
	codeFileRe       = regexp.MustCompile(`^  - \[([ x])\] (.+)`)
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

		if matches := sectionRe.FindStringSubmatch(line); matches != nil {
			section := strings.TrimSpace(matches[1])
			if section == "Artifacts" {
				inArtifacts = true
				continue
			} else if inArtifacts {
				break
			}
			continue
		}

		if !inArtifacts {
			continue
		}

		if subsectionRe.MatchString(line) {
			continue
		}

		if matches := inlineArtifactRe.FindStringSubmatch(line); matches != nil {
			checked := matches[1] == "x"
			designFile := matches[2]
			codeFilesStr := matches[3]

			if current != nil {
				artifacts = append(artifacts, *current)
				current = nil
			}

			artifact := Artifact{DesignFile: designFile}

			if codeFilesStr != "" {
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

		if matches := designFileRe.FindStringSubmatch(line); matches != nil {
			if current != nil {
				artifacts = append(artifacts, *current)
			}
			current = &Artifact{DesignFile: matches[1]}
			continue
		}

		if matches := codeFileRe.FindStringSubmatch(line); matches != nil && current != nil {
			current.CodeFiles = append(current.CodeFiles, CodeFile{
				Path:    strings.TrimSpace(matches[2]),
				Checked: matches[1] == "x",
				Line:    lineNum,
			})
		}
	}

	if current != nil {
		artifacts = append(artifacts, *current)
	}

	return artifacts, scanner.Err()
}

// CRC: crc-Parser.md | R326
// ParseGaps reads the Gaps section of design.md through the dependency's gaps reader: a gap
// is a bullet at any depth whose head is `X<n>:`, nested ones included, its text folded
// across continuation lines; a fenced example is body; a permanent gap with a checkbox or a
// tracked one without is a deviation the reader lists. What the reader could not read is
// dropped here; ParseGapsReport carries it.
func ParseGaps(path string) ([]Gap, error) {
	gaps, _, err := ParseGapsReport(path)
	return gaps, err
}

// CRC: crc-Parser.md | R326
// ParseGapsReport is ParseGaps with the reader's unread list beside the gaps.
func ParseGapsReport(path string) ([]Gap, []minispecsdom.Unread, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	g := minispecsdom.ParseGaps(string(data))
	var out []Gap
	for _, item := range g.Items() {
		out = append(out, Gap{
			ID:          item.ID,
			Type:        item.Type,
			Description: item.Text,
			Resolved:    item.Checked,
			HasCheckbox: item.Checkbox,
			Line:        item.Line(),
		})
	}
	return out, g.Unread(), nil
}

// GapTypes are the gap classes in the order they are reported. R192
var GapTypes = []string{"S", "R", "D", "C", "I", "O", "A", "T"}

// IsGapType reports whether s names one of the gap classes.
//
// The membership test lives with the list so a caller validating a type and a caller
// reporting the set cannot disagree about which classes exist. The gap regexes above
// still spell the same set a third time; they are the remaining copy.
func IsGapType(s string) bool {
	return slices.Contains(GapTypes, s)
}

// CRC: crc-Parser.md | R192
// NextGapNum returns the next free number for one gap type.
//
// Factored out of update's nextGapID so a read-only caller can ask the same question
// without importing a package that writes. The formatting stays with the writer, which
// is the only place an ID string is minted.
func NextGapNum(gaps []Gap, gapType string) int {
	maxNum := 0
	for _, g := range gaps {
		if g.Type != gapType {
			continue
		}
		if num, err := strconv.Atoi(strings.TrimPrefix(g.ID, gapType)); err == nil {
			maxNum = max(maxNum, num)
		}
	}
	return maxNum + 1
}
