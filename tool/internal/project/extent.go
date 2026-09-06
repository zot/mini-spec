// CRC: crc-Git.md | Seq: seq-alarm-freshness.md#1.5.1, seq-alarm-freshness.md#1.5.2, seq-alarm-freshness.md#1.5.3 | R303, R304, R305, R306
package project

import (
	"strings"

	"github.com/zot/simple-dom/sdom"
	"github.com/zot/simple-dom/sdom/schema"
)

// siteExtent resolves an `**Inject:**` symbol in Go source to its 1-based line range, and
// says how many declarations answered to the name.
//
// **The reader computes the range; git is only asked when those lines changed.** Handing
// git a pattern asks it to *find* the symbol as well as bound it, and git bounds a
// declaration at the line before the next one — so its range carried the successor's doc
// comment and the trailing blank line, and the first bounded match could be a use or a
// comment rather than the declaration (gaps O10, O11, measured 2026-09-04).
//
// **The declaring line plus the groups that follow it.** A declaration is not a node in
// the dependency's DOM: the keyword and the name are narrow typed siblings with the
// receiver and parameter groups between them. So the extent starts at the keyword's line
// and runs through the line on which every bracket group opened inside it has closed —
// the group rule, under which `func f() {` is one start though its parens close mid-line.
//
// **No comment is in the range, and that is measured rather than reasoned.** The first
// repair drafted on old-sdom attached the declaration's own doc comment; run against the
// live corpus it turned three verified alarms stale, every one a traceability line
// rewritten inside the doc block. A comment is not the code an injection proves.
//
// `Type.Method` is matched on the receiver group's type, so three `Parse` methods on
// three types are three symbols; a bare name that two declarations answer to is counted,
// not guessed at, because the anchor would have been watching an arbitrary one of them.
func siteExtent(src, symbol string) (start, end, count int) {
	bp := sdom.NewBracketParser(&sdom.LangGo)
	d := sdom.Parse(src, 0, bp)
	ctx := bp.Context()
	if err := schema.Go(d, ctx); err != nil {
		return 0, 0, 0
	}
	typ, method, isMethod := strings.Cut(symbol, ".")
	want := symbol
	if isMethod {
		want = method
	}
	nodes := d.Nodes()
	for i, n := range nodes {
		kw, ok := n.(*sdom.DeclarationType)
		if !ok {
			continue
		}
		names, err := ctx.DeclarationNames(kw)
		if err != nil {
			return 0, 0, 0
		}
		for _, name := range names {
			text, _ := name.Render()
			if text != want {
				continue
			}
			// R304 — the receiver group sits between the keyword and the name; its type
			// is the last word of its text, star stripped.
			if isMethod && receiverType(d, ctx, i, name) != typ {
				continue
			}
			count++
			if count > 1 {
				continue
			}
			// R305 — a grouped declaration's member starts at its own line; a plain one
			// at its keyword's.
			from := kw.Location().Offset()
			if len(names) > 1 {
				from = name.Location().Offset()
			}
			start = d.Line(from)
			end = groupEnd(d, ctx, d.IndexOf(name), d.Line(name.Location().Offset()))
		}
	}
	return start, end, count
}

// R304
// receiverType is the type named by the receiver group between the keyword at kw and the
// name, or "" when there is none: `(d *Doc)` and `(Doc)` both answer `Doc`, and a type
// parameter list is dropped.
func receiverType(d *sdom.Doc, ctx *sdom.BracketContext, kw int, name sdom.Node) string {
	nodes := d.Nodes()
	nameAt := d.IndexOf(name)
	for i := kw + 1; i < nameAt; i++ {
		op, ok := nodes[i].(*sdom.Opener)
		if !ok {
			continue
		}
		if text, _ := op.Render(); text != "(" {
			continue
		}
		inner := ctx.InnerText(op)
		if at := strings.IndexByte(inner, '['); at >= 0 {
			inner = inner[:at]
		}
		fields := strings.Fields(inner)
		if len(fields) == 0 {
			return ""
		}
		return strings.TrimPrefix(fields[len(fields)-1], "*")
	}
	return ""
}

// Seq: seq-alarm-freshness.md#1.5.2 | R305
// groupEnd walks forward from the name and returns the last line any bracket group
// opened within the extent closes on. A node that begins past the extent ends the walk,
// which is what keeps the successor's doc comment out.
func groupEnd(d *sdom.Doc, ctx *sdom.BracketContext, from, end int) int {
	nodes := d.Nodes()
	for i := from + 1; i < len(nodes); i++ {
		n := nodes[i]
		if d.Line(n.Location().Offset()) > end {
			break
		}
		op, ok := n.(*sdom.Opener)
		if !ok {
			continue
		}
		if cl := ctx.Closer(op); cl != nil {
			end = max(end, d.Line(cl.Location().Offset()))
		}
	}
	return end
}
