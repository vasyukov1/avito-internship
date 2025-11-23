# Pull Request Reviewer Service

Автор: Васюков Александр

Pull Request Reviewer Service — это микросервис, отвечающий за:
- создание Pull Request’ов,
- автоматическое назначение ревьюеров,
- пересоздание и переназначение ревьюеров,
- просмотр списка PR по пользователю,
- пометку PR как merged,
- хранение данных в PostgreSQL и корректную работу под нагрузкой.

Сервис реализован на Go, использует Gin, pgx, транзакции и round-robin балансировку при выборе ревьюеров.

---

### 🛠 Функционал

1. Создание Pull Request
- `POST /pullRequest/create`
- создаёт новый PR,
- сохраняет автора, имя и статус,
- автоматически назначает ревьюеров,
- выбор ревьюеров – через round-robin, реализованный на уровне БД.

2. Назначение ревьюеров  
Во время создания PR сервис:
- Поднимает транзакцию.
- Блокирует таблицу `reviewer_pointer` через `SELECT ... FOR UPDATE`.
- Выбирает следующих ревьюеров согласно индексу.
- Обновляет указатель.
- Записывает ревьюеров в `pull_request_shorts`.

3. Переназначение ревьюера
- `POST /pullRequest/reassign`
- Позволяет заменить одного ревьюера другим:
- выполняется через транзакцию,
- удаляет старого ревьюера,
- добавляет нового,
- возвращает актуальное состояние PR.

4. Получение PR текущего пользователя

- `GET /pullRequest/getReview`
- Возвращает список PR, где пользователь назначен ревьюером,
- данные агрегируются через `JOIN pull_request_shorts`.

5. Пометка PR как Merged
- `POST /pullRequest/merge`
- идемпотентная операция,
- повторный вызов всегда возвращает корректное финальное состояние PR.
- если PR уже в статусе merged, повторный вызов просто вернёт его как merged.

---

### 🏛 Архитектура
```md
delivery/        → HTTP-контроллеры (Gin)
usecase/         → бизнес-процессы
repository/      → работа с PostgreSQL (pgx, транзакции)
domain/          → модели
migrations/      → SQL-миграции
```
---

### 🏠 Структура проекта

```go
.
├── cmd/
│   └── server/
│       └── main.go
├── docs/
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── delivery/
│   │   ├── error_handler.go
│   │   ├── handler.go
│   │   ├── health_handler.go
│   │   ├── middleware.go
│   │   ├── pr_handler.go
│   │   ├── router.go
│   │   ├── team_handler.go
│   │   └── user_handler.go
│   ├── domain/
│   │   ├── errors.go
│   │   ├── pull_request.go
│   │   ├── team.go
│   │   └── user.go
│   ├── logger/
│   │   └── logger.go
│   ├── metrics/
│   │   └── metrics.go
│   ├── repository/
│   │   ├── pgstorage.go
│   │   ├── pr_repo.go
│   │   ├── team_repo.go
│   │   └── user_repo.go
│   └── usecase/
│       └── server.go
├── migrations/
├── monitoring/
│   ├── grafana/
│   └── prometheus/
├── .env.example
├── .girignore
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── openapi.yml
└── README.md
```

---

### 🗄 Применяемые паттерны

1. Транзакционный репозиторий

В критичных местах (AssignReviewers, Reassign, Merge) используется транзакция уровня pgx.

2. Round-robin через БД

Используется таблица:

```sql
CREATE TABLE reviewer_pointer (
    id          INTEGER PRIMARY KEY CHECK (id = 1),
    last_index  INTEGER NOT NULL
);
```

Преимущества:
- гарантированная последовательность,
- отсутствие конкурентных конфликтов,
- корректность при большом числе параллельных запросов.

3. ON CONFLICT DO NOTHING

Устраняет дубликаты ревьюеров при повторных вызовах.

4. Идемпотентность

Merge-операция повторяемая без сайд-эффектов.

---

### ⚙ Конфигурация

Через переменные окружения:

```bash
DB_HOST=  
DB_PORT=  
DB_USER=  
DB_PASSWORD=  
DB_NAME=  
LOG_LEVEL=  
```
---

## Нагрузочное тестирование

Для проверки стабильности сервиса были проведены три вида нагрузочных тестов с использованием k6:

1. Создание Pull Request (100 VU, 30s)
2. Параллельный merge для одного PR (идемпотентность)

Все тесты находятся в директории `tests/`.

Чтобы запустить, надо установить k6:
```bash
brew install k6
```

- Запуск тестов:
```bash
k6 run tests/pr_create.js
k6 run tests/merge_parallel.js
```

---

### 🚀 Запуск


Запуск через Makefile
```bash
make full-up
```