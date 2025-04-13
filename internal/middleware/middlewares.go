package middleware

import (
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/config"
)

// Middleware manager struct
type Manager struct {
	cfg    *config.Config
	logger *zap.Logger
}

// Middleware manager constructor
func NewManager(cfg *config.Config, logger *zap.Logger) *Manager {
	return &Manager{
		cfg:    cfg,
		logger: logger,
	}
}
