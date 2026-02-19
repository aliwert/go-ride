package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/aliwert/go-ride/internal/modules/identity/domain/entity"
)

// postgres unique_violation error code
const uniqueViolation = "23505"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateEmail = errors.New("email already exists")
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		string(user.Role),
		string(user.Status),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return ErrDuplicateEmail
		}
		return fmt.Errorf("postgres: create user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at
		FROM users WHERE id = $1`

	return r.scanUser(ctx, query, id)
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, role, status, created_at, updated_at
		FROM users WHERE email = $1`

	return r.scanUser(ctx, query, email)
}

func (r *PostgresUserRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.Status) error {
	query := `UPDATE users SET status = $1 WHERE id = $2`

	tag, err := r.pool.Exec(ctx, query, string(status), id)
	if err != nil {
		return fmt.Errorf("postgres: update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// centralises the row-to-entity mapping so every finder stays DRY.
func (r *PostgresUserRepository) scanUser(ctx context.Context, query string, args ...any) (*entity.User, error) {
	var u entity.User
	var role, status string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&role,
		&status,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("postgres: scan user: %w", err)
	}

	u.Role = entity.Role(role)
	u.Status = entity.Status(status)

	return &u, nil
}
