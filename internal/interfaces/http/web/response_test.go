package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testComponent struct {
	body string
}

func (c testComponent) Render(_ context.Context, w io.Writer) error {
	_, err := w.Write([]byte(c.body))
	return err
}

func TestRedirectForHTMXRequestUsesHXRedirect(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/login", nil)
	request.Header.Set("HX-Request", "true")

	Redirect(recorder, request, "/app", http.StatusSeeOther)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("HX-Redirect"); got != "/app" {
		t.Fatalf("HX-Redirect = %q, want /app", got)
	}
	if got := recorder.Header().Values("Vary"); len(got) == 0 || got[0] != "HX-Request" {
		t.Fatalf("Vary = %v, want HX-Request", got)
	}
}

func TestRenderAddsHTMXVaryHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/app", nil)

	render(recorder, request, http.StatusOK, testComponent{body: "<div>ok</div>"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", got)
	}
	if got := recorder.Header().Values("Vary"); len(got) == 0 || got[0] != "HX-Request" {
		t.Fatalf("Vary = %v, want HX-Request", got)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "<div>ok</div>") {
		t.Fatalf("body = %q, want rendered component", body)
	}
}

func TestRenderNormalizesHTMXClientErrorsTo200(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/login", nil)
	request.Header.Set("HX-Request", "true")

	render(recorder, request, http.StatusBadRequest, testComponent{body: "<div>invalid</div>"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); !strings.Contains(body, "<div>invalid</div>") {
		t.Fatalf("body = %q, want rendered component", body)
	}
}
