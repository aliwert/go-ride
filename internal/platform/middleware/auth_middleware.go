package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/aliwert/go-ride/internal/platform/apierror"
)

// RequireAuth returns a fiber handler that validates JWT bearer tokens.
// protected routes must present a valid access token; public routes (login, register) should not use this.
func RequireAuth(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return apierror.NewUnauthorized("MISSING_AUTH_HEADER", "missing authorization header")
		}

		// expect "Bearer <token>" format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return apierror.NewUnauthorized("INVALID_AUTH_FORMAT", "invalid authorization format, expected: Bearer <token>")
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			// ensure the signing method is HMAC to prevent algorithm-swap attacks
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return apierror.NewUnauthorized("INVALID_TOKEN", "invalid or expired token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return apierror.NewUnauthorized("INVALID_CLAIMS", "invalid token claims")
		}

		userID, _ := claims.GetSubject()
		if userID == "" {
			return apierror.NewUnauthorized("MISSING_SUBJECT", "token missing subject claim")
		}

		role, _ := claims["role"].(string)

		// store identity in fiber locals so downstream handlers can access them
		c.Locals("userID", userID)
		c.Locals("role", role)

		return c.Next()
	}
}
