package main

import (
	"fmt"

	pbUser "github.com/vvatanabe/go-grpc-microservices/proto/user"
	"github.com/vvatanabe/go-grpc-microservices/shared/inmemory"
)

type Store interface {
	CreateUser(user *pbUser.User) (*pbUser.User, error)
	FindUser(userID uint64) (*pbUser.User, error)
	FindUserByEmail(email string) (*pbUser.User, error)
}

func NewStoreOnMemory() *StoreOnMemory {
	return &StoreOnMemory{inmemory.NewIndexMap()}
}

type StoreOnMemory struct {
	users *inmemory.IndexMap
}

func (s *StoreOnMemory) CreateUser(user *pbUser.User) (*pbUser.User, error) {
	if _, err := s.FindUserByEmail(user.Email); err == nil {
		return nil, fmt.Errorf("already exists user %s", user.Email)
	}
	newUser := *user
	idx := s.users.Index()
	newUser.Id = idx
	s.users.Set(idx, &newUser)
	return &newUser, nil
}

func (s *StoreOnMemory) FindUser(userID uint64) (*pbUser.User, error) {
	value, ok := s.users.Get(userID)
	if !ok {
		return nil, fmt.Errorf("not found user %d", userID)
	}
	return value.(*pbUser.User), nil
}

func (s *StoreOnMemory) FindUserByEmail(email string) (*pbUser.User, error) {
	var user *pbUser.User
	s.users.Range(func(idx uint64, value interface{}) bool {
		u := value.(*pbUser.User)
		if u.Email == email {
			user = u
			return false
		}
		return true
	})
	if user == nil {
		return nil, fmt.Errorf("not found user")
	}
	return user, nil
}
