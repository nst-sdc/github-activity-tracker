package models

import "gorm.io/gorm"

type Project struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"not null"`
    OrgID     uint           `gorm:"not null"`
    Org       Org            `gorm:"foreignKey:OrgID"`
    PRs       []PR           `gorm:"foreignKey:ProjectID"`
    CreatedAt gorm.DeletedAt `gorm:"autoCreateTime"`
}