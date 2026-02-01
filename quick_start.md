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

go run ./cmd/cli -cmd create "fail" -worker-id worker1
```

отправка задачи в конкретное время 
```bash
go run ./cmd/cli -cmd create "run at 19:00" -at 19:00
go run ./cmd/cli -cmd create "run at 19:30" -at 19:30
go run ./cmd/cli -cmd create "run at 19:00:00" -at 19:00:00

go run ./cmd/cli -cmd create "only worker1 at 19" -at 19:00 -worker-id worker1
go run ./cmd/cli -cmd create "only worker2 at 20:30" -at 20:30 -worker-id worker2
```

