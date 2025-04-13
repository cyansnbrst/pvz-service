package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/dtos"
	"github.com/cyansnbrst/pvz-service/internal/server"
)

type HandlersTestSuite struct {
	BaseTestSuite
}

func TestHandlersSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (s *HandlersTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
}

func (s *HandlersTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *HandlersTestSuite) TestPostDummyLogin() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())

	tests := []struct {
		name           string
		payload        any
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful login",
			payload: pvzapi.PostDummyLoginJSONRequestBody{
				Role: pvzapi.PostDummyLoginJSONBodyRoleEmployee,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid role",
			payload: pvzapi.PostDummyLoginJSONRequestBody{
				Role: "admin",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "missing role",
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			body, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/dummyLogin", ts.URL),
				bytes.NewReader(body),
			)
			s.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				s.NoError(err)
				return
			}

			var tokenResp dtos.Token
			err = json.NewDecoder(resp.Body).Decode(&tokenResp)
			s.NoError(err)
			s.NotEmpty(tokenResp.Value)
		})
	}
}

func (s *HandlersTestSuite) TestPostRegister() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())

	tests := []struct {
		name           string
		payload        any
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successfull registration",
			payload: pvzapi.PostRegisterJSONRequestBody{
				Email:    "moderator@test.com",
				Password: "secure123",
				Role:     pvzapi.Moderator,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "duplicate email",
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)",
					"test@test.com",
					"password",
					"employee",
				)

				s.Require().NoError(err)
			},
			payload: pvzapi.PostRegisterJSONRequestBody{
				Email:    "test@test.com",
				Password: "secure123",
				Role:     "employee",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "invalid role",
			payload: pvzapi.PostRegisterJSONRequestBody{
				Email:    "test@test.com",
				Password: "secure123",
				Role:     "role",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "missing email",
			payload: map[string]any{
				"password": "secure123",
				"role":     "employee",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.dbPool.Exec(context.Background(), "DELETE FROM users")
			s.Require().NoError(err)

			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			body, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/register", ts.URL),
				bytes.NewReader(body),
			)
			s.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				s.NoError(err)
				return
			}

			var userResp pvzapi.User
			err = json.NewDecoder(resp.Body).Decode(&userResp)
			s.NoError(err)
			s.NotEmpty(userResp.Id)
		})
	}
}

func (s *HandlersTestSuite) TestPostLogin() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())

	testPassword := "password"
	hashedPassword, err := argon2id.CreateHash(testPassword, argon2id.DefaultParams)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(
		context.Background(),
		"INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)",
		"test@test.com",
		string(hashedPassword),
		pvzapi.Moderator,
	)
	s.Require().NoError(err)

	tests := []struct {
		name           string
		payload        any
		expectedStatus int
		wantErr        bool
	}{
		{
			name: "successful login",
			payload: pvzapi.PostLoginJSONRequestBody{
				Email:    "test@test.com",
				Password: testPassword,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid email",
			payload: pvzapi.PostLoginJSONRequestBody{
				Email:    "adfasdfsa@test.com",
				Password: testPassword,
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "incorrect password",
			payload: pvzapi.PostLoginJSONRequestBody{
				Email:    "test@test.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name: "missing password",
			payload: map[string]any{
				"email": "test@test.com",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			body, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/login", ts.URL),
				bytes.NewReader(body),
			)
			s.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				s.NoError(err)
				return
			}

			var tokenResp map[string]string
			err = json.NewDecoder(resp.Body).Decode(&tokenResp)
			s.NoError(err)
			s.NotEmpty(tokenResp["token"])
		})
	}
}

func (s *HandlersTestSuite) TestPostPvz() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())

	moderatorToken := s.Login(ts, "moderator")
	employeeToken := s.Login(ts, "employee")

	pvzID := uuid.New()

	tests := []struct {
		name           string
		token          string
		payload        any
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name:  "successful pvz creation by moderator",
			token: moderatorToken,
			payload: map[string]any{
				"city": "Москва",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "access denied for other user",
			token: employeeToken,
			payload: map[string]any{
				"city": "Москва",
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name:           "missing city",
			token:          moderatorToken,
			payload:        map[string]any{},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "duplicate pvz",
			token: moderatorToken,
			payload: map[string]any{
				"id":   pvzID,
				"city": "Москва",
			},
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
					pvzID, "Москва")
				s.Require().NoError(err)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "invalid city",
			token: moderatorToken,
			payload: map[string]any{
				"city": "Астана",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			token:          moderatorToken,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.dbPool.Exec(context.Background(), "DELETE FROM pvzs")
			s.Require().NoError(err)

			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			body, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/pvz", ts.URL),
				bytes.NewReader(body),
			)
			s.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")

			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				s.NoError(err)
				return
			}

			var pvzResp pvzapi.PVZ
			err = json.NewDecoder(resp.Body).Decode(&pvzResp)
			s.NoError(err)
			s.NotEmpty(pvzResp.Id)
			if p, ok := tt.payload.(pvzapi.PostPvzJSONRequestBody); ok {
				s.Equal(*p.Id, pvzResp.Id)
				s.Equal(p.City, pvzResp.City)
			}
		})
	}
}

func (s *HandlersTestSuite) TestPostReceptions() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	moderatorToken := s.Login(ts, "moderator")
	employeeToken := s.Login(ts, "employee")

	pvzID := uuid.New()
	_, err := s.dbPool.Exec(context.Background(),
		"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
		pvzID, "Москва")
	s.Require().NoError(err)

	tests := []struct {
		name           string
		token          string
		payload        any
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name:  "successful reception creation",
			token: employeeToken,
			payload: pvzapi.PostReceptionsJSONRequestBody{
				PvzId: pvzID,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "access denied for other user",
			token: moderatorToken,
			payload: pvzapi.PostReceptionsJSONRequestBody{
				PvzId: pvzID,
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name:  "missing id",
			token: employeeToken,
			payload: pvzapi.PostReceptionsJSONRequestBody{
				PvzId: uuid.Nil,
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "reception conflict",
			token: employeeToken,
			payload: pvzapi.PostReceptionsJSONRequestBody{
				PvzId: pvzID,
			},
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"INSERT INTO receptions (pvz_id) VALUES ($1)",
					pvzID)
				s.Require().NoError(err)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			token:          employeeToken,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.dbPool.Exec(context.Background(), "DELETE FROM receptions")
			s.Require().NoError(err)

			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest(
				http.MethodPost,
				ts.URL+"/receptions",
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if !tt.wantErr {
				var reception pvzapi.Reception
				s.NoError(json.NewDecoder(resp.Body).Decode(&reception))
				s.Equal(pvzID, reception.PvzId)
			}
		})
	}
}

func (s *HandlersTestSuite) TestInvalidToken() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	token := "Bearer token"

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "invalid token structure",
			token:          token,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("%s/pvz", ts.URL)

			req, _ := http.NewRequest(http.MethodGet, url, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)
		})
	}
}

func (s *HandlersTestSuite) TestPostProducts() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	moderatorToken := s.Login(ts, "moderator")
	employeeToken := s.Login(ts, "employee")

	pvzID := uuid.New()

	_, err := s.dbPool.Exec(context.Background(),
		"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
		pvzID, "Москва")
	s.Require().NoError(err)

	tests := []struct {
		name           string
		token          string
		payload        any
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name:  "successful product addition",
			token: employeeToken,
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"INSERT INTO receptions (pvz_id) VALUES ($1)",
					pvzID)
				s.Require().NoError(err)
			},
			payload: pvzapi.PostProductsJSONRequestBody{
				PvzId: pvzID,
				Type:  pvzapi.PostProductsJSONBodyTypeОдежда,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:  "access denied for moderator",
			token: moderatorToken,
			payload: map[string]any{
				"pvz_id": pvzID,
				"type":   pvzapi.PostProductsJSONBodyTypeОдежда,
			},
			expectedStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name:  "missing pvz_id",
			token: employeeToken,
			payload: map[string]any{
				"type": pvzapi.PostProductsJSONBodyTypeОдежда,
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "invalid product type",
			token: employeeToken,
			payload: pvzapi.PostProductsJSONRequestBody{
				PvzId: pvzID,
				Type:  "косметика",
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "no open reception",
			token: employeeToken,
			payload: pvzapi.PostProductsJSONRequestBody{
				PvzId: pvzID,
				Type:  pvzapi.PostProductsJSONBodyTypeОдежда,
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid json",
			token:          employeeToken,
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.dbPool.Exec(context.Background(), "DELETE FROM products")
			s.Require().NoError(err)
			_, err = s.dbPool.Exec(context.Background(), "DELETE FROM receptions")
			s.Require().NoError(err)

			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			body, err := json.Marshal(tt.payload)
			s.Require().NoError(err)

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/products", ts.URL),
				bytes.NewReader(body),
			)
			s.Require().NoError(err)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				err = json.NewDecoder(resp.Body).Decode(&errResp)
				s.NoError(err)
			} else {
				var product pvzapi.Product
				err = json.NewDecoder(resp.Body).Decode(&product)
				s.NoError(err)
				s.NotEmpty(product.Id)
			}
		})
	}
}

func (s *HandlersTestSuite) TestPostPvzPvzIdDeleteLastProduct() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	moderatorToken := s.Login(ts, "moderator")
	employeeToken := s.Login(ts, "employee")

	pvzID := uuid.New()

	_, err := s.dbPool.Exec(context.Background(),
		"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
		pvzID, "Москва")
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		"INSERT INTO receptions (pvz_id) VALUES ($1)",
		pvzID)
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		"INSERT INTO products (reception_id, type) VALUES ((SELECT id FROM receptions WHERE pvz_id = $1), $2)",
		pvzID, "обувь")
	s.Require().NoError(err)

	tests := []struct {
		name           string
		token          string
		pvzId          uuid.UUID
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "successful product deletion",
			token:          employeeToken,
			pvzId:          pvzID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "access denied",
			token:          moderatorToken,
			pvzId:          pvzID,
			expectedStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name:           "no open reception",
			token:          employeeToken,
			pvzId:          uuid.New(),
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "no products in reception",
			token: employeeToken,
			pvzId: pvzID,
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"DELETE FROM products WHERE reception_id = (SELECT id FROM receptions WHERE pvz_id = $1)",
					pvzID)
				s.Require().NoError(err)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/pvz/%s/delete_last_product", ts.URL, tt.pvzId),
				nil,
			)
			s.Require().NoError(err)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				s.NoError(json.NewDecoder(resp.Body).Decode(&errResp))
			} else {
				var count int
				err := s.dbPool.QueryRow(context.Background(),
					"SELECT COUNT(*) FROM products WHERE reception_id = (SELECT id FROM receptions WHERE pvz_id = $1)",
					pvzID).Scan(&count)
				s.Require().NoError(err)
				s.Equal(0, count)
			}
		})
	}
}

func (s *HandlersTestSuite) TestPostPvzPvzIdCloseLastReception() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	moderatorToken := s.Login(ts, "moderator")
	employeeToken := s.Login(ts, "employee")

	pvzID := uuid.New()

	_, err := s.dbPool.Exec(context.Background(),
		"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
		pvzID, "Москва")
	s.Require().NoError(err)

	_, err = s.dbPool.Exec(context.Background(),
		"INSERT INTO receptions (pvz_id) VALUES ($1)",
		pvzID)
	s.Require().NoError(err)

	tests := []struct {
		name           string
		token          string
		pvzId          uuid.UUID
		prepareDB      func()
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "successful reception closing",
			token:          employeeToken,
			pvzId:          pvzID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "access denied for moderator",
			token:          moderatorToken,
			pvzId:          pvzID,
			expectedStatus: http.StatusForbidden,
			wantErr:        true,
		},
		{
			name:           "no open reception",
			token:          employeeToken,
			pvzId:          uuid.New(),
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "already closed reception",
			token: employeeToken,
			pvzId: pvzID,
			prepareDB: func() {
				_, err := s.dbPool.Exec(context.Background(),
					"UPDATE receptions SET status = 'close' WHERE pvz_id = $1",
					pvzID)
				s.Require().NoError(err)
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			_, err := s.dbPool.Exec(context.Background(),
				"UPDATE receptions SET status = 'in_progress' WHERE pvz_id = $1", pvzID)
			s.Require().NoError(err)

			if tt.prepareDB != nil {
				tt.prepareDB()
			}

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/pvz/%s/close_last_reception", ts.URL, tt.pvzId),
				nil,
			)
			s.Require().NoError(err)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				s.NoError(json.NewDecoder(resp.Body).Decode(&errResp))
			} else {
				var reception pvzapi.Reception
				s.NoError(json.NewDecoder(resp.Body).Decode(&reception))
				s.Equal(pvzapi.Close, reception.Status)
			}
		})
	}
}

func (s *HandlersTestSuite) TestGetPvz() {
	app := server.NewServer(s.cfg, zap.NewNop(), s.dbPool)
	ts := httptest.NewServer(app.RegisterHandlers())
	defer ts.Close()

	employeeToken := s.Login(ts, "employee")

	now := time.Now()
	dates := []time.Time{
		now.Add(-72 * time.Hour),
		now.Add(-48 * time.Hour),
		now.Add(-24 * time.Hour),
		now,
		now.Add(24 * time.Hour),
	}

	pvzIDs := make([]uuid.UUID, 5)
	for i := 0; i < 5; i++ {
		pvzIDs[i] = uuid.New()
		_, err := s.dbPool.Exec(context.Background(),
			"INSERT INTO pvzs (id, city) VALUES ($1, $2)",
			pvzIDs[i], "Москва")
		s.Require().NoError(err)
	}

	for i, pvzID := range pvzIDs {
		_, err := s.dbPool.Exec(context.Background(),
			`INSERT INTO receptions (pvz_id, date_time) 
			 VALUES ($1, $2)`,
			pvzID, dates[i])
		s.Require().NoError(err)
	}

	tests := []struct {
		name           string
		token          string
		queryParams    url.Values
		expectedCount  int
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "no filters",
			token:          employeeToken,
			expectedCount:  5,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "filter by date range",
			token: employeeToken,
			queryParams: url.Values{
				"startDate": []string{dates[1].Format(time.RFC3339)},
				"endDate":   []string{dates[3].Format(time.RFC3339)},
			},
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "invalid date range",
			token: employeeToken,
			queryParams: url.Values{
				"startDate": []string{dates[3].Format(time.RFC3339)},
				"endDate":   []string{dates[1].Format(time.RFC3339)},
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:  "pagination - first page",
			token: employeeToken,
			queryParams: url.Values{
				"page":  []string{"1"},
				"limit": []string{"2"},
			},
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "pagination - second page",
			token: employeeToken,
			queryParams: url.Values{
				"page":  []string{"2"},
				"limit": []string{"2"},
			},
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "pagination - last page",
			token: employeeToken,
			queryParams: url.Values{
				"page":  []string{"3"},
				"limit": []string{"2"},
			},
			expectedCount:  1,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "combined filters",
			token: employeeToken,
			queryParams: url.Values{
				"startDate": []string{dates[0].Format(time.RFC3339)},
				"endDate":   []string{dates[2].Format(time.RFC3339)},
				"page":      []string{"1"},
				"limit":     []string{"2"},
			},
			expectedCount:  2,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := fmt.Sprintf("%s/pvz", ts.URL)
			if len(tt.queryParams) > 0 {
				url += "?" + tt.queryParams.Encode()
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			s.Require().NoError(err)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := http.DefaultClient.Do(req)
			s.Require().NoError(err)
			defer resp.Body.Close()

			s.Equal(tt.expectedStatus, resp.StatusCode)

			if tt.wantErr {
				var errResp pvzapi.Error
				s.NoError(json.NewDecoder(resp.Body).Decode(&errResp))
				return
			}

			var pvzs []dtos.PVZWithReceptions
			s.NoError(json.NewDecoder(resp.Body).Decode(&pvzs))
			s.Len(pvzs, tt.expectedCount)
		})
	}
}
