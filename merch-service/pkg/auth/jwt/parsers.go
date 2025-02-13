package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"cyansnbrst/merch-service/pkg/auth"
)

// Parse JWT token (HMAC)
func ParseJWT(tokenString, secret string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenMalformed) || errors.Is(err, jwt.ErrTokenExpired) {
			return 0, auth.ErrInvalidToken
		}
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, auth.ErrInvalidToken
		}
		return int64(userID), nil
	}

	return 0, auth.ErrInvalidToken
}
