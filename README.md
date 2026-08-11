# Avito Recap — Backend Engineering Portfolio

Портфолио-репозиторий с моей зоной ответственности в проекте персональных
итогов года.

Моя роль начиналась как **Backend 2 — Data Layer**: PostgreSQL, Redis,
миграции, репозитории и целостность данных. По мере развития проекта зона
ответственности расширилась: я добавил наблюдаемость слоя данных и разработал
backend-конструктор динамических recap-карточек для администратора.

Репозиторий содержит мою реализацию и минимальные доменные/архитектурные
контракты, необходимые для её сборки и демонстрации. Frontend и посторонние
части командного приложения сюда не переносились.

## Мой вклад

### 1. Надёжный слой хранения

- спроектировал и развивал схему PostgreSQL;
- написал версионируемые SQL-миграции `up/down`;
- реализовал PostgreSQL adapters через `pgx/v5`;
- добавил пул соединений `pgxpool`;
- реализовал CRUD профилей с изоляцией по `account_id`;
- добавил UPSERT итогов года по `(user_id, year)`;
- преобразовал ошибки PostgreSQL в доменные ошибки.

### 2. Транзакционная регистрация

Создание аккаунта и первой серверной сессии выполняется атомарно:

```text
BEGIN
  INSERT account
  INSERT session
COMMIT
```

При ошибке создания сессии происходит `ROLLBACK`. В базе не остаётся аккаунт,
в который невозможно войти.

### 3. Redis-кэш с graceful degradation

Для готовых recap реализован read-through/write-through decorator:

```text
READ:
  Redis hit  -> вернуть recap
  Redis miss -> PostgreSQL -> Redis -> вернуть recap

WRITE:
  PostgreSQL -> Redis
```

- ключ: `recap:v1:{user_id}:{year}`;
- TTL: 24 часа;
- PostgreSQL остаётся источником истины;
- ошибка Redis не ломает пользовательский запрос;
- повреждённая cache-запись удаляется и восстанавливается из PostgreSQL.

### 4. Метрики PostgreSQL и Redis

Я добавил Prometheus collectors для контроля состояния инфраструктуры:

**PostgreSQL pool**

- acquired, idle, total и max connections;
- количество попыток получения соединения;
- время ожидания соединений;
- отменённые acquire-операции;
- количество созданных соединений.

**Redis recap cache**

- cache hit;
- cache miss;
- ошибки `get`, `set` и `delete`.

Метрики позволяют увидеть деградацию кэша, нехватку соединений и рост времени
ожидания до того, как проблема станет массовой пользовательской ошибкой.

### 5. Admin Card Definition Builder

Я разработал backend-конструктор, позволяющий администратору создавать новые
виды recap-карточек без изменения frontend-кода.

Администратор задаёт:

- внутреннее название и пользовательский заголовок;
- глобальную область действия или конкретный профиль;
- статистику: просмотры, избранное, покупки, продажи, контакты и другие;
- анализ: итог, среднее за месяц или максимальное значение месяца;
- условие показа: `always`, `gt`, `gte`, `lt`, `lte`, `eq`;
- оформление: layout, theme, icon и суффикс значения;
- порядок, доступность и возможность публикации карточки.

Доступные endpoint:

```text
GET    /api/admin/card-definitions/options
GET    /api/admin/card-definitions
POST   /api/admin/card-definitions
DELETE /api/admin/card-definitions/{id}
```

Все admin endpoint защищены сессией и проверкой `is_admin`. Новые аккаунты не
получают административные права автоматически.

Правила используют фиксированный allowlist метрик и операций. Администратор не
может передать произвольный SQL или выполнить код на сервере.

При генерации recap:

1. backend загружает активные определения;
2. вычисляет выбранную статистику;
3. применяет способ анализа и условие;
4. формирует обычный `RecapCard`;
5. сохраняет результат вместе с recap.

Карточки остаются совместимыми с существующим frontend-контрактом. Обзор и
финальная карточка сохраняются, а общий набор ограничен девятью слайдами.

## Почему эта часть важна

Мой вклад закрывает три критических свойства продукта:

1. **Сохранность данных.** Итоги, профили и сессии хранятся атомарно и
   предсказуемо. Без этого генерация recap не могла бы безопасно повторяться.
2. **Доступность.** Redis ускоряет чтение, но не становится единой точкой
   отказа: при его недоступности сервис продолжает работать через PostgreSQL.
3. **Управляемость продукта.** Метрики показывают состояние data layer, а
   конструктор карточек превращает жёстко заданный набор слайдов в управляемый
   механизм продуктовых сценариев.

Таким образом, работа Backend 2 оказалась не вспомогательной обвязкой, а
фундаментом для авторизации, генерации итогов, производительности,
диагностики и дальнейшего развития контента.

## Архитектура

```text
HTTP admin API
      |
      v
Card Definition Service ---> Card Definition Repository ---> PostgreSQL
      |
      v
Generator ---> ProfileMetrics ---> Configured RecapCard

Recap Repository Port
      |
      v
CachedRecapRepository ---> Redis
      |
      +------------------> PostgreSQL source of truth

pgxpool.Stat + cache events ---> Prometheus collectors
```

Слои связаны через интерфейсы портов: use cases не зависят от конкретной базы
данных или Redis-клиента.

## Структура

```text
internal/
  config/                    конфигурация PostgreSQL и Redis
  delivery/
    handlers/                admin API конструктора карточек
    middleware/              authentication и RequireAdmin
  domain/                    account, recap и card definition
  observability/             Prometheus-метрики data layer
  repository/
    postgres/                PostgreSQL adapters
    redis/                   fail-open recap cache
  usecase/
    auth/                    транзакционная регистрация
    carddefinition/          валидация и управление шаблонами
    generator/               расчёт и вставка динамических карточек
    ports/                   repository interfaces
migrations/                  схема и migration runner
tests/unit/                  тесты генератора карточек
pkg/errors/                  доменные ошибки
```

## Локальная проверка

```bash
cp .env.example .env
docker compose up -d
go test ./...
go vet ./...
```

Интеграционный тест транзакции:

```bash
DB_HOST=localhost \
DB_PORT=5446 \
DB_USER=result_year \
DB_PASSWORD=result_year_dev_password \
DB_NAME=result_year \
go test ./internal/repository/postgres -run Transaction -v
```

## Технологии

- Go 1.22
- PostgreSQL 17
- Redis 8
- `pgx/v5`, `pgxpool`
- `go-redis/v9`
- Prometheus client
- Docker Compose

## Командный контекст

Это выделенная часть командного проекта. Доменные модели, repository ports,
базовый генератор встроенных карточек и HTTP helpers включены как необходимый
контекст для компиляции авторских изменений.

В конструкторе моя реализация сосредоточена в `card_definition.go`,
`usecase/carddefinition/`, `configured_cards.go`, admin handler, middleware и
миграции. Основной вклад репозитория — PostgreSQL/Redis adapters,
транзакционная целостность, инфраструктурные метрики и backend-конструктор
recap-карточек.
