package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Username   string         `gorm:"unique;not null" json:"username"`
	Password   string         `gorm:"not null" json:"-"`
	Role       string         `gorm:"not null" json:"role"`
	NeedsReset bool           `gorm:"not null" json:"needReset"`
	CreatedAt  time.Time      `json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}
