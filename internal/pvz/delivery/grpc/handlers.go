package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cyansnbrst/pvz-service/internal/pvz"
	"github.com/cyansnbrst/pvz-service/pkg/converters"
	pvz_v1 "github.com/cyansnbrst/pvz-service/protos/gen/proto/pvz"
)

// PVZ handlers struct
type pvzHandlers struct {
	pvzUC  pvz.UseCase
	logger *zap.Logger
	pvz_v1.UnimplementedPVZServiceServer
}

// PVZ handlers constructor
func NewPVZHandlers(gRPCServer *grpc.Server, pvzUC pvz.UseCase, logger *zap.Logger) {
	pvz_v1.RegisterPVZServiceServer(gRPCServer, &pvzHandlers{pvzUC: pvzUC, logger: logger})
}

// Get all PVZs
func (h *pvzHandlers) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	pvzs, err := h.pvzUC.GetPVZList(ctx)
	if err != nil {
		h.logger.Error("failed to fetch pvzs", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to fetch pvzs")
	}

	rpvzs := make([]*pvz_v1.PVZ, 0, len(pvzs))
	for _, p := range pvzs {
		rpvzs = append(rpvzs, converters.ToProtoPVZ(p))
	}

	return &pvz_v1.GetPVZListResponse{Pvzs: rpvzs}, nil
}
