package env

import (
	"os"
	"strconv"
)

func Get(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func Int(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}
