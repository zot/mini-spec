// CRC: crc-Phase.md | Seq: seq-phase.md
package phase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
)

// Result contains the output of a phase validation
type Result struct {
	Phase    string   `json:"phase"`
	Passed   bool     `json:"passed"`
	Findings any      `json:"findings,omitempty"`
	Issues   []string `json:"issues,omitempty"`
}

// Phase runs phase-specific validations
type Phase struct {
	Project *project.Project
	Query   *query.Query
}

// New creates a new Phase instance
func New(p *project.Project) *Phase {
	return &Phase{
		Project: p,
		Query:   query.New(p),
	}
}

// RunSpec validates spec files exist and are non-empty (R44)
func (ph *Phase) RunSpec() *Result {
	result := &Result{Phase: "spec"}

	specsDir := filepath.Join(ph.Project.RootPath, "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("specs/ directory: %v", err))
		return result
	}

	var specFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			specFiles = append(specFiles, e.Name())

			// Check non-empty
			path := filepath.Join(specsDir, e.Name())
			info, err := os.Stat(path)
			if err == nil && info.Size() == 0 {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: empty file", e.Name()))
			}
		}
	}

	if len(specFiles) == 0 {
		result.Issues = append(result.Issues, "no spec files found in specs/")
	}

	result.Findings = map[string]any{
		"specs_found": specFiles,
	}
	result.Passed = len(result.Issues) == 0
	return result
}

// RunRequirements validates requirements.md format and spec sources (R45)
func (ph *Phase) RunRequirements() *Result {
	result := &Result{Phase: "requirements"}

	reqs, err := ph.Query.Requirements()
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("requirements.md: %v", err))
		return result
	}

	if len(reqs) == 0 {
		result.Issues = append(result.Issues, "no requirements found")
		return result
	}

	// Check sequential numbering
	var found []string
	var inferred []string
	sources := make(map[string][]string)
	expectedNum := 1

	for _, r := range reqs {
		found = append(found, r.ID)
		if r.Inferred {
			inferred = append(inferred, r.ID)
		}
		if r.Source != "" {
			sources[r.Source] = append(sources[r.Source], r.ID)
		}

		// Check sequential
		var num int
		fmt.Sscanf(r.ID, "R%d", &num)
		if num != expectedNum {
			result.Issues = append(result.Issues, fmt.Sprintf("non-sequential: expected R%d, found %s", expectedNum, r.ID))
		}
		expectedNum = num + 1
	}

	// Check spec sources exist (R41)
	checked := make(map[string]bool)
	for _, r := range reqs {
		if r.Source == "" || checked[r.Source] {
			continue
		}
		checked[r.Source] = true

		specPath := filepath.Join(ph.Project.RootPath, r.Source)
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: referenced as Source but file missing", r.Source))
		}
	}

	result.Findings = map[string]any{
		"found":    found,
		"sources":  sources,
		"inferred": inferred,
	}
	result.Passed = len(result.Issues) == 0
	return result
}

// RunDesign validates design files, CRC cards, and requirement coverage (R46)
func (ph *Phase) RunDesign() *Result {
	result := &Result{Phase: "design"}

	// Get requirements for validation
	reqs, err := ph.Query.Requirements()
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("requirements.md: %v", err))
		return result
	}

	validReqs := make(map[string]bool)
	for _, r := range reqs {
		validReqs[r.ID] = true
	}

	// Parse artifacts
	artifacts, err := ph.Query.Artifacts()
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("design.md Artifacts: %v", err))
		return result
	}

	// Get CRC cards and check requirements
	crcCards := make(map[string][]string)
	files, _ := ph.Project.GlobCRCCards()
	for _, path := range files {
		card, err := parser.ParseCRCCard(path)
		if err != nil {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: parse error", filepath.Base(path)))
			continue
		}

		relPath := filepath.Base(path)
		crcCards[relPath] = card.Requirements

		if len(card.Requirements) == 0 {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: no Requirements field", relPath))
		}

		for _, reqID := range card.Requirements {
			if !validReqs[reqID] {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: references unknown %s", relPath, reqID))
			}
		}
	}

	// Compute coverage
	covered := make(map[string]bool)
	for _, reqs := range crcCards {
		for _, reqID := range reqs {
			covered[reqID] = true
		}
	}

	var coveredList, uncoveredList []string
	for _, r := range reqs {
		if covered[r.ID] {
			coveredList = append(coveredList, r.ID)
		} else {
			uncoveredList = append(uncoveredList, r.ID)
		}
	}

	if len(uncoveredList) > 0 {
		result.Issues = append(result.Issues, fmt.Sprintf("uncovered requirements: %s", strings.Join(uncoveredList, ", ")))
	}

	// Check all design files are listed in Artifacts (R40)
	patterns := []string{"crc-*.md", "seq-*.md", "ui-*.md", "test-*.md", "manifest-*.md"}
	designFiles := make(map[string]bool)
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(ph.Project.DesignPath(pattern))
		for _, m := range matches {
			designFiles[filepath.Base(m)] = true
		}
	}

	listedFiles := make(map[string]bool)
	for _, art := range artifacts {
		listedFiles[art.DesignFile] = true
	}

	for file := range designFiles {
		if !listedFiles[file] {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: not listed in Artifacts", file))
		}
	}

	result.Findings = map[string]any{
		"crc_cards":    crcCards,
		"covered":      coveredList,
		"uncovered":    uncoveredList,
		"design_files": len(designFiles),
	}
	result.Passed = len(result.Issues) == 0
	return result
}

// RunImplementation validates code files and traceability comments (R47)
func (ph *Phase) RunImplementation() *Result {
	result := &Result{Phase: "implementation"}

	artifacts, err := ph.Query.Artifacts()
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("design.md Artifacts: %v", err))
		return result
	}

	var codeFiles []map[string]any

	for _, art := range artifacts {
		for _, cf := range art.CodeFiles {
			fileInfo := map[string]any{
				"path":    cf.Path,
				"checked": cf.Checked,
			}

			fullPath := filepath.Join(ph.Project.RootPath, cf.Path)
			_, err := os.Stat(fullPath)
			exists := err == nil
			fileInfo["exists"] = exists

			if !exists && cf.Checked {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: checked but file missing", cf.Path))
			}

			if exists {
				ext := filepath.Ext(cf.Path)
				pattern := ph.Project.CommentPattern(ext)
				trace, err := parser.ParseTraceability(fullPath, pattern)
				if err == nil {
					fileInfo["crc_refs"] = trace.CRCRefs
					fileInfo["seq_refs"] = trace.SeqRefs

					if len(trace.CRCRefs) == 0 {
						result.Issues = append(result.Issues, fmt.Sprintf("%s: missing traceability comment", cf.Path))
					}

					// Check refs exist (R42)
					for _, ref := range trace.CRCRefs {
						refPath := ph.Project.DesignPath(ref)
						if _, err := os.Stat(refPath); os.IsNotExist(err) {
							result.Issues = append(result.Issues, fmt.Sprintf("%s: references %s which does not exist", cf.Path, ref))
						}
					}
					for _, ref := range trace.SeqRefs {
						refPath := ph.Project.DesignPath(ref)
						if _, err := os.Stat(refPath); os.IsNotExist(err) {
							result.Issues = append(result.Issues, fmt.Sprintf("%s: references %s which does not exist", cf.Path, ref))
						}
					}
				}
			}

			codeFiles = append(codeFiles, fileInfo)
		}
	}

	result.Findings = map[string]any{
		"code_files": codeFiles,
	}
	result.Passed = len(result.Issues) == 0
	return result
}

// RunGaps validates gaps section structure (R48)
func (ph *Phase) RunGaps() *Result {
	result := &Result{Phase: "gaps"}

	gaps, err := ph.Query.Gaps()
	if err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("design.md Gaps: %v", err))
		return result
	}

	var open, resolved []string
	seen := make(map[string]bool)

	for _, g := range gaps {
		if seen[g.ID] {
			result.Issues = append(result.Issues, fmt.Sprintf("duplicate gap ID: %s", g.ID))
		}
		seen[g.ID] = true

		if g.Resolved {
			resolved = append(resolved, g.ID)
		} else {
			open = append(open, g.ID)
		}
	}

	result.Findings = map[string]any{
		"open":     open,
		"resolved": resolved,
		"total":    len(gaps),
	}
	result.Passed = len(result.Issues) == 0
	return result
}

// FormatText returns a human-readable text representation
func (r *Result) FormatText() string {
	var sb strings.Builder

	if findings, ok := r.Findings.(map[string]any); ok {
		for key, val := range findings {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, val))
		}
	}

	if len(r.Issues) > 0 {
		sb.WriteString("\nissues:\n")
		for _, issue := range r.Issues {
			sb.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}

	status := "OK"
	if !r.Passed {
		status = "FAILED"
	}
	sb.WriteString(fmt.Sprintf("\nphase: %s %s\n", r.Phase, status))

	return sb.String()
}
