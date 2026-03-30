package example

import (
	"context"
	"fmt"

	domain "github.com/kvitrvn/go-ygg/internal/domain/example"
)

// UseCase orchestrates domain logic for the Example bounded context.
type UseCase struct {
	repo domain.Repository
}

func NewUseCase(repo domain.Repository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Get(ctx context.Context, id string) (*domain.Example, error) {
	e, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get example: %w", err)
	}
	return e, nil
}

func (uc *UseCase) Create(ctx context.Context, id, name string) (*domain.Example, error) {
	e := domain.NewExample(id, name)
	if err := uc.repo.Save(ctx, e); err != nil {
		return nil, fmt.Errorf("create example: %w", err)
	}
	return e, nil
}
