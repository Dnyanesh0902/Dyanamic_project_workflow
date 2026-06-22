package repository

import (
	"project-workflow-backend/dto"
	"project-workflow-backend/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	Create(req dto.ProjectRepoCreateRequest) (*model.Project, error)
	GetByUUID(req dto.ProjectByUUIDRequest) (*dto.ProjectResponse, error)
	List(req dto.ProjectListRequest) ([]dto.ProjectResponse, error)
	GetWorkflowSteps() ([]model.WorkflowStep, error)
	CreateApproval(req dto.ProjectApprovalCreateRequest) error
	UpdateApprovalStatusDirect(req dto.ProjectApprovalDirectUpdateRequest) error
	UpdateProjectStatusDirect(req dto.ProjectStatusDirectUpdateRequest) error
	UpdateProjectDetails(req dto.ProjectDetailsRepoUpdateRequest) error
	ResetSubsequentApprovals(projectID int, sequence int) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) Create(req dto.ProjectRepoCreateRequest) (*model.Project, error) {
	project := &model.Project{
		UUID:        req.UUID,
		ProjectName: req.ProjectName,
		Description: req.Description,
		Budget:      req.Budget,
		Status:      "Pending",
		CreatedBy:   req.UserID,
	}
	if err := r.db.Create(project).Error; err != nil {
		logrus.Errorf("DB Error creating project %s (user: %d): %v", req.ProjectName, req.UserID, err)
		return nil, err
	}
	return project, nil
}

func assembleProjects(rows []dto.FlatRow) []dto.ProjectResponse {
	projectMap := make(map[int]*dto.ProjectResponse)
	var orderedProjectIDs []int

	for _, row := range rows {
		p, exists := projectMap[row.ProjectID]
		if !exists {
			p = &dto.ProjectResponse{
				ID:          row.ProjectID,
				UUID:        row.ProjectUUID,
				ProjectName: row.ProjectName,
				Description: row.ProjectDescription,
				Budget:      row.ProjectBudget,
				Status:      row.ProjectStatus,
				CreatedBy:   row.ProjectCreatedBy,
				UpdatedBy:   row.ProjectUpdatedBy,
				CreatedAt:   row.ProjectCreatedAt,
				UpdatedAt:   row.ProjectUpdatedAt,
				DeletedAt:   row.ProjectDeletedAt,
				Approvals:   []dto.ProjectApprovalResponse{},
			}
			projectMap[row.ProjectID] = p
			orderedProjectIDs = append(orderedProjectIDs, row.ProjectID)
		}

		if row.ApprovalID != nil {
			var actionBy string
			if row.UserName != nil {
				actionBy = *row.UserName
			}

			var remarks string
			if row.ApprovalRemarks != nil {
				remarks = *row.ApprovalRemarks
			}

			appStatus := "Pending"
			if row.ApprovalStatus != nil {
				appStatus = *row.ApprovalStatus
			}

			stepName := ""
			if row.StepName != nil {
				stepName = *row.StepName
			}

			roleName := ""
			if row.StepRoleName != nil {
				roleName = *row.StepRoleName
			}

			sequence := 0
			if row.StepStepSequence != nil {
				sequence = *row.StepStepSequence
			}

			approval := dto.ProjectApprovalResponse{
				ID:       *row.ApprovalID,
				StepName: stepName,
				RoleName: roleName,
				Sequence: sequence,
				Status:   appStatus,
				ActionBy: actionBy,
				ActionAt: row.ApprovalActionAt,
				Remarks:  remarks,
			}
			p.Approvals = append(p.Approvals, approval)
		}
	}

	var projects []dto.ProjectResponse
	for _, id := range orderedProjectIDs {
		projects = append(projects, *projectMap[id])
	}
	return projects
}

func (r *projectRepository) GetByUUID(req dto.ProjectByUUIDRequest) (*dto.ProjectResponse, error) {
	var rows []dto.FlatRow
	err := r.db.Table("projects p").
		Select(`
			p.id AS project_id, p.uuid AS project_uuid, p.project_name, p.description AS project_description,
			p.budget AS project_budget, p.status AS project_status, p.created_by AS project_created_by,
			p.updated_by AS project_updated_by, p.created_at AS project_created_at, p.updated_at AS project_updated_at,
			p.deleted_at AS project_deleted_at,
			pa.id AS approval_id, pa.status AS approval_status, pa.action_by AS approval_action_by,
			pa.action_at AS approval_action_at, pa.remarks AS approval_remarks,
			ws.id AS step_id, ws.step_name, ws.role_name AS step_role_name, ws.step_sequence,
			u.id AS user_id, u.name AS user_name
		`).
		Joins("LEFT JOIN project_approvals pa ON p.id = pa.project_id").
		Joins("LEFT JOIN workflow_steps ws ON pa.step_id = ws.id").
		Joins("LEFT JOIN users u ON pa.action_by = u.id").
		Where("p.uuid = ? AND p.deleted_at IS NULL", req.UUID).
		Order("ws.step_sequence ASC").
		Scan(&rows).Error

	if err != nil {
		logrus.Errorf("DB Error getting project by UUID %s: %v", req.UUID, err)
		return nil, err
	}

	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	projects := assembleProjects(rows)
	if len(projects) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &projects[0], nil
}

func (r *projectRepository) List(req dto.ProjectListRequest) ([]dto.ProjectResponse, error) {
	var rows []dto.FlatRow
	err := r.db.Table("projects p").
		Select(`
			p.id AS project_id, p.uuid AS project_uuid, p.project_name, p.description AS project_description,
			p.budget AS project_budget, p.status AS project_status, p.created_by AS project_created_by,
			p.updated_by AS project_updated_by, p.created_at AS project_created_at, p.updated_at AS project_updated_at,
			p.deleted_at AS project_deleted_at,
			pa.id AS approval_id, pa.status AS approval_status, pa.action_by AS approval_action_by,
			pa.action_at AS approval_action_at, pa.remarks AS approval_remarks,
			ws.id AS step_id, ws.step_name, ws.role_name AS step_role_name, ws.step_sequence,
			u.id AS user_id, u.name AS user_name
		`).
		Joins("LEFT JOIN project_approvals pa ON p.id = pa.project_id").
		Joins("LEFT JOIN workflow_steps ws ON pa.step_id = ws.id").
		Joins("LEFT JOIN users u ON pa.action_by = u.id").
		Where("p.deleted_at IS NULL").
		Order("p.id DESC, ws.step_sequence ASC").
		Scan(&rows).Error

	if err != nil {
		logrus.Errorf("DB Error listing projects: %v", err)
		return nil, err
	}

	allProjects := assembleProjects(rows)

	if req.UserRole == "Admin" {
		return allProjects, nil
	}

	var filtered []dto.ProjectResponse
	for _, p := range allProjects {
		if p.CreatedBy == req.UserID {
			filtered = append(filtered, p)
			continue
		}

		var activeStep *dto.ProjectApprovalResponse
		for i := range p.Approvals {
			app := &p.Approvals[i]
			if app.Status == "Pending" {
				if activeStep == nil || app.Sequence < activeStep.Sequence {
					activeStep = app
				}
			}
		}

		if activeStep != nil && activeStep.RoleName == req.UserRole && p.Status == "Pending" {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}

func (r *projectRepository) GetWorkflowSteps() ([]model.WorkflowStep, error) {
	var steps []model.WorkflowStep
	if err := r.db.Order("step_sequence ASC").Find(&steps).Error; err != nil {
		logrus.Errorf("DB Error getting workflow steps: %v", err)
		return nil, err
	}
	return steps, nil
}

func (r *projectRepository) CreateApproval(req dto.ProjectApprovalCreateRequest) error {
	approval := &model.ProjectApproval{
		ProjectID: req.ProjectID,
		StepID:    req.StepID,
		Status:    req.Status,
	}
	if err := r.db.Create(approval).Error; err != nil {
		logrus.Errorf("DB Error creating project approval log: %v", err)
		return err
	}
	return nil
}

func (r *projectRepository) UpdateApprovalStatusDirect(req dto.ProjectApprovalDirectUpdateRequest) error {
	updates := map[string]interface{}{
		"status":    req.Status,
		"action_by": req.ActionBy,
		"action_at": req.ActionAt,
		"remarks":   req.Remarks,
	}
	if err := r.db.Model(&model.ProjectApproval{}).Where("id = ?", req.ApprovalID).Updates(updates).Error; err != nil {
		logrus.Errorf("DB Error updating project approval status directly ID %d: %v", req.ApprovalID, err)
		return err
	}
	return nil
}

func (r *projectRepository) ResetSubsequentApprovals(projectID int, sequence int) error {
	var approvalIDs []int
	err := r.db.Table("project_approvals pa").
		Select("pa.id").
		Joins("JOIN workflow_steps ws ON pa.step_id = ws.id").
		Where("pa.project_id = ? AND ws.step_sequence > ?", projectID, sequence).
		Scan(&approvalIDs).Error
	if err != nil {
		logrus.Errorf("DB Error finding subsequent approvals to reset: %v", err)
		return err
	}

	if len(approvalIDs) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"status":    "Pending",
		"action_by": nil,
		"action_at": nil,
		"remarks":   "",
	}
	if err := r.db.Model(&model.ProjectApproval{}).Where("id IN ?", approvalIDs).Updates(updates).Error; err != nil {
		logrus.Errorf("DB Error resetting subsequent approvals: %v", err)
		return err
	}
	return nil
}

func (r *projectRepository) UpdateProjectStatusDirect(req dto.ProjectStatusDirectUpdateRequest) error {
	updates := map[string]interface{}{
		"status":     req.Status,
		"updated_by": req.UpdatedBy,
	}
	if err := r.db.Model(&model.Project{}).Where("id = ?", req.ProjectID).Updates(updates).Error; err != nil {
		logrus.Errorf("DB Error updating project status directly ID %d: %v", req.ProjectID, err)
		return err
	}
	return nil
}

func (r *projectRepository) UpdateProjectDetails(req dto.ProjectDetailsRepoUpdateRequest) error {
	updates := map[string]interface{}{
		"updated_by": req.UserID,
	}
	if req.ProjectName != "" {
		updates["project_name"] = req.ProjectName
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Budget > 0 {
		updates["budget"] = req.Budget
	}

	if err := r.db.Model(&model.Project{}).Where("uuid = ?", req.UUID).Updates(updates).Error; err != nil {
		logrus.Errorf("DB Error updating project details UUID %s: %v", req.UUID, err)
		return err
	}
	return nil
}
