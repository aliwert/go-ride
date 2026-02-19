package repository

import (
	"context"

	"github.com/aliwert/go-ride/internal/modules/identity/domain/entity"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.Status) error
}
