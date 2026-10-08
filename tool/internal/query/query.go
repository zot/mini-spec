// CRC: crc-Query.md | Seq: seq-query.md | R102
package query

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
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

// Migrations lists in-flight migration spec files (specs/migrations/*.md, R79).
// Excludes specs/migrations/complete/. Returns relative paths from project root,
// sorted lexically. Returns an empty slice (not error) when no migrations dir exists.
func (q *Query) Migrations() ([]string, error) {
	dir := q.Project.MigrationsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		rel, err := filepath.Rel(q.Project.RootPath, filepath.Join(dir, e.Name()))
		if err != nil {
			rel = filepath.Join("specs", "migrations", e.Name())
		}
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths, nil
}

// specMDToken matches a per-feature spec filename as a whole token, so a
// short name (search.md) is not matched inside a longer one (fuzzy-search.md).
var specMDToken = regexp.MustCompile(`[A-Za-z0-9_-]+[.]md`)

// UnindexedSpecs lists per-feature specs (specs/*.md, non-recursive) not
// referenced in the root index specs/index.md (R102). The index file itself
// and files under specs/migrations/ are excluded; matching is by exact .md
// token. Returns every spec when specs/index.md is absent (nothing indexed
// yet). Relative paths from project root, sorted; empty when all are indexed.
func (q *Query) UnindexedSpecs() ([]string, error) {
	dir := q.Project.SpecsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	indexBytes, err := os.ReadFile(filepath.Join(dir, "index.md"))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	indexed := make(map[string]bool)
	for _, tok := range specMDToken.FindAllString(string(indexBytes), -1) {
		indexed[tok] = true
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") || e.Name() == "index.md" {
			continue
		}
		if indexed[e.Name()] {
			continue
		}
		rel, err := filepath.Rel(q.Project.RootPath, filepath.Join(dir, e.Name()))
		if err != nil {
			rel = filepath.Join("specs", e.Name())
		}
		paths = append(paths, rel)
	}
	sort.Strings(paths)
	return paths, nil
}

// CRC: crc-Query.md | R519
// Traceability reads one code file through the harvest, path as the caller gave it. A file
// the harvest cannot read comes back with the reason, never as a file with no comments.
func (q *Query) Traceability(path string) (parser.FileHarvest, *parser.UnreadFile, error) {
	configured, err := q.Project.Languages()
	if err != nil {
		return parser.FileHarvest{}, nil, err
	}
	return parser.HarvestFile("", path, configured)
}

// CRC: crc-Query.md | R519
// TraceabilityAll is the harvest over every code file in Artifacts, in manifest order, its
// unread list included.
func (q *Query) TraceabilityAll() (parser.Harvest, error) {
	artifacts, err := q.Artifacts()
	if err != nil {
		return parser.Harvest{}, err
	}
	configured, err := q.Project.Languages()
	if err != nil {
		return parser.Harvest{}, err
	}
	return parser.HarvestArtifacts(q.Project.RootPath, artifacts, configured)
}

// CommentFormsEntry is how one extension writes a traceability comment, and what else it
// reads. R518
type CommentFormsEntry struct {
	Extension string                     `json:"extension"`
	Write     minispecsdom.CommentForm   `json:"write"`
	Reads     []minispecsdom.CommentForm `json:"reads"`
}

// CRC: crc-Query.md | R518
// CommentForms reports, for every extension a table reads — built in or configured — the
// comment form to write and the forms read, sorted by extension.
func (q *Query) CommentForms() ([]CommentFormsEntry, error) {
	configured, err := q.Project.Languages()
	if err != nil {
		return nil, err
	}
	var out []CommentFormsEntry
	for _, ext := range minispecsdom.Extensions(configured) {
		lang, _ := minispecsdom.LanguageFor(ext, configured)
		write, reads := minispecsdom.CommentForms(lang)
		out = append(out, CommentFormsEntry{Extension: ext, Write: write, Reads: reads})
	}
	return out, nil
}

// NextIDSource is one file's contribution to a next-ID answer, and the evidence that
// makes the answer checkable. R197
type NextIDSource struct {
	Name    string
	Present bool
	Count   int
}

// NextIDResult is the answer for one class. R189
//
// ByType is populated only for gaps, whose numbering runs a separate sequence per type,
// so a single Next would have to pick one arbitrarily. R192
type NextIDResult struct {
	Class   string
	Next    string
	ByType  map[string]string
	Sources []NextIDSource
}

// CRC: crc-Query.md | Seq: seq-query.md | R189, R191, R192, R193
// NextID returns the next free identifier for a class of permanent, never-reused number.
//
// The class selects the root as well as the count: items are repository-scoped and live
// outside any design root, gaps and requirements are design-scoped. R191
func (q *Query) NextID(class string) (*NextIDResult, error) {
	switch class {
	case "item":
		return NextItemID()
	case "gap":
		return q.nextGapIDs()
	case "req":
		return q.nextReqID()
	default:
		return nil, fmt.Errorf("unknown class %q (expected item, gap, or req)", class)
	}
}

// CRC: crc-Query.md | R190, R191, R194, R195, R198
// NextItemID reads the queue files at the repository root.
//
// A package function rather than a method, deliberately: it touches no design-root
// state, and a signature that demanded a Project would be claiming a dependency it does
// not have. That claim was not free — routing it through the method made the command
// refuse to run in this very repository, whose design roots are tool/ and example/
// while the queue sits above both. R198
func NextItemID() (*NextIDResult, error) {
	repoRoot, err := project.RepoRoot()
	if err != nil {
		return nil, err
	}
	scan, err := parser.ScanTrajectory(repoRoot)
	if err != nil {
		return nil, err
	}
	// Neither file present is not an empty queue -- it is a project with no trajectory
	// layer, which has no next ID at all. R194
	if !scan.AnyPresent() {
		return nil, parser.ErrNoTrajectoryFiles
	}
	res := &NextIDResult{Class: "item", Next: fmt.Sprintf("#%d", scan.MaxItemID()+1)}
	for _, f := range scan.Files {
		res.Sources = append(res.Sources, NextIDSource{Name: f.Name, Present: f.Present, Count: len(f.IDs)})
	}
	return res, nil
}

// CRC: crc-Query.md | R192, R549
// nextGapIDs answers for every gap type, since each runs its own sequence.
func (q *Query) nextGapIDs() (*NextIDResult, error) {
	gaps, err := q.Gaps()
	if err != nil {
		return nil, err
	}
	res := &NextIDResult{
		Class:   "gap",
		ByType:  make(map[string]string, len(parser.GapTypes)),
		Sources: []NextIDSource{{Name: "design.md", Present: true, Count: len(gaps)}},
	}
	for _, t := range parser.GapTypes {
		res.ByType[t] = fmt.Sprintf("%s%d", t, parser.NextGapNum(gaps, t))
	}
	// R549 — a `Tn` named only by a requirement's retired marker is still taken.
	reqs, err := q.Requirements()
	if err != nil {
		return nil, err
	}
	res.ByType["T"] = fmt.Sprintf("T%d", parser.NextTNum(gaps, reqs))
	res.Sources = append(res.Sources, NextIDSource{Name: "requirements.md", Present: true, Count: len(reqs)})
	return res, nil
}

// CRC: crc-Query.md | R193
// nextReqID counts retired requirements too: a retired Rn keeps its number permanently,
// so skipping it would hand out one that is already taken.
func (q *Query) nextReqID() (*NextIDResult, error) {
	reqs, err := q.Requirements()
	if err != nil {
		return nil, err
	}
	maxNum := 0
	for _, r := range reqs {
		if n, err := strconv.Atoi(strings.TrimPrefix(r.ID, "R")); err == nil {
			maxNum = max(maxNum, n)
		}
	}
	return &NextIDResult{
		Class:   "req",
		Next:    fmt.Sprintf("R%d", maxNum+1),
		Sources: []NextIDSource{{Name: "requirements.md", Present: true, Count: len(reqs)}},
	}, nil
}
