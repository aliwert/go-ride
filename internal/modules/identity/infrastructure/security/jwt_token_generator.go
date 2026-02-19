package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/aliwert/go-ride/internal/modules/identity/domain/entity"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type JwtTokenGenerator struct {
	secret []byte
}

func NewJwtTokenGenerator(secret string) *JwtTokenGenerator {
	if secret == "" {
		// fallback for local dev only — in production this must be set via env
		secret = "go-ride-dev-secret-do-not-use-in-prod"
	}
	return &JwtTokenGenerator{secret: []byte(secret)}
}

func (g *JwtTokenGenerator) GenerateTokens(user *entity.User) (string, string, error) {
	accessToken, err := g.generateToken(user, accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("jwt: access token generation failed: %w", err)
	}

	refreshToken, err := g.generateToken(user, refreshTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("jwt: refresh token generation failed: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (g *JwtTokenGenerator) generateToken(user *entity.User, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": string(user.Role),
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
		"jti":  uuid.New().String(), // unique token id — useful for future revocation
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(g.secret)
}
