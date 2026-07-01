package storage

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

func testClient() *Client {
	return &Client{publicURL: "https://cdn.torrin.app", signingKey: []byte("secret")}
}

func parse(signed string) (path string, expires int64, user, sig string) {
	u := strings.TrimPrefix(signed, "https://cdn.torrin.app/")
	q := strings.SplitN(u, "?", 2)
	path = q[0]
	for _, kv := range strings.Split(q[1], "&") {
		k, v, _ := strings.Cut(kv, "=")
		switch k {
		case "expires":
			expires, _ = strconv.ParseInt(v, 10, 64)
		case "u":
			user = v
		case "sig":
			sig = v
		}
	}
	return
}

func TestSignVerifyRoundTrip(t *testing.T) {
	c := testClient()
	path, exp, user, sig := parse(c.SignURL("abc/file_0/movie.mkv", time.Hour))
	if !c.Verify(path, exp, user, sig) {
		t.Fatal("valid signature rejected")
	}
}

func TestSignURLNodeRouting(t *testing.T) {
	c := testClient()
	c.SetNodeBases(map[string]string{"oldbox": "https://old.torrin.app:8084"})

	// Known node → its base; empty/unknown → the local publicURL.
	if got := c.SignURLNode("oldbox", "h/file_0/m.mkv", time.Hour); !strings.HasPrefix(got, "https://old.torrin.app:8084/") {
		t.Fatalf("oldbox not routed: %s", got)
	}
	if got := c.SignURLNode("", "h/file_0/m.mkv", time.Hour); !strings.HasPrefix(got, "https://cdn.torrin.app/") {
		t.Fatalf("empty node should use publicURL: %s", got)
	}
	if got := c.SignURLNode("ghost", "h/file_0/m.mkv", time.Hour); !strings.HasPrefix(got, "https://cdn.torrin.app/") {
		t.Fatalf("unknown node should fall back: %s", got)
	}

	// Host is NOT part of the signature: a link signed for one node verifies anywhere.
	signed := c.SignURLNode("oldbox", "h/file_0/m.mkv", time.Hour)
	p := strings.TrimPrefix(signed, "https://old.torrin.app:8084/")
	q := strings.SplitN(p, "?", 2)
	var exp int64
	var sig string
	for _, kv := range strings.Split(q[1], "&") {
		k, v, _ := strings.Cut(kv, "=")
		if k == "expires" {
			exp, _ = strconv.ParseInt(v, 10, 64)
		} else if k == "sig" {
			sig = v
		}
	}
	if !c.Verify(q[0], exp, "", sig) {
		t.Fatal("node-routed signature should verify regardless of host")
	}
}

func TestVerifyExpired(t *testing.T) {
	c := testClient()
	path, _, user, sig := parse(c.SignURL("abc/movie.mkv", time.Hour))
	if c.Verify(path, time.Now().Add(-time.Minute).Unix(), user, sig) {
		t.Fatal("expired link accepted")
	}
}

func TestVerifyTampered(t *testing.T) {
	c := testClient()
	path, exp, user, _ := parse(c.SignURL("abc/movie.mkv", time.Hour))
	if c.Verify(path, exp, user, "deadbeef") {
		t.Fatal("tampered signature accepted")
	}
}

func TestUserScopedSignature(t *testing.T) {
	c := testClient()
	path, exp, user, sig := parse(c.SignURLWithUser("abc/movie.mkv", "user-1", time.Hour))
	if user != "user-1" {
		t.Fatalf("user not in url: %q", user)
	}
	if !c.Verify(path, exp, "user-1", sig) {
		t.Fatal("valid user signature rejected")
	}
	if c.Verify(path, exp, "", sig) {
		t.Fatal("user link verified without the user")
	}
}
