package model

import "time"

type ProjectApproval struct {
	ID       int          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ProjectID int          `gorm:"column:project_id;not null" json:"project_id"`
	StepID   int          `gorm:"column:step_id;not null" json:"step_id"`
	Status   string       `gorm:"column:status;default:'Pending';not null" json:"status"` // 'Pending', 'Approved', 'Rejected'
	ActionBy *int         `gorm:"column:action_by" json:"action_by"`
	ActionAt *time.Time   `gorm:"column:action_at" json:"action_at"`
	Remarks  string       `gorm:"column:remarks" json:"remarks"`
	Step     WorkflowStep `gorm:"foreignKey:StepID" json:"step"`
	User     *User        `gorm:"foreignKey:ActionBy" json:"user"`
}

func (ProjectApproval) TableName() string {
	return "project_approvals"
}
