package model

import (
	"time"
)

// UserModel GORMモデル（データベース用）
type UserModel struct {
	ID           string    `gorm:"primaryKey;type:varchar(36)"`
	Username     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	PasswordHash string    `gorm:"type:varchar(255);not null"`
	Role         string    `gorm:"type:enum('admin','editor','viewer');not null;default:'viewer'"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TableName テーブル名を指定
func (UserModel) TableName() string {
	return "users"
}
