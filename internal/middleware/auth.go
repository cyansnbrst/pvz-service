package middleware

import (
	"errors"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/cyansnbrst/pvz-service/pkg/auth"
	"github.com/cyansnbrst/pvz-service/pkg/auth/jwt"
	hh "github.com/cyansnbrst/pvz-service/pkg/http_helpers"
)

// Authentication middleware
func (mw *Manager) Authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	excludedPaths := map[string]bool{
		"/login":      true,
		"/register":   true,
		"/dummyLogin": true,
	}

	return func(c echo.Context) error {
		currentPath := c.Path()

		if excludedPaths[currentPath] {
			return next(c)
		}

		authHeader := c.Request().Header.Get(echo.HeaderAuthorization)

		if authHeader == "" {
			return hh.AccessDeniedResponse(c)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return hh.AccessDeniedResponse(c)
		}

		userRole, err := jwt.ParseJWT(token, mw.cfg.App.JWTSecretKey)
		if err != nil {
			if errors.Is(err, auth.ErrInvalidToken) {
				return hh.AccessDeniedResponse(c)
			}

			return hh.ServerErrorResponse(c, mw.logger, err)
		}

		ContextSetUserRole(c, userRole)
		return next(c)
	}
}
