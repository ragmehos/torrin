package download

import "testing"

func TestIsOptional(t *testing.T) {
	skip := []string{"cover.jpg", "poster.PNG", "info.nfo", "files.sfv", "readme.txt"}
	keep := []string{"movie.mkv", "movie.part01.rar", "data.r00", "archive.par2", "ep.s01e02.mp4"}
	for _, n := range skip {
		if !isOptional(n) {
			t.Errorf("%q should be optional (skipped)", n)
		}
	}
	for _, n := range keep {
		if isOptional(n) {
			t.Errorf("%q should NOT be optional (par2 kept for repair, media kept)", n)
		}
	}
}
