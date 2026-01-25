// CRC: crc-Parser.md
package parser

// Requirement represents a single requirement from requirements.md
type Requirement struct {
	ID       string // e.g., "R1"
	Text     string
	Source   string // spec file path
	Inferred bool
	Line     int
}

// CRCCard represents a parsed CRC card
type CRCCard struct {
	Name         string
	Requirements []string // e.g., ["R1", "R3", "R7"]
	Sequences    []string // e.g., ["seq-login.md", "seq-auth.md"]
	Path         string
	ReqLine      int // line number of Requirements field
}

// Artifact represents a design file with its associated code files
type Artifact struct {
	DesignFile string
	CodeFiles  []CodeFile
}

// CodeFile represents a code file listed in Artifacts
type CodeFile struct {
	Path    string
	Checked bool
	Line    int
}

// Gap represents an item in the Gaps section
type Gap struct {
	ID          string // e.g., "D1", "R2", "S1"
	Type        string // S, R, D, C, O
	Description string
	Resolved    bool
	Line        int
}

// Traceability represents traceability comments found in a code file
type Traceability struct {
	CRCRefs []string
	SeqRefs []string
}
