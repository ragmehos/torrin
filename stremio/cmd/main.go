package main

import (
	"net/http"

	"github.com/torrin-app/torrin/shared/service"
	"github.com/torrin-app/torrin/stremio/internal/addon"
)

func main() {
	service.RunDB("stremio", "7000", func(d service.DB) http.Handler {
		return addon.New(d.Users, d.Jobs, d.Store).Handler()
	})
}
