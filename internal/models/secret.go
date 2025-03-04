package models

import (
	"gorm.io/gorm"
	"time"
)

type SecretType string

const (
	TypeLogin  SecretType = "login"
	TypeText   SecretType = "text"
	TypeBinary SecretType = "binary"
	TypeCard   SecretType = "card"
)

// Secret описывает секрет
type Secret struct {
	gorm.Model
	UserID         uint       `gorm:"not null"` // ID пользователя
	Name           string     `gorm:"not null"` // Имя секрета
	Type           SecretType `gorm:"not null"` // Тип секрета (логин, текст, бинарные данные, карта)
	Metadata       string     // Метаинформация (например, теги, категории)
	Data           []byte     `gorm:"not null"`      // Данные
	Version        int        `gorm:"not null"`      // Версия данных для синхронизации
	LocalUpdatedAt time.Time  `gorm:"-"`             // Время последнего обновления локально, в gorm игнорируем
	Deleted        bool       `gorm:"default:false"` // Флаг актуальность секрета
	Synced         bool       `json:"synced"`        // Флаг синхронизации секрета с сервером
}
