package web

import (
	"context"

	appiam "github.com/kvitrvn/go-ygg/internal/application/iam"
)

type authContextKey struct{}

func WithAuthContext(ctx context.Context, auth *appiam.AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, auth)
}

func AuthFromContext(ctx context.Context) *appiam.AuthContext {
	auth, _ := ctx.Value(authContextKey{}).(*appiam.AuthContext)
	return auth
}
