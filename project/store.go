package main

import (
	"fmt"
	"sort"

	pbProject "github.com/vvatanabe/go-grpc-microservices/proto/project"
	"github.com/vvatanabe/go-grpc-microservices/shared/inmemory"
)

type Store interface {
	CreateProject(project *pbProject.Project) (*pbProject.Project, error)
	FindProject(projectID, userID uint64) (*pbProject.Project, error)
	FindProjects(userID uint64) ([]*pbProject.Project, error)
	UpdateProject(project *pbProject.Project) (*pbProject.Project, error)
}

func NewStoreOnMemory() *StoreOnMemory {
	return &StoreOnMemory{inmemory.NewIndexMap()}
}

type StoreOnMemory struct {
	projects *inmemory.IndexMap
}

func (s *StoreOnMemory) CreateProject(project *pbProject.Project) (*pbProject.Project, error) {
	newProject := *project
	idx := s.projects.Index()
	newProject.Id = idx
	s.projects.Set(idx, &newProject)
	return &newProject, nil
}

func (s *StoreOnMemory) FindProject(projectID, userID uint64) (*pbProject.Project, error) {
	value, ok := s.projects.Get(projectID)
	project := value.(*pbProject.Project)
	if !ok {
		return nil, fmt.Errorf("not found project")
	}
	if project.UserId != userID {
		return nil, fmt.Errorf("not found project")
	}
	return project, nil
}

func (s *StoreOnMemory) FindProjects(userID uint64) ([]*pbProject.Project, error) {
	var projects []*pbProject.Project
	s.projects.Range(func(idx uint64, value interface{}) bool {
		project := value.(*pbProject.Project)
		if project.UserId == userID {
			projects = append(projects, project)
		}
		return true
	})
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].CreatedAt.Seconds > projects[j].CreatedAt.Seconds
	})
	return projects, nil
}

func (s *StoreOnMemory) UpdateProject(project *pbProject.Project) (*pbProject.Project, error) {
	newProject := *project
	s.projects.Set(newProject.Id, &newProject)
	return &newProject, nil
}
