package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const profileColumns = `
	id, name, avatar, registered_at, profile_type, likes, chats_count,
	own_ads, viewed_ads`

type UserRepository struct {
	pool *pgxpool.Pool
}

type rowScanner interface {
	Scan(dest ...any) error
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, accountID, id uuid.UUID) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+profileColumns+` FROM users WHERE account_id = $1 AND id = $2`,
		accountID,
		id,
	)
	user, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &user, nil
}

func (r *UserRepository) ListProfiles(ctx context.Context, accountID uuid.UUID) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+profileColumns+` FROM users WHERE account_id = $1 ORDER BY name, id`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (r *UserRepository) Create(ctx context.Context, accountID uuid.UUID, user *domain.User) error {
	ownAds, viewedAds, err := marshalProfileActivity(user)
	if err != nil {
		return err
	}
	const query = `
		INSERT INTO users (
			id, account_id, name, avatar, registered_at, profile_type,
			likes, chats_count, own_ads, viewed_ads
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	if _, err := r.pool.Exec(ctx, query,
		user.ID,
		accountID,
		user.Name,
		user.Avatar,
		user.RegisteredAt,
		user.ProfileType,
		user.Likes,
		user.ChatsCount,
		ownAds,
		viewedAds,
	); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) Update(ctx context.Context, accountID uuid.UUID, user *domain.User) error {
	ownAds, viewedAds, err := marshalProfileActivity(user)
	if err != nil {
		return err
	}
	const query = `
		UPDATE users SET
			name = $3,
			avatar = $4,
			registered_at = $5,
			profile_type = $6,
			likes = $7,
			chats_count = $8,
			own_ads = $9,
			viewed_ads = $10,
			updated_at = NOW()
		WHERE account_id = $1 AND id = $2`
	result, err := r.pool.Exec(ctx, query,
		accountID,
		user.ID,
		user.Name,
		user.Avatar,
		user.RegisteredAt,
		user.ProfileType,
		user.Likes,
		user.ChatsCount,
		ownAds,
		viewedAds,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, accountID, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE account_id = $1 AND id = $2`, accountID, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func scanUser(scanner rowScanner) (domain.User, error) {
	var user domain.User
	var ownAdsJSON []byte
	var viewedAdsJSON []byte
	if err := scanner.Scan(
		&user.ID,
		&user.Name,
		&user.Avatar,
		&user.RegisteredAt,
		&user.ProfileType,
		&user.Likes,
		&user.ChatsCount,
		&ownAdsJSON,
		&viewedAdsJSON,
	); err != nil {
		return domain.User{}, err
	}
	if err := json.Unmarshal(ownAdsJSON, &user.OwnAds); err != nil {
		return domain.User{}, fmt.Errorf("decode own ads: %w", err)
	}
	if err := json.Unmarshal(viewedAdsJSON, &user.Views); err != nil {
		return domain.User{}, fmt.Errorf("decode viewed ads: %w", err)
	}
	if user.OwnAds == nil {
		user.OwnAds = []domain.OwnAd{}
	}
	if user.Views == nil {
		user.Views = []domain.ViewedAd{}
	}
	return user, nil
}

func marshalProfileActivity(user *domain.User) ([]byte, []byte, error) {
	ownAds, err := json.Marshal(user.OwnAds)
	if err != nil {
		return nil, nil, fmt.Errorf("encode own ads: %w", err)
	}
	viewedAds, err := json.Marshal(user.Views)
	if err != nil {
		return nil, nil, fmt.Errorf("encode viewed ads: %w", err)
	}
	return ownAds, viewedAds, nil
}
