package main

import (
	"fmt"

	"sort"

	pbTask "github.com/vvatanabe/go-grpc-microservices/proto/task"
	"github.com/vvatanabe/go-grpc-microservices/shared/inmemory"
)

type Store interface {
	CreateTask(task *pbTask.Task) (*pbTask.Task, error)
	FindTask(taskID, userID uint64) (*pbTask.Task, error)
	FindTasks(userID uint64) ([]*pbTask.Task, error)
	FindProjectTasks(projectID, userID uint64) ([]*pbTask.Task, error)
	UpdateTask(task *pbTask.Task) (*pbTask.Task, error)
}

func NewStoreOnMemory() *StoreOnMemory {
	return &StoreOnMemory{inmemory.NewIndexMap()}
}

type StoreOnMemory struct {
	tasks *inmemory.IndexMap
}

func (s *StoreOnMemory) CreateTask(task *pbTask.Task) (*pbTask.Task, error) {
	newTask := *task
	idx := s.tasks.Index()
	newTask.Id = idx
	s.tasks.Set(idx, &newTask)
	return &newTask, nil
}

func (s *StoreOnMemory) FindTask(taskID, userID uint64) (*pbTask.Task, error) {
	value, ok := s.tasks.Get(taskID)
	task := value.(*pbTask.Task)
	if !ok {
		return nil, fmt.Errorf("not found task")
	}
	if task.UserId != userID {
		return nil, fmt.Errorf("not found task")
	}
	return task, nil
}

func (s *StoreOnMemory) FindTasks(userID uint64) ([]*pbTask.Task, error) {
	var tasks []*pbTask.Task
	s.tasks.Range(func(idx uint64, value interface{}) bool {
		task := value.(*pbTask.Task)
		if task.UserId == userID {
			tasks = append(tasks, task)
		}
		return true
	})
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.Seconds > tasks[j].UpdatedAt.Seconds
	})
	return tasks, nil
}

func (s *StoreOnMemory) FindProjectTasks(projectID, userID uint64) ([]*pbTask.Task, error) {
	var tasks []*pbTask.Task
	s.tasks.Range(func(idx uint64, value interface{}) bool {
		task := value.(*pbTask.Task)
		if task.ProjectId == projectID && task.UserId == userID {
			tasks = append(tasks, task)
		}
		return true
	})
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.Seconds > tasks[j].UpdatedAt.Seconds
	})
	return tasks, nil
}

func (s *StoreOnMemory) UpdateTask(task *pbTask.Task) (*pbTask.Task, error) {
	newTask := *task
	s.tasks.Set(newTask.Id, &newTask)
	return &newTask, nil
}
