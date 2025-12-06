package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"task_shelduler/internal/database"
	"task_shelduler/internal/models"
	"task_shelduler/internal/queue"
	"task_shelduler/internal/repository"
	"task_shelduler/internal/service"
	schedulerv1 "task_shelduler/pkg/proto/v1"
)

func main() {
	var (
		dbPath          = flag.String("db", "internal/migrations/data/task_scheduler.db", "path to SQLite database file")
		port            = flag.Int("port", 50051, "gRPC server port")
		kafkaBrokers    = flag.String("kafka-brokers", "localhost:9092", "comma-separated list of Kafka brokers")
		kafkaTopic      = flag.String("kafka-topic", "tasks", "Kafka topic for tasks")
		kafkaPartitions = flag.Int("kafka-partitions", 10, "number of Kafka topic partitions (for load distribution across workers)")
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
	brokers := parseBrokers(*kafkaBrokers)

	if err := queue.EnsureTopic(brokers, *kafkaTopic, *kafkaPartitions); err != nil {
		log.Printf("Warning: failed to ensure topic exists (may already exist): %v", err)
	}

	producer := queue.NewKafkaProducer(brokers, *kafkaTopic)
	defer producer.Close()

	// планировщик
	go startSchedulerLoop(taskRepo, producer, brokers)

	// gRPC-сервис
	schedulerSvc := service.NewSchedulerService(taskRepo, attemptRepo)

	// gRPC-сервер
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	schedulerv1.RegisterSchedulerServiceServer(grpcServer, schedulerSvc)

	reflection.Register(grpcServer)

	log.Printf("Scheduler gRPC server listening on :%d", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

func startSchedulerLoop(taskRepo repository.TaskRepository, producer *queue.KafkaProducer, brokers []string) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := scheduleReadyTasks(ctx, taskRepo, producer, brokers)
		cancel()
		if err != nil {
			log.Printf("scheduler loop error: %v", err)
		}
	}
}

func scheduleReadyTasks(ctx context.Context, taskRepo repository.TaskRepository, producer *queue.KafkaProducer, brokers []string) error {
	readyTasks, err := taskRepo.GetReadyTasks(10)
	if err != nil {
		return fmt.Errorf("get ready tasks: %w", err)
	}

	if len(readyTasks) == 0 {

		return nil
	}

	log.Printf("Found %d ready task(s) to schedule", len(readyTasks))

	baseTopic := "tasks"
	createdTopics := make(map[string]bool)

	for _, t := range readyTasks {
		log.Printf("Scheduling task: %s (payload: %s, worker_id: %s)", t.ID, t.Payload, t.WorkerID)

		payload := map[string]string{
			"task_id":    t.ID,
			"request_id": t.RequestID,
			"worker_id":  t.WorkerID,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			log.Printf("failed to marshal task payload: %v", err)
			continue
		}

		var targetTopic string
		partitionKey := t.ID

		if t.WorkerID != "" {
			targetTopic = fmt.Sprintf("%s-worker-%s", baseTopic, t.WorkerID)
			partitionKey = t.WorkerID
		} else {
			targetTopic = baseTopic
		}

		if !createdTopics[targetTopic] {
			if err := queue.EnsureTopic(brokers, targetTopic, 1); err != nil {
				log.Printf("Warning: failed to ensure topic %s exists: %v", targetTopic, err)
			} else {
				createdTopics[targetTopic] = true
			}
		}

		if err := producer.PublishTaskToTopic(ctx, targetTopic, []byte(partitionKey), data); err != nil {
			log.Printf("failed to publish task %s to Kafka topic %s: %v", t.ID, targetTopic, err)
			continue
		}

		if err := taskRepo.UpdateTaskStatus(t.ID, models.TaskStatusQueued); err != nil {
			log.Printf("failed to update task %s status to QUEUED: %v", t.ID, err)
			continue
		}

		log.Printf("Task %s successfully published to topic %s and marked as QUEUED", t.ID, targetTopic)
	}

	return nil
}

func parseBrokers(brokersStr string) []string {
	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return brokers
}
