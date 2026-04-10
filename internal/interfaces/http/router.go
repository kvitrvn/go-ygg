package apphttp

import (
	"errors"
	"net/http"
	"net/url"
	"time"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	domain "github.com/kvitrvn/go-ygg/internal/domain/iam"
	"github.com/kvitrvn/go-ygg/internal/interfaces/http/handler"
	"github.com/kvitrvn/go-ygg/internal/interfaces/http/web"
)

func newRouter(iamService *appiam.Service, cookieConfig web.CookieConfig, appBaseURL string, sessionTTL time.Duration) http.Handler {
	mux := http.NewServeMux()
	h := web.NewHandler(iamService, cookieConfig, sessionTTL)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))

	mux.HandleFunc("GET /healthz", handler.Health)
	mux.HandleFunc("GET /version", handler.Version)
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /signup", h.ShowSignUp)
	mux.HandleFunc("POST /signup", h.SignUp)
	mux.HandleFunc("GET /login", h.ShowLogin)
	mux.HandleFunc("POST /login", h.SignIn)
	mux.HandleFunc("POST /logout", h.SignOut)
	mux.Handle("GET /app", requireAuthenticated(http.HandlerFunc(h.Dashboard)))
	mux.Handle("POST /app/tenants/switch", requireAuthenticated(http.HandlerFunc(h.SwitchTenant)))
	mux.Handle("GET /app/organizations/new", requireAuthenticated(http.HandlerFunc(h.ShowOrganizationCreate)))
	mux.Handle("POST /app/organizations", requireAuthenticated(http.HandlerFunc(h.CreateOrganization)))
	mux.Handle("GET /app/members", requireAuthenticated(http.HandlerFunc(h.Members)))
	mux.Handle("GET /app/invitations/new", requireRole(http.HandlerFunc(h.ShowInvitationCreate), domain.RoleOwner, domain.RoleAdmin))
	mux.Handle("POST /app/invitations", requireRole(http.HandlerFunc(h.CreateInvitation), domain.RoleOwner, domain.RoleAdmin))
	mux.HandleFunc("GET /invitations/{token}", h.ShowInvitation)
	mux.HandleFunc("POST /invitations/{token}/accept", h.AcceptInvitation)

	return sessionMiddleware(iamService, cookieConfig)(csrfMiddleware(appBaseURL)(mux))
}

func sessionMiddleware(iamService *appiam.Service, cookieConfig web.CookieConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := web.ReadSessionToken(r, cookieConfig)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			auth, err := iamService.ResolveSession(r.Context(), token)
			if err != nil {
				if errors.Is(err, domain.ErrUnauthenticated) {
					web.ClearSessionCookie(w, cookieConfig)
					next.ServeHTTP(w, r)
					return
				}
				http.Error(w, "session resolution failed", http.StatusInternalServerError)
				return
			}

			next.ServeHTTP(w, r.WithContext(web.WithAuthContext(r.Context(), auth)))
		})
	}
}

func csrfMiddleware(appBaseURL string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		baseURL, _ := url.Parse(appBaseURL)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
				if !sameOrigin(r, baseURL) {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func requireAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if web.AuthFromContext(r.Context()) == nil {
			web.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireRole(next http.Handler, roles ...domain.Role) http.Handler {
	return requireAuthenticated(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := web.AuthFromContext(r.Context())
		if auth == nil {
			web.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		for _, role := range roles {
			if auth.ActiveMembership.Membership.Role == role {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
}

func sameOrigin(r *http.Request, baseURL *url.URL) bool {
	if baseURL == nil {
		return true
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		u, err := url.Parse(origin)
		return err == nil && u.Scheme == baseURL.Scheme && u.Host == baseURL.Host
	}
	if referer := r.Header.Get("Referer"); referer != "" {
		u, err := url.Parse(referer)
		return err == nil && u.Scheme == baseURL.Scheme && u.Host == baseURL.Host
	}
	return false
}
