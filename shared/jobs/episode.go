package jobs

import (
	"strings"

	tnp "github.com/torrin-app/torrent-name-parser"
)

// MatchesEpisodeFile applies job metadata when it is available, but prefers
// explicit episode information in the filename. That keeps season packs
// seekable while preventing one episode request from returning every file in
// the pack.
func MatchesEpisodeFile(j *Job, fileName string, season, episode int) bool {
	if season <= 0 || episode <= 0 {
		return true
	}
	if j != nil && j.Season > 0 && j.Season != season {
		return false
	}

	info, err := tnp.ParseName(strings.ReplaceAll(fileName, ".", " "))
	if err == nil && info.Episode > 0 {
		if info.Episode != episode {
			return false
		}
		if len(info.Seasons) == 0 {
			return season == 1
		}
		for _, fileSeason := range info.Seasons {
			if fileSeason == season {
				return true
			}
		}
		return false
	}

	return j != nil && j.Episode == episode && (j.Season == 0 || j.Season == season)
}

// FilesForEpisode returns the matching files with stable original indexes.
// Some legacy rows did not persist File.Index, so their slice position is the
// fallback used by the storage key convention.
func FilesForEpisode(j *Job, files []File, season, episode int) []File {
	out := make([]File, 0, len(files))
	for position, file := range files {
		if file.Index == 0 && position > 0 {
			file.Index = position
		}
		if MatchesEpisodeFile(j, file.Name, season, episode) {
			out = append(out, file)
		}
	}
	return out
}
