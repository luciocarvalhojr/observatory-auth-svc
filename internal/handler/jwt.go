package handler

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luciocarvalhojr/observatory-auth-svc/internal/domain"
)

// issueToken creates a signed JWT for the given identity.
func issueToken(identity *domain.Identity, secret string, ttl time.Duration) (string, error) {
	now := time.Now()

	claims := domain.Claims{
		Email:  identity.Email,
		Name:   identity.Name,
		Groups: identity.Groups,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   identity.Subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "observatory-auth-svc",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}

	return signed, nil
}

// verifyToken validates a signed JWT and returns the claims.
func verifyToken(tokenStr, secret string) (*domain.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &domain.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}

	claims, ok := token.Claims.(*domain.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("verify token: invalid claims")
	}

	return claims, nil
}
