package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
)

type storedSession struct {
	accountID uuid.UUID
	expiresAt time.Time
}

type fakeAuthRepository struct {
	accounts map[string]domain.Account
	hashes   map[string]string
	sessions map[string]storedSession
}

func newFakeAuthRepository() *fakeAuthRepository {
	return &fakeAuthRepository{
		accounts: make(map[string]domain.Account),
		hashes:   make(map[string]string),
		sessions: make(map[string]storedSession),
	}
}

func (r *fakeAuthRepository) CreateAccountWithSession(
	_ context.Context,
	account *domain.Account,
	passwordHash string,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	if _, exists := r.accounts[account.Login]; exists {
		return apperrors.ErrLoginTaken
	}
	r.accounts[account.Login] = *account
	r.hashes[account.Login] = passwordHash
	r.sessions[string(tokenHash)] = storedSession{
		accountID: account.ID,
		expiresAt: expiresAt,
	}
	return nil
}

func (r *fakeAuthRepository) GetAccountByLogin(
	_ context.Context,
	login string,
) (*domain.Account, string, error) {
	account, exists := r.accounts[login]
	if !exists {
		return nil, "", apperrors.ErrNotFound
	}
	return &account, r.hashes[login], nil
}

func (r *fakeAuthRepository) CreateSession(
	_ context.Context,
	accountID uuid.UUID,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	r.sessions[string(tokenHash)] = storedSession{accountID: accountID, expiresAt: expiresAt}
	return nil
}

func (r *fakeAuthRepository) GetAccountBySessionHash(
	_ context.Context,
	tokenHash []byte,
	now time.Time,
) (*domain.Account, error) {
	session, exists := r.sessions[string(tokenHash)]
	if !exists || !session.expiresAt.After(now) {
		return nil, apperrors.ErrNotFound
	}
	for _, account := range r.accounts {
		if account.ID == session.accountID {
			result := account
			return &result, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *fakeAuthRepository) DeleteSession(_ context.Context, tokenHash []byte) error {
	delete(r.sessions, string(tokenHash))
	return nil
}

func TestServiceRegistrationAndSessionLifecycle(t *testing.T) {
	repository := newFakeAuthRepository()
	service := New(repository, time.Hour)
	service.now = func() time.Time { return time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC) }

	account, token, err := service.Register(context.Background(), " Alice ", "strong-pass")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if account.Login != "alice" || token == "" {
		t.Fatalf("account/token = %#v/%q, want normalized login and token", account, token)
	}
	if repository.hashes["alice"] == "strong-pass" {
		t.Fatal("password was stored in plaintext")
	}
	valid, err := (argon2Hasher{}).Compare("strong-pass", repository.hashes["alice"])
	if err != nil || !valid {
		t.Fatalf("stored password hash is invalid: valid=%v err=%v", valid, err)
	}
	if _, exists := repository.sessions[string(sessionHash(token))]; !exists {
		t.Fatal("repository does not contain the session token hash")
	}

	current, err := service.Authenticate(context.Background(), token)
	if err != nil || current.ID != account.ID {
		t.Fatalf("authenticate = %#v, %v", current, err)
	}
	if err := service.Logout(context.Background(), token); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.Authenticate(context.Background(), token); !errors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("authenticate after logout error = %v, want unauthorized", err)
	}
}

func TestServiceRejectsInvalidRegistrationAndCredentials(t *testing.T) {
	repository := newFakeAuthRepository()
	service := New(repository, time.Hour)

	if _, _, err := service.Register(context.Background(), "a!", "strong-pass"); !errors.Is(err, apperrors.ErrInvalidLogin) {
		t.Fatalf("invalid login error = %v", err)
	}
	if _, _, err := service.Register(context.Background(), "alice", "short"); !errors.Is(err, apperrors.ErrWeakPassword) {
		t.Fatalf("weak password error = %v", err)
	}
	if _, _, err := service.Login(context.Background(), "unknown", "strong-pass"); !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("unknown login error = %v", err)
	}

	if _, _, err := service.Register(context.Background(), "alice", "strong-pass"); err != nil {
		t.Fatalf("register valid account: %v", err)
	}
	if _, _, err := service.Register(context.Background(), "ALICE", "another-pass"); !errors.Is(err, apperrors.ErrLoginTaken) {
		t.Fatalf("duplicate login error = %v", err)
	}
	if _, _, err := service.Login(context.Background(), "alice", "wrong-pass"); !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v", err)
	}
	if account, token, err := service.Login(context.Background(), "ALICE", "strong-pass"); err != nil || account.Login != "alice" || token == "" {
		t.Fatalf("valid login = %#v, %q, %v", account, token, err)
	}
}
