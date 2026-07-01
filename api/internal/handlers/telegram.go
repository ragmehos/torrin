package handlers

import (
	"net/http"
	"time"

	"github.com/torrin-app/torrin/api/internal/middleware"
	"github.com/torrin-app/torrin/api/internal/web"
)

func (s *Server) registerTelegramRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("POST /api/telegram/link-code", authMW(http.HandlerFunc(s.telegramLinkCode)))
}

func (s *Server) telegramLinkCode(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	code := genLinkCode()
	if err := s.Users.CreateTelegramLinkCode(r.Context(), code, user.ID, 15*time.Minute); err != nil {
		web.WriteError(w, 500, "could not create link code")
		return
	}
	resp := map[string]any{"code": code, "expires_in": 900}
	if s.TGBotUsername != "" {
		resp["bot"] = s.TGBotUsername
		resp["deeplink"] = "https://t.me/" + s.TGBotUsername + "?start=" + code
	}
	web.WriteJSON(w, 200, resp)
}

func genLinkCode() string { return randCode(8) }
