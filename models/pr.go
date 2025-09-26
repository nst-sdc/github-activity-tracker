package models

import "gorm.io/gorm"

type PR struct {
    ID         uint           `gorm:"primaryKey"`
    Status     string         `gorm:"not null;default:'open'"`
    OrgID      uint           `gorm:"not null"`
    Org        Org            `gorm:"foreignKey:OrgID"`
    UserID     uint           `gorm:"not null"`
    User       User           `gorm:"foreignKey:UserID"`
    ProjectID  uint           `gorm:"not null"`
    Project    Project        `gorm:"foreignKey:ProjectID"`
    MonthID    uint           `gorm:"not null"`
    Month      Month          `gorm:"foreignKey:MonthID"`
    CreatedAt  gorm.DeletedAt `gorm:"autoCreateTime"`
}