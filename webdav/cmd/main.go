package main

import (
	"net/http"

	"github.com/torrin-app/torrin/shared/service"
	"github.com/torrin-app/torrin/webdav/internal/server"
)

func main() {
	service.RunDB("webdav", "9092", func(d service.DB) http.Handler {
		return server.New(d.Users, d.Jobs, d.Store).Handler()
	})
}
