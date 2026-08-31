package release

import (
	"slices"

	ptt "github.com/MunifTanjim/go-ptt"
)

func MatchesEpisode(title string, season, episode int) bool {
	if season < 0 || episode <= 0 {
		return false
	}
	info := ptt.Parse(title)
	if info.Error() != nil {
		return false
	}
	if len(info.Episodes) > 0 {
		if !slices.Contains(info.Episodes, episode) {
			return false
		}
		if len(info.Seasons) == 0 {
			return season == 1
		}
		return slices.Contains(info.Seasons, season)
	}
	return slices.Contains(info.Seasons, season)
}
