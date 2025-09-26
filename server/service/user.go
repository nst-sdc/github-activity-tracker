package service

import (
	"github.com/nst-sdc/github-activity-tracker/model"
	"github.com/nst-sdc/github-activity-tracker/store"
)

type UserService struct {
	Store store.UserStore
}

func NewUserService(store store.UserStore) *UserService {
	return &UserService{Store: store}
}

func (u *UserService) AddGitId(gitId model.GitId) (model.GitId, error) {
	return u.Store.Save(gitId)
}

func (u *UserService) GetUserByID(id string) (model.GitId, error) {
	return u.Store.GetByID(id)
}

func (u *UserService) GetUserByGitHubID(githubID string) (model.GitId, error) {
	return u.Store.GetByGitHubID(githubID)
}

func (u *UserService) GetAllUsers() ([]model.GitId, error) {
	return u.Store.GetAll()
}

func (u *UserService) DeleteUser(id string) error {
	return u.Store.Delete(id)
}
