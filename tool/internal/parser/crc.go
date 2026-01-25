// CRC: crc-Parser.md | Seq: seq-parse.md
package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	crcNameRe     = regexp.MustCompile(`^# (.+)`)
	crcReqsRe     = regexp.MustCompile(`^\*\*Requirements:\*\*\s*(.*)`)
	crcSeqHdrRe   = regexp.MustCompile(`^## Sequences`)
	crcListItemRe = regexp.MustCompile(`^- (.+\.md)`)
)

// ParseCRCCard parses a CRC card file
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
			reqStr := strings.TrimSpace(matches[1])
			if reqStr != "" {
				parts := strings.Split(reqStr, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						card.Requirements = append(card.Requirements, p)
					}
				}
			}
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
