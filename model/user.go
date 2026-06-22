package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int            `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID      string         `gorm:"column:uuid" json:"uuid"`
	Name      string         `gorm:"column:name" json:"name"`
	Email     string         `gorm:"column:email" json:"email"`
	Password  string         `gorm:"column:password" json:"password"`
	Rolename  string         `gorm:"column:role_name" json:"role_name"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedBy int            `gorm:"column:updated_by" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (User) TableName() string {
	return "users"
}
