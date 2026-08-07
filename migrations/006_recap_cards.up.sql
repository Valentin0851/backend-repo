ALTER TABLE recaps
    ADD COLUMN summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN cards JSONB NOT NULL DEFAULT '[]'::jsonb;
