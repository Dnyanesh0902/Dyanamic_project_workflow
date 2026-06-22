package route

import (
	"os"
	"project-workflow-backend/app"
	"project-workflow-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(application *app.App) *gin.Engine {

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "OPTIONS", "PUT", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
		MaxAge:           86400,
	}))
	router.Use(middleware.RateLimiter())
	router.Use(middleware.RequestLoggerMiddleware())

	if os.Getenv("APP_ENV") == "production" {
		router.Use(middleware.PostmanCheck())
	}

	userGroup := router.Group("/api/users")
	{

		userGroup.POST("/create", application.UserController.Create)
		userGroup.POST("/login", application.UserController.Login)

		protected := userGroup.Group("")
		protected.Use(middleware.TokenAuthentication(), middleware.Timeout(10))
		{
			protected.POST("/list", application.UserController.List)
			protected.POST("/details", application.UserController.Details)
			protected.POST("/update", application.UserController.Update)
			protected.POST("/delete", application.UserController.Delete)
		}
	}

	projectGroup := router.Group("/api/projects")
	projectGroup.Use(middleware.TokenAuthentication(), middleware.Timeout(10))
	{
		projectGroup.POST("/create", application.ProjectController.Create)
		projectGroup.POST("/approve", application.ProjectController.Approve)
		projectGroup.POST("/list", application.ProjectController.List)
		projectGroup.POST("/details", application.ProjectController.Details)
		projectGroup.POST("/update", application.ProjectController.Update)
		projectGroup.POST("/approval-update", application.ProjectController.UpdateApprovalStatus)
	}

	return router
}
