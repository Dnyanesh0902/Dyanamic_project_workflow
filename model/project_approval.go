package model

import "time"

type ProjectApproval struct {
	ID       int32        `gorm:"column:id;primaryKey;autoIncrement;type:int" json:"id"`
	ProjectID int32        `gorm:"column:project_id;type:int;not null" json:"project_id"`
	StepID   int32        `gorm:"column:step_id;type:int;not null" json:"step_id"`
	Status   string       `gorm:"column:status;type:varchar(20);default:'Pending';not null" json:"status"` // 'Pending', 'Approved', 'Rejected'
	ActionBy *int32       `gorm:"column:action_by;type:int" json:"action_by"`
	ActionAt *time.Time   `gorm:"column:action_at" json:"action_at"`
	Remarks  string       `gorm:"column:remarks;type:text" json:"remarks"`
	Step     WorkflowStep `gorm:"foreignKey:StepID" json:"step"`
	User     *User        `gorm:"foreignKey:ActionBy" json:"user"`
}

func (ProjectApproval) TableName() string {
	return "project_approvals"
}
