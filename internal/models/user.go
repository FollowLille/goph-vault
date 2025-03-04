package models

import "gorm.io/gorm"

// User описывает пользователя
type User struct {
	gorm.Model
	Username        string `gorm:"unique;not null"`
	Password        string `gorm:"not null"`
	LastSyncVersion int
}
