package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"task_shelduler/internal/models"
	"task_shelduler/internal/repository"
	schedulerv1 "task_shelduler/pkg/proto/v1"
)

type SchedulerService struct {
	schedulerv1.UnimplementedSchedulerServiceServer

	tasks    repository.TaskRepository
	attempts repository.AttemptRepository
}

func NewSchedulerService(
	tasks repository.TaskRepository,
	attempts repository.AttemptRepository,
) *SchedulerService {
	return &SchedulerService{
		tasks:    tasks,
		attempts: attempts,
	}
}

func (s *SchedulerService) CreateTask(ctx context.Context, req *schedulerv1.CreateTaskRequest) (*schedulerv1.CreateTaskResponse, error) {
	now := time.Now().UTC()

	task := &models.Task{
		ID:          generateID(),
		Payload:     req.GetPayload(),
		Status:      models.TaskStatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
		ScheduledAt: req.GetScheduledAt().AsTime(),
		Attempt:     0,
		LastError:   "",
		RequestID:   req.GetRequestId(),
		Priority:    req.GetPriority(),
	}

	// requirements позже

	if err := s.tasks.CreateTask(task); err != nil {
		return nil, err
	}

	return &schedulerv1.CreateTaskResponse{
		Task: mapTaskToProto(task),
	}, nil
}

func (s *SchedulerService) GetTask(ctx context.Context, req *schedulerv1.GetTaskRequest) (*schedulerv1.GetTaskResponse, error) {
	task, err := s.tasks.GetTask(req.GetId())
	if err != nil {
		return nil, err
	}

	return &schedulerv1.GetTaskResponse{
		Task: mapTaskToProto(task),
	}, nil
}

func (s *SchedulerService) ListTasks(ctx context.Context, req *schedulerv1.ListTasksRequest) (*schedulerv1.ListTasksResponse, error) {
	status := models.TaskStatus(req.GetStatus())

	tasks, err := s.tasks.ListTasks(status, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, err
	}

	resp := &schedulerv1.ListTasksResponse{
		Tasks: make([]*schedulerv1.Task, 0, len(tasks)),
		Total: int32(len(tasks)),
	}

	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, mapTaskToProto(t))
	}

	return resp, nil
}

func (s *SchedulerService) UpdateTaskStatus(ctx context.Context, req *schedulerv1.UpdateTaskStatusRequest) (*schedulerv1.UpdateTaskStatusResponse, error) {
	status := models.TaskStatus(req.GetStatus())

	if err := s.tasks.UpdateTaskStatus(req.GetTaskId(), status); err != nil {
		return nil, err
	}

	task, err := s.tasks.GetTask(req.GetTaskId())
	if err != nil {
		return nil, err
	}

	return &schedulerv1.UpdateTaskStatusResponse{
		Task: mapTaskToProto(task),
	}, nil
}

func (s *SchedulerService) ReportAttempt(ctx context.Context, req *schedulerv1.ReportAttemptRequest) (*schedulerv1.ReportAttemptResponse, error) {
	now := time.Now().UTC()

	attempt := &models.Attempt{
		ID:         generateID(),
		TaskID:     req.GetTaskId(),
		StartedAt:  now,
		FinishedAt: ptrTime(now),
		Status:     models.AttemptStatus(req.GetStatus()),
		Error:      req.GetError(),
	}

	if err := s.attempts.CreateAttempt(attempt); err != nil {
		return nil, err
	}

	switch req.GetStatus() {
	case schedulerv1.AttemptStatus_ATTEMPT_STATUS_SUCCESS:
		_ = s.tasks.UpdateTaskStatusWithError(req.GetTaskId(), models.TaskStatusSuccess, "")
	case schedulerv1.AttemptStatus_ATTEMPT_STATUS_FAILED:
		_ = s.tasks.UpdateTaskStatusWithError(req.GetTaskId(), models.TaskStatusFailed, req.GetError())
	}

	return &schedulerv1.ReportAttemptResponse{
		Attempt: mapAttemptToProto(attempt),
	}, nil
}

func mapTaskToProto(t *models.Task) *schedulerv1.Task {
	if t == nil {
		return nil
	}

	return &schedulerv1.Task{
		Id:          t.ID,
		Payload:     t.Payload,
		Status:      schedulerv1.TaskStatus(t.Status),
		CreatedAt:   toTimestamp(t.CreatedAt),
		UpdatedAt:   toTimestamp(t.UpdatedAt),
		ScheduledAt: toTimestamp(t.ScheduledAt),
		Attempt:     int32(t.Attempt),
		LastError:   t.LastError,
		RequestId:   t.RequestID,
		// Requirements и Priority добавить позже
	}
}

func mapAttemptToProto(a *models.Attempt) *schedulerv1.Attempt {
	if a == nil {
		return nil
	}

	return &schedulerv1.Attempt{
		Id:        a.ID,
		TaskId:    a.TaskID,
		StartedAt: toTimestamp(a.StartedAt),
		FinishedAt: func() *timestamppb.Timestamp {
			if a.FinishedAt == nil {
				return nil
			}
			return toTimestamp(*a.FinishedAt)
		}(),
		Status: schedulerv1.AttemptStatus(a.Status),
		Error:  a.Error,
	}
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func generateID() string {
	return uuid.NewString()
}
