# Быстрый запуск

скрипт для быстрого запуска проекта

запуск docker контейнеров
```bash
make docker-up
```

запуск миграций в бд
```bash
make migrate
```

запуск сервиса для приема задач
```bash
go run ./cmd/app
```

запуск воркера для выполнения задач
```bash
go run ./cmd/worker
```

создание задач через cli
```bash
go run ./cmd/cli -cmd create "for all workers"

go run ./cmd/cli -cmd create "for all workers2"

go run ./cmd/cli -cmd create "for worker 1" -worker-id worker1

go run ./cmd/cli -cmd create "for worker 2" -worker-id worker2
```