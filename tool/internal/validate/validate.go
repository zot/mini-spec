// CRC: crc-Validate.md | Seq: seq-validate.md | R68, R69, R70, R72, R76, R78, R84, R85, R86, R88, R90, R91, R92, R93, R97, R98, R99, R100, R101, R188
package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/alarm"
	"github.com/zot/minispec/internal/parser"
	"github.com/zot/minispec/internal/project"
	"github.com/zot/minispec/internal/query"
	"github.com/zot/minispec/internal/minispecsdom"
)

// ValidationResult contains issues bucketed by category. R84
type ValidationResult struct {
	UncoveredReqs         []string            // R numbers
	MissingImplCoverage   []string            // R numbers
	DuplicateReqs         []string            // R numbers
	ReqNumberingGaps      []string            // R numbers (missing in sequence)
	UnknownCRCRefs        map[string][]string // file -> []Rn
	MissingArtifacts      []string            // code paths
	MissingTraceability   []string            // code paths
	MissingDesignRefs     map[string][]string // code path -> []missing-ref
	UnlistedDesignFiles   []string            // design filenames
	MissingSpecSources    []string            // spec paths
	MalformedSpecSources  []string            // Source values that don't look like clean .md paths (R91)
	SuspiciousSourceLines []string            // lines that look like Source markers but don't match the canonical pattern (R91)
	MissingCRCSequences   map[string][]string // crc filename -> []seq-ref
	MissingSeqFragments   map[string][]string // code path -> []ref with unresolved #fragment (R97)
	SeqNumberingGaps      map[string][]string // seq filename -> []missing dotted id (R98, R99)
	SeqDuplicateIDs       map[string][]string // seq filename -> []duplicated dotted id (R100)
	CheckboxedPermanent   []string            // gap IDs
	DuplicateGapIDs       []string            // gap IDs
	OrphanCRCNoReqField   []string            // crc filenames
	// VoidedAlarms are recorded fault injections whose proof has expired, or whose
	// injection site no longer resolves. R179, R182, R183
	VoidedAlarms []string
	// ReaderDisagreement is the second opinion over the design documents the dependency's
	// readers own — the gaps section, requirements.md, every test design: an independent
	// line scan against what the reader returned, per file. A reader that lost a file's
	// tail to one unclosed span reports nothing wrong; the line scan says how much of the
	// file it actually saw. R327
	ReaderDisagreement []string
	// Unread is what those readers could not read, per file — coverage, not an issue. R328
	Unread map[string]int
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
		MissingSeqFragments: make(map[string][]string),
		SeqNumberingGaps:    make(map[string][]string),
		SeqDuplicateIDs:     make(map[string][]string),
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
	result.MissingSpecSources, result.MalformedSpecSources = v.checkSpecSources(reqs)
	if issues, err := parser.ScanSourceLineIssues(v.Project.RequirementsPath()); err == nil {
		for _, iss := range issues {
			result.SuspiciousSourceLines = append(result.SuspiciousSourceLines,
				fmt.Sprintf("line %d: %s", iss.LineNum, strings.TrimSpace(iss.Line)))
		}
	}

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
				file, fragment := parser.SplitSeqRef(ref)
				if _, err := os.Stat(v.Project.DesignPath(file)); os.IsNotExist(err) {
					result.MissingDesignRefs[cf.Path] = append(result.MissingDesignRefs[cf.Path], ref)
					continue
				}
				if fragment == "" {
					continue
				}
				doc, err := parser.ParseSeqDoc(v.Project.DesignPath(file))
				if err != nil || !doc.Has(fragment) {
					result.MissingSeqFragments[cf.Path] = append(result.MissingSeqFragments[cf.Path], ref)
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

	v.validateSeqNumbering(result)

	v.checkAlarmFreshness(result)
	v.checkReaderAgreement(result)
	dedupAndSortAll(result)
	return result, nil
}

// validateSeqNumbering parses every seq-*.md file and records per-K
// contiguity gaps and duplicate dotted ids. Unnumbered files are skipped.
// R98, R99, R100, R101
func (v *Validate) validateSeqNumbering(result *ValidationResult) {
	matches, err := filepath.Glob(v.Project.DesignPath("seq-*.md"))
	if err != nil {
		return
	}
	for _, path := range matches {
		doc, err := parser.ParseSeqDoc(path)
		if err != nil || !doc.Numbered() {
			continue
		}
		name := filepath.Base(path)
		if gaps := doc.NumberingGaps(); len(gaps) > 0 {
			result.SeqNumberingGaps[name] = append(result.SeqNumberingGaps[name], gaps...)
		}
		if len(doc.Dupes) > 0 {
			result.SeqDuplicateIDs[name] = append(result.SeqDuplicateIDs[name], doc.Dupes...)
		}
	}
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

// isCleanSpecPath reports whether s looks like a relative .md path with no
// embedded annotations. Allowed: A-Z, a-z, 0-9, `_`, `.`, `/`, `-`. Rejected:
// leading `/`, leading `-`, missing `.md` suffix, any other character (spaces,
// parens, backticks, etc.). R91
func isCleanSpecPath(s string) bool {
	if s == "" || !strings.HasSuffix(s, ".md") {
		return false
	}
	if s[0] == '/' || s[0] == '-' {
		return false
	}
	for _, ch := range s {
		switch {
		case ch >= 'A' && ch <= 'Z':
		case ch >= 'a' && ch <= 'z':
		case ch >= '0' && ch <= '9':
		case ch == '_', ch == '.', ch == '/', ch == '-':
		default:
			return false
		}
	}
	return true
}

// checkSpecSources walks every Source path in every requirement and bins them
// into "missing" (clean path but file not on disk) and "malformed" (path
// shape doesn't match — has spaces, parens, absolute leading slash, missing
// .md suffix, etc.). R90, R91
func (v *Validate) checkSpecSources(reqs []parser.Requirement) (missing, malformed []string) {
	checked := make(map[string]bool)
	for _, r := range reqs {
		for _, src := range r.Sources {
			if checked[src] {
				continue
			}
			checked[src] = true
			if !isCleanSpecPath(src) {
				malformed = append(malformed, src)
				continue
			}
			if _, ok := v.Project.ResolveSpecSource(src); !ok {
				missing = append(missing, src)
			}
		}
	}
	return
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
	r.MalformedSpecSources = dedupStrings(r.MalformedSpecSources)
	r.SuspiciousSourceLines = dedupStrings(r.SuspiciousSourceLines)
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
	for k, v := range r.MissingSeqFragments {
		r.MissingSeqFragments[k] = dedupStrings(v)
	}
	for k, v := range r.SeqNumberingGaps {
		r.SeqNumberingGaps[k] = dedupStrings(v)
	}
	for k, v := range r.SeqDuplicateIDs {
		r.SeqDuplicateIDs[k] = dedupStrings(v)
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
		len(r.MalformedSpecSources) > 0 ||
		len(r.SuspiciousSourceLines) > 0 ||
		len(r.MissingCRCSequences) > 0 ||
		len(r.MissingSeqFragments) > 0 ||
		len(r.SeqNumberingGaps) > 0 ||
		len(r.SeqDuplicateIDs) > 0 ||
		len(r.CheckboxedPermanent) > 0 ||
		len(r.DuplicateGapIDs) > 0 ||
		len(r.OrphanCRCNoReqField) > 0 ||
		len(r.VoidedAlarms) > 0 ||
		len(r.ReaderDisagreement) > 0
}

// FormatText returns the issues-only text report. R84, R88
func (r *ValidationResult) FormatText() string {
	if !r.HasIssues() {
		return r.unreadNote() + "phase: validate OK\n"
	}

	var sb strings.Builder
	sb.WriteString("issues:\n")
	// R327 — listed first, because every finding below it reads through the document
	// reader alone.
	if len(r.ReaderDisagreement) > 0 {
		sb.WriteString("  the two readers disagree:\n")
		for _, d := range r.ReaderDisagreement {
			fmt.Fprintf(&sb, "    %s\n", d)
		}
	}

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
	if len(r.MalformedSpecSources) > 0 {
		fmt.Fprintf(&sb, "  malformed Source values: %s\n", strings.Join(r.MalformedSpecSources, "; "))
	}
	if len(r.SuspiciousSourceLines) > 0 {
		fmt.Fprintf(&sb, "  suspicious Source lines: %s\n", strings.Join(r.SuspiciousSourceLines, "; "))
	}
	if len(r.MissingCRCSequences) > 0 {
		fmt.Fprintf(&sb, "  CRC sequences not found: %s\n", formatFileMap(r.MissingCRCSequences, joinComma))
	}
	if len(r.MissingSeqFragments) > 0 {
		fmt.Fprintf(&sb, "  missing seq anchors: %s\n", formatFileMap(r.MissingSeqFragments, joinComma))
	}
	if len(r.SeqNumberingGaps) > 0 {
		fmt.Fprintf(&sb, "  seq numbering gaps: %s\n", formatFileMap(r.SeqNumberingGaps, joinComma))
	}
	if len(r.SeqDuplicateIDs) > 0 {
		fmt.Fprintf(&sb, "  seq duplicate IDs: %s\n", formatFileMap(r.SeqDuplicateIDs, joinComma))
	}
	if len(r.VoidedAlarms) > 0 {
		// One line each rather than a joined list: every entry names a different
		// document, test and site, and a comma-joined run of those is unreadable.
		sb.WriteString("  fire alarms whose proof has expired:\n")
		for _, a := range r.VoidedAlarms {
			fmt.Fprintf(&sb, "    %s\n", a)
		}
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

	sb.WriteString(r.sourceFixInstructions())

	sb.WriteString(r.unreadNote())
	sb.WriteString("\nphase: validate FAILED\n")
	return sb.String()
}

// sourceFormatFix describes the canonical `**Source:**` format. R92
const sourceFormatFix = `  Source lines in requirements.md must match this exact format:
    **Source:** path/to/spec.md
  Multiple sources are allowed, comma-separated:
    **Source:** path/a.md, path/b.md
  Each path must be relative (no leading slash), end in ` + "`.md`" + `, and
  contain no annotations, parenthetical comments, spaces, or backticks. If
  context about a source needs to be recorded, put it in the requirement
  text or in the spec file itself, not the Source line.
`

// missingSourceFix names the three legitimate repairs for a Source whose spec
// is gone, and forecloses the fourth — deleting the orphaned requirements —
// because the numbering gap that opens invites a renumber, and a renumber
// repoints every anchor while leaving every check green. R188
const missingSourceFix = `  A missing Source means its spec was renamed, merged, or deleted. Repair by
  case — never by deleting the requirements:
    renamed  rewrite the **Source:** to the new path
    merged   repoint the **Source:** at the absorbing spec; nothing retires
    deleted  retire each requirement (minispec update retire <Rn> - "<spec>
             deleted"), move them under a **Source:** specs/deleted.md block,
             and record the spec's name, a one-line description, and its
             requirement numbers in specs/deleted.md
  Requirement numbers are never renumbered and never reused. Deleting the
  orphaned requirements opens a numbering gap, and renumbering to close it
  points every design and code anchor at a different requirement — which
  nothing detects, because all the numbers still exist.
`

// sourceFixInstructions returns the crank-handle block for whichever Source
// diagnostics fired, or "" when none did. Both halves sit under one header.
// R92, R188
func (r *ValidationResult) sourceFixInstructions() string {
	var halves []string
	if len(r.MalformedSpecSources) > 0 || len(r.SuspiciousSourceLines) > 0 {
		halves = append(halves, sourceFormatFix)
	}
	if len(r.MissingSpecSources) > 0 {
		halves = append(halves, missingSourceFix)
	}
	if len(halves) == 0 {
		return ""
	}
	// Each half ends in a newline, so joining on one puts a blank line between them.
	return "\nfix instructions:\n" + strings.Join(halves, "\n")
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

// CRC: crc-Validate.md | Seq: seq-alarm-freshness.md#2.1 | R179, R183, R184
// checkAlarmFreshness reports recorded fault injections whose proof has expired.
//
// Only stale and unresolvable alarms reach the result. An alarm that was never recorded
// as pulled, or that names no injection site, is real information and belongs to
// `query alarms`: a project adopting the convention carries many of each, the counts
// fall slowly, and a line reporting a non-zero number on every run for months is the
// recurring nag this project distinguishes from a closable gripe.
//
// Silent when anything goes wrong reading the documents. This check is an addition to
// an existing validator, and a malformed test design must not break the checks a
// project already depends on.
func (v *Validate) checkAlarmFreshness(result *ValidationResult) {
	docs, err := filepath.Glob(filepath.Join(v.Project.DesignDir, "test-*.md"))
	if err != nil || len(docs) == 0 {
		return
	}
	var alarms []parser.Alarm
	for _, d := range docs {
		found, err := parser.ParseTestDoc(d)
		if err != nil {
			continue
		}
		alarms = append(alarms, found...)
	}
	if len(alarms) == 0 {
		return
	}
	git := project.NewGit(v.Project.RootPath)
	for _, a := range alarm.Voided(alarm.Assess(alarms, git)) {
		switch a.State {
		case alarm.Stale:
			result.VoidedAlarms = append(result.VoidedAlarms, fmt.Sprintf(
				"%s: %s — proof pulled %s, but %s changed %s",
				a.Alarm.Doc, a.Alarm.Test, a.Alarm.Pulled.Format("2006-01-02"), a.Site, a.Changed))
		case alarm.Unresolvable:
			result.VoidedAlarms = append(result.VoidedAlarms, fmt.Sprintf(
				"%s: %s — git cannot resolve %s, so the anchor no longer points at anything",
				a.Alarm.Doc, a.Alarm.Test, a.Site))
		}
	}
}

// R328
// unreadNote is the coverage statement: what the design-document readers could not read,
// printed whether or not anything else fired, because a reader takes silence about coverage
// as a claim of completeness.
func (r *ValidationResult) unreadNote() string {
	if len(r.Unread) == 0 {
		return ""
	}
	names := make([]string, 0, len(r.Unread))
	total := 0
	for name, n := range r.Unread {
		names = append(names, name)
		total += n
	}
	sort.Strings(names)
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = fmt.Sprintf("%s (%d)", name, r.Unread[name])
	}
	return fmt.Sprintf("note: %d line(s) were not read — %s. An entry-like line outside the shape,\n"+
		"      or a group never closed, which takes the rest of its file with it; every check\n"+
		"      above is blind to what they hold.\n", total, strings.Join(parts, ", "))
}

var (
	gapLineRe  = regexp.MustCompile(`^\s*- (?:\[[ x]\] )?([SRDCIOAT]\d+):`)
	reqLineRe  = regexp.MustCompile(`^- \*\*(?:~~)?(R\d+):`)
	gapsHeadRe = regexp.MustCompile(`^## Gaps\s*$`)
)

// CRC: crc-Validate.md | R327, R328
// checkReaderAgreement is the second opinion over the design documents the dependency's
// readers own. The line scan is deliberately naive — a regex over lines, bounded for the
// gaps section by its heading and the next level-2 heading — so it cannot share the
// reader's blind spot: a reader that lost a file's tail to one unclosed span agrees with
// itself forever, and only a scan built on nothing it built on can say how much of the
// file it actually saw. Measured 2026-09-07 on this repository's test-Update.md: six
// alarm entries by line, two by the reader, one unread line reported.
func (v *Validate) checkReaderAgreement(result *ValidationResult) {
	findings, unread := readerAgreement(v.Project.DesignDir)
	result.ReaderDisagreement = append(result.ReaderDisagreement, findings...)
	if len(unread) == 0 {
		return
	}
	if result.Unread == nil {
		result.Unread = map[string]int{}
	}
	for name, n := range unread {
		result.Unread[name] = n
	}
}

// readerAgreement runs the three comparisons over a design directory and returns the
// findings and the per-file unread counts. Separate from the Validate so a test can run
// it over a directory with no project.
func readerAgreement(designDir string) (findings []string, unread map[string]int) {
	unread = map[string]int{}
	note := func(name string, n int) {
		if n > 0 {
			unread[name] = n
		}
	}
	designPath := filepath.Join(designDir, "design.md")
	if src, err := os.ReadFile(designPath); err == nil {
		if gaps, left, gerr := parser.ParseGapsReport(designPath); gerr == nil {
			byReader := map[string]bool{}
			for _, g := range gaps {
				byReader[g.ID] = true
			}
			byLine := scanIDs(gapsSection(string(src)), gapLineRe)
			findings = append(findings, disagreements("design.md", "gap", byLine, byReader)...)
			note("design.md", len(left))
		}
	}
	reqPath := filepath.Join(designDir, "requirements.md")
	if src, err := os.ReadFile(reqPath); err == nil {
		if reqs, left, rerr := parser.ParseRequirementsReport(reqPath); rerr == nil {
			byReader := map[string]bool{}
			for _, r := range reqs {
				byReader[r.ID] = true
			}
			byLine := scanIDs(strings.Split(string(src), "\n"), reqLineRe)
			findings = append(findings, disagreements("requirements.md", "requirement", byLine, byReader)...)
			note("requirements.md", len(left))
		}
	}
	// Test designs compare entry counts rather than titles: a title is display, and the
	// reader's may differ from the heading's text.
	docs, _ := filepath.Glob(filepath.Join(designDir, "test-*.md"))
	for _, doc := range docs {
		src, err := os.ReadFile(doc)
		if err != nil {
			continue
		}
		byLine := 0
		for _, line := range strings.Split(string(src), "\n") {
			if strings.HasPrefix(line, "## Test:") {
				byLine++
			}
		}
		td := minispecsdom.ParseTestDoc(string(src))
		name := filepath.Base(doc)
		if byReader := len(td.Tests()); byReader != byLine {
			findings = append(findings, fmt.Sprintf("%s: %d test entries by line, %d by the document reader", name, byLine, byReader))
		}
		note(name, len(td.Unread()))
	}
	return findings, unread
}

// gapsSection returns the lines of the design document's `## Gaps` section, the heading
// exclusive to the next level-2 heading: the line scan's bound.
func gapsSection(src string) []string {
	var lines []string
	in := false
	for _, line := range strings.Split(src, "\n") {
		switch {
		case gapsHeadRe.MatchString(line):
			in = true
		case strings.HasPrefix(line, "## "):
			in = false
		case in:
			lines = append(lines, line)
		}
	}
	return lines
}

// scanIDs collects the IDs the pattern captures, one line at a time — the naive reading
// the document reader is held against.
func scanIDs(lines []string, pattern *regexp.Regexp) map[string]bool {
	ids := map[string]bool{}
	for _, line := range lines {
		if m := pattern.FindStringSubmatch(line); m != nil {
			ids[m[1]] = true
		}
	}
	return ids
}

// disagreements names the IDs one reading has and the other does not, both ways, sorted,
// consecutive numbers collapsed to ranges: a reader that lost a file's tail names hundreds.
func disagreements(file, kind string, byLine, byReader map[string]bool) []string {
	var out []string
	if only := onlyIn(byLine, byReader); len(only) > 0 {
		out = append(out, fmt.Sprintf("%s: the line scan read %s %s that the document reader returned no entry for", file, kind, idRanges(only)))
	}
	if only := onlyIn(byReader, byLine); len(only) > 0 {
		out = append(out, fmt.Sprintf("%s: the document reader returned %s %s that the line scan did not read", file, kind, idRanges(only)))
	}
	return out
}

// idRanges collapses sorted IDs of one letter-prefix shape into `O3-O7` runs per prefix.
func idRanges(ids []string) string {
	byPrefix := map[string][]int{}
	var order []string
	for _, id := range ids {
		prefix, n, numbered := splitID(id)
		if !numbered {
			byPrefix[id] = nil
			order = append(order, id)
			continue
		}
		if _, seen := byPrefix[prefix]; !seen {
			order = append(order, prefix)
		}
		byPrefix[prefix] = append(byPrefix[prefix], n)
	}
	var parts []string
	for _, prefix := range order {
		nums := byPrefix[prefix]
		if nums == nil {
			parts = append(parts, prefix)
			continue
		}
		sort.Ints(nums)
		start, prev := nums[0], nums[0]
		flush := func() {
			if start == prev {
				parts = append(parts, fmt.Sprintf("%s%d", prefix, start))
			} else {
				parts = append(parts, fmt.Sprintf("%s%d-%s%d", prefix, start, prefix, prev))
			}
		}
		for _, n := range nums[1:] {
			if n == prev+1 {
				prev = n
				continue
			}
			flush()
			start, prev = n, n
		}
		flush()
	}
	return strings.Join(parts, ", ")
}

// splitID divides an ID like `O37` into its letter prefix and its number. An ID that does
// not end in one is not numbered, and stands alone as its own prefix.
func splitID(id string) (prefix string, n int, numbered bool) {
	i := 0
	for i < len(id) && (id[i] < '0' || id[i] > '9') {
		i++
	}
	n, err := strconv.Atoi(id[i:])
	if err != nil {
		return id, 0, false
	}
	return id[:i], n, true
}

func onlyIn(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
