// example controller
// to control db and store data

package controllers

import (
	"go-orm-app/config"
	"go-orm-app/models"
)

func CreateUser(name string, email string) models.User {
	user := models.User{Name: name, Email: email}
	config.DB.Create(&user)
	return user
}

func GetUsers() []models.User {
	var users []models.User
	config.DB.Find(&users)
	return users
}
