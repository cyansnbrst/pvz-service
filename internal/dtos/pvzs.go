package dtos

import "github.com/cyansnbrst/pvz-service/gen/pvzapi"

// PVZ with receptions response struct
type PVZWithReceptions struct {
	PVZ        pvzapi.PVZ              `json:"pvz"`
	Receptions []ReceptionWithProducts `json:"receptions"`
}

// Reception with products response struct
type ReceptionWithProducts struct {
	Reception pvzapi.Reception `json:"reception"`
	Products  []pvzapi.Product `json:"products"`
}
