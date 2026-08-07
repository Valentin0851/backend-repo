CREATE TABLE accounts (
    id UUID PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (login = LOWER(login)),
    CHECK (CHAR_LENGTH(login) BETWEEN 3 AND 32)
);

INSERT INTO accounts (id, login, password_hash, created_at) VALUES (
    '99999999-9999-4999-8999-999999999999',
    'nikita',
    '$argon2id$v=19$m=19456,t=2,p=1$H7eaNLiQRPnkW97cUoyUBw$1gSVVGrLCuY1ORViVB7c8CgI29gueEN7WkKL+4dsm2E',
    '2026-01-01T00:00:00Z'
);

ALTER TABLE users
    ADD COLUMN account_id UUID REFERENCES accounts(id) ON DELETE CASCADE;

UPDATE users
SET account_id = '99999999-9999-4999-8999-999999999999';

ALTER TABLE users
    ALTER COLUMN account_id SET NOT NULL;

CREATE INDEX users_account_id_idx ON users (account_id);

CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE CHECK (OCTET_LENGTH(token_hash) = 32),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sessions_account_id_idx ON sessions (account_id);
CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
