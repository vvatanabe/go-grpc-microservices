package main

import (
	"errors"
	"sort"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	pbActivity "github.com/vvatanabe/go-grpc-microservices/proto/activity"
	"github.com/vvatanabe/go-grpc-microservices/shared/inmemory"
)

type Store interface {
	CreateActivity(activity *pbActivity.Activity) (*pbActivity.Activity, error)
	FindActivities(userID uint64) ([]*pbActivity.Activity, error)
}

func NewStoreOnMemory() *StoreOnMemory {
	return &StoreOnMemory{inmemory.NewIndexMap()}
}

type StoreOnMemory struct {
	activities *inmemory.IndexMap
}

func (s *StoreOnMemory) CreateActivity(activity *pbActivity.Activity) (*pbActivity.Activity, error) {
	if kindOf(activity.Content) == KindUnknown {
		return nil, errors.New("unknown activity content")
	}
	newActivity := *activity
	idx := s.activities.Index()
	newActivity.Id = idx
	s.activities.Set(idx, &newActivity)
	return &newActivity, nil
}

func (s *StoreOnMemory) FindActivities(userID uint64) ([]*pbActivity.Activity, error) {
	var activities []*pbActivity.Activity
	s.activities.Range(func(idx uint64, value interface{}) bool {
		activity := value.(*pbActivity.Activity)
		if activity.UserId == userID {
			activities = append(activities, activity)
		}
		return true
	})
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].CreatedAt.Seconds > activities[j].CreatedAt.Seconds
	})
	return activities, nil
}

type Kind int32

const (
	KindUnknown          Kind = 0
	KindCreateTask       Kind = 1
	KindUpdateTaskStatus Kind = 2
	KindCreateProject    Kind = 3
)

func kindOf(any *any.Any) Kind {
	if msg := new(pbActivity.CreateTaskContent); ptypes.Is(any, msg) {
		return KindCreateTask
	}
	if msg := new(pbActivity.UpdateTaskStatusContent); ptypes.Is(any, msg) {
		return KindUpdateTaskStatus
	}
	if msg := new(pbActivity.CreateProjectContent); ptypes.Is(any, msg) {
		return KindCreateProject
	}
	return KindUnknown
}
