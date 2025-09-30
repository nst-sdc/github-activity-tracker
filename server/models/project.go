package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	OrgID     uint           `gorm:"not null"`
	Org       Org            `gorm:"foreignKey:OrgID"`
	PRs       []PR           `gorm:"foreignKey:ProjectID"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
