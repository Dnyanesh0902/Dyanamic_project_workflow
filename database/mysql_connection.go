package database

import (
	"project-workflow-backend/config"
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
	AttendaceCmrfDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Error("Failed to connect to the database:", err)
		return err
	}
	sqlDB, err := AttendaceCmrfDB.DB()
	if err != nil {
		logrus.Error("Failed to set up 'AttendaceCmrfDB' database connection pool: ", err)
	}
	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}
