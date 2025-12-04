package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"task_shelduler/internal/database"
	"task_shelduler/internal/queue"
	"task_shelduler/internal/repository"
	schedulerv1 "task_shelduler/pkg/proto/v1"
)

func main() {
	var (
		kafkaBrokers = flag.String("kafka-brokers", "localhost:9092", "comma-separated list of Kafka brokers")
		kafkaTopic   = flag.String("kafka-topic", "tasks", "Kafka topic for tasks")
		groupID      = flag.String("group-id", "worker-group", "Kafka consumer group ID")
		grpcAddr     = flag.String("grpc-addr", "localhost:50051", "gRPC server address")
		dbPath       = flag.String("db", "internal/migrations/data/task_scheduler.db", "path to SQLite database file")
		workerID     = flag.String("worker-id", "", "unique worker ID (e.g. worker1, worker2). If empty, it will be assigned automatically.")
	)
	flag.Parse()

	if *workerID == "" {
		db, err := database.OpenDB(*dbPath)
		if err != nil {
			log.Fatalf("Failed to open database for worker registration: %v", err)
		}
		defer db.Close()

		workerRepo := repository.NewWorkerRepository(db)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		id, err := workerRepo.RegisterWorker(ctx, "worker")
		if err != nil {
			log.Fatalf("Failed to register worker: %v", err)
		}
		*workerID = id
	}

	brokers := parseBrokers(*kafkaBrokers)
	consumer := queue.NewKafkaConsumer(brokers, *kafkaTopic, *groupID)
	defer consumer.Close()

	conn, err := grpc.NewClient(*grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := schedulerv1.NewSchedulerServiceClient(conn)

	log.Printf("Worker started. ID=%s, topic=%s", *workerID, *kafkaTopic)

	ctx := context.Background()
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Printf("Failed to read message from Kafka: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if err := processTask(ctx, client, msg.Value, *workerID); err != nil {
			log.Printf("Failed to process task: %v", err)

			continue
		}

		log.Printf("Task processed successfully: key=%s", string(msg.Key))
	}
}

func processTask(ctx context.Context, client schedulerv1.SchedulerServiceClient, data []byte, workerID string) error {

	var payload map[string]string
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", err)
	}

	taskID := payload["task_id"]
	requestID := payload["request_id"]
	targetWorkerID := payload["worker_id"]

	if taskID == "" {
		return fmt.Errorf("task_id is empty")
	}

	if workerID != "" && targetWorkerID != "" && targetWorkerID != workerID {
		log.Printf("Skip task %s: target_worker_id=%s, worker_id=%s", taskID, targetWorkerID, workerID)
		return nil
	}

	log.Printf("Processing task: %s (request_id: %s)", taskID, requestID)

	_, err := client.UpdateTaskStatus(ctx, &schedulerv1.UpdateTaskStatusRequest{
		TaskId:    taskID,
		Status:    schedulerv1.TaskStatus_TASK_STATUS_RUNNING,
		RequestId: requestID,
	})
	if err != nil {
		return fmt.Errorf("failed to update task status to RUNNING: %w", err)
	}

	success := executeTask(taskID)

	attemptStatus := schedulerv1.AttemptStatus_ATTEMPT_STATUS_SUCCESS
	errorMsg := ""
	if !success {
		attemptStatus = schedulerv1.AttemptStatus_ATTEMPT_STATUS_FAILED
		errorMsg = "Task execution failed"
	}

	_, err = client.ReportAttempt(ctx, &schedulerv1.ReportAttemptRequest{
		TaskId:    taskID,
		Status:    attemptStatus,
		Error:     errorMsg,
		RequestId: requestID,
	})
	if err != nil {
		return fmt.Errorf("failed to report attempt: %w", err)
	}

	return nil
}

func executeTask(taskID string) bool {

	time.Sleep(100 * time.Millisecond)

	log.Printf("Executing task: %s", taskID)

	// демо возвращаем успех

	return true
}

func parseBrokers(brokersStr string) []string {
	brokers := strings.Split(brokersStr, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}
	return brokers
}
