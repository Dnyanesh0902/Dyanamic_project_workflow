package model

type WorkflowStep struct {
	ID           int32  `gorm:"column:id;primaryKey;autoIncrement;type:int" json:"id"`
	StepName     string `gorm:"column:step_name;type:varchar(100);not null" json:"step_name"`
	RoleName     string `gorm:"column:role_name;type:varchar(50);not null" json:"role_name"`
	StepSequence int32  `gorm:"column:step_sequence;type:int;not null;unique" json:"step_sequence"`
}

func (WorkflowStep) TableName() string {
	return "workflow_steps"
}
