package app

import (
	"project-workflow-backend/controller"
	"project-workflow-backend/database"
	"project-workflow-backend/repository"
	"project-workflow-backend/service"
)

type App struct {
	ConsentRequestController *ConsentRequestControllers
	UserController           controller.UserController
	ProjectController        controller.ProjectController
}
type ConsentRequestControllers struct {
}

type Services struct {
	UserSvc    service.UserService
	ProjectSvc service.ProjectService
}
type Repositories struct {
	UserRepo    repository.UserRepository
	ProjectRepo repository.ProjectRepository
}

func InitApp() *App {

	repos := initRepositories()
	services := initServices(repos)
	controllers := initControllers(services)

	return &App{
		ConsentRequestController: controllers,
		UserController:           controller.NewUserController(services.UserSvc),
		ProjectController:        controller.NewProjectController(services.ProjectSvc),
	}
}
func initRepositories() *Repositories {
	return &Repositories{
		UserRepo:    repository.NewUserRepository(database.AttendaceCmrfDB),
		ProjectRepo: repository.NewProjectRepository(database.AttendaceCmrfDB),
	}
}
func initServices(repos *Repositories) *Services {
	return &Services{
		UserSvc:    service.NewUserService(repos.UserRepo),
		ProjectSvc: service.NewProjectService(repos.ProjectRepo),
	}
}
func initControllers(services *Services) *ConsentRequestControllers {
	return &ConsentRequestControllers{}
}
