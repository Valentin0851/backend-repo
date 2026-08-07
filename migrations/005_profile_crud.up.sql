ALTER TABLE users
    ADD COLUMN likes INTEGER NOT NULL DEFAULT 0 CHECK (likes >= 0),
    ADD COLUMN own_ads JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN viewed_ads JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE users
SET likes = COALESCE((metrics ->> 'likes')::integer, 0);

UPDATE users SET
    viewed_ads = '[
      {"title":"Смартфон","category":"Электроника","subcategory":"Смартфоны","price":118000,"viewCount":7,"lastViewedAt":"2026-03-12T14:10:00Z","isFavorite":true,"favoritedAt":"2026-03-10T18:20:00Z","isPurchased":true,"purchasedAt":"2026-03-12T14:10:00Z"},
      {"title":"Наушники","category":"Электроника","subcategory":"Аудио","price":12990,"viewCount":3,"lastViewedAt":"2026-05-04T16:45:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-05-04T16:45:00Z"},
      {"title":"Чехол для телефона","category":"Аксессуары","price":490,"viewCount":2,"lastViewedAt":"2026-01-17T13:30:00Z","isFavorite":true,"favoritedAt":"2026-01-15T11:05:00Z","isPurchased":true,"purchasedAt":"2026-01-17T13:30:00Z"}
    ]'::jsonb,
    own_ads = '[
      {"title":"Планшет","category":"Электроника","subcategory":"Планшеты","price":28500,"viewCount":214,"isArchived":true,"isSold":true,"soldAt":"2026-02-20T00:00:00Z","review":{"comment":"Планшет в отличном состоянии, всё как в описании.","rating":5,"createdAt":"2026-02-21T00:00:00Z"}},
      {"title":"Кресло","category":"Для дома","subcategory":"Мебель","price":7500,"viewCount":86,"isArchived":true,"isSold":true,"soldAt":"2026-06-18T00:00:00Z"}
    ]'::jsonb
WHERE id = '11111111-1111-4111-8111-111111111111';

UPDATE users SET
    viewed_ads = '[
      {"title":"Акустическая гитара","category":"Хобби и отдых","subcategory":"Музыкальные инструменты","price":64900,"viewCount":11,"lastViewedAt":"2026-05-21T19:20:00Z","isFavorite":true,"favoritedAt":"2026-05-12T12:10:00Z","isPurchased":true,"purchasedAt":"2026-05-21T19:20:00Z"},
      {"title":"Велосипедный шлем","category":"Хобби и отдых","subcategory":"Спорт","price":5400,"viewCount":5,"lastViewedAt":"2026-04-02T16:05:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-04-02T16:05:00Z"},
      {"title":"Комплект струн","category":"Хобби и отдых","subcategory":"Музыкальные инструменты","price":850,"viewCount":2,"lastViewedAt":"2026-02-06T18:00:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-02-06T18:00:00Z"}
    ]'::jsonb,
    own_ads = '[
      {"title":"Игровая консоль","category":"Электроника","subcategory":"Игровые приставки","price":32000,"viewCount":326,"isArchived":true,"isSold":true,"soldAt":"2026-03-04T00:00:00Z","review":{"comment":"Быстро договорились, всё работает.","rating":5,"createdAt":"2026-03-05T00:00:00Z"}},
      {"title":"Сноуборд","category":"Хобби и отдых","subcategory":"Спорт","price":18500,"viewCount":189,"isArchived":true,"isSold":true,"soldAt":"2026-01-26T00:00:00Z"}
    ]'::jsonb
WHERE id = '22222222-2222-4222-8222-222222222222';

UPDATE users SET
    viewed_ads = '[
      {"title":"Робот-пылесос","category":"Для дома и дачи","subcategory":"Бытовая техника","price":54990,"viewCount":9,"lastViewedAt":"2026-04-09T17:30:00Z","isFavorite":true,"favoritedAt":"2026-04-01T10:15:00Z","isPurchased":true,"purchasedAt":"2026-04-09T17:30:00Z"},
      {"title":"Торшер","category":"Для дома и дачи","subcategory":"Мебель и интерьер","price":8990,"viewCount":4,"lastViewedAt":"2026-05-28T20:10:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-05-28T20:10:00Z"},
      {"title":"Набор полотенец","category":"Для дома и дачи","subcategory":"Текстиль","price":1290,"viewCount":3,"lastViewedAt":"2026-07-02T14:05:00Z","isFavorite":true,"favoritedAt":"2026-06-25T18:00:00Z","isPurchased":true,"purchasedAt":"2026-07-02T14:05:00Z"}
    ]'::jsonb,
    own_ads = '[
      {"title":"Детская кроватка","category":"Товары для детей","subcategory":"Детская мебель","price":11200,"viewCount":164,"isArchived":true,"isSold":true,"soldAt":"2026-02-14T00:00:00Z"},
      {"title":"Кофемашина","category":"Для дома и дачи","subcategory":"Бытовая техника","price":19700,"viewCount":241,"isArchived":true,"isSold":true,"soldAt":"2026-06-03T00:00:00Z","review":{"comment":"Спасибо за чистую и исправную кофемашину.","rating":5,"createdAt":"2026-06-04T00:00:00Z"}}
    ]'::jsonb
WHERE id = '33333333-3333-4333-8333-333333333333';

UPDATE users SET
    viewed_ads = '[
      {"title":"Комплект колёс","category":"Транспорт","subcategory":"Запчасти","price":84000,"viewCount":8,"lastViewedAt":"2026-03-30T19:40:00Z","isFavorite":true,"favoritedAt":"2026-03-22T09:30:00Z","isPurchased":true,"purchasedAt":"2026-03-30T19:40:00Z"},
      {"title":"Видеорегистратор","category":"Транспорт","subcategory":"Автоэлектроника","price":15400,"viewCount":6,"lastViewedAt":"2026-06-11T12:20:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-06-11T12:20:00Z"},
      {"title":"Держатель для телефона","category":"Транспорт","subcategory":"Автоаксессуары","price":990,"viewCount":4,"lastViewedAt":"2026-01-24T11:55:00Z","isFavorite":false,"isPurchased":true,"purchasedAt":"2026-01-24T11:55:00Z"}
    ]'::jsonb,
    own_ads = '[
      {"title":"Горный велосипед","category":"Хобби и отдых","subcategory":"Спорт","price":46800,"viewCount":302,"isArchived":true,"isSold":true,"soldAt":"2026-04-16T00:00:00Z","review":{"comment":"Велосипед соответствует описанию, продавца рекомендую.","rating":5,"createdAt":"2026-04-17T00:00:00Z"}},
      {"title":"Автомагнитола","category":"Транспорт","subcategory":"Автоэлектроника","price":9400,"viewCount":117,"isArchived":true,"isSold":true,"soldAt":"2026-05-08T00:00:00Z"}
    ]'::jsonb
WHERE id = '44444444-4444-4444-8444-444444444444';
