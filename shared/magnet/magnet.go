package magnet

import (
	"encoding/base32"
	"encoding/hex"
	"net/url"
	"strings"
)

func Valid(s string) bool {
	if len(s) != 40 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func normalize(h string) string {
	h = strings.TrimSpace(h)
	switch len(h) {
	case 40:
		if l := strings.ToLower(h); Valid(l) {
			return l
		}
	case 32:
		if b, err := base32.StdEncoding.DecodeString(strings.ToUpper(h)); err == nil && len(b) == 20 {
			return hex.EncodeToString(b)
		}
	}
	return ""
}

func Hash(magnet string) string {
	s := strings.TrimSpace(magnet)
	if h := normalize(s); h != "" {
		return h
	}
	const marker = "xt=urn:btih:"
	if i := strings.Index(s, marker); i >= 0 {
		h := s[i+len(marker):]
		if end := strings.Index(h, "&"); end >= 0 {
			h = h[:end]
		}
		return normalize(h)
	}
	return ""
}

func DisplayName(magnet string) string {
	u, err := url.Parse(magnet)
	if err != nil {
		return ""
	}
	return u.Query().Get("dn")
}

func Build(hash, name string) string {
	m := "magnet:?xt=urn:btih:" + strings.ToLower(hash)
	if name != "" {
		m += "&dn=" + url.QueryEscape(name)
	}
	return m
}

func Find(text string) string {
	i := strings.Index(text, "magnet:?")
	if i < 0 {
		return ""
	}
	rest := text[i:]
	if end := strings.IndexAny(rest, " \t\r\n\"'<>"); end >= 0 {
		return rest[:end]
	}
	return rest
}
