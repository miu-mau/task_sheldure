package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"task_shelduler/internal/database"
	"task_shelduler/internal/models"
	"task_shelduler/internal/queue"
	"task_shelduler/internal/repository"
	"task_shelduler/internal/service"
	schedulerv1 "task_shelduler/pkg/proto/v1"
)

func main() {
	var (
		dbPath       = flag.String("db", "internal/migrations/data/task_scheduler.db", "path to SQLite database file")
		port         = flag.Int("port", 50051, "gRPC server port")
		kafkaBrokers = flag.String("kafka-brokers", "localhost:9092", "comma-separated list of Kafka brokers")
		kafkaTopic   = flag.String("kafka-topic", "tasks", "Kafka topic for tasks")
	)
	flag.Parse()

	db, err := database.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Миграции
	if err := database.RunMigrations(db, "internal/migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Репозитории
	taskRepo := repository.NewTaskRepository(db)
	attemptRepo := repository.NewAttemptRepository(db)

	// Kafka продюсер
	brokers := []string{*kafkaBrokers}
	producer := queue.NewKafkaProducer(brokers, *kafkaTopic)
	defer producer.Close()

	// планировщик
	go startSchedulerLoop(taskRepo, producer)

	// gRPC-сервис
	schedulerSvc := service.NewSchedulerService(taskRepo, attemptRepo)

	// gRPC-сервер
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	schedulerv1.RegisterSchedulerServiceServer(grpcServer, schedulerSvc)

	log.Printf("Scheduler gRPC server listening on :%d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

func startSchedulerLoop(taskRepo repository.TaskRepository, producer *queue.KafkaProducer) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := scheduleReadyTasks(ctx, taskRepo, producer)
		cancel()
		if err != nil {
			log.Printf("scheduler loop error: %v", err)
		}
	}
}

func scheduleReadyTasks(ctx context.Context, taskRepo repository.TaskRepository, producer *queue.KafkaProducer) error {
	readyTasks, err := taskRepo.GetReadyTasks(10)
	if err != nil {
		return fmt.Errorf("get ready tasks: %w", err)
	}

	for _, t := range readyTasks {
		payload := map[string]string{
			"task_id":    t.ID,
			"request_id": t.RequestID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("failed to marshal task payload: %v", err)
			continue
		}

		if err := producer.PublishTask(ctx, []byte(t.ID), data); err != nil {
			log.Printf("failed to publish task %s to Kafka: %v", t.ID, err)
			continue
		}

		if err := taskRepo.UpdateTaskStatus(t.ID, models.TaskStatusQueued); err != nil {
			log.Printf("failed to update task %s status to QUEUED: %v", t.ID, err)
			continue
		}
	}

	return nil
}
