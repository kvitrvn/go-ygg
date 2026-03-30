package persistence

import (
	"context"
	"errors"

	domain "github.com/kvitrvn/go-ygg/internal/domain/example"
)

// TODO: Add your DB driver. Example with pgx v5:
//
//	import (
//	    "github.com/jackc/pgx/v5/pgxpool"
//	    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
//	)
//
// Then inject *pgxpool.Pool and replace the stub methods with real queries.

// ExampleRepository implements domain.Repository.
type ExampleRepository struct {
	// pool *pgxpool.Pool
}

func NewExampleRepository() *ExampleRepository {
	return &ExampleRepository{}
}

func (r *ExampleRepository) FindByID(_ context.Context, _ string) (*domain.Example, error) {
	return nil, errors.New("not implemented: inject a DB connection")
}

func (r *ExampleRepository) Save(_ context.Context, _ *domain.Example) error {
	return errors.New("not implemented: inject a DB connection")
}

func (r *ExampleRepository) Delete(_ context.Context, _ string) error {
	return errors.New("not implemented: inject a DB connection")
}
