package models

import "gorm.io/gorm"

type Org struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"unique;not null"`
    URL       string         `gorm:"not null"`
    Projects  []Project      `gorm:"foreignKey:OrgID"`
    PRs       []PR           `gorm:"foreignKey:OrgID"`
    Users     []User         `gorm:"many2many:user_orgs"`
    CreatedAt gorm.DeletedAt `gorm:"autoCreateTime"`
}