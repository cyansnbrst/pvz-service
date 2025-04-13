package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/models"
	"github.com/cyansnbrst/pvz-service/internal/pvz"
	"github.com/cyansnbrst/pvz-service/pkg/db"
)

// Database interface
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// PVZ repository struct
type pvzRepo struct {
	db DB
}

// PVZ repository constructor
func NewPVZRepo(db DB) pvz.Repository {
	return &pvzRepo{db: db}
}

// Get user by email
func (r *pvzRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.GetUserByEmail"

	query := `
		SELECT id, email, password_hash, role
		FROM users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

// Create a new user
func (r *pvzRepo) CreateUser(ctx context.Context, user models.User) error {
	const op = "repository.CreateUser"

	query := `
        INSERT INTO users (id, email, password_hash, role)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (email) DO NOTHING
        RETURNING id
    `

	var id string
	err := r.db.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrDuplicateEmail
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Create a new pvz
func (r *pvzRepo) CreatePVZ(ctx context.Context, pvz models.PVZ) error {
	const op = "repository.CreatePVZ"

	query := `
        INSERT INTO pvzs (id, city, registration_date)
        VALUES ($1, $2, $3)
        ON CONFLICT (id) DO NOTHING
        RETURNING id
    `

	var id string
	err := r.db.QueryRow(ctx, query,
		pvz.ID,
		pvz.City,
		pvz.RegistrationDate,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrDuplicatePVZ
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Create a new reception
func (r *pvzRepo) CreateReception(ctx context.Context, receptionID, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "repository.CreateReception"

	query := `
		INSERT INTO receptions (id, pvz_id, status)
		SELECT $1, $2, $3::VARCHAR
		WHERE EXISTS (
			SELECT 1 FROM pvzs WHERE id = $2
		)
		AND NOT EXISTS (
			SELECT 1 FROM receptions WHERE pvz_id = $2 AND status = $3::VARCHAR
		)
		RETURNING id, date_time, pvz_id, status
	`

	defaultStatus := string(pvzapi.InProgress)

	var reception models.Reception
	err := r.db.QueryRow(ctx, query, receptionID, pvzID, defaultStatus).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PvzID,
		&reception.Status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrReceptionConflict
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &reception, nil
}

// Add a product in the reception
func (r *pvzRepo) AddProduct(ctx context.Context, productID, pvzID uuid.UUID, productType string) (*models.Product, error) {
	const op = "repository.AddProduct"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				log.Printf("%s: failed to rollback transaction: %v", op, rbErr)
			}
		}
	}()

	query := `
		SELECT id 
		FROM receptions 
        WHERE pvz_id = $1 AND status = $2::VARCHAR
        LIMIT 1
		FOR UPDATE
	`

	allowedStatus := string(pvzapi.InProgress)

	var receptionID uuid.UUID
	err = tx.QueryRow(ctx, query,
		pvzID,
		allowedStatus,
	).Scan(&receptionID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrNoOpenReception
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query = `
		INSERT INTO products (id, type, reception_id)
		VALUES ($1, $2, $3)
		RETURNING id, date_time, type, reception_id
	`

	var product models.Product
	err = tx.QueryRow(ctx, query,
		productID,
		productType,
		receptionID,
	).Scan(
		&product.ID,
		&product.DateTime,
		&product.Type,
		&product.ReceptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &product, nil
}

// Delete the last product from the reception
func (r *pvzRepo) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	const op = "repository.DeleteLastProduct"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				log.Printf("%s: failed to rollback transaction: %v", op, rbErr)
			}
		}
	}()

	query := `
        SELECT id 
        FROM receptions 
        WHERE pvz_id = $1 AND status = $2::VARCHAR
        LIMIT 1
        FOR UPDATE
    `

	allowedStatus := string(pvzapi.InProgress)

	var receptionID uuid.UUID
	err = tx.QueryRow(ctx, query,
		pvzID,
		allowedStatus,
	).Scan(&receptionID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrNoOpenReception
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	query = `
        DELETE FROM products
        WHERE id = (
            SELECT id 
            FROM products 
            WHERE reception_id = $1
            ORDER BY date_time DESC
            LIMIT 1
        )
        RETURNING id
    `

	var deletedID uuid.UUID
	err = tx.QueryRow(ctx, query, receptionID).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.ErrNoProducts
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Close the last reception in pvz
func (r *pvzRepo) CloseLastReception(ctx context.Context, pvzID uuid.UUID) (*models.Reception, error) {
	const op = "repository.CloseLastReception"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				log.Printf("%s: failed to rollback transaction: %v", op, rbErr)
			}
		}
	}()

	query := `
        SELECT id, date_time, pvz_id
        FROM receptions 
        WHERE pvz_id = $1 AND status = $2::VARCHAR
        LIMIT 1
        FOR UPDATE
    `

	allowedStatus := string(pvzapi.InProgress)

	var reception models.Reception
	err = tx.QueryRow(ctx, query,
		pvzID,
		allowedStatus,
	).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PvzID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, db.ErrNoOpenReception
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query = `
        UPDATE receptions
        SET status = $1
        WHERE id = $2
    `

	newStatus := string(pvzapi.Close)

	_, err = tx.Exec(ctx, query, newStatus, reception.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	reception.Status = newStatus

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &reception, nil
}

// Get the list of pvzs with their receptions and products with pagination by PVZ count
func (r *pvzRepo) GetPVZs(ctx context.Context, startDate, endDate *time.Time, limit, offset uint64) ([]*models.PVZWithReceptions, error) {
	const op = "repository.GetPVZs"

	pvzQueryBuilder := sq.
		Select("DISTINCT p.id").
		From("pvzs p").
		LeftJoin("receptions r ON r.pvz_id = p.id")

	var conditions sq.And
	if startDate != nil {
		conditions = append(conditions, sq.GtOrEq{"r.date_time": *startDate})
	}
	if endDate != nil {
		conditions = append(conditions, sq.LtOrEq{"r.date_time": *endDate})
	}

	if len(conditions) > 0 {
		pvzQueryBuilder = pvzQueryBuilder.Where(conditions)
	}

	pvzQueryBuilder = pvzQueryBuilder.
		Limit(limit).
		Offset(offset)

	pvzQuery, pvzArgs, err := pvzQueryBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := r.db.Query(ctx, pvzQuery, pvzArgs...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var pvzIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pvzIDs = append(pvzIDs, id)
	}
	rows.Close()

	if len(pvzIDs) == 0 {
		return []*models.PVZWithReceptions{}, nil
	}

	queryBuilder := sq.
		Select(
			"p.id", "p.city", "p.registration_date",
			"r.id", "r.date_time", "r.status",
			"pr.id", "pr.type", "pr.date_time",
		).
		From("pvzs p").
		LeftJoin("receptions r ON r.pvz_id = p.id").
		LeftJoin("products pr ON pr.reception_id = r.id").
		Where(sq.Eq{"p.id": pvzIDs}).
		OrderBy("r.date_time DESC")

	if len(conditions) > 0 {
		queryBuilder = queryBuilder.Where(conditions)
	}

	query, args, err := queryBuilder.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err = r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var result []*models.PVZWithReceptions
	var currentPVZ *models.PVZWithReceptions
	var currentReception *models.ReceptionWithProducts

	for rows.Next() {
		var (
			pvzID           uuid.UUID
			pvzCity         string
			pvzRegistration time.Time
			receptionID     *uuid.UUID
			receptionDate   *time.Time
			receptionStatus *string
			productID       *uuid.UUID
			productType     *string
			productDate     *time.Time
		)

		if err := rows.Scan(
			&pvzID,
			&pvzCity,
			&pvzRegistration,
			&receptionID,
			&receptionDate,
			&receptionStatus,
			&productID,
			&productType,
			&productDate,
		); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if currentPVZ == nil || currentPVZ.PVZ.ID != pvzID {
			currentPVZ = &models.PVZWithReceptions{
				PVZ: models.PVZ{
					ID:               pvzID,
					City:             pvzCity,
					RegistrationDate: pvzRegistration,
				},
				Receptions: []*models.ReceptionWithProducts{},
			}
			result = append(result, currentPVZ)
			currentReception = nil
		}

		if receptionID != nil {
			if currentReception == nil || currentReception.Reception.ID != *receptionID {
				currentReception = &models.ReceptionWithProducts{
					Reception: models.Reception{
						ID:       *receptionID,
						PvzID:    pvzID,
						DateTime: *receptionDate,
						Status:   *receptionStatus,
					},
					Products: []*models.Product{},
				}
				currentPVZ.Receptions = append(currentPVZ.Receptions, currentReception)
			}

			if productID != nil {
				product := &models.Product{
					ID:          *productID,
					DateTime:    *productDate,
					Type:        *productType,
					ReceptionID: *receptionID,
				}
				currentReception.Products = append(currentReception.Products, product)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

// Get list of all pvzs
func (r *pvzRepo) GetPVZList(ctx context.Context) ([]models.PVZ, error) {
	const op = "repository.GetPVZList"

	query := `
		SELECT id, city, registration_date
		FROM pvzs
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var pvzs []models.PVZ
	for rows.Next() {
		var pvz models.PVZ
		err := rows.Scan(
			&pvz.ID,
			&pvz.City,
			&pvz.RegistrationDate,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		pvzs = append(pvzs, pvz)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pvzs, nil
}
