package minispecsdom

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func loadReqs(t *testing.T) (*Requirements, string) {
	t.Helper()
	src, err := os.ReadFile("testdata/requirements-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return ParseRequirements(string(src)), string(src)
}

func reqIDs(r *Requirements) string {
	var ids []string
	for _, q := range r.Requirements() {
		ids = append(ids, q.ID)
	}
	return strings.Join(ids, ",")
}

// CRC: crc-Requirements.md | R437, R438, R439, R440, R441
func TestRequirementsReadsTheFixture(t *testing.T) {
	r, src := loadReqs(t)
	var titles []string
	for _, s := range r.Sections() {
		titles = append(titles, s.Title)
	}
	if got := strings.Join(titles, "|"); got != "Requirements|Feature: node protocol|Notes|Feature: fences|Feature: empty" {
		t.Errorf("sections = %s", got)
	}
	if ids := reqIDs(r); ids != "R1,R2,R3,R4,R5,R6,R7" {
		t.Errorf("ids = %s", ids)
	}
	np, notes := r.Section("Feature: node protocol")[0], r.Section("Notes")[0]
	if np.Source != "specs/node-protocol.md" || np.Level != 2 || np.Line() != 3 || notes.Parent != np || notes.Level != 3 {
		t.Errorf("sections: %+v / %+v", np, notes)
	}
	r1, r3, r4, r5 := r.Requirement("R1"), r.Requirement("R3"), r.Requirement("R4"), r.Requirement("R5")
	if !strings.HasSuffix(r1.Text, "ignores provenance.") || r1.Section != np || r1.Line() != 6 || r1.Number != 1 {
		t.Errorf("R1: %+v", r1)
	}
	if !r3.Retired || r3.RetiredBy != "T1" || r3.Replacement != "R7" || r3.Text != "A node carries a parent pointer." {
		t.Errorf("R3: %+v", r3)
	}
	if !r4.Retired || r4.RetiredBy != "T2" || r4.Replacement != "" || r4.Text != "Nodes are registered with the document." {
		t.Errorf("R4: %+v", r4)
	}
	if r5.Section != notes {
		t.Error("R5 does not belong to the sub-heading")
	}
	if r.Requirement("R99") != nil || r.Section("Feature: quoted") != nil {
		t.Error("the fenced example was read")
	}
	if len(r.Unread()) != 0 {
		t.Errorf("unread = %v", r.Unread())
	}
	if out, _ := r.Render(); out != src {
		t.Error("render is not byte-exact")
	}
}

// CRC: crc-Requirements.md | R438, R439, R441
func TestRequirementsDeviations(t *testing.T) {
	r := ParseRequirements("## A\n**Source:** x\n**Source:** y\n\n- **R1:** one\n- **~~R2:~~** struck with no clause\n- **R1:** again\n- a bare bullet\n")
	if s := r.Section("A")[0]; s.Source != "x" {
		t.Errorf("source = %q", s.Source)
	}
	if len(r.Requirement("R2").Deviations()) != 1 || r.Requirement("R1").Text != "one" || len(r.Requirements()[2].Deviations()) != 1 {
		t.Errorf("deviations: %+v", r.Requirements())
	}
	if u := r.Unread(); len(u) != 4 || u[0].Text != "**Source:** y" || !strings.HasSuffix(u[3].Text, "a bare bullet") {
		t.Errorf("unread = %v", u)
	}
	var de *DeviationError
	if err := r.Retire("R2", "T3", "no replacement"); !errors.As(err, &de) {
		t.Errorf("write over a deviant entry: %v", err)
	}
}

// CRC: crc-Requirements.md | R442, R444
func TestRequirementsAdd(t *testing.T) {
	r, src := loadReqs(t)
	if err := r.Add("Feature: node protocol", "R8", "a new one, before the notes"); err != nil {
		t.Fatal(err)
	}
	out, _ := r.Render()
	if !strings.Contains(out, "registered with the document.\n- **R8:** a new one, before the notes\n\n### Notes") {
		t.Errorf("placement:\n%s", out)
	}
	if len(out) != len(src)+len("- **R8:** a new one, before the notes\n") {
		t.Error("changed more than the added line")
	}
	if err := r.Add("Feature: empty", "R9", "first of its section"); err != nil {
		t.Fatal(err)
	}
	out, _ = r.Render()
	if !strings.HasSuffix(out, "## Feature: empty\n**Source:** specs/empty.md\n- **R9:** first of its section\n") {
		t.Errorf("empty section:\n%s", out)
	}
	if err := r.Add("Feature: fences", "R10", "after the fence and the entries"); err != nil {
		t.Fatal(err)
	}
	out, _ = r.Render()
	if !strings.Contains(out, "one line.\n- **R10:** after the fence and the entries\n\n## Feature: empty") {
		t.Errorf("fence section:\n%s", out)
	}
	if err := r.Add("Feature: fences", "R6", "x"); !errors.Is(err, ErrReqExists) {
		t.Errorf("exists: %v", err)
	}
	if err := r.Add("Feature: fences", "Q6", "x"); !errors.Is(err, ErrBadReqID) {
		t.Errorf("bad id: %v", err)
	}
	if err := r.Add("Feature: none", "R11", "x"); !errors.Is(err, ErrNoSection) {
		t.Errorf("no section: %v", err)
	}
	two := ParseRequirements("## A\n- **R1:** a\n## A\n- **R2:** b\n")
	if err := two.Add("A", "R3", "x"); !errors.Is(err, ErrManySections) {
		t.Errorf("many: %v", err)
	}
}

// CRC: crc-Requirements.md | R443, R444
func TestRequirementsRetire(t *testing.T) {
	r, src := loadReqs(t)
	if err := r.Retire("R1", "T3", "see R7"); err != nil {
		t.Fatal(err)
	}
	out, _ := r.Render()
	if !strings.Contains(out, "- **~~R1:~~** (Retired T3 — see R7) A `Node` renders its bytes and reports its location; equality is per kind and\n  ignores provenance.\n") {
		t.Errorf("retire:\n%s", out)
	}
	if len(out) != len(src)+len("~~~~(Retired T3 — see R7) ") {
		t.Error("changed more than the head")
	}
	if q := r.Requirement("R1"); !q.Retired || q.Replacement != "R7" || !strings.HasSuffix(q.Text, "ignores provenance.") {
		t.Errorf("read back: %+v", q)
	}
	if err := r.Retire("R1", "T4", "no replacement"); !errors.Is(err, ErrRetired) {
		t.Errorf("twice: %v", err)
	}
	if err := r.Retire("R2", "T4", "see nothing"); !errors.Is(err, ErrBadClause) {
		t.Errorf("clause: %v", err)
	}
	if err := r.Retire("R2", "X4", "no replacement"); !errors.Is(err, ErrBadClause) {
		t.Errorf("tn: %v", err)
	}
	if err := r.Retire("R42", "T4", "no replacement"); !errors.Is(err, ErrNoRequirement) {
		t.Errorf("absent: %v", err)
	}
	if err := r.Retire("R2", "T4", "no replacement"); err != nil {
		t.Fatal(err)
	}
	if q := r.Requirement("R2"); q.Text != "(inferred) A compound's children tile its span." || q.Replacement != "" {
		t.Errorf("no replacement: %+v", q)
	}
}
