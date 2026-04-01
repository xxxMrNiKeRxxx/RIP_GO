package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ClaimUserID      = "user_id"
	ClaimIsModerator = "is_moderator"
)

// JWTSecret ключ подписи
func JWTSecret() []byte {
	s := os.Getenv("JWT_KEY")
	if s == "" {
		return []byte("default-dev-only-unsafe")
	}
	return []byte(s)
}

// GenerateToken выпускает JWT с идентификатором пользователя и признаком модератора
func GenerateToken(userID uint, isModerator bool) (string, error) {
	claims := jwt.MapClaims{
		ClaimUserID:      float64(userID),
		ClaimIsModerator: isModerator,
		"exp":            time.Now().Add(24 * time.Hour).Unix(),
		"iat":            time.Now().Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(JWTSecret())
}

// ParseAndValidateToken проверяет подпись и возвращает claims
func ParseAndValidateToken(tokenString string) (jwt.MapClaims, error) {
	tok, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return JWTSecret(), nil
	})
	if err != nil || tok == nil || !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	c, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return c, nil
}

// UserIDFromClaims извлекает uint user_id из JWT claims
func UserIDFromClaims(c jwt.MapClaims) (uint, error) {
	v, ok := c[ClaimUserID]
	if !ok {
		return 0, fmt.Errorf("no user_id")
	}
	switch x := v.(type) {
	case float64:
		if x < 1 {
			return 0, fmt.Errorf("invalid user_id")
		}
		return uint(x), nil
	case int:
		if x < 1 {
			return 0, fmt.Errorf("invalid user_id")
		}
		return uint(x), nil
	default:
		return 0, fmt.Errorf("invalid user_id type")
	}
}
// IsModeratorFromClaims возвращает флаг модератора из JWT
// Учитывает bool и float64
func IsModeratorFromClaims(c jwt.MapClaims) bool {
	v, ok := c[ClaimIsModerator]
	if !ok {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case int:
		return x != 0
	default:
		return false
	}
}
