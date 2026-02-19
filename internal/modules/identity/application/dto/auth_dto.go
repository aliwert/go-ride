package dto

import "time"

type RegisterUserRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Role      string
}

type LoginRequest struct {
	Email    string
	Password string
}

// returned after a successful register or login.
// never includes the password hash — only safe-to-expose user fields.
type AuthResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}
