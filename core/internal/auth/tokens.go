package auth

import (
	"errors"
	"fmt"
	"time"

	"stock-market/backend/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidTokenPayload = errors.New("invalid token payload")
	ErrInvalidTokenType    = errors.New("invalid token type")
)

type tokenClaims struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(cfg Config, subject model.AuthTokenSubject) (string, error) {
	return signToken(cfg.AccessJWTSecret, subject, "access", cfg.AccessTokenTTL())
}

func GenerateRefreshToken(cfg Config, subject model.AuthTokenSubject) (string, error) {
	return signToken(cfg.RefreshJWTSecret, subject, "refresh", cfg.RefreshTokenTTL())
}

func VerifyAccessToken(cfg Config, token string) (model.AuthTokenPayload, error) {
	return verifyTypedToken(cfg.AccessJWTSecret, token, "access")
}

func VerifyRefreshToken(cfg Config, token string) (model.AuthTokenPayload, error) {
	return verifyTypedToken(cfg.RefreshJWTSecret, token, "refresh")
}

func signToken(secret string, subject model.AuthTokenSubject, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		UserID:    subject.UserID,
		Username:  subject.Username,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signed, nil
}

func verifyTypedToken(secret, token, expectedType string) (model.AuthTokenPayload, error) {
	parsed, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return model.AuthTokenPayload{}, err
	}

	claims, ok := parsed.Claims.(*tokenClaims)
	if !ok || !parsed.Valid || claims.UserID == "" || claims.Username == "" || claims.TokenType == "" {
		return model.AuthTokenPayload{}, ErrInvalidTokenPayload
	}

	if claims.TokenType != expectedType {
		return model.AuthTokenPayload{}, ErrInvalidTokenType
	}

	return model.AuthTokenPayload{
		AuthTokenSubject: model.AuthTokenSubject{
			UserID:   claims.UserID,
			Username: claims.Username,
		},
		TokenType: claims.TokenType,
	}, nil
}

func IsTokenExpired(err error) bool {
	return errors.Is(err, jwt.ErrTokenExpired)
}

func IsTokenInvalid(err error) bool {
	return errors.Is(err, jwt.ErrTokenMalformed) ||
		errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
		errors.Is(err, jwt.ErrTokenUnverifiable) ||
		errors.Is(err, ErrInvalidTokenPayload) ||
		errors.Is(err, ErrInvalidTokenType)
}
