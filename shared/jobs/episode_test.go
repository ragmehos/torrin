package jobs

import "testing"

func TestFilesForEpisodeFiltersSeasonAndEpisode(t *testing.T) {
	j := &Job{Season: 12, Episode: 1, Files: []File{
		{Name: "Paw.Patrol.S05E01.mkv"},
		{Name: "Paw.Patrol.S12E01.mkv"},
		{Name: "Paw Patrol 12x02.mkv"},
		{Name: "opaque-video.mkv"},
	}}
	got := FilesForEpisode(j, j.Files, 12, 2)
	if len(got) != 1 || got[0].Name != "Paw Patrol 12x02.mkv" || got[0].Index != 2 {
		t.Fatalf("files = %+v, want only S12E02 at index 2", got)
	}
}

func TestFilesForEpisodeDoesNotUseStalePackMetadata(t *testing.T) {
	j := &Job{Season: 12, Episode: 1, Files: []File{
		{Name: "Paw.Patrol.S05E01.mkv"},
		{Name: "Paw.Patrol.S12E01.mkv"},
	}}
	got := FilesForEpisode(j, j.Files, 5, 1)
	if len(got) != 1 || got[0].Name != "Paw.Patrol.S05E01.mkv" {
		t.Fatalf("files = %+v, want only requested S05E01", got)
	}
}

func TestFilesForEpisodeExplicitFilenameBeatsWrongFolder(t *testing.T) {
	j := &Job{Files: []File{{Name: "Season 12/Paw.Patrol.S05E01.mkv"}}}
	if got := FilesForEpisode(j, j.Files, 12, 1); len(got) != 0 {
		t.Fatalf("wrong folder overrode explicit filename: %+v", got)
	}
	if got := FilesForEpisode(j, j.Files, 5, 1); len(got) != 1 {
		t.Fatalf("explicit S05E01 did not match: %+v", got)
	}
}

func TestFilesForEpisodeLeavesMoviesUnfiltered(t *testing.T) {
	j := &Job{Files: []File{{Name: "Movie.mkv"}, {Name: "Movie.extra.mkv"}}}
	got := FilesForEpisode(j, j.Files, 0, 0)
	if len(got) != 2 {
		t.Fatalf("movie files = %d, want 2", len(got))
	}
}

func TestFilesForEpisodeKeepsEveryRealMatch(t *testing.T) {
	files := []File{
		{Name: "Paw.Patrol.S12E03.alt.mkv", Size: 900},
		{Name: "Paw.Patrol.S12E03.mkv", Size: 1000},
		{Name: "Paw.Patrol.S12E04.mkv", Size: 1000},
	}
	got := FilesForEpisode(&Job{}, files, 12, 3)
	if len(got) != 2 {
		t.Fatalf("files = %+v, want both S12E03 matches", got)
	}
}

func TestFilesForEpisodeMatchesMultiEpisodeFile(t *testing.T) {
	file := File{Name: "Paw.Patrol.S12E01-E03.mkv"}
	for episode := 1; episode <= 3; episode++ {
		if got := FilesForEpisode(&Job{}, []File{file}, 12, episode); len(got) != 1 {
			t.Fatalf("episode %d did not match multi-episode file: %+v", episode, got)
		}
	}
	if got := FilesForEpisode(&Job{}, []File{file}, 12, 4); len(got) != 0 {
		t.Fatalf("episode 4 incorrectly matched: %+v", got)
	}
}
