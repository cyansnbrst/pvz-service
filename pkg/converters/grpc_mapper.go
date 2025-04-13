package converters

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cyansnbrst/pvz-service/internal/models"
	pvz_v1 "github.com/cyansnbrst/pvz-service/protos/gen/proto/pvz"
)

// PVZ model to PVZ proto
func ToProtoPVZ(m models.PVZ) *pvz_v1.PVZ {
	return &pvz_v1.PVZ{
		Id:               m.ID.String(),
		RegistrationDate: timestamppb.New(m.RegistrationDate),
		City:             m.City,
	}
}
