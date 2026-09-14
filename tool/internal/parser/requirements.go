// CRC: crc-Parser.md | Seq: seq-parse.md | R77, R90, R91
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
)

var (
	sourceRe   = regexp.MustCompile(`^\*\*Source:\*\*\s*(.+)`)
	inferredRe = regexp.MustCompile(`^\(inferred\)\s*`)
	// suspiciousSourceRe matches lines that look like attempted Source markers
	// (markdown emphasis + "Source" + a colon somewhere) but don't necessarily
	// match the canonical sourceRe pattern. R91
	suspiciousSourceRe = regexp.MustCompile(`^[*_~]*[Ss]ource[*_]*\s*:`)
)

// splitSourceList splits a `**Source:**` value into trimmed, non-empty paths.
// R90: a Source line may carry a comma-separated list of paths.
func splitSourceList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// CRC: crc-Parser.md | R326
// ParseRequirements reads requirements.md through the dependency's requirements reader: a
// section is a heading at any level with its own content, a requirement is a column-0
// `**Rn:**` or `**~~Rn:~~**` bullet with its text folded and the retired clause read out,
// and a fenced example is body. A section with no `**Source:**` of its own inherits its
// nearest ancestor's, which is what a `### Notes` under a `## Feature:` always meant.
func ParseRequirements(path string) ([]Requirement, error) {
	reqs, _, err := ParseRequirementsReport(path)
	return reqs, err
}

// CRC: crc-Parser.md | R326
// ParseRequirementsReport is ParseRequirements with the reader's unread list beside it.
func ParseRequirementsReport(path string) ([]Requirement, []minispecsdom.Unread, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	r := minispecsdom.ParseRequirements(string(data))
	var out []Requirement
	for _, q := range r.Requirements() {
		text := strings.TrimSpace(q.Text)
		inferred := inferredRe.MatchString(text)
		out = append(out, Requirement{
			ID:       q.ID,
			Text:     inferredRe.ReplaceAllString(text, ""),
			Sources:  splitSourceList(sectionSource(q.Section)),
			Inferred: inferred,
			Retired:  q.Retired,
			Line:     q.Line(),
		})
	}
	return out, r.Unread(), nil
}

// sectionSource is the nearest `**Source:**` at or above a section.
func sectionSource(s *minispecsdom.Section) string {
	for ; s != nil; s = s.Parent {
		if s.Source != "" {
			return s.Source
		}
	}
	return ""
}

// ScanSourceLineIssues re-scans a requirements.md file for lines that look
// like Source markers but don't match the canonical `**Source:** ...` pattern.
// Lines that match the canonical pattern are ignored. R91
func ScanSourceLineIssues(path string) ([]SourceLineIssue, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var issues []SourceLineIssue
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if sourceRe.MatchString(line) {
			continue
		}
		if suspiciousSourceRe.MatchString(line) {
			issues = append(issues, SourceLineIssue{LineNum: lineNum, Line: line})
		}
	}
	return issues, scanner.Err()
}
