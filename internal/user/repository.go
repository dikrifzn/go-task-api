package user

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email already registered")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	name, email, passwordHash string,
) (User, error) {
	query := `
		INSERT INTO users (name, email, password_hash)
		VALUES (?, ?, ?)
	`

	result, err := r.db.ExecContext(
		ctx, query, name, email, passwordHash,
	)
	if err != nil {
		// SQLite memakai kode constraint untuk email duplikat.
		if isUniqueViolation(err) {
			return User{}, ErrEmailExists
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("get user ID: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE email = ?
	`

	var u User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by email: %w", err)
	}

	return u, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id int64,
) (User, error) {
	query := `
		SELECT id, name, email, password_hash, created_at
		FROM users
		WHERE id = ?
	`

	var u User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by ID: %w", err)
	}

	return u, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	userID int64,
	tokenHash string,
	expiresAt time.Time,
) error {
	query := `
        INSERT INTO sessions (user_id, token_hash, expires_at)
        VALUES (?, ?, ?)
    `

	_, err := r.db.ExecContext(
		ctx, query, userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	return nil
}

func (r *Repository) GetUserByToken(
	ctx context.Context,
	tokenHash string,
) (User, error) {
	query := `
        SELECT u.id, u.name, u.email, u.password_hash, u.created_at
        FROM users u
        JOIN sessions s ON s.user_id = u.id
        WHERE s.token_hash = ? AND s.expires_at > ?
    `

	var u User
	err := r.db.QueryRowContext(
		ctx, query, tokenHash, time.Now().UTC(),
	).Scan(
		&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("get user by token: %w", err)
	}

	return u, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Implementasi ini khusus untuk driver SQLite modernc.org/sqlite.
func isUniqueViolation(err error) bool {
	return err != nil &&
		(strings.Contains(strings.ToLower(err.Error()), "unique constraint") ||
			strings.Contains(strings.ToLower(err.Error()), "constraint failed"))
}
