package handler

import (
	"net/http"

	"github.com/kvitrvn/go-ygg/internal/interfaces/http/templates"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = templates.HomePage(false).Render(r.Context(), w)
}
