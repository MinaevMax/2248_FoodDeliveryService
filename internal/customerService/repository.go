package foodservice

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, comment *models.Customer, log *log.Logger) (*models.Customer, error)
	Update(ctx context.Context, comment *models.Customer, log *log.Logger) (*models.Customer, error)
	Delete(ctx context.Context, commentID uuid.UUID, log *log.Logger) error
	GetByID(ctx context.Context, commentID uuid.UUID, log *log.Logger) (*models.Customer, error)
}
