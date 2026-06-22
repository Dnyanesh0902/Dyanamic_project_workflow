package database

import (
	"project-workflow-backend/config"
	"project-workflow-backend/model"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var AttendaceCmrfDB *gorm.DB

func InitCampMgmtDB() error {
	dbConfig := config.GetPrimaryMySQLDBConfig()
	dsn := dbConfig.Username + ":" + dbConfig.Password + "@tcp(" + dbConfig.Host + ":" + dbConfig.Port + ")/" + dbConfig.Database + "?charset=utf8mb4&parseTime=True&loc=Local&sql_mode=''"
	logrus.Info("CrowdFunding@dsn : ", dsn)
	var err error
	for i := 0; i < 10; i++ {
		AttendaceCmrfDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		logrus.Warnf("Failed to connect to the database (attempt %d/10): %v. Retrying in 2 seconds...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		logrus.Error("Failed to connect to the database after 10 attempts:", err)
		return err
	}
	sqlDB, err := AttendaceCmrfDB.DB()
	if err != nil {
		logrus.Error("Failed to set up 'AttendaceCmrfDB' database connection pool: ", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto Migrate the database schema
	err = AttendaceCmrfDB.AutoMigrate(
		&model.User{},
		&model.Project{},
		&model.ProjectApproval{},
		&model.WorkflowStep{},
	)
	if err != nil {
		logrus.Error("Failed to auto migrate database schema:", err)
		return err
	}

	// Seed default workflow steps if table is empty
	var count int64
	if err := AttendaceCmrfDB.Model(&model.WorkflowStep{}).Count(&count).Error; err == nil && count == 0 {
		steps := []model.WorkflowStep{
			{StepName: "Relationship Manager Review", RoleName: "RM", StepSequence: 1},
			{StepName: "Branch Head Approval", RoleName: "BH", StepSequence: 2},
			{StepName: "Vertical Head Verification", RoleName: "VH", StepSequence: 3},
		}
		if err := AttendaceCmrfDB.Create(&steps).Error; err != nil {
			logrus.Error("Failed to seed workflow steps:", err)
		} else {
			logrus.Info("Workflow steps seeded successfully.")
		}
	}

	return nil
}
