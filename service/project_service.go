package service

import (
	"errors"
	"fmt"
	"project-workflow-backend/dto"
	"project-workflow-backend/repository"
	"project-workflow-backend/util"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ProjectService interface {
	Create(req dto.ProjectServiceCreateRequest) (*dto.ProjectResponse, error)
	Approve(req dto.ProjectServiceApproveRequest) (*dto.ProjectResponse, error)
	List(req dto.ProjectServiceListRequest) ([]dto.ProjectResponse, error)
	Details(req dto.ProjectByUUIDRequest) (*dto.ProjectResponse, error)
	Update(req dto.ProjectServiceUpdateRequest) (*dto.ProjectResponse, error)
	UpdateApprovalStatus(req dto.ProjectServiceApprovalUpdateRequest) (*dto.ProjectResponse, error)
}

type projectService struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) Create(req dto.ProjectServiceCreateRequest) (*dto.ProjectResponse, error) {
	steps, err := s.repo.GetWorkflowSteps()
	if err != nil {
		return nil, err
	}
	if len(steps) == 0 {
		logrus.Warn("Workflow steps are not configured in the database")
		return nil, errors.New("workflow steps are not configured in the database. Please seed the 'workflow_steps' table first")
	}
	projectUUID := util.WithoutHypenGenUUID()
	if projectUUID == "" {
		projectUUID = uuid.New().String()
	}

	project, err := s.repo.Create(dto.ProjectRepoCreateRequest{
		ProjectName: req.ProjectName,
		Description: req.Description,
		Budget:      req.Budget,
		UUID:        projectUUID,
		UserID:      req.UserID,
	})
	if err != nil {
		return nil, err
	}

	for _, step := range steps {
		err := s.repo.CreateApproval(dto.ProjectApprovalCreateRequest{
			ProjectID: int(project.ID),
			StepID:    int(step.ID),
			Status:    "Pending",
		})
		if err != nil {
			return nil, err
		}
	}
	fullProject, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: project.UUID})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Project created successfully with dynamic approvals: UUID=%s, StepsCount=%d", fullProject.UUID, len(steps))
	return fullProject, nil
}

func (s *projectService) Approve(req dto.ProjectServiceApproveRequest) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.ProjectUUID})
	if err != nil {
		logrus.Warnf("Project approval failed: Project UUID %s not found", req.ProjectUUID)
		return nil, errors.New("project not found")
	}

	if project.Status == "Rejected" {
		logrus.Warnf("Project approval rejected: Project %s is already in Rejected state", project.UUID)
		return nil, errors.New("project has already been rejected")
	}
	if project.Status == "Approved" {
		logrus.Warnf("Project approval rejected: Project %s is already fully Approved", project.UUID)
		return nil, errors.New("project has already been fully approved")
	}
	var activeApproval *dto.ProjectApprovalResponse
	for i := range project.Approvals {
		app := &project.Approvals[i]
		if app.Status == "Pending" {
			if activeApproval == nil || app.Sequence < activeApproval.Sequence {
				activeApproval = app
			}
		}
	}

	if activeApproval == nil {
		logrus.Warnf("Project approval rejected: No pending approvals found for project %s", project.UUID)
		return nil, errors.New("project has already been fully approved")
	}

	if req.UserRole != activeApproval.RoleName {
		logrus.Warnf("Unauthorized approval step: user %d (role: %s) attempted approval on project %s (requires: %s)",
			req.UserID, req.UserRole, project.UUID, activeApproval.RoleName)
		return nil, fmt.Errorf("only a %s can approve or reject at this stage (%s)", activeApproval.RoleName, activeApproval.StepName)
	}

	now := time.Now()
	err = s.repo.UpdateApprovalStatusDirect(dto.ProjectApprovalDirectUpdateRequest{
		ApprovalID: activeApproval.ID,
		Status:     req.Status,
		ActionBy:   req.UserID,
		ActionAt:   now,
		Remarks:    req.Remarks,
	})
	if err != nil {
		return nil, err
	}

	previousStatus := project.Status
	newStatus := previousStatus
	if req.Status == "Rejected" {
		newStatus = "Rejected"
	} else {
		hasMorePending := false
		for _, app := range project.Approvals {
			if app.ID == activeApproval.ID {
				continue
			}
			if app.Status == "Pending" {
				hasMorePending = true
				break
			}
		}
		if !hasMorePending {
			newStatus = "Approved"
		}
	}

	err = s.repo.UpdateProjectStatusDirect(dto.ProjectStatusDirectUpdateRequest{
		ProjectID: project.ID,
		Status:    newStatus,
		UpdatedBy: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	refreshedProject, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.ProjectUUID})
	if err != nil {
		return nil, err
	}
	logrus.Infof("Project status transitioned successfully: UUID=%s, Step=%s, Action=%s, PreviousStatus=%s, NewStatus=%s, ProcessedBy=%d (%s)",
		project.UUID, activeApproval.StepName, req.Status, previousStatus, newStatus, req.UserID, req.UserRole)

	return refreshedProject, nil
}

func (s *projectService) List(req dto.ProjectServiceListRequest) ([]dto.ProjectResponse, error) {
	return s.repo.List(dto.ProjectListRequest{
		UserID:   req.UserID,
		UserRole: req.UserRole,
	})
}

func (s *projectService) Details(req dto.ProjectByUUIDRequest) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByUUID(req)
	if err != nil {
		logrus.Warnf("Project details not found: UUID %s", req.UUID)
		return nil, errors.New("project not found")
	}
	return project, nil
}

func (s *projectService) Update(req dto.ProjectServiceUpdateRequest) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.UUID})
	if err != nil {
		logrus.Warnf("Project update failed: UUID %s not found", req.UUID)
		return nil, errors.New("project not found")
	}

	err = s.repo.UpdateProjectDetails(dto.ProjectDetailsRepoUpdateRequest{
		UUID:        req.UUID,
		ProjectName: req.ProjectName,
		Description: req.Description,
		Budget:      req.Budget,
		UserID:      req.UserID,
	})
	if err != nil {
		return nil, err
	}

	// Reset all approvals (sequence > 0 resets all workflow steps)
	err = s.repo.ResetSubsequentApprovals(project.ID, 0)
	if err != nil {
		return nil, err
	}

	// Update project overall status back to Pending
	err = s.repo.UpdateProjectStatusDirect(dto.ProjectStatusDirectUpdateRequest{
		ProjectID: project.ID,
		Status:    "Pending",
		UpdatedBy: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	refreshed, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.UUID})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Project updated and workflow reset to Pending: UUID=%s, UpdatedBy=%d", refreshed.UUID, req.UserID)
	return refreshed, nil
}

func (s *projectService) UpdateApprovalStatus(req dto.ProjectServiceApprovalUpdateRequest) (*dto.ProjectResponse, error) {
	project, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.ProjectUUID})
	if err != nil {
		logrus.Warnf("Project approval update failed: Project UUID %s not found", req.ProjectUUID)
		return nil, errors.New("project not found")
	}

	targetRole := req.RoleName
	if targetRole == "" {
		if req.UserRole == "Admin" {
			return nil, errors.New("admin must specify the role_name of the step to update")
		}
		targetRole = req.UserRole
	}

	if req.UserRole != "Admin" && req.UserRole != targetRole {
		logrus.Warnf("Unauthorized approval update: user %d (role: %s) tried to update %s approval status on project %s",
			req.UserID, req.UserRole, targetRole, project.UUID)
		return nil, fmt.Errorf("only a %s or Admin can update this approval status", targetRole)
	}

	var targetApproval *dto.ProjectApprovalResponse
	for i := range project.Approvals {
		app := &project.Approvals[i]
		if app.RoleName == targetRole {
			targetApproval = app
			break
		}
	}

	if targetApproval == nil {
		return nil, fmt.Errorf("approval step for role %s not found in this project workflow", targetRole)
	}

	// 1. Enforce sequence: Cannot set a step to "Approved" if any prior step is not "Approved"
	if req.Status == "Approved" {
		for _, app := range project.Approvals {
			if app.Sequence < targetApproval.Sequence {
				if app.Status != "Approved" {
					return nil, fmt.Errorf("cannot approve this step because the previous step (%s) is not approved (current status: %s)", app.StepName, app.Status)
				}
			}
		}
	}

	now := time.Now()
	err = s.repo.UpdateApprovalStatusDirect(dto.ProjectApprovalDirectUpdateRequest{
		ApprovalID: targetApproval.ID,
		Status:     req.Status,
		ActionBy:   req.UserID,
		ActionAt:   now,
		Remarks:    req.Remarks,
	})
	if err != nil {
		return nil, err
	}

	// 2. Reset subsequent steps to "Pending" if the status is updated to "Rejected" or "Pending"
	if req.Status == "Rejected" || req.Status == "Pending" {
		err = s.repo.ResetSubsequentApprovals(project.ID, targetApproval.Sequence)
		if err != nil {
			return nil, err
		}
	}

	project, err = s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.ProjectUUID})
	if err != nil {
		return nil, err
	}

	hasRejected := false
	hasPending := false
	for _, app := range project.Approvals {
		if app.Status == "Rejected" {
			hasRejected = true
		} else if app.Status == "Pending" {
			hasPending = true
		}
	}

	previousStatus := project.Status
	newStatus := previousStatus
	if hasRejected {
		newStatus = "Rejected"
	} else if hasPending {
		newStatus = "Pending"
	} else {
		newStatus = "Approved"
	}

	err = s.repo.UpdateProjectStatusDirect(dto.ProjectStatusDirectUpdateRequest{
		ProjectID: project.ID,
		Status:    newStatus,
		UpdatedBy: req.UserID,
	})
	if err != nil {
		return nil, err
	}

	refreshed, err := s.repo.GetByUUID(dto.ProjectByUUIDRequest{UUID: req.ProjectUUID})
	if err != nil {
		return nil, err
	}

	logrus.Infof("Project approval status updated: UUID=%s, Step=%s, Action=%s, PreviousStatus=%s, NewStatus=%s, ProcessedBy=%d (%s)",
		project.UUID, targetRole, req.Status, previousStatus, newStatus, req.UserID, req.UserRole)
	return refreshed, nil
}
