package usecase

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/aliwert/go-ride/internal/modules/identity/application/dto"
	"github.com/aliwert/go-ride/internal/modules/identity/application/port"
	"github.com/aliwert/go-ride/internal/modules/identity/domain/entity"
	"github.com/aliwert/go-ride/internal/modules/identity/domain/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyTaken  = errors.New("email already taken")
	ErrAccountSuspended   = errors.New("account suspended")
	ErrInvalidRole        = errors.New("invalid role")
)

type AuthUseCase struct {
	userRepo repository.UserRepository
	tokenGen port.TokenGenerator
}

func NewAuthUseCase(repo repository.UserRepository, tokenGen port.TokenGenerator) *AuthUseCase {
	return &AuthUseCase{
		userRepo: repo,
		tokenGen: tokenGen,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, req *dto.RegisterUserRequest) (*dto.AuthResponse, error) {
	role := entity.Role(req.Role)
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	// hash the plain-text password before it ever touches the domain or persistence
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(req.Email, string(hash), req.FirstName, req.LastName, role)

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := uc.tokenGen.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return buildAuthResponse(user, accessToken, refreshToken), nil
}

func (uc *AuthUseCase) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if user.Status == entity.StatusSuspended {
		return nil, ErrAccountSuspended
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := uc.tokenGen.GenerateTokens(user)
	if err != nil {
		return nil, err
	}

	return buildAuthResponse(user, accessToken, refreshToken), nil
}

// buildAuthResponse maps the domain entity to a transport-safe DTO.
// password hash is deliberately excluded.
func buildAuthResponse(user *entity.User, accessToken, refreshToken string) *dto.AuthResponse {
	return &dto.AuthResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.Role),
		Status:       string(user.Status),
		CreatedAt:    user.CreatedAt,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}
