package pvz

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/models"
)

// PVZ usecase interface
type UseCase interface {
	GenerateJWT(ctx context.Context, role pvzapi.UserRole) (string, error)
	Register(ctx context.Context, email, password, role string) (models.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	CreatePVZ(ctx context.Context, id *uuid.UUID, city string, registrationDate *time.Time) (models.PVZ, error)
	CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error)
	AddProduct(ctx context.Context, pvzID uuid.UUID, productType string) (models.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error
	CloseLastReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error)
	GetPVZs(ctx context.Context, params pvzapi.GetPvzParams) ([]*models.PVZWithReceptions, error)
	GetPVZList(ctx context.Context) ([]models.PVZ, error)
}
