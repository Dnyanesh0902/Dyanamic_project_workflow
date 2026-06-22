package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"project-workflow-backend/app"
	"project-workflow-backend/database"
	"project-workflow-backend/route"
	"project-workflow-backend/util"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func init() {
	envFile := flag.String("env", ".env", "specify the env file name")
	flag.Parse()

	if err := godotenv.Load(*envFile); err != nil {
		logrus.Fatalf("Error loading %s file: %v", *envFile, err)
	}
	util.InitializeLogger()
	logrus.Infof("Environment variables loaded successfully.")
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Initialize the database
	if err := database.InitCampMgmtDB(); err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer func() {
		if database.AttendaceCmrfDB != nil {
			if db, err := database.AttendaceCmrfDB.DB(); err == nil {
				db.Close()
			}
		}
	}()
	fmt.Println("DB Initialized !!!")

	app := app.InitApp()
	router := route.SetupRouter(app)

	router.Run(":" + os.Getenv("APP_PORT"))
}
