package converters

import (
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/dtos"
	"github.com/cyansnbrst/pvz-service/internal/models"
)

// User model to user response
func ToResponseUser(m models.User) pvzapi.User {
	return pvzapi.User{
		Email: openapi_types.Email(m.Email),
		Id:    &m.ID,
		Role:  pvzapi.UserRole(m.Role),
	}
}

// PVZ model to PVZ response
func ToResponsePVZ(m models.PVZ) pvzapi.PVZ {
	return pvzapi.PVZ{
		Id:               &m.ID,
		City:             pvzapi.PVZCity(m.City),
		RegistrationDate: &m.RegistrationDate,
	}
}

// Reception model to reception response
func ToResponseReception(m models.Reception) pvzapi.Reception {
	return pvzapi.Reception{
		Id:       &m.ID,
		DateTime: m.DateTime,
		PvzId:    m.PvzID,
		Status:   pvzapi.ReceptionStatus(m.Status),
	}
}

// Product model to product response
func ToResponseProduct(m models.Product) pvzapi.Product {
	return pvzapi.Product{
		Id:          &m.ID,
		DateTime:    &m.DateTime,
		ReceptionId: m.ReceptionID,
		Type:        pvzapi.ProductType(m.Type),
	}
}

// PVZ with receptions model to PVZ with receptions response
func ToResponsePVZWithReceptions(m *models.PVZWithReceptions) dtos.PVZWithReceptions {
	pvz := ToResponsePVZ(m.PVZ)

	receptions := make([]dtos.ReceptionWithProducts, len(m.Receptions))
	for i, r := range m.Receptions {
		reception := ToResponseReception(r.Reception)

		products := make([]pvzapi.Product, len(r.Products))
		for j, p := range r.Products {
			products[j] = ToResponseProduct(*p)
		}

		receptions[i] = dtos.ReceptionWithProducts{
			Reception: reception,
			Products:  products,
		}
	}

	return dtos.PVZWithReceptions{
		PVZ:        pvz,
		Receptions: receptions,
	}
}
