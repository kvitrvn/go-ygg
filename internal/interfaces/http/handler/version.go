package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kvitrvn/go-ygg/internal/version"
)

func Version(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(version.Get())
}
