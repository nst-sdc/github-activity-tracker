package service

import (
	"github-activity-tracker/models"
	"github-activity-tracker/store"
)

type UserService struct {
	Store store.UserStore
}

func NewUserService(store store.UserStore) *UserService {
	return &UserService{Store: store}
}

func (u *UserService) AddGitId(user models.User) (models.User, error) {
	return u.Store.Save(user)
}

func (u *UserService) GetUserByID(id string) (models.User, error) {
	return u.Store.GetByID(id)
}

func (u *UserService) GetUserByGitHubID(githubID string) (models.User, error) {
	return u.Store.GetByGitHubID(githubID)
}

func (u *UserService) GetAllUsers() ([]models.User, error) {
	return u.Store.GetAll()
}

func (u *UserService) DeleteUser(id string) error {
	return u.Store.Delete(id)
}
