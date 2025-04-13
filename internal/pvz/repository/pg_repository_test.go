package repository

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/models"
	"github.com/cyansnbrst/pvz-service/pkg/db"
)

var ErrRandomError = errors.New("random error")

func TestPVZRepo_GetUserByEmail(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	testUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed_pass",
		Role:         "admin",
	}

	tests := []struct {
		name          string
		email         string
		mockSetup     func()
		expected      *models.User
		expectedError error
	}{
		{
			name:  "user found",
			email: testUser.Email,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(testUser.ID, testUser.Email, testUser.PasswordHash, testUser.Role)

				dbMock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = \\$1").
					WithArgs(testUser.Email).
					WillReturnRows(rows)
			},
			expected:      testUser,
			expectedError: nil,
		},
		{
			name:  "user not found",
			email: "test@example.com",
			mockSetup: func() {
				dbMock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnError(pgx.ErrNoRows)
			},
			expected:      nil,
			expectedError: db.ErrUserNotFound,
		},
		{
			name:  "query error",
			email: "test@example.com",
			mockSetup: func() {
				dbMock.ExpectQuery("SELECT id, email, password_hash, role FROM users WHERE email = \\$1").
					WithArgs("test@example.com").
					WillReturnError(ErrRandomError)
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetUserByEmail(context.Background(), tt.email)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZRepo_CreateUser(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	testUser := models.User{
		ID:           uuid.New(),
		Email:        "new@example.com",
		PasswordHash: "secure_hash",
		Role:         "user",
	}

	tests := []struct {
		name          string
		user          models.User
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful creation",
			user: testUser,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id"}).
					AddRow(testUser.ID.String())
				dbMock.ExpectQuery("INSERT INTO users.*RETURNING id").
					WithArgs(testUser.ID, testUser.Email, testUser.PasswordHash, testUser.Role).
					WillReturnRows(rows)
			},
			expectedError: nil,
		},
		{
			name: "duplicate email",
			user: testUser,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO users.*RETURNING id").
					WithArgs(testUser.ID, testUser.Email, testUser.PasswordHash, testUser.Role).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: db.ErrDuplicateEmail,
		},
		{
			name: "random error",
			user: testUser,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO users.*RETURNING id").
					WithArgs(testUser.ID, testUser.Email, testUser.PasswordHash, testUser.Role).
					WillReturnError(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.CreateUser(context.Background(), tt.user)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZRepo_CreatePVZ(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	testPVZ := models.PVZ{
		ID:               uuid.New(),
		City:             "Москва",
		RegistrationDate: time.Now(),
	}

	tests := []struct {
		name          string
		pvz           models.PVZ
		mockSetup     func()
		expectedError error
	}{
		{
			name: "successful creation",
			pvz:  testPVZ,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id"}).
					AddRow(testPVZ.ID.String())
				dbMock.ExpectQuery("INSERT INTO pvzs.*RETURNING id").
					WithArgs(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate).
					WillReturnRows(rows)
			},
			expectedError: nil,
		},
		{
			name: "duplicate PVZ",
			pvz:  testPVZ,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO pvzs.*RETURNING id").
					WithArgs(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedError: db.ErrDuplicatePVZ,
		},
		{
			name: "query error",
			pvz:  testPVZ,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO pvzs.*RETURNING id").
					WithArgs(testPVZ.ID, testPVZ.City, testPVZ.RegistrationDate).
					WillReturnError(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.CreatePVZ(context.Background(), tt.pvz)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZRepo_CreateReception(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	receptionID := uuid.New()
	pvzID := uuid.New()
	defaultStatus := string(pvzapi.InProgress)

	expectedReception := &models.Reception{
		ID:       receptionID,
		PvzID:    pvzID,
		Status:   defaultStatus,
		DateTime: time.Now(),
	}

	tests := []struct {
		name          string
		receptionID   uuid.UUID
		pvzID         uuid.UUID
		mockSetup     func()
		expected      *models.Reception
		expectedError error
	}{
		{
			name:        "successful creation",
			receptionID: receptionID,
			pvzID:       pvzID,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "date_time", "pvz_id", "status"}).
					AddRow(expectedReception.ID, expectedReception.DateTime, expectedReception.PvzID, expectedReception.Status)
				dbMock.ExpectQuery("INSERT INTO receptions.*RETURNING id, date_time, pvz_id, status").
					WithArgs(receptionID, pvzID, defaultStatus).
					WillReturnRows(rows)
			},
			expected:      expectedReception,
			expectedError: nil,
		},
		{
			name:        "conflict (already exists)",
			receptionID: receptionID,
			pvzID:       pvzID,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO receptions.*RETURNING id, date_time, pvz_id, status").
					WithArgs(receptionID, pvzID, defaultStatus).
					WillReturnError(pgx.ErrNoRows)
			},
			expected:      nil,
			expectedError: db.ErrReceptionConflict,
		},
		{
			name:        "query error",
			receptionID: receptionID,
			pvzID:       pvzID,
			mockSetup: func() {
				dbMock.ExpectQuery("INSERT INTO receptions.*RETURNING id, date_time, pvz_id, status").
					WithArgs(receptionID, pvzID, defaultStatus).
					WillReturnError(ErrRandomError)
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.CreateReception(context.Background(), tt.receptionID, tt.pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZRepo_AddProduct(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	productID := uuid.New()
	pvzID := uuid.New()
	productType := "обувь"
	receptionID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		mockSetup     func()
		expected      *models.Product
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func() {
				dbMock.ExpectBegin()

				rowsReception := pgxmock.NewRows([]string{"id"}).
					AddRow(receptionID)
				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(rowsReception)

				rowsProduct := pgxmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
					AddRow(productID, now, productType, receptionID)
				dbMock.ExpectQuery("INSERT INTO products.*RETURNING id, date_time, type, reception_id").
					WithArgs(productID, productType, receptionID).
					WillReturnRows(rowsProduct)

				dbMock.ExpectCommit()
			},
			expected: &models.Product{
				ID:          productID,
				DateTime:    now,
				Type:        productType,
				ReceptionID: receptionID,
			},
			expectedError: nil,
		},
		{
			name: "no open reception",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnError(pgx.ErrNoRows)

				dbMock.ExpectRollback()
			},
			expected:      nil,
			expectedError: db.ErrNoOpenReception,
		},
		{
			name: "insert product error",
			mockSetup: func() {
				dbMock.ExpectBegin()

				rowsReception := pgxmock.NewRows([]string{"id"}).
					AddRow(receptionID)
				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(rowsReception)

				dbMock.ExpectQuery("INSERT INTO products.*RETURNING id, date_time, type, reception_id").
					WithArgs(productID, productType, receptionID).
					WillReturnError(ErrRandomError)

				dbMock.ExpectRollback()
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
		{
			name: "begin transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin().WillReturnError(ErrRandomError)
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
		{
			name: "commit transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin()

				rowsReception := pgxmock.NewRows([]string{"id"}).
					AddRow(receptionID)
				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(rowsReception)

				rowsProduct := pgxmock.NewRows([]string{"id", "date_time", "type", "reception_id"}).
					AddRow(productID, now, productType, receptionID)
				dbMock.ExpectQuery("INSERT INTO products.*RETURNING id, date_time, type, reception_id").
					WithArgs(productID, productType, receptionID).
					WillReturnRows(rowsProduct)

				dbMock.ExpectCommit().WillReturnError(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.AddProduct(context.Background(), productID, pvzID, productType)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZRepo_DeleteLastProduct(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(receptionID))

				dbMock.ExpectQuery("DELETE FROM products.*RETURNING id").
					WithArgs(receptionID).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(productID))

				dbMock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name: "no open reception",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnError(pgx.ErrNoRows)

				dbMock.ExpectRollback()
			},
			expectedError: db.ErrNoOpenReception,
		},
		{
			name: "no products to delete",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(receptionID))

				dbMock.ExpectQuery("DELETE FROM products.*RETURNING id").
					WithArgs(receptionID).
					WillReturnError(pgx.ErrNoRows)

				dbMock.ExpectRollback()
			},
			expectedError: db.ErrNoProducts,
		},
		{
			name: "begin transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin().WillReturnError(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
		{
			name: "commit transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(receptionID))

				dbMock.ExpectQuery("DELETE FROM products.*RETURNING id").
					WithArgs(receptionID).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(productID))

				dbMock.ExpectCommit().WillReturnError(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
		{
			name: "delete product error",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(receptionID))

				dbMock.ExpectQuery("DELETE FROM products.*RETURNING id").
					WithArgs(receptionID).
					WillReturnError(ErrRandomError)

				dbMock.ExpectRollback()
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := repo.DeleteLastProduct(context.Background(), pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZRepo_CloseLastReception(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	pvzID := uuid.New()
	receptionID := uuid.New()

	now := time.Now()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError error
		expected      *models.Reception
	}{
		{
			name: "success",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id, date_time, pvz_id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id", "date_time", "pvz_id"}).
						AddRow(receptionID, now, pvzID))

				dbMock.ExpectExec("UPDATE receptions SET status =.*").
					WithArgs(string(pvzapi.Close), receptionID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				dbMock.ExpectCommit()
			},
			expected: &models.Reception{
				ID:       receptionID,
				DateTime: now,
				PvzID:    pvzID,
				Status:   string(pvzapi.Close),
			},
			expectedError: nil,
		},
		{
			name: "no open reception",
			mockSetup: func() {
				dbMock.ExpectBegin()
				dbMock.ExpectQuery("SELECT id, date_time, pvz_id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnError(pgx.ErrNoRows)
				dbMock.ExpectRollback()
			},
			expected:      nil,
			expectedError: db.ErrNoOpenReception,
		},
		{
			name: "begin transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin().WillReturnError(ErrRandomError)
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
		{
			name: "update reception error",
			mockSetup: func() {
				dbMock.ExpectBegin()
				dbMock.ExpectQuery("SELECT id, date_time, pvz_id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id", "date_time", "pvz_id"}).
						AddRow(receptionID, time.Now(), pvzID))

				dbMock.ExpectExec("UPDATE receptions SET status =.*").
					WithArgs(string(pvzapi.Close), receptionID).
					WillReturnError(ErrRandomError)

				dbMock.ExpectRollback()
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
		{
			name: "commit transaction error",
			mockSetup: func() {
				dbMock.ExpectBegin()

				dbMock.ExpectQuery("SELECT id, date_time, pvz_id FROM receptions.*FOR UPDATE").
					WithArgs(pvzID, string(pvzapi.InProgress)).
					WillReturnRows(pgxmock.NewRows([]string{"id", "date_time", "pvz_id"}).
						AddRow(receptionID, time.Now(), pvzID))

				dbMock.ExpectExec("UPDATE receptions SET status =.*").
					WithArgs(string(pvzapi.Close), receptionID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				dbMock.ExpectCommit().WillReturnError(ErrRandomError)
			},
			expected:      nil,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.CloseLastReception(context.Background(), pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZRepo_GetPVZs(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := &pvzRepo{db: dbMock}

	startDate := time.Now().Add(2 * time.Hour)
	endDate := startDate.Add(-2 * time.Hour)
	var limit uint64 = 10
	var offset uint64 = 0

	pvzID := uuid.New()
	receptionID := uuid.New()
	productID := uuid.New()

	regDate := time.Now().AddDate(-1, 0, 0)
	recDate := time.Now().Add(-2 * time.Hour)
	prodDate := time.Now().Add(-1 * time.Hour)

	status := "in_progress"
	productType := "обувь"

	tests := []struct {
		name          string
		mockSetup     func()
		expected      []*models.PVZWithReceptions
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func() {
				dbMock.ExpectQuery(regexp.QuoteMeta(`
					SELECT DISTINCT p.id FROM pvzs p 
					LEFT JOIN receptions r ON r.pvz_id = p.id 
					WHERE (r.date_time >= $1 AND r.date_time <= $2) 
					LIMIT 10 OFFSET 0
				`)).
					WithArgs(startDate, endDate).
					WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(pvzID))

				dbMock.ExpectQuery(regexp.QuoteMeta(`
					SELECT p.id, p.city, p.registration_date, 
						   r.id, r.date_time, r.status, 
						   pr.id, pr.type, pr.date_time 
					FROM pvzs p 
					LEFT JOIN receptions r ON r.pvz_id = p.id 
					LEFT JOIN products pr ON pr.reception_id = r.id 
					WHERE p.id IN ($1) AND (r.date_time >= $2 AND r.date_time <= $3) 
					ORDER BY r.date_time DESC
				`)).
					WithArgs(pvzID, startDate, endDate).
					WillReturnRows(pgxmock.NewRows([]string{
						"p.id", "p.city", "p.registration_date",
						"r.id", "r.date_time", "r.status",
						"pr.id", "pr.type", "pr.date_time",
					}).AddRow(
						pvzID, "Москва", regDate,
						&receptionID, &recDate, &status,
						&productID, &productType, &prodDate,
					))
			},
			expected: []*models.PVZWithReceptions{
				{
					PVZ: models.PVZ{
						ID:               pvzID,
						City:             "Москва",
						RegistrationDate: regDate,
					},
					Receptions: []*models.ReceptionWithProducts{
						{
							Reception: models.Reception{
								ID:       receptionID,
								DateTime: recDate,
								Status:   status,
								PvzID:    pvzID,
							},
							Products: []*models.Product{
								{
									ID:          productID,
									Type:        productType,
									DateTime:    prodDate,
									ReceptionID: receptionID,
								},
							},
						},
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			ctx := context.Background()
			result, err := repo.GetPVZs(ctx, &startDate, &endDate, limit, offset)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZRepo_GetPVZList(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer dbMock.Close()

	repo := NewPVZRepo(dbMock)

	testPVZs := []models.PVZ{
		{
			ID:               uuid.New(),
			City:             "Москва",
			RegistrationDate: time.Now(),
		},
		{
			ID:               uuid.New(),
			City:             "Казань",
			RegistrationDate: time.Now().Add(-24 * time.Hour),
		},
	}

	tests := []struct {
		name           string
		mockSetup      func()
		expectedResult []models.PVZ
		expectedError  error
	}{
		{
			name: "successful get list",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "city", "registration_date"}).
					AddRow(testPVZs[0].ID, testPVZs[0].City, testPVZs[0].RegistrationDate).
					AddRow(testPVZs[1].ID, testPVZs[1].City, testPVZs[1].RegistrationDate)
				dbMock.ExpectQuery("SELECT id, city, registration_date FROM pvzs").
					WillReturnRows(rows)
			},
			expectedResult: testPVZs,
			expectedError:  nil,
		},
		{
			name: "empty list",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "city", "registration_date"})
				dbMock.ExpectQuery("SELECT id, city, registration_date FROM pvzs").
					WillReturnRows(rows)
			},
			expectedResult: []models.PVZ(nil),
			expectedError:  nil,
		},
		{
			name: "database error",
			mockSetup: func() {
				dbMock.ExpectQuery("SELECT id, city, registration_date FROM pvzs").
					WillReturnError(ErrRandomError)
			},
			expectedResult: nil,
			expectedError:  ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := repo.GetPVZList(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
