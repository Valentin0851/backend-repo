ALTER TABLE accounts
    ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE card_definitions (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (CHAR_LENGTH(name) BETWEEN 1 AND 100),
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('statistic', 'highlight')),
    metric TEXT NOT NULL CHECK (metric IN (
        'total_views',
        'favorites',
        'chats',
        'purchases',
        'sales',
        'listing_views',
        'contacts',
        'reviews',
        'activity_days',
        'categories',
        'deals'
    )),
    analysis TEXT NOT NULL CHECK (analysis IN ('total', 'monthly_average', 'monthly_max')),
    condition_operator TEXT NOT NULL CHECK (condition_operator IN ('always', 'gt', 'gte', 'lt', 'lte', 'eq')),
    condition_value DOUBLE PRECISION,
    title TEXT NOT NULL CHECK (CHAR_LENGTH(title) BETWEEN 1 AND 160),
    description TEXT NOT NULL DEFAULT '' CHECK (CHAR_LENGTH(description) <= 500),
    value_suffix TEXT NOT NULL DEFAULT '' CHECK (CHAR_LENGTH(value_suffix) <= 40),
    layout TEXT NOT NULL CHECK (layout IN ('statistic', 'hero')),
    theme TEXT NOT NULL CHECK (CHAR_LENGTH(theme) BETWEEN 1 AND 50),
    icon TEXT NOT NULL CHECK (CHAR_LENGTH(icon) BETWEEN 1 AND 50),
    shareable BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INTEGER NOT NULL DEFAULT 100 CHECK (sort_order >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_by UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        (condition_operator = 'always' AND condition_value IS NULL)
        OR
        (condition_operator <> 'always' AND condition_value IS NOT NULL)
    ),
    CHECK (
        analysis = 'total'
        OR metric IN ('total_views', 'favorites', 'purchases', 'sales')
    )
);

CREATE INDEX card_definitions_active_target_idx
    ON card_definitions (target_user_id, sort_order, created_at)
    WHERE is_active = TRUE;
