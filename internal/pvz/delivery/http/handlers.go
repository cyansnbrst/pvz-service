package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/internal/dtos"
	"github.com/cyansnbrst/pvz-service/internal/middleware"
	"github.com/cyansnbrst/pvz-service/internal/pvz"
	"github.com/cyansnbrst/pvz-service/internal/pvz/usecase"
	"github.com/cyansnbrst/pvz-service/pkg/converters"
	"github.com/cyansnbrst/pvz-service/pkg/db"
	hh "github.com/cyansnbrst/pvz-service/pkg/http_helpers"
	"github.com/cyansnbrst/pvz-service/pkg/metric"
)

// PVZ handlers struct
type pvzHandlers struct {
	pvzUC   pvz.UseCase
	logger  *zap.Logger
	metrics metric.Metrics
}

// PVZ handlers constructor
func NewPVZHandlers(pvzUC pvz.UseCase, logger *zap.Logger, metrics metric.Metrics) pvzapi.ServerInterface {
	return &pvzHandlers{
		pvzUC:   pvzUC,
		logger:  logger,
		metrics: metrics,
	}
}

// Gives a JWT token for the specified role
func (h *pvzHandlers) PostDummyLogin(c echo.Context) error {
	var req pvzapi.PostDummyLoginJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.Role == "" {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	if req.Role != pvzapi.PostDummyLoginJSONBodyRoleEmployee && req.Role != pvzapi.PostDummyLoginJSONBodyRoleModerator {
		return hh.BadRequestResponse(c, usecase.ErrInvalidRole)
	}

	tokenStr, err := h.pvzUC.GenerateJWT(c.Request().Context(), pvzapi.UserRole(req.Role))
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	resp := &dtos.Token{Value: tokenStr}

	return c.JSON(http.StatusOK, resp)
}

// Register user with the desired role
func (h *pvzHandlers) PostRegister(c echo.Context) error {
	var req pvzapi.PostRegisterJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.Email == "" || req.Password == "" || req.Role == "" {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	if req.Role != pvzapi.Employee && req.Role != pvzapi.Moderator {
		return hh.BadRequestResponse(c, usecase.ErrInvalidRole)
	}

	user, err := h.pvzUC.Register(c.Request().Context(), string(req.Email), req.Password, string(req.Role))
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	resp := converters.ToResponseUser(user)

	return c.JSON(http.StatusCreated, resp)
}

// Login user
func (h *pvzHandlers) PostLogin(c echo.Context) error {
	var req pvzapi.PostLoginJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.Email == "" || req.Password == "" {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	tokenStr, err := h.pvzUC.Login(c.Request().Context(), string(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, db.ErrUserNotFound) || errors.Is(err, usecase.ErrIncorrectPassword) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": tokenStr})
}

// Create a new PVZ (moderator only)
func (h *pvzHandlers) PostPvz(c echo.Context) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleModerator {
		return hh.AccessDeniedResponse(c)
	}

	var req pvzapi.PostPvzJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.City == "" {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	allowedCities := map[pvzapi.PVZCity]bool{
		pvzapi.Казань:         true,
		pvzapi.Москва:         true,
		pvzapi.СанктПетербург: true,
	}

	if !allowedCities[req.City] {
		return hh.BadRequestResponse(c, usecase.ErrInvalidCity)
	}

	pvz, err := h.pvzUC.CreatePVZ(c.Request().Context(), req.Id, string(req.City), req.RegistrationDate)
	if err != nil {
		if errors.Is(err, db.ErrDuplicatePVZ) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if h.metrics != nil {
		h.metrics.IncPVZCreated()
	}

	resp := converters.ToResponsePVZ(pvz)

	return c.JSON(http.StatusCreated, resp)
}

// Create a new reception (employee only)
func (h *pvzHandlers) PostReceptions(c echo.Context) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleEmployee {
		return hh.AccessDeniedResponse(c)
	}

	var req pvzapi.PostReceptionsJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.PvzId == uuid.Nil {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	reception, err := h.pvzUC.CreateReception(c.Request().Context(), req.PvzId)
	if err != nil {
		if errors.Is(err, db.ErrReceptionConflict) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if h.metrics != nil {
		h.metrics.IncReceptionsCreated()
	}

	resp := converters.ToResponseReception(reception)

	return c.JSON(http.StatusCreated, resp)
}

// Add a product to the reception (employee only)
func (h *pvzHandlers) PostProducts(c echo.Context) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleEmployee {
		return hh.AccessDeniedResponse(c)
	}

	var req pvzapi.PostProductsJSONRequestBody

	if err := c.Bind(&req); err != nil {
		return hh.BadRequestResponse(c, err)
	}

	if req.PvzId == uuid.Nil || req.Type == "" {
		return hh.BadRequestResponse(c, fmt.Errorf("missing field(s)"))
	}

	allowedTypes := map[pvzapi.PostProductsJSONBodyType]bool{
		pvzapi.PostProductsJSONBodyTypeОбувь:       true,
		pvzapi.PostProductsJSONBodyTypeОдежда:      true,
		pvzapi.PostProductsJSONBodyTypeЭлектроника: true,
	}

	if !allowedTypes[req.Type] {
		return hh.BadRequestResponse(c, usecase.ErrInvalidType)
	}

	product, err := h.pvzUC.AddProduct(c.Request().Context(), req.PvzId, string(req.Type))
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if h.metrics != nil {
		h.metrics.IncProductsAdded()
	}

	resp := converters.ToResponseProduct(product)

	return c.JSON(http.StatusCreated, resp)
}

// Delete last product from the reception
func (h *pvzHandlers) PostPvzPvzIdDeleteLastProduct(c echo.Context, uuid openapi_types.UUID) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleEmployee {
		return hh.AccessDeniedResponse(c)
	}

	err = h.pvzUC.DeleteLastProduct(c.Request().Context(), uuid)
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) || errors.Is(err, db.ErrNoProducts) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	return c.NoContent(http.StatusOK)
}

// Close last reception for the pvz
func (h *pvzHandlers) PostPvzPvzIdCloseLastReception(c echo.Context, uuid openapi_types.UUID) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleEmployee {
		return hh.AccessDeniedResponse(c)
	}

	reception, err := h.pvzUC.CloseLastReception(c.Request().Context(), uuid)
	if err != nil {
		if errors.Is(err, db.ErrNoOpenReception) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	resp := converters.ToResponseReception(reception)

	return c.JSON(http.StatusOK, resp)
}

// Get a list of pvzs
func (h *pvzHandlers) GetPvz(c echo.Context, params pvzapi.GetPvzParams) error {
	role, err := middleware.ContextGetUserRole(c)
	if err != nil {
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	if role != pvzapi.UserRoleEmployee && role != pvzapi.UserRoleModerator {
		return hh.AccessDeniedResponse(c)
	}

	pvzs, err := h.pvzUC.GetPVZs(c.Request().Context(), params)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidDateRange) {
			return hh.BadRequestResponse(c, err)
		}
		return hh.ServerErrorResponse(c, h.logger, err)
	}

	resp := make([]dtos.PVZWithReceptions, len(pvzs))
	for i, pvz := range pvzs {
		resp[i] = converters.ToResponsePVZWithReceptions(pvz)
	}

	return c.JSON(http.StatusOK, resp)
}
