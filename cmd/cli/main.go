package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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
		atTime   = flag.String("at", "", "run at specific time: \"19:00\", \"19:00:00\" or RFC3339 (e.g. 2025-02-01T19:00:00Z)")
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
		createTask(ctx, client, *workerID, *atTime, flag.Args())
	case "get":
		getTask(ctx, client, flag.Args())
	case "list":
		listTasks(ctx, client, flag.Args())
	default:
		printUsage()
		os.Exit(1)
	}
}

func createTask(ctx context.Context, client schedulerv1.SchedulerServiceClient, workerIDFlag string, atTime string, args []string) {

	positionals := collectCreatePositionals(args)

	var payload, requestID string
	switch {
	case len(positionals) == 0:
		log.Fatal("Error: payload is required. Usage: cli -cmd create <payload> [request_id] [-worker-id <worker_id>] [-at <time>]")
	case positionals[0] == "create" && len(positionals) < 2:
		log.Fatal("Error: payload is required. Usage: cli -cmd create <payload> [request_id] [-worker-id <worker_id>] [-at <time>]")
	case positionals[0] == "create":
		payload = positionals[1]
		if len(positionals) >= 3 {
			requestID = positionals[2]
		}
	default:

		payload = positionals[0]
		if len(positionals) >= 2 {
			requestID = positionals[1]
		}
	}
	if requestID == "" {
		requestID = fmt.Sprintf("req-%d", time.Now().Unix())
	}

	workerID := workerIDFlag
	for i := 0; i < len(args); i++ {
		if (args[i] == "-worker-id" || args[i] == "--worker-id") && i+1 < len(args) && workerID == "" {
			workerID = args[i+1]
			break
		}
	}

	if atTime == "" {
		for i := 0; i < len(args); i++ {
			if args[i] == "-at" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				atTime = args[i+1]
				break
			}
		}
	}

	scheduledAt := time.Now().UTC()
	if atTime != "" {
		t, err := parseScheduledTime(atTime)
		if err != nil {
			log.Fatalf("Invalid -at time %q: %v (use 19:00, 19:00:00 or RFC3339)", atTime, err)
		}
		scheduledAt = t
		fmt.Printf("Task will be sent to scheduler at: %s (status TIME)\n", scheduledAt.Format(time.RFC3339))
	}

	req := &schedulerv1.CreateTaskRequest{
		Payload:     payload,
		RequestId:   requestID,
		WorkerId:    workerID,
		ScheduledAt: timestamppb.New(scheduledAt),
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
	if task.GetWorkerId() != "" {
		fmt.Printf("  Worker ID: %s\n", task.GetWorkerId())
	}
	fmt.Printf("  Created: %s\n", task.GetCreatedAt().AsTime().Format(time.RFC3339))
	fmt.Printf("  Scheduled: %s\n", task.GetScheduledAt().AsTime().Format(time.RFC3339))
}

func collectCreatePositionals(args []string) []string {
	var out []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-at", "-worker-id", "--worker-id":
			i++
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			continue
		}
		out = append(out, args[i])
	}
	return out
}

func parseScheduledTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Now().UTC(), nil
	}

	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	now := time.Now()
	loc := now.Location()
	var target time.Time
	// "00:00:00"
	if t, err := time.ParseInLocation("15:04:05", s, loc); err == nil {
		target = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc)
	} else if t, err := time.ParseInLocation("15:04", s, loc); err == nil {
		// "00:00"
		target = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc)
	} else if t, err := time.ParseInLocation("15", s, loc); err == nil {
		// "00"
		target = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), 0, 0, 0, loc)
	} else {
		return time.Time{}, fmt.Errorf("unsupported format")
	}
	if !target.After(now) {
		target = target.AddDate(0, 0, 1)
	}
	return target.UTC(), nil
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
	fmt.Println("  go run ./cmd/cli -cmd create <payload> [request_id] [-worker-id <worker_id>] [-at <time>]")
	fmt.Println("  go run ./cmd/cli -cmd get <task_id>")
	fmt.Println("  go run ./cmd/cli -cmd list [limit] [offset] [status]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/cli -cmd create 'test task'")
	fmt.Println("  go run ./cmd/cli -cmd create 'run at 19:00' -at 19:00")
	fmt.Println("  go run ./cmd/cli -cmd create 'my task' 'req-123'")
	fmt.Println("  go run ./cmd/cli -cmd create 'resize avatar' 'req-123' -worker-id worker1")
	fmt.Println("  go run ./cmd/cli -cmd get abc-123-def-456")
	fmt.Println("  go run ./cmd/cli -cmd list 10 0")
	fmt.Println("  go run ./cmd/cli -cmd list 20 0 2  # status 2 = QUEUED")
	fmt.Println()
	fmt.Println("Status values:")
	fmt.Println("  0 = UNSPECIFIED")
	fmt.Println("  1 = DRAFT   (готово к отправке сразу)")
	fmt.Println("  2 = QUEUED")
	fmt.Println("  3 = RUNNING")
	fmt.Println("  4 = SUCCESS")
	fmt.Println("  5 = FAILED")
	fmt.Println("  6 = CANCELLED")
	fmt.Println("  7 = TIME    (запланировано на время, отправится когда наступит scheduled_at)")
}
