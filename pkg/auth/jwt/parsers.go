package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/cyansnbrst/pvz-service/gen/pvzapi"
	"github.com/cyansnbrst/pvz-service/pkg/auth"
)

// Parse JWT token (HMAC)
func ParseJWT(tokenString, secret string) (pvzapi.UserRole, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrSignatureInvalid) {
			return "", auth.ErrInvalidToken
		}
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		role, ok := claims["role"].(string)
		if !ok {
			return "", auth.ErrInvalidToken
		}
		return pvzapi.UserRole(role), nil
	}

	return "", auth.ErrInvalidToken
}
