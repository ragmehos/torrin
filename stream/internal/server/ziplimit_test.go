package server

import "testing"

func TestZipGuard(t *testing.T) {
	g := newZipGuard(2)
	if !g.acquire("u1") {
		t.Fatal("first acquire should pass")
	}
	if !g.acquire("u1") {
		t.Fatal("second acquire should pass")
	}
	if g.acquire("u1") {
		t.Fatal("third acquire should be blocked")
	}
	if !g.acquire("u2") {
		t.Fatal("other user should not be blocked")
	}
	g.release("u1")
	if !g.acquire("u1") {
		t.Fatal("acquire should pass after release")
	}
	if !g.acquire("") {
		t.Fatal("empty user is unbounded")
	}
}
