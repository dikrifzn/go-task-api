package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidInput = errors.New("invalid input")
var ErrInvalidCredentials = errors.New("email or password is incorrect")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
) (User, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if name == "" || len(name) > 100 {
		return User{}, ErrInvalidInput
	}

	if len(email) > 254 {
		return User{}, ErrInvalidInput
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		return User{}, ErrInvalidInput
	}

	if len(req.Password) < 8 || len(req.Password) > 72 {
		return User{}, ErrInvalidInput
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	return s.repo.Create(
		ctx, name, email, string(passwordHash),
	)
}

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	u, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return AuthResponse{}, fmt.Errorf("generate token: %w", err)
	}

	token := hex.EncodeToString(randomBytes)
	tokenHash := hashToken(token)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)

	if err := s.repo.CreateSession(
		ctx, u.ID, tokenHash, expiresAt,
	); err != nil {
		return AuthResponse{}, err
	}

	// Jangan pernah mengirim hash password ke client.
	u.PasswordHash = ""

	return AuthResponse{
		User:  u,
		Token: token,
	}, nil
}
