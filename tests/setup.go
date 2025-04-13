package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/suite"

	"github.com/cyansnbrst/pvz-service/config"
	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/dtos"
)

type BaseTestSuite struct {
	suite.Suite
	pool      *dockertest.Pool
	container *dockertest.Resource
	dbPool    *pgxpool.Pool
	cfg       *config.Config
}

func (s *BaseTestSuite) SetupSuite() {
	pool, err := dockertest.NewPool("")
	s.Require().NoError(err)

	err = pool.Client.Ping()
	s.Require().NoError(err)

	pool.MaxWait = 120 * time.Second
	s.pool = pool

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15-alpine3.18",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	s.Require().NoError(err)

	container.Expire(120)
	s.container = container

	hostAndPort := container.GetHostPort("5432/tcp")
	dbDSN := fmt.Sprintf("postgres://postgres:postgres@%s/test?sslmode=disable", hostAndPort)

	err = pool.Retry(func() error {
		ctx := context.Background()
		dbPool, err := pgxpool.New(ctx, dbDSN)
		if err != nil {
			return err
		}
		s.dbPool = dbPool
		return dbPool.Ping(ctx)
	})
	s.Require().NoError(err)

	s.runMigrations(dbDSN)

	envPath := filepath.ToSlash(filepath.Join("..", ".env"))
	err = godotenv.Load(envPath)
	s.Require().NoError(err)

	configPath := filepath.ToSlash(filepath.Join("..", "config", "config-local.yml"))
	s.cfg, err = config.LoadConfig(configPath)
	s.Require().NoError(err)
}

func (s *BaseTestSuite) runMigrations(dbDSN string) {
	db, err := sql.Open("pgx", dbDSN)
	s.Require().NoError(err)

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	s.Require().NoError(err)

	migrationsPath := "file://" + filepath.ToSlash(filepath.Join("..", "migrations"))
	migrate, err := migrate.NewWithDatabaseInstance(
		migrationsPath,
		"pgx",
		driver,
	)
	s.Require().NoError(err)

	err = migrate.Up()
	s.Require().NoError(err)
}

func (s *BaseTestSuite) TearDownSuite() {
	if s.dbPool != nil {
		s.dbPool.Close()
	}

	if s.container != nil {
		err := s.pool.Purge(s.container)
		if err != nil {
			log.Printf("failed to purge PostgreSQL container: %v", err)
		}
	}
}

func (s *BaseTestSuite) Login(ts *httptest.Server, role string) string {
	loginReq := pvzapi.PostDummyLoginJSONBody{
		Role: pvzapi.PostDummyLoginJSONBodyRole(role),
	}
	loginBody, err := json.Marshal(loginReq)
	s.Require().NoError(err)

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/dummyLogin", bytes.NewReader(loginBody))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Equal(http.StatusOK, resp.StatusCode)

	var authResp dtos.Token
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	s.Require().NoError(err)
	s.NotEmpty(authResp.Value)

	return authResp.Value
}
