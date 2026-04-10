package apphttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
	"github.com/kvitrvn/go-ygg/internal/interfaces/http/web"
)

// Server wraps net/http.Server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, iamService *appiam.Service, cookieConfig web.CookieConfig, appBaseURL string, sessionTTL time.Duration) *Server {
	router := newRouter(iamService, cookieConfig, appBaseURL, sessionTTL)
	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start listens and blocks until SIGINT/SIGTERM is received, then shuts down gracefully.
func (s *Server) Start() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}
