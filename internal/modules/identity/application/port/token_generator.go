package port

import "github.com/aliwert/go-ride/internal/modules/identity/domain/entity"

// abstracts JWT (or any other token scheme) away from the
// use-case layer. the actual signing logic lives in the infrastructure layer
// so the application core stays free of crypto/jwt dependencies.
type TokenGenerator interface {
	GenerateTokens(user *entity.User) (accessToken string, refreshToken string, err error)
}
