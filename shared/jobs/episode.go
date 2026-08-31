package jobs

import (
	"slices"

	ptt "github.com/MunifTanjim/go-ptt"
)

// MatchesEpisodeFile applies job metadata when it is available, but prefers
// explicit episode information in the filename. That keeps season packs
// seekable while preventing one episode request from returning every file in
// the pack.
func MatchesEpisodeFile(j *Job, fileName string, season, episode int) bool {
	if season < 0 || episode <= 0 {
		return true
	}

	info := ptt.Parse(fileName)
	if info.Error() == nil && len(info.Episodes) > 0 {
		if !slices.Contains(info.Episodes, episode) {
			return false
		}
		if len(info.Seasons) == 0 {
			if j != nil && (j.Season > 0 || j.Episode > 0) {
				return j.Season == season
			}
			return season == 1
		}
		return slices.Contains(info.Seasons, season)
	}

	return j != nil && j.Season == season && j.Episode == episode
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
