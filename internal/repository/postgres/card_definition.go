package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Valentin0851/avito-recap-backend/internal/domain"
	apperrors "github.com/Valentin0851/avito-recap-backend/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

const cardDefinitionColumns = `
	id, name, target_user_id, kind, metric, analysis, condition_operator,
	condition_value, title, description, value_suffix, layout, theme, icon,
	shareable, sort_order, is_active, created_by, created_at, updated_at`

type CardDefinitionRepository struct {
	pool *pgxpool.Pool
}

func NewCardDefinitionRepository(pool *pgxpool.Pool) *CardDefinitionRepository {
	return &CardDefinitionRepository{pool: pool}
}

func (r *CardDefinitionRepository) Create(ctx context.Context, definition *domain.CardDefinition) error {
	const query = `
		INSERT INTO card_definitions (
			id, name, target_user_id, kind, metric, analysis, condition_operator,
			condition_value, title, description, value_suffix, layout, theme, icon,
			shareable, sort_order, is_active, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`
	if _, err := r.pool.Exec(ctx, query, cardDefinitionValues(definition)...); err != nil {
		if isForeignKeyViolation(err) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("insert card definition: %w", err)
	}
	return nil
}

func (r *CardDefinitionRepository) List(ctx context.Context) ([]domain.CardDefinition, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+cardDefinitionColumns+` FROM card_definitions ORDER BY sort_order, created_at, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list card definitions: %w", err)
	}
	return scanCardDefinitions(rows)
}

func (r *CardDefinitionRepository) Update(ctx context.Context, definition *domain.CardDefinition) error {
	const query = `
		UPDATE card_definitions SET
			name = $2, target_user_id = $3, kind = $4, metric = $5,
			analysis = $6, condition_operator = $7, condition_value = $8,
			title = $9, description = $10, value_suffix = $11, layout = $12,
			theme = $13, icon = $14, shareable = $15, sort_order = $16,
			is_active = $17, updated_at = $20
		WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, cardDefinitionValues(definition)...)
	if err != nil {
		if isForeignKeyViolation(err) {
			return apperrors.ErrNotFound
		}
		return fmt.Errorf("update card definition: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CardDefinitionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM card_definitions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete card definition: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *CardDefinitionRepository) ListActiveForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]domain.CardDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+cardDefinitionColumns+`
		FROM card_definitions
		WHERE is_active = TRUE AND (target_user_id IS NULL OR target_user_id = $1)
		ORDER BY sort_order, created_at, id`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list active card definitions: %w", err)
	}
	return scanCardDefinitions(rows)
}

func cardDefinitionValues(definition *domain.CardDefinition) []any {
	return []any{
		definition.ID,
		definition.Name,
		definition.TargetUserID,
		definition.Kind,
		definition.Metric,
		definition.Analysis,
		definition.ConditionOperator,
		definition.ConditionValue,
		definition.Title,
		definition.Description,
		definition.ValueSuffix,
		definition.Layout,
		definition.Theme,
		definition.Icon,
		definition.Shareable,
		definition.SortOrder,
		definition.IsActive,
		definition.CreatedBy,
		definition.CreatedAt,
		definition.UpdatedAt,
	}
}

func scanCardDefinitions(rows pgx.Rows) ([]domain.CardDefinition, error) {
	defer rows.Close()
	definitions := make([]domain.CardDefinition, 0)
	for rows.Next() {
		definition, err := scanCardDefinition(rows)
		if err != nil {
			return nil, fmt.Errorf("scan card definition: %w", err)
		}
		definitions = append(definitions, definition)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate card definitions: %w", err)
	}
	return definitions, nil
}

func scanCardDefinition(scanner rowScanner) (domain.CardDefinition, error) {
	var definition domain.CardDefinition
	var targetUserID pgtype.UUID
	var conditionValue pgtype.Float8
	err := scanner.Scan(
		&definition.ID,
		&definition.Name,
		&targetUserID,
		&definition.Kind,
		&definition.Metric,
		&definition.Analysis,
		&definition.ConditionOperator,
		&conditionValue,
		&definition.Title,
		&definition.Description,
		&definition.ValueSuffix,
		&definition.Layout,
		&definition.Theme,
		&definition.Icon,
		&definition.Shareable,
		&definition.SortOrder,
		&definition.IsActive,
		&definition.CreatedBy,
		&definition.CreatedAt,
		&definition.UpdatedAt,
	)
	if err != nil {
		return domain.CardDefinition{}, err
	}
	if targetUserID.Valid {
		id := uuid.UUID(targetUserID.Bytes)
		definition.TargetUserID = &id
	}
	if conditionValue.Valid {
		value := conditionValue.Float64
		definition.ConditionValue = &value
	}
	return definition, nil
}

func isForeignKeyViolation(err error) bool {
	var databaseError *pgconn.PgError
	return errors.As(err, &databaseError) && databaseError.Code == "23503"
}
