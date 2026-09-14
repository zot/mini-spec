package minispecsdom

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/zot/simple-dom/sdom"
)

// CRC: crc-Done.md | R402, R403, R404, R405, R411, R445
//
// unbalanced lists what the base's context reports as paired with nothing, each as an
// Unread naming the marker at its line: every opener never closed — at end of input, or
// because a rejected run ended its group — and every closer that closes nothing — a stray
// one, or that rejected run.
//
// One helper for four readers: a fence or span that runs to end of file takes every later
// entry with it and leaves nothing entry-like to list, so each reader must say this itself
// rather than wait to notice.
func unbalanced(ctx *sdom.BracketContext) []Unread {
	doc := ctx.Doc()
	var out []Unread
	for _, o := range ctx.Unclosed() {
		marker, _ := o.Render()
		out = append(out, Unread{doc.Line(o.Location().Offset()), fmt.Sprintf("`%s` never closed", marker)})
	}
	for _, c := range ctx.Unpaired() {
		marker, _ := c.Render()
		out = append(out, Unread{doc.Line(c.Location().Offset()), fmt.Sprintf("`%s` closes nothing", marker)})
	}
	for _, m := range ctx.Demoted() { // R445
		out = append(out, Unread{doc.Line(m.Run.Location().Offset() + m.Offset), fmt.Sprintf("`%s` never closed, read as text", m.Marker)})
	}
	return out
}

// CRC: crc-Done.md | R402, R403
// byLine orders an unread list by line, stably: a closer that closes nothing can sit
// anywhere in a file, so the kinds interleave.
func byLine(u []Unread) {
	slices.SortStableFunc(u, func(a, b Unread) int { return cmp.Compare(a.Line, b.Line) })
}
