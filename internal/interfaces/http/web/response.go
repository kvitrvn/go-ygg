package web

import "net/http"

func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func AddHTMLVary(w http.ResponseWriter) {
	w.Header().Add("Vary", "HX-Request")
}

func Redirect(w http.ResponseWriter, r *http.Request, location string, status int) {
	AddHTMLVary(w)
	if IsHTMXRequest(r) {
		w.Header().Set("HX-Redirect", location)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, location, status)
}
