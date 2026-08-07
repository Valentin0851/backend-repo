INSERT INTO users (id, name, avatar, registered_at, profile_type) VALUES
    ('11111111-1111-4111-8111-111111111111', 'Анна Смирнова', 'https://randomuser.me/api/portraits/women/44.jpg', '2018-04-14T00:00:00Z', 'mixed'),
    ('22222222-2222-4222-8222-222222222222', 'Михаил Орлов', 'https://randomuser.me/api/portraits/men/32.jpg', '2021-09-03T00:00:00Z', 'seller'),
    ('33333333-3333-4333-8333-333333333333', 'Елена Коваль', 'https://randomuser.me/api/portraits/women/68.jpg', '2016-11-28T00:00:00Z', 'buyer'),
    ('44444444-4444-4444-8444-444444444444', 'Даниил Волков', 'https://randomuser.me/api/portraits/men/75.jpg', '2023-02-19T00:00:00Z', 'mixed')
ON CONFLICT (id) DO NOTHING;

INSERT INTO categories (id, name) VALUES
    ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', 'Электроника'),
    ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', 'Авто'),
    ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', 'Хобби и отдых'),
    ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', 'Для дома и дачи'),
    ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', 'Одежда и аксессуары')
ON CONFLICT (id) DO NOTHING;

WITH profile_actions (user_id, action_type, action_count, day_span, category_offset) AS (
    VALUES
        ('11111111-1111-4111-8111-111111111111'::uuid, 'view',     1200, 310, 0),
        ('11111111-1111-4111-8111-111111111111'::uuid, 'message',    65, 310, 0),
        ('11111111-1111-4111-8111-111111111111'::uuid, 'favorite',  140, 310, 0),
        ('11111111-1111-4111-8111-111111111111'::uuid, 'purchase',   12, 310, 0),
        ('11111111-1111-4111-8111-111111111111'::uuid, 'sale',        7, 310, 0),
        ('22222222-2222-4222-8222-222222222222'::uuid, 'view',      320, 180, 1),
        ('22222222-2222-4222-8222-222222222222'::uuid, 'message',    75, 180, 1),
        ('22222222-2222-4222-8222-222222222222'::uuid, 'favorite',   40, 180, 1),
        ('22222222-2222-4222-8222-222222222222'::uuid, 'purchase',    4, 180, 1),
        ('22222222-2222-4222-8222-222222222222'::uuid, 'sale',       13, 180, 1),
        ('33333333-3333-4333-8333-333333333333'::uuid, 'view',      800, 210, 2),
        ('33333333-3333-4333-8333-333333333333'::uuid, 'message',    35, 210, 2),
        ('33333333-3333-4333-8333-333333333333'::uuid, 'favorite',  220, 210, 2),
        ('33333333-3333-4333-8333-333333333333'::uuid, 'purchase',   15, 210, 2),
        ('33333333-3333-4333-8333-333333333333'::uuid, 'sale',        2, 210, 2),
        ('44444444-4444-4444-8444-444444444444'::uuid, 'view',      180,  75, 3),
        ('44444444-4444-4444-8444-444444444444'::uuid, 'message',    15,  75, 3),
        ('44444444-4444-4444-8444-444444444444'::uuid, 'favorite',   25,  75, 3),
        ('44444444-4444-4444-8444-444444444444'::uuid, 'purchase',    3,  75, 3),
        ('44444444-4444-4444-8444-444444444444'::uuid, 'sale',        1,  75, 3)
), expanded AS (
    SELECT pa.*, generate_series(1, pa.action_count) AS sequence_number
    FROM profile_actions pa
)
INSERT INTO actions (id, user_id, type, category_id, created_at)
SELECT
    md5(user_id::text || ':' || action_type || ':' || sequence_number::text)::uuid,
    user_id,
    action_type,
    (ARRAY[
        'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1'::uuid,
        'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2'::uuid,
        'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3'::uuid,
        'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4'::uuid,
        'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5'::uuid
    ])[1 + mod(sequence_number + category_offset, 5)],
    '2025-01-01T00:00:00Z'::timestamptz
        + mod(sequence_number - 1, day_span) * interval '1 day'
        + mod(sequence_number * 37, 86400) * interval '1 second'
FROM expanded
ON CONFLICT (id) DO NOTHING;
