package pvz

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/cyansnbrst/pvz-service/internal/models"
)

// PVZ repository interface
type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) error
	CreatePVZ(ctx context.Context, pvz models.PVZ) error
	CreateReception(ctx context.Context, receptionID, pvzID uuid.UUID) (*models.Reception, error)
	AddProduct(ctx context.Context, productID, pvzID uuid.UUID, productType string) (*models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error)
	GetPVZs(ctx context.Context, startDate, endDate *time.Time, limit, offset uint64) ([]*models.PVZWithReceptions, error)
	GetPVZList(ctx context.Context) ([]models.PVZ, error)
}
