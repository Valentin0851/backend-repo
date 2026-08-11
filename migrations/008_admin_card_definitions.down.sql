DROP TABLE IF EXISTS card_definitions;

ALTER TABLE accounts
    DROP COLUMN IF EXISTS is_admin;
