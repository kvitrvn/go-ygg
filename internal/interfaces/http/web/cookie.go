package web

import (
	"net/http"
	"time"
)

type CookieConfig struct {
	Name   string
	Secure bool
}

func ReadSessionToken(r *http.Request, cfg CookieConfig) string {
	cookie, err := r.Cookie(cfg.Name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func SetSessionCookie(w http.ResponseWriter, cfg CookieConfig, token string, ttl time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
	})
}

func ClearSessionCookie(w http.ResponseWriter, cfg CookieConfig) {
	http.SetCookie(w, &http.Cookie{
		Name:     cfg.Name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}
