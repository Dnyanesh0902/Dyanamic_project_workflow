package controller

import (
	"encoding/json"
	"project-workflow-backend/dto"
	"project-workflow-backend/service"
	"project-workflow-backend/util"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type ProjectController interface {
	Create(c *gin.Context)
	Approve(c *gin.Context)
	List(c *gin.Context)
	Details(c *gin.Context)
	Update(c *gin.Context)
	UpdateApprovalStatus(c *gin.Context)
}

type projectController struct {
	svc service.ProjectService
}

func NewProjectController(svc service.ProjectService) ProjectController {
	return &projectController{svc: svc}
}

func (ctrl *projectController) Create(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController Create panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	userIDStr := c.Request.Header.Get("auth_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		util.UnauthorizedResponse(c, "Unauthorized: missing user ID")
		return
	}

	var req dto.ProjectCreateRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		util.ValidationResponse(c, "Invalid request payload")
		return
	}

	validationResp := util.ValidateRequest(c, req)
	if validationResp != nil {
		logrus.Error("Validation failed:", validationResp)
		util.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.svc.Create(dto.ProjectServiceCreateRequest{
		UserID:               userID,
		ProjectCreateRequest: req,
	})
	if err != nil {
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.CreatedResponse(c, "Project created successfully", res)
}

func (ctrl *projectController) Approve(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController Approve panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	userIDStr := c.Request.Header.Get("auth_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		util.UnauthorizedResponse(c, "Unauthorized: missing user ID")
		return
	}

	userRole := c.Request.Header.Get("user_type")
	if userRole == "" {
		util.UnauthorizedResponse(c, "Unauthorized: missing user role")
		return
	}

	var req dto.ProjectApprovalRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		util.ValidationResponse(c, "Invalid request payload")
		return
	}

	validationResp := util.ValidateRequest(c, req)
	if validationResp != nil {
		logrus.Error("Validation failed:", validationResp)
		util.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.svc.Approve(dto.ProjectServiceApproveRequest{
		UserID:                 userID,
		UserRole:               userRole,
		ProjectApprovalRequest: req,
	})
	if err != nil {
		if err.Error() == "project not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		if err.Error() == "project has already been rejected" ||
			err.Error() == "project has already been fully approved" ||
			err.Error() == "invalid approval stage or transition not allowed" ||
			strings.Contains(err.Error(), "only a") {
			util.BadRequestResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Project status updated successfully", res)
}

func (ctrl *projectController) List(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController List panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	userIDStr := c.Request.Header.Get("auth_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		util.UnauthorizedResponse(c, "Unauthorized: missing user ID")
		return
	}

	userRole := c.Request.Header.Get("user_type")

	res, err := ctrl.svc.List(dto.ProjectServiceListRequest{
		UserID:   userID,
		UserRole: userRole,
	})
	if err != nil {
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Projects retrieved successfully", res)
}

func (ctrl *projectController) Details(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController Details panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.ProjectByUUIDRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		util.ValidationResponse(c, "Invalid request payload")
		return
	}

	validationResp := util.ValidateRequest(c, req)
	if validationResp != nil {
		logrus.Error("Validation failed:", validationResp)
		util.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.svc.Details(req)
	if err != nil {
		if err.Error() == "project not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Project details retrieved successfully", res)
}

func (ctrl *projectController) Update(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController Update panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	userIDStr := c.Request.Header.Get("auth_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		util.UnauthorizedResponse(c, "Unauthorized: missing user ID")
		return
	}

	var req dto.ProjectUpdateRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		util.ValidationResponse(c, "Invalid request payload")
		return
	}

	validationResp := util.ValidateRequest(c, req)
	if validationResp != nil {
		logrus.Error("Validation failed:", validationResp)
		util.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.svc.Update(dto.ProjectServiceUpdateRequest{
		UserID:               userID,
		ProjectUpdateRequest: req,
	})
	if err != nil {
		if err.Error() == "project not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Project updated successfully", res)
}

func (ctrl *projectController) UpdateApprovalStatus(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("ProjectController UpdateApprovalStatus panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	userIDStr := c.Request.Header.Get("auth_user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		util.UnauthorizedResponse(c, "Unauthorized: missing user ID")
		return
	}

	userRole := c.Request.Header.Get("user_type")
	if userRole == "" {
		util.UnauthorizedResponse(c, "Unauthorized: missing user role")
		return
	}

	var req dto.ProjectApprovalUpdateRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		util.ValidationResponse(c, "Invalid request payload")
		return
	}

	validationResp := util.ValidateRequest(c, req)
	if validationResp != nil {
		logrus.Error("Validation failed:", validationResp)
		util.ValidationResponse(c, validationResp.(string))
		return
	}

	res, err := ctrl.svc.UpdateApprovalStatus(
		dto.ProjectServiceApprovalUpdateRequest{
			UserID:                       userID,
			UserRole:                     userRole,
			ProjectApprovalUpdateRequest: req,
		})
	if err != nil {
		if err.Error() == "project not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "only a") {
			util.BadRequestResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Project approval status updated successfully", res)
}
