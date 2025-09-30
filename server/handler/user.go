package handler

import (
	"errors"

	"github-activity-tracker/models"
	"github-activity-tracker/service"

	"gofr.dev/pkg/gofr"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) UserHandler {
	return UserHandler{userService: userService}
}

func (h UserHandler) AddGitId(ctx *gofr.Context) (interface{}, error) {
	var u models.User

	err := ctx.Bind(&u)

	if err != nil {
		return nil, errors.New("Invalid request format")
	}

	// Validate that GitHub ID is provided
	if u.GithubUser == "" {
		return nil, errors.New("GitHub ID is required")
	}

	user, err := h.userService.AddGitId(u)

	if err != nil {
		return nil, err
	}

	return user, nil
}
