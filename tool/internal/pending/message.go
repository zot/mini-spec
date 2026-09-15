// CRC: crc-CommitMessage.md | Seq: seq-queue-item.md#4 | R479, R480, R481, R482, R483
package pending

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/zot/minispec/internal/minispecsdom"
	"github.com/zot/minispec/internal/project"
)

// Message is a composed commit message and the items it names.
type Message struct {
	Subject string `json:"subject"`
	Text    string `json:"text"` // the whole message, for `git commit -F`
	Items   []int  `json:"items"`
	Amend   bool   `json:"amend"`
}

// ErrNothingToCommit is every done entry already named by a commit.
var ErrNothingToCommit = errors.New("every done entry is named by a commit; nothing to compose")

// OnRemoteError is the amend refusal: HEAD is in a remote branch someone else may hold.
type OnRemoteError struct{ Branches []string }

func (e *OnRemoteError) Error() string {
	return fmt.Sprintf("HEAD is on %s; amending a shared commit rewrites history someone else holds — make a follow-up commit instead", strings.Join(e.Branches, ", "))
}

// item is one uncommitted done entry: its header and the file's lines beneath it.
type item struct {
	ids   []int
	title string
	body  string
}

// CRC: crc-CommitMessage.md | Seq: seq-queue-item.md#4 | R479, R480, R481, R482, R483
//
// Compose writes the message for the items no commit names yet. Git says which those are:
// the done file is private and never in a commit, so the tool asks which `#N` the messages
// on HEAD's history name and keeps every entry with a number none names. With amend, HEAD's
// message comes back unchanged with the new items after it — the previous message is part
// of the record — and a HEAD on a remote branch is refused.
func Compose(repoRoot string, amend bool) (Message, error) {
	if err := missingTrajectory(repoRoot); err != nil { // R332
		return Message{}, err
	}
	git := project.NewGit(repoRoot)
	if !git.IsRepo() {
		return Message{}, project.ErrNoGit
	}
	items, newest, err := uncommitted(repoRoot, git)
	if err != nil {
		return Message{}, err
	}
	if len(items) == 0 {
		return Message{}, fmt.Errorf("%w (the newest, %s, is named)", ErrNothingToCommit, newest)
	}
	var ids []int
	var titles []string
	var body strings.Builder
	for _, it := range items {
		ids = append(ids, it.ids...)
		titles = append(titles, it.title)
		fmt.Fprintf(&body, "\n%s — %s\n", idList(it.ids), it.title)
		if it.body != "" {
			body.WriteString(it.body + "\n")
		}
	}
	list := idList(ids)
	msg := Message{Items: ids, Amend: amend}
	if !amend {
		msg.Subject = fmt.Sprintf("%s: %s", list, strings.Join(titles, "; "))
		msg.Text = fmt.Sprintf("%s\n\nItems %s.\n%s", msg.Subject, list, body.String())
		return msg, nil
	}
	// R483 — refused before the message is read: a shared commit is not ours to rewrite.
	branches, err := git.HeadOnRemote()
	if err != nil {
		return Message{}, err
	}
	if len(branches) > 0 {
		return Message{}, &OnRemoteError{Branches: branches}
	}
	existing, err := git.HeadMessage()
	if err != nil {
		return Message{}, err
	}
	existing = strings.TrimRight(existing, "\n")
	msg.Subject, _, _ = strings.Cut(existing, "\n")
	// R482 — the existing message first, byte for byte; the new items after it.
	msg.Text = fmt.Sprintf("%s\n\nAlso lands %s.\n%s", existing, list, body.String())
	return msg, nil
}

// CRC: crc-CommitMessage.md | Seq: seq-queue-item.md#4.3 | R479
// uncommitted is every done entry with an identifier no commit names, oldest first, and the
// newest entry's identifiers for the refusal.
func uncommitted(repoRoot string, git *project.Git) ([]item, string, error) {
	src, err := os.ReadFile(filepath.Join(repoRoot, doneFile))
	if err != nil {
		return nil, "", err
	}
	named, err := git.NamedItems()
	if err != nil {
		return nil, "", err
	}
	text := string(src)
	lines := strings.Split(text, "\n")
	entries := minispecsdom.ParseDone(text).Entries()
	var out []item
	newest := ""
	for i, e := range entries { // most recent first
		if len(e.IDs) == 0 {
			continue
		}
		if newest == "" {
			newest = idList(e.IDs)
		}
		// The run ends at the first named entry — a hash in its slot, or every identifier in a
		// message: everything older is history, whatever its slot says. Measured on this
		// repository's first run: two August entries with neither, composed by a per-entry rule.
		if e.Commit != "" || allNamed(e.IDs, named) {
			break
		}
		end := len(lines)
		if i+1 < len(entries) {
			end = entries[i+1].Line() - 1
		}
		out = append(out, item{ids: e.IDs, title: strings.TrimSuffix(e.Title, "."), body: entryBody(lines[e.Line():end])})
	}
	slices.Reverse(out) // oldest first: the order the items finished
	return out, newest, nil
}

func allNamed(ids []int, named map[int]bool) bool {
	for _, id := range ids {
		if !named[id] {
			return false
		}
	}
	return true
}

// entryBody is the done file's lines beneath an entry's header, the two-space indent
// removed and blank edges trimmed; a horizontal rule ends it.
func entryBody(lines []string) string {
	var out []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "---" {
			break
		}
		out = append(out, strings.TrimPrefix(l, "  "))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func idList(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "#" + strconv.Itoa(id)
	}
	return strings.Join(parts, ", ")
}
