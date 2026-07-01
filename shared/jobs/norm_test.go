package jobs

import "testing"

func TestNormTitle(t *testing.T) {
	cases := map[string]string{
		"The Matrix":            "matrix",    // stopword dropped
		"The Lord of the Rings": "lordrings", // stopwords "the","of" dropped
		"Avatar 2009":           "avatar",    // year token dropped
		"Mr. Robot":             "mrrobot",   // punctuation collapsed
		"2012":                  "2012",      // year-only falls back to raw
		"The":                   "the",       // all-stopword falls back to raw
	}
	for in, want := range cases {
		if got := NormTitle(in); got != want {
			t.Errorf("NormTitle(%q)=%q want %q", in, got, want)
		}
	}
}

func TestTitleNormFromName(t *testing.T) {
	if got := titleNormFromName("The.Matrix.1999.1080p.BluRay.x264"); got != "matrix" {
		t.Errorf("titleNormFromName matrix release = %q", got)
	}
}
