package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) CreateAccountWithSession(
	ctx context.Context,
	account *domain.Account,
	passwordHash string,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin account registration transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := insertAccount(ctx, tx, account, passwordHash); err != nil {
		return err
	}
	if err := insertSession(ctx, tx, account.ID, tokenHash, expiresAt); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit account registration transaction: %w", err)
	}
	return nil
}

func insertAccount(
	ctx context.Context,
	tx pgx.Tx,
	account *domain.Account,
	passwordHash string,
) error {
	_, err := tx.Exec(ctx,
		`INSERT INTO accounts (id, login, password_hash, created_at) VALUES ($1, $2, $3, $4)`,
		account.ID,
		account.Login,
		passwordHash,
		account.CreatedAt,
	)
	if err == nil {
		return nil
	}
	var databaseError *pgconn.PgError
	if errors.As(err, &databaseError) && databaseError.Code == "23505" {
		return apperrors.ErrLoginTaken
	}
	return fmt.Errorf("insert account: %w", err)
}

func (r *AuthRepository) GetAccountByLogin(ctx context.Context, login string) (*domain.Account, string, error) {
	var account domain.Account
	var passwordHash string
	err := r.pool.QueryRow(ctx,
		`SELECT id, login, password_hash, created_at FROM accounts WHERE login = $1`,
		login,
	).Scan(&account.ID, &account.Login, &passwordHash, &account.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", apperrors.ErrNotFound
	}
	if err != nil {
		return nil, "", fmt.Errorf("select account by login: %w", err)
	}
	return &account, passwordHash, nil
}

func (r *AuthRepository) CreateSession(
	ctx context.Context,
	accountID uuid.UUID,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	return insertSession(ctx, r.pool, accountID, tokenHash, expiresAt)
}

type sessionExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertSession(
	ctx context.Context,
	executor sessionExecutor,
	accountID uuid.UUID,
	tokenHash []byte,
	expiresAt time.Time,
) error {
	const query = `
		INSERT INTO sessions (id, account_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)`
	if _, err := executor.Exec(ctx, query, uuid.New(), accountID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetAccountBySessionHash(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
) (*domain.Account, error) {
	const query = `
		SELECT accounts.id, accounts.login, accounts.created_at
		FROM sessions
		JOIN accounts ON accounts.id = sessions.account_id
		WHERE sessions.token_hash = $1 AND sessions.expires_at > $2`
	var account domain.Account
	if err := r.pool.QueryRow(ctx, query, tokenHash, now).Scan(
		&account.ID,
		&account.Login,
		&account.CreatedAt,
	); errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("select account by session: %w", err)
	}
	return &account, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
