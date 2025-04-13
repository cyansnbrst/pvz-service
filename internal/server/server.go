package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/cyansnbrst/pvz-service/config"
)

// Server struct
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	db         *pgxpool.Pool
	httpServer *http.Server
	grpcServer *grpc.Server
}

// New server constructor
func NewServer(cfg *config.Config, logger *zap.Logger, db *pgxpool.Pool) *Server {
	return &Server{
		config: cfg,
		logger: logger,
		db:     db,
	}
}

// Run server
func (s *Server) Run() error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.App.HTTPPort),
		Handler:      s.RegisterHandlers(),
		IdleTimeout:  s.config.App.IdleTimeout,
		ReadTimeout:  s.config.App.ReadTimeout,
		WriteTimeout: s.config.App.WriteTimeout,
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.App.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen gRPC: %w", err)
	}

	s.grpcServer = grpc.NewServer()
	s.RegisterServices()

	shutDownError := make(chan error, 2)

	go func() {
		s.logger.Info("starting HTTP server",
			zap.String("addr", s.httpServer.Addr),
		)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			shutDownError <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	go func() {
		s.logger.Info("starting gRPC server",
			zap.String("addr", lis.Addr().String()),
		)
		if err := s.grpcServer.Serve(lis); err != nil {
			shutDownError <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit

		s.logger.Info("shutting down servers",
			zap.String("signal", sig.String()),
		)

		ctx, cancel := context.WithTimeout(context.Background(), s.config.App.ShutdownTimeout)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			shutDownError <- fmt.Errorf("HTTP shutdown error: %w", err)
		}

		s.grpcServer.GracefulStop()

		shutDownError <- nil
	}()

	err = <-shutDownError
	if err != nil {
		return err
	}

	s.logger.Info("servers stopped successfully")

	return nil
}
