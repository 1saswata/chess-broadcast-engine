package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var key = []byte("my-super-secret-key")

// TODO(saswata): TECH DEBT - Sprint 7
// Remove matchID from JWT payload once the `matches` table is implemented.
// Match authorization should be a database/cache lookup, not a stateless claim.
func GenerateToken(userID string, matchID int32, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["id"] = userID
	claims["role"] = role
	claims["matchID"] = matchID
	claims["exp"] = time.Now().Add(time.Minute * 60).Unix()

	tokenString, err := token.SignedString(key)
	return tokenString, err
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Not authorized")
		}
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if token.Valid {
		return token.Claims.(jwt.MapClaims), nil
	}
	return nil, fmt.Errorf("Not authorized")
}
