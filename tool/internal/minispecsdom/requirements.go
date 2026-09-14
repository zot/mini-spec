package minispecsdom

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CRC: crc-Requirements.md | R438, R439
//
// The shapes a requirements file takes: a heading at any level, a `**Source:**` line, and
// the requirement entry in its live and retired forms.
var (
	headingRe    = regexp.MustCompile(`^(#{1,6}) (.*)$`)
	reqHeadRe    = regexp.MustCompile(`^- \*\*(~~)?R(\d+):(?:~~)?\*\*[ \t]*(.*)$`)
	retiredRe    = regexp.MustCompile(`^\(Retired (T\d+) — (?:see (R\d+)|no replacement)\)\s*(.*)$`)
	reqSourceRe  = regexp.MustCompile(`^\*\*Source:\*\*[ \t]*(.*)$`)
	reqKeyRe     = regexp.MustCompile(`^R(\d+)$`)
	retireTnRe   = regexp.MustCompile(`^T\d+$`)
	retireClause = regexp.MustCompile(`^(?:see R\d+|no replacement)$`)
)

// CRC: crc-Requirements.md | R442, R443
var (
	ErrNoRequirement = errors.New("minispecsdom: no entry carries that requirement ID")
	ErrBadReqID      = errors.New("minispecsdom: a requirement ID is R and a number, like R5")
	ErrReqExists     = errors.New("minispecsdom: the document already carries that requirement ID")
	ErrRetired       = errors.New("minispecsdom: the requirement is already retired")
	ErrBadClause     = errors.New("minispecsdom: a retirement clause is `see R<m>` or `no replacement`, and its Tn is `T<n>`")
	ErrManySections  = errors.New("minispecsdom: several sections carry that title; the reader does not pick")
)

// CRC: crc-Requirements.md | R437
//
// Requirements is the requirements schema: it embeds the markdown base, owns the document,
// and reads every heading as a section with its own content and its entries.
type Requirements struct {
	markdownDoc

	sections []*Section
	items    []*Requirement
	unread   []Unread
}

// CRC: crc-Requirements.md | R437, R440
type Section struct {
	Title  string
	Level  int
	Source string
	Parent *Section

	line        int
	own         span // the heading line through the section's own content
	lastContent int  // one past the last non-blank line of the own content
	source      bool
}

// CRC: crc-Requirements.md | R438, R439, R441
type Requirement struct {
	ID          string
	Number      int
	Text        string
	Retired     bool
	RetiredBy   string
	Replacement string
	Section     *Section

	line       int
	head       span
	headText   string // the head line's text after the key, clause included
	deviations []Deviation
}

// CRC: crc-Requirements.md | R441
func (s *Section) Line() int { return s.line }

// CRC: crc-Requirements.md | R441
func (q *Requirement) Line() int { return q.line }

// CRC: crc-Requirements.md | R441
func (q *Requirement) Deviations() []Deviation { return q.deviations }

// CRC: crc-Requirements.md | Seq: seq-requirements.md#1 | R437
func ParseRequirements(src string) *Requirements {
	r := &Requirements{}
	r.parse(src)
	return r
}

// parse reads src with the markdown base and derives the view over the result.
func (r *Requirements) parse(src string) {
	r.parseBase(src)
	r.scan()
}

// reload re-reads the document from its bytes, so the view and the array are rebuilt together.
func (r *Requirements) reload() {
	src, _ := r.doc.Render()
	r.parse(src)
}

func (r *Requirements) Sections() []*Section         { return r.sections }
func (r *Requirements) Requirements() []*Requirement { return r.items }
func (r *Requirements) Unread() []Unread             { return r.unread }

// CRC: crc-Requirements.md | R437
func (r *Requirements) Section(title string) []*Section {
	var out []*Section
	for _, s := range r.sections {
		if s.Title == title {
			out = append(out, s)
		}
	}
	return out
}

// CRC: crc-Requirements.md | R441
func (r *Requirements) Requirement(id string) *Requirement {
	for _, q := range r.items {
		if q.ID == id {
			return q
		}
	}
	return nil
}

// CRC: crc-Requirements.md | Seq: seq-requirements.md#1.2 | R437, R438, R439, R440, R441
//
// scan renders the document to lines and walks them: a heading outside a code group opens
// a section, a keyed bullet opens an entry, the Source line names the section's source, and
// any other line folds into the open entry until a bullet, a blank line or a heading.
func (r *Requirements) scan() {
	r.sections, r.items, r.unread = nil, nil, nil
	src, _ := r.doc.Render()
	seen := map[string]bool{}
	var sec *Section
	var cur *Requirement
	off := 0
	closeSection := func(at int) {
		if sec != nil {
			sec.own.end = at
			sec = nil
		}
	}
	for _, l := range strings.SplitAfter(src, "\n") {
		lineStart := off
		off += len(l)
		line := strings.TrimRight(l, "\n")
		if strings.TrimSpace(line) == "" { // a blank line, or the split's trailing empty piece
			cur = nil
			continue
		}
		if r.inCode(lineStart) {
			if sec != nil {
				sec.lastContent = off
			}
			continue
		}
		if m := headingRe.FindStringSubmatch(line); m != nil {
			closeSection(lineStart)
			cur = nil
			s := &Section{
				Title:       strings.TrimSpace(m[2]),
				Level:       len(m[1]),
				line:        r.doc.Line(lineStart),
				own:         span{lineStart, off},
				lastContent: off,
			}
			s.Parent = r.parentOf(s.Level)
			r.sections = append(r.sections, s)
			sec = s
			continue
		}
		if sec != nil {
			sec.lastContent = off
		}
		if m := reqHeadRe.FindStringSubmatch(line); m != nil {
			cur = r.newRequirement(m, span{lineStart, off}, sec, seen)
			continue
		}
		if m := reqSourceRe.FindStringSubmatch(line); m != nil && sec != nil {
			cur = nil
			if sec.source {
				r.unread = append(r.unread, Unread{r.doc.Line(lineStart), line})
			} else {
				sec.Source, sec.source = strings.TrimSpace(m[1]), true
			}
			continue
		}
		if strings.HasPrefix(line, "- ") {
			cur = nil
			r.unread = append(r.unread, Unread{r.doc.Line(lineStart), line})
			continue
		}
		if cur != nil {
			cur.Text += " " + strings.TrimSpace(line)
		}
	}
	closeSection(r.total)
	r.unread = append(r.unread, unbalanced(r.ctx)...)
	byLine(r.unread)
}

// CRC: crc-Requirements.md | Seq: seq-requirements.md#1.3 | R438, R439, R441
// newRequirement reads one head line into an entry, with its deviations, and records it.
func (r *Requirements) newRequirement(m []string, head span, sec *Section, seen map[string]bool) *Requirement {
	struck, digits, text := m[1] != "", m[2], strings.TrimSpace(m[3])
	q := &Requirement{ID: "R" + digits, Text: text, Retired: struck, Section: sec, line: r.doc.Line(head.start), head: head, headText: text}
	q.Number, _ = strconv.Atoi(digits)
	if struck {
		if c := retiredRe.FindStringSubmatch(text); c != nil {
			q.RetiredBy, q.Replacement, q.Text = c[1], c[2], strings.TrimSpace(c[3])
		} else {
			q.deviations = append(q.deviations, Deviation{"a retired requirement names its Tn and its replacement: `(Retired Tn — see Rm)` or `(Retired Tn — no replacement)`", m[0]})
		}
	}
	if seen[q.ID] {
		q.deviations = append(q.deviations, Deviation{"a requirement ID appears once in the document", q.ID})
	}
	seen[q.ID] = true
	for _, d := range q.deviations {
		r.unread = append(r.unread, Unread{q.line, fmt.Sprintf("%s — %s: %s", q.ID, d.Rule, d.Target)})
	}
	r.items = append(r.items, q)
	return q
}

// parentOf is the nearest section read so far that is shallower than level, or nil at the
// top level.
func (r *Requirements) parentOf(level int) *Section {
	for i := len(r.sections) - 1; i >= 0; i-- {
		if r.sections[i].Level < level {
			return r.sections[i]
		}
	}
	return nil
}

// CRC: crc-Requirements.md | Seq: seq-requirements.md#2.2 | R442, R444
func (r *Requirements) Add(title, id, text string) error {
	if !reqKeyRe.MatchString(id) {
		return ErrBadReqID
	}
	if r.Requirement(id) != nil {
		return ErrReqExists
	}
	found := r.Section(title)
	if len(found) == 0 {
		return ErrNoSection
	}
	if len(found) > 1 {
		return ErrManySections
	}
	sec := found[0]
	text = strings.TrimSpace(text)
	line := "- **" + id + ":** " + text + "\n"
	at := sec.lastContent
	if err := r.doc.Mutate(func() error { return r.replaceSpan(at, at, line) }); err != nil {
		return err
	}
	r.reload()
	got := r.Requirement(id)
	ok := got != nil && got.Text == text && got.Section != nil && got.Section.Title == title && len(got.deviations) == 0
	mustReadBack("Requirements", "Add", id, ok, line, fmt.Sprintf("%+v", got))
	return nil
}

// CRC: crc-Requirements.md | Seq: seq-requirements.md#2.3 | R443, R444
func (r *Requirements) Retire(id, tn, clause string) error {
	if !retireTnRe.MatchString(tn) || !retireClause.MatchString(clause) {
		return ErrBadClause
	}
	q := r.Requirement(id)
	if q == nil {
		return ErrNoRequirement
	}
	if len(q.deviations) > 0 {
		return &DeviationError{Key: id, Deviations: q.deviations}
	}
	if q.Retired {
		return ErrRetired
	}
	line := "- **~~" + id + ":~~** (Retired " + tn + " — " + clause + ") " + q.headText + "\n"
	if err := r.doc.Mutate(func() error { return r.replaceSpan(q.head.start, q.head.end, line) }); err != nil {
		return err
	}
	r.reload()
	got := r.Requirement(id)
	ok := got != nil && got.Retired && got.RetiredBy == tn && strings.HasPrefix(got.Text, q.headText) && len(got.deviations) == 0
	mustReadBack("Requirements", "Retire", id, ok, line, fmt.Sprintf("%+v", got))
	return nil
}
