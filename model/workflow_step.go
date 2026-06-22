package model

type WorkflowStep struct {
	ID           int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	StepName     string `gorm:"column:step_name;not null" json:"step_name"`
	RoleName     string `gorm:"column:role_name;not null" json:"role_name"`
	StepSequence int    `gorm:"column:step_sequence;not null;unique" json:"step_sequence"`
}

func (WorkflowStep) TableName() string {
	return "workflow_steps"
}
