package jobs

import "testing"

func TestMatchesEpisodeFile(t *testing.T) {
	tests := []struct {
		name     string
		job      *Job
		fileName string
		season   int
		episode  int
		want     bool
	}{
		{name: "filename exact", job: &Job{}, fileName: "Paw.Patrol.S05E01.mkv", season: 5, episode: 1, want: true},
		{name: "filename wrong episode", job: &Job{}, fileName: "Paw.Patrol.S05E02.mkv", season: 5, episode: 1},
		{name: "filename wrong season", job: &Job{}, fileName: "Paw.Patrol.S12E01.mkv", season: 5, episode: 1},
		{name: "season pack metadata", job: &Job{Season: 5}, fileName: "Paw.Patrol.S05E01.mkv", season: 5, episode: 1, want: true},
		{name: "wrong pack metadata", job: &Job{Season: 12}, fileName: "Paw.Patrol.S12E01.mkv", season: 5, episode: 1},
		{name: "exact metadata fallback", job: &Job{Season: 5, Episode: 1}, fileName: "opaque-video.mkv", season: 5, episode: 1, want: true},
		{name: "metadata does not override filename", job: &Job{Season: 5, Episode: 1}, fileName: "Paw.Patrol.S05E02.mkv", season: 5, episode: 1},
		{name: "movie is unfiltered", job: &Job{}, fileName: "anything.mkv", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesEpisodeFile(tt.job, tt.fileName, tt.season, tt.episode); got != tt.want {
				t.Fatalf("MatchesEpisodeFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilesForEpisodePreservesOriginalIndex(t *testing.T) {
	j := &Job{Season: 5, Files: []File{
		{Name: "Paw.Patrol.S05E01.mkv"},
		{Name: "Paw.Patrol.S05E02.mkv"},
	}}
	got := FilesForEpisode(j, j.Files, 5, 2)
	if len(got) != 1 {
		t.Fatalf("files = %d, want 1", len(got))
	}
	if got[0].Index != 1 {
		t.Fatalf("index = %d, want original index 1", got[0].Index)
	}
}
