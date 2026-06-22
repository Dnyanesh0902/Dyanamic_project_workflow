package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        int32          `gorm:"column:id;primaryKey;autoIncrement;type:int" json:"id"`
	UUID      string         `gorm:"column:uuid;type:varchar(36);not null;unique" json:"uuid"`
	Name      string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Email     string         `gorm:"column:email;type:varchar(255);not null;unique" json:"email"`
	Password  string         `gorm:"column:password;type:varchar(255);not null" json:"password"`
	Rolename  string         `gorm:"column:role_name;type:varchar(50);not null" json:"role_name"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	UpdatedBy int32          `gorm:"column:updated_by;type:int" json:"updated_by"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

func (User) TableName() string {
	return "users"
}
