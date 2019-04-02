package main

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	pbTask "github.com/vvatanabe/go-grpc-microservices/proto/task"
	"github.com/vvatanabe/go-grpc-microservices/shared/md"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TaskService struct {
	store          Store
	activityClient pbActivity.ActivityServiceClient
	projectClient  pbProject.ProjectServiceClient
}

func (s *TaskService) CreateTask(
	ctx context.Context,
	req *pbTask.CreateTaskRequest,
) (*pbTask.CreateTaskResponse, error) {
	if req.GetName() == "" {
		// gRPCのステータスコード付きのerrorを生成する
		return nil, status.Error(codes.InvalidArgument,
			"empty task name")
	}
	// ProjectServiceのクライアントスタブでプロジェクトを取得する
	resp, err := s.projectClient.FindProject(ctx,
		&pbProject.FindProjectRequest{
			ProjectId: req.GetProjectId(),
		})
	if err != nil {
		return nil, status.Error(
			codes.NotFound, err.Error())
	}
	// メタデータからUserIDを取得する
	userID := md.GetUserIDFromContext(ctx)
	// protobufのTimestamp型で現在日時を取得する
	now := ptypes.TimestampNow()
	// タスクを保存する
	task, err := s.store.CreateTask(&pbTask.Task{
		Name:      req.GetName(),
		Status:    pbTask.Status_WAITING,
		UserId:    userID,
		ProjectId: resp.Project.GetId(),
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}
	// アクティビティの内容をAny型に変換する
	content := &pbActivity.CreateTaskContent{
		TaskId:   task.GetId(),
		TaskName: task.GetName()}
	any, err := ptypes.MarshalAny(content)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}
	// ActivityServiceのクライアントスタブでアクティビティを作成する
	if _, err := s.activityClient.CreateActivity(ctx,
		&pbActivity.CreateActivityRequest{
			Content: any,
		}); err != nil {
		return nil, err
	}
	return &pbTask.CreateTaskResponse{Task: task}, nil
}

func (s *TaskService) FindTasks(
	ctx context.Context,
	_ *empty.Empty,
) (*pbTask.FindTasksResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	tasks, err := s.store.FindTasks(userID)
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}
	return &pbTask.FindTasksResponse{Tasks: tasks}, nil
}

func (s *TaskService) FindProjectTasks(ctx context.Context,
	req *pbTask.FindProjectTasksRequest) (*pbTask.FindProjectTasksResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	tasks, err := s.store.FindProjectTasks(req.GetProjectId(), userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pbTask.FindProjectTasksResponse{Tasks: tasks}, nil
}

func (s *TaskService) UpdateTask(
	ctx context.Context,
	req *pbTask.UpdateTaskRequest,
) (*pbTask.UpdateTaskResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"empty task name")
	}
	if req.GetStatus() == pbTask.Status_UNKNOWN {
		return nil, status.Error(
			codes.InvalidArgument,
			"unknown task status")
	}
	userID := md.GetUserIDFromContext(ctx)
	task, err := s.store.FindTask(req.GetTaskId(), userID)
	if err != nil {
		return nil, status.Error(
			codes.NotFound, err.Error())
	}
	updatedTask, err := s.store.UpdateTask(&pbTask.Task{
		Id:        task.Id,
		Name:      req.GetName(),
		Status:    req.GetStatus(),
		ProjectId: task.GetProjectId(),
		UserId:    task.GetUserId(),
		CreatedAt: task.GetCreatedAt(),
		UpdatedAt: ptypes.TimestampNow(),
	})
	if err != nil {
		return nil, status.Error(
			codes.InvalidArgument, err.Error())
	}
	if task.GetStatus() == updatedTask.GetStatus() {
		return &pbTask.UpdateTaskResponse{Task: updatedTask}, nil
	}
	any, err := ptypes.MarshalAny(&pbActivity.UpdateTaskStatusContent{
		TaskId:     updatedTask.GetId(),
		TaskName:   updatedTask.GetName(),
		TaskStatus: updatedTask.GetStatus(),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err := s.activityClient.CreateActivity(ctx,
		&pbActivity.CreateActivityRequest{Content: any}); err != nil {
		return nil, err
	}
	return &pbTask.UpdateTaskResponse{Task: updatedTask}, nil
}
