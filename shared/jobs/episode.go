package jobs

import (
	"slices"
	"strings"

	tnp "github.com/torrin-app/torrent-name-parser"
)

// MatchesEpisodeFile applies job metadata when it is available, but prefers
// explicit episode information in the filename. That keeps season packs
// seekable while preventing one episode request from returning every file in
// the pack. single reports whether the job holds a single video file; the job
// metadata fallback is only trustworthy then, since a pack shares one
// Season/Episode across every file.
func MatchesEpisodeFile(j *Job, fileName string, season, episode int, single bool) bool {
	if season < 0 || episode <= 0 {
		return true
	}

	info, err := tnp.ParseName(strings.ReplaceAll(fileName, ".", " "))
	if err == nil && len(info.Episodes) > 0 {
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

	return single && j != nil && j.Season == season && j.Episode == episode
}

// FilesForEpisode returns the matching files with stable original indexes.
// When a row never persisted File.Index (every index is zero) the slice
// position is used as the storage-key index; when any index is set they are
// trusted as-is so a legitimate index 0 is not overwritten.
func FilesForEpisode(j *Job, files []File, season, episode int) []File {
	indexed := false
	for _, f := range files {
		if f.Index > 0 {
			indexed = true
			break
		}
	}
	single := len(files) == 1
	out := make([]File, 0, len(files))
	for position, file := range files {
		if !indexed {
			file.Index = position
		}
		if MatchesEpisodeFile(j, file.Name, season, episode, single) {
			out = append(out, file)
		}
	}
	return out
}
