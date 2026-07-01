package handlers

import "testing"

func TestExtractInfoHash(t *testing.T) {
	hash := "c12fe1c06bba254a9dc9f519b335aa7c1367a88a"
	cases := map[string]string{
		"magnet:?xt=urn:btih:" + hash + "&dn=movie": hash,
		hash: hash, // bare hash
		"magnet:?xt=urn:btih:C12FE1C06BBA254A9DC9F519B335AA7C1367A88A": hash, // uppercase
		"not a magnet":              "",
		"magnet:?xt=urn:btih:short": "",
	}
	for in, want := range cases {
		if got := extractInfoHash(in); got != want {
			t.Errorf("extractInfoHash(%q) = %q, want %q", in, got, want)
		}
	}
}
