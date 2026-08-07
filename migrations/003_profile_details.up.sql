ALTER TABLE users
    ADD COLUMN avatar_fallback TEXT NOT NULL DEFAULT '',
    ADD COLUMN accent_color TEXT NOT NULL DEFAULT '#00aaff',
    ADD COLUMN chats_count INTEGER NOT NULL DEFAULT 0 CHECK (chats_count >= 0),
    ADD COLUMN favorite_category TEXT NOT NULL DEFAULT '',
    ADD COLUMN metrics JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN purchases JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN sales JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN listing_views JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE users SET
    avatar_fallback = 'АС',
    accent_color = '#00aaff',
    chats_count = 43,
    favorite_category = 'Электроника',
    metrics = '{"activeDays":163,"city":"Москва","createdListings":25,"favoriteListings":32,"likes":148,"rating":4.9,"reviews":17}'::jsonb,
    purchases = '[{"title":"Смартфон","category":"Электроника","price":118000,"date":"2026-03-12T00:00:00Z"},{"title":"Наушники","category":"Аудио","price":12990,"date":"2026-05-04T00:00:00Z"},{"title":"Чехол для телефона","category":"Аксессуары","price":490,"date":"2026-01-17T00:00:00Z"}]'::jsonb,
    sales = '[{"title":"Планшет","category":"Электроника","price":28500,"date":"2026-02-20T00:00:00Z","inquiriesCount":14},{"title":"Кресло","category":"Для дома","price":7500,"date":"2026-06-18T00:00:00Z","inquiriesCount":9}]'::jsonb,
    listing_views = '[{"title":"Ноутбук для работы","category":"Ноутбуки","likes":12,"viewedAt":"2026-07-10T20:15:00Z","viewCount":7},{"title":"Фотоаппарат","category":"Фототехника","likes":8,"viewedAt":"2026-06-29T13:40:00Z","viewCount":3}]'::jsonb
WHERE id = '11111111-1111-4111-8111-111111111111';

UPDATE users SET
    avatar_fallback = 'МО',
    accent_color = '#965eeb',
    chats_count = 51,
    favorite_category = 'Хобби и отдых',
    metrics = '{"activeDays":109,"city":"Казань","createdListings":31,"favoriteListings":19,"likes":86,"rating":4.8,"reviews":12}'::jsonb,
    purchases = '[{"title":"Акустическая гитара","category":"Музыкальные инструменты","price":64900,"date":"2026-05-21T00:00:00Z"},{"title":"Велосипедный шлем","category":"Спорт","price":5400,"date":"2026-04-02T00:00:00Z"},{"title":"Комплект струн","category":"Музыкальные инструменты","price":850,"date":"2026-02-06T00:00:00Z"}]'::jsonb,
    sales = '[{"title":"Игровая консоль","category":"Игры","price":32000,"date":"2026-03-04T00:00:00Z","inquiriesCount":21},{"title":"Сноуборд","category":"Спорт","price":18500,"date":"2026-01-26T00:00:00Z","inquiriesCount":17}]'::jsonb,
    listing_views = '[{"title":"Электрогитара","category":"Музыкальные инструменты","likes":15,"viewedAt":"2026-07-17T21:05:00Z","viewCount":11},{"title":"Палатка на 4 человека","category":"Туризм","likes":6,"viewedAt":"2026-06-12T09:25:00Z","viewCount":4}]'::jsonb
WHERE id = '22222222-2222-4222-8222-222222222222';

UPDATE users SET
    avatar_fallback = 'ЕК',
    accent_color = '#04e061',
    chats_count = 27,
    favorite_category = 'Для дома и дачи',
    metrics = '{"activeDays":202,"city":"Санкт-Петербург","createdListings":16,"favoriteListings":47,"likes":219,"rating":5.0,"reviews":28}'::jsonb,
    purchases = '[{"title":"Робот-пылесос","category":"Бытовая техника","price":54990,"date":"2026-04-09T00:00:00Z"},{"title":"Торшер","category":"Мебель и интерьер","price":8990,"date":"2026-05-28T00:00:00Z"},{"title":"Набор полотенец","category":"Текстиль","price":1290,"date":"2026-07-02T00:00:00Z"}]'::jsonb,
    sales = '[{"title":"Детская кроватка","category":"Товары для детей","price":11200,"date":"2026-02-14T00:00:00Z","inquiriesCount":18},{"title":"Кофемашина","category":"Бытовая техника","price":19700,"date":"2026-06-03T00:00:00Z","inquiriesCount":11}]'::jsonb,
    listing_views = '[{"title":"Диван","category":"Мебель и интерьер","likes":23,"viewedAt":"2026-07-24T18:30:00Z","viewCount":9},{"title":"Увлажнитель воздуха","category":"Бытовая техника","likes":10,"viewedAt":"2026-07-08T11:45:00Z","viewCount":5}]'::jsonb
WHERE id = '33333333-3333-4333-8333-333333333333';

UPDATE users SET
    avatar_fallback = 'ДВ',
    accent_color = '#ff4053',
    chats_count = 36,
    favorite_category = 'Транспорт',
    metrics = '{"activeDays":87,"city":"Екатеринбург","createdListings":22,"favoriteListings":14,"likes":63,"rating":4.7,"reviews":9}'::jsonb,
    purchases = '[{"title":"Комплект колёс","category":"Запчасти","price":84000,"date":"2026-03-30T00:00:00Z"},{"title":"Видеорегистратор","category":"Автоэлектроника","price":15400,"date":"2026-06-11T00:00:00Z"},{"title":"Держатель для телефона","category":"Автоаксессуары","price":990,"date":"2026-01-24T00:00:00Z"}]'::jsonb,
    sales = '[{"title":"Горный велосипед","category":"Спорт","price":46800,"date":"2026-04-16T00:00:00Z","inquiriesCount":26},{"title":"Автомагнитола","category":"Автоэлектроника","price":9400,"date":"2026-05-08T00:00:00Z","inquiriesCount":13}]'::jsonb,
    listing_views = '[{"title":"Кроссовый мотоцикл","category":"Мототехника","likes":19,"viewedAt":"2026-07-22T19:10:00Z","viewCount":8},{"title":"Багажник на крышу","category":"Автоаксессуары","likes":7,"viewedAt":"2026-07-03T08:55:00Z","viewCount":4}]'::jsonb
WHERE id = '44444444-4444-4444-8444-444444444444';
