// CRC: crc-Update.md | Seq: seq-update.md | R313, R314, R315, R316
package update

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/zot/minispec/internal/parser"
	"github.com/zot/simple-dom/minispecsdom"
)

// Ranger resolves an injection site to a line range, in history and on disk. It is what
// SetInject decides `void` with, and an interface so the decision is testable against a
// stated world rather than a repository. R315
type Ranger interface {
	SiteRange(file, symbol string) (start, end int, err error)
	DiskRange(file, symbol string) (start, end int, err error)
}

// CRC: crc-Update.md | R310
// resolveAlarm turns a `<doc>#<n>` key into the document's path and the number, refusing
// anything else by name rather than by returning nothing. The reader answers whether the
// number names an alarm.
func (u *Update) resolveAlarm(key string) (path string, n int, err error) {
	name, num, ok := strings.Cut(key, "#")
	if !ok || name == "" {
		return "", 0, fmt.Errorf("an alarm is named <document>#<n>, not %q", key)
	}
	n, cerr := strconv.Atoi(num)
	if cerr != nil || n <= 0 {
		return "", 0, fmt.Errorf("an alarm's number is a positive integer, not %q", num)
	}
	path = filepath.Join(u.Project.DesignDir, name)
	if _, serr := os.Stat(path); serr != nil {
		return "", 0, fmt.Errorf("no document %s in the design directory", name)
	}
	return path, n, nil
}

// editTestDoc reads the document through the dependency, applies one write, and renders it
// back through the atomic file write. A refusal from the reader — no such alarm, a
// deviation on the entry — reaches the caller with no byte written. R316
func editTestDoc(path string, write func(td *minispecsdom.TestDoc) error) error {
	return parser.EditFile(path, func(src string) (string, error) {
		td := minispecsdom.ParseTestDoc(src)
		if err := write(td); err != nil {
			return "", err
		}
		return td.Render()
	})
}

// CRC: crc-Update.md | Seq: seq-update.md | R314
// SetPulled records a fire alarm as pulled: the date from `now` — the **system** clock, since
// a session that took its date from its opening greeting once wrote two days of stamps and
// staled three alarms on the next commit — then the body, read by the caller from a file
// byte for byte. The reader folds the previous line after it as history: the leading date
// is what the census reads, so a re-pull moves it, and a pull's history is the only record
// of what an injection used to do.
func (u *Update) SetPulled(key, body string, now time.Time) error {
	path, n, err := u.resolveAlarm(key)
	if err != nil {
		return err
	}
	return editTestDoc(path, func(td *minispecsdom.TestDoc) error {
		return td.SetPulled(n, now.Format("2006-01-02"), strings.TrimSpace(body))
	})
}

// CRC: crc-Update.md | Seq: seq-update.md | R315
// SetInject re-sites a fire alarm, and voids its record when the sites resolve to different
// code. Cleared reports whether a `**Pulled:**` line was demoted.
//
// **The comparison is of resolved extents, not of the field's text.** Text is right about a
// move and wrong about a rename, a re-formatting, and a disambiguation from `Parse` to
// `Head.Parse` — all read as a move to a string comparison, all would void a proof still
// good. Extents are self-sorting: an anchor rewritten to the declaration it already
// resolved to keeps its record, and one naming other code clears it. **Asymmetric on
// purpose:** the old sites resolve in HEAD, where a renamed symbol still exists, and the new
// ones on disk, where the rename is; old-sdom resolved both on disk and could not tell a
// rename from a move (its `O120`). A site resolving to nothing counts as different — the
// conservative direction, since understating a proof costs one re-pull and overstating it
// leaves a date vouching for a function nobody checked.
//
// Rewriting the sites to what they already name changes nothing and clears nothing, so the
// verb stays usable for tidying.
func (u *Update) SetInject(key string, sites []parser.AlarmSite, r Ranger) (cleared bool, err error) {
	path, n, err := u.resolveAlarm(key)
	if err != nil {
		return false, err
	}
	if len(sites) == 0 {
		return false, fmt.Errorf("refusing to write an empty **Inject:** — an alarm with no site is unanchored, which is a state to record rather than a value to write")
	}
	err = editTestDoc(path, func(td *minispecsdom.TestDoc) error {
		entry := td.Alarm(n)
		if entry == nil {
			return fmt.Errorf("no alarm numbered %d in %s", n, filepath.Base(path))
		}
		was := make([]parser.AlarmSite, 0, len(entry.Inject))
		for _, s := range entry.Inject {
			was = append(was, parser.AlarmSite{File: s.File, Symbol: s.Symbol})
		}
		if slices.Equal(was, sites) {
			return errUnchanged
		}
		void := entry.Pulled != nil && !sameCode(r, was, sites)
		cleared = void
		out := make([]minispecsdom.Site, len(sites))
		for i, s := range sites {
			out[i] = minispecsdom.Site{File: s.File, Symbol: s.Symbol}
		}
		return td.SetInject(n, out, void)
	})
	if err == errUnchanged {
		return false, nil
	}
	return cleared, err
}

var errUnchanged = errors.New("unchanged")

// R315
// sameCode reports whether the old sites, resolved in HEAD, and the new sites, resolved on
// disk, name the same line ranges.
func sameCode(r Ranger, was, now []parser.AlarmSite) bool {
	inHead, ok := extents(was, r.SiteRange)
	if !ok {
		return false
	}
	onDisk, ok := extents(now, r.DiskRange)
	if !ok {
		return false
	}
	return maps.Equal(inHead, onDisk)
}

// extents is the set of line ranges the sites resolve to through `resolve`, keyed by file
// and range. Not ok when any site resolves to nothing, which counts as different code.
func extents(sites []parser.AlarmSite, resolve func(file, symbol string) (int, int, error)) (map[string]bool, bool) {
	out := map[string]bool{}
	for _, s := range sites {
		start, end, err := resolve(s.File, s.Symbol)
		if err != nil {
			return nil, false
		}
		out[fmt.Sprintf("%s:%d-%d", s.File, start, end)] = true
	}
	return out, true
}

// NumberedDoc is what one document contributed to a numbering run.
type NumberedDoc struct {
	Path     string
	Assigned []int // the numbers written, in document order
}

// CRC: crc-Update.md | Seq: seq-update.md | R311, R313
// NumberAlarms gives every alarm lacking an `**Alarm:**` field the next free number in its
// document, writing the field, over every `design/test-*.md` when no path is named.
//
// **Append-only, idempotent, and it writes nothing but the added lines** — the reader's
// rules, restated because they are the requirement's: a number is an identifier and never
// a position, so an existing one is kept whatever order it sits in, a freed one is never
// handed out again, and a second run assigns nothing. A migration that also reflowed a
// paragraph would be one nobody can review.
func (u *Update) NumberAlarms(paths []string) ([]NumberedDoc, error) {
	if len(paths) == 0 {
		found, err := filepath.Glob(filepath.Join(u.Project.DesignDir, "test-*.md"))
		if err != nil {
			return nil, err
		}
		paths = found
	}
	slices.Sort(paths)
	var out []NumberedDoc
	for _, p := range paths {
		var assigned []int
		err := editTestDoc(p, func(td *minispecsdom.TestDoc) error {
			var werr error
			assigned, werr = td.NumberAlarms()
			return werr
		})
		if err != nil {
			return out, fmt.Errorf("%s: %w", p, err)
		}
		out = append(out, NumberedDoc{Path: p, Assigned: assigned})
	}
	return out, nil
}
