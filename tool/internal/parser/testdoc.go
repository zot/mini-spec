// CRC: crc-Parser.md | Seq: seq-alarm-freshness.md#1.1 | R178, R316
package parser

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zot/minispec/internal/minispecsdom"
)

// CRC: crc-Parser.md | R178
// AlarmSite is one place a fault injection edits: a file and a symbol declared in it.
//
// The symbol names what the injection *changes*, never something it merely consults.
// That distinction is the whole reason the field is parsed rather than read out of the
// prose — "return early when `Lookup` already knows the path" edits the caller, and an
// anchor recorded on `Lookup` would point every future check at the wrong function.
type AlarmSite struct {
	File   string
	Symbol string
}

func (s AlarmSite) String() string { return s.File + ":" + s.Symbol }

// CRC: crc-Parser.md | R178, R310
// Alarm is one recorded fault injection from a test design.
//
// Pulled is zero when the document records no verification. That absence is data, not
// a missing value to be defaulted: an alarm with no pull date is a *prescription* —
// an injection someone wrote down and may never have run — and the two read identically
// in prose. ID is zero when the entry carries no `**Alarm:**` field: unmigrated, and
// reported as such rather than numbered by position (R312).
type Alarm struct {
	Doc    string // the test document's base name
	Test   string // the `## Test:` heading it sits under
	Prose  string // the `**Fire alarm:**` body
	Sites  []AlarmSite
	Pulled time.Time
	Code   []string // the `**Code:**` files, as written
	ID     int      // the `**Alarm:**` number, 0 when unnumbered
	Line   int      // the entry's heading line, 1-based
}

// HasPulled reports whether a verification date was recorded.
func (a Alarm) HasPulled() bool { return !a.Pulled.IsZero() }

// Numbered reports whether the entry carries an `**Alarm:**` field. R312
func (a Alarm) Numbered() bool { return a.ID != 0 }

// Key is the alarm's name, `<doc>#<n>`, or "" while it is unnumbered. R310
func (a Alarm) Key() string {
	if !a.Numbered() {
		return ""
	}
	return a.Doc + "#" + strconv.Itoa(a.ID)
}

// CRC: crc-Parser.md | Seq: seq-alarm-freshness.md#1.1 | R178, R316
// ParseTestDoc reads the fire alarms a test design records, through the dependency's
// test-document reader: an entry is a `## Test:` heading's region, a field is
// `**Name:**` at a line head outside any code group, the two prose fields fold across
// wrapped lines, and a fenced example is body. What the reader could not read is dropped
// here; ParseTestDocReport carries it.
func ParseTestDoc(path string) ([]Alarm, error) {
	alarms, _, err := ParseTestDocReport(path)
	return alarms, err
}

// CRC: crc-Parser.md | R178, R316
// ParseTestDocReport is ParseTestDoc with the reader's unread list beside the alarms.
func ParseTestDocReport(path string) ([]Alarm, []minispecsdom.Unread, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	td := minispecsdom.ParseTestDoc(string(data))
	doc := filepath.Base(path)
	var out []Alarm
	for _, e := range td.Tests() {
		if !e.HasAlarm() {
			continue
		}
		out = append(out, alarmOf(doc, e))
	}
	return out, td.Unread(), nil
}

// R178
// alarmOf is the census's view of one entry. A site with no file or no symbol is
// dropped rather than guessed at — half an anchor points somewhere, and somewhere is
// worse than nowhere — and a `**Pulled:**` whose date does not parse records nothing:
// the reader carries what was written, the judgment is this tool's.
func alarmOf(doc string, e *minispecsdom.TestEntry) Alarm {
	a := Alarm{Doc: doc, Test: e.Title, Prose: e.FireAlarm, Code: e.Code, ID: e.Alarm, Line: e.Line()}
	for _, s := range e.Inject {
		file, symbol := strings.TrimSpace(s.File), strings.TrimSpace(s.Symbol)
		if file == "" || symbol == "" {
			continue
		}
		a.Sites = append(a.Sites, AlarmSite{File: file, Symbol: symbol})
	}
	if e.Pulled != nil {
		if when, perr := time.Parse("2006-01-02", e.Pulled.Date); perr == nil {
			a.Pulled = when
		}
	}
	return a
}
