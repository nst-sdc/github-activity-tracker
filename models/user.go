package models

import "gorm.io/gorm"

type User struct {
    ID         uint           `gorm:"primaryKey"`
    Name       string
    Email      string         `gorm:"unique;not null"`
    GithubUser string         `gorm:"unique;not null"`
    PRs        []PR           `gorm:"foreignKey:UserID"`
    Orgs       []Org          `gorm:"many2many:user_orgs"`
    CreatedAt  gorm.DeletedAt `gorm:"autoCreateTime"`
}


