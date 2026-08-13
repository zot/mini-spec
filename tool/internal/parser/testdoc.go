// CRC: crc-Parser.md | Seq: seq-alarm-freshness.md#1.1 | R178
package parser

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

// CRC: crc-Parser.md | R178
// Alarm is one recorded fault injection from a test design.
//
// Pulled is zero when the document records no verification. That absence is data, not
// a missing value to be defaulted: an alarm with no pull date is a *prescription* —
// an injection someone wrote down and may never have run — and the two read identically
// in prose.
type Alarm struct {
	Doc    string // the test document's base name
	Test   string // the `## Test:` heading it sits under
	Prose  string // the `**Fire alarm:**` body
	Sites  []AlarmSite
	Pulled time.Time
}

// HasPulled reports whether a verification date was recorded.
func (a Alarm) HasPulled() bool { return !a.Pulled.IsZero() }

var (
	alarmRe  = regexp.MustCompile(`^\*\*Fire alarm:\*\*\s*(.*)$`)
	injectRe = regexp.MustCompile(`^\*\*Inject:\*\*\s*(.*)$`)
	pulledRe = regexp.MustCompile(`^\*\*Pulled:\*\*\s*(\d{4}-\d{2}-\d{2})`)
	fieldRe  = regexp.MustCompile(`^(\*\*[A-Z]|##)`)
)

// CRC: crc-Parser.md | Seq: seq-alarm-freshness.md#1.1 | R178
// ParseTestDoc reads the fire alarms a test design records.
//
// Fields are matched only at the **start of a line**, which is what keeps a mention
// inside prose from being read as a field. Continuation lines are folded into the alarm
// body until the next field or heading, because these documents wrap — and a reader
// that assumed one line per field would silently see half of every alarm. That failure
// is not hypothetical: three separate greps during this feature's own measurement
// reported absence because a phrase broke across lines.
func ParseTestDoc(path string) ([]Alarm, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	doc := filepath.Base(path)
	lines := strings.Split(string(data), "\n")

	var out []Alarm
	test := "(file-level)"
	for i, line := range lines {
		if strings.HasPrefix(line, "## Test:") {
			test = strings.TrimSpace(strings.TrimPrefix(line, "## Test:"))
			continue
		}
		m := alarmRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		// Fold the wrapped remainder of the alarm, then read the fields that follow it.
		prose, fieldsAt := foldProse(lines, i, m[1])
		sites, pulled := readAlarmFields(lines, fieldsAt)
		out = append(out, Alarm{Doc: doc, Test: test, Prose: prose, Sites: sites, Pulled: pulled})
	}
	return out, nil
}

// foldProse joins a `**Fire alarm:**` value with the wrapped lines that continue it, and
// reports where the fold stopped — the index at which that alarm's fields begin.
//
// It ends at a blank line, the next field, or the next heading, and never consumes the
// line that stopped it. That is also why the caller can keep walking from where it left
// off instead of skipping ahead: a line folded here matched neither `## Test:` nor
// `**Fire alarm:**` by construction, since both of those start a field or heading and so
// would have ended the fold. Advancing the caller's cursor past the body would be an
// optimisation paid for in index arithmetic, which is the reading this parser can least
// afford to get subtly wrong.
func foldProse(lines []string, start int, first string) (string, int) {
	body := []string{strings.TrimSpace(first)}
	i := start + 1
	for ; i < len(lines) && strings.TrimSpace(lines[i]) != "" && !fieldRe.MatchString(lines[i]); i++ {
		body = append(body, strings.TrimSpace(lines[i]))
	}
	return strings.Join(body, " "), i
}

// readAlarmFields collects the Inject sites and Pulled date belonging to the alarm that
// ends at `from`, stopping at the next test heading so one alarm never adopts the next
// one's fields.
func readAlarmFields(lines []string, from int) ([]AlarmSite, time.Time) {
	var sites []AlarmSite
	var pulled time.Time
	for k := from; k < len(lines); k++ {
		line := lines[k]
		if strings.HasPrefix(line, "## ") || alarmRe.MatchString(line) {
			break
		}
		if m := injectRe.FindStringSubmatch(line); m != nil {
			sites = append(sites, parseSites(m[1])...)
			continue
		}
		if m := pulledRe.FindStringSubmatch(line); m != nil {
			if when, err := time.Parse("2006-01-02", m[1]); err == nil {
				pulled = when
			}
		}
	}
	return sites, pulled
}

// parseSites splits a comma-separated `**Inject:**` value into file/symbol pairs. An
// entry with no colon is dropped rather than guessed at: half an anchor points
// somewhere, and somewhere is worse than nowhere.
func parseSites(value string) []AlarmSite {
	var out []AlarmSite
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(strings.Trim(strings.TrimSpace(part), "`"))
		if part == "" {
			continue
		}
		file, symbol, ok := strings.Cut(part, ":")
		if !ok || strings.TrimSpace(file) == "" || strings.TrimSpace(symbol) == "" {
			continue
		}
		out = append(out, AlarmSite{File: strings.TrimSpace(file), Symbol: strings.TrimSpace(symbol)})
	}
	return out
}
