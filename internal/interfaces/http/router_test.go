package apphttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuthenticatedRedirectsHTMXRequests(t *testing.T) {
	handler := requireAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/app", nil)
	request.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("HX-Redirect"); got != "/login" {
		t.Fatalf("HX-Redirect = %q, want /login", got)
	}
}
