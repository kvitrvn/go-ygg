package apphttp

import (
	"net/http"

	"github.com/kvitrvn/go-ygg/internal/interfaces/http/handler"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))

	mux.HandleFunc("GET /healthz", handler.Health)
	mux.HandleFunc("GET /version", handler.Version)
	mux.HandleFunc("GET /", handler.Home)

	return mux
}
