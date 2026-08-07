package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	"github.com/Valentin0851/avito-recap-backend/internal/usecase/ports"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

const (
	minLoginLength    = 3
	maxLoginLength    = 32
	minPasswordLength = 8
	maxPasswordBytes  = 128
	sessionTokenBytes = 32
)

type Service struct {
	repository ports.AuthRepository
	hasher     passwordHasher
	sessionTTL time.Duration
	now        func() time.Time
}

func New(repository ports.AuthRepository, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}
	return &Service{
		repository: repository,
		hasher:     argon2Hasher{},
		sessionTTL: sessionTTL,
		now:        time.Now,
	}
}

func (s *Service) Register(ctx context.Context, login, password string) (*domain.Account, string, error) {
	login, err := normalizeLogin(login)
	if err != nil {
		return nil, "", err
	}
	if err := validatePassword(password); err != nil {
		return nil, "", err
	}

	passwordHash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, "", fmt.Errorf("hash password: %w", err)
	}
	now := s.now().UTC()
	account := &domain.Account{ID: uuid.New(), Login: login, CreatedAt: now}

	token, tokenHash, expiresAt, err := s.newSession(now)
	if err != nil {
		return nil, "", err
	}
	if err := s.repository.CreateAccountWithSession(
		ctx,
		account,
		passwordHash,
		tokenHash,
		expiresAt,
	); err != nil {
		return nil, "", fmt.Errorf("create account with session: %w", err)
	}
	return account, token, nil
}

func (s *Service) Login(ctx context.Context, login, password string) (*domain.Account, string, error) {
	login, err := normalizeLogin(login)
	if err != nil || password == "" || len(password) > maxPasswordBytes {
		return nil, "", apperrors.ErrInvalidCredentials
	}

	account, passwordHash, err := s.repository.GetAccountByLogin(ctx, login)
	if errors.Is(err, apperrors.ErrNotFound) {
		_, _ = s.hasher.Hash(password)
		return nil, "", apperrors.ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", fmt.Errorf("get account: %w", err)
	}
	valid, err := s.hasher.Compare(password, passwordHash)
	if err != nil {
		return nil, "", fmt.Errorf("verify password: %w", err)
	}
	if !valid {
		return nil, "", apperrors.ErrInvalidCredentials
	}

	now := s.now().UTC()
	token, err := s.createSession(ctx, account.ID, now)
	if err != nil {
		return nil, "", err
	}
	return account, token, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*domain.Account, error) {
	if token == "" {
		return nil, apperrors.ErrUnauthorized
	}
	account, err := s.repository.GetAccountBySessionHash(ctx, sessionHash(token), s.now().UTC())
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, apperrors.ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return account, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.repository.DeleteSession(ctx, sessionHash(token)); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Service) createSession(ctx context.Context, accountID uuid.UUID, now time.Time) (string, error) {
	token, tokenHash, expiresAt, err := s.newSession(now)
	if err != nil {
		return "", err
	}
	if err := s.repository.CreateSession(ctx, accountID, tokenHash, expiresAt); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func (s *Service) newSession(now time.Time) (string, []byte, time.Time, error) {
	randomBytes := make([]byte, sessionTokenBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", nil, time.Time{}, fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(randomBytes)
	return token, sessionHash(token), now.Add(s.sessionTTL), nil
}

func sessionHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func normalizeLogin(login string) (string, error) {
	login = strings.ToLower(strings.TrimSpace(login))
	length := utf8.RuneCountInString(login)
	if length < minLoginLength || length > maxLoginLength {
		return "", apperrors.ErrInvalidLogin
	}
	for _, character := range login {
		if !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '_' && character != '-' && character != '.' {
			return "", apperrors.ErrInvalidLogin
		}
	}
	return login, nil
}

func validatePassword(password string) error {
	if utf8.RuneCountInString(password) < minPasswordLength || len(password) > maxPasswordBytes {
		return apperrors.ErrWeakPassword
	}
	return nil
}
