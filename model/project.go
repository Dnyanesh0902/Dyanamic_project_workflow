package model

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	ID          int               `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UUID        string            `gorm:"column:uuid;not null;unique" json:"uuid"`
	ProjectName string            `gorm:"column:project_name;not null" json:"project_name"`
	Description string            `gorm:"column:description" json:"description"`
	Budget      float64           `gorm:"column:budget;type:decimal(15,2);not null" json:"budget"`
	Status      string            `gorm:"column:status;default:'Pending';not null" json:"status"` // 'Pending', 'Approved', 'Rejected'
	CreatedBy   int               `gorm:"column:created_by;not null" json:"created_by"`
	UpdatedBy   *int              `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt   time.Time         `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time         `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt    `gorm:"column:deleted_at;index" json:"deleted_at"`
	Approvals   []ProjectApproval `gorm:"foreignKey:ProjectID" json:"approvals"`
}

func (Project) TableName() string {
	return "projects"
}
