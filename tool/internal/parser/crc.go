// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/zot/simple-dom/sdom"
)

var (
	crcNameRe     = regexp.MustCompile(`^# (.+)`)
	crcReqsRe     = regexp.MustCompile(`^\*\*Requirements:\*\*\s*(.*)`)
	crcSeqHdrRe   = regexp.MustCompile(`^## Sequences`)
	crcListItemRe = regexp.MustCompile(`^- (.+\.md)`)
)

// ParseCRCCard parses a CRC card file
// CRC: crc-Parser.md | R6, R532
func ParseCRCCard(path string) (CRCCard, error) {
	file, err := os.Open(path)
	if err != nil {
		return CRCCard{}, err
	}
	defer file.Close()

	card := CRCCard{Path: path}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	inSequences := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Extract name from first # heading
		if card.Name == "" {
			if matches := crcNameRe.FindStringSubmatch(line); matches != nil {
				card.Name = strings.TrimSpace(matches[1])
				continue
			}
		}

		// Extract requirements
		if matches := crcReqsRe.FindStringSubmatch(line); matches != nil {
			card.ReqLine = lineNum
			refs, others := RequirementsField(strings.TrimSpace(matches[1]))
			for _, n := range refs {
				card.Requirements = append(card.Requirements, "R"+strconv.Itoa(n))
			}
			card.Requirements = append(card.Requirements, others...)
			continue
		}

		// Detect Sequences section
		if crcSeqHdrRe.MatchString(line) {
			inSequences = true
			continue
		}

		// New section ends Sequences parsing
		if strings.HasPrefix(line, "## ") {
			inSequences = false
			continue
		}

		// Extract sequence refs when in Sequences section
		if inSequences {
			if matches := crcListItemRe.FindStringSubmatch(line); matches != nil {
				card.Sequences = append(card.Sequences, strings.TrimSpace(matches[1]))
			}
		}
	}

	return card, scanner.Err()
}

// CRC: crc-Parser.md | R532
// RequirementsField reads a card's Requirements field through the grammar traceability
// comments use, so a range names every member. The list is read from the head; every comma
// token after it that is itself a whole requirement list joins the refs too, so a ref written
// after a stray word is still a ref. What remains is kept, as written and in order, for validate
// to report as unknown references and for a rewrite to carry through unchanged.
func RequirementsField(field string) (refs []int, others []string) {
	rest := field
	if list, after, ok := sdom.ParseRequirementList(field, sdom.Loc{}); ok {
		refs = list.Items()
		rest = after
	}
	for _, tok := range strings.Split(rest, ",") {
		if tok = strings.TrimSpace(tok); tok == "" {
			continue
		}
		if list, after, ok := sdom.ParseRequirementList(tok, sdom.Loc{}); ok && strings.TrimSpace(after) == "" {
			refs = append(refs, list.Items()...)
			continue
		}
		others = append(others, tok)
	}
	return refs, others
}
