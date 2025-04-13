package httphelpers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
)

const (
	msgServerError  = "the server encountered a problem and could not process your request"
	msgAccessDenied = "access denied"
)

// Log an error
func logError(c echo.Context, l *zap.Logger, err error) {
	l.Error("an error occurred",
		zap.String("request_method", c.Request().Method),
		zap.String("request_url", c.Request().URL.String()),
		zap.Error(err),
	)
}

// Error response
func errorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, pvzapi.Error{Message: message})
}

// Server error response (500)
func ServerErrorResponse(c echo.Context, l *zap.Logger, err error) error {
	logError(c, l, err)
	return errorResponse(c, http.StatusInternalServerError, msgServerError)
}

// Bad request response (400)
func BadRequestResponse(c echo.Context, err error) error {
	return errorResponse(c, http.StatusBadRequest, err.Error())
}

// Access denied (403)
func AccessDeniedResponse(c echo.Context) error {
	return errorResponse(c, http.StatusForbidden, msgAccessDenied)
}
