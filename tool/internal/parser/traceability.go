// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

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

	// Build the traceability regex from the comment pattern
	if commentPattern == "" {
		commentPattern = `(?://|--|#)\s*`
	}
	pattern := fmt.Sprintf(`%sCRC:\s*([^\|]+)(?:\|\s*Seq:\s*(.+))?`, commentPattern)
	traceRe, err := regexp.Compile(pattern)
	if err != nil {
		return Traceability{}, fmt.Errorf("invalid comment pattern %q: %w", commentPattern, err)
	}

	trace := Traceability{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if matches := traceRe.FindStringSubmatch(line); matches != nil {
			trace.CRCRefs = append(trace.CRCRefs, splitRefs(matches[1], commentCloser)...)
			if len(matches) > 2 && matches[2] != "" {
				trace.SeqRefs = append(trace.SeqRefs, splitRefs(matches[2], commentCloser)...)
			}
		}
	}

	return trace, scanner.Err()
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
