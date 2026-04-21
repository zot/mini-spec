// CRC: crc-Parser.md | Seq: seq-parse.md | R67
package parser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.go")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseTraceability_Basic(t *testing.T) {
	path := writeTemp(t, "// CRC: crc-DB.md | Seq: seq-crud.md\nfunc main() {}\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.CRCRefs) != 1 || trace.CRCRefs[0] != "crc-DB.md" {
		t.Errorf("CRCRefs = %v, want [crc-DB.md]", trace.CRCRefs)
	}
	if len(trace.SeqRefs) != 1 || trace.SeqRefs[0] != "seq-crud.md" {
		t.Errorf("SeqRefs = %v, want [seq-crud.md]", trace.SeqRefs)
	}
	if len(trace.ReqRefs) != 0 {
		t.Errorf("ReqRefs = %v, want []", trace.ReqRefs)
	}
}

func TestParseTraceability_WithReqRefs(t *testing.T) {
	path := writeTemp(t, "// CRC: crc-DB.md | Seq: seq-write-actor.md | R1054, R1055, R1056\nfunc Copy() {}\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.CRCRefs) != 1 || trace.CRCRefs[0] != "crc-DB.md" {
		t.Errorf("CRCRefs = %v, want [crc-DB.md]", trace.CRCRefs)
	}
	if len(trace.SeqRefs) != 1 || trace.SeqRefs[0] != "seq-write-actor.md" {
		t.Errorf("SeqRefs = %v, want [seq-write-actor.md]", trace.SeqRefs)
	}
	want := []string{"R1054", "R1055", "R1056"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}

func TestParseTraceability_CRCOnlyWithReqRefs(t *testing.T) {
	path := writeTemp(t, "// CRC: crc-DB.md | R1053\nfunc main() {}\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.CRCRefs) != 1 || trace.CRCRefs[0] != "crc-DB.md" {
		t.Errorf("CRCRefs = %v, want [crc-DB.md]", trace.CRCRefs)
	}
	if len(trace.SeqRefs) != 0 {
		t.Errorf("SeqRefs = %v, want []", trace.SeqRefs)
	}
	want := []string{"R1053"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}

func TestParseTraceability_MultipleLines(t *testing.T) {
	content := "// CRC: crc-DB.md | R5, R6\n// CRC: crc-DB.md | Seq: seq-crud.md | R7\nfunc main() {}\n"
	path := writeTemp(t, content)
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R5", "R6", "R7"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}
