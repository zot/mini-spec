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

func TestParseTraceability_BareAnnotation(t *testing.T) {
	// R104: a `// Rn: desc` field/line annotation counts, no CRC ref.
	path := writeTemp(t, "// R1190: tag lines end with two trailing spaces\nvar x int\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R1190"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
	if len(trace.CRCRefs) != 0 {
		t.Errorf("CRCRefs = %v, want []", trace.CRCRefs)
	}
}

func TestParseTraceability_BareListTrailing(t *testing.T) {
	// R104: a trailing `code // Rn, Rn` annotation counts; leading refs only.
	path := writeTemp(t, "cache.Invalidate() // R1057, R1064\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R1057", "R1064"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}

func TestParseTraceability_BareNoColon(t *testing.T) {
	// R104: a bare ref with no description counts.
	path := writeTemp(t, "// R2693\nvar y int\n")
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R2693"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}

func TestParseTraceability_RangeExpansion(t *testing.T) {
	// R105: ranges expand in both the bare annotation and the CRC tail;
	// ranges and comma lists may be mixed.
	content := "// R5-R8: bare range\n" +
		"// CRC: crc-X.md | R10-R12\n" +
		"// R20-R22, R25: mixed\n" +
		"func f() {}\n"
	path := writeTemp(t, content)
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R5", "R6", "R7", "R8", "R10", "R11", "R12", "R20", "R21", "R22", "R25"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}

func TestParseTraceability_AlternationPattern(t *testing.T) {
	// R106: an alternation comment prefix (HTML + embedded JS) composes
	// correctly — the HTML CRC comment and the JS bare range both harvest.
	content := "<!-- CRC: crc-Foo.md | R500, R501 -->\n" +
		"// R600-R603: ranged bare annotation in <script>\n" +
		"foo();\n"
	path := writeTemp(t, content)
	trace, err := ParseTraceability(path, `<!--\s*|//\s*`, " -->")
	if err != nil {
		t.Fatal(err)
	}
	wantCRC := []string{"crc-Foo.md"}
	if !reflect.DeepEqual(trace.CRCRefs, wantCRC) {
		t.Errorf("CRCRefs = %v, want %v", trace.CRCRefs, wantCRC)
	}
	wantReq := []string{"R500", "R501", "R600", "R601", "R602", "R603"}
	if !reflect.DeepEqual(trace.ReqRefs, wantReq) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, wantReq)
	}
}

func TestParseTraceability_ProseNotCounted(t *testing.T) {
	// R104: refs not leading the comment are prose, not annotations.
	content := "// see R5 for the rationale\n// computed lazily (R6)\n// the R7 path is taken here\nfunc f() {}\n"
	path := writeTemp(t, content)
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.ReqRefs) != 0 {
		t.Errorf("ReqRefs = %v, want [] (prose mentions must not count)", trace.ReqRefs)
	}
}

func TestParseTraceability_BareAndCRCMixed(t *testing.T) {
	// R104: bare annotations and CRC lines coexist in one file.
	content := "// CRC: crc-Config.md | R624, R625\ntype Config struct {\n\tTTL string // R646: duration\n}\n"
	path := writeTemp(t, content)
	trace, err := ParseTraceability(path, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"R624", "R625", "R646"}
	if !reflect.DeepEqual(trace.ReqRefs, want) {
		t.Errorf("ReqRefs = %v, want %v", trace.ReqRefs, want)
	}
}
