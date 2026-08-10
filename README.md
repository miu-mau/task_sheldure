# Project Title

*Design and Development of a Distributed Business Process Task Scheduler for the Technical Department of a Financial Organization*

## Brief Explanation of the Topic:

The scheduler receives tasks from the user or other services via gRPC. It then stores them in a database, and then, based on a condition or, for example, a scheduled task, it sends them to a queue. Services then take tasks from this queue and execute them, changing the task status upon completion.

## Scheduler Implementation Tools

*Language and Environment*: Go (Golang) — the primary development language.

*Service Interaction*: gRPC

*Data Storage*: SQLite (used for demonstration and migration via goose).

*Message Queue*: Apache Kafka (library github.com/segmentio/kafka-go).

*Task Scheduling*: Background loop with ticker (checked every 5 seconds) for processing tasks by scheduled_at. Cron scheduling support via the schedules table (implementation in development).

*Monitoring and Logging*: Built-in Go log package (structured logging).

*Testing*: Testing in Go. gRPC can be tested via Go CLI clients or Postman. (in development)

*Infrastructure*: Docker/docker-compose




# Дипломное название проекта 

*Проектирование и разработка распределённого планировщика задач бизнес-процессов технического отдела финансовой организации*

## краткое обьяснение темы:

Планировщик принимает задачи из вне через gRPC от пользователя или от других сервисов. Потом он сохраняет их в бд, потом по условию или, например, по условному расписанию он их отправляет в очередь. Уже потом сервисы с этой очереди забирают себе задачи  и выполняют их меняя статус задания после завершения.

## Инструменты для реализации планировщика


*Язык и среда*: Go (Golang) — основной язык разработки.


*Взаимодействие сервисов*: gRPC


*Хранение данных*: SQLite (используется для демонстрации, миграции через goose).


*Очередь сообщений*: Apache Kafka (библиотека github.com/segmentio/kafka-go).


*Планирование задач*: Фоновый цикл с тикером (проверка каждые 5 секунд) для обработки задач по scheduled_at. Поддержка cron-расписаний через таблицу schedules (реализация в разработке).


*Мониторинг и логирование*: Встроенный log пакет Go (структурированное логирование).


*Тестирование*: testing в Go. gRPC можно тестировать через Go CLI-клиенты или Postman. (в разработке)


*Инфраструктура*: Docker/docker-compose
