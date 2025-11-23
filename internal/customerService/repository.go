package foodservice

import (
	"context"
	"log"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, comment *models.Customer, log *log.Logger) (*models.Customer, error)
	Update(ctx context.Context, comment *models.Customer, log *log.Logger) (*models.Customer, error)
	Delete(ctx context.Context, customerID uuid.UUID, log *log.Logger) error
	GetByID(customerID uuid.UUID, log *log.Logger) (*models.Customer, error)
}
