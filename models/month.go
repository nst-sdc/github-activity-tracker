package models

import "gorm.io/gorm"

type Month struct {
    ID             uint           `gorm:"primaryKey"`
    Name           string         `gorm:"not null"`
    TotalPR        int            `gorm:"default:0"`
    TopContributor string         // Github username or UserID
    HighestPROnOrg string         // Org name or OrgID
    PRs            []PR           `gorm:"foreignKey:MonthID"`
    CreatedAt      gorm.DeletedAt `gorm:"autoCreateTime"`
}