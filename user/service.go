package main

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	"golang.org/x/crypto/bcrypt"

	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	pbUser "github.com/vvatanabe/go-grpc-microservices/proto/user"
	"github.com/vvatanabe/go-grpc-microservices/shared/md"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	store         Store
	projectClient pbProject.ProjectServiceClient
}

const defaultProjectName = "default"

func (s *UserService) CreateUser(ctx context.Context,
	req *pbUser.CreateUserRequest) (*pbUser.CreateUserResponse, error) {
	if req.Email == "" || len(req.Password) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "empty email or password")
	}
	passwordHash, err := bcrypt.GenerateFromPassword(req.Password, bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	user, err := s.store.CreateUser(&pbUser.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    ptypes.TimestampNow(),
	})
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	ctx = md.AddUserIDToContext(ctx, user.Id)
	if _, err := s.projectClient.CreateProject(ctx, &pbProject.CreateProjectRequest{
		Name: defaultProjectName,
	}); err != nil {
		return nil, err
	}
	return &pbUser.CreateUserResponse{User: user}, nil
}

func (s *UserService) FindUser(ctx context.Context,
	req *pbUser.FindUserRequest) (*pbUser.FindUserResponse, error) {
	user, err := s.store.FindUser(req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pbUser.FindUserResponse{User: user}, nil
}

func (s *UserService) VerifyUser(ctx context.Context,
	req *pbUser.VerifyUserRequest) (*pbUser.VerifyUserResponse, error) {
	user, err := s.store.FindUserByEmail(req.Email)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err := bcrypt.CompareHashAndPassword(user.PasswordHash, req.Password); err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}
	return &pbUser.VerifyUserResponse{User: user}, nil
}
