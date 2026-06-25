// CRC: crc-Parser.md | Seq: seq-parse.md | R67, R71, R104
package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var reqRefRe = regexp.MustCompile(`R\d+`)

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
	pattern := fmt.Sprintf(`%sCRC:\s*([^\|]+)(?:\|\s*Seq:\s*([^\|]+))?(.*)`, commentPattern)
	traceRe, err := regexp.Compile(pattern)
	if err != nil {
		return Traceability{}, fmt.Errorf("invalid comment pattern %q: %w", commentPattern, err)
	}
	// R104: a bare annotation leads with the ref(s) immediately after the
	// comment leader (e.g. `// R5: desc`, `// R5, R6`, trailing `foo() // R7`).
	// Only the leading comma-separated refs match; prose like `// see R5` does
	// not, because the ref does not follow the leader.
	bareRe, err := regexp.Compile(fmt.Sprintf(`%s(R\d+\b(?:\s*,\s*R\d+\b)*)`, commentPattern))
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
			// R104: bare annotation — collect only the leading refs.
			trace.ReqRefs = append(trace.ReqRefs, reqRefRe.FindAllString(m[1], -1)...)
		}
	}

	return trace, scanner.Err()
}

func extractReqRefs(s string, commentCloser string) []string {
	if commentCloser != "" {
		s = strings.TrimSuffix(s, strings.TrimSpace(commentCloser))
	}
	return reqRefRe.FindAllString(s, -1)
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
