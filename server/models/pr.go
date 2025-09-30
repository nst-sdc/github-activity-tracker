package models

import (
	"time"

	"gorm.io/gorm"
)

type PR struct {
	ID        uint           `gorm:"primaryKey"`
	Title     string         `gorm:"not null"`
	Status    string         `gorm:"not null;default:'open'"`
	URL       string         `gorm:"not null"`
	Merged    bool           `gorm:"default:false"`
	OrgID     uint           `gorm:""`
	Org       Org            `gorm:"foreignKey:OrgID"`
	UserID    uint           `gorm:"not null"`
	User      User           `gorm:"foreignKey:UserID"`
	ProjectID uint           `gorm:""`
	Project   Project        `gorm:"foreignKey:ProjectID"`
	MonthID   uint           `gorm:"not null"`
	Month     Month          `gorm:"foreignKey:MonthID"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
