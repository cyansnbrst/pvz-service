package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/cyansnbrst/pvz-service/config"
	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/models"
	"github.com/cyansnbrst/pvz-service/internal/pvz"
	"github.com/cyansnbrst/pvz-service/pkg/db"
)

// PVZ usecase struct
type pvzUC struct {
	cfg     *config.Config
	pvzRepo pvz.Repository
}

var (
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrInvalidRole       = errors.New("invalid role")
	ErrInvalidCity       = errors.New("invalid city")
	ErrInvalidType       = errors.New("invalid type")
	ErrInvalidDateRange  = errors.New("invalid date range")
)

// PVZ usecase constructor
func NewPVZUseCase(cfg *config.Config, pvzRepo pvz.Repository) pvz.UseCase {
	return &pvzUC{
		cfg:     cfg,
		pvzRepo: pvzRepo,
	}
}

// Generates JWT token for the given role
func (u *pvzUC) GenerateJWT(ctx context.Context, role pvzapi.UserRole) (string, error) {
	const op = "PVZ.GenerateJWT"

	expirationTime := jwt.NewNumericDate(time.Now().Add(u.cfg.App.JWTTokenTTL))
	claims := jwt.MapClaims{
		"role": role,
		"exp":  expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(u.cfg.App.JWTSecretKey))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return signedToken, nil
}

// Register user
func (u *pvzUC) Register(ctx context.Context, email, password, role string) (models.User, error) {
	const op = "PVZ.Register"

	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	uuid := uuid.New()

	var newUser models.User
	newUser.Email = email
	newUser.ID = uuid
	newUser.Role = role
	newUser.PasswordHash = hashedPassword

	err = u.pvzRepo.CreateUser(ctx, newUser)
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			return models.User{}, err
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return newUser, nil
}

// Login user
func (u *pvzUC) Login(ctx context.Context, email, password string) (string, error) {
	const op = "PVZ.Login"

	user, err := u.pvzRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) {
			return "", err
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = u.validatePassword(user, password)
	if err != nil {
		if errors.Is(err, ErrIncorrectPassword) {
			return "", err
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return u.GenerateJWT(ctx, pvzapi.UserRole(user.Role))
}

// Validate password
func (u *pvzUC) validatePassword(user *models.User, password string) error {
	const op = "PVZ.ValidatePassword"

	match, err := argon2id.ComparePasswordAndHash(password, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !match {
		return ErrIncorrectPassword
	}

	return nil
}

// Create PVZ
func (u *pvzUC) CreatePVZ(ctx context.Context, id *uuid.UUID, city string, registrationDate *time.Time) (models.PVZ, error) {
	const op = "PVZ.CreatePVZ"

	var pvzID uuid.UUID
	if id == nil {
		pvzID = uuid.New()
	} else {
		pvzID = *id
	}

	var regDate time.Time
	if registrationDate == nil {
		regDate = time.Now()
	} else {
		regDate = *registrationDate
	}

	newPVZ := models.PVZ{
		ID:               pvzID,
		City:             city,
		RegistrationDate: regDate,
	}

	err := u.pvzRepo.CreatePVZ(ctx, newPVZ)
	if err != nil {
		if errors.Is(err, db.ErrDuplicatePVZ) {
			return models.PVZ{}, err
		}
		return models.PVZ{}, fmt.Errorf("%s: %w", op, err)
	}

	return newPVZ, nil
}

// Create a new reception for the pvz
func (u *pvzUC) CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	const op = "PVZ.CreateReception"

	uuid := uuid.New()

	reception, err := u.pvzRepo.CreateReception(ctx, uuid, pvzID)
	if err != nil {
		if errors.Is(err, db.ErrReceptionConflict) {
			return models.Reception{}, err
		}
		return models.Reception{}, fmt.Errorf("%s: %w", op, err)
	}

	return *reception, nil
}

// Add a new product for the reception
func (u *pvzUC) AddProduct(ctx context.Context, pvzID uuid.UUID, productType string) (models.Product, error) {
	const op = "PVZ.AddProduct"

	uuid := uuid.New()

	product, err := u.pvzRepo.AddProduct(ctx, uuid, pvzID, productType)
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) {
			return models.Product{}, err
		}
		return models.Product{}, fmt.Errorf("%s: %w", op, err)
	}

	return *product, nil
}

// Delete the last product from the reception
func (u *pvzUC) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const op = "PVZ.DeleteLastProduct"

	err := u.pvzRepo.DeleteLastProduct(ctx, pvzID)
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) || errors.Is(err, db.ErrNoProducts) {
			return err
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Close the last reception in the pvz
func (u *pvzUC) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	const op = "PVZ.CloseLastReception"

	reception, err := u.pvzRepo.CloseLastReception(ctx, pvzID)
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) {
			return models.Reception{}, err
		}
		return models.Reception{}, fmt.Errorf("%s: %w", op, err)
	}

	return *reception, nil
}

// List of PVZs with their receptions and products
func (u *pvzUC) GetPVZs(ctx context.Context, params pvzapi.GetPvzParams) ([]*models.PVZWithReceptions, error) {
	const op = "PVZ.GetPVZs"

	if params.Page == nil {
		defaultPage := 1
		params.Page = &defaultPage
	}
	if params.Limit == nil {
		defaultLimit := 10
		params.Limit = &defaultLimit
	}

	if *params.Page < 1 {
		*params.Page = 1
	}
	if *params.Limit < 1 || *params.Limit > 30 {
		*params.Limit = 10
	}

	if params.StartDate != nil && params.EndDate != nil && params.StartDate.After(*params.EndDate) {
		return nil, ErrInvalidDateRange
	}

	offset := (*params.Page - 1) * *params.Limit

	limitU := uint64(*params.Limit)
	offsetU := uint64(offset) //nolint:gosec

	pvzs, err := u.pvzRepo.GetPVZs(ctx, params.StartDate, params.EndDate, limitU, offsetU)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzs, nil
}

// List of all created PVZs
func (u *pvzUC) GetPVZList(ctx context.Context) ([]models.PVZ, error) {
	const op = "PVZ.GetPVZList"

	pvzs, err := u.pvzRepo.GetPVZList(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzs, nil
}
