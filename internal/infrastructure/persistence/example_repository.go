package persistence

import (
	"context"
	"errors"

	domain "github.com/kvitrvn/go-ygg/internal/domain/example"
)

// TODO: Add your DB driver blank import here, for example:
//
//	_ "github.com/golang-migrate/migrate/v4/database/postgres"
//	_ "github.com/lib/pq"
//
// Then replace the stub methods with real DB queries.

// ExampleRepository implements domain.Repository.
type ExampleRepository struct {
	// db *sql.DB  ← inject your DB connection
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
