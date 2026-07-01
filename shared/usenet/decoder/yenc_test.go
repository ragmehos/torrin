package decoder

import (
	"bytes"
	"testing"
)

func TestDecode(t *testing.T) {
	// bytes 0..4 yEnc-encode (no escapes needed) to (b+42): '*' '+' ',' '-' '.'
	msg := "=ybegin part=1 total=1 line=128 size=5 name=test.bin\n*+,-.\n=yend size=5 part=1\n"
	r, err := Decode([]byte(msg))
	if err != nil {
		t.Fatal(err)
	}
	if r.Filename != "test.bin" {
		t.Errorf("filename = %q", r.Filename)
	}
	if !bytes.Equal(r.Data, []byte{0, 1, 2, 3, 4}) {
		t.Errorf("data = %v, want [0 1 2 3 4]", r.Data)
	}
}

func TestDecodeNoYenc(t *testing.T) {
	if _, err := Decode([]byte("not yenc data")); err == nil {
		t.Error("expected error for non-yenc input")
	}
}
