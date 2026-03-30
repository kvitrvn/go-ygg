package example

import "context"

// Repository is the outbound port for Example persistence.
// Implement it in internal/infrastructure/persistence/.
type Repository interface {
	FindByID(ctx context.Context, id string) (*Example, error)
	Save(ctx context.Context, e *Example) error
	Delete(ctx context.Context, id string) error
}
