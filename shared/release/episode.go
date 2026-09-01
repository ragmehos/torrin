package release

import (
	"slices"
	"strings"

	tnp "github.com/torrin-app/torrent-name-parser"
)

func MatchesEpisode(title string, season, episode int) bool {
	if season < 0 || episode <= 0 {
		return false
	}
	info, err := tnp.ParseName(strings.ReplaceAll(title, ".", " "))
	if err != nil {
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
