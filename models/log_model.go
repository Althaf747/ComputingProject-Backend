package models

import (
	"time"
)

type Log struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Authorized bool      `gorm:"not null" json:"authorized"`
	Confidence float64   `gorm:"not null" json:"confidence"`
	Name       string    `gorm:"not null" json:"name"`
	Role       string    `gorm:"not null" json:"role"`
	Timestamp  string    `gorm:"not null" json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Log) TableName() string {
	return "logs"
}
