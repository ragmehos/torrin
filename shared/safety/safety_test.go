package safety

import "testing"

func TestScreen(t *testing.T) {
	SetTerms([]string{"badterm"}, []string{"flagme"})

	if v := Screen("a badterm here"); !v.Blocked || !v.Ban {
		t.Error("hard term should block + ban")
	}
	if v := Screen("contains flagme"); !v.Blocked || v.Ban {
		t.Error("soft term should block without ban")
	}
	if v := Screen("clean movie 1080p"); v.Blocked {
		t.Error("clean text should pass")
	}
	if v := Screen("http://foo.onion/x"); !v.Blocked || !v.Ban {
		t.Error(".onion should block + ban")
	}
}

func TestScreenWholeWord(t *testing.T) {
	SetTerms([]string{"9yo"}, nil)
	if v := Screen(`"tc45fc19yOA9vw2YHbF.part16.rar"`); v.Blocked {
		t.Error("term inside a hash must not match (false positive)")
	}
	if v := Screen("some 9yo thing"); !v.Blocked || !v.Ban {
		t.Error("standalone term should still block + ban")
	}
}
