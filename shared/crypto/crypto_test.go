package crypto

import "testing"

const testKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func TestRoundTrip(t *testing.T) {
	c, err := New(testKey)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{"rd-api-key-xyz", "", "TRN-ABCD-EFGH"} {
		ct := c.Encrypt(p)
		if p != "" && ct == p {
			t.Errorf("%q was not encrypted", p)
		}
		if got := c.Decrypt(ct); got != p {
			t.Errorf("round-trip: got %q want %q", got, p)
		}
	}
}

func TestNilAndLegacyPlaintext(t *testing.T) {
	var nilC *Cipher
	if nilC.Encrypt("x") != "x" || nilC.Decrypt("x") != "x" {
		t.Error("nil cipher must be a no-op (encryption disabled)")
	}
	c, _ := New(testKey)
	if c.Decrypt("legacy-plaintext") != "legacy-plaintext" {
		t.Error("legacy plaintext without the marker must pass through unchanged")
	}
}

func TestEmptyKeyDisabled(t *testing.T) {
	c, err := New("")
	if err != nil || c != nil {
		t.Fatalf("empty key should give a nil cipher and no error; got %v, %v", c, err)
	}
}
