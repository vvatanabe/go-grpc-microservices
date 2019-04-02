package main

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	"github.com/vvatanabe/go-grpc-microservices/shared/md"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProjectService struct {
	store          Store
	activityClient pbActivity.ActivityServiceClient
}

func (s *ProjectService) CreateProject(ctx context.Context,
	req *pbProject.CreateProjectRequest) (*pbProject.CreateProjectResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "empty project name")
	}
	userID := md.GetUserIDFromContext(ctx)
	project, err := s.store.CreateProject(&pbProject.Project{
		Name:      req.Name,
		UserId:    userID,
		CreatedAt: ptypes.TimestampNow(),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	any, err := ptypes.MarshalAny(&pbActivity.CreateProjectContent{
		ProjectId:   project.Id,
		ProjectName: project.Name,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if _, err := s.activityClient.CreateActivity(
		ctx,
		&pbActivity.CreateActivityRequest{Content: any}); err != nil {
		return nil, err
	}
	return &pbProject.CreateProjectResponse{Project: project}, nil
}

func (s *ProjectService) FindProject(ctx context.Context,
	req *pbProject.FindProjectRequest) (*pbProject.FindProjectResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	project, err := s.store.FindProject(req.ProjectId, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pbProject.FindProjectResponse{Project: project}, nil
}

func (s *ProjectService) FindProjects(ctx context.Context,
	_ *empty.Empty) (*pbProject.FindProjectsResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	projects, err := s.store.FindProjects(userID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &pbProject.FindProjectsResponse{Projects: projects}, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context,
	req *pbProject.UpdateProjectRequest) (*pbProject.UpdateProjectResponse, error) {
	if req.ProjectName == "" {
		return nil, status.Error(codes.InvalidArgument, "empty project name")
	}
	userID := md.GetUserIDFromContext(ctx)
	project, err := s.store.FindProject(req.ProjectId, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	project.Name = req.ProjectName
	if _, err := s.store.UpdateProject(project); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pbProject.UpdateProjectResponse{Project: project}, nil
}
