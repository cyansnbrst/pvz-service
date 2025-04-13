package server

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	mm "github.com/cyansnbrst/pvz-service/internal/middleware"
	"github.com/cyansnbrst/pvz-service/internal/pvz/delivery/http"
	"github.com/cyansnbrst/pvz-service/internal/pvz/repository"
	"github.com/cyansnbrst/pvz-service/internal/pvz/usecase"
	"github.com/cyansnbrst/pvz-service/pkg/metric"
)

// Register server handlers
func (s *Server) RegisterHandlers() *echo.Echo {
	metrics, err := metric.CreateMetrics(s.config.Metrics.URL, s.config.Metrics.ServiceName)
	if err != nil {
		s.logger.Info("CreateMetrics", zap.Error(err))
	}
	s.logger.Info("metrics server started",
		zap.String("available URL", s.config.Metrics.URL),
		zap.String("service name", s.config.Metrics.ServiceName),
	)

	e := echo.New()

	e.Use(middleware.Recover())

	pvzRepo := repository.NewPVZRepo(s.db)
	pvzUC := usecase.NewPVZUseCase(s.config, pvzRepo)
	pvzHandlers := http.NewPVZHandlers(pvzUC, s.logger, metrics)

	mw := mm.NewManager(s.config, s.logger)
	e.Use(mw.Authenticate)
	e.Use(mw.MetricsMiddleware(metrics))

	pvzapi.RegisterHandlers(e, pvzHandlers)

	return e
}
