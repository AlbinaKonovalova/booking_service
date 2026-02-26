# booking_service

Implementation of domain-driven design logic

---

## Архитектура

Clean Architecture + DDD Lite (Domain First):

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Handlers                          │
│                   (adapters/http/handlers)                  │
├─────────────────────────────────────────────────────────────┤
│                    Application Services                     │
│                      (application/)                         │
├─────────────────────────────────────────────────────────────┤
│                      Domain Layer                           │
│                       (domain/)                             │
├─────────────────────────────────────────────────────────────┤
│              Ports (Interfaces)    │    Adapters            │
│                 (ports/)           │  (adapters/repository) │
└─────────────────────────────────────────────────────────────┘
```

---

## Структура проекта

```
booking_service/
├── cmd/
│   └── server/
│       └── main.go                 # Точка входа
├── internal/
│   ├── domain/                     # Бизнес-логика (ядро)
│   ├── application/                # Use Cases
│   ├── ports/
│   │   ├── input/                  # Входящие порты
│   │   └── output/                 # Исходящие порты (репозитории)
│   ├── adapters/
│   │   ├── http/
│   │   │   ├── server.go           # HTTP сервер + middleware
│   │   │   └── handlers/           # HTTP хендлеры
│   │   ├── repository/
│   │   │   └── postgres/           # PostgreSQL реализация репозиториев
│   │   └── scheduler/              # Фоновый scheduler (expire + complete)
│   └── config/
│       └── config.go               # Парсинг YAML + ENV
├── config/
│   └── config.yaml
├── migrations/                     # Goose миграции
├── docker-compose.yml
├── Dockerfile
├── Dockerfile.migrate
├── Makefile
├── go.mod
└── README.md
```

---

## Быстрый старт

### Требования
- Go 1.24+
- PostgreSQL 15+
- Docker & Docker Compose (опционально)

### Запуск с Docker (рекомендуется)

```bash
git clone https://github.com/AlbinaKonovalova/booking_service.git
cd booking_service

# Запустить всё одной командой (БД + миграции + сервер)
docker-compose up -d --build

# Проверить что сервер работает
curl http://localhost:8080/health
```

При запуске автоматически поднимается PostgreSQL, применяются миграции (goose) и запускается сервер на порту 8080.

### Запуск локально

```bash
git clone https://github.com/AlbinaKonovalova/booking_service.git
cd booking_service

# Создать базу данных
createdb -U postgres -h localhost booking_service

# Установить goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Применить миграции
make migrate-up

# Запустить сервер
make run
```

---

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/booking_service?sslmode=disable` |
| `SERVER_PORT` | Порт HTTP сервера | `8080` |
| `SERVER_READ_TIMEOUT` | Таймаут чтения запроса | `10s` |
| `SERVER_WRITE_TIMEOUT` | Таймаут записи ответа | `10s` |
| `SERVER_SHUTDOWN_TIMEOUT` | Таймаут graceful shutdown | `5s` |
| `DB_MAX_OPEN_CONNS` | Макс. открытых соединений к БД | `25` |
| `DB_MAX_IDLE_CONNS` | Макс. idle соединений к БД | `5` |
| `DB_CONN_MAX_LIFETIME` | Время жизни соединения | `5m` |
| `HOTEL_TIMEZONE` | Таймзона отеля (IANA) | `UTC` |
| `LOG_LEVEL` | Уровень логирования (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Формат логов | `json` |
| `EXPIRATION_INTERVAL` | Интервал проверки просроченных бронирований | `5m` |
| `COMPLETION_TIME` | Время ежедневного завершения бронирований (HH:MM UTC) | `11:50` |

---

## API Endpoints

### Создание ресурса — POST /resource

Создаёт новый ресурс (номер в отеле).

```bash
curl -X POST http://localhost:8080/resource \
  -H 'Content-Type: application/json' \
  -d '{"name":"Room A"}'
```

Response `201 Created`:
```json
{
  "id": "6b74a2e7-8813-4322-9db5-69e9e2bf060e",
  "name": "Room A",
  "created_at": "2026-02-25T08:46:10.543269Z"
}
```

Ошибки:

| Статус | Код | Описание |
|--------|-----|----------|
| `400` | `BAD_REQUEST` | Невалидный JSON |
| `422` | `INVALID_INPUT` | Пустое имя ресурса |

### Создание бронирования — POST /booking

Создаёт бронирование ресурса на временной интервал.

Бизнес-правила:
- `start_time = check_in`
- Вычисляется день периода `D`: если время >= 12:00 — D = сегодня, иначе D = вчера
- Окно заезда: `D 12:00 <= start_time < (D+1) 02:00`
- `end_time = (D+N) 12:00` — минимальный N (1..365), чтобы `check_out <= end_time`
- Проверяется отсутствие пересечений с активными бронями (CREATED, CONFIRMED)
- Стык разрешён: `existing.end == new.start` — это не пересечение

```bash
curl -X POST http://localhost:8080/booking \
  -H 'Content-Type: application/json' \
  -d '{
    "resource_id": "<uuid>",
    "check_in": "2026-02-25T14:00:00Z",
    "check_out": "2026-02-26T08:00:00Z"
  }'
```

Response `201 Created`:
```json
{
  "id": "<uuid>",
  "resource_id": "<uuid>",
  "start_time": "2026-02-25T14:00:00Z",
  "end_time": "2026-02-26T12:00:00Z",
  "check_in": "2026-02-25T14:00:00Z",
  "check_out": "2026-02-26T08:00:00Z",
  "status": "CREATED",
  "created_at": "2026-02-25T08:50:00Z"
}
```

`end_time` всегда на границе 12:00 UTC. Если `check_out` > 12:00, бронируется дополнительный период (N=2).

Ошибки:

| Статус | Код | Описание |
|--------|-----|----------|
| `400` | `BAD_REQUEST` | Невалидный JSON / формат UUID / формат времени |
| `404` | `NOT_FOUND` | Ресурс не найден |
| `409` | `BOOKING_OVERLAP` | Пересечение с существующей бронью |
| `410` | `ALREADY_REMOVED` | Ресурс удалён |
| `422` | `BOOKING_IN_PAST` | Бронирование в прошлом |
| `422` | `BOOKING_NOT_AVAILABLE` | check_in вне допустимого окна заезда |
| `422` | `INVALID_TIME_RANGE` | check_in >= check_out |
| `422` | `BOOKING_TOO_LONG` | Длительность > 365 периодов |

### Подтверждение бронирования — POST /booking/{id}/confirm

Переводит бронирование из CREATED в CONFIRMED.

Если `now > start_time` и статус CREATED — бронь автоматически переводится в EXPIRED, подтверждение запрещается.

```bash
curl -X POST http://localhost:8080/booking/<booking_id>/confirm
```

Response `200 OK`:
```json
{
  "id": "<uuid>",
  "status": "CONFIRMED"
}
```

Ошибки:

| Статус | Код | Описание |
|--------|-----|----------|
| `400` | `BAD_REQUEST` | Невалидный формат UUID |
| `404` | `NOT_FOUND` | Бронирование не найдено |
| `409` | `INVALID_STATUS_TRANSITION` | Статус не CREATED |
| `409` | `BOOKING_EXPIRED` | Бронирование автоматически истекло |

### Отмена бронирования — POST /booking/{id}/cancel

Отменяет бронирование из статусов CREATED или CONFIRMED.

Если `now > start_time` и статус CREATED — бронь автоматически переводится в EXPIRED, отмена запрещается.

```bash
curl -X POST http://localhost:8080/booking/<booking_id>/cancel
```

Response `200 OK`:
```json
{
  "id": "<uuid>",
  "status": "CANCELLED"
}
```

Ошибки:

| Статус | Код | Описание                                |
|--------|-----|-----------------------------------------|
| `400` | `BAD_REQUEST` | Невалидный формат UUID                  |
| `404` | `NOT_FOUND` | Бронирование не найдено                 |
| `409` | `INVALID_STATUS_TRANSITION` | Статус EXPIRED, CANCELLED или COMPLETED |
| `409` | `BOOKING_EXPIRED` | Бронирование автоматически истекло      |

### Удаление ресурса — DELETE /resource/{id}

Soft delete ресурса. Запрещено если есть активные бронирования (CREATED / CONFIRMED).

```bash
curl -X DELETE http://localhost:8080/resource/<resource_id>
```

Response `200 OK`:
```json
{
  "id": "<uuid>",
  "status": "removed"
}
```

Ошибки:

| Статус | Код | Описание |
|--------|-----|----------|
| `400` | `BAD_REQUEST` | Невалидный формат UUID |
| `404` | `NOT_FOUND` | Ресурс не найден |
| `409` | `HAS_ACTIVE_BOOKINGS` | Есть активные бронирования |
| `410` | `ALREADY_REMOVED` | Ресурс уже удалён |

### Список бронирований ресурса — GET /resource/{id}/bookings

Возвращает список бронирований ресурса (в том числе для удалённых ресурсов).
Опциональная фильтрация по статусу через query-параметр `?status=`.

```bash
curl http://localhost:8080/resource/<resource_id>/bookings
curl http://localhost:8080/resource/<resource_id>/bookings?status=CONFIRMED
```

Response `200 OK`:
```json
[
  {
    "id": "<uuid>",
    "resource_id": "<uuid>",
    "start_time": "2026-02-25T14:00:00Z",
    "end_time": "2026-02-26T12:00:00Z",
    "check_in": "2026-02-25T14:00:00Z",
    "check_out": "2026-02-26T08:00:00Z",
    "status": "CREATED",
    "created_at": "2026-02-25T08:50:00Z"
  }
]
```

Ошибки:

| Статус | Код | Описание |
|--------|-----|----------|
| `400` | `BAD_REQUEST` | Невалидный формат UUID |
| `404` | `NOT_FOUND` | Ресурс не найден |

---

## Автоматическое истечение и завершение

Фоновый scheduler выполняет два действия:

1. **CREATED -> EXPIRED** — если `now > start_time` и бронь в статусе CREATED (по интервалу, по умолчанию каждые 5 минут)
2. **CONFIRMED -> COMPLETED** — если `now >= end_time` и бронь в статусе CONFIRMED (ежедневно в настраиваемое время, по умолчанию 11:50 UTC)

### Статусы бронирования

| Статус | Описание | Блокирует ресурс |
|--------|----------|-----------------|
| `CREATED` | Создано, ожидает подтверждения | да |
| `CONFIRMED` | Подтверждено | да |
| `CANCELLED` | Отменено | нет |
| `EXPIRED` | Истекло (не подтверждено вовремя) | нет |
| `COMPLETED` | Завершено (период проживания окончен) | нет |

### Допустимые переходы

```
CREATED   -> CONFIRMED  (confirm)
CREATED   -> CANCELLED  (cancel)
CREATED   -> EXPIRED    (авто-экспайр)
CONFIRMED -> CANCELLED  (cancel)
CONFIRMED -> COMPLETED  (авто-завершение)
```

---

## Make-команды

```bash
make run            # Запуск сервера локально
make build          # Сборка бинарника
make test           # Запуск тестов
make lint           # Линтинг кода
make migrate-up     # Применить миграции
make migrate-down   # Откатить миграции
make migrate-status # Статус миграций
make docker-run     # Собрать и запустить через docker-compose
make docker-down    # Остановить docker-compose
make docker-logs    # Логи приложения
make docker-reset   # Сбросить volumes и пересобрать
make clean          # Очистить артефакты сборки
```

---

## База данных

### Таблица resources

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | `UUID PK` | Уникальный идентификатор |
| `name` | `TEXT NOT NULL` | Название ресурса |
| `created_at` | `TIMESTAMPTZ NOT NULL` | Дата создания |
| `removed_at` | `TIMESTAMPTZ` | Дата удаления (soft delete) |

### Таблица bookings

| Колонка | Тип | Описание |
|---------|-----|----------|
| `id` | `UUID PK` | Уникальный идентификатор |
| `resource_id` | `UUID FK -> resources` | Ресурс бронирования |
| `start_time` | `TIMESTAMPTZ NOT NULL` | Начало блокировки (= check_in) |
| `end_time` | `TIMESTAMPTZ NOT NULL` | Конец блокировки (граница периода 12:00) |
| `check_in` | `TIMESTAMPTZ NOT NULL` | Желаемый заезд |
| `check_out` | `TIMESTAMPTZ NOT NULL` | Желаемый выезд |
| `status` | `TEXT NOT NULL` | Статус бронирования |
| `created_at` | `TIMESTAMPTZ NOT NULL` | Дата создания |


---

## Примечания

- **DDD Lite** — домен не зависит от фреймворков и инфраструктуры, вся бизнес-логика в `internal/domain/`
- **Clean Architecture** — зависимости направлены внутрь (adapters -> application -> domain)
- **Транзакции** — `TxManager` абстрагирует `database/sql.Tx`, application layer не знает о SQL
- **Конкурентность** — `SELECT ... FOR UPDATE` при создании/подтверждении/отмене бронирований
- **Soft delete** — ресурсы помечаются `removed_at`, не удаляются физически
- **Авто-экспайр** — фоновый scheduler переводит просроченные CREATED -> EXPIRED и завершённые CONFIRMED -> COMPLETED
