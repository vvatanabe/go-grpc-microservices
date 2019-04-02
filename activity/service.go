package main

import (
	"context"
	"database/sql"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	"github.com/vvatanabe/go-grpc-microservices/shared/md"
)

type ActivityService struct {
	db    *sql.DB
	store Store
}

func (s *ActivityService) CreateActivity(ctx context.Context,
	req *pbActivity.CreateActivityRequest) (*empty.Empty, error) {
	userID := md.GetUserIDFromContext(ctx)
	if _, err := s.store.CreateActivity(&pbActivity.Activity{
		Content:   req.Content,
		UserId:    userID,
		CreatedAt: ptypes.TimestampNow(),
	}); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *ActivityService) FindActivities(ctx context.Context,
	_ *empty.Empty) (*pbActivity.FindActivitiesResponse, error) {
	userID := md.GetUserIDFromContext(ctx)
	activities, err := s.store.FindActivities(userID)
	if err != nil {
		return nil, err
	}
	return &pbActivity.FindActivitiesResponse{Activities: activities}, nil
}
