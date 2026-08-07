# Avito Recap — Backend Data Layer

Портфолио-репозиторий с моей зоной ответственности в командном проекте
[GoOffer Hackathon Avito](https://github.com/NikName2021/GoOffer_HackathonAvito).

Здесь выделен Backend 2: PostgreSQL, Redis, миграции, репозитории и обеспечение
целостности данных. Frontend и HTTP delivery-слой исходного проекта намеренно
не включены.

## Моя зона ответственности

- проектирование и развитие схемы PostgreSQL;
- SQL-миграции `up/down` и тестовые данные;
- реализация repository adapters через `pgx/v5`;
- пул соединений PostgreSQL;
- UPSERT итогов года по `(user_id, year)`;
- Redis-кэш для готовых recap;
- транзакционная регистрация аккаунта и первой сессии;
- преобразование ошибок PostgreSQL в доменные ошибки;
- интеграционные и unit-тесты слоя данных;
- Docker-инфраструктура PostgreSQL и Redis.

## Что реализовано

### PostgreSQL

Репозитории:

- `UserRepository` — CRUD профилей с изоляцией по `account_id`;
- `ActionRepository` — выборка действий пользователя за год;
- `RecapRepository` — сохранение и чтение итогов года;
- `AuthRepository` — аккаунты и серверные сессии.

Для итогов года используется атомарный UPSERT:

```sql
ON CONFLICT (user_id, year) DO UPDATE
```

Это позволяет безопасно повторно генерировать recap без дублирования строк.

### Транзакции

Создание аккаунта и первой сессии выполняется в одной транзакции:

```text
BEGIN
  INSERT accounts
  INSERT sessions
COMMIT
```

Если сессия не создаётся, выполняется `ROLLBACK`, поэтому в базе не остаётся
аккаунт без рабочей сессии.

Одиночные `INSERT`, `UPDATE`, `DELETE` и UPSERT дополнительно в транзакции не
оборачиваются: один SQL statement уже атомарен в PostgreSQL.

### Redis

Для recap реализован read-through/write-through cache:

```text
GET recap:
  Redis hit  -> вернуть значение
  Redis miss -> PostgreSQL -> записать в Redis -> вернуть значение

SAVE recap:
  PostgreSQL -> обновить Redis
```

- TTL: 24 часа;
- ключ: `recap:v1:{user_id}:{year}`;
- PostgreSQL остаётся источником истины;
- при недоступности Redis запрос обслуживается через PostgreSQL;
- повреждённая запись кэша удаляется и восстанавливается из БД.

### Миграции

Миграции встроены в Go-бинарник через `embed.FS` и применяются в
лексикографическом порядке. Каждая миграция выполняется:

- под PostgreSQL advisory lock;
- в отдельной транзакции;
- с записью версии в `schema_migrations`.

Схема включает профили, категории, действия, recap, аккаунты, сессии и данные
для расширенной аналитики.

## Архитектура

```text
Use cases
    |
    v
Repository ports
    |
    +--> PostgreSQL adapters --> PostgreSQL
    |
    +--> CachedRecapRepository --> Redis
                 |
                 +-------------> PostgreSQL fallback
```

Репозитории зависят от доменных контрактов, а бизнес-логика не зависит от
конкретной базы данных.

## Структура

```text
internal/
  config/                 конфигурация PostgreSQL и Redis
  domain/                 доменные структуры, необходимые для контрактов
  repository/
    postgres/             PostgreSQL adapters
    redis/                Redis cache и cached repository decorator
  usecase/
    ports/                интерфейсы репозиториев
    auth/                 сценарий регистрации с атомарной записью
migrations/               SQL up/down и встроенный migration runner
pkg/errors/               доменные ошибки
```

## Локальный запуск инфраструктуры

```bash
cp .env.example .env
docker compose up -d
```

Проверка:

```bash
docker compose ps
docker compose exec postgres pg_isready -U result_year -d result_year
docker compose exec redis redis-cli ping
```

## Тесты

```bash
go test ./...
go vet ./...
```

Для интеграционного теста транзакции:

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
- Docker Compose

## Командный контекст

Этот репозиторий является выделенной частью командного проекта. Доменные модели
и repository ports включены как необходимые контракты для сборки и демонстрации
слоя данных. Мой основной вклад — содержимое `repository/`, миграции,
транзакционная целостность, Redis-кэш и инфраструктура хранения данных.
