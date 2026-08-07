ALTER TABLE users
    DROP COLUMN IF EXISTS listing_views,
    DROP COLUMN IF EXISTS sales,
    DROP COLUMN IF EXISTS purchases,
    DROP COLUMN IF EXISTS metrics,
    DROP COLUMN IF EXISTS favorite_category,
    DROP COLUMN IF EXISTS chats_count,
    DROP COLUMN IF EXISTS accent_color,
    DROP COLUMN IF EXISTS avatar_fallback;
