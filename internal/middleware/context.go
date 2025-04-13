package middleware

import (
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
)

const RoleContextKey = "role"

// Set user role to the context
func ContextSetUserRole(c echo.Context, role pvzapi.UserRole) {
	c.Set(RoleContextKey, role)
}

// Get user role from the context
func ContextGetUserRole(c echo.Context) (pvzapi.UserRole, error) {
	role, ok := c.Get(RoleContextKey).(pvzapi.UserRole)
	if !ok {
		return "", errors.New("incorrect role")
	}
	return role, nil
}
