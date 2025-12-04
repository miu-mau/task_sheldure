package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	schedulerv1 "task_shelduler/pkg/proto/v1"
)

func main() {
	var (
		grpcAddr = flag.String("addr", "localhost:50051", "gRPC server address")
		command  = flag.String("cmd", "", "command: create, get, list")
		workerID = flag.String("worker-id", "", "target worker ID for created tasks (e.g. worker1)")
	)
	flag.Parse()

	conn, err := grpc.NewClient(*grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := schedulerv1.NewSchedulerServiceClient(conn)
	ctx := context.Background()

	switch *command {
	case "create":
		createTask(ctx, client, *workerID, flag.Args())
	case "get":
		getTask(ctx, client, flag.Args())
	case "list":
		listTasks(ctx, client, flag.Args())
	default:
		printUsage()
		os.Exit(1)
	}
}

func createTask(ctx context.Context, client schedulerv1.SchedulerServiceClient, workerIDFlag string, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: cli -cmd create <payload> [worker_id] [request_id]")
	}

	payload := args[0]
	workerID := workerIDFlag
	requestID := fmt.Sprintf("req-%d", time.Now().Unix())

	if len(args) >= 2 {
		workerID = args[1]
	}

	if len(args) >= 3 {
		requestID = args[2]
	}

	req := &schedulerv1.CreateTaskRequest{
		Payload:   payload,
		RequestId: requestID,
		WorkerId:  workerID,

		ScheduledAt: timestamppb.Now(),
	}

	resp, err := client.CreateTask(ctx, req)
	if err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}

	task := resp.GetTask()
	fmt.Println("✓ Task created successfully!")
	fmt.Printf("  ID: %s\n", task.GetId())
	fmt.Printf("  Payload: %s\n", task.GetPayload())
	fmt.Printf("  Status: %s\n", task.GetStatus().String())
	fmt.Printf("  Request ID: %s\n", task.GetRequestId())
	fmt.Printf("  Created: %s\n", task.GetCreatedAt().AsTime().Format(time.RFC3339))
}

func getTask(ctx context.Context, client schedulerv1.SchedulerServiceClient, args []string) {
	if len(args) < 1 {
		log.Fatal("Usage: cli -cmd get <task_id>")
	}

	taskID := args[0]
	req := &schedulerv1.GetTaskRequest{Id: taskID}

	resp, err := client.GetTask(ctx, req)
	if err != nil {
		log.Fatalf("Failed to get task: %v", err)
	}

	task := resp.GetTask()
	printTask(task)
}

func listTasks(ctx context.Context, client schedulerv1.SchedulerServiceClient, args []string) {
	limit := int32(10)
	offset := int32(0)
	status := schedulerv1.TaskStatus_TASK_STATUS_UNSPECIFIED

	if len(args) >= 1 {
		fmt.Sscanf(args[0], "%d", &limit)
	}
	if len(args) >= 2 {
		fmt.Sscanf(args[1], "%d", &offset)
	}
	if len(args) >= 3 {
		var statusInt int
		fmt.Sscanf(args[2], "%d", &statusInt)
		status = schedulerv1.TaskStatus(statusInt)
	}

	req := &schedulerv1.ListTasksRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}

	resp, err := client.ListTasks(ctx, req)
	if err != nil {
		log.Fatalf("Failed to list tasks: %v", err)
	}

	tasks := resp.GetTasks()
	fmt.Printf("Found %d task(s):\n\n", len(tasks))

	for i, task := range tasks {
		fmt.Printf("Task #%d:\n", i+1)
		printTask(task)
		if i < len(tasks)-1 {
			fmt.Println()
		}
	}
}

func printTask(task *schedulerv1.Task) {
	fmt.Printf("  ID: %s\n", task.GetId())
	fmt.Printf("  Payload: %s\n", task.GetPayload())
	fmt.Printf("  Status: %s\n", task.GetStatus().String())
	fmt.Printf("  Request ID: %s\n", task.GetRequestId())
	fmt.Printf("  Attempt: %d\n", task.GetAttempt())
	if task.GetLastError() != "" {
		fmt.Printf("  Last Error: %s\n", task.GetLastError())
	}
	fmt.Printf("  Created: %s\n", task.GetCreatedAt().AsTime().Format(time.RFC3339))
	fmt.Printf("  Updated: %s\n", task.GetUpdatedAt().AsTime().Format(time.RFC3339))
	fmt.Printf("  Scheduled: %s\n", task.GetScheduledAt().AsTime().Format(time.RFC3339))
}

func printUsage() {
	fmt.Println("Task Scheduler CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/cli -cmd create <payload> [request_id] [-worker-group <group>]")
	fmt.Println("  go run ./cmd/cli -cmd get <task_id>")
	fmt.Println("  go run ./cmd/cli -cmd list [limit] [offset] [status]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/cli -cmd create 'test task'")
	fmt.Println("  go run ./cmd/cli -cmd create 'my task' 'req-123'")
	fmt.Println("  go run ./cmd/cli -cmd create 'image resize' 'req-123' -worker-group image")
	fmt.Println("  go run ./cmd/cli -cmd get abc-123-def-456")
	fmt.Println("  go run ./cmd/cli -cmd list 10 0")
	fmt.Println("  go run ./cmd/cli -cmd list 20 0 2  # status 2 = QUEUED")
	fmt.Println()
	fmt.Println("Status values:")
	fmt.Println("  0 = UNSPECIFIED")
	fmt.Println("  1 = DRAFT")
	fmt.Println("  2 = QUEUED")
	fmt.Println("  3 = RUNNING")
	fmt.Println("  4 = SUCCESS")
	fmt.Println("  5 = FAILED")
	fmt.Println("  6 = CANCELLED")
}
