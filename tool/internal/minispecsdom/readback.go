package minispecsdom

import "fmt"

// CRC: crc-PendingSdom.md | R413, R414, R415, R416
//
// ReadBackError is what a write panics with when the reader cannot read it back: the
// writer produced bytes its own reader does not see as the write that was asked for. It
// is a library invariant, not caller input, so it is a panic and not an error a caller
// could swallow; a tool recovers it into a refusal naming the file, and the file stays
// untouched because nothing is written until render returns. The message is a crank
// handle: which reader, which write, on which key, what was wanted and what came back.
// It proves the tree describes the bytes as the write intended; it does not prove the
// write addressed the right region, which no re-parse can.
type ReadBackError struct {
	Reader, Write, Key, Want, Got string
}

func (e *ReadBackError) Error() string {
	return fmt.Sprintf("minispecsdom: %s.%s on %q did not read back: want %s, got %s", e.Reader, e.Write, e.Key, e.Want, e.Got)
}

// CRC: crc-PendingSdom.md | R413, R414, R415, R416
// mustReadBack is the guard at the tail of every write path: ok is the reader's own
// verdict on the write it just made, after the re-read.
func mustReadBack(reader, write, key string, ok bool, want, got string) {
	if !ok {
		panic(&ReadBackError{Reader: reader, Write: write, Key: key, Want: want, Got: got})
	}
}
