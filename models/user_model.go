package models

import "time"

type User struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"unique;not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"`
	OldPassword string    `gorm:"not null;default:''" json:"-"`
	Role        string    `gorm:"not null" json:"role"`
	NeedsReset  bool      `gorm:"not null" json:"needReset"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FCMToken    string    `gorm:"size:500" json:"-"`
}

func (User) TableName() string {
	return "users"
}
