package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config struct
type Config struct {
	App        App        `yaml:"app"`
	PostgreSQL PostgreSQL `yaml:"postgres"`
	Metrics    Metrics    `yaml:"metrics"`
}

// App config struct
type App struct {
	HTTPPort        int64         `yaml:"http_port" env:"APP_HTTP_PORT" env-required:"true"`
	GRPCPort        int64         `yaml:"grpc_port" env:"APP_GRPC_PORT" env-required:"true"`
	Env             string        `yaml:"env" env:"APP_ENV" env-required:"true"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env:"APP_IDLE_TIMEOUT" env-required:"true"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env:"APP_READ_TIMEOUT" env-required:"true"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env:"APP_WRITE_TIMEOUT" env-required:"true"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"APP_SHUTDOWN_TIMEOUT" env-required:"true"`
	JWTTokenTTL     time.Duration `yaml:"jwt_token_ttl" env:"JWT_TOKEN_TTL" env-required:"true"`
	JWTSecretKey    string        `env:"JWT_SECRET_KEY" env-required:"true"`
}

// PostgreSQL config struct
type PostgreSQL struct {
	Host        string        `env:"POSTGRES_DOCKER_HOST" env-required:"true"`
	Port        int64         `env:"POSTGRES_PORT" env-required:"true"`
	User        string        `env:"POSTGRES_USER" env-required:"true"`
	Password    string        `env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName      string        `env:"POSTGRES_DB" env-required:"true"`
	SSLMode     string        `env:"POSTGRES_SSLMODE" env-required:"true"`
	MaxPoolSize int32         `yaml:"max_pool_size" env:"POSTGRES_MAX_POOL_SIZE" env-required:"true"`
	ConnTimeout time.Duration `yaml:"conn_timeout" env:"POSTGRES_CONN_TIMEOUT" env-required:"true"`
	Driver      string        `yaml:"driver" env:"POSTGRES_DRIVER" env-required:"true"`
}

// Metrics config struct
type Metrics struct {
	URL         string `env:"METRICS_PORT"`
	ServiceName string `env:"METRICS_SERVICE_NAME"`
}

// Load config file from given path and env variables
func LoadConfig(filename string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(filepath.Join(".", filename), &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read env variables: %w", err)
	}

	return &cfg, nil
}
