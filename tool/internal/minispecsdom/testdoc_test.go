package minispecsdom

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func loadTestDoc(t *testing.T) (*TestDoc, string) {
	t.Helper()
	src, err := os.ReadFile("testdata/testdoc-sample.md")
	if err != nil {
		t.Fatal(err)
	}
	return ParseTestDoc(string(src)), string(src)
}

// CRC: crc-TestDoc.md | R418, R419, R420, R421, R422, R424
func TestTestDocReadsTheFixture(t *testing.T) {
	d, src := loadTestDoc(t)
	tests := d.Tests()
	if len(tests) != 4 {
		t.Fatalf("entries = %d, want 4", len(tests))
	}
	e := tests[0]
	if e.Title != "a same-day change is verified, and the next day is stale" || e.Line() != 4 {
		t.Errorf("first entry: %q at %d", e.Title, e.Line())
	}
	if !strings.HasPrefix(e.FireAlarm, "make the comparison inclusive") || !strings.HasSuffix(e.FireAlarm, "which is the whole property") {
		t.Errorf("fire alarm did not fold: %q", e.FireAlarm)
	}
	if len(e.Inject) != 1 || e.Inject[0] != (Site{"internal/alarm/alarm.go", "assessOne"}) {
		t.Errorf("inject = %v", e.Inject)
	}
	if e.Pulled == nil || e.Pulled.Date != "2026-09-06" || !strings.HasPrefix(e.Pulled.Body, "rang again") {
		t.Errorf("pulled = %+v", e.Pulled)
	}
	if len(e.Code) != 1 || e.Code[0] != "internal/alarm/alarm_test.go" || e.Alarm != 3 {
		t.Errorf("code = %v alarm = %d", e.Code, e.Alarm)
	}
	if d.Alarm(3) != e || d.Alarm(0) != nil || d.Alarm(99) != nil {
		t.Error("Alarm(n) does not resolve as expected")
	}
	if tests[1].HasAlarm() || tests[1].Alarm != 0 || tests[1].Pulled != nil {
		t.Errorf("second entry is not an alarm: %+v", tests[1])
	}
	if !tests[2].HasAlarm() || tests[2].Pulled != nil || len(tests[2].Inject) != 2 || tests[2].Inject[1].Symbol != "routeRefusal" {
		t.Errorf("third entry: %+v", tests[2])
	}
	if tests[3].Alarm != 1 || d.Alarm(99) != nil {
		t.Errorf("the fenced fields were read as fields: alarm=%d", tests[3].Alarm)
	}
	if u := d.Unread(); len(u) != 1 || u[0].Text != "Notes" {
		t.Errorf("unread = %v, want the Notes heading alone", u)
	}
	if out, _ := d.Render(); out != src {
		t.Error("render is not byte-exact")
	}
}

// CRC: crc-TestDoc.md | R422, R423
func TestTestDocDeviations(t *testing.T) {
	var de *DeviationError
	d := ParseTestDoc("## Test: doubled\n**Fire alarm:** a\n**Alarm:** 1\n**Alarm:** 2\n**Pulled:** yesterday\n\n## Test: bad number\n**Fire alarm:** b\n**Alarm:** x\n")
	e := d.Tests()[0]
	if e.Alarm != 1 || len(e.Deviations()) != 2 {
		t.Fatalf("first read %d with %v", e.Alarm, e.Deviations())
	}
	if e.Pulled != nil {
		t.Error("a Pulled with no date read as a record")
	}
	if d.Tests()[1].Alarm != 0 || len(d.Tests()[1].Deviations()) != 1 {
		t.Errorf("bad number: %+v", d.Tests()[1])
	}
	orphan := ParseTestDoc("## Test: numbered, no alarm\n**Alarm:** 5\n**Refs:** x\n")
	if o := orphan.Tests()[0]; len(o.Deviations()) != 1 {
		t.Errorf("an Alarm with no Fire alarm is not a deviation: %+v", o)
	} else if err := orphan.SetPulled(5, "2026-09-06", "x"); !errors.As(err, &de) {
		t.Errorf("write over the orphan number: %v", err)
	}
	if len(d.Unread()) != 3 {
		t.Errorf("unread = %v", d.Unread())
	}
	if err := d.SetPulled(1, "2026-09-06", "x"); !errors.As(err, &de) {
		t.Errorf("write over deviations: %v", err)
	}
	if err := d.SetPulled(7, "2026-09-06", "x"); !errors.Is(err, ErrNoAlarm) {
		t.Errorf("absent number: %v", err)
	}
}

// CRC: crc-TestDoc.md | R425, R426
func TestTestDocSetPulled(t *testing.T) {
	d, src := loadTestDoc(t)
	if err := d.SetPulled(3, "2026-09-07", "rang: `boom`"); err != nil {
		t.Fatal(err)
	}
	out, _ := d.Render()
	want := "**Pulled:** 2026-09-07 — rang: `boom` *Earlier —* 2026-09-06 — rang again after `assessOne` gained the ambiguous-site case: `same-day = \"stale\", want verified`; restore byte-clean. First pulled 2026-08-13, same signature\n**Refs:**"
	if !strings.Contains(out, want) {
		t.Errorf("fold missing:\n%s", out)
	}
	if strings.Count(out, "**Pulled:**") != 1 {
		t.Error("the old line was not replaced")
	}
	if e := d.Alarm(3); e.Pulled.Date != "2026-09-07" {
		t.Errorf("read back %+v", e.Pulled)
	}
	// Outside the entry nothing moved: the other three entries render as before.
	for _, part := range []string{"## Test: pending entries", "## Test: the pre-`track`", "## Notes\n\nA level-2"} {
		if !strings.Contains(out, part) {
			t.Errorf("lost %q", part)
		}
	}
	if len(out) != len(src)+len("2026-09-07 — rang: `boom` *Earlier —* ") {
		t.Errorf("length moved by %d, want the prefix alone", len(out)-len(src))
	}

	// A first pull goes after Inject.
	d, _ = loadTestDoc(t)
	if _, err := d.NumberAlarms(); err != nil {
		t.Fatal(err)
	}
	e := d.Tests()[2]
	if err := d.SetPulled(e.Alarm, "2026-09-07", "rang"); err != nil {
		t.Fatal(err)
	}
	out, _ = d.Render()
	if !strings.Contains(out, "bootstrap.go:routeRefusal\n**Pulled:** 2026-09-07 — rang\n**Refs:** seq-bootstrap") {
		t.Errorf("first pull not after Inject:\n%s", out)
	}

	// With no Inject, after Fire alarm.
	d = ParseTestDoc("## Test: bare\n**Fire alarm:** break it\nand more\n**Alarm:** 1\n**Refs:** x\n")
	if err := d.SetPulled(1, "2026-09-07", "rang"); err != nil {
		t.Fatal(err)
	}
	if out, _ = d.Render(); out != "## Test: bare\n**Fire alarm:** break it\nand more\n**Pulled:** 2026-09-07 — rang\n**Alarm:** 1\n**Refs:** x\n" {
		t.Errorf("got %q", out)
	}
}

// CRC: crc-TestDoc.md | R427
func TestTestDocSetInject(t *testing.T) {
	d, _ := loadTestDoc(t)
	sites := []Site{{"a.go", "A"}, {"b.go", "B"}}
	if err := d.SetInject(3, sites, false); err != nil {
		t.Fatal(err)
	}
	out, _ := d.Render()
	if !strings.Contains(out, "**Inject:** a.go:A, b.go:B\n**Pulled:** 2026-09-06") {
		t.Errorf("rewrite without void:\n%s", out)
	}
	d, _ = loadTestDoc(t)
	if err := d.SetInject(3, sites, true); err != nil {
		t.Fatal(err)
	}
	out, _ = d.Render()
	want := "**Inject:** a.go:A, b.go:B\n*Pulled at `internal/alarm/alarm.go:assessOne` — 2026-09-06 — rang again after `assessOne` gained the ambiguous-site case: `same-day = \"stale\", want verified`; restore byte-clean. First pulled 2026-08-13, same signature — and the site has since moved, so this is history rather than a record.*\n**Refs:**"
	if !strings.Contains(out, want) {
		t.Errorf("demotion:\n%s", out)
	}
	if e := d.Alarm(3); e.Pulled != nil {
		t.Error("the demoted line still reads as a record")
	}
	if err := d.SetInject(3, nil, false); !errors.Is(err, ErrEmptyInject) {
		t.Errorf("empty: %v", err)
	}
	d = ParseTestDoc("## Test: x\n**Fire alarm:** a\n**Alarm:** 1\n")
	if err := d.SetInject(1, sites, false); !errors.Is(err, ErrNoInject) {
		t.Errorf("no inject: %v", err)
	}
}

// CRC: crc-TestDoc.md | R428
func TestTestDocNumberAlarms(t *testing.T) {
	d, src := loadTestDoc(t)
	got, err := d.NumberAlarms()
	if err != nil || len(got) != 1 || got[0] != 4 {
		t.Fatalf("assigned %v, %v; want [4]", got, err)
	}
	out, _ := d.Render()
	if !strings.Contains(out, "question\n**Alarm:** 4\n**Fire alarm:** swap the message") {
		t.Errorf("not above Fire alarm:\n%s", out)
	}
	if d.Alarm(4) == nil || d.Alarm(3).Alarm != 3 || d.Alarm(1).Alarm != 1 {
		t.Error("numbers moved")
	}
	if len(out) != len(src)+len("**Alarm:** 4\n") {
		t.Errorf("changed more than the added line: %d bytes", len(out)-len(src))
	}
	again, err := d.NumberAlarms()
	if err != nil || again != nil {
		t.Errorf("second run assigned %v", again)
	}
	if out2, _ := d.Render(); out2 != out {
		t.Error("second run changed bytes")
	}
	if d.Tests()[1].Alarm != 0 {
		t.Error("an entry with no Fire alarm was numbered")
	}
}

// CRC: crc-TestDoc.md | Seq: seq-testdoc.md#1.2 | R446
func TestATitleWithACodeSpanOrBoldReadsWhole(t *testing.T) {
	d, _ := loadTestDoc(t)
	want := map[string]bool{
		"the pre-`track` refusal asks the intent and stops": false,
		"pending entries read through **the dependency**":   false,
	}
	for _, e := range d.Tests() {
		if _, ok := want[e.Title]; ok {
			want[e.Title] = true
		}
	}
	for title, seen := range want {
		if !seen {
			t.Errorf("no entry titled %q — the title stopped at a marker", title)
		}
	}
}
