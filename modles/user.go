// example schema.. 

package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100;not null"`
	Email     string    `gorm:"uniqueIndex;size:150"`
	CreatedAt time.Time
	UpdatedAt time.Time
}