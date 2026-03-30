package apphttp

import (
	"net/http"

	"github.com/kvitrvn/go-ygg/internal/interfaces/http/handler"
)

func newRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handler.Health)
	mux.HandleFunc("GET /version", handler.Version)

	return mux
}
