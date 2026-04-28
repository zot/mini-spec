// CRC: crc-Phase.md | Seq: seq-phase.md | R44, R45, R46, R47, R48, R49, R50, R76, R87
package phase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
	"github.com/zot/minispec/internal/validate"
)

// Result is the output of a phase validation. Body is the issues-only block
// (without trailing status line); Passed signals success.
type Result struct {
	Phase  string `json:"phase"`
	Passed bool   `json:"passed"`
	Body   string `json:"body,omitempty"`
}

// Phase runs phase-specific validations
type Phase struct {
	Project *project.Project
	Query   *query.Query
}

// New creates a new Phase instance
func New(p *project.Project) *Phase {
	return &Phase{Project: p, Query: query.New(p)}
}

// RunSpec validates spec files exist and are non-empty (R44)
func (ph *Phase) RunSpec() *Result {
	specsDir := filepath.Join(ph.Project.RootPath, "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return &Result{Phase: "spec", Body: fmt.Sprintf("  specs/ directory: %v\n", err)}
	}

	var issues []string
	specCount := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		specCount++
		info, err := os.Stat(filepath.Join(specsDir, e.Name()))
		if err == nil && info.Size() == 0 {
			issues = append(issues, fmt.Sprintf("  empty spec: %s", e.Name()))
		}
	}
	if specCount == 0 {
		issues = append(issues, "  no spec files found in specs/")
	}

	r := &Result{Phase: "spec", Passed: len(issues) == 0}
	if !r.Passed {
		r.Body = strings.Join(issues, "\n") + "\n"
	}
	return r
}

// runSubset runs full validation and emits only the issue categories listed.
func (ph *Phase) runSubset(name string, categories []string) *Result {
	v := validate.New(ph.Project)
	full, err := v.Run()
	if err != nil {
		return &Result{Phase: name, Body: fmt.Sprintf("  %v\n", err)}
	}
	filtered := filterResult(full, categories)
	r := &Result{Phase: name, Passed: !filtered.HasIssues()}
	if !r.Passed {
		body := filtered.FormatText()
		body = strings.TrimSuffix(body, "\nphase: validate FAILED\n")
		body = strings.TrimPrefix(body, "issues:\n")
		r.Body = body
	}
	return r
}

// RunRequirements validates requirement-level issues (R45, R87).
func (ph *Phase) RunRequirements() *Result {
	return ph.runSubset("requirements",
		[]string{"DuplicateReqs", "ReqNumberingGaps", "MissingSpecSources"})
}

// RunDesign validates design-level issues (R46, R87).
func (ph *Phase) RunDesign() *Result {
	return ph.runSubset("design",
		[]string{"UncoveredReqs", "UnknownCRCRefs", "UnlistedDesignFiles",
			"MissingCRCSequences", "OrphanCRCNoReqField"})
}

// RunImplementation validates code-level issues (R47, R87).
func (ph *Phase) RunImplementation() *Result {
	return ph.runSubset("implementation",
		[]string{"MissingArtifacts", "MissingTraceability", "MissingDesignRefs", "MissingImplCoverage"})
}

// RunGaps validates Gaps section structure (R48, R76, R87).
func (ph *Phase) RunGaps() *Result {
	return ph.runSubset("gaps",
		[]string{"DuplicateGapIDs", "CheckboxedPermanent"})
}

// filterResult returns a copy of r with only the listed categories preserved.
func filterResult(r *validate.ValidationResult, keep []string) *validate.ValidationResult {
	out := &validate.ValidationResult{
		UnknownCRCRefs:      map[string][]string{},
		MissingDesignRefs:   map[string][]string{},
		MissingCRCSequences: map[string][]string{},
	}
	for _, k := range keep {
		switch k {
		case "UncoveredReqs":
			out.UncoveredReqs = r.UncoveredReqs
		case "MissingImplCoverage":
			out.MissingImplCoverage = r.MissingImplCoverage
		case "DuplicateReqs":
			out.DuplicateReqs = r.DuplicateReqs
		case "ReqNumberingGaps":
			out.ReqNumberingGaps = r.ReqNumberingGaps
		case "UnknownCRCRefs":
			out.UnknownCRCRefs = r.UnknownCRCRefs
		case "MissingArtifacts":
			out.MissingArtifacts = r.MissingArtifacts
		case "MissingTraceability":
			out.MissingTraceability = r.MissingTraceability
		case "MissingDesignRefs":
			out.MissingDesignRefs = r.MissingDesignRefs
		case "UnlistedDesignFiles":
			out.UnlistedDesignFiles = r.UnlistedDesignFiles
		case "MissingSpecSources":
			out.MissingSpecSources = r.MissingSpecSources
		case "MissingCRCSequences":
			out.MissingCRCSequences = r.MissingCRCSequences
		case "CheckboxedPermanent":
			out.CheckboxedPermanent = r.CheckboxedPermanent
		case "DuplicateGapIDs":
			out.DuplicateGapIDs = r.DuplicateGapIDs
		case "OrphanCRCNoReqField":
			out.OrphanCRCNoReqField = r.OrphanCRCNoReqField
		}
	}
	return out
}

// FormatText returns the phase output: issues block (if any) plus status line.
func (r *Result) FormatText() string {
	status := "OK"
	if !r.Passed {
		status = "FAILED"
	}
	if r.Body == "" {
		return fmt.Sprintf("phase: %s %s\n", r.Phase, status)
	}
	var sb strings.Builder
	sb.WriteString("issues:\n")
	sb.WriteString(r.Body)
	if !strings.HasSuffix(r.Body, "\n") {
		sb.WriteString("\n")
	}
	fmt.Fprintf(&sb, "\nphase: %s %s\n", r.Phase, status)
	return sb.String()
}
