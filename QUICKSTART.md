# Быстрый старт

## Базовый функционал готов!

### Что реализовано:

1. ✅ **Репозитории** - работа с SQLite (tasks, schedules, attempts)
2. ✅ **gRPC сервис** - CreateTask, GetTask, ListTasks, UpdateTaskStatus, ReportAttempt
3. ✅ **Kafka интеграция** - продюсер и консьюмер
4. ✅ **Планировщик** - автоматически берёт готовые задачи и отправляет в Kafka
5. ✅ **Воркер** - читает задачи из Kafka и выполняет их

## Запуск

### 1. Подготовка

```bash
# Установить зависимости
go mod download

# Применить миграции
make migrate
```

### 2. Запустить Kafka

```bash
# Запустить Kafka через docker-compose
make docker-up
# или
docker-compose up -d

# Kafka будет доступна на localhost:9092
# Проверить логи: make docker-logs
```

### 3. Запустить планировщик (gRPC сервер + планировщик)

```bash
go run ./cmd/app
# или
go run ./cmd/app -port 50051 -kafka-brokers localhost:9092 -kafka-topic tasks
```

### 4. Запустить воркер(ы)

В отдельном терминале:

```bash
go run ./cmd/worker
# или
go run ./cmd/worker -kafka-brokers localhost:9092 -kafka-topic tasks -group-id worker-group -grpc-addr localhost:50051
```

**Можно запустить любое количество воркеров** - Kafka автоматически распределит партиции между всеми воркерами в consumer group:
- Если воркеров меньше чем партиций - каждый воркер обработает несколько партиций
- Если воркеров больше чем партиций - лишние воркеры будут простаивать (это нормально)
- Kafka автоматически перераспределит партиции при добавлении/удалении воркеров

По умолчанию топик создаётся с 10 партициями (можно изменить через `-kafka-partitions`).

## Тестирование

### Создать задачу через Go CLI:

```bash
# Создать задачу
go run ./cmd/cli -cmd create "test task"

# Создать задачу с request_id
go run ./cmd/cli -cmd create "my task" "req-123"

# Получить задачу по ID
go run ./cmd/cli -cmd get <task_id>

# Список задач
go run ./cmd/cli -cmd list 10 0

# Список задач с фильтром по статусу (2 = QUEUED)
go run ./cmd/cli -cmd list 20 0 2
```

**Статусы задач:**
- 0 = UNSPECIFIED
- 1 = DRAFT
- 2 = QUEUED
- 3 = RUNNING
- 4 = SUCCESS
- 5 = FAILED
- 6 = CANCELLED

**См. также:** `scripts/test_commands.md` для подробных инструкций

## Как это работает:

1. **Создание задачи**: `CreateTask` → сохраняется в БД со статусом `DRAFT`
2. **Планировщик**: каждые 5 секунд берёт готовые задачи (`DRAFT`, `scheduled_at <= now`) и отправляет в Kafka, меняя статус на `QUEUED`
3. **Воркер**: читает из Kafka, обновляет статус на `RUNNING`, "выполняет" задачу, вызывает `ReportAttempt` → статус становится `SUCCESS` или `FAILED`

## Структура проекта:

```
cmd/
  app/        - планировщик (gRPC сервер + планировщик)
  worker/     - воркер (читает из Kafka и выполняет задачи)

internal/
  database/   - подключение к SQLite и миграции
  models/     - модели данных
  repository/ - репозитории для работы с БД
  service/    - gRPC сервис
  queue/      - Kafka продюсер и консьюмер
```

