package controller

import (
	"encoding/json"
	"project-workflow-backend/dto"
	"project-workflow-backend/service"
	"project-workflow-backend/util"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserController interface {
	Create(c *gin.Context)
	Login(c *gin.Context)
	List(c *gin.Context)
	Details(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)
}

type userController struct {
	svc service.UserService
}

func NewUserController(svc service.UserService) UserController {
	return &userController{svc: svc}
}

func (ctrl *userController) Create(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController Create panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.UserCreateRequest
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

	res, err := ctrl.svc.Create(req)
	if err != nil {
		if err.Error() == "email already in use" {
			util.ConflictResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.CreatedResponse(c, "User created successfully", res)
}

func (ctrl *userController) Login(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController Login panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.UserLoginRequest
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

	res, err := ctrl.svc.Login(req)
	if err != nil {
		util.UnauthorizedResponse(c, err.Error())
		return
	}

	util.SuccessResponse(c, "Login successful", res)
}

func (ctrl *userController) List(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController List panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	res, err := ctrl.svc.List()
	if err != nil {
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "Users retrieved successfully", res)
}

func (ctrl *userController) Details(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController Details panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.UserByUUIDRequest
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
		if err.Error() == "user not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "User details retrieved successfully", res)
}

func (ctrl *userController) Update(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController Update panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.UserUpdateRequest
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

	res, err := ctrl.svc.Update(req)
	if err != nil {
		if err.Error() == "user not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		if err.Error() == "email already in use" {
			util.ConflictResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "User updated successfully", res)
}

func (ctrl *userController) Delete(c *gin.Context) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logrus.Error("UserController Delete panic:", panicInfo)
			util.InternalServerErrorResponse(c, panicInfo.(error))
		}
	}()

	var req dto.UserByUUIDRequest
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

	err := ctrl.svc.Delete(req)
	if err != nil {
		if err.Error() == "user not found" {
			util.NotFoundResponse(c, err.Error())
			return
		}
		util.InternalServerErrorResponse(c, err)
		return
	}

	util.SuccessResponse(c, "User deleted successfully", nil)
}
