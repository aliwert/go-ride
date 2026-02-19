package entity

import (
	"time"

	"github.com/google/uuid"
)

// role

type Role string

const (
	RoleRider  Role = "RIDER"
	RoleDriver Role = "DRIVER"
	RoleAdmin  Role = "ADMIN"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleRider, RoleDriver, RoleAdmin:
		return true
	}
	return false
}

// status

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusSuspended:
		return true
	}
	return false
}

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email, passwordHash, firstName, lastName string, role Role) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		Status:       StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
