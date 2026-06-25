// CRC: crc-Parser.md | Seq: seq-parse.md | R67, R71, R104, R105, R106
package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// reqRefTokenRe matches a single ref `Rn` or an inclusive range `Rn-Rm`
// (the second `R` is optional: `R5-8` and `R5-R8` both parse). R105.
var reqRefTokenRe = regexp.MustCompile(`R(\d+)(?:\s*-\s*R?(\d+))?`)

// expandReqRefs extracts every Rn ref from s, expanding `Rn-Rm` ranges into
// each member so a range-form annotation covers the whole span (R105). A
// reversed range (`R8-R5`) contributes just the low ref.
func expandReqRefs(s string) []string {
	var refs []string
	for _, m := range reqRefTokenRe.FindAllStringSubmatch(s, -1) {
		if m[2] == "" {
			refs = append(refs, "R"+m[1])
			continue
		}
		lo, _ := strconv.Atoi(m[1])
		hi, _ := strconv.Atoi(m[2])
		if hi < lo {
			refs = append(refs, "R"+m[1])
			continue
		}
		for n := lo; n <= hi; n++ {
			refs = append(refs, "R"+strconv.Itoa(n))
		}
	}
	return refs
}

// ParseTraceability scans a code file for traceability comments.
// The commentPattern is a regex for the comment prefix (e.g., `//\s*` for Go).
// If commentPattern is empty, a default pattern matching // or -- is used.
// The commentCloser is the closing delimiter for block-comment languages (e.g., "}" for Pascal).
// If empty, only the built-in closers (-->, */) are stripped.
func ParseTraceability(path string, commentPattern string, commentCloser string) (Traceability, error) {
	file, err := os.Open(path)
	if err != nil {
		return Traceability{}, err
	}
	defer file.Close()

	if commentPattern == "" {
		commentPattern = `(?://|--|#)\s*`
	}
	// R106: wrap the comment prefix in a non-capturing group so an alternation
	// pattern (e.g. `<!--\s*|//\s*` for HTML with embedded JS) composes
	// correctly — otherwise the `|` binds loosely and the first alternative
	// matches without requiring `CRC:`.
	pattern := fmt.Sprintf(`(?:%s)CRC:\s*([^\|]+)(?:\|\s*Seq:\s*([^\|]+))?(.*)`, commentPattern)
	traceRe, err := regexp.Compile(pattern)
	if err != nil {
		return Traceability{}, fmt.Errorf("invalid comment pattern %q: %w", commentPattern, err)
	}
	// R104: a bare annotation leads with the ref(s) immediately after the
	// comment leader (e.g. `// R5: desc`, `// R5, R6`, trailing `foo() // R7`).
	// Only the leading comma-separated refs match; prose like `// see R5` does
	// not, because the ref does not follow the leader.
	bareRe, err := regexp.Compile(fmt.Sprintf(`(?:%s)(R\d+(?:\s*-\s*R?\d+)?(?:\s*,\s*R\d+(?:\s*-\s*R?\d+)?)*)`, commentPattern))
	if err != nil {
		return Traceability{}, fmt.Errorf("invalid comment pattern %q: %w", commentPattern, err)
	}

	trace := Traceability{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := traceRe.FindStringSubmatch(line); matches != nil {
			trace.CRCRefs = append(trace.CRCRefs, splitRefs(matches[1], commentCloser)...)
			if matches[2] != "" {
				trace.SeqRefs = append(trace.SeqRefs, splitRefs(matches[2], commentCloser)...)
			}
			if matches[3] != "" {
				trace.ReqRefs = append(trace.ReqRefs, extractReqRefs(matches[3], commentCloser)...)
			}
		} else if m := bareRe.FindStringSubmatch(line); m != nil {
			// R104, R105: bare annotation — collect only the leading refs,
			// expanding any Rn-Rm ranges.
			trace.ReqRefs = append(trace.ReqRefs, expandReqRefs(m[1])...)
		}
	}

	return trace, scanner.Err()
}

func extractReqRefs(s string, commentCloser string) []string {
	if commentCloser != "" {
		s = strings.TrimSuffix(s, strings.TrimSpace(commentCloser))
	}
	return expandReqRefs(s)
}

// splitRefs splits a comma-separated ref string into trimmed, non-empty parts.
// It strips the comment closer (from config) if provided.
func splitRefs(s string, commentCloser string) []string {
	var refs []string
	for _, ref := range strings.Split(s, ",") {
		ref = strings.TrimSpace(ref)
		if commentCloser != "" {
			ref = strings.TrimSuffix(ref, strings.TrimSpace(commentCloser))
		}
		ref = strings.TrimSpace(ref)
		if ref != "" {
			refs = append(refs, ref)
		}
	}
	return refs
}
