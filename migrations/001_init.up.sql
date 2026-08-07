CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    avatar TEXT NOT NULL DEFAULT '',
    registered_at TIMESTAMPTZ NOT NULL,
    profile_type TEXT NOT NULL CHECK (profile_type IN ('buyer', 'seller', 'mixed'))
);

CREATE TABLE categories (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE actions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('view', 'message', 'favorite', 'purchase', 'sale')),
    category_id UUID NOT NULL REFERENCES categories(id),
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX actions_user_created_at_idx ON actions (user_id, created_at);

CREATE TABLE recaps (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    year INTEGER NOT NULL CHECK (year >= 2000 AND year <= 2100),
    total_views INTEGER NOT NULL DEFAULT 0 CHECK (total_views >= 0),
    total_messages INTEGER NOT NULL DEFAULT 0 CHECK (total_messages >= 0),
    total_favorites INTEGER NOT NULL DEFAULT 0 CHECK (total_favorites >= 0),
    total_purchases INTEGER NOT NULL DEFAULT 0 CHECK (total_purchases >= 0),
    total_sales INTEGER NOT NULL DEFAULT 0 CHECK (total_sales >= 0),
    top_categories JSONB NOT NULL DEFAULT '[]'::jsonb,
    achievements JSONB NOT NULL DEFAULT '[]'::jsonb,
    activity_days INTEGER NOT NULL DEFAULT 0 CHECK (activity_days >= 0 AND activity_days <= 366),
    generated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, year)
);
