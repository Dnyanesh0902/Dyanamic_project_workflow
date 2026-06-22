package model

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          int32             `gorm:"column:id;primaryKey;autoIncrement;type:int" json:"id"`
	UUID        string            `gorm:"column:uuid;type:varchar(36);not null;unique" json:"uuid"`
	ProjectName string            `gorm:"column:project_name;type:varchar(255);not null" json:"project_name"`
	Description string            `gorm:"column:description;type:text" json:"description"`
	Budget      float64           `gorm:"column:budget;type:decimal(15,2);not null" json:"budget"`
	Status      string            `gorm:"column:status;type:varchar(20);default:'Pending';not null" json:"status"` // 'Pending', 'Approved', 'Rejected'
	CreatedBy   int32             `gorm:"column:created_by;type:int;not null" json:"created_by"`
	UpdatedBy   *int32            `gorm:"column:updated_by;type:int" json:"updated_by"`
	CreatedAt   time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"column:deleted_at;index" json:"deleted_at"`
	Approvals   []ProjectApproval `gorm:"foreignKey:ProjectID" json:"approvals"`
}

func (Project) TableName() string {
	return "projects"
}
