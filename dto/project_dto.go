package dto

import "time"

type ProjectCreateRequest struct {
	ProjectName string  `json:"project_name" binding:"required"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget" binding:"required,min=0"`
}

type ProjectApprovalRequest struct {
	ProjectUUID string `json:"project_uuid" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=Approved Rejected"`
	Remarks     string `json:"remarks"`
}

type ProjectApprovalUpdateRequest struct {
	ProjectUUID string `json:"project_uuid" binding:"required"`
	RoleName    string `json:"role_name" binding:"omitempty,oneof=RM BH VH"`
	Status      string `json:"status" binding:"required,oneof=Pending Approved Rejected"`
	Remarks     string `json:"remarks"`
}

type ProjectByUUIDRequest struct {
	UUID string `json:"uuid" binding:"required"`
}

type ProjectUpdateRequest struct {
	UUID        string  `json:"uuid" binding:"required"`
	ProjectName string  `json:"project_name"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget" binding:"omitempty,min=0"`
}

type ProjectApprovalResponse struct {
	ID       int        `json:"id"`
	StepName string     `json:"step_name"`
	RoleName string     `json:"role_name"`
	Sequence int        `json:"sequence"`
	Status   string     `json:"status"`
	ActionBy string     `json:"action_by,omitempty"`
	ActionAt *time.Time `json:"action_at,omitempty"`
	Remarks  string     `json:"remarks,omitempty"`
}

type ProjectResponse struct {
	ID          int                       `json:"id"`
	UUID        string                    `json:"uuid"`
	ProjectName string                    `json:"project_name"`
	Description string                    `json:"description"`
	Budget      float64                   `json:"budget"`
	Status      string                    `json:"status"`
	CreatedBy   int                       `json:"created_by"`
	UpdatedBy   *int                      `json:"updated_by,omitempty"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	DeletedAt   *time.Time                `json:"deleted_at,omitempty"`
	Approvals   []ProjectApprovalResponse `json:"approvals"`
}

type FlatRow struct {
	ProjectID          int        `gorm:"column:project_id"`
	ProjectUUID        string     `gorm:"column:project_uuid"`
	ProjectName        string     `gorm:"column:project_name"`
	ProjectDescription string     `gorm:"column:project_description"`
	ProjectBudget      float64    `gorm:"column:project_budget"`
	ProjectStatus      string     `gorm:"column:project_status"`
	ProjectCreatedBy   int        `gorm:"column:project_created_by"`
	ProjectUpdatedBy   *int       `gorm:"column:project_updated_by"`
	ProjectCreatedAt   time.Time  `gorm:"column:project_created_at"`
	ProjectUpdatedAt   time.Time  `gorm:"column:project_updated_at"`
	ProjectDeletedAt   *time.Time `gorm:"column:project_deleted_at"`

	ApprovalID       *int       `gorm:"column:approval_id"`
	ApprovalStatus   *string    `gorm:"column:approval_status"`
	ApprovalActionBy *int       `gorm:"column:approval_action_by"`
	ApprovalActionAt *time.Time `gorm:"column:approval_action_at"`
	ApprovalRemarks  *string    `gorm:"column:approval_remarks"`

	StepID           *int    `gorm:"column:step_id"`
	StepName         *string `gorm:"column:step_name"`
	StepRoleName     *string `gorm:"column:step_role_name"`
	StepStepSequence *int    `gorm:"column:step_sequence"`

	UserID   *int    `gorm:"column:user_id"`
	UserName *string `gorm:"column:user_name"`
}

type ProjectRepoCreateRequest struct {
	ProjectName string  `json:"project_name"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	UUID        string  `json:"uuid"`
	UserID      int     `json:"user_id"`
}

type ProjectListRequest struct {
	UserID   int    `json:"user_id"`
	UserRole string `json:"user_role"`
}

type ProjectApprovalCreateRequest struct {
	ProjectID int    `json:"project_id"`
	StepID    int    `json:"step_id"`
	Status    string `json:"status"`
}

type ProjectApprovalDirectUpdateRequest struct {
	ApprovalID int       `json:"approval_id"`
	Status     string    `json:"status"`
	ActionBy   int       `json:"action_by"`
	ActionAt   time.Time `json:"action_at"`
	Remarks    string    `json:"remarks"`
}

type ProjectStatusDirectUpdateRequest struct {
	ProjectID int    `json:"project_id"`
	Status    string `json:"status"`
	UpdatedBy int    `json:"updated_by"`
}

type ProjectDetailsRepoUpdateRequest struct {
	UUID        string  `json:"uuid"`
	ProjectName string  `json:"project_name"`
	Description string  `json:"description"`
	Budget      float64 `json:"budget"`
	UserID      int     `json:"user_id"`
}

type ProjectServiceCreateRequest struct {
	UserID int `json:"user_id"`
	ProjectCreateRequest
}

type ProjectServiceApproveRequest struct {
	UserID   int    `json:"user_id"`
	UserRole string `json:"user_role"`
	ProjectApprovalRequest
}

type ProjectServiceListRequest struct {
	UserID   int    `json:"user_id"`
	UserRole string `json:"user_role"`
}

type ProjectServiceUpdateRequest struct {
	UserID int `json:"user_id"`
	ProjectUpdateRequest
}

type ProjectServiceApprovalUpdateRequest struct {
	UserID   int    `json:"user_id"`
	UserRole string `json:"user_role"`
	ProjectApprovalUpdateRequest
}
