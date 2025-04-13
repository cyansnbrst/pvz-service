package server

import (
	grpcapp "github.com/cyansnbrst/pvz-service/internal/pvz/delivery/grpc"
	"github.com/cyansnbrst/pvz-service/internal/pvz/repository"
	"github.com/cyansnbrst/pvz-service/internal/pvz/usecase"
)

// Register server services
func (s *Server) RegisterServices() {
	pvzRepo := repository.NewPVZRepo(s.db)
	pvzUC := usecase.NewPVZUseCase(s.config, pvzRepo)
	grpcapp.NewPVZHandlers(s.grpcServer, pvzUC, s.logger)
}
