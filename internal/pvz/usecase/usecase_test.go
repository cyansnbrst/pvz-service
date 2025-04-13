package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/cyansnbrst/pvz-service/config"
	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/models"
	mock_pvz "github.com/cyansnbrst/pvz-service/internal/pvz/mock"
	"github.com/cyansnbrst/pvz-service/pkg/db"
)

var ErrRandomError = errors.New("random error")

func TestPVZUC_GenerateJWT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{
		App: config.App{
			JWTSecretKey: "secret",
			JWTTokenTTL:  time.Hour * 1,
		},
	}

	pvzUC := NewPVZUseCase(cfg, nil)

	user := &models.User{
		Role: "moderator",
	}

	token, err := pvzUC.GenerateJWT(context.Background(), pvzapi.UserRole(user.Role))
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (any, error) {
		return []byte(cfg.App.JWTSecretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
}

func TestPVZUC_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	tests := []struct {
		name          string
		email         string
		password      string
		role          string
		mockSetup     func()
		expectedUser  models.User
		expectedError error
	}{
		{
			name:     "successful registration",
			email:    "test@example.com",
			password: "password",
			role:     "employee",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, user models.User) error {
						_, err := argon2id.ComparePasswordAndHash("password", user.PasswordHash)
						assert.NoError(t, err)
						return nil
					},
				)
			},
			expectedUser: models.User{
				Email: "test@example.com",
				Role:  "employee",
			},
			expectedError: nil,
		},
		{
			name:     "duplicate email",
			email:    "test@example.com",
			password: "password",
			role:     "user",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(db.ErrDuplicateEmail)
			},
			expectedUser:  models.User{},
			expectedError: db.ErrDuplicateEmail,
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password",
			role:     "user",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(ErrRandomError)
			},
			expectedUser:  models.User{},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := pvzUC.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.role, user.Role)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.PasswordHash)
			} else {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			}
		})
	}
}

func TestPVZUC_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	correctPassword := "password123"
	hashedPassword, _ := argon2id.CreateHash(correctPassword, argon2id.DefaultParams)
	testUser := &models.User{
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         "employee",
	}

	tests := []struct {
		name          string
		email         string
		password      string
		mockSetup     func()
		expectToken   bool
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: correctPassword,
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(testUser, nil)
			},
			expectToken:   true,
			expectedError: nil,
		},
		{
			name:     "user not found",
			email:    "test@example.com",
			password: correctPassword,
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, db.ErrUserNotFound)
			},
			expectToken:   false,
			expectedError: db.ErrUserNotFound,
		},
		{
			name:     "incorrect password",
			email:    "test@example.com",
			password: "password",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(testUser, nil)
			},
			expectToken:   false,
			expectedError: ErrIncorrectPassword,
		},
		{
			name:     "repository error",
			email:    "test@example.com",
			password: "password",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, ErrRandomError)
			},
			expectToken:   false,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			token, err := pvzUC.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestPVZUC_CreatePVZ(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	testUUID := uuid.MustParse("a1b2c3d4-e5f6-7890-1234-567890abcdef")
	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		inputID       *uuid.UUID
		inputCity     string
		inputRegDate  *time.Time
		mockSetup     func()
		expectedPVZ   models.PVZ
		expectedError error
	}{
		{
			name:         "success with generated ID and time",
			inputID:      nil,
			inputCity:    "Казань",
			inputRegDate: nil,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreatePVZ(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, pvz models.PVZ) error {
						assert.NotEqual(t, uuid.Nil, pvz.ID)
						assert.Equal(t, "Казань", pvz.City)
						assert.False(t, pvz.RegistrationDate.IsZero())
						return nil
					})
			},
			expectedError: nil,
		},
		{
			name:         "success with provided ID and time",
			inputID:      &testUUID,
			inputCity:    "Казань",
			inputRegDate: &testTime,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreatePVZ(gomock.Any(), models.PVZ{
						ID:               testUUID,
						City:             "Казань",
						RegistrationDate: testTime,
					}).
					Return(nil)
			},
			expectedPVZ: models.PVZ{
				ID:               testUUID,
				City:             "Казань",
				RegistrationDate: testTime,
			},
			expectedError: nil,
		},
		{
			name:         "duplicate PVZ error",
			inputID:      &testUUID,
			inputCity:    "Казань",
			inputRegDate: &testTime,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreatePVZ(gomock.Any(), gomock.Any()).
					Return(db.ErrDuplicatePVZ)
			},
			expectedError: db.ErrDuplicatePVZ,
		},
		{
			name:         "repository error",
			inputID:      nil,
			inputCity:    "Казань",
			inputRegDate: nil,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreatePVZ(gomock.Any(), gomock.Any()).
					Return(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.CreatePVZ(context.Background(), tt.inputID, tt.inputCity, tt.inputRegDate)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, models.PVZ{}, result)
			} else {
				assert.NoError(t, err)
				if tt.inputID != nil {
					assert.Equal(t, *tt.inputID, result.ID)
				} else {
					assert.NotEqual(t, uuid.Nil, result.ID)
				}
				assert.Equal(t, tt.inputCity, result.City)
				if tt.inputRegDate != nil {
					assert.Equal(t, *tt.inputRegDate, result.RegistrationDate)
				} else {
					assert.False(t, result.RegistrationDate.IsZero())
				}
			}
		})
	}
}

func TestPVZUC_CreateReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	testPVZID := uuid.MustParse("a1b2c3d4-e5f6-7890-1234-567890abcdef")
	testReception := &models.Reception{
		ID:     uuid.New(),
		PvzID:  testPVZID,
		Status: "in_progress",
	}

	tests := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func()
		expected      models.Reception
		expectedError error
	}{
		{
			name:  "successful reception creation",
			pvzID: testPVZID,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreateReception(gomock.Any(), gomock.Any(), testPVZID).
					Return(testReception, nil)
			},
			expected:      *testReception,
			expectedError: nil,
		},
		{
			name:  "reception conflict error",
			pvzID: testPVZID,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreateReception(gomock.Any(), gomock.Any(), testPVZID).
					Return(nil, db.ErrReceptionConflict)
			},
			expected:      models.Reception{},
			expectedError: db.ErrReceptionConflict,
		},
		{
			name:  "repository error",
			pvzID: testPVZID,
			mockSetup: func() {
				mockRepo.EXPECT().
					CreateReception(gomock.Any(), gomock.Any(), testPVZID).
					Return(nil, ErrRandomError)
			},
			expected:      models.Reception{},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.CreateReception(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.Equal(t, tt.pvzID, result.PvzID)
				assert.Equal(t, "in_progress", result.Status)
			}
		})
	}
}

func TestPVZUC_AddProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	testID := uuid.MustParse("a1b2c3d4-e5f6-7890-1234-567890abcdef")
	testTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	testProduct := &models.Product{
		ID:          uuid.New(),
		Type:        "обувь",
		DateTime:    testTime,
		ReceptionID: testID,
	}

	tests := []struct {
		name          string
		pvzID         uuid.UUID
		productType   string
		mockSetup     func()
		expected      models.Product
		expectedError error
	}{
		{
			name:        "successful product addition",
			pvzID:       uuid.New(),
			productType: "обувь",
			mockSetup: func() {
				mockRepo.EXPECT().
					AddProduct(gomock.Any(), gomock.Any(), gomock.Any(), "обувь").
					Return(testProduct, nil)
			},
			expected:      *testProduct,
			expectedError: nil,
		},
		{
			name:        "no open reception error",
			pvzID:       uuid.New(),
			productType: "обувь",
			mockSetup: func() {
				mockRepo.EXPECT().
					AddProduct(gomock.Any(), gomock.Any(), gomock.Any(), "обувь").
					Return(nil, db.ErrNoOpenReception)
			},
			expected:      models.Product{},
			expectedError: db.ErrNoOpenReception,
		},
		{
			name:        "repository error",
			pvzID:       uuid.New(),
			productType: "обувь",
			mockSetup: func() {
				mockRepo.EXPECT().
					AddProduct(gomock.Any(), gomock.Any(), gomock.Any(), "обувь").
					Return(nil, ErrRandomError)
			},
			expected:      models.Product{},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.AddProduct(context.Background(), tt.pvzID, tt.productType)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, result.ID)
				assert.Equal(t, tt.productType, result.Type)
				assert.NotEqual(t, uuid.Nil, result.ReceptionID)
			}
		})
	}
}

func TestPVZUC_DeleteLastProduct(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	tests := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func()
		expectedError error
	}{
		{
			name:  "successful product deletion",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:  "no open reception error",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Any()).
					Return(db.ErrNoOpenReception)
			},
			expectedError: db.ErrNoOpenReception,
		},
		{
			name:  "no products error",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Any()).
					Return(db.ErrNoProducts)
			},
			expectedError: db.ErrNoProducts,
		},
		{
			name:  "repository error",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					DeleteLastProduct(gomock.Any(), gomock.Any()).
					Return(ErrRandomError)
			},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			err := pvzUC.DeleteLastProduct(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPVZUC_CloseLastReception(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}

	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	testReception := &models.Reception{
		ID:       uuid.New(),
		PvzID:    uuid.New(),
		Status:   "in_progress",
		DateTime: time.Now(),
	}

	tests := []struct {
		name          string
		pvzID         uuid.UUID
		mockSetup     func()
		expected      models.Reception
		expectedError error
	}{
		{
			name:  "successful reception closing",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					CloseLastReception(gomock.Any(), gomock.Any()).
					Return(testReception, nil)
			},
			expected:      *testReception,
			expectedError: nil,
		},
		{
			name:  "no open reception error",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					CloseLastReception(gomock.Any(), gomock.Any()).
					Return(nil, db.ErrNoOpenReception)
			},
			expected:      models.Reception{},
			expectedError: db.ErrNoOpenReception,
		},
		{
			name:  "repository error",
			pvzID: uuid.New(),
			mockSetup: func() {
				mockRepo.EXPECT().
					CloseLastReception(gomock.Any(), gomock.Any()).
					Return(nil, ErrRandomError)
			},
			expected:      models.Reception{},
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.CloseLastReception(context.Background(), tt.pvzID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Equal(t, tt.expected, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPVZUC_GetPVZs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{}
	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	now := time.Now()
	testTime := now.Add(-time.Hour)

	testPVZs := []*models.PVZWithReceptions{
		{
			PVZ: models.PVZ{
				ID:               uuid.New(),
				City:             "Москва",
				RegistrationDate: now.AddDate(0, -2, 0),
			},
			Receptions: []*models.ReceptionWithProducts{{
				Reception: models.Reception{
					ID:       uuid.New(),
					DateTime: testTime,
					PvzID:    uuid.New(),
					Status:   "in_progress",
				},
				Products: []*models.Product{{
					ID:          uuid.New(),
					Type:        "обувь",
					DateTime:    testTime,
					ReceptionID: uuid.New(),
				}},
			}},
		},
		{
			PVZ: models.PVZ{
				ID:               uuid.New(),
				City:             "Москва",
				RegistrationDate: now.AddDate(0, -1, 0),
			},
			Receptions: []*models.ReceptionWithProducts{{
				Reception: models.Reception{
					ID:       uuid.New(),
					DateTime: testTime.Add(-time.Hour),
					PvzID:    uuid.New(),
					Status:   "in_progress",
				},
				Products: []*models.Product{{
					ID:          uuid.New(),
					Type:        "обувь",
					DateTime:    testTime.Add(-time.Hour),
					ReceptionID: uuid.New(),
				}},
			}},
		},
		{
			PVZ: models.PVZ{
				ID:               uuid.New(),
				City:             "Москва",
				RegistrationDate: now,
			},
			Receptions: []*models.ReceptionWithProducts{{
				Reception: models.Reception{
					ID:       uuid.New(),
					DateTime: testTime.Add(-2 * time.Hour),
					PvzID:    uuid.New(),
					Status:   "in_progress",
				},
				Products: []*models.Product{{
					ID:          uuid.New(),
					Type:        "обувь",
					DateTime:    testTime.Add(-2 * time.Hour),
					ReceptionID: uuid.New(),
				}},
			}},
		},
	}

	tests := []struct {
		name          string
		params        pvzapi.GetPvzParams
		mockSetup     func()
		expectedCount int
		expectedError error
	}{
		{
			name: "default pagination",
			params: pvzapi.GetPvzParams{
				Page:  nil,
				Limit: nil,
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(10), uint64(0)).
					Return(testPVZs[:2], nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "custom pagination - second page",
			params: pvzapi.GetPvzParams{
				Page:  ptrToInt(2),
				Limit: ptrToInt(2),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(2), uint64(2)).
					Return(testPVZs[2:], nil)
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name: "limit exceeds total count",
			params: pvzapi.GetPvzParams{
				Page:  ptrToInt(1),
				Limit: ptrToInt(5),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(5), uint64(0)).
					Return(testPVZs, nil)
			},
			expectedCount: 3,
			expectedError: nil,
		},
		{
			name: "page correction",
			params: pvzapi.GetPvzParams{
				Page:  ptrToInt(-1),
				Limit: ptrToInt(5),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(5), uint64(0)).
					Return(testPVZs[:2], nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "limit correction",
			params: pvzapi.GetPvzParams{
				Page:  ptrToInt(2),
				Limit: ptrToInt(0),
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(10), uint64(10)).
					Return(testPVZs[2:], nil)
			},
			expectedCount: 1,
			expectedError: nil,
		},
		{
			name: "invalid date range",
			params: pvzapi.GetPvzParams{
				StartDate: &now,
				EndDate:   &testTime,
			},
			mockSetup:     func() {},
			expectedCount: 0,
			expectedError: ErrInvalidDateRange,
		},
		{
			name: "repository error",
			params: pvzapi.GetPvzParams{
				Page:  nil,
				Limit: nil,
			},
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZs(gomock.Any(), nil, nil, uint64(10), uint64(0)).
					Return(nil, ErrRandomError)
			},
			expectedCount: 0,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.GetPVZs(context.Background(), tt.params)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func TestPVZUC_GetPVZList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &config.Config{App: config.App{}}
	mockRepo := mock_pvz.NewMockRepository(ctrl)
	pvzUC := NewPVZUseCase(cfg, mockRepo)

	testPVZs := []models.PVZ{
		{
			ID:               uuid.New(),
			City:             "Москва",
			RegistrationDate: time.Now(),
		},
		{
			ID:               uuid.New(),
			City:             "Москва",
			RegistrationDate: time.Now(),
		},
	}

	tests := []struct {
		name          string
		mockSetup     func()
		expectedCount int
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZList(gomock.Any()).
					Return(testPVZs, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "repository error",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetPVZList(gomock.Any()).
					Return(nil, ErrRandomError)
			},
			expectedCount: 0,
			expectedError: ErrRandomError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			result, err := pvzUC.GetPVZList(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}
		})
	}
}

func ptrToInt(i int) *int { return &i }
