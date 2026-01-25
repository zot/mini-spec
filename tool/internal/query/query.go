// CRC: crc-Query.md | Seq: seq-query.md
package query

import (
	"path/filepath"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
)

// Query provides read-only operations on design files
type Query struct {
	Project *project.Project
}

// New creates a new Query instance
func New(p *project.Project) *Query {
	return &Query{Project: p}
}

// Requirements lists all requirements from requirements.md
func (q *Query) Requirements() ([]parser.Requirement, error) {
	return parser.ParseRequirements(q.Project.RequirementsPath())
}

// CoverageResult maps requirement IDs to files that reference them
type CoverageResult struct {
	Coverage map[string][]string // Rn -> []file paths
	CRCCards map[string][]string // file -> []Rn
}

// Coverage shows which design files reference each requirement
func (q *Query) Coverage() (*CoverageResult, error) {
	result := &CoverageResult{
		Coverage: make(map[string][]string),
		CRCCards: make(map[string][]string),
	}

	// Get all requirements first to initialize map
	reqs, err := q.Requirements()
	if err != nil {
		return nil, err
	}
	for _, r := range reqs {
		result.Coverage[r.ID] = []string{}
	}

	// Parse all CRC cards
	files, err := q.Project.GlobCRCCards()
	if err != nil {
		return nil, err
	}

	for _, path := range files {
		card, err := parser.ParseCRCCard(path)
		if err != nil {
			continue // Skip unparseable files
		}
		result.CRCCards[card.Path] = card.Requirements
		for _, reqID := range card.Requirements {
			result.Coverage[reqID] = append(result.Coverage[reqID], card.Path)
		}
	}

	return result, nil
}

// Uncovered returns requirements with no design file references
func (q *Query) Uncovered() ([]string, error) {
	cov, err := q.Coverage()
	if err != nil {
		return nil, err
	}

	var uncovered []string
	for id, files := range cov.Coverage {
		if len(files) == 0 {
			uncovered = append(uncovered, id)
		}
	}
	return uncovered, nil
}

// OrphanDesigns returns CRC cards with no/empty Requirements field
func (q *Query) OrphanDesigns() ([]string, error) {
	files, err := q.Project.GlobCRCCards()
	if err != nil {
		return nil, err
	}

	var orphans []string
	for _, path := range files {
		card, err := parser.ParseCRCCard(path)
		if err != nil {
			continue
		}
		if len(card.Requirements) == 0 {
			orphans = append(orphans, path)
		}
	}
	return orphans, nil
}

// Artifacts lists all artifacts with checkbox states
func (q *Query) Artifacts() ([]parser.Artifact, error) {
	return parser.ParseArtifacts(q.Project.DesignMdPath())
}

// Gaps lists all gap items from design.md
func (q *Query) Gaps() ([]parser.Gap, error) {
	return parser.ParseGaps(q.Project.DesignMdPath())
}

// Traceability checks a code file for CRC/Seq comments
func (q *Query) Traceability(path string) (parser.Traceability, error) {
	ext := filepath.Ext(path)
	pattern := q.Project.CommentPattern(ext)
	return parser.ParseTraceability(path, pattern)
}

// TraceabilityAll checks all code files in Artifacts
func (q *Query) TraceabilityAll() (map[string]parser.Traceability, error) {
	artifacts, err := q.Artifacts()
	if err != nil {
		return nil, err
	}

	result := make(map[string]parser.Traceability)
	for _, art := range artifacts {
		for _, cf := range art.CodeFiles {
			trace, err := q.Traceability(cf.Path)
			if err != nil {
				// File might not exist yet
				result[cf.Path] = parser.Traceability{}
				continue
			}
			result[cf.Path] = trace
		}
	}
	return result, nil
}

// CommentPatterns returns the configured comment patterns per file extension
func (q *Query) CommentPatterns() map[string]string {
	return q.Project.Config.CommentPatterns
}
