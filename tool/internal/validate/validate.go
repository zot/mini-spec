// CRC: crc-Validate.md | Seq: seq-validate.md | R68, R69, R70, R72, R76, R78, R84, R85, R86, R88
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
)

// ValidationResult contains issues bucketed by category. R84
type ValidationResult struct {
	UncoveredReqs        []string            // R numbers
	MissingImplCoverage  []string            // R numbers
	DuplicateReqs        []string            // R numbers
	ReqNumberingGaps     []string            // R numbers (missing in sequence)
	UnknownCRCRefs       map[string][]string // file -> []Rn
	MissingArtifacts     []string            // code paths
	MissingTraceability  []string            // code paths
	MissingDesignRefs    map[string][]string // code path -> []missing-ref
	UnlistedDesignFiles  []string            // design filenames
	MissingSpecSources   []string            // spec paths
	MissingCRCSequences  map[string][]string // crc filename -> []seq-ref
	CheckboxedPermanent  []string            // gap IDs
	DuplicateGapIDs      []string            // gap IDs
	OrphanCRCNoReqField  []string            // crc filenames
}

// Validate runs all structural validations
type Validate struct {
	Project *project.Project
	Query   *query.Query
}

// New creates a new Validate instance
func New(p *project.Project) *Validate {
	return &Validate{Project: p, Query: query.New(p)}
}

// Run executes all validations and returns the bucketed result.
func (v *Validate) Run() (*ValidationResult, error) {
	result := &ValidationResult{
		UnknownCRCRefs:      make(map[string][]string),
		MissingDesignRefs:   make(map[string][]string),
		MissingCRCSequences: make(map[string][]string),
	}

	reqs, err := v.Query.Requirements()
	if err != nil {
		return nil, fmt.Errorf("requirements.md: %w", err)
	}
	validReqs, retired, dups, numberingGaps := summarizeRequirements(reqs)
	result.DuplicateReqs = dups
	result.ReqNumberingGaps = numberingGaps

	cards, err := v.parseAllCRCCards()
	if err != nil {
		return nil, err
	}
	for _, c := range cards {
		if len(c.Requirements) == 0 {
			result.OrphanCRCNoReqField = append(result.OrphanCRCNoReqField, filepath.Base(c.Path))
			continue
		}
		for _, ref := range c.Requirements {
			if !validReqs[ref] {
				name := filepath.Base(c.Path)
				result.UnknownCRCRefs[name] = append(result.UnknownCRCRefs[name], ref)
			}
		}
	}

	gaps, err := v.Query.Gaps()
	if err != nil {
		return nil, fmt.Errorf("design.md Gaps: %w", err)
	}
	approvedReqs := approvedGapReqs(gaps)
	seenGap := make(map[string]bool)
	for _, g := range gaps {
		if seenGap[g.ID] {
			result.DuplicateGapIDs = append(result.DuplicateGapIDs, g.ID)
		}
		seenGap[g.ID] = true
		if g.HasCheckbox && (g.Type == "A" || g.Type == "T") {
			result.CheckboxedPermanent = append(result.CheckboxedPermanent, g.ID)
		}
	}

	covered := make(map[string]bool)
	for _, c := range cards {
		for _, ref := range c.Requirements {
			covered[ref] = true
		}
	}
	for id := range approvedReqs {
		covered[id] = true
	}
	for _, r := range reqs {
		if r.Retired || covered[r.ID] {
			continue
		}
		result.UncoveredReqs = append(result.UncoveredReqs, r.ID)
	}

	artifacts, err := v.Query.Artifacts()
	if err != nil {
		return nil, fmt.Errorf("design.md Artifacts: %w", err)
	}
	result.UnlistedDesignFiles = v.unlistedDesignFiles(artifacts)
	result.MissingSpecSources = v.missingSpecSources(reqs)

	for _, c := range cards {
		for _, seq := range c.Sequences {
			if _, err := os.Stat(v.Project.DesignPath(seq)); os.IsNotExist(err) {
				name := filepath.Base(c.Path)
				result.MissingCRCSequences[name] = append(result.MissingCRCSequences[name], seq)
			}
		}
	}

	implCovered := make(map[string]bool)
	for _, art := range artifacts {
		for _, cf := range art.CodeFiles {
			fullPath := filepath.Join(v.Project.RootPath, cf.Path)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				if cf.Checked {
					result.MissingArtifacts = append(result.MissingArtifacts, cf.Path)
				}
				continue
			}

			ext := filepath.Ext(cf.Path)
			pattern := v.Project.CommentPattern(ext)
			closer := v.Project.CommentCloser(ext)
			trace, err := parser.ParseTraceability(fullPath, pattern, closer)
			if err != nil {
				continue
			}
			if len(trace.CRCRefs) == 0 {
				result.MissingTraceability = append(result.MissingTraceability, cf.Path)
			}

			// CRC and Seq refs must resolve to files in design/.
			// dedupAndSortAll handles deduplication of the result list.
			for _, ref := range trace.CRCRefs {
				if _, err := os.Stat(v.Project.DesignPath(ref)); os.IsNotExist(err) {
					result.MissingDesignRefs[cf.Path] = append(result.MissingDesignRefs[cf.Path], ref)
				}
			}
			for _, ref := range trace.SeqRefs {
				if _, err := os.Stat(v.Project.DesignPath(ref)); os.IsNotExist(err) {
					result.MissingDesignRefs[cf.Path] = append(result.MissingDesignRefs[cf.Path], ref)
				}
			}

			for _, ref := range trace.ReqRefs {
				if !validReqs[ref] && !retired[ref] {
					result.MissingDesignRefs[cf.Path] = append(result.MissingDesignRefs[cf.Path], ref)
				}
				implCovered[ref] = true
			}
		}
	}

	for _, r := range reqs {
		if r.Retired || approvedReqs[r.ID] || implCovered[r.ID] {
			continue
		}
		result.MissingImplCoverage = append(result.MissingImplCoverage, r.ID)
	}

	dedupAndSortAll(result)
	return result, nil
}

// summarizeRequirements returns a set of valid Rn IDs (any), the subset that are
// retired, the list of duplicates, and the list of missing numbers in sequence.
func summarizeRequirements(reqs []parser.Requirement) (valid map[string]bool, retired map[string]bool, dups []string, numGaps []string) {
	valid = make(map[string]bool)
	retired = make(map[string]bool)
	seen := make(map[int]bool)
	var nums []int
	for _, r := range reqs {
		valid[r.ID] = true
		if r.Retired {
			retired[r.ID] = true
		}
		n, _ := strconv.Atoi(strings.TrimPrefix(r.ID, "R"))
		if seen[n] {
			dups = append(dups, r.ID)
		}
		seen[n] = true
		nums = append(nums, n)
	}
	if len(nums) > 0 {
		sort.Ints(nums)
		for missing := 1; missing < nums[0]; missing++ {
			numGaps = append(numGaps, fmt.Sprintf("R%d", missing))
		}
		for i := 1; i < len(nums); i++ {
			for missing := nums[i-1] + 1; missing < nums[i]; missing++ {
				numGaps = append(numGaps, fmt.Sprintf("R%d", missing))
			}
		}
	}
	return
}

// parseAllCRCCards walks design/crc-*.md and returns parsed cards.
func (v *Validate) parseAllCRCCards() ([]parser.CRCCard, error) {
	files, err := v.Project.GlobCRCCards()
	if err != nil {
		return nil, err
	}
	cards := make([]parser.CRCCard, 0, len(files))
	for _, p := range files {
		c, err := parser.ParseCRCCard(p)
		if err != nil {
			continue
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (v *Validate) unlistedDesignFiles(artifacts []parser.Artifact) []string {
	patterns := []string{"crc-*.md", "seq-*.md", "ui-*.md", "test-*.md", "manifest-*.md"}
	listed := make(map[string]bool)
	for _, a := range artifacts {
		listed[a.DesignFile] = true
	}
	var unlisted []string
	for _, pat := range patterns {
		matches, _ := filepath.Glob(v.Project.DesignPath(pat))
		for _, m := range matches {
			base := filepath.Base(m)
			if !listed[base] {
				unlisted = append(unlisted, base)
			}
		}
	}
	return unlisted
}

func (v *Validate) missingSpecSources(reqs []parser.Requirement) []string {
	checked := make(map[string]bool)
	var missing []string
	for _, r := range reqs {
		if r.Source == "" || checked[r.Source] {
			continue
		}
		checked[r.Source] = true
		if _, err := os.Stat(filepath.Join(v.Project.RootPath, r.Source)); os.IsNotExist(err) {
			missing = append(missing, r.Source)
		}
	}
	return missing
}

// approvedGapReqRe matches Rn or Rn-Rm in approved-gap descriptions.
var approvedGapReqRe = regexp.MustCompile(`R(\d+)(?:-R(\d+))?`)

// approvedGapReqs extracts requirement IDs referenced by approved (A-type)
// gaps. R65
func approvedGapReqs(gaps []parser.Gap) map[string]bool {
	reqs := make(map[string]bool)
	for _, g := range gaps {
		if g.Type != "A" {
			continue
		}
		for _, m := range approvedGapReqRe.FindAllStringSubmatch(g.Description, -1) {
			lo, _ := strconv.Atoi(m[1])
			hi := lo
			if m[2] != "" {
				hi, _ = strconv.Atoi(m[2])
			}
			for n := lo; n <= hi; n++ {
				reqs[fmt.Sprintf("R%d", n)] = true
			}
		}
	}
	return reqs
}

// dedupAndSortAll deduplicates and sorts every list field in the result.
func dedupAndSortAll(r *ValidationResult) {
	r.UncoveredReqs = dedupReqIDs(r.UncoveredReqs)
	r.MissingImplCoverage = dedupReqIDs(r.MissingImplCoverage)
	r.DuplicateReqs = dedupReqIDs(r.DuplicateReqs)
	r.ReqNumberingGaps = dedupReqIDs(r.ReqNumberingGaps)
	r.MissingArtifacts = dedupStrings(r.MissingArtifacts)
	r.MissingTraceability = dedupStrings(r.MissingTraceability)
	r.UnlistedDesignFiles = dedupStrings(r.UnlistedDesignFiles)
	r.MissingSpecSources = dedupStrings(r.MissingSpecSources)
	r.CheckboxedPermanent = dedupStrings(r.CheckboxedPermanent)
	r.DuplicateGapIDs = dedupStrings(r.DuplicateGapIDs)
	r.OrphanCRCNoReqField = dedupStrings(r.OrphanCRCNoReqField)
	for k, v := range r.UnknownCRCRefs {
		r.UnknownCRCRefs[k] = dedupReqIDs(v)
	}
	for k, v := range r.MissingDesignRefs {
		r.MissingDesignRefs[k] = dedupStrings(v)
	}
	for k, v := range r.MissingCRCSequences {
		r.MissingCRCSequences[k] = dedupStrings(v)
	}
}

func dedupStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func dedupReqIDs(in []string) []string {
	seen := make(map[int]bool, len(in))
	nums := make([]int, 0, len(in))
	for _, id := range in {
		n, err := strconv.Atoi(strings.TrimPrefix(id, "R"))
		if err != nil || seen[n] {
			continue
		}
		seen[n] = true
		nums = append(nums, n)
	}
	sort.Ints(nums)
	out := make([]string, len(nums))
	for i, n := range nums {
		out[i] = fmt.Sprintf("R%d", n)
	}
	return out
}

// FormatRanges collapses consecutive Rn IDs into hyphenated ranges. R85
// Inputs are expected to be deduplicated and sorted; non-Rn entries are ignored.
func FormatRanges(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	nums := make([]int, 0, len(ids))
	for _, id := range ids {
		n, err := strconv.Atoi(strings.TrimPrefix(id, "R"))
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}
	if len(nums) == 0 {
		return ""
	}

	var parts []string
	start, prev := nums[0], nums[0]
	flush := func() {
		if start == prev {
			parts = append(parts, fmt.Sprintf("R%d", start))
		} else {
			parts = append(parts, fmt.Sprintf("R%d-%d", start, prev))
		}
	}
	for i := 1; i < len(nums); i++ {
		if nums[i] == prev+1 {
			prev = nums[i]
			continue
		}
		flush()
		start, prev = nums[i], nums[i]
	}
	flush()
	return strings.Join(parts, ", ")
}

// HasIssues returns true if any issues were found.
func (r *ValidationResult) HasIssues() bool {
	return len(r.UncoveredReqs) > 0 ||
		len(r.MissingImplCoverage) > 0 ||
		len(r.DuplicateReqs) > 0 ||
		len(r.ReqNumberingGaps) > 0 ||
		len(r.UnknownCRCRefs) > 0 ||
		len(r.MissingArtifacts) > 0 ||
		len(r.MissingTraceability) > 0 ||
		len(r.MissingDesignRefs) > 0 ||
		len(r.UnlistedDesignFiles) > 0 ||
		len(r.MissingSpecSources) > 0 ||
		len(r.MissingCRCSequences) > 0 ||
		len(r.CheckboxedPermanent) > 0 ||
		len(r.DuplicateGapIDs) > 0 ||
		len(r.OrphanCRCNoReqField) > 0
}

// FormatText returns the issues-only text report. R84, R88
func (r *ValidationResult) FormatText() string {
	if !r.HasIssues() {
		return "phase: validate OK\n"
	}

	var sb strings.Builder
	sb.WriteString("issues:\n")

	if s := FormatRanges(r.UncoveredReqs); s != "" {
		fmt.Fprintf(&sb, "  uncovered requirements: %s\n", s)
	}
	if s := FormatRanges(r.MissingImplCoverage); s != "" {
		fmt.Fprintf(&sb, "  missing impl coverage: %s\n", s)
	}
	if s := FormatRanges(r.DuplicateReqs); s != "" {
		fmt.Fprintf(&sb, "  duplicate requirements: %s\n", s)
	}
	if s := FormatRanges(r.ReqNumberingGaps); s != "" {
		fmt.Fprintf(&sb, "  numbering gaps: %s\n", s)
	}
	if len(r.UnknownCRCRefs) > 0 {
		fmt.Fprintf(&sb, "  unknown CRC refs: %s\n", formatFileMap(r.UnknownCRCRefs, FormatRanges))
	}
	if len(r.MissingArtifacts) > 0 {
		fmt.Fprintf(&sb, "  missing artifacts: %s\n", strings.Join(r.MissingArtifacts, ", "))
	}
	if len(r.MissingTraceability) > 0 {
		fmt.Fprintf(&sb, "  missing traceability: %s\n", strings.Join(r.MissingTraceability, ", "))
	}
	if len(r.MissingDesignRefs) > 0 {
		fmt.Fprintf(&sb, "  missing design refs: %s\n", formatFileMap(r.MissingDesignRefs, joinComma))
	}
	if len(r.UnlistedDesignFiles) > 0 {
		fmt.Fprintf(&sb, "  unlisted design files: %s\n", strings.Join(r.UnlistedDesignFiles, ", "))
	}
	if len(r.MissingSpecSources) > 0 {
		fmt.Fprintf(&sb, "  missing spec sources: %s\n", strings.Join(r.MissingSpecSources, ", "))
	}
	if len(r.MissingCRCSequences) > 0 {
		fmt.Fprintf(&sb, "  CRC sequences not found: %s\n", formatFileMap(r.MissingCRCSequences, joinComma))
	}
	if len(r.OrphanCRCNoReqField) > 0 {
		fmt.Fprintf(&sb, "  CRCs without Requirements field: %s\n", strings.Join(r.OrphanCRCNoReqField, ", "))
	}
	if len(r.CheckboxedPermanent) > 0 {
		fmt.Fprintf(&sb, "  permanent gaps with checkbox: %s\n", strings.Join(r.CheckboxedPermanent, ", "))
	}
	if len(r.DuplicateGapIDs) > 0 {
		fmt.Fprintf(&sb, "  duplicate gap IDs: %s\n", strings.Join(r.DuplicateGapIDs, ", "))
	}

	sb.WriteString("\nphase: validate FAILED\n")
	return sb.String()
}

// formatFileMap renders a map of file -> []ref entries, sorted by key, using
// the supplied renderer to stringify each value list.
func formatFileMap(m map[string][]string, render func([]string) string) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s → %s", k, render(m[k])))
	}
	return strings.Join(parts, "; ")
}

func joinComma(s []string) string { return strings.Join(s, ", ") }
