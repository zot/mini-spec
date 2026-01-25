// CRC: crc-Validate.md | Seq: seq-validate.md
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
)

// ValidationResult contains all findings from validation
type ValidationResult struct {
	Requirements RequirementsFindings
	CRCCards     CRCFindings
	Coverage     CoverageFindings
	Artifacts    ArtifactsFindings
	Gaps         GapsFindings
	Issues       []string
}

type RequirementsFindings struct {
	Found    []string            // List of Rn IDs found
	Sources  map[string][]string // source -> []Rn
	Inferred []string            // List of inferred Rn IDs
}

type CRCFindings struct {
	Cards map[string][]string // file -> []Rn
}

type CoverageFindings struct {
	Covered   []string
	Uncovered []string
}

type ArtifactsFindings struct {
	Artifacts []ArtifactFinding
}

type ArtifactFinding struct {
	DesignFile string
	CodeFiles  []CodeFileFinding
}

type CodeFileFinding struct {
	Path    string
	Checked bool
	Exists  bool
}

type GapsFindings struct {
	Gaps []parser.Gap
}

// Validate runs all structural validations
type Validate struct {
	Project *project.Project
	Query   *query.Query
}

// New creates a new Validate instance
func New(p *project.Project) *Validate {
	return &Validate{
		Project: p,
		Query:   query.New(p),
	}
}

// Run executes all validations and returns results
func (v *Validate) Run() (*ValidationResult, error) {
	result := &ValidationResult{
		Requirements: RequirementsFindings{
			Sources: make(map[string][]string),
		},
		CRCCards: CRCFindings{
			Cards: make(map[string][]string),
		},
	}

	// Validate requirements
	if err := v.validateRequirements(result); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("requirements.md: %v", err))
	}

	// Validate CRC cards
	if err := v.validateCRCCards(result); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("CRC cards: %v", err))
	}

	// Compute coverage
	v.computeCoverage(result)

	// Validate artifacts
	if err := v.validateArtifacts(result); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("artifacts: %v", err))
	}

	// Validate gaps
	if err := v.validateGaps(result); err != nil {
		result.Issues = append(result.Issues, fmt.Sprintf("gaps: %v", err))
	}

	// Validate traceability (R29, R42)
	v.validateTraceability(result)

	// Validate artifacts completeness (R40)
	v.validateArtifactsCompleteness(result)

	// Validate spec sources (R41)
	v.validateSpecSources(result)

	// Validate CRC sequences (R43)
	v.validateCRCSequences(result)

	return result, nil
}

func (v *Validate) validateRequirements(result *ValidationResult) error {
	reqs, err := v.Query.Requirements()
	if err != nil {
		return err
	}

	// Build findings
	expectedNum := 1
	for _, r := range reqs {
		result.Requirements.Found = append(result.Requirements.Found, r.ID)
		if r.Inferred {
			result.Requirements.Inferred = append(result.Requirements.Inferred, r.ID)
		}
		if r.Source != "" {
			result.Requirements.Sources[r.Source] = append(result.Requirements.Sources[r.Source], r.ID)
		}

		// Check sequential numbering
		numStr := strings.TrimPrefix(r.ID, "R")
		num, _ := strconv.Atoi(numStr)
		if num != expectedNum {
			result.Issues = append(result.Issues, fmt.Sprintf("non-sequential: expected R%d, found %s", expectedNum, r.ID))
		}
		expectedNum = num + 1
	}

	return nil
}

func (v *Validate) validateCRCCards(result *ValidationResult) error {
	files, err := v.Project.GlobCRCCards()
	if err != nil {
		return err
	}

	// Build set of valid Rn IDs
	validReqs := make(map[string]bool)
	for _, id := range result.Requirements.Found {
		validReqs[id] = true
	}

	for _, path := range files {
		card, err := parser.ParseCRCCard(path)
		if err != nil {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: parse error: %v", filepath.Base(path), err))
			continue
		}

		relPath := filepath.Base(path)
		result.CRCCards.Cards[relPath] = card.Requirements

		if len(card.Requirements) == 0 {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: no Requirements field", relPath))
		}

		// Check for invalid references
		for _, reqID := range card.Requirements {
			if !validReqs[reqID] {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: references unknown %s", relPath, reqID))
			}
		}
	}

	return nil
}

func (v *Validate) computeCoverage(result *ValidationResult) {
	covered := make(map[string]bool)

	for _, reqs := range result.CRCCards.Cards {
		for _, reqID := range reqs {
			covered[reqID] = true
		}
	}

	for _, id := range result.Requirements.Found {
		if covered[id] {
			result.Coverage.Covered = append(result.Coverage.Covered, id)
		} else {
			result.Coverage.Uncovered = append(result.Coverage.Uncovered, id)
		}
	}

	if len(result.Coverage.Uncovered) > 0 {
		result.Issues = append(result.Issues, fmt.Sprintf("uncovered requirements: %s", strings.Join(result.Coverage.Uncovered, ", ")))
	}
}

func (v *Validate) validateArtifacts(result *ValidationResult) error {
	artifacts, err := v.Query.Artifacts()
	if err != nil {
		return err
	}

	for _, art := range artifacts {
		finding := ArtifactFinding{DesignFile: art.DesignFile}

		for _, cf := range art.CodeFiles {
			cfFinding := CodeFileFinding{
				Path:    cf.Path,
				Checked: cf.Checked,
			}

			// Check if file exists
			fullPath := filepath.Join(v.Project.RootPath, cf.Path)
			_, err := os.Stat(fullPath)
			cfFinding.Exists = err == nil

			if !cfFinding.Exists && cf.Checked {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: listed but not found", cf.Path))
			}

			finding.CodeFiles = append(finding.CodeFiles, cfFinding)
		}

		result.Artifacts.Artifacts = append(result.Artifacts.Artifacts, finding)
	}

	return nil
}

func (v *Validate) validateGaps(result *ValidationResult) error {
	gaps, err := v.Query.Gaps()
	if err != nil {
		return err
	}

	result.Gaps.Gaps = gaps

	// Check for duplicate IDs
	seen := make(map[string]bool)
	for _, g := range gaps {
		if seen[g.ID] {
			result.Issues = append(result.Issues, fmt.Sprintf("duplicate gap ID: %s", g.ID))
		}
		seen[g.ID] = true
	}

	return nil
}

func (v *Validate) validateTraceability(result *ValidationResult) {
	for _, art := range result.Artifacts.Artifacts {
		for _, cf := range art.CodeFiles {
			if !cf.Exists {
				continue
			}

			fullPath := filepath.Join(v.Project.RootPath, cf.Path)
			ext := filepath.Ext(cf.Path)
			pattern := v.Project.CommentPattern(ext)
			trace, err := parser.ParseTraceability(fullPath, pattern)
			if err != nil {
				continue
			}

			if len(trace.CRCRefs) == 0 {
				result.Issues = append(result.Issues, fmt.Sprintf("%s: missing traceability comment", cf.Path))
			}

			// R42: Check that referenced design files exist
			allRefs := append(trace.CRCRefs, trace.SeqRefs...)
			for _, ref := range allRefs {
				refPath := v.Project.DesignPath(ref)
				if _, err := os.Stat(refPath); os.IsNotExist(err) {
					result.Issues = append(result.Issues, fmt.Sprintf("%s: references %s which does not exist", cf.Path, ref))
				}
			}
		}
	}
}

// R40: Check all design files in design/ are listed in Artifacts
func (v *Validate) validateArtifactsCompleteness(result *ValidationResult) {
	// Get all design files that should be tracked
	patterns := []string{"crc-*.md", "seq-*.md", "ui-*.md", "test-*.md", "manifest-*.md"}
	designFiles := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(v.Project.DesignPath(pattern))
		if err != nil {
			continue
		}
		for _, m := range matches {
			designFiles[filepath.Base(m)] = true
		}
	}

	// Build set of files listed in Artifacts
	listedFiles := make(map[string]bool)
	for _, art := range result.Artifacts.Artifacts {
		listedFiles[art.DesignFile] = true
	}

	// Check for unlisted design files
	for file := range designFiles {
		if !listedFiles[file] {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: not listed in Artifacts", file))
		}
	}
}

// R41: Check Source fields reference existing spec files
func (v *Validate) validateSpecSources(result *ValidationResult) {
	reqs, err := v.Query.Requirements()
	if err != nil {
		return
	}

	checked := make(map[string]bool)
	for _, r := range reqs {
		if r.Source == "" || checked[r.Source] {
			continue
		}
		checked[r.Source] = true

		specPath := filepath.Join(v.Project.RootPath, r.Source)
		if _, err := os.Stat(specPath); os.IsNotExist(err) {
			result.Issues = append(result.Issues, fmt.Sprintf("%s: referenced as Source but file missing", r.Source))
		}
	}
}

// R43: Check files in CRC Sequences sections exist
func (v *Validate) validateCRCSequences(result *ValidationResult) {
	files, err := v.Project.GlobCRCCards()
	if err != nil {
		return
	}

	for _, path := range files {
		card, err := parser.ParseCRCCard(path)
		if err != nil {
			continue
		}

		relPath := filepath.Base(path)
		for _, seq := range card.Sequences {
			seqPath := v.Project.DesignPath(seq)
			if _, err := os.Stat(seqPath); os.IsNotExist(err) {
				result.Issues = append(result.Issues, fmt.Sprintf("%s Sequences: %s does not exist", relPath, seq))
			}
		}
	}
}

// HasIssues returns true if any issues were found
func (r *ValidationResult) HasIssues() bool {
	return len(r.Issues) > 0
}

// FormatText returns a human-readable text representation
func (r *ValidationResult) FormatText() string {
	var sb strings.Builder

	sb.WriteString("requirements.md:\n")
	sb.WriteString(fmt.Sprintf("  found: %s\n", strings.Join(r.Requirements.Found, ", ")))
	if len(r.Requirements.Inferred) > 0 {
		sb.WriteString(fmt.Sprintf("  inferred: %s\n", strings.Join(r.Requirements.Inferred, ", ")))
	}

	// Sources
	var sources []string
	for src := range r.Requirements.Sources {
		sources = append(sources, src)
	}
	sort.Strings(sources)
	for _, src := range sources {
		reqs := r.Requirements.Sources[src]
		sb.WriteString(fmt.Sprintf("  source %s: %s\n", src, formatRange(reqs)))
	}

	sb.WriteString("\ndesign files:\n")
	for file, reqs := range r.CRCCards.Cards {
		if len(reqs) > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", file, strings.Join(reqs, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("  %s: (no Requirements field)\n", file))
		}
	}

	sb.WriteString("\ncoverage:\n")
	sb.WriteString(fmt.Sprintf("  covered: %s\n", strings.Join(r.Coverage.Covered, ", ")))
	if len(r.Coverage.Uncovered) > 0 {
		sb.WriteString(fmt.Sprintf("  uncovered: %s\n", strings.Join(r.Coverage.Uncovered, ", ")))
	}

	sb.WriteString("\nartifacts:\n")
	for _, art := range r.Artifacts.Artifacts {
		sb.WriteString(fmt.Sprintf("  %s:\n", art.DesignFile))
		for _, cf := range art.CodeFiles {
			mark := " "
			if cf.Checked {
				mark = "x"
			}
			suffix := ""
			if !cf.Exists {
				suffix = " (missing)"
			}
			sb.WriteString(fmt.Sprintf("    [%s] %s%s\n", mark, cf.Path, suffix))
		}
	}

	if len(r.Gaps.Gaps) > 0 {
		sb.WriteString("\ngaps:\n")
		for _, g := range r.Gaps.Gaps {
			mark := " "
			if g.Resolved {
				mark = "x"
			}
			sb.WriteString(fmt.Sprintf("  [%s] %s: %s\n", mark, g.ID, g.Description))
		}
	}

	if len(r.Issues) > 0 {
		sb.WriteString("\nissues:\n")
		for _, issue := range r.Issues {
			sb.WriteString(fmt.Sprintf("  - %s\n", issue))
		}
	}

	return sb.String()
}

func formatRange(reqs []string) string {
	if len(reqs) == 0 {
		return "(none)"
	}
	if len(reqs) == 1 {
		return reqs[0]
	}
	return fmt.Sprintf("%s-%s", reqs[0], reqs[len(reqs)-1])
}
