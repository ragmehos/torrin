package release

import (
	"strings"

	tnp "github.com/torrin-app/torrent-name-parser"
)

func MatchesEpisode(title string, season, episode int) bool {
	info, err := tnp.ParseName(strings.ReplaceAll(title, ".", " "))
	if err != nil {
		return false
	}
	if info.Episode == 0 {
		if info.Season == season {
			return true
		}
		for _, s := range info.Seasons {
			if s == season {
				return true
			}
		}
		return false
	}
	return info.Season == season && info.Episode == episode
}
